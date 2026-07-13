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

			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/trips", nil)
	ctx := context.WithValue(req.Context(), ctxKeyRequestID{}, "log-fields-id")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	Logger(logger)(next).ServeHTTP(rec, req)

	entry := decodeLastLogLine(t, &buf)
	assert.Equal(t, "HTTP request", entry["msg"])
	assert.Equal(t, "traiveler-api", entry["service"])
	assert.Equal(t, http.MethodPost, entry["method"])
	assert.Equal(t, "/trips", entry["path"])
	assert.EqualValues(t, http.StatusCreated, entry["status"])
	assert.Equal(t, "log-fields-id", entry["request_id"])
	require.Contains(t, entry, "duration_ms")
	assert.GreaterOrEqual(t, entry["duration_ms"], float64(0))
}

func TestLogger_NoRequestIDInContext_LogsEmptyString(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

func TestLogger_UserIDInContext_LogsUserID(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	ctx := context.WithValue(req.Context(), ctxKeyUserID{}, "9f1c1a1e-2b3d-4c5e-8f6a-1234567890ab")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	Logger(logger)(next).ServeHTTP(rec, req)

	entry := decodeLastLogLine(t, &buf)
	assert.Equal(t, "9f1c1a1e-2b3d-4c5e-8f6a-1234567890ab", entry["user_id"])
}

func TestLogger_NoUserIDInContext_OmitsUserIDField(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	rec := httptest.NewRecorder()

	Logger(logger)(next).ServeHTTP(rec, req)

	entry := decodeLastLogLine(t, &buf)
	_, ok := entry["user_id"]
	assert.False(t, ok, "user_id key must be entirely absent for unauthenticated requests")
}

func TestLogger_ResponseStatus5xx_IncludesStackField(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	rec := httptest.NewRecorder()

	Logger(logger)(next).ServeHTTP(rec, req)

	entry := decodeLastLogLine(t, &buf)
	require.Contains(t, entry, "stack")
	stack, ok := entry["stack"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, stack)
}

func TestLogger_ResponseStatusBelow5xx_OmitsStackField(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{name: "2xx omits stack", statusCode: http.StatusOK},
		{name: "4xx omits stack", statusCode: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := newTestLogger(&buf)

			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			req := httptest.NewRequest(http.MethodGet, "/trips", nil)
			rec := httptest.NewRecorder()

			Logger(logger)(next).ServeHTTP(rec, req)

			entry := decodeLastLogLine(t, &buf)
			_, ok := entry["stack"]
			assert.False(t, ok, "stack key must be absent for non-5xx responses")
		})
	}
}
