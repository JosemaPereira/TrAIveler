//go:build test

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/config"
	"github.com/JosemaPereira/TrAIveler/backend/internal/database"
	dbmocks "github.com/JosemaPereira/TrAIveler/backend/internal/database/mocks"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
)

const (
	gatedUserID       = "33333333-3333-3333-3333-333333333333"
	probePath         = "/api/v1/auth-probe"
	anyAccessTokenVal = "any-access-token-value"
)

// fakeTokenValidator is a hand fake for middleware.TokenValidator (an internal
// port — hand-faked rather than mockery-generated, per docs/mock-standards.md).
// It keeps these wiring tests free of real RSA keys, and its call counter is what
// distinguishes a gated route from an ungated one: Authenticate only consults a
// validator on routes it actually guards.
type fakeTokenValidator struct {
	claims    middleware.AuthClaims
	err       error
	callCount int
}

func (f *fakeTokenValidator) ValidateToken(_ context.Context, _ string) (middleware.AuthClaims, error) {
	f.callCount++
	return f.claims, f.err
}

// withTokenValidator injects a stand-in for the JWT-backed validator behind the
// Authenticate gate. It is the 008-T209 test seam: NewHTTPServer otherwise builds
// its validator internally from real signing keys.
func withTokenValidator(validator middleware.TokenValidator) serverOption {
	return func(s *HTTPServer) { s.tokenValidator = validator }
}

// withProtectedRoutes mounts extra routes inside the authenticated /api/v1
// group, so a test can observe what a handler behind the gate actually sees.
// It drives HTTPServer.extraProtectedRoutes and shares that field's lifetime:
// both go away once a real subscription-gated route exists to assert
// 008-T209's context propagation against.
func withProtectedRoutes(register func(chi.Router)) serverOption {
	return func(s *HTTPServer) { s.extraProtectedRoutes = register }
}

// newGatedServer builds a server whose gate accepts every presented cookie and
// reports the given claims, optionally mounting extra probe routes inside the
// authenticated group. Uses the default (non-production) test config.
func newGatedServer(t *testing.T, db database.Client, extra ...serverOption) (*HTTPServer, *fakeTokenValidator) {
	t.Helper()
	return newGatedServerWithConfig(t, db, testConfig(t), extra...)
}

// newGatedServerWithConfig is newGatedServer parametrized by cfg, so a test can
// exercise environment-dependent behavior — e.g. /swagger/* only being gated in
// production (issue #192 Task C) — without duplicating the fake-validator wiring.
func newGatedServerWithConfig(
	t *testing.T, db database.Client, cfg *config.Config, extra ...serverOption,
) (*HTTPServer, *fakeTokenValidator) {
	t.Helper()

	validator := &fakeTokenValidator{
		claims: middleware.AuthClaims{UserID: gatedUserID, HasSubscription: true},
	}
	opts := append([]serverOption{withTokenValidator(validator)}, extra...)

	srv, err := NewHTTPServer(db, cfg, testLogger(t), opts...)
	require.NoError(t, err)

	return srv, validator
}

// productionTestConfig is testConfig with Environment set to production, for
// tests proving production-only behavior such as /swagger/* gating (issue
// #192 Task C).
func productionTestConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := testConfig(t)
	cfg.Environment = "production"
	return cfg
}

// doRequest issues one request against the server's real route table, optionally
// presenting an access_token cookie.
func doRequest(t *testing.T, srv *HTTPServer, method, path string, withCookie bool) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, http.NoBody)
	if withCookie {
		req.AddCookie(&http.Cookie{Name: "access_token", Value: anyAccessTokenVal})
	}
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)
	return rec
}

// requestWithCookie issues a request that *always* presents an access_token
// cookie (the body is what's optional — pass "" to omit it). Always sending the
// cookie is the point: a route's gated/ungated status is then observable purely
// from whether the fake validator was consulted.
func requestWithCookie(t *testing.T, srv *HTTPServer, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader = http.NoBody
	if body != "" {
		reader = bytes.NewBufferString(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: anyAccessTokenVal})
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)
	return rec
}

func decodeErrorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	code, _ := body["error"].(string)
	return code
}

func TestHTTPServer_Routes_RegisteringPublicAndProtectedGroupsDoesNotPanic(t *testing.T) {
	// chi panics when the same pattern is routed twice on one tree, which is the
	// hazard the public/protected split has to avoid.
	assert.NotPanics(t, func() {
		_, _ = newGatedServer(t, dbmocks.NewMockClient(t))
	})
}

func TestHTTPServer_ProtectedRoutes_NoAccessTokenCookie_Returns401AuthenticationRequired(t *testing.T) {
	// /swagger/* is deliberately not in this table: unlike the routes below, it
	// is gated only in production (issue #192 Task C) — see
	// TestHTTPServer_SwaggerRoutes_ProductionConfig_NoAccessTokenCookie_Returns401AuthenticationRequired
	// and TestHTTPServer_SwaggerRoutes_NonProductionConfig_ServesWithoutAuthentication.
	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "when logout is requested", method: http.MethodPost, path: "/api/v1/auth/logout"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			srv, validator := newGatedServer(t, dbmocks.NewMockClient(t))

			// Act
			rec := doRequest(t, srv, testCase.method, testCase.path, false)

			// Assert
			require.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Equal(t, "authentication_required", decodeErrorCode(t, rec))
			assert.Zero(t, validator.callCount,
				"a missing cookie must be rejected before any token validation")
		})
	}
}

func TestHTTPServer_ProtectedRoutes_InvalidAccessToken_Returns401AuthenticationRequired(t *testing.T) {
	// Exercised against /api/v1/auth/logout rather than /swagger/* because the
	// latter is only conditionally gated (issue #192 Task C); logout is gated
	// in every environment, which is what this test's own name promises.
	// Arrange
	srv, validator := newGatedServer(t, dbmocks.NewMockClient(t))
	validator.err = errors.New("invalid or expired token")

	// Act
	rec := doRequest(t, srv, http.MethodPost, "/api/v1/auth/logout", true)

	// Assert
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "authentication_required", decodeErrorCode(t, rec))
	assert.Equal(t, 1, validator.callCount, "a presented cookie must be validated")
}

func TestHTTPServer_ProtectedRoutes_ExpiredAccessToken_Returns401TokenExpired(t *testing.T) {
	// A validly-signed-but-expired token gets its own code end-to-end through the
	// real route table, so the frontend can tell "refresh me" (008-T150) apart from
	// every other 401, which stays the uniform authentication_required. Exercised
	// against /api/v1/auth/logout for the same reason as the InvalidAccessToken
	// test above: /swagger/* is only conditionally gated (issue #192 Task C).
	// Arrange
	srv, validator := newGatedServer(t, dbmocks.NewMockClient(t))
	validator.err = middleware.ErrTokenExpired

	// Act
	rec := doRequest(t, srv, http.MethodPost, "/api/v1/auth/logout", true)

	// Assert
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "token_expired", decodeErrorCode(t, rec))
	assert.Equal(t, 1, validator.callCount, "a presented cookie must be validated")
}

func TestHTTPServer_PublicRoutes_ReachableWithoutAuthentication(t *testing.T) {
	testCases := []struct {
		name     string
		method   string
		path     string
		body     string
		wantCode int
	}{
		{
			name:     "when the health probe is requested",
			method:   http.MethodGet,
			path:     "/healthz",
			wantCode: http.StatusOK,
		},
		{
			name:     "when registration is posted a malformed body",
			method:   http.MethodPost,
			path:     "/api/v1/auth/register",
			body:     `{not json`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "when login is posted a malformed body",
			method:   http.MethodPost,
			path:     "/api/v1/auth/login",
			body:     `{bad`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:   "when refresh is called without a refresh cookie",
			method: http.MethodPost,
			path:   "/api/v1/auth/refresh",
			// The handler itself answers 401 for a missing refresh cookie; the point
			// here is that the Authenticate gate never runs (callCount stays zero).
			wantCode: http.StatusUnauthorized,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			mockDB := dbmocks.NewMockClient(t)
			if testCase.path == "/healthz" {
				mockDB.EXPECT().Ping(mock.Anything).Return(nil)
			}
			srv, validator := newGatedServer(t, mockDB)

			// Act: present an access_token cookie anyway — on a gated route the
			// validator would be consulted, on a public one it must not be.
			rec := requestWithCookie(t, srv, testCase.method, testCase.path, testCase.body)

			// Assert
			assert.Equal(t, testCase.wantCode, rec.Code)
			assert.Zero(t, validator.callCount,
				"a public route must not be guarded by the Authenticate gate")
		})
	}
}

func TestHTTPServer_ProtectedRoutes_ValidAccessTokenCookie_ReachesHandlerWithUserContext(t *testing.T) {
	// Arrange
	var (
		gotUserID    string
		gotUserOK    bool
		gotHasSub    bool
		gotHasSubOK  bool
		probeReached bool
	)
	probe := withProtectedRoutes(func(r chi.Router) {
		r.Get("/auth-probe", func(w http.ResponseWriter, r *http.Request) {
			probeReached = true
			gotUserID, gotUserOK = middleware.UserIDFromContext(r.Context())
			gotHasSub, gotHasSubOK = middleware.HasSubscriptionFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})
	})
	srv, validator := newGatedServer(t, dbmocks.NewMockClient(t), probe)

	// Act
	rec := doRequest(t, srv, http.MethodGet, probePath, true)

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, probeReached, "a valid cookie must let the request through to the handler")
	assert.Equal(t, 1, validator.callCount)
	assert.True(t, gotUserOK, "middleware.UserIDFromContext must resolve inside the protected group")
	assert.Equal(t, gatedUserID, gotUserID)
	assert.True(t, gotHasSubOK, "middleware.HasSubscriptionFromContext must resolve inside the protected group")
	assert.True(t, gotHasSub)
}

// TestHTTPServer_SwaggerRoutes_NonProductionConfig_ServesWithoutAuthentication
// is issue #192 Task C: /swagger/index.html sitting behind the Authenticate
// gate is real onboarding friction locally (401 on first load, before any
// login), so outside production the route is reachable with no cookie at all.
func TestHTTPServer_SwaggerRoutes_NonProductionConfig_ServesWithoutAuthentication(t *testing.T) {
	// Arrange
	srv, validator := newGatedServer(t, dbmocks.NewMockClient(t))

	// Act: no cookie presented at all.
	rec := doRequest(t, srv, http.MethodGet, "/swagger/doc.json", false)

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Zero(t, validator.callCount, "a non-production /swagger/* request must not consult the gate")

	var doc map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	assert.Equal(t, "2.0", doc["swagger"])
}

// TestHTTPServer_SwaggerRoutes_ProductionConfig_NoAccessTokenCookie_Returns401AuthenticationRequired
// preserves the pre-existing production behavior (specs/009-api-documentation/research.md's
// "Auth gating for Swagger UI ahead of Sprint 5" decision): /swagger/* stays
// gated in production, issue #192 Task C only relaxes it outside production.
func TestHTTPServer_SwaggerRoutes_ProductionConfig_NoAccessTokenCookie_Returns401AuthenticationRequired(t *testing.T) {
	// Arrange
	srv, validator := newGatedServerWithConfig(t, dbmocks.NewMockClient(t), productionTestConfig(t))

	// Act
	rec := doRequest(t, srv, http.MethodGet, "/swagger/doc.json", false)

	// Assert
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "authentication_required", decodeErrorCode(t, rec))
	assert.Zero(t, validator.callCount, "a missing cookie must be rejected before any token validation")
}

// TestHTTPServer_SwaggerRoutes_ProductionConfig_ValidAccessTokenCookie_ServesTheContract
// is the production counterpart of
// TestHTTPServer_SwaggerRoutes_NonProductionConfig_ServesWithoutAuthentication.
func TestHTTPServer_SwaggerRoutes_ProductionConfig_ValidAccessTokenCookie_ServesTheContract(t *testing.T) {
	// Arrange
	srv, validator := newGatedServerWithConfig(t, dbmocks.NewMockClient(t), productionTestConfig(t))

	// Act
	rec := doRequest(t, srv, http.MethodGet, "/swagger/doc.json", true)

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, validator.callCount)

	var doc map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	assert.Equal(t, "2.0", doc["swagger"])
}

func TestHTTPServer_ProtectedAuthRoutes_LogoutWithValidCookie_Returns204(t *testing.T) {
	// Arrange: no refresh cookie, so logout short-circuits without touching the DB.
	srv, validator := newGatedServer(t, dbmocks.NewMockClient(t))

	// Act
	rec := doRequest(t, srv, http.MethodPost, "/api/v1/auth/logout", true)

	// Assert
	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, 1, validator.callCount)
}
