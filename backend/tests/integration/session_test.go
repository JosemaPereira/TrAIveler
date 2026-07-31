package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// testAccountPassword satisfies internal/auth's password policy (length plus
// upper/lower/digit); it is only ever used by throwaway containerized accounts.
const testAccountPassword = "IntegrationPass1!"

// registerTestSession creates a fresh account through the public
// POST /api/v1/auth/register route and returns the session cookies the server
// set. Since 008-T208 gated /api/v1 and /swagger/*, black-box tests that hit a
// protected route need a genuine session, and registration is the only entry
// point that mints one without pre-seeded data: it is public, DB-backed, and —
// with no payment_method_token — never touches the plans/subscriptions tables.
//
// The caller must have applied migrations to the database the server runs
// against, otherwise the users table does not exist.
func registerTestSession(t *testing.T, baseURL string) []*http.Cookie {
	t.Helper()

	payload, err := json.Marshal(map[string]string{
		"email":     fmt.Sprintf("integration-%d@example.com", time.Now().UnixNano()),
		"password":  testAccountPassword,
		"full_name": "Integration Test User",
	})
	require.NoError(t, err)

	resp, err := http.Post(baseURL+"/api/v1/auth/register", "application/json", bytes.NewReader(payload))
	require.NoError(t, err, "POST /api/v1/auth/register must succeed at the transport level")
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode,
		"registration must succeed so the rest of the test has a real session")

	cookies := resp.Cookies()
	require.NotEmpty(t, cookies, "registration must set the session cookies")
	require.NotNil(t, cookieNamed(cookies, "access_token"),
		"registration must set an access_token cookie for the authenticated route group")

	return cookies
}

// cookieNamed returns the named cookie, or nil when absent.
func cookieNamed(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

// newAuthenticatedRequest builds a request carrying the session cookies. They
// are attached per-request rather than via an http.CookieJar because these
// tests share http.DefaultClient, which has no jar; a jar would also have to be
// threaded through every call site for no gain here.
func newAuthenticatedRequest(t *testing.T, method, url string, cookies []*http.Cookie, body []byte) *http.Request {
	t.Helper()

	var reader io.Reader = http.NoBody
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	require.NoError(t, err)

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req
}

// authenticatedGet issues an authenticated GET against a protected route.
func authenticatedGet(t *testing.T, client *http.Client, url string, cookies []*http.Cookie) *http.Response {
	t.Helper()

	resp, err := client.Do(newAuthenticatedRequest(t, http.MethodGet, url, cookies, nil))
	require.NoError(t, err, "GET %s must succeed at the transport level", url)
	return resp
}

// setupAPITestServer boots a real backend API process against a
// disposable Postgres testcontainer and returns its base URL, so tests in
// this file can make genuine HTTP requests against the actual
// cmd/api/routes.go route table instead of a hand-built substitute.
//
// Migrations are applied because every route worth black-box testing sits in
// the authenticated group (008-T208): reaching one means registering a real
// account through registerTestSession, which needs the users table to exist.
func setupAPITestServer(t *testing.T) string {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test requiring a container runtime in short mode")
	}

	ctx := context.Background()
	databaseURL := startPostgresContainer(t, ctx)
	applyMigrations(t, databaseURL)
	binPath := buildAPIBinary(t)

	return startAPIServer(t, ctx, binPath, databaseURL)
}
