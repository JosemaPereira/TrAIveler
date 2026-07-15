package observability

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withCapturedDefaultLogger temporarily replaces slog's global default logger
// with one writing to a buffer, restoring the original when the test ends.
// LogSecurityEvent has no logger parameter (mirrors internal/errors.HandleError's
// established precedent for signatures with no room for a constructor-injected
// logger), so slog.Default() is the only way to observe its output.
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

func TestLogSecurityEvent_AuthLoginFailure_LogsAllSecurityEventFields(t *testing.T) {
	buf := withCapturedDefaultLogger(t)

	LogSecurityEvent(
		"c1a1e2b3-4d5e-6f7a-8b9c-0d1e2f3a4b5c",
		EventAuthLoginFailure,
		"9f1c1a1e-2b3d-4c5e-8f6a-1234567890ab",
		SeverityWarning,
		"203.0.113.42",
		"Mozilla/5.0",
		map[string]any{"user_email": "user@example.com", "failure_reason": "invalid_credentials"},
	)

	entry := decodeLastLogLine(t, buf)
	assert.Equal(t, "c1a1e2b3-4d5e-6f7a-8b9c-0d1e2f3a4b5c", entry["correlation_id"])
	assert.Equal(t, "auth_login_failure", entry["event_type"])
	assert.Equal(t, "9f1c1a1e-2b3d-4c5e-8f6a-1234567890ab", entry["user_id"])
	assert.Equal(t, "WARN", entry["level"])
	assert.Equal(t, "203.0.113.42", entry["ip_address"])
	assert.Equal(t, "Mozilla/5.0", entry["user_agent"])
	require.Contains(t, entry, "details")
	details, ok := entry["details"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "user@example.com", details["user_email"])
	assert.Equal(t, "invalid_credentials", details["failure_reason"])
	require.Contains(t, entry, "time")
}

func TestLogSecurityEvent_ValidationSQLInjection_LogsAllSecurityEventFields(t *testing.T) {
	buf := withCapturedDefaultLogger(t)

	LogSecurityEvent(
		"a2b2f3c4-5d6e-7f8a-9b0c-1d2e3f4a5b6c",
		EventValidationSQLInjection,
		"",
		SeverityError,
		"198.51.100.7",
		"curl/8.0",
		map[string]any{"endpoint": "/api/v1/trips", "payload_preview": "' OR 1=1--"},
	)

	entry := decodeLastLogLine(t, buf)
	assert.Equal(t, "a2b2f3c4-5d6e-7f8a-9b0c-1d2e3f4a5b6c", entry["correlation_id"])
	assert.Equal(t, "validation_sql_injection", entry["event_type"])
	assert.Equal(t, "", entry["user_id"])
	assert.Equal(t, "ERROR", entry["level"])
	assert.Equal(t, "198.51.100.7", entry["ip_address"])
	assert.Equal(t, "curl/8.0", entry["user_agent"])
	require.Contains(t, entry, "details")
	details, ok := entry["details"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "/api/v1/trips", details["endpoint"])
}

func TestLogSecurityEvent_AuthLoginSuccess_LogsInfoLevel(t *testing.T) {
	buf := withCapturedDefaultLogger(t)

	LogSecurityEvent(
		"b3c3a4d5-6e7f-8a9b-0c1d-2e3f4a5b6c7d",
		EventAuthLoginSuccess,
		"9f1c1a1e-2b3d-4c5e-8f6a-1234567890ab",
		SeverityInfo,
		"203.0.113.42",
		"Mozilla/5.0",
		map[string]any{"user_email": "user@example.com"},
	)

	entry := decodeLastLogLine(t, buf)
	assert.Equal(t, "auth_login_success", entry["event_type"])
	assert.Equal(t, "INFO", entry["level"])
}
