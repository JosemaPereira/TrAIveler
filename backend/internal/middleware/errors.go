package middleware

import (
	"encoding/json"
	"net/http"
)

// errorEnvelope is the standard error response body defined in
// docs/api-design-standards.md §7.
type errorEnvelope struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// writeErrorEnvelope writes a JSON error response following the project's
// standard error format, pulling the request ID from context when present.
func writeErrorEnvelope(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	requestID, _ := RequestIDFromContext(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorEnvelope{
		Error:     code,
		Message:   message,
		RequestID: requestID,
	})
}
