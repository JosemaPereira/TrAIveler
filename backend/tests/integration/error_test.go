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

	"github.com/JosemaPereira/TrAIveler/backend/internal/database/migrations"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/example"
)

// lockHoldStatementTimeoutMillis is the statement_timeout (ms) applied only
// to the app's own Postgres connections: short enough to keep the test fast,
// long enough that the app's startup queries aren't canceled before the
// exclusive lock below is even acquired.
const lockHoldStatementTimeoutMillis = 500

// applyMigrations runs goose migrations against databaseURL. Not reused from
// repository_integration_test.go's setupRepositoryTestDB because this test
// also needs to start the app against a separate, statement_timeout-appended
// connection string, which that helper doesn't support.
func applyMigrations(t *testing.T, databaseURL string) {
	t.Helper()

	sqlDB, err := sql.Open("pgx", databaseURL)
	require.NoError(t, err, "failed to open database/sql connection for migrations")
	defer sqlDB.Close()

	require.NoError(t, migrations.SetDialect())
	require.NoError(t, goose.Up(sqlDB, migrations.Dir), "failed to apply goose migrations")
}

// createExample POSTs a row via the running app and returns its ID, so the
// DB-timeout scenario below has something real to GET.
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

// holdExclusiveTableLock locks the examples table in ACCESS EXCLUSIVE mode
// on a connection independent of the app's pool — blocking even a plain
// SELECT — until the returned release func runs. statement_timeout counts
// time spent waiting on a lock, so once this is held, any connection with a
// statement_timeout (including the app's) gets its blocked query canceled
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
// (005-T118) uses a real Postgres-canceled query, not a short-circuited one,
// to prove the resulting 500 internal_error carries a correlation ID
// consistent between the response header and the structured error body.
// scanOne wraps any non-ErrNoRows failure as a plain error (not a
// *domainerrors.DomainError), so HandleError maps it to the unmapped
// internal_error/500 path.
func TestGetExample_DBQueryCanceledByStatementTimeout_ReturnsCorrelatedInternalError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test requiring a container runtime in short mode")
	}

	ctx := context.Background()
	databaseURL := startPostgresContainer(ctx, t)
	applyMigrations(t, databaseURL)

	// Only the app gets the short statement_timeout; pgx forwards unrecognized
	// connection-string keys into the connection's startup parameters, so this
	// applies to every pooled connection the app opens, not just the first.
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
