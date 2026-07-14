package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// healthCheckResponse mirrors cmd/api/server.go's HealthCheckResponse;
// duplicated because cmd/api is `package main` and can't be imported.
type healthCheckResponse struct {
	Status        string  `json:"status"`
	Version       string  `json:"version"`
	UptimeSeconds float64 `json:"uptime_seconds"`
}

// TestHealthz_NoIncomingRequestID_ReturnsOKWithGeneratedCorrelationID (005-T117)
// asserts GET /healthz always returns 200 with a generated X-Request-ID and a
// healthy body, against a real backend + Postgres testcontainer.
func TestHealthz_NoIncomingRequestID_ReturnsOKWithGeneratedCorrelationID(t *testing.T) {
	baseURL := setupSwaggerTestServer(t)

	resp, err := http.Get(baseURL + "/healthz")
	require.NoError(t, err, "GET /healthz must succeed at the transport level")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode,
		"per spec 002's validation rules /healthz must always return 200, regardless of DB health")

	requestID := resp.Header.Get("X-Request-ID")
	assert.NotEmpty(t, requestID,
		"expected middleware.RequestID to set a generated X-Request-ID response header")

	var body healthCheckResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body), "response body must be valid JSON")
	assert.Equal(t, "ok", body.Status,
		"the testcontainer backing this test is healthy, so status must be \"ok\", not \"degraded\"")
	assert.NotEmpty(t, body.Version, "expected a non-empty version string")
	assert.GreaterOrEqual(t, body.UptimeSeconds, 0.0, "expected a non-negative uptime")
}

// TestHealthz_WithIncomingRequestIDHeader_EchoesSameValueBack (005-T117)
// proves middleware.RequestID reuses an incoming X-Request-ID rather than
// overwriting it with a freshly generated one.
func TestHealthz_WithIncomingRequestIDHeader_EchoesSameValueBack(t *testing.T) {
	baseURL := setupSwaggerTestServer(t)

	const incomingRequestID = "test-correlation-id-12345"
	req, err := http.NewRequest(http.MethodGet, baseURL+"/healthz", nil)
	require.NoError(t, err, "failed to build request")
	req.Header.Set("X-Request-ID", incomingRequestID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "GET /healthz must succeed at the transport level")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, incomingRequestID, resp.Header.Get("X-Request-ID"),
		"middleware.RequestID must reuse an incoming X-Request-ID header rather than "+
			"overwrite it with a freshly generated UUID")
}
