//go:build test

package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/internal/auth"
	authjwt "github.com/JosemaPereira/TrAIveler/backend/internal/auth/jwt"
	authmocks "github.com/JosemaPereira/TrAIveler/backend/internal/auth/mocks"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
	"github.com/JosemaPereira/TrAIveler/backend/internal/subscription"
)

const (
	testUserID   = "11111111-1111-1111-1111-111111111111"
	handlerEmail = "ada@example.com"
)

// fakeTokenIssuer is a hand fake for the one-method TokenIssuer port (preferred
// over a generated mock per docs/mock-standards.md); it records args and returns
// a canned pair.
type fakeTokenIssuer struct {
	pair      auth.TokenPair
	err       error
	called    bool
	gotUserID string
	gotHasSub bool
}

func (f *fakeTokenIssuer) IssueTokens(_ context.Context, userID string, hasSubscription bool) (auth.TokenPair, error) {
	f.called = true
	f.gotUserID = userID
	f.gotHasSub = hasSubscription
	return f.pair, f.err
}

// fakeTokenRefresher is a hand fake for the one-method TokenRefresher port,
// matching the fakeTokenIssuer convention above.
type fakeTokenRefresher struct {
	pair     auth.TokenPair
	err      error
	called   bool
	gotToken string
}

func (f *fakeTokenRefresher) RefreshToken(_ context.Context, rawRefreshToken string) (auth.TokenPair, error) {
	f.called = true
	f.gotToken = rawRefreshToken
	return f.pair, f.err
}

func validPair() auth.TokenPair {
	return auth.TokenPair{
		AccessToken:      "access.jwt.value",
		RefreshToken:     "refresh-opaque-value",
		AccessExpiresAt:  time.Now().Add(24 * time.Hour),
		RefreshExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
}

func testUser(hasSubscription bool) auth.User {
	return auth.User{ID: testUserID, Email: handlerEmail, FullName: "Ada Lovelace", HasSubscription: hasSubscription}
}

// newAuthRouter mounts the handler under /api/v1 on a real chi.Router, matching
// production wiring (cmd/api/routes.go): one /api/v1 subtree holding a public
// group and an authenticated group, with RequestID middleware so correlation ids
// populate. No Authenticate gate is mounted here — these are transport-level
// unit tests; the gate is covered by cmd/api's wiring tests (008-T209).
func newAuthRouter(t *testing.T, h *auth.Handler) http.Handler {
	t.Helper()
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Route("/api/v1", func(r chi.Router) {
		r.Group(h.RegisterPublicRoutes)
		r.Group(h.RegisterProtectedRoutes)
	})
	return r
}

func doJSON(t *testing.T, router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func cookieByName(recorder *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range recorder.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body
}

// --- Register -------------------------------------------------------------

func TestUnitHandleRegister_FreeUserNoToken_Returns201WithSessionCookies(t *testing.T) {
	svc := authmocks.NewMockAccountService(t)
	svc.EXPECT().
		Register(mock.Anything, mock.MatchedBy(func(req auth.RegisterRequest) bool {
			return req.Email == handlerEmail && req.PaymentMethodToken == ""
		})).
		Return(&auth.RegisterResponse{User: testUser(false)}, nil).
		Once()
	issuer := &fakeTokenIssuer{pair: validPair()}

	h := auth.NewHandler(svc, issuer, &fakeTokenRefresher{}, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{Domain: "localhost", Secure: true})
	rec := doJSON(t, newAuthRouter(t, h), http.MethodPost, "/api/v1/auth/register",
		`{"email":"ada@example.com","password":"CorrectHorse1!","full_name":"Ada Lovelace"}`)

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.True(t, issuer.called)
	assert.False(t, issuer.gotHasSub, "free user token must carry has_subscription=false")

	access := cookieByName(rec, "access_token")
	require.NotNil(t, access)
	assert.Equal(t, "access.jwt.value", access.Value)
	assert.True(t, access.HttpOnly)
	assert.True(t, access.Secure)
	assert.Equal(t, http.SameSiteStrictMode, access.SameSite)
	assert.Equal(t, "/", access.Path)
	refresh := cookieByName(rec, "refresh_token")
	require.NotNil(t, refresh)
	assert.Equal(t, "refresh-opaque-value", refresh.Value)
	assert.True(t, refresh.HttpOnly)

	body := decodeBody(t, rec)
	user := body["user"].(map[string]any)
	assert.Equal(t, handlerEmail, user["email"])
	assert.Equal(t, false, user["has_subscription"])
	_, hasSubField := body["subscription"]
	assert.False(t, hasSubField, "no subscription object without a payment token")
}

func TestUnitHandleRegister_PaidUserWithToken_Returns201WithSubscribedToken(t *testing.T) {
	svc := authmocks.NewMockAccountService(t)
	svc.EXPECT().
		Register(mock.Anything, mock.MatchedBy(func(req auth.RegisterRequest) bool {
			return req.PaymentMethodToken == "tok_visa"
		})).
		Return(&auth.RegisterResponse{
			User:         testUser(true),
			Subscription: &subscription.Subscription{ID: "sub-1"},
		}, nil).
		Once()
	issuer := &fakeTokenIssuer{pair: validPair()}

	h := auth.NewHandler(svc, issuer, &fakeTokenRefresher{}, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
	rec := doJSON(t, newAuthRouter(t, h), http.MethodPost, "/api/v1/auth/register",
		`{"email":"ada@example.com","password":"CorrectHorse1!","full_name":"Ada Lovelace","payment_method_token":"tok_visa"}`)

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, testUserID, issuer.gotUserID)
	assert.True(t, issuer.gotHasSub, "paid user token must carry has_subscription=true")
	_, hasSubField := decodeBody(t, rec)["subscription"]
	assert.True(t, hasSubField)
}

func TestUnitHandleRegister_MalformedJSON_Returns400InvalidRequest(t *testing.T) {
	svc := authmocks.NewMockAccountService(t)
	h := auth.NewHandler(svc, &fakeTokenIssuer{}, &fakeTokenRefresher{}, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
	rec := doJSON(t, newAuthRouter(t, h), http.MethodPost, "/api/v1/auth/register", `{not json`)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "invalid_request", decodeBody(t, rec)["error"])
	svc.AssertNotCalled(t, "Register")
}

func TestUnitHandleRegister_DuplicateEmail_Returns409(t *testing.T) {
	svc := authmocks.NewMockAccountService(t)
	svc.EXPECT().Register(mock.Anything, mock.Anything).
		Return(nil, domainerrors.Conflict("Email already registered")).Once()

	h := auth.NewHandler(svc, &fakeTokenIssuer{}, &fakeTokenRefresher{}, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
	rec := doJSON(t, newAuthRouter(t, h), http.MethodPost, "/api/v1/auth/register",
		`{"email":"ada@example.com","password":"CorrectHorse1!","full_name":"Ada Lovelace"}`)

	require.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, "conflict", decodeBody(t, rec)["error"])
	assert.Nil(t, cookieByName(rec, "access_token"))
}

func TestUnitHandleRegister_ValidationError_Returns422(t *testing.T) {
	svc := authmocks.NewMockAccountService(t)
	svc.EXPECT().Register(mock.Anything, mock.Anything).
		Return(nil, domainerrors.Validation("One or more fields failed validation",
			domainerrors.ValidationError{Field: "password", Error: "too weak"})).Once()

	h := auth.NewHandler(svc, &fakeTokenIssuer{}, &fakeTokenRefresher{}, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
	rec := doJSON(t, newAuthRouter(t, h), http.MethodPost, "/api/v1/auth/register",
		`{"email":"bad","password":"weak","full_name":"Ada"}`)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Equal(t, "validation_failed", decodeBody(t, rec)["error"])
}

func TestUnitHandleRegister_TokenIssuanceFails_Returns500(t *testing.T) {
	svc := authmocks.NewMockAccountService(t)
	svc.EXPECT().Register(mock.Anything, mock.Anything).
		Return(&auth.RegisterResponse{User: testUser(false)}, nil).Once()
	issuer := &fakeTokenIssuer{err: errors.New("signing key unavailable")}

	h := auth.NewHandler(svc, issuer, &fakeTokenRefresher{}, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
	rec := doJSON(t, newAuthRouter(t, h), http.MethodPost, "/api/v1/auth/register",
		`{"email":"ada@example.com","password":"CorrectHorse1!","full_name":"Ada Lovelace"}`)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Nil(t, cookieByName(rec, "access_token"))
}

// --- Login ----------------------------------------------------------------

func TestUnitHandleLogin_ValidCredentials_Returns200WithSessionCookies(t *testing.T) {
	svc := authmocks.NewMockAccountService(t)
	svc.EXPECT().
		Login(mock.Anything, mock.MatchedBy(func(req auth.LoginRequest) bool {
			return req.Email == handlerEmail
		}), mock.Anything, mock.Anything).
		Return(&auth.LoginResponse{User: testUser(true)}, nil).
		Once()
	issuer := &fakeTokenIssuer{pair: validPair()}

	h := auth.NewHandler(svc, issuer, &fakeTokenRefresher{}, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{Secure: true})
	rec := doJSON(t, newAuthRouter(t, h), http.MethodPost, "/api/v1/auth/login",
		`{"email":"ada@example.com","password":"CorrectHorse1!"}`)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, issuer.gotHasSub)
	require.NotNil(t, cookieByName(rec, "access_token"))
	require.NotNil(t, cookieByName(rec, "refresh_token"))
	user := decodeBody(t, rec)["user"].(map[string]any)
	assert.Equal(t, handlerEmail, user["email"])
}

func TestUnitHandleLogin_MalformedJSON_Returns400(t *testing.T) {
	svc := authmocks.NewMockAccountService(t)
	h := auth.NewHandler(svc, &fakeTokenIssuer{}, &fakeTokenRefresher{}, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
	rec := doJSON(t, newAuthRouter(t, h), http.MethodPost, "/api/v1/auth/login", `{bad`)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	svc.AssertNotCalled(t, "Login")
}

func TestUnitHandleLogin_InvalidCredentials_Returns401NoCookies(t *testing.T) {
	svc := authmocks.NewMockAccountService(t)
	svc.EXPECT().Login(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, domainerrors.Unauthorized("Invalid credentials")).Once()
	issuer := &fakeTokenIssuer{}

	h := auth.NewHandler(svc, issuer, &fakeTokenRefresher{}, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
	rec := doJSON(t, newAuthRouter(t, h), http.MethodPost, "/api/v1/auth/login",
		`{"email":"ada@example.com","password":"wrong"}`)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "authentication_required", decodeBody(t, rec)["error"])
	assert.False(t, issuer.called)
	assert.Nil(t, cookieByName(rec, "access_token"))
}

func TestUnitHandleLogin_RateLimited_Returns429WithRetryAfter(t *testing.T) {
	svc := authmocks.NewMockAccountService(t)
	svc.EXPECT().Login(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, domainerrors.RateLimited(8)).Once()

	h := auth.NewHandler(svc, &fakeTokenIssuer{}, &fakeTokenRefresher{}, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
	rec := doJSON(t, newAuthRouter(t, h), http.MethodPost, "/api/v1/auth/login",
		`{"email":"ada@example.com","password":"wrong"}`)

	require.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Equal(t, "8", rec.Header().Get("Retry-After"))
}

// --- Logout ---------------------------------------------------------------

func postLogout(t *testing.T, router http.Handler, refreshCookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	if refreshCookie != nil {
		req.AddCookie(refreshCookie)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestUnitHandleLogout_WithValidRefreshCookie_RevokesAndClearsCookies(t *testing.T) {
	raw := "presented-refresh-token"
	repo := authmocks.NewMockRefreshTokenRepository(t)
	repo.EXPECT().
		GetRefreshTokenByHash(mock.Anything, authjwt.HashRefreshToken(raw)).
		Return(&auth.RefreshToken{ID: "rt-1"}, nil).
		Once()
	repo.EXPECT().RevokeRefreshToken(mock.Anything, "rt-1").Return(nil).Once()

	h := auth.NewHandler(authmocks.NewMockAccountService(t), &fakeTokenIssuer{}, &fakeTokenRefresher{}, repo, auth.CookieConfig{})
	rec := postLogout(t, newAuthRouter(t, h), &http.Cookie{Name: "refresh_token", Value: raw})

	require.Equal(t, http.StatusNoContent, rec.Code)
	access := cookieByName(rec, "access_token")
	require.NotNil(t, access)
	assert.True(t, access.MaxAge < 0, "access cookie must be expired")
	assert.Empty(t, access.Value)
	refresh := cookieByName(rec, "refresh_token")
	require.NotNil(t, refresh)
	assert.True(t, refresh.MaxAge < 0)
}

func TestUnitHandleLogout_NoRefreshCookie_ClearsCookiesWithoutRevoking(t *testing.T) {
	repo := authmocks.NewMockRefreshTokenRepository(t)

	h := auth.NewHandler(authmocks.NewMockAccountService(t), &fakeTokenIssuer{}, &fakeTokenRefresher{}, repo, auth.CookieConfig{})
	rec := postLogout(t, newAuthRouter(t, h), nil)

	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.NotNil(t, cookieByName(rec, "access_token"))
	repo.AssertNotCalled(t, "GetRefreshTokenByHash")
	repo.AssertNotCalled(t, "RevokeRefreshToken")
}

func TestUnitHandleLogout_UnknownRefreshToken_IsIdempotent204(t *testing.T) {
	raw := "stale-token"
	repo := authmocks.NewMockRefreshTokenRepository(t)
	repo.EXPECT().
		GetRefreshTokenByHash(mock.Anything, authjwt.HashRefreshToken(raw)).
		Return(nil, domainerrors.NotFound("refresh token", "hash")).
		Once()

	h := auth.NewHandler(authmocks.NewMockAccountService(t), &fakeTokenIssuer{}, &fakeTokenRefresher{}, repo, auth.CookieConfig{})
	rec := postLogout(t, newAuthRouter(t, h), &http.Cookie{Name: "refresh_token", Value: raw})

	require.Equal(t, http.StatusNoContent, rec.Code)
	repo.AssertNotCalled(t, "RevokeRefreshToken")
}

func TestUnitHandleLogout_RevokeStorageError_Returns500(t *testing.T) {
	raw := "presented-refresh-token"
	repo := authmocks.NewMockRefreshTokenRepository(t)
	repo.EXPECT().GetRefreshTokenByHash(mock.Anything, mock.Anything).
		Return(&auth.RefreshToken{ID: "rt-1"}, nil).Once()
	repo.EXPECT().RevokeRefreshToken(mock.Anything, "rt-1").
		Return(errors.New("db down")).Once()

	h := auth.NewHandler(authmocks.NewMockAccountService(t), &fakeTokenIssuer{}, &fakeTokenRefresher{}, repo, auth.CookieConfig{})
	rec := postLogout(t, newAuthRouter(t, h), &http.Cookie{Name: "refresh_token", Value: raw})

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Nil(t, cookieByName(rec, "access_token"), "cookies must not be cleared when revocation fails")
}

// --- Refresh --------------------------------------------------------------

func postRefresh(t *testing.T, router http.Handler, refreshCookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	if refreshCookie != nil {
		req.AddCookie(refreshCookie)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func newRefreshHandler(t *testing.T, refresher auth.TokenRefresher) *auth.Handler {
	t.Helper()
	return auth.NewHandler(
		authmocks.NewMockAccountService(t),
		&fakeTokenIssuer{},
		refresher,
		authmocks.NewMockRefreshTokenRepository(t),
		auth.CookieConfig{Domain: "localhost", Secure: true},
	)
}

func TestUnitHandleRefresh_ValidRefreshCookie_Returns200AndRotatesBothCookies(t *testing.T) {
	// Arrange
	pair := validPair()
	pair.UserID = testUserID
	refresher := &fakeTokenRefresher{pair: pair}
	h := newRefreshHandler(t, refresher)

	// Act
	rec := postRefresh(t, newAuthRouter(t, h), &http.Cookie{Name: "refresh_token", Value: "presented-refresh-token"})

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, refresher.called, "the handler must delegate to the token refresher")
	assert.Equal(t, "presented-refresh-token", refresher.gotToken,
		"the raw cookie value must be forwarded to the refresher")
	assert.Equal(t, "Token refreshed successfully", decodeBody(t, rec)["message"])

	access := cookieByName(rec, "access_token")
	require.NotNil(t, access)
	assert.Equal(t, pair.AccessToken, access.Value)
	assert.True(t, access.HttpOnly)
	assert.True(t, access.Secure)
	assert.Equal(t, http.SameSiteStrictMode, access.SameSite)

	refresh := cookieByName(rec, "refresh_token")
	require.NotNil(t, refresh)
	assert.Equal(t, pair.RefreshToken, refresh.Value,
		"the rotated refresh token must replace the presented one")
}

func TestUnitHandleRefresh_MissingOrEmptyCookie_Returns401WithoutCallingTheRefresher(t *testing.T) {
	testCases := []struct {
		name   string
		cookie *http.Cookie
	}{
		{name: "when no refresh_token cookie is presented", cookie: nil},
		{name: "when the refresh_token cookie is empty", cookie: &http.Cookie{Name: "refresh_token", Value: ""}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			refresher := &fakeTokenRefresher{}
			h := newRefreshHandler(t, refresher)

			// Act
			rec := postRefresh(t, newAuthRouter(t, h), testCase.cookie)

			// Assert
			require.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Equal(t, "authentication_required", decodeBody(t, rec)["error"],
				"the uniform 401 envelope is the snake_case catalog code, not api.md's stale INVALID_REFRESH_TOKEN")
			assert.False(t, refresher.called, "a missing token must short-circuit before any store lookup")
			assert.Nil(t, cookieByName(rec, "access_token"), "no session cookie may be set on a failed refresh")
		})
	}
}

func TestUnitHandleRefresh_RefresherRejectsToken_Returns401NoCookies(t *testing.T) {
	// Arrange
	// The fake returns exactly what the real jwt.Refresher returns for any bad
	// token, so this test breaks if that error stops being the uniform one.
	refresher := &fakeTokenRefresher{err: domainerrors.Unauthorized(authjwt.RefreshFailureMessage)}
	h := newRefreshHandler(t, refresher)

	// Act
	rec := postRefresh(t, newAuthRouter(t, h), &http.Cookie{Name: "refresh_token", Value: "revoked-token"})

	// Assert
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "authentication_required", decodeBody(t, rec)["error"])
	assert.Nil(t, cookieByName(rec, "access_token"))
	assert.Nil(t, cookieByName(rec, "refresh_token"))
}

// A client must not be able to tell an absent refresh cookie from an invalid,
// expired, or revoked one: the two paths are produced by different code (the
// handler short-circuits before ever reaching the Refresher), so the identical
// response is an invariant that has to be asserted, not assumed.
func TestUnitHandleRefresh_MissingCookieAndRejectedToken_AreIndistinguishable(t *testing.T) {
	// Arrange
	missingCookie := postRefresh(t, newAuthRouter(t, newRefreshHandler(t, &fakeTokenRefresher{})), nil)

	rejected := postRefresh(t,
		newAuthRouter(t, newRefreshHandler(t,
			&fakeTokenRefresher{err: domainerrors.Unauthorized(authjwt.RefreshFailureMessage)})),
		&http.Cookie{Name: "refresh_token", Value: "revoked-token"})

	// Assert
	require.Equal(t, missingCookie.Code, rejected.Code)

	missingBody, rejectedBody := decodeBody(t, missingCookie), decodeBody(t, rejected)
	assert.Equal(t, missingBody["error"], rejectedBody["error"])
	assert.Equal(t, missingBody["message"], rejectedBody["message"],
		"a missing refresh cookie must not be distinguishable from a rejected token by its message")
	assert.Equal(t, authjwt.RefreshFailureMessage, missingBody["message"],
		"both paths must emit the message contracts/api.md specifies for this endpoint")
}

func TestUnitHandleRefresh_StorageFailure_Returns500NoCookies(t *testing.T) {
	// Arrange
	refresher := &fakeTokenRefresher{err: errors.New("db down")}
	h := newRefreshHandler(t, refresher)

	// Act
	rec := postRefresh(t, newAuthRouter(t, h), &http.Cookie{Name: "refresh_token", Value: "some-token"})

	// Assert
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Nil(t, cookieByName(rec, "access_token"))
}

// --- Route split (008-T208) -----------------------------------------------

func TestUnitRegisterRoutes_PublicAndProtectedGroupsCoexistWithoutPanicking(t *testing.T) {
	// Arrange
	h := newRefreshHandler(t, &fakeTokenRefresher{pair: validPair()})

	// Act / Assert: chi panics if the same pattern is mounted twice on one
	// routing tree, so registering full paths from two sibling groups (rather
	// than two r.Route("/auth", ...) subtrees) must be panic-free.
	var router http.Handler
	require.NotPanics(t, func() { router = newAuthRouter(t, h) })

	testCases := []struct {
		name     string
		path     string
		wantCode int
	}{
		{name: "when the public register route is called with a malformed body", path: "/api/v1/auth/register", wantCode: http.StatusBadRequest},
		{name: "when the public login route is called with a malformed body", path: "/api/v1/auth/login", wantCode: http.StatusBadRequest},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			rec := doJSON(t, router, http.MethodPost, testCase.path, `{bad`)
			assert.Equal(t, testCase.wantCode, rec.Code)
		})
	}
}

// --- Current user ---------------------------------------------------------

// fakeClaimsValidator is a hand fake for middleware.TokenValidator (one method,
// per docs/mock-standards.md), letting these tests drive the real gate without
// RSA keys.
type fakeClaimsValidator struct {
	userID string
}

func (f *fakeClaimsValidator) ValidateToken(_ context.Context, _ string) (middleware.AuthClaims, error) {
	return middleware.AuthClaims{UserID: f.userID}, nil
}

// newGatedAuthRouter mounts the protected routes behind the real
// middleware.Authenticate. Driving the actual gate is both closer to production
// wiring than hand-injecting a context value and cheaper: the alternative would
// have been exporting a context setter from internal/middleware purely for
// tests, which is the kind of production-code test seam this repo already
// regrets elsewhere.
func newGatedAuthRouter(t *testing.T, h *auth.Handler, userID string) http.Handler {
	t.Helper()
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(&fakeClaimsValidator{userID: userID}))
			h.RegisterProtectedRoutes(r)
		})
	})
	return r
}

// doAuthenticatedGet issues a GET carrying the access_token cookie the gate
// requires. Its value is irrelevant — fakeClaimsValidator accepts anything —
// but it must be present, since Authenticate rejects an absent cookie before
// ever calling the validator.
func doAuthenticatedGet(t *testing.T, router http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "any-non-empty-token"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestUnitHandleCurrentUser_Authenticated_Returns200WithUser(t *testing.T) {
	svc := authmocks.NewMockAccountService(t)
	svc.EXPECT().
		CurrentUser(mock.Anything, testUserID).
		Return(&auth.CurrentUserResponse{User: testUser(true)}, nil).
		Once()

	h := auth.NewHandler(svc, &fakeTokenIssuer{}, &fakeTokenRefresher{}, nil, auth.CookieConfig{})
	rec := doAuthenticatedGet(t, newGatedAuthRouter(t, h, testUserID), "/api/v1/auth/me")

	require.Equal(t, http.StatusOK, rec.Code)
	body := decodeBody(t, rec)
	user, ok := body["user"].(map[string]any)
	require.True(t, ok, "response must wrap the account in a \"user\" envelope")
	assert.Equal(t, testUserID, user["id"])
	assert.Equal(t, handlerEmail, user["email"])
	assert.Equal(t, true, user["has_subscription"])
	assert.NotContains(t, user, "role", "role is never serialized (models.go json:\"-\")")
}

// A GET must not mint or rotate anything.
func TestUnitHandleCurrentUser_Authenticated_SetsNoCookies(t *testing.T) {
	svc := authmocks.NewMockAccountService(t)
	svc.EXPECT().
		CurrentUser(mock.Anything, testUserID).
		Return(&auth.CurrentUserResponse{User: testUser(false)}, nil).
		Once()

	issuer := &fakeTokenIssuer{}
	h := auth.NewHandler(svc, issuer, &fakeTokenRefresher{}, nil, auth.CookieConfig{})
	rec := doAuthenticatedGet(t, newGatedAuthRouter(t, h, testUserID), "/api/v1/auth/me")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Result().Cookies(), "a read-only probe must not set cookies")
	assert.False(t, issuer.called, "a read-only probe must not mint tokens")
}

// A token whose subject no longer exists (deleted account) is a dead session,
// not a missing resource: the client must be told to log in again, not shown a
// 404 for its own identity.
func TestUnitHandleCurrentUser_DeletedAccount_Returns401NotNotFound(t *testing.T) {
	svc := authmocks.NewMockAccountService(t)
	svc.EXPECT().
		CurrentUser(mock.Anything, testUserID).
		Return(nil, domainerrors.NotFound("user", testUserID)).
		Once()

	h := auth.NewHandler(svc, &fakeTokenIssuer{}, &fakeTokenRefresher{}, nil, auth.CookieConfig{})
	rec := doAuthenticatedGet(t, newGatedAuthRouter(t, h, testUserID), "/api/v1/auth/me")

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "authentication_required", decodeBody(t, rec)["error"])
}

// Defense in depth: the route only mounts behind Authenticate, so a missing id
// means the gate was misconfigured. Fail closed rather than panic or leak.
func TestUnitHandleCurrentUser_NoUserInContext_Returns401(t *testing.T) {
	svc := authmocks.NewMockAccountService(t) // no expects: must not be reached

	h := auth.NewHandler(svc, &fakeTokenIssuer{}, &fakeTokenRefresher{}, nil, auth.CookieConfig{})
	rec := doJSON(t, newAuthRouter(t, h), http.MethodGet, "/api/v1/auth/me", "")

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "authentication_required", decodeBody(t, rec)["error"])
}
