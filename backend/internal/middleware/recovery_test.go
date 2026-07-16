package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecovery_NextPanics_Returns500WithErrorEnvelope(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(t, &buf)

	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom: unexpected nil pointer")
	})

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	ctx := context.WithValue(req.Context(), ctxKeyRequestID{}, "recovery-test-id")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		Recovery(logger)(next).ServeHTTP(rec, req)
	})

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body errorEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "internal_error", body.Error)
	assert.Equal(t, "An unexpected error occurred", body.Message)
	assert.Equal(t, "recovery-test-id", body.RequestID)
}

func TestRecovery_NextPanics_DoesNotLeakPanicDetailsInResponseBody(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(t, &buf)

	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("sensitive stack trace detail: db_password=secret")
	})

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	rec := httptest.NewRecorder()

	Recovery(logger)(next).ServeHTTP(rec, req)

	assert.NotContains(t, rec.Body.String(), "sensitive stack trace detail")
	assert.NotContains(t, rec.Body.String(), "goroutine")
}

func TestRecovery_NextPanics_LogsPanicAndStackAtErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(t, &buf)

	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	ctx := context.WithValue(req.Context(), ctxKeyRequestID{}, "recovery-log-id")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	Recovery(logger)(next).ServeHTTP(rec, req)

	entry := decodeLastLogLine(t, &buf)
	assert.Equal(t, "ERROR", entry["level"])
	assert.Equal(t, "recovery-log-id", entry["request_id"])
	assert.Contains(t, entry["panic"], "boom")
	assert.Contains(t, entry, "stack")
}

func TestRecovery_NextDoesNotPanic_PassesThroughUnchanged(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(t, &buf)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	Recovery(logger)(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
	assert.Empty(t, buf.String())
}

func TestRecovery_NoRequestIDInContext_UsesEmptyStringWithoutPanicking(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(t, &buf)

	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		Recovery(logger)(next).ServeHTTP(rec, req)
	})

	var body errorEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "", body.RequestID)
}
