package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID_NoIncomingHeader_GeneratesAndSetsResponseHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	RequestID(next).ServeHTTP(rec, req)

	got := rec.Header().Get("X-Request-ID")
	assert.NotEmpty(t, got)
}

func TestRequestID_IncomingHeader_ReusesValueVerbatim(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "incoming-fixed-id")
	rec := httptest.NewRecorder()

	RequestID(next).ServeHTTP(rec, req)

	assert.Equal(t, "incoming-fixed-id", rec.Header().Get("X-Request-ID"))
}

func TestRequestID_ContextCarriesValue_InsideNext(t *testing.T) {
	var idFromContext string
	var idPresent bool

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idFromContext, idPresent = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "context-check-id")
	rec := httptest.NewRecorder()

	RequestID(next).ServeHTTP(rec, req)

	require.True(t, idPresent)
	assert.Equal(t, "context-check-id", idFromContext)
}

func TestRequestIDFromContext_NoValueSet_ReturnsFalse(t *testing.T) {
	_, ok := RequestIDFromContext(httptest.NewRequest(http.MethodGet, "/", nil).Context())

	assert.False(t, ok)
}
