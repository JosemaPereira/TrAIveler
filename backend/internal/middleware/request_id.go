// Package middleware provides standard library-compatible HTTP middleware
// (func(http.Handler) http.Handler) shared across the backend's HTTP layer.
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// requestIDHeader is the canonical header name used to propagate the request ID
// between clients, this service, and any downstream services.
const requestIDHeader = "X-Request-ID"

// ctxKeyRequestID is an unexported type so context values set here can never
// collide with a key from another package (see revive's context-keys-type rule).
type ctxKeyRequestID struct{}

// RequestID assigns a unique identifier to each request, reusing an incoming
// X-Request-ID header when present or generating a new UUID v4 otherwise. The
// ID is stored in the request context and echoed back on the response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}

		w.Header().Set(requestIDHeader, id)
		ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext returns the request ID stored by RequestID, and whether
// one was present in the context.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKeyRequestID{}).(string)
	return id, ok
}
