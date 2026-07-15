package observability

import (
	"context"
	"log/slog"
)

// Severity is a SecurityEvent's severity level; must match the
// security_events table's severity CHECK constraint verbatim.
type Severity string

// Severity levels.
const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

// EventType is a SecurityEvent's event type; values are DB-persisted strings
// and must stay in sync verbatim with data-model.md's "Event Types" table.
type EventType string

// Event types, one per row of data-model.md's "Event Types" table.
const (
	EventAuthLoginSuccess          EventType = "auth_login_success"
	EventAuthLoginFailure          EventType = "auth_login_failure"
	EventAuthTokenRefresh          EventType = "auth_token_refresh"
	EventAuthLogout                EventType = "auth_logout"
	EventAuthPasswordChange        EventType = "auth_password_change"
	EventAuthzDenied               EventType = "authz_denied"
	EventValidationPromptInjection EventType = "validation_prompt_injection"
	EventValidationSQLInjection    EventType = "validation_sql_injection"
	EventValidationXSSInjection    EventType = "validation_xss_injection"
	EventConcurrencyConflict       EventType = "concurrency_conflict"
)

func severityToLevel(severity Severity) slog.Level {
	switch severity {
	case SeverityError:
		return slog.LevelError
	case SeverityWarning:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

// LogSecurityEvent emits one structured JSON log line for a security event.
// It does not persist to the security_events table (Phase 3, out of scope
// here) and does not sanitize details — per docs/security.md, callers must
// never include passwords, tokens, or full payloads in details.
func LogSecurityEvent(
	correlationID string,
	eventType EventType,
	userID string,
	severity Severity,
	ipAddress, userAgent string,
	details map[string]any,
) {
	slog.Default().LogAttrs(context.Background(), severityToLevel(severity), "security event",
		slog.String("correlation_id", correlationID),
		slog.String("event_type", string(eventType)),
		slog.String("user_id", userID),
		slog.String("ip_address", ipAddress),
		slog.String("user_agent", userAgent),
		slog.Any("details", details),
	)
}
