package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// healthCheckResponse mirrors cmd/api/server.go's HealthCheckResponse wire
// shape. Duplicated here rather than imported because cmd/api is `package
// main` and cannot be imported from any other package — see this package's
// doc comment in swagger_test.go for the same constraint applied there.
type healthCheckResponse struct {
	Status        string  `json:"status"`
	Version       string  `json:"version"`
	UptimeSeconds float64 `json:"uptime_seconds"`
}

// TestHealthz_NoIncomingRequestID_ReturnsOKWithGeneratedCorrelationID is
// 005-T117. It asserts GET /healthz against a real running backend process
// (real Chi router, real middleware.RequestID, real database.Client backed
// by a genuinely healthy Postgres testcontainer):
//   - always returns HTTP 200, per spec 002's validation rule that HTTP
//     status never reflects DB health (only the body's "status" field does);
//   - carries a non-empty X-Request-ID response header, proving
//     middleware.RequestID actually ran ahead of the handler;
//   - decodes into the real HealthCheckResponse shape with status "ok",
//     since the container backing this test is healthy.
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

// TestHealthz_WithIncomingRequestIDHeader_EchoesSameValueBack is also
// 005-T117. It proves genuine correlation-ID propagation through the real
// middleware chain — not merely "a header exists" — by sending a request
// with a pre-set X-Request-ID header and asserting the response echoes back
// that exact same value, per middleware/request_id.go's documented
// "reuse an incoming header when present" behavior.
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
