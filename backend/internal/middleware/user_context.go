package middleware

import "context"

// ctxKeyUserID is an unexported type so context values set here can never
// collide with a key from another package (see revive's context-keys-type rule).
type ctxKeyUserID struct{}

// UserIDFromContext returns the authenticated user's ID stored in the request
// context (e.g. by a future JWT authentication middleware), and whether one
// was present. Absent for unauthenticated requests.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKeyUserID{}).(string)
	return id, ok
}
