package observability

import "github.com/google/uuid"

// GenerateCorrelationID returns a new, randomly generated UUID (v4) suitable
// for use as a SecurityEvent's correlation_id (see
// specs/004-security-auth-model/data-model.md, "Entity 4: SecurityEvent").
// It is a pure generator with no HTTP or context awareness — callers that
// need to propagate an existing correlation ID across a request should reuse
// it rather than generating a new one; see internal/middleware/request_id.go
// for the separate, already-shipped HTTP request-ID concern.
func GenerateCorrelationID() string {
	return uuid.NewString()
}
