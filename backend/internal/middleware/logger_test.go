package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func decodeLastLogLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	require.NotEmpty(t, lines)

	var entry map[string]any
	require.NoError(t, json.Unmarshal(lines[len(lines)-1], &entry))
	return entry
}

func TestLogger_StatusCode_DeterminesLogLevel(t *testing.T) {
	tests := []struct {
		name       string
		wantLevel  string
		statusCode int
	}{
		{name: "2xx logs at INFO", statusCode: http.StatusOK, wantLevel: "INFO"},
		{name: "3xx logs at INFO", statusCode: http.StatusFound, wantLevel: "INFO"},
		{name: "4xx logs at WARN", statusCode: http.StatusNotFound, wantLevel: "WARN"},
		{name: "5xx logs at ERROR", statusCode: http.StatusInternalServerError, wantLevel: "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := newTestLogger(&buf)

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			req := httptest.NewRequest(http.MethodGet, "/trips", nil)
			rec := httptest.NewRecorder()

			Logger(logger)(next).ServeHTTP(rec, req)

			entry := decodeLastLogLine(t, &buf)
			assert.Equal(t, tt.wantLevel, entry["level"])
		})
	}
}

func TestLogger_RequestCompletes_LogsExpectedFields(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/trips", nil)
	ctx := context.WithValue(req.Context(), ctxKeyRequestID{}, "log-fields-id")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	Logger(logger)(next).ServeHTTP(rec, req)

	entry := decodeLastLogLine(t, &buf)
	assert.Equal(t, "HTTP request", entry["msg"])
	assert.Equal(t, http.MethodPost, entry["method"])
	assert.Equal(t, "/trips", entry["path"])
	assert.EqualValues(t, http.StatusCreated, entry["status"])
	assert.Equal(t, "log-fields-id", entry["request_id"])
	assert.Contains(t, entry, "duration_ms")
}

func TestLogger_NoRequestIDInContext_LogsEmptyString(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		Logger(logger)(next).ServeHTTP(rec, req)
	})

	entry := decodeLastLogLine(t, &buf)
	assert.Equal(t, "", entry["request_id"])
}
