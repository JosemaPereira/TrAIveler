// The test build tag matches internal/database/mocks (mockery-generated,
// same tag), keeping mock-dependent tests out of default `go build`/`go
// vet` and requiring `-tags=test` (see Makefile) to compile them.
//go:build test
// +build test

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/config"
	dbmocks "github.com/JosemaPereira/TrAIveler/backend/internal/database/mocks"
)

// testConfig returns a minimal, valid *config.Config for tests that don't
// need config.Load()'s environment-variable plumbing.
func testConfig(t *testing.T) *config.Config {
	t.Helper()

	return &config.Config{
		Server: config.ServerConfig{
			Port:         8080,
			AllowedCORS:  "http://localhost:5173",
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// testLogger returns a *slog.Logger that discards output, keeping test runs
// quiet while still exercising the real Logger/Recovery middleware.
func testLogger(t *testing.T) *slog.Logger {
	t.Helper()

	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestHTTPServer_Healthz_DatabasePingSucceeds_Returns200OkBody(t *testing.T) {
	mockDB := dbmocks.NewMockClient(t)
	mockDB.EXPECT().Ping(mock.Anything).Return(nil)

	srv := NewHTTPServer(mockDB, testConfig(t), testLogger(t))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	srv.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body HealthCheckResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "ok", body.Status)
	assert.NotEmpty(t, body.Version)
	assert.GreaterOrEqual(t, body.UptimeSeconds, float64(0))
}

func TestHTTPServer_Healthz_DatabasePingFails_Returns200DegradedBody(t *testing.T) {
	mockDB := dbmocks.NewMockClient(t)
	mockDB.EXPECT().Ping(mock.Anything).Return(errors.New("connection refused"))

	srv := NewHTTPServer(mockDB, testConfig(t), testLogger(t))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	srv.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "HTTP status must stay 200 even when a dependency is degraded")
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body HealthCheckResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "degraded", body.Status)
	assert.NotEmpty(t, body.Version)
	assert.GreaterOrEqual(t, body.UptimeSeconds, float64(0))
}

func TestHTTPServer_Healthz_DatabasePingExceedsTwoSecondTimeout_Returns200DegradedPromptly(t *testing.T) {
	mockDB := dbmocks.NewMockClient(t)
	mockDB.EXPECT().Ping(mock.Anything).RunAndReturn(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})

	srv := NewHTTPServer(mockDB, testConfig(t), testLogger(t))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	start := time.Now()
	srv.Router().ServeHTTP(rec, req)
	elapsed := time.Since(start)

	assert.Equal(t, http.StatusOK, rec.Code, "HTTP status must stay 200 even when the ping times out")
	assert.Less(t, elapsed, 3*time.Second, "handler must bound the ping to ~2s, not hang indefinitely")

	var body HealthCheckResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "degraded", body.Status)
}

func TestHTTPServer_MiddlewareChain_SetsRequestIDHeaderOnResponse(t *testing.T) {
	mockDB := dbmocks.NewMockClient(t)
	mockDB.EXPECT().Ping(mock.Anything).Return(nil)

	srv := NewHTTPServer(mockDB, testConfig(t), testLogger(t))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	srv.Router().ServeHTTP(rec, req)

	assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
}

func TestHTTPServer_MiddlewareChain_IncomingRequestIDIsEchoedBack(t *testing.T) {
	mockDB := dbmocks.NewMockClient(t)
	mockDB.EXPECT().Ping(mock.Anything).Return(nil)

	srv := NewHTTPServer(mockDB, testConfig(t), testLogger(t))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-ID", "incoming-id-123")
	rec := httptest.NewRecorder()

	srv.Router().ServeHTTP(rec, req)

	assert.Equal(t, "incoming-id-123", rec.Header().Get("X-Request-ID"))
}

func TestHTTPServer_MiddlewareChain_PanicInHandler_Returns500WithoutCrashingProcess(t *testing.T) {
	mockDB := dbmocks.NewMockClient(t)

	srv := NewHTTPServer(mockDB, testConfig(t), testLogger(t))
	srv.router.Get("/panic-test", func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom: simulated handler panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic-test", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		srv.Router().ServeHTTP(rec, req)
	})

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHTTPServer_MiddlewareChain_DisallowedOrigin_NoAccessControlHeaders(t *testing.T) {
	mockDB := dbmocks.NewMockClient(t)
	mockDB.EXPECT().Ping(mock.Anything).Return(nil)

	srv := NewHTTPServer(mockDB, testConfig(t), testLogger(t))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()

	srv.Router().ServeHTTP(rec, req)

	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHTTPServer_MiddlewareChain_RateLimitDisabledByDefault_AllowsRepeatRequests(t *testing.T) {
	mockDB := dbmocks.NewMockClient(t)
	mockDB.EXPECT().Ping(mock.Anything).Return(nil)

	// testConfig leaves RateLimit at its zero value (Requests: 0 == disabled),
	// so the chain must not throttle even under repeated calls from one client.
	srv := NewHTTPServer(mockDB, testConfig(t), testLogger(t))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		srv.Router().ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("X-RateLimit-Limit"),
			"a disabled global rate limit must not set X-RateLimit headers")
	}
}

func TestHTTPServer_MiddlewareChain_RateLimitConfigured_Throttles(t *testing.T) {
	mockDB := dbmocks.NewMockClient(t)
	mockDB.EXPECT().Ping(mock.Anything).Return(nil)

	cfg := testConfig(t)
	cfg.RateLimit = config.RateLimitConfig{Requests: 1, Window: time.Minute}
	srv := NewHTTPServer(mockDB, cfg, testLogger(t))

	first := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	firstRec := httptest.NewRecorder()
	srv.Router().ServeHTTP(firstRec, first)
	assert.Equal(t, http.StatusOK, firstRec.Code)

	second := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	secondRec := httptest.NewRecorder()
	srv.Router().ServeHTTP(secondRec, second)

	assert.Equal(t, http.StatusTooManyRequests, secondRec.Code,
		"a second request from the same client within the window must be throttled")
	assert.NotEmpty(t, secondRec.Header().Get("Retry-After"))
}

func TestHTTPServer_Lifecycle_StartServeShutdownGracefully(t *testing.T) {
	mockDB := dbmocks.NewMockClient(t)
	mockDB.EXPECT().Ping(mock.Anything).Return(nil)

	srv := NewHTTPServer(mockDB, testConfig(t), testLogger(t))

	httpServer := &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: srv.Router(),
	}

	listener, err := net.Listen("tcp", httpServer.Addr)
	require.NoError(t, err)

	serveErrCh := make(chan error, 1)
	go func() {
		serveErrCh <- httpServer.Serve(listener)
	}()

	resp, err := http.Get(fmt.Sprintf("http://%s/healthz", listener.Addr().String()))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	assert.NoError(t, httpServer.Shutdown(ctx))
	assert.ErrorIs(t, <-serveErrCh, http.ErrServerClosed)
}

func TestHTTPServer_Shutdown_WaitsForInFlightRequestToCompleteBeforeReturning(t *testing.T) {
	mockDB := dbmocks.NewMockClient(t)

	srv := NewHTTPServer(mockDB, testConfig(t), testLogger(t))

	requestStarted := make(chan struct{})
	srv.router.Get("/slow", func(w http.ResponseWriter, _ *http.Request) {
		close(requestStarted)
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	httpServer := &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: srv.Router(),
	}

	listener, err := net.Listen("tcp", httpServer.Addr)
	require.NoError(t, err)

	go func() {
		_ = httpServer.Serve(listener)
	}()

	type result struct {
		status int
		err    error
	}
	resultCh := make(chan result, 1)
	go func() {
		resp, err := http.Get(fmt.Sprintf("http://%s/slow", listener.Addr().String()))
		if err != nil {
			resultCh <- result{err: err}
			return
		}
		defer resp.Body.Close()
		resultCh <- result{status: resp.StatusCode}
	}()

	<-requestStarted

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, httpServer.Shutdown(ctx))

	select {
	case res := <-resultCh:
		require.NoError(t, res.err)
		assert.Equal(t, http.StatusOK, res.status)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for in-flight request to complete")
	}
}
