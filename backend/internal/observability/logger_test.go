package observability

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withCapturedDefaultLogger swaps slog's default logger for a buffer-backed
// one, restored via t.Cleanup.
func withCapturedDefaultLogger(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	original := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() {
		slog.SetDefault(original)
	})
	return &buf
}

func decodeLastLogLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	require.NotEmpty(t, lines)

	var entry map[string]any
	require.NoError(t, json.Unmarshal(lines[len(lines)-1], &entry))
	return entry
}

func TestLogSecurityEvent(t *testing.T) {
	tests := []struct {
		name          string
		correlationID string
		eventType     EventType
		userID        string
		severity      Severity
		ipAddress     string
		userAgent     string
		details       map[string]any
		wantEventType string
		wantLevel     string
	}{
		{
			name:          "when a login failure is logged it should record every security event field at WARN level",
			correlationID: "c1a1e2b3-4d5e-6f7a-8b9c-0d1e2f3a4b5c",
			eventType:     EventAuthLoginFailure,
			userID:        "9f1c1a1e-2b3d-4c5e-8f6a-1234567890ab",
			severity:      SeverityWarning,
			ipAddress:     "203.0.113.42",
			userAgent:     "Mozilla/5.0",
			details:       map[string]any{"user_email": "user@example.com", "failure_reason": "invalid_credentials"},
			wantEventType: "auth_login_failure",
			wantLevel:     "WARN",
		},
		{
			name:          "when a SQL injection attempt is logged without a user it should record an empty user_id at ERROR level",
			correlationID: "a2b2f3c4-5d6e-7f8a-9b0c-1d2e3f4a5b6c",
			eventType:     EventValidationSQLInjection,
			userID:        "",
			severity:      SeverityError,
			ipAddress:     "198.51.100.7",
			userAgent:     "curl/8.0",
			details:       map[string]any{"endpoint": "/api/v1/trips", "payload_preview": "' OR 1=1--"},
			wantEventType: "validation_sql_injection",
			wantLevel:     "ERROR",
		},
		{
			name:          "when a registration is logged it should record the auth_registration event at INFO level",
			correlationID: "c4d4b5e6-7f8a-9b0c-1d2e-3f4a5b6c7d8e",
			eventType:     EventAuthRegistration,
			userID:        "0a1b2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d",
			severity:      SeverityInfo,
			ipAddress:     "203.0.113.9",
			userAgent:     "Mozilla/5.0",
			details:       map[string]any{"has_subscription": true},
			wantEventType: "auth_registration",
			wantLevel:     "INFO",
		},
		{
			name:          "when a login success is logged it should record the event at INFO level",
			correlationID: "b3c3a4d5-6e7f-8a9b-0c1d-2e3f4a5b6c7d",
			eventType:     EventAuthLoginSuccess,
			userID:        "9f1c1a1e-2b3d-4c5e-8f6a-1234567890ab",
			severity:      SeverityInfo,
			ipAddress:     "203.0.113.42",
			userAgent:     "Mozilla/5.0",
			details:       map[string]any{"user_email": "user@example.com"},
			wantEventType: "auth_login_success",
			wantLevel:     "INFO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := withCapturedDefaultLogger(t)

			LogSecurityEvent(
				tt.correlationID,
				tt.eventType,
				tt.userID,
				tt.severity,
				tt.ipAddress,
				tt.userAgent,
				tt.details,
			)

			entry := decodeLastLogLine(t, buf)
			assertSecurityEventEntry(t, entry, tt.correlationID, tt.wantEventType, tt.userID, tt.wantLevel, tt.ipAddress, tt.userAgent, tt.details)
		})
	}
}

// assertSecurityEventEntry verifies the shared field contract of a decoded
// security event log line, including a subset match on details.
func assertSecurityEventEntry(
	t *testing.T,
	entry map[string]any,
	correlationID, eventType, userID, level, ipAddress, userAgent string,
	details map[string]any,
) {
	t.Helper()

	assert.Equal(t, correlationID, entry["correlation_id"])
	assert.Equal(t, eventType, entry["event_type"])
	assert.Equal(t, userID, entry["user_id"])
	assert.Equal(t, level, entry["level"])
	assert.Equal(t, ipAddress, entry["ip_address"])
	assert.Equal(t, userAgent, entry["user_agent"])
	require.Contains(t, entry, "time")

	require.Contains(t, entry, "details")
	loggedDetails, ok := entry["details"].(map[string]any)
	require.True(t, ok)
	for key, want := range details {
		assert.Equal(t, want, loggedDetails[key])
	}
}
