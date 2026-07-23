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
// production wiring, with RequestID middleware so correlation ids populate.
func newAuthRouter(t *testing.T, h *auth.Handler) http.Handler {
	t.Helper()
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Route("/api/v1", func(r chi.Router) {
		h.RegisterRoutes(r)
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

	h := auth.NewHandler(svc, issuer, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{Domain: "localhost", Secure: true})
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

	h := auth.NewHandler(svc, issuer, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
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
	h := auth.NewHandler(svc, &fakeTokenIssuer{}, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
	rec := doJSON(t, newAuthRouter(t, h), http.MethodPost, "/api/v1/auth/register", `{not json`)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "invalid_request", decodeBody(t, rec)["error"])
	svc.AssertNotCalled(t, "Register")
}

func TestUnitHandleRegister_DuplicateEmail_Returns409(t *testing.T) {
	svc := authmocks.NewMockAccountService(t)
	svc.EXPECT().Register(mock.Anything, mock.Anything).
		Return(nil, domainerrors.Conflict("Email already registered")).Once()

	h := auth.NewHandler(svc, &fakeTokenIssuer{}, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
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

	h := auth.NewHandler(svc, &fakeTokenIssuer{}, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
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

	h := auth.NewHandler(svc, issuer, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
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

	h := auth.NewHandler(svc, issuer, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{Secure: true})
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
	h := auth.NewHandler(svc, &fakeTokenIssuer{}, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
	rec := doJSON(t, newAuthRouter(t, h), http.MethodPost, "/api/v1/auth/login", `{bad`)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	svc.AssertNotCalled(t, "Login")
}

func TestUnitHandleLogin_InvalidCredentials_Returns401NoCookies(t *testing.T) {
	svc := authmocks.NewMockAccountService(t)
	svc.EXPECT().Login(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, domainerrors.Unauthorized("Invalid credentials")).Once()
	issuer := &fakeTokenIssuer{}

	h := auth.NewHandler(svc, issuer, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
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

	h := auth.NewHandler(svc, &fakeTokenIssuer{}, authmocks.NewMockRefreshTokenRepository(t), auth.CookieConfig{})
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

	h := auth.NewHandler(authmocks.NewMockAccountService(t), &fakeTokenIssuer{}, repo, auth.CookieConfig{})
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

	h := auth.NewHandler(authmocks.NewMockAccountService(t), &fakeTokenIssuer{}, repo, auth.CookieConfig{})
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

	h := auth.NewHandler(authmocks.NewMockAccountService(t), &fakeTokenIssuer{}, repo, auth.CookieConfig{})
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

	h := auth.NewHandler(authmocks.NewMockAccountService(t), &fakeTokenIssuer{}, repo, auth.CookieConfig{})
	rec := postLogout(t, newAuthRouter(t, h), &http.Cookie{Name: "refresh_token", Value: raw})

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Nil(t, cookieByName(rec, "access_token"), "cookies must not be cleared when revocation fails")
}
