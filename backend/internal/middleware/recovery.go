package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recovery returns middleware that recovers from panics in downstream
// handlers, logs the panic value and stack trace at ERROR level, and
// responds with a generic 500 error envelope instead of crashing the
// server or leaking implementation details to the client. If next already
// wrote response headers before panicking, the 500 status is silently
// dropped by net/http and the envelope is appended best-effort — the
// original partial response cannot be undone.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer recoverAndRespond(logger, w, r)
			next.ServeHTTP(w, r)
		})
	}
}

// recoverAndRespond catches a panic (if any), logs it, and writes the
// standard internal-error envelope. It is a no-op when no panic occurred.
func recoverAndRespond(logger *slog.Logger, w http.ResponseWriter, r *http.Request) {
	rec := recover()
	if rec == nil {
		return
	}

	requestID, _ := RequestIDFromContext(r.Context())
	logger.LogAttrs(r.Context(), slog.LevelError, "panic recovered",
		slog.Any("panic", rec),
		slog.String("stack", string(debug.Stack())),
		slog.String("request_id", requestID),
	)

	writeErrorEnvelope(w, r, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
}
