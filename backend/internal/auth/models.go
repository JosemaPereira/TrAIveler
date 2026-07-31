package auth

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/subscription"
)

// Role values accepted by User.Role. These mirror the CHECK constraint on the
// users table (migrations/001_create_users_table.sql). Role is immutable
// after creation (docs/data-model.md §User).
const (
	// RoleAdmin owns trips and approves/rejects suggestions; the default role
	// assigned at registration.
	RoleAdmin = "admin"
	// RolePartner is a collaborator: views shared trips and submits suggestions.
	RolePartner = "partner"
)

// maxFullNameLength bounds full_name per the register contract (1-100 chars,
// specs/008-auth-collaboration-ux/contracts/api.md).
const maxFullNameLength = 100

// User is the authenticated account domain model.
//
// The field set matches the live users table (migrations/001 +
// 005_alter_users_add_auth_fields.sql), which is the source of truth. Role is
// included even though issue #167's field list omits it: it is a NOT NULL
// column central to authorization (docs/data-model.md §User), so a faithful
// domain model must carry it. subscription_id and last_login_at are omitted:
// the former is not a column (migration 005 dropped it in favor of
// has_subscription), and the latter is not needed by the auth DTOs here.
//
// db tags drive repository row scanning; json tags drive the API wire format.
// Sensitive/internal fields are excluded from JSON (json:"-"): the password
// hash is never returned (docs/data-model.md §User), and neither are the
// failed-login counters, email-verification flag, or optimistic-lock version,
// none of which appear in the register/login response contract.
type User struct {
	ID                  string     `json:"id" db:"id"`
	Email               string     `json:"email" db:"email"`
	PasswordHash        string     `json:"-" db:"password_hash"`
	Role                string     `json:"-" db:"role"`
	FullName            string     `json:"full_name" db:"full_name"`
	HasSubscription     bool       `json:"has_subscription" db:"has_subscription"`
	FailedLoginAttempts int        `json:"-" db:"failed_login_attempts"`
	LastFailedLoginAt   *time.Time `json:"-" db:"last_failed_login_at"`
	EmailVerified       bool       `json:"-" db:"email_verified"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"-" db:"updated_at"`
	Version             int64      `json:"-" db:"version"`
}

// RefreshToken is a stored refresh-token row (migrations/002). Only the SHA-256
// hash of the opaque token is persisted, never the token itself
// (docs/security.md). RevokedAt is nil while the token is live. The type is
// server-internal, so it carries db tags only.
type RefreshToken struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	CreatedAt time.Time  `db:"created_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}

// Validate checks the User's domain invariants (email present, role valid),
// collecting every failure. Input-shape validation of untrusted request
// bodies lives on RegisterRequest/LoginRequest; this guards the constructed
// domain entity before it is persisted. Returns nil when u is valid.
func (u *User) Validate() *domainerrors.DomainError {
	var fields []domainerrors.ValidationError

	if strings.TrimSpace(u.Email) == "" {
		fields = append(fields, domainerrors.ValidationError{
			Field: "email",
			Error: "Email is required",
		})
	}

	if u.Role != RoleAdmin && u.Role != RolePartner {
		fields = append(fields, domainerrors.ValidationError{
			Field: "role",
			Error: "Role must be 'admin' or 'partner'",
		})
	}

	if len(fields) > 0 {
		return domainerrors.Validation("One or more fields failed validation", fields...)
	}

	return nil
}

// RegisterRequest is the POST /auth/register request body. PaymentMethodToken
// is optional: when present it triggers (stub) subscription creation, marking
// the new user as subscribed.
type RegisterRequest struct {
	Email              string `json:"email"`
	Password           string `json:"password"`
	FullName           string `json:"full_name"`
	PaymentMethodToken string `json:"payment_method_token,omitempty"`
}

// Validate checks the registration input: a well-formed email, a full name of
// 1-100 characters, and a password meeting the strength rules (delegated to
// ValidatePassword). All failing fields are collected so the client can fix
// them in one round trip.
func (r *RegisterRequest) Validate() *domainerrors.DomainError {
	var fields []domainerrors.ValidationError

	fields = append(fields, validateEmailField(r.Email)...)
	fields = append(fields, validateFullNameField(r.FullName)...)
	fields = append(fields, passwordFields(r.Password)...)

	if len(fields) > 0 {
		return domainerrors.Validation("One or more fields failed validation", fields...)
	}

	return nil
}

// RegisterResponse is the POST /auth/register success body. Subscription is
// populated only when the request carried a payment token (json omitempty),
// matching the API contract's "only present if payment_method_token provided".
type RegisterResponse struct {
	User         User                       `json:"user"`
	Subscription *subscription.Subscription `json:"subscription,omitempty"`
}

// LoginRequest is the POST /auth/login request body.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate checks the login input: a well-formed email and a non-empty
// password. Login deliberately does not enforce password *strength* (only
// presence) so the rules can evolve without locking out existing accounts;
// credential correctness is checked later against the stored hash.
func (r *LoginRequest) Validate() *domainerrors.DomainError {
	var fields []domainerrors.ValidationError

	fields = append(fields, validateEmailField(r.Email)...)

	if r.Password == "" {
		fields = append(fields, domainerrors.ValidationError{
			Field: "password",
			Error: "Password is required",
		})
	}

	if len(fields) > 0 {
		return domainerrors.Validation("One or more fields failed validation", fields...)
	}

	return nil
}

// LoginResponse is the POST /auth/login success body.
type LoginResponse struct {
	User User `json:"user"`
}

// CurrentUserResponse is the GET /auth/me success body. It wraps the user in the
// same "user" envelope register and login use, so a client can decode all three
// with one type rather than special-casing this one.
type CurrentUserResponse struct {
	User User `json:"user"`
}

// validateEmailField reports an email that is empty or not RFC 5322-parseable.
// It uses net/mail rather than a hand-rolled regex, which under-/over-matches
// the RFC in subtle ways.
func validateEmailField(email string) []domainerrors.ValidationError {
	if strings.TrimSpace(email) == "" {
		return []domainerrors.ValidationError{{Field: "email", Error: "Email is required"}}
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return []domainerrors.ValidationError{{Field: "email", Error: "Email must be a valid email address"}}
	}
	return nil
}

// validateFullNameField reports a full name that is empty (after trimming) or
// longer than maxFullNameLength.
func validateFullNameField(fullName string) []domainerrors.ValidationError {
	trimmed := strings.TrimSpace(fullName)
	if trimmed == "" {
		return []domainerrors.ValidationError{{Field: "full_name", Error: "Full name is required"}}
	}
	if len([]rune(trimmed)) > maxFullNameLength {
		return []domainerrors.ValidationError{{
			Field: "full_name",
			Error: "Full name must be at most 100 characters",
		}}
	}
	return nil
}

// passwordFields adapts ValidatePassword's combined error into the flat
// ValidationError slice the request validators accumulate. ValidatePassword
// returns a *DomainError carrying per-rule field entries; those are lifted
// out so a caller sees each failed password rule alongside its other field
// errors.
func passwordFields(password string) []domainerrors.ValidationError {
	err := ValidatePassword(password)
	if err == nil {
		return nil
	}
	var domainErr *domainerrors.DomainError
	if errors.As(err, &domainErr) && len(domainErr.Fields) > 0 {
		return domainErr.Fields
	}
	return []domainerrors.ValidationError{{Field: "password", Error: "Password is invalid"}}
}
