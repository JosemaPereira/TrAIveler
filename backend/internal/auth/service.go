package auth

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"

	"github.com/JosemaPereira/TrAIveler/backend/internal/auth/ratelimit"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
	"github.com/JosemaPereira/TrAIveler/backend/internal/observability"
	"github.com/JosemaPereira/TrAIveler/backend/internal/subscription"
)

// subscriptionCreator is the auth service's minimal, consumer-owned port for
// creating a subscription during registration. It is satisfied structurally by
// *subscription.Service; keeping it local (rather than importing the concrete
// service type) narrows the dependency to the single method auth needs and keeps
// the seam mockable (Local-Port pattern, see docs patterns).
type subscriptionCreator interface {
	CreateSubscription(ctx context.Context, userID, planID, token string) (*subscription.Subscription, error)
}

// dbUnavailableRetryAfterSeconds is the Retry-After hint attached to a
// service_unavailable response caused by a database connectivity failure
// during register or login (issue #207 Sub-item 1). Matches
// jwt.Refresher's own constant for the same class of failure on
// POST /auth/refresh (issue #192 Task A) and the documented example in
// specs/008-auth-collaboration-ux/contracts/api.md.
const dbUnavailableRetryAfterSeconds = 30

// classifyStorageErr converts a DB-backed repository/subscriber call's
// failure into the error Register/Login should return: a
// service_unavailable DomainError (domainerrors.ServiceUnavailableFromDB) for
// a database connectivity failure, so errors.HandleError maps it to 503 per
// specs/008-auth-collaboration-ux/contracts/api.md's "Database unavailable";
// otherwise err unchanged, so a normal repository error (not-found,
// constraint violation, ...) keeps falling through to its existing handling
// (generic internal_error/500, or a caller-specific domain error).
//
// This mirrors jwt.classifyStorageErr (internal/auth/jwt/refresher.go,
// unexported there too) rather than sharing it: auth already imports auth/jwt
// for token issuance, so importing back would cycle. Both call the same
// underlying errors.ServiceUnavailableFromDB.
func classifyStorageErr(err error) error {
	if svcErr := domainerrors.ServiceUnavailableFromDB(err, dbUnavailableRetryAfterSeconds); svcErr != nil {
		return svcErr
	}
	return err
}

// Service contains authentication business logic: registration (uniqueness,
// hashing, optional subscription) and login (credential validation, progressive
// rate limiting, security-event logging). It has no knowledge of HTTP, SQL, or
// JWT/cookie minting — token/session issuance is composed at a higher layer.
type Service struct {
	users      UserRepository
	subscriber subscriptionCreator
	limiter    *ratelimit.Limiter
	bcryptCost int
}

// NewService builds a Service from its collaborators: the user repository, the
// subscription-creation port, the login rate limiter, and the bcrypt cost used
// when hashing new passwords (wired from config.Auth.BcryptCost; an out-of-range
// value is corrected by HashPassword).
func NewService(users UserRepository, subscriber subscriptionCreator, limiter *ratelimit.Limiter, bcryptCost int) *Service {
	return &Service{users: users, subscriber: subscriber, limiter: limiter, bcryptCost: bcryptCost}
}

// Register creates a new admin user after validating the request and enforcing
// email uniqueness. Ordering is FK-safe: the user row is inserted first (the
// subscriptions.user_id foreign key requires it), then — only when a payment
// token is supplied — a subscription is charged/created and the user is flipped
// to has_subscription=true via an optimistic-locked update. A duplicate email is
// a 409 conflict (enumeration is not a concern on register). Passwords, hashes,
// and tokens are never logged.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if _, err := s.users.GetUserByEmail(ctx, req.Email); err == nil {
		return nil, domainerrors.Conflict("Email already registered")
	} else if !isNotFound(err) {
		return nil, classifyStorageErr(err)
	}

	hash, err := HashPassword(req.Password, s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &User{
		ID:           uuid.NewString(),
		Email:        req.Email,
		PasswordHash: hash,
		Role:         RoleAdmin,
		FullName:     req.FullName,
	}
	if err := s.users.CreateUser(ctx, user); err != nil {
		return nil, classifyStorageErr(err)
	}

	sub, err := s.maybeCreateSubscription(ctx, user, req.PaymentMethodToken)
	if err != nil {
		return nil, err
	}

	observability.LogSecurityEvent(
		correlationID(ctx),
		observability.EventAuthRegistration,
		user.ID,
		observability.SeverityInfo,
		"", "",
		map[string]any{"has_subscription": user.HasSubscription},
	)

	return &RegisterResponse{User: *user, Subscription: sub}, nil
}

// maybeCreateSubscription creates a subscription for the just-persisted user when
// a payment token is present, flipping has_subscription true and persisting that
// change. With no token it is a no-op returning (nil, nil). The stub provider
// always succeeds, so the error path is defensive.
func (s *Service) maybeCreateSubscription(
	ctx context.Context, user *User, paymentToken string,
) (*subscription.Subscription, error) {
	if paymentToken == "" {
		return nil, nil //nolint:nilnil // "no subscription" is a valid, expected outcome, not an error.
	}

	sub, err := s.subscriber.CreateSubscription(ctx, user.ID, subscription.DefaultPlanID, paymentToken)
	if err != nil {
		return nil, classifyStorageErr(err)
	}

	user.HasSubscription = true
	if err := s.users.UpdateUser(ctx, user); err != nil {
		return nil, classifyStorageErr(err)
	}

	return sub, nil
}

// CurrentUser resolves the account behind an authenticated request, given the
// user id the access token carried. It is the read side of session bootstrap:
// the browser holds only HTTP-only cookies it cannot inspect, so after a page
// reload this is the only way to learn who is logged in.
//
// The record comes from the database rather than the token's claims on purpose —
// has_subscription can go stale within an access token's lifetime (see
// auth/jwt.Claims), and a caller asking "who am I" wants the current answer.
//
// An unknown id yields the repository's NotFound unchanged; mapping that onto a
// session-level response is the handler's job, since the service has no notion
// of a session.
func (s *Service) CurrentUser(ctx context.Context, userID string) (*CurrentUserResponse, error) {
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &CurrentUserResponse{User: *user}, nil
}

// Login validates credentials with progressive, email-keyed rate limiting and
// emits security events. It deliberately does NOT mint JWTs or set cookies —
// that composition happens at a higher layer. The rate-limit check runs before
// bcrypt (so a locked-out account never pays the hashing cost), and unknown
// email and bad password return an identical 401 to avoid user enumeration.
func (s *Service) Login(
	ctx context.Context, req LoginRequest, ipAddress, userAgent string,
) (*LoginResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	correlation := correlationID(ctx)

	if delay := s.limiter.CheckRateLimit(req.Email); delay > 0 {
		s.logLoginFailure(correlation, "", req.Email, "rate_limited", ipAddress, userAgent)
		retryAfter := int(math.Ceil(delay.Seconds()))
		return nil, domainerrors.RateLimited(retryAfter)
	}

	user, err := s.users.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if isNotFound(err) {
			s.limiter.RecordFailure(req.Email)
			s.logLoginFailure(correlation, "", req.Email, "unknown_email", ipAddress, userAgent)
			return nil, domainerrors.Unauthorized("Invalid credentials")
		}
		return nil, classifyStorageErr(err)
	}

	if ComparePassword(user.PasswordHash, req.Password) != nil {
		s.limiter.RecordFailure(req.Email)
		user.FailedLoginAttempts++
		if err := s.users.UpdateUser(ctx, user); err != nil {
			return nil, classifyStorageErr(err)
		}
		s.logLoginFailure(correlation, user.ID, req.Email, "invalid_credentials", ipAddress, userAgent)
		return nil, domainerrors.Unauthorized("Invalid credentials")
	}

	s.limiter.Reset(req.Email)
	if user.FailedLoginAttempts > 0 {
		user.FailedLoginAttempts = 0
		if err := s.users.UpdateUser(ctx, user); err != nil {
			return nil, classifyStorageErr(err)
		}
	}

	observability.LogSecurityEvent(
		correlation, observability.EventAuthLoginSuccess, user.ID, observability.SeverityInfo,
		ipAddress, userAgent, map[string]any{"user_email": req.Email},
	)

	return &LoginResponse{User: *user}, nil
}

// logLoginFailure emits a warning-level auth_login_failure event with the given
// reason, keeping the login flow's several failure branches DRY. userID is empty
// when the account could not be resolved (unknown email or pre-lookup lockout).
func (s *Service) logLoginFailure(correlation, userID, email, reason, ipAddress, userAgent string) {
	observability.LogSecurityEvent(
		correlation, observability.EventAuthLoginFailure, userID, observability.SeverityWarning,
		ipAddress, userAgent, map[string]any{"user_email": email, "failure_reason": reason},
	)
}

// correlationID pulls the request correlation/trace id from the context (set by
// the request-id middleware), returning "" when absent so unit tests and
// background callers still log cleanly.
func correlationID(ctx context.Context) string {
	id, _ := middleware.RequestIDFromContext(ctx)
	return id
}
