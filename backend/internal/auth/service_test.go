//go:build test

package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/internal/auth"
	authmocks "github.com/JosemaPereira/TrAIveler/backend/internal/auth/mocks"
	"github.com/JosemaPereira/TrAIveler/backend/internal/auth/ratelimit"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/subscription"
)

const (
	testEmail    = "ada@example.com"
	testPassword = "CorrectHorse1!"
	testFullName = "Ada Lovelace"
	testToken    = "tok_visa_demo"
	// testBcryptCost is bcrypt.MinCost (4): the tests don't assert a specific cost,
	// so the cheapest valid cost keeps hashing fast.
	testBcryptCost = 4
)

// fakeConnectivityError stands in for a pgx connection-level failure (dial
// refused, timeout, DNS failure, ...), duplicated here rather than imported
// from another package's test file (same reasoning as
// internal/auth/jwt/refresher_test.go's own copy: pgx's real connectivity
// errors are unexported and only constructible by actually dialing, so this
// minimal type satisfies the same duck-typed `interface{ SafeToRetry() bool
// }` that pgconn.SafeToRetry checks for via errors.As — see
// errors.ServiceUnavailableFromDB).
type fakeConnectivityError struct {
	cause error
}

func (e *fakeConnectivityError) Error() string     { return "dial: " + e.cause.Error() }
func (e *fakeConnectivityError) Unwrap() error     { return e.cause }
func (e *fakeConnectivityError) SafeToRetry() bool { return true }

// newConnRefusedError builds a connection-level failure, e.g. a database
// outage during POST /auth/register or POST /auth/login (issue #207 Sub-item 1).
func newConnRefusedError() error {
	return &fakeConnectivityError{cause: errors.New("connection refused")}
}

// fakeSubscriptionCreator is a hand-written stand-in for the unexported
// subscriptionCreator port (single method, satisfied structurally by
// *subscription.Service). Per docs/mock-standards.md a hand fake is preferred
// over a generated mock for a one-method consumer-owned port.
type fakeSubscriptionCreator struct {
	sub    *subscription.Subscription
	err    error
	called bool
	gotArg struct {
		userID, planID, token string
	}
}

func (f *fakeSubscriptionCreator) CreateSubscription(
	_ context.Context, userID, planID, token string,
) (*subscription.Subscription, error) {
	f.called = true
	f.gotArg.userID, f.gotArg.planID, f.gotArg.token = userID, planID, token
	return f.sub, f.err
}

func hashFor(t *testing.T) string {
	t.Helper()
	hash, err := auth.HashPassword(testPassword, testBcryptCost)
	require.NoError(t, err)
	return hash
}

func validRegisterRequest() auth.RegisterRequest {
	return auth.RegisterRequest{Email: testEmail, Password: testPassword, FullName: testFullName}
}

// --- Register -------------------------------------------------------------

// TestUnitRegister_NewEmailWithoutToken_CreatesUnsubscribedUser verifies the
// happy path with no payment token: the user is created with has_subscription
// false and no subscription is attempted.
func TestUnitRegister_NewEmailWithoutToken_CreatesUnsubscribedUser(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	users.EXPECT().
		GetUserByEmail(mock.Anything, testEmail).
		Return(nil, domainerrors.NotFound("user", testEmail)).
		Once()
	users.EXPECT().
		CreateUser(mock.Anything, mock.MatchedBy(func(u *auth.User) bool {
			return u.ID != "" && u.Email == testEmail && u.Role == auth.RoleAdmin &&
				u.PasswordHash != "" && u.PasswordHash != testPassword && !u.HasSubscription
		})).
		Return(nil).
		Once()
	creator := &fakeSubscriptionCreator{}

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Register(context.Background(), validRegisterRequest())

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.User.HasSubscription)
	assert.Nil(t, resp.Subscription)
	assert.False(t, creator.called, "subscription must not be created without a token")
}

// TestUnitRegister_NewEmailWithToken_CreatesSubscribedUser verifies that a
// payment token drives subscription creation (with the default plan), flips
// has_subscription true, persists the update, and returns the subscription.
func TestUnitRegister_NewEmailWithToken_CreatesSubscribedUser(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	users.EXPECT().
		GetUserByEmail(mock.Anything, testEmail).
		Return(nil, domainerrors.NotFound("user", testEmail)).
		Once()
	users.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(nil).Once()
	users.EXPECT().
		UpdateUser(mock.Anything, mock.MatchedBy(func(u *auth.User) bool {
			return u.HasSubscription
		})).
		Return(nil).
		Once()

	wantSub := &subscription.Subscription{ID: "sub-1", Status: subscription.StatusActive}
	creator := &fakeSubscriptionCreator{sub: wantSub}

	req := validRegisterRequest()
	req.PaymentMethodToken = testToken

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Register(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.User.HasSubscription)
	require.NotNil(t, resp.Subscription)
	assert.Equal(t, "sub-1", resp.Subscription.ID)

	assert.True(t, creator.called)
	assert.Equal(t, subscription.DefaultPlanID, creator.gotArg.planID)
	assert.Equal(t, testToken, creator.gotArg.token)
	assert.NotEmpty(t, creator.gotArg.userID, "subscription must be created for the persisted user id")
}

// TestUnitRegister_ExistingEmail_ReturnsConflict verifies that a taken email is
// rejected with a 409 conflict before any user is created.
func TestUnitRegister_ExistingEmail_ReturnsConflict(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	users.EXPECT().
		GetUserByEmail(mock.Anything, testEmail).
		Return(&auth.User{ID: "existing", Email: testEmail}, nil).
		Once()
	creator := &fakeSubscriptionCreator{}

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Register(context.Background(), validRegisterRequest())

	require.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "conflict", domainErr.Code)
}

// TestUnitRegister_InvalidRequest_ReturnsValidationWithoutTouchingRepo verifies
// input validation runs before any repository call.
func TestUnitRegister_InvalidRequest_ReturnsValidationWithoutTouchingRepo(t *testing.T) {
	users := authmocks.NewMockUserRepository(t) // no expects: must not be touched
	creator := &fakeSubscriptionCreator{}

	req := validRegisterRequest()
	req.Email = "" // invalid

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Register(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "validation_failed", domainErr.Code)
	assert.False(t, creator.called)
}

// TestUnitRegister_LookupFails_PropagatesUnexpectedError verifies a non-NotFound
// lookup error is propagated rather than treated as "email available".
func TestUnitRegister_LookupFails_PropagatesUnexpectedError(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	dbErr := errors.New("connection reset")
	users.EXPECT().GetUserByEmail(mock.Anything, testEmail).Return(nil, dbErr).Once()
	creator := &fakeSubscriptionCreator{}

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Register(context.Background(), validRegisterRequest())

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, dbErr)
}

// assertServiceUnavailable verifies err is a "service_unavailable" DomainError
// carrying the 30-second Retry-After hint, matching /auth/refresh's own
// classification (jwt.classifyStorageErr, issue #192 Task A) so register/login
// behave consistently under a database outage (issue #207 Sub-item 1).
func assertServiceUnavailable(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "service_unavailable", domainErr.Code)
	assert.Equal(t, 30, domainErr.Details["retry_after_seconds"])
}

// TestUnitRegister_LookupFailsWithConnectivityError_ReturnsServiceUnavailable
// verifies a database outage during the email-uniqueness lookup maps to a 503,
// not the generic 500 a plain repository error would fall through to.
func TestUnitRegister_LookupFailsWithConnectivityError_ReturnsServiceUnavailable(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	users.EXPECT().GetUserByEmail(mock.Anything, testEmail).Return(nil, newConnRefusedError()).Once()
	creator := &fakeSubscriptionCreator{}

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Register(context.Background(), validRegisterRequest())

	assert.Nil(t, resp)
	assertServiceUnavailable(t, err)
}

// TestUnitRegister_CreateUserFailsWithConnectivityError_ReturnsServiceUnavailable
// verifies a database outage while persisting the new user maps to a 503.
func TestUnitRegister_CreateUserFailsWithConnectivityError_ReturnsServiceUnavailable(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	users.EXPECT().
		GetUserByEmail(mock.Anything, testEmail).
		Return(nil, domainerrors.NotFound("user", testEmail)).
		Once()
	users.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(newConnRefusedError()).Once()
	creator := &fakeSubscriptionCreator{}

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Register(context.Background(), validRegisterRequest())

	assert.Nil(t, resp)
	assertServiceUnavailable(t, err)
}

// TestUnitRegister_SubscriptionCreationFailsWithConnectivityError_ReturnsServiceUnavailable
// verifies a database outage inside subscription creation (charge succeeds,
// persistence fails) maps to a 503 rather than a generic 500.
func TestUnitRegister_SubscriptionCreationFailsWithConnectivityError_ReturnsServiceUnavailable(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	users.EXPECT().
		GetUserByEmail(mock.Anything, testEmail).
		Return(nil, domainerrors.NotFound("user", testEmail)).
		Once()
	users.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(nil).Once()
	creator := &fakeSubscriptionCreator{err: newConnRefusedError()}

	req := validRegisterRequest()
	req.PaymentMethodToken = testToken

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Register(context.Background(), req)

	assert.Nil(t, resp)
	assertServiceUnavailable(t, err)
}

// TestUnitRegister_SubscriptionFlipUpdateFailsWithConnectivityError_ReturnsServiceUnavailable
// verifies a database outage while flipping has_subscription=true after a
// successful subscription creation maps to a 503.
func TestUnitRegister_SubscriptionFlipUpdateFailsWithConnectivityError_ReturnsServiceUnavailable(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	users.EXPECT().
		GetUserByEmail(mock.Anything, testEmail).
		Return(nil, domainerrors.NotFound("user", testEmail)).
		Once()
	users.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(nil).Once()
	users.EXPECT().UpdateUser(mock.Anything, mock.Anything).Return(newConnRefusedError()).Once()

	wantSub := &subscription.Subscription{ID: "sub-1", Status: subscription.StatusActive}
	creator := &fakeSubscriptionCreator{sub: wantSub}

	req := validRegisterRequest()
	req.PaymentMethodToken = testToken

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Register(context.Background(), req)

	assert.Nil(t, resp)
	assertServiceUnavailable(t, err)
}

// --- Login ----------------------------------------------------------------

// TestUnitLogin_ValidCredentials_ReturnsUser verifies the happy path: correct
// password, no prior failures, so no counter update is needed.
func TestUnitLogin_ValidCredentials_ReturnsUser(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	user := &auth.User{ID: "u1", Email: testEmail, PasswordHash: hashFor(t)}
	users.EXPECT().GetUserByEmail(mock.Anything, testEmail).Return(user, nil).Once()
	creator := &fakeSubscriptionCreator{}

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Login(context.Background(),
		auth.LoginRequest{Email: testEmail, Password: testPassword}, "203.0.113.1", "Mozilla/5.0")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "u1", resp.User.ID)
}

// TestUnitLogin_UnknownEmail_ReturnsUnauthorizedWithoutEnumeration verifies an
// unknown email yields the same generic 401 as a bad password and records a
// failed attempt.
func TestUnitLogin_UnknownEmail_ReturnsUnauthorizedWithoutEnumeration(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	users.EXPECT().
		GetUserByEmail(mock.Anything, testEmail).
		Return(nil, domainerrors.NotFound("user", testEmail)).
		Once()
	creator := &fakeSubscriptionCreator{}

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Login(context.Background(),
		auth.LoginRequest{Email: testEmail, Password: testPassword}, "203.0.113.1", "Mozilla/5.0")

	require.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "authentication_required", domainErr.Code)
	assert.Equal(t, "Invalid credentials", domainErr.Message)
}

// TestUnitLogin_WrongPassword_ReturnsUnauthorizedAndIncrementsFailedAttempts
// verifies a bad password yields a generic 401 and increments the persisted
// failed-attempt counter.
func TestUnitLogin_WrongPassword_ReturnsUnauthorizedAndIncrementsFailedAttempts(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	user := &auth.User{
		ID: "u1", Email: testEmail, PasswordHash: hashFor(t), FailedLoginAttempts: 0,
	}
	users.EXPECT().GetUserByEmail(mock.Anything, testEmail).Return(user, nil).Once()
	users.EXPECT().
		UpdateUser(mock.Anything, mock.MatchedBy(func(u *auth.User) bool {
			return u.FailedLoginAttempts == 1
		})).
		Return(nil).
		Once()
	creator := &fakeSubscriptionCreator{}

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Login(context.Background(),
		auth.LoginRequest{Email: testEmail, Password: "WrongPassword9!"}, "203.0.113.1", "Mozilla/5.0")

	require.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "authentication_required", domainErr.Code)
	assert.Equal(t, "Invalid credentials", domainErr.Message)
}

// TestUnitLogin_RateLimited_ReturnsTooManyRequestsBeforeLookup verifies that
// once the limiter trips, login returns a 429 carrying retry_after and never
// reaches the user lookup or bcrypt.
func TestUnitLogin_RateLimited_ReturnsTooManyRequestsBeforeLookup(t *testing.T) {
	users := authmocks.NewMockUserRepository(t) // no expects: lookup must not happen
	creator := &fakeSubscriptionCreator{}

	limiter := ratelimit.New()
	for range 5 { // freeAttempts, so the next CheckRateLimit returns a delay
		limiter.RecordFailure(testEmail)
	}

	svc := auth.NewService(users, creator, limiter, testBcryptCost)
	resp, err := svc.Login(context.Background(),
		auth.LoginRequest{Email: testEmail, Password: testPassword}, "203.0.113.1", "Mozilla/5.0")

	require.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "rate_limit_exceeded", domainErr.Code)
	retryAfter, ok := domainErr.Details["retry_after_seconds"].(int)
	require.True(t, ok, "expected integer retry_after_seconds in details")
	assert.GreaterOrEqual(t, retryAfter, 1)
}

// TestUnitLogin_SuccessAfterFailures_ResetsFailedAttempts verifies a successful
// login clears a non-zero failed-attempt counter and persists it.
func TestUnitLogin_SuccessAfterFailures_ResetsFailedAttempts(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	user := &auth.User{
		ID: "u1", Email: testEmail, PasswordHash: hashFor(t), FailedLoginAttempts: 3,
	}
	users.EXPECT().GetUserByEmail(mock.Anything, testEmail).Return(user, nil).Once()
	users.EXPECT().
		UpdateUser(mock.Anything, mock.MatchedBy(func(u *auth.User) bool {
			return u.FailedLoginAttempts == 0
		})).
		Return(nil).
		Once()
	creator := &fakeSubscriptionCreator{}

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Login(context.Background(),
		auth.LoginRequest{Email: testEmail, Password: testPassword}, "203.0.113.1", "Mozilla/5.0")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "u1", resp.User.ID)
}

// TestUnitLogin_InvalidRequest_ReturnsValidationWithoutTouchingRepo verifies
// input validation runs before any repository call.
func TestUnitLogin_InvalidRequest_ReturnsValidationWithoutTouchingRepo(t *testing.T) {
	users := authmocks.NewMockUserRepository(t) // no expects
	creator := &fakeSubscriptionCreator{}

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Login(context.Background(),
		auth.LoginRequest{Email: testEmail, Password: ""}, "203.0.113.1", "Mozilla/5.0")

	require.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "validation_failed", domainErr.Code)
}

// TestUnitLogin_LookupFailsWithConnectivityError_ReturnsServiceUnavailable
// verifies a database outage during the credential lookup maps to a 503,
// not the uniform 401 an unknown email would get.
func TestUnitLogin_LookupFailsWithConnectivityError_ReturnsServiceUnavailable(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	users.EXPECT().GetUserByEmail(mock.Anything, testEmail).Return(nil, newConnRefusedError()).Once()
	creator := &fakeSubscriptionCreator{}

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Login(context.Background(),
		auth.LoginRequest{Email: testEmail, Password: testPassword}, "203.0.113.1", "Mozilla/5.0")

	assert.Nil(t, resp)
	assertServiceUnavailable(t, err)
}

// TestUnitLogin_FailedAttemptUpdateFailsWithConnectivityError_ReturnsServiceUnavailable
// verifies a database outage while persisting the incremented failed-attempt
// counter (after a wrong password) maps to a 503.
func TestUnitLogin_FailedAttemptUpdateFailsWithConnectivityError_ReturnsServiceUnavailable(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	user := &auth.User{
		ID: "u1", Email: testEmail, PasswordHash: hashFor(t), FailedLoginAttempts: 0,
	}
	users.EXPECT().GetUserByEmail(mock.Anything, testEmail).Return(user, nil).Once()
	users.EXPECT().UpdateUser(mock.Anything, mock.Anything).Return(newConnRefusedError()).Once()
	creator := &fakeSubscriptionCreator{}

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Login(context.Background(),
		auth.LoginRequest{Email: testEmail, Password: "WrongPassword9!"}, "203.0.113.1", "Mozilla/5.0")

	assert.Nil(t, resp)
	assertServiceUnavailable(t, err)
}

// TestUnitLogin_ResetFailedAttemptsUpdateFailsWithConnectivityError_ReturnsServiceUnavailable
// verifies a database outage while clearing a previously non-zero failed-
// attempt counter on a successful login maps to a 503.
func TestUnitLogin_ResetFailedAttemptsUpdateFailsWithConnectivityError_ReturnsServiceUnavailable(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	user := &auth.User{
		ID: "u1", Email: testEmail, PasswordHash: hashFor(t), FailedLoginAttempts: 3,
	}
	users.EXPECT().GetUserByEmail(mock.Anything, testEmail).Return(user, nil).Once()
	users.EXPECT().UpdateUser(mock.Anything, mock.Anything).Return(newConnRefusedError()).Once()
	creator := &fakeSubscriptionCreator{}

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.Login(context.Background(),
		auth.LoginRequest{Email: testEmail, Password: testPassword}, "203.0.113.1", "Mozilla/5.0")

	assert.Nil(t, resp)
	assertServiceUnavailable(t, err)
}

// --- CurrentUser ----------------------------------------------------------

// TestUnitCurrentUser_ExistingUser_ReturnsUser verifies the happy path: the id
// carried by a valid access token resolves to the account it was minted for.
func TestUnitCurrentUser_ExistingUser_ReturnsUser(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	user := &auth.User{ID: "u1", Email: testEmail, FullName: testFullName, HasSubscription: true}
	users.EXPECT().GetUserByID(mock.Anything, "u1").Return(user, nil).Once()
	creator := &fakeSubscriptionCreator{}

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.CurrentUser(context.Background(), "u1")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "u1", resp.User.ID)
	assert.Equal(t, testEmail, resp.User.Email)
	assert.True(t, resp.User.HasSubscription,
		"has_subscription must come from the row, not the token's possibly-stale claim")
}

// TestUnitCurrentUser_UnknownUser_PropagatesNotFound verifies the repository's
// NotFound reaches the caller unchanged. Translating it to the uniform 401 is
// the handler's job: the service has no notion of a session.
func TestUnitCurrentUser_UnknownUser_PropagatesNotFound(t *testing.T) {
	users := authmocks.NewMockUserRepository(t)
	users.EXPECT().
		GetUserByID(mock.Anything, "ghost").
		Return(nil, domainerrors.NotFound("user", "ghost")).
		Once()
	creator := &fakeSubscriptionCreator{}

	svc := auth.NewService(users, creator, ratelimit.New(), testBcryptCost)
	resp, err := svc.CurrentUser(context.Background(), "ghost")

	require.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}
