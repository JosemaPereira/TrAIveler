package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/felixge/httpsnoop"
)

// serviceName is the fixed value emitted on every structured log entry's
// "service" field, identifying the emitting process to log aggregators.
const serviceName = "traiveler-api"

// Logger returns middleware that logs one structured JSON line per completed
// request via the given slog.Logger, including service, method, path,
// status, duration, the request ID propagated by RequestID, and — when the
// request is authenticated — the user ID propagated via UserIDFromContext.
// A 5xx response also gets a "stack" field (same shape Recovery uses for
// panics), so operator-actionable failures are traceable even when the
// handler returned an error instead of panicking.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metrics := httpsnoop.CaptureMetrics(next, w, r)

			requestID, _ := RequestIDFromContext(r.Context())
			attrs := []slog.Attr{
				slog.String("service", serviceName),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", metrics.Code),
				slog.Int64("duration_ms", metrics.Duration.Milliseconds()),
				slog.String("request_id", requestID),
			}

			// user_id must never appear (not even as an empty string) on
			// unauthenticated requests, per NFR-PRIV-002 — only append it
			// when a value is actually present in the context.
			if userID, ok := UserIDFromContext(r.Context()); ok {
				attrs = append(attrs, slog.String("user_id", userID))
			}

			// A handler that returns a 5xx without panicking (e.g. an
			// unmapped error going through errors.HandleError ->
			// internal_error -> 500) never reaches Recovery's panic-only
			// stack capture, so it would otherwise be logged with no stack
			// trace at all. Capture one here too, in the same field shape
			// Recovery already uses, whenever the response itself signals a
			// server-side failure.
			if metrics.Code >= http.StatusInternalServerError {
				attrs = append(attrs, slog.String("stack", string(debug.Stack())))
			}

			logger.LogAttrs(r.Context(), levelForStatus(metrics.Code), "HTTP request", attrs...)
		})
	}
}

// levelForStatus maps an HTTP status code to a log severity: 5xx is an
// operator-actionable failure, 4xx is a client-caused warning, everything
// else is routine informational traffic.
func levelForStatus(status int) slog.Level {
	switch {
	case status >= http.StatusInternalServerError:
		return slog.LevelError
	case status >= http.StatusBadRequest:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}
