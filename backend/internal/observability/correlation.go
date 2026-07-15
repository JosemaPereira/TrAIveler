package observability

import "github.com/google/uuid"

// GenerateCorrelationID returns a new UUID v4 for a SecurityEvent's
// correlation_id. Callers propagating an existing correlation ID should
// reuse it rather than generating a new one.
func GenerateCorrelationID() string {
	return uuid.NewString()
}
