package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver, used for goose migrations and the lock-holding connection below
	"github.com/pressly/goose/v3"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/example"
)

// migrationsDir points at the shared goose migrations directory from this
// package's location (backend/tests/integration -> backend/migrations),
// mirroring internal/example/repository_integration_test.go's constant of
// the same name in its own package (same depth: both packages live two
// directories below backend/).
const migrationsDir = "../../migrations"

// lockHoldStatementTimeoutMillis is the statement_timeout (in milliseconds)
// applied to the app's own Postgres connections for this test only. Short
// enough to keep the test fast, long enough that the app's real startup
// queries (pool warm-up, /healthz ping) aren't spuriously canceled before
// the exclusive lock below is even acquired.
const lockHoldStatementTimeoutMillis = 500

// applyMigrations runs goose migrations against databaseURL so the real
// examples table exists before the app or the lock-holding connection touch
// it. Mirrors internal/example/repository_integration_test.go's
// setupRepositoryTestDB migration step; factored out separately here (rather
// than reused as one function) because this file's setup additionally needs
// to start the app against a *different*, statement_timeout-appended
// connection string, which that helper does not support.
func applyMigrations(t *testing.T, databaseURL string) {
	t.Helper()

	sqlDB, err := sql.Open("pgx", databaseURL)
	require.NoError(t, err, "failed to open database/sql connection for migrations")
	defer sqlDB.Close()

	require.NoError(t, goose.SetDialect("postgres"))
	require.NoError(t, goose.Up(sqlDB, migrationsDir), "failed to apply goose migrations")
}

// createExample POSTs a valid example to the running app and returns the
// created resource's ID, so the DB-timeout scenario below has a real row to
// GET.
func createExample(t *testing.T, baseURL string) string {
	t.Helper()

	payload, err := json.Marshal(map[string]string{
		"name":  "DB Timeout Test Example",
		"email": fmt.Sprintf("db-timeout-test-%d@example.com", time.Now().UnixNano()),
	})
	require.NoError(t, err)

	resp, err := http.Post(baseURL+"/api/v1/examples", "application/json", bytes.NewReader(payload))
	require.NoError(t, err, "POST /api/v1/examples must succeed at the transport level")
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode, "expected the setup POST to succeed")

	var created example.Example
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created), "response body must be valid JSON")
	require.NotEmpty(t, created.ID, "expected the created example to have a generated ID")

	return created.ID
}

// holdExclusiveTableLock opens a database/sql connection independent of the
// app's own connection pool, starts a transaction on it, and locks the
// examples table in ACCESS EXCLUSIVE mode — Postgres's most restrictive lock
// mode, which conflicts with even the plain SELECT the app's GET handler
// issues. The lock is held (the transaction deliberately left open, neither
// committed nor rolled back) until the returned release function runs.
// statement_timeout counts time spent waiting to acquire a lock, so once
// this lock is held, any other connection configured with a
// statement_timeout — including the app's — has its blocked query canceled
// by Postgres once that timeout elapses.
func holdExclusiveTableLock(ctx context.Context, t *testing.T, databaseURL string) (release func()) {
	t.Helper()

	db, err := sql.Open("pgx", databaseURL)
	require.NoError(t, err, "failed to open lock-holding database/sql handle")

	conn, err := db.Conn(ctx)
	require.NoError(t, err, "failed to acquire a dedicated connection for the lock")

	tx, err := conn.BeginTx(ctx, nil)
	require.NoError(t, err, "failed to begin the lock-holding transaction")

	_, err = tx.ExecContext(ctx, "LOCK TABLE examples IN ACCESS EXCLUSIVE MODE")
	require.NoError(t, err, "failed to acquire ACCESS EXCLUSIVE lock on examples")

	return func() {
		if err := tx.Rollback(); err != nil {
			t.Logf("failed to roll back lock-holding transaction: %v", err)
		}
		if err := conn.Close(); err != nil {
			t.Logf("failed to close lock-holding connection: %v", err)
		}
		if err := db.Close(); err != nil {
			t.Logf("failed to close lock-holding database/sql handle: %v", err)
		}
	}
}

// TestGetExample_DBQueryCanceledByStatementTimeout_ReturnsCorrelatedInternalError
// is 005-T118. It exercises a realistic DB-timeout failure — a real
// Postgres-enforced statement_timeout canceling the app's own blocked query
// while waiting on a genuine table lock held by a second connection — rather
// than an artificial short-circuit that would not occur in production.
//
// internal/example/repository.go's scanOne wraps any error that is not
// pgx.ErrNoRows as a plain `fmt.Errorf("find example: %w", err)`, which is
// NOT a *domainerrors.DomainError. errors.HandleError therefore maps it to
// the unmapped "internal_error"/500 path. This test proves that path still
// carries a correlation ID that matches between the response header and the
// structured JSON body, exactly like the happy path does.
//
// The statement_timeout connection parameter approach worked as planned: pgx
// (via pgconn.ParseConfig) forwards any key in the connection string/URL it
// does not recognize as a native libpq option into the connection's startup
// RuntimeParams, which Postgres applies exactly as `SET statement_timeout =
// ...` would for every connection opened with that config — including every
// pooled connection the app's pgxpool later opens, not just the first. No
// fallback (options=-c ...) was needed.
func TestGetExample_DBQueryCanceledByStatementTimeout_ReturnsCorrelatedInternalError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test requiring a container runtime in short mode")
	}

	ctx := context.Background()
	databaseURL := startPostgresContainer(ctx, t)
	applyMigrations(t, databaseURL)

	// Only the app's own connection string gets the short statement_timeout;
	// the migration and lock-holding connections above/below use the plain
	// databaseURL so they are never themselves subject to it.
	appDatabaseURL := fmt.Sprintf("%s&statement_timeout=%d", databaseURL, lockHoldStatementTimeoutMillis)

	binPath := buildAPIBinary(t)
	baseURL := startAPIServer(ctx, t, binPath, appDatabaseURL)

	exampleID := createExample(t, baseURL)

	release := holdExclusiveTableLock(ctx, t, databaseURL)
	defer release()

	// Bounded client-side timeout: if the statement_timeout wiring were ever
	// broken, this fails fast instead of hanging until the surrounding go
	// test timeout.
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/api/v1/examples/" + exampleID)
	require.NoError(t, err,
		"GET must succeed at the transport level; the expected failure is a 500 response body, not a transport error")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode,
		"a statement_timeout-canceled query is a plain wrapped error (not a *domainerrors.DomainError), "+
			"so errors.HandleError must map it to 500 internal_error")

	requestIDHeader := resp.Header.Get("X-Request-ID")
	assert.NotEmpty(t, requestIDHeader, "expected a correlation ID on the response header even for the 500 path")

	var body domainerrors.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body), "error response body must be valid JSON")
	assert.Equal(t, "internal_error", body.Error,
		"an unwrapped, non-DomainError failure must surface as the unmapped internal_error code")
	assert.NotEmpty(t, body.RequestID, "expected the structured error body to carry a request_id")
	assert.Equal(t, requestIDHeader, body.RequestID,
		"the response header's X-Request-ID and the error body's request_id must be the same correlation ID")
}
