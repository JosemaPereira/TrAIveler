package middleware

import "context"

// ctxKeyUserID / ctxKeyHasSubscription are unexported types so context values
// set here can never collide with a key from another package (see revive's
// context-keys-type rule).
type (
	ctxKeyUserID          struct{}
	ctxKeyHasSubscription struct{}
)

// UserIDFromContext returns the authenticated user's ID stored in the request
// context by Authenticate, and whether one was present. Absent for
// unauthenticated requests.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKeyUserID{}).(string)
	return id, ok
}

// HasSubscriptionFromContext reports whether the authenticated user had an
// active subscription when the access token was issued, as attached by
// Authenticate. The second return is false for unauthenticated requests. The
// flag can go stale within the token's lifetime — flows needing an immediate
// effect must not rely on it (see auth/jwt.Claims).
func HasSubscriptionFromContext(ctx context.Context) (bool, bool) {
	has, ok := ctx.Value(ctxKeyHasSubscription{}).(bool)
	return has, ok
}
