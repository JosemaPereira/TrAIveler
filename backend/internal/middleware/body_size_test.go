package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const oneMB = 1024 * 1024

func TestBodySize_UnderLimit_PassesThroughToNext(t *testing.T) {
	payload := strings.Repeat("a", oneMB)
	var readBody string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		readBody = string(b)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/trips", strings.NewReader(payload))
	rec := httptest.NewRecorder()

	BodySize(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, payload, readBody)
}

func TestBodySize_ContentLengthExceedsLimit_Returns413AndSkipsNext(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	req := httptest.NewRequest(http.MethodPost, "/trips", strings.NewReader("small"))
	req.ContentLength = 11 * oneMB
	rec := httptest.NewRecorder()

	BodySize(next).ServeHTTP(rec, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)

	var body errorEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "payload_too_large", body.Error)
	assert.Equal(t, "Request body exceeds the 10 MB limit", body.Message)
}

func TestBodySize_ActualBodyExceedsLimitWithUnknownContentLength_Returns413(t *testing.T) {
	oversized := bytes.Repeat([]byte("a"), 11*oneMB)
	// io.MultiReader hides the underlying type from net/http's ContentLength
	// sniffing, simulating a chunked-transfer request with no declared length.
	body := io.MultiReader(bytes.NewReader(oversized))

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
	})

	req := httptest.NewRequest(http.MethodPost, "/trips", body)
	require.LessOrEqual(t, req.ContentLength, int64(10*oneMB))
	rec := httptest.NewRecorder()

	BodySize(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)

	var envelope errorEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	assert.Equal(t, "payload_too_large", envelope.Error)
	assert.Equal(t, "Request body exceeds the 10 MB limit", envelope.Message)
}

func TestBodySize_RequestIDInContext_IncludesItInEnvelope(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest(http.MethodPost, "/trips", strings.NewReader("small"))
	req.ContentLength = 11 * oneMB
	req.Header.Set("X-Request-ID", "body-size-test-id")
	rec := httptest.NewRecorder()

	RequestID(BodySize(next)).ServeHTTP(rec, req)

	var envelope errorEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	assert.Equal(t, "body-size-test-id", envelope.RequestID)
}
