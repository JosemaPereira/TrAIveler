package middleware

import (
	"context"
	"errors"
	"net/http"
)

// accessTokenCookie is the HTTP-only cookie carrying the JWT access token, per
// docs/security.md ("Token Storage: HTTP-only, Secure, SameSite=Strict").
const accessTokenCookie = "access_token"

// ErrTokenExpired is the sentinel a TokenValidator returns (typically wrapped)
// when the presented access token is validly signed by us but its exp has
// passed. Authenticate matches it with errors.Is and answers 401 token_expired
// instead of the uniform authentication_required, so the frontend can silently
// refresh (008-T150). It is a deliberately narrow exception to the anti-
// enumeration rule: only a token we truly issued — proven by a valid signature —
// can qualify, so it leaks nothing about account or token existence. Every other
// failure (missing/malformed token, bad signature, unknown key, wrong issuer)
// stays authentication_required. The TokenValidator port signature is unchanged;
// the composition-root adapter maps auth/jwt's expiry error onto this sentinel.
var ErrTokenExpired = errors.New("middleware: access token expired")

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
// HasSubscriptionFromContext). A missing, empty, malformed, or otherwise invalid
// token yields a 401 authentication_required envelope, so nothing leaks about why
// auth failed. The single exception is a validly-signed-but-expired token, which
// the validator flags via ErrTokenExpired and which yields 401 token_expired
// (the frontend's silent-refresh trigger). Either way next is never reached.
func Authenticate(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(accessTokenCookie)
			if err != nil || cookie.Value == "" {
				// An absent token is not "expired": stay uniform.
				writeUnauthorized(w, r)
				return
			}

			claims, err := validator.ValidateToken(r.Context(), cookie.Value)
			if err != nil {
				if errors.Is(err, ErrTokenExpired) {
					writeTokenExpired(w, r)
					return
				}
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

// writeTokenExpired emits a 401 envelope for a validly-signed-but-expired access
// token. It shares the schema of writeUnauthorized — only the wire code differs
// ("token_expired", docs/api-design-standards.md §7) — because the frontend keys
// off the code, not the message, to trigger a silent refresh. The message stays
// honest and non-leaky.
func writeTokenExpired(w http.ResponseWriter, r *http.Request) {
	writeErrorEnvelope(w, r, http.StatusUnauthorized, "token_expired", "Access token expired")
}
