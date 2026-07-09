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

// TestNewClient_Success verifies that a client can be created with a valid database URL
func TestNewClient_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, connStr := setupPostgresContainer(ctx, t)
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	// Create client
	client, err := NewClient(ctx, connStr)
	require.NoError(t, err, "NewClient should succeed with valid connection string")
	require.NotNil(t, client, "client should not be nil")
	defer client.Close()

	// Verify pool is initialized and accessible
	pool := client.Pool()
	assert.NotNil(t, pool, "pool should be initialized")

	// Verify connection pool stats
	stats := pool.Stat()
	assert.GreaterOrEqual(t, stats.TotalConns(), int32(5), "should have at least MinConns (5) connections")
	assert.LessOrEqual(t, stats.MaxConns(), int32(25), "MaxConns should be 25")
}

// TestNewClient_InvalidURL verifies that NewClient fails with an invalid database URL
func TestNewClient_InvalidURL(t *testing.T) {
	ctx := context.Background()

	invalidURLs := []string{
		"",
		"invalid-url",
		"postgresql://invalid:invalid@nonexistent:5432/db?sslmode=disable",
	}

	for _, url := range invalidURLs {
		t.Run(url, func(t *testing.T) {
			client, err := NewClient(ctx, url)
			assert.Error(t, err, "NewClient should fail with invalid URL")
			assert.Nil(t, client, "client should be nil on error")
		})
	}
}

// TestNewClient_ContextCancellation verifies that NewClient respects context cancellation
func TestNewClient_ContextCancellation(t *testing.T) {
	// Create a context that's already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client, err := NewClient(ctx, "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	assert.Error(t, err, "NewClient should fail with canceled context")
	assert.Nil(t, client, "client should be nil on error")
}

// TestPing_Success verifies that Ping succeeds with a healthy connection
func TestPing_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()

	pgContainer, connStr := setupPostgresContainer(ctx, t)
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	client, err := NewClient(ctx, connStr)
	require.NoError(t, err)
	defer client.Close()

	// Ping should succeed
	err = client.Ping(ctx)
	assert.NoError(t, err, "Ping should succeed with healthy connection")
}

// TestPing_ClosedConnection verifies that Ping fails after Close
func TestPing_ClosedConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()

	pgContainer, connStr := setupPostgresContainer(ctx, t)
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	client, err := NewClient(ctx, connStr)
	require.NoError(t, err)

	// Close the client
	err = client.Close()
	require.NoError(t, err)

	// Ping should fail after close
	err = client.Ping(ctx)
	assert.Error(t, err, "Ping should fail after Close")
}

// TestPing_ContextTimeout verifies that Ping respects context timeout
func TestPing_ContextTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()

	pgContainer, connStr := setupPostgresContainer(ctx, t)
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	client, err := NewClient(ctx, connStr)
	require.NoError(t, err)
	defer client.Close()

	// Create context with very short timeout
	pingCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout expires

	err = client.Ping(pingCtx)
	assert.Error(t, err, "Ping should fail with expired context")
}

// TestClose_GracefulShutdown verifies that Close properly shuts down the pool
func TestClose_GracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()

	pgContainer, connStr := setupPostgresContainer(ctx, t)
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	client, err := NewClient(ctx, connStr)
	require.NoError(t, err)

	// Get initial connection count
	stats := client.Pool().Stat()
	initialConns := stats.TotalConns()
	assert.Greater(t, initialConns, int32(0), "should have active connections")

	// Close should succeed
	err = client.Close()
	assert.NoError(t, err, "Close should succeed")

	// Ping should fail after close
	err = client.Ping(ctx)
	assert.Error(t, err, "Ping should fail after Close")
	assert.Contains(t, err.Error(), "closed", "error should mention closed pool")
}

// TestConnectionPoolLimits verifies that connection pool respects min/max limits
func TestConnectionPoolLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()

	pgContainer, connStr := setupPostgresContainer(ctx, t)
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	client, err := NewClient(ctx, connStr)
	require.NoError(t, err)
	defer client.Close()

	// Check pool configuration
	stats := client.Pool().Stat()
	assert.Equal(t, int32(25), stats.MaxConns(), "MaxConns should be 25")

	// Verify minimum connections are created
	assert.GreaterOrEqual(t, stats.TotalConns(), int32(5), "should have at least MinConns (5)")
}

// TestNewClient_RetryLogic verifies retry behavior on connection failure
func TestNewClient_RetryLogic(t *testing.T) {
	ctx := context.Background()

	// Use invalid port to trigger connection failure
	invalidURL := "postgresql://postgres:postgres@localhost:9999/db?sslmode=disable"

	start := time.Now()
	client, err := NewClient(ctx, invalidURL)
	elapsed := time.Since(start)

	// Should fail after retries
	assert.Error(t, err, "NewClient should fail after retries")
	assert.Nil(t, client, "client should be nil on error")

	// Should have taken at least 4 seconds (3 retries * 2 seconds between attempts)
	// We allow some margin for execution overhead
	assert.GreaterOrEqual(t, elapsed.Seconds(), 3.5, "should retry with delays")
}

// setupPostgresContainer is a test helper that starts a PostgreSQL testcontainer.
// It returns the container instance and connection string for use in integration tests.
// The container uses postgres:16-alpine with test credentials (testuser/testpass/testdb).
// Waits for PostgreSQL to be fully ready before returning (2 occurrences of "ready" message).
//
// The caller is responsible for terminating the container in a defer statement.
func setupPostgresContainer(ctx context.Context, t *testing.T) (*postgres.PostgresContainer, string) {
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

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	return pgContainer, connStr
}

// TestPool_ConcurrentOperations verifies thread-safety of pool operations
func TestPool_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()

	pgContainer, connStr := setupPostgresContainer(ctx, t)
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	client, err := NewClient(ctx, connStr)
	require.NoError(t, err)
	defer client.Close()

	// Run 50 concurrent Ping operations
	const numGoroutines = 50
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			errChan <- client.Ping(ctx)
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-errChan
		assert.NoError(t, err, "concurrent Ping should succeed")
	}
}

// TestPool_AcquireReleaseCycle verifies connection acquisition and release
func TestPool_AcquireReleaseCycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()

	pgContainer, connStr := setupPostgresContainer(ctx, t)
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	client, err := NewClient(ctx, connStr)
	require.NoError(t, err)
	defer client.Close()

	// Acquire connection
	conn, err := client.Pool().Acquire(ctx)
	require.NoError(t, err, "should acquire connection")
	require.NotNil(t, conn, "connection should not be nil")

	// Use connection
	var result int
	err = conn.QueryRow(ctx, "SELECT 1").Scan(&result)
	assert.NoError(t, err, "query should succeed")
	assert.Equal(t, 1, result, "should get expected result")

	// Release connection
	conn.Release()

	// Pool should still be healthy
	err = client.Ping(ctx)
	assert.NoError(t, err, "pool should remain healthy after release")
}
