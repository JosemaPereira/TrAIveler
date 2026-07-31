package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// meResponse mirrors auth.CurrentUserResponse on the wire. Declared here rather
// than imported so these black-box tests assert against the JSON the server
// actually emits, not against the struct that produced it — a json tag typo
// would otherwise pass unnoticed.
type meResponse struct {
	User struct {
		ID              string `json:"id"`
		Email           string `json:"email"`
		FullName        string `json:"full_name"`
		HasSubscription bool   `json:"has_subscription"`
		CreatedAt       string `json:"created_at"`
	} `json:"user"`
}

// TestAuthMe_WithSessionCookies_ReturnsTheRegisteredAccount is the load-bearing
// test for 001-T021: it proves a client holding nothing but the HTTP-only
// session cookies can re-derive who it is. That is exactly what the frontend
// does after a page reload, where no user data survives in memory.
func TestAuthMe_WithSessionCookies_ReturnsTheRegisteredAccount(t *testing.T) {
	baseURL := setupAPITestServer(t)
	sessionCookies := registerTestSession(t, baseURL)

	client := &http.Client{Timeout: 5 * time.Second}
	resp := authenticatedGet(t, client, baseURL+"/api/v1/auth/me", sessionCookies)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body meResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.NotEmpty(t, body.User.ID)
	assert.Contains(t, body.User.Email, "@example.com")
	assert.Equal(t, "Integration Test User", body.User.FullName)
	assert.False(t, body.User.HasSubscription,
		"registerTestSession sends no payment token, so the account is a Free User")
	assert.NotEmpty(t, body.User.CreatedAt)
}

// TestAuthMe_NeverLeaksSensitiveFields guards the json:"-" tags on the User
// model. A future field added without one would silently start shipping to
// every client; asserting on the raw JSON keys catches that, whereas decoding
// into a typed struct would quietly discard the extras.
func TestAuthMe_NeverLeaksSensitiveFields(t *testing.T) {
	baseURL := setupAPITestServer(t)
	sessionCookies := registerTestSession(t, baseURL)

	client := &http.Client{Timeout: 5 * time.Second}
	resp := authenticatedGet(t, client, baseURL+"/api/v1/auth/me", sessionCookies)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var envelope map[string]json.RawMessage
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))
	var user map[string]any
	require.NoError(t, json.Unmarshal(envelope["user"], &user))

	for _, forbidden := range []string{
		"password_hash", "role", "version", "failed_login_attempts",
		"last_failed_login_at", "email_verified", "updated_at",
	} {
		assert.NotContains(t, user, forbidden,
			"%s carries json:\"-\" and must never reach a client", forbidden)
	}
	assert.Len(t, user, 5, "the wire shape is exactly id/email/full_name/has_subscription/created_at")
}

// TestAuthMe_WithoutSession_Returns401 pins that the endpoint really sits behind
// the gate. Without this, a routing mistake that mounted it publicly would leak
// nothing by itself (there is no id to look up) but would silently break the
// frontend's "401 means not signed in" contract.
func TestAuthMe_WithoutSession_Returns401(t *testing.T) {
	baseURL := setupAPITestServer(t)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/api/v1/auth/me")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "authentication_required", body["error"],
		"an absent token is not token_expired: the frontend must not try to refresh")
}

// TestAuthMe_AfterLogout_AccessTokenOutlivesTheSession documents a real and
// security-relevant property rather than asserting the comfortable answer: the
// access token is a stateless JWT, so logout cannot invalidate it. Logout
// revokes the *refresh* token and expires the cookies, which is what stops a
// real browser — but a caller that kept a copy of the old access token keeps
// working until it expires (24h, config.Auth.JWTExpiration).
//
// If this test ever starts failing with a 401, that is not a regression: it
// means access tokens gained server-side revocation (a denylist), and the
// trade-off recorded here should be revisited in docs/security.md.
func TestAuthMe_AfterLogout_AccessTokenOutlivesTheSession(t *testing.T) {
	baseURL := setupAPITestServer(t)
	sessionCookies := registerTestSession(t, baseURL)

	client := &http.Client{Timeout: 5 * time.Second}

	logout, err := client.Do(newAuthenticatedRequest(t, http.MethodPost, baseURL+"/api/v1/auth/logout", sessionCookies, nil))
	require.NoError(t, err)
	require.NoError(t, logout.Body.Close())
	require.Equal(t, http.StatusNoContent, logout.StatusCode)

	for _, cookie := range logout.Cookies() {
		if cookie.Name == "access_token" || cookie.Name == "refresh_token" {
			assert.Negative(t, cookie.MaxAge,
				"logout must expire %s so a real browser stops sending it", cookie.Name)
		}
	}

	// Deliberately replaying the pre-logout cookies, which a browser would no
	// longer hold.
	resp := authenticatedGet(t, client, baseURL+"/api/v1/auth/me", sessionCookies)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode,
		"a stateless access token is not revoked by logout; cookie clearing is what ends the session")
}
