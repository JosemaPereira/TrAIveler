package middleware

import (
	"context"
	"net/http"
)

// accessTokenCookie is the HTTP-only cookie carrying the JWT access token, per
// docs/security.md ("Token Storage: HTTP-only, Secure, SameSite=Strict").
const accessTokenCookie = "access_token"

// AuthClaims is the minimal set of authorization facts Authenticate propagates
// downstream. It is a middleware-local port, deliberately decoupled from
// auth/jwt.Claims: internal/errors imports this package, and auth/jwt imports
// internal/errors, so importing auth/jwt here would form a cycle. The caller
// wiring a real validator (in package main) adapts *jwt.Claims into this.
type AuthClaims struct {
	UserID          string
	HasSubscription bool
}

// TokenValidator verifies an access token and returns its claims. Authenticate
// depends on this narrow interface so it can be tested without real RSA keys
// and stays independent of the concrete token implementation.
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (AuthClaims, error)
}

// Authenticate returns middleware that requires a valid access_token cookie.
// It validates the JWT via validator and, on success, attaches the user ID and
// subscription flag to the request context (read them with UserIDFromContext /
// HasSubscriptionFromContext). Any missing, empty, or invalid token yields a
// 401 authentication_required envelope and next is never reached — the generic
// message avoids leaking whether the token was absent, malformed, or expired.
func Authenticate(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(accessTokenCookie)
			if err != nil || cookie.Value == "" {
				writeUnauthorized(w, r)
				return
			}

			claims, err := validator.ValidateToken(r.Context(), cookie.Value)
			if err != nil {
				writeUnauthorized(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyUserID{}, claims.UserID)
			ctx = context.WithValue(ctx, ctxKeyHasSubscription{}, claims.HasSubscription)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// writeUnauthorized emits the standard 401 envelope. The wire code
// "authentication_required" matches internal/errors.Unauthorized and
// docs/api-design-standards.md §7.
func writeUnauthorized(w http.ResponseWriter, r *http.Request) {
	writeErrorEnvelope(w, r, http.StatusUnauthorized, "authentication_required", "Authentication required")
}
