package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Default pool-size values used by tests that don't care about a specific
// configuration, mirroring config.DatabaseConfig's documented defaults
// (DB_MIN_CONNECTIONS=5, DB_MAX_CONNECTIONS=25).
const (
	testDefaultMinConns = 5
	testDefaultMaxConns = 25
)

// setupPostgresContainer is a test helper that starts a PostgreSQL testcontainer
// and returns its connection string for use in integration tests.
// The container uses postgres:16-alpine with test credentials (testuser/testpass/testdb).
// Waits for PostgreSQL to be fully ready before returning (2 occurrences of "ready" message).
//
// The container is terminated automatically via t.Cleanup when the test ends.
func setupPostgresContainer(t *testing.T, ctx context.Context) string {
	t.Helper()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err, "failed to start postgres container")
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	return connStr
}

func TestNewClient_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	connStr := setupPostgresContainer(t, ctx)

	client, err := NewClient(ctx, connStr, testDefaultMinConns, testDefaultMaxConns)

	require.NoError(t, err, "NewClient should succeed with valid connection string")
	require.NotNil(t, client, "client should not be nil")
	defer client.Close()
	pool := client.Pool()
	assert.NotNil(t, pool, "pool should be initialized")
	stats := pool.Stat()
	assert.GreaterOrEqual(t, stats.TotalConns(), int32(5), "should have at least MinConns (5) connections")
	assert.LessOrEqual(t, stats.MaxConns(), int32(25), "MaxConns should be 25")
}

// TestNewClient_ConfigurablePoolSize asserts that NewClient applies the
// min/max connection values it is given, rather than a hardcoded pool size.
// Uses non-default values (2/10) so the assertion cannot pass by coincidence
// against the old hardcoded 5/25 behavior.
func TestNewClient_ConfigurablePoolSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	connStr := setupPostgresContainer(t, ctx)

	const wantMinConns = 2
	const wantMaxConns = 10

	client, err := NewClient(ctx, connStr, wantMinConns, wantMaxConns)

	require.NoError(t, err, "NewClient should succeed with valid connection string")
	require.NotNil(t, client, "client should not be nil")
	defer client.Close()
	poolConfig := client.Pool().Config()
	assert.Equal(t, int32(wantMinConns), poolConfig.MinConns, "pool should apply the caller-provided MinConns")
	assert.Equal(t, int32(wantMaxConns), poolConfig.MaxConns, "pool should apply the caller-provided MaxConns")
}

func TestNewClient_InvalidURL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		url  string
	}{
		{name: "when the URL is empty it should return an error", url: ""},
		{name: "when the URL is malformed it should return an error", url: "invalid-url"},
		{
			name: "when the host does not exist it should return an error",
			url:  "postgresql://invalid:invalid@nonexistent:5432/db?sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(ctx, tt.url, testDefaultMinConns, testDefaultMaxConns)

			assert.Error(t, err, "NewClient should fail with invalid URL")
			assert.Nil(t, client, "client should be nil on error")
		})
	}
}

func TestNewClient_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client, err := NewClient(ctx, "postgresql://user:pass@localhost:5432/db?sslmode=disable", testDefaultMinConns, testDefaultMaxConns)

	assert.Error(t, err, "NewClient should fail with canceled context")
	assert.Nil(t, client, "client should be nil on error")
}

func TestPing_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	connStr := setupPostgresContainer(t, ctx)
	client, err := NewClient(ctx, connStr, testDefaultMinConns, testDefaultMaxConns)
	require.NoError(t, err)
	defer client.Close()

	err = client.Ping(ctx)

	assert.NoError(t, err, "Ping should succeed with healthy connection")
}

func TestPing_ClosedConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	connStr := setupPostgresContainer(t, ctx)
	client, err := NewClient(ctx, connStr, testDefaultMinConns, testDefaultMaxConns)
	require.NoError(t, err)

	err = client.Close()
	require.NoError(t, err)

	err = client.Ping(ctx)
	assert.Error(t, err, "Ping should fail after Close")
}

func TestPing_ContextTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	connStr := setupPostgresContainer(t, ctx)
	client, err := NewClient(ctx, connStr, testDefaultMinConns, testDefaultMaxConns)
	require.NoError(t, err)
	defer client.Close()
	pingCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond) // Ensure timeout expires

	err = client.Ping(pingCtx)

	assert.Error(t, err, "Ping should fail with expired context")
}

func TestClose_GracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	connStr := setupPostgresContainer(t, ctx)
	client, err := NewClient(ctx, connStr, testDefaultMinConns, testDefaultMaxConns)
	require.NoError(t, err)
	stats := client.Pool().Stat()
	initialConns := stats.TotalConns()
	assert.Greater(t, initialConns, int32(0), "should have active connections")

	err = client.Close()

	assert.NoError(t, err, "Close should succeed")
	err = client.Ping(ctx)
	assert.Error(t, err, "Ping should fail after Close")
	assert.Contains(t, err.Error(), "closed", "error should mention closed pool")
}

func TestConnectionPoolLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	connStr := setupPostgresContainer(t, ctx)

	client, err := NewClient(ctx, connStr, testDefaultMinConns, testDefaultMaxConns)

	require.NoError(t, err)
	defer client.Close()
	stats := client.Pool().Stat()
	assert.Equal(t, int32(25), stats.MaxConns(), "MaxConns should be 25")
	assert.GreaterOrEqual(t, stats.TotalConns(), int32(5), "should have at least MinConns (5)")
}

func TestNewClient_RetryLogic(t *testing.T) {
	ctx := context.Background()
	invalidURL := "postgresql://postgres:postgres@localhost:9999/db?sslmode=disable"

	start := time.Now()
	client, err := NewClient(ctx, invalidURL, testDefaultMinConns, testDefaultMaxConns)
	elapsed := time.Since(start)

	assert.Error(t, err, "NewClient should fail after retries")
	assert.Nil(t, client, "client should be nil on error")
	// 3 retries * 2s delay = 4s minimum; 3.5s leaves margin for execution overhead.
	assert.GreaterOrEqual(t, elapsed.Seconds(), 3.5, "should retry with delays")
}

func TestPool_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	connStr := setupPostgresContainer(t, ctx)
	client, err := NewClient(ctx, connStr, testDefaultMinConns, testDefaultMaxConns)
	require.NoError(t, err)
	defer client.Close()
	const numGoroutines = 50
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			errChan <- client.Ping(ctx)
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		err := <-errChan
		assert.NoError(t, err, "concurrent Ping should succeed")
	}
}

func TestPool_AcquireReleaseCycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	connStr := setupPostgresContainer(t, ctx)
	client, err := NewClient(ctx, connStr, testDefaultMinConns, testDefaultMaxConns)
	require.NoError(t, err)
	defer client.Close()

	conn, err := client.Pool().Acquire(ctx)
	require.NoError(t, err, "should acquire connection")
	require.NotNil(t, conn, "connection should not be nil")
	var result int
	err = conn.QueryRow(ctx, "SELECT 1").Scan(&result)
	assert.NoError(t, err, "query should succeed")
	assert.Equal(t, 1, result, "should get expected result")
	conn.Release()

	err = client.Ping(ctx)
	assert.NoError(t, err, "pool should remain healthy after release")
}
