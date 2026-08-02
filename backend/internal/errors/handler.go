package errors

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
)

// ErrorResponse is the wire shape of the standard error envelope
// (docs/api-design-standards.md §7). It is wire-compatible with, but
// distinct from, middleware's package-private errorEnvelope: this exported
// version additionally carries "fields" (validation errors) and "details".
// The two types can't be unified because this package already imports
// middleware (for RequestIDFromContext below) — middleware importing back
// would create an import cycle. Exported (rather than the package-private
// name used elsewhere in this file) so swag doc annotations across the
// codebase can reference it directly, e.g. `@Failure 404 {object}
// errors.ErrorResponse` in internal/auth/handler.go.
type ErrorResponse struct {
	Error     string            `json:"error"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id"`
	Fields    []ValidationError `json:"fields,omitempty"`
	Details   map[string]any    `json:"details,omitempty"`
}

// HandleError writes the standard JSON error envelope for err. Recognized
// *DomainError values are mapped to their catalog status/code; any other
// error is treated as an unmapped internal_error/500 — its real message is
// logged server-side (via slog.Default, since this signature takes no
// logger) but never sent to the client.
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	var domainErr *DomainError
	if !errors.As(err, &domainErr) {
		requestID, _ := middleware.RequestIDFromContext(r.Context())
		slog.Default().LogAttrs(r.Context(), slog.LevelError, "unhandled internal error",
			slog.Any("error", err),
			slog.String("request_id", requestID),
		)
		domainErr = &DomainError{Code: "internal_error", Message: "An unexpected error occurred"}
	}

	writeErrorResponse(w, r, domainErr)
}

// writeErrorResponse encodes domainErr as the standard JSON error envelope.
// A "service_unavailable" domainErr (see ServiceUnavailable in types.go)
// additionally gets a Retry-After response header, echoing the
// "retry_after_seconds" value already present in its Details — e.g. the AI
// client factory's provider-unavailable translation (client.go,
// docs/roadmap.md 005-T113) surfacing an exhausted-retries failure this way.
func writeErrorResponse(w http.ResponseWriter, r *http.Request, domainErr *DomainError) {
	requestID, _ := middleware.RequestIDFromContext(r.Context())

	w.Header().Set("Content-Type", "application/json")
	if domainErr.Code == "service_unavailable" || domainErr.Code == "rate_limit_exceeded" {
		if retryAfterSeconds, ok := domainErr.Details["retry_after_seconds"].(int); ok {
			w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
		}
	}
	w.WriteHeader(statusForCode(domainErr.Code))
	if err := json.NewEncoder(w).Encode(ErrorResponse{
		Error:     domainErr.Code,
		Message:   domainErr.Message,
		RequestID: requestID,
		Fields:    domainErr.Fields,
		Details:   domainErr.Details,
	}); err != nil {
		slog.Default().Error("failed to write error response", "error", err, "request_id", requestID)
	}
}

// statusForCode maps a domain error code to its HTTP status per the catalog
// in docs/api-design-standards.md §7-8. Unrecognized codes default to 500,
// so the mapping stays safe to extend with new codes later.
func statusForCode(code string) int {
	switch code {
	case "not_found":
		return http.StatusNotFound
	case "validation_failed":
		return http.StatusUnprocessableEntity
	case "authentication_required":
		return http.StatusUnauthorized
	case "forbidden":
		return http.StatusForbidden
	case "conflict":
		return http.StatusConflict
	case "service_unavailable":
		return http.StatusServiceUnavailable
	case "rate_limit_exceeded":
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
