package middleware

import (
	"log/slog"
	"net/http"

	"github.com/felixge/httpsnoop"
)

// Logger returns middleware that logs one structured JSON line per completed
// request via the given slog.Logger, including method, path, status,
// duration, and the request ID propagated by RequestID.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metrics := httpsnoop.CaptureMetrics(next, w, r)

			requestID, _ := RequestIDFromContext(r.Context())
			logger.LogAttrs(r.Context(), levelForStatus(metrics.Code), "HTTP request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", metrics.Code),
				slog.Int64("duration_ms", metrics.Duration.Milliseconds()),
				slog.String("request_id", requestID),
			)
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
