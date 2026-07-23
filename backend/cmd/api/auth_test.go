//go:build test

package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbmocks "github.com/JosemaPereira/TrAIveler/backend/internal/database/mocks"
)

// These tests verify only that the auth routes are mounted and reachable through
// the real composition root (NewHTTPServer -> buildAuthComponents -> routes).
// They exercise paths that never touch the DB (invalid bodies rejected before any
// repository call, cookie-less logout short-circuits), so the mock DB needs no
// expectations. Full behavior is covered by internal/auth's handler unit tests.

func postTo(t *testing.T, srv *HTTPServer, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)
	return rec
}

func TestHTTPServer_AuthRoutes_RegisterMountedAndValidatesInput(t *testing.T) {
	srv := mustNewHTTPServer(t, dbmocks.NewMockClient(t), testConfig(t), testLogger(t))

	// Malformed JSON is rejected at the transport layer (400), proving the route
	// is mounted without reaching the service/DB.
	malformed := postTo(t, srv, "/api/v1/auth/register", `{not json`)
	assert.Equal(t, http.StatusBadRequest, malformed.Code)

	// A well-formed but invalid body fails validation in the service (422) before
	// any DB lookup.
	invalid := postTo(t, srv, "/api/v1/auth/register",
		`{"email":"not-an-email","password":"weak","full_name":""}`)
	assert.Equal(t, http.StatusUnprocessableEntity, invalid.Code)
}

func TestHTTPServer_AuthRoutes_LoginMountedAndValidatesInput(t *testing.T) {
	srv := mustNewHTTPServer(t, dbmocks.NewMockClient(t), testConfig(t), testLogger(t))

	malformed := postTo(t, srv, "/api/v1/auth/login", `{bad`)
	assert.Equal(t, http.StatusBadRequest, malformed.Code)

	invalid := postTo(t, srv, "/api/v1/auth/login", `{"email":"not-an-email","password":""}`)
	assert.Equal(t, http.StatusUnprocessableEntity, invalid.Code)
}

func TestHTTPServer_AuthRoutes_LogoutMountedAndClearsCookies(t *testing.T) {
	srv := mustNewHTTPServer(t, dbmocks.NewMockClient(t), testConfig(t), testLogger(t))

	// Cookie-less logout is idempotent: it clears cookies and returns 204 without
	// any revocation (no DB call), confirming the route is mounted.
	rec := postTo(t, srv, "/api/v1/auth/logout", "")

	require.Equal(t, http.StatusNoContent, rec.Code)
	var clearedAccess bool
	for _, c := range rec.Result().Cookies() {
		if c.Name == "access_token" && c.MaxAge < 0 {
			clearedAccess = true
		}
	}
	assert.True(t, clearedAccess, "logout must expire the access_token cookie")
}
