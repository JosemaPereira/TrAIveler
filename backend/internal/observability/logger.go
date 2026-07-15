package observability

import (
	"context"
	"log/slog"
)

// Severity is a SecurityEvent's severity level, matching the database
// CHECK (severity IN ('info', 'warning', 'error')) constraint on the
// security_events table (see backend/migrations/004_create_security_events_table.sql
// and specs/004-security-auth-model/data-model.md, "Entity 4: SecurityEvent").
type Severity string

// Severity levels, mirroring the security_events table's severity CHECK
// constraint verbatim.
const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

// EventType is a SecurityEvent's event type, matching the database's
// event_type column. Values are DB-persisted strings and must stay in sync
// verbatim with data-model.md's "Event Types" table.
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

// severityToLevel maps a SecurityEvent Severity to the slog.Level it is
// logged at, so log aggregators (CloudWatch Logs, via the ECS awslogs
// driver reading structured stdout JSON — see internal/middleware.Logger for
// the established precedent) can filter/alarm on severity the same way they
// already do for HTTP request logs.
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

// LogSecurityEvent emits one structured JSON log line for a security-relevant
// event (authentication, authorization, input validation, concurrency
// conflicts — see specs/004-security-auth-model/data-model.md, "Entity 4:
// SecurityEvent"), mirroring that entity's attributes so the log line carries
// everything a later DB-persistence step (Spec 004 Phase 3, out of scope
// here) would need to write a security_events row.
//
// userID, ipAddress, and userAgent are nullable columns in the schema;
// callers pass an empty string when the value is not available (e.g. a
// login-failure event before a user is identified, or a background job with
// no request IP). They are always included in the log line, empty or not, so
// every SecurityEvent log line has the same fixed field shape.
//
// It uses slog.Default() rather than taking a *slog.Logger parameter,
// matching this codebase's established precedent for call sites with no room
// for a constructor-injected logger (see internal/errors.HandleError).
//
// LogSecurityEvent does not write to the security_events database table —
// persistence is a later Phase 3 concern. It also does not sanitize details;
// per FR-050/docs/security.md, callers are responsible for never including
// passwords, tokens, or full payloads in details.
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
