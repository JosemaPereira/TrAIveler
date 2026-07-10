package errors_test

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withCapturedDefaultLogger temporarily replaces slog's global default logger
// with one writing to buf, restoring the original when the test ends. This is
// only needed because HandleError's signature (w, r, err) has no logger
// parameter to inject, unlike the rest of this codebase's constructor-injected
// loggers (see middleware.Logger/Recovery).
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

// requestWithID builds a request whose context carries requestID, using the
// real middleware.RequestID chain (its context key is unexported and cannot
// be set directly from outside the middleware package).
func requestWithID(method, target, requestID string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	if requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}

	var captured *http.Request
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = r
	})
	middleware.RequestID(next).ServeHTTP(httptest.NewRecorder(), req)
	return captured
}

func TestHandleError_DomainErrorCodes_MapToExpectedStatus(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"not found", domainerrors.NotFound("trip", "trip-123"), http.StatusNotFound, "not_found"},
		{"validation failed", domainerrors.Validation("bad input"), http.StatusUnprocessableEntity, "validation_failed"},
		{"unauthorized", domainerrors.Unauthorized("no token"), http.StatusUnauthorized, "authentication_required"},
		{"forbidden", domainerrors.Forbidden("not allowed"), http.StatusForbidden, "forbidden"},
		{"conflict", domainerrors.Conflict("version mismatch"), http.StatusConflict, "conflict"},
		{"unmapped plain error falls back to internal_error", stderrors.New("boom"), http.StatusInternalServerError, "internal_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withCapturedDefaultLogger(t)

			rec := httptest.NewRecorder()
			req := requestWithID(http.MethodGet, "/trips/trip-123", "")

			domainerrors.HandleError(rec, req, tt.err)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			var body map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, tt.wantCode, body["error"])
		})
	}
}

func TestHandleError_ValidationError_IncludesFieldsInResponse(t *testing.T) {
	withCapturedDefaultLogger(t)

	err := domainerrors.Validation("One or more fields failed validation",
		domainerrors.ValidationError{Field: "email", Error: "Email address is already registered"},
	)

	rec := httptest.NewRecorder()
	req := requestWithID(http.MethodPost, "/trips", "")

	domainerrors.HandleError(rec, req, err)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	fields, ok := body["fields"].([]any)
	require.True(t, ok, "expected fields array in response body")
	require.Len(t, fields, 1)

	field, ok := fields[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "email", field["field"])
	assert.Equal(t, "Email address is already registered", field["error"])
}

func TestHandleError_NonValidationError_OmitsFieldsFromResponse(t *testing.T) {
	withCapturedDefaultLogger(t)

	rec := httptest.NewRecorder()
	req := requestWithID(http.MethodGet, "/trips/trip-123", "")

	domainerrors.HandleError(rec, req, domainerrors.NotFound("trip", "trip-123"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	_, present := body["fields"]
	assert.False(t, present, "fields key must be absent for non-validation errors")
}

func TestHandleError_NonDomainError_DoesNotLeakRawMessageToClient(t *testing.T) {
	withCapturedDefaultLogger(t)

	rec := httptest.NewRecorder()
	req := requestWithID(http.MethodGet, "/trips/trip-123", "")

	domainerrors.HandleError(rec, req, stderrors.New("pq: connection refused to db_password=secret host"))

	assert.NotContains(t, rec.Body.String(), "db_password")
	assert.NotContains(t, rec.Body.String(), "connection refused")

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "internal_error", body["error"])
	assert.NotEmpty(t, body["message"])
}

func TestHandleError_NonDomainError_LogsRawErrorViaDefaultLogger(t *testing.T) {
	buf := withCapturedDefaultLogger(t)

	rec := httptest.NewRecorder()
	req := requestWithID(http.MethodGet, "/trips/trip-123", "logging-test-id")

	domainerrors.HandleError(rec, req, stderrors.New("pq: connection refused"))

	logged := buf.String()
	assert.Contains(t, logged, "connection refused")
	assert.Contains(t, logged, "logging-test-id")
}

func TestHandleError_WrappedDomainError_StillResolvesViaErrorsAs(t *testing.T) {
	withCapturedDefaultLogger(t)

	wrapped := fmt.Errorf("repository lookup failed: %w", domainerrors.NotFound("trip", "trip-123"))

	rec := httptest.NewRecorder()
	req := requestWithID(http.MethodGet, "/trips/trip-123", "")

	domainerrors.HandleError(rec, req, wrapped)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "not_found", body["error"])
}

func TestHandleError_RequestIDPresentInContext_PopulatesResponseField(t *testing.T) {
	withCapturedDefaultLogger(t)

	rec := httptest.NewRecorder()
	req := requestWithID(http.MethodGet, "/trips/trip-123", "req-abc123")

	domainerrors.HandleError(rec, req, domainerrors.NotFound("trip", "trip-123"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "req-abc123", body["request_id"])
}

func TestHandleError_RequestIDAbsentFromContext_UsesEmptyString(t *testing.T) {
	withCapturedDefaultLogger(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/trips/trip-123", nil)

	assert.NotPanics(t, func() {
		domainerrors.HandleError(rec, req, domainerrors.NotFound("trip", "trip-123"))
	})

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "", body["request_id"])
}

func TestHandleError_ResponseBody_IncludesMessageField(t *testing.T) {
	withCapturedDefaultLogger(t)

	rec := httptest.NewRecorder()
	req := requestWithID(http.MethodGet, "/trips/trip-123", "")

	domainerrors.HandleError(rec, req, domainerrors.Conflict("version mismatch"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "version mismatch", body["message"])
}
