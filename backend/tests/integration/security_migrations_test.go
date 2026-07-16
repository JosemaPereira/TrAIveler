// Package integration: security_migrations_test.go verifies the Spec 004
// security/auth data model migrations (004-T006 to 004-T009, issue #138).
// This is a migrations-only slice: no Go domain code exists yet for
// users/refresh_tokens/jwt_signing_keys/security_events, so the only way to
// verify the schema is by applying goose migrations against a real Postgres
// testcontainer and inspecting information_schema directly.
//
// 004-T010 (trips.version) is no longer deferred: as of issue #142
// (G-008-MIGRATIONS) it lands inside the trips CREATE migration (010), and is
// covered by spec008_migrations_test.go in this same package. 004-T011
// (itinerary_items.version) STAYS deferred — no migration in either spec's
// Sprint 5 set creates an `itinerary_items` table, so its version column
// cannot be added yet. An earlier version of this file included placeholder
// trips/itinerary_items migrations plus a test standing up minimal tables to
// validate them in isolation; that empirically proved goose.Up applies
// migrations in strict version order and stops at the first failure, so
// shipping a migration that references a not-yet-created table would break
// every goose.Up(sqlDB, migrations.Dir) call in the repo — including
// internal/example's and tests/integration/error_test.go's — not just this
// package's own assertions. That regression guard is why 004-T011 remains
// held until a migration creates `itinerary_items`; see docs/roadmap.md row
// 004-T011 for the tracking note.
package integration

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver, used only to run goose migrations
	"github.com/pressly/goose/v3"

	"github.com/JosemaPereira/TrAIveler/backend/internal/database/migrations"
)

// coreVersion is the goose version of the last Phase-1/2/3 security table
// migration (004-T006 to 004-T009: users, refresh_tokens, jwt_signing_keys,
// security_events). Stopping goose UpTo/DownTo here deliberately avoids
// touching the pre-existing examples-table migration (version
// 20260710120000, which sorts after these).
const coreVersion = 4

// openMigrationDB opens a database/sql handle against a fresh postgres
// testcontainer and configures goose's dialect, without applying any
// migrations. Reuses startPostgresContainer (swagger_test.go) rather than
// duplicating the container setup a third time in this package.
func openMigrationDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	connStr := startPostgresContainer(t, ctx)

	sqlDB, err := sql.Open("pgx", connStr)
	require.NoError(t, err, "failed to open database/sql connection for migrations")
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, migrations.SetDialect())
	return sqlDB
}

// columnInfo mirrors the subset of information_schema.columns this test
// needs to assert on.
type columnInfo struct {
	dataType   string
	isNullable string
	defaultVal sql.NullString
}

func queryColumn(t *testing.T, db *sql.DB, table, column string) (columnInfo, bool) {
	t.Helper()

	var info columnInfo
	row := db.QueryRow(
		`SELECT data_type, is_nullable, column_default
		 FROM information_schema.columns
		 WHERE table_name = $1 AND column_name = $2`,
		table, column,
	)
	err := row.Scan(&info.dataType, &info.isNullable, &info.defaultVal)
	if err == sql.ErrNoRows {
		return columnInfo{}, false
	}
	require.NoError(t, err, "failed to query information_schema.columns for %s.%s", table, column)
	return info, true
}

func indexExists(t *testing.T, db *sql.DB, indexName string) bool {
	t.Helper()

	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = $1)`,
		indexName,
	).Scan(&exists)
	require.NoError(t, err, "failed to query pg_indexes for %s", indexName)
	return exists
}

func checkConstraintExists(t *testing.T, db *sql.DB, table, checkFragment string) bool {
	t.Helper()

	var count int
	err := db.QueryRow(
		`SELECT COUNT(*)
		 FROM information_schema.table_constraints tc
		 JOIN information_schema.check_constraints cc
		   ON tc.constraint_name = cc.constraint_name
		   AND tc.constraint_schema = cc.constraint_schema
		 WHERE tc.table_name = $1 AND cc.check_clause LIKE $2`,
		table, "%"+checkFragment+"%",
	).Scan(&count)
	require.NoError(t, err, "failed to query check constraints for %s", table)
	return count > 0
}

// TestSecurityMigrations_CoreTables_CreatesUsersRefreshTokensJWTKeysSecurityEvents
// (004-T006 to 004-T009) applies migrations 001-004 and asserts the users,
// refresh_tokens, jwt_signing_keys, and security_events tables exist with the
// exact columns/constraints/indexes specified in
// specs/004-security-auth-model/data-model.md's "Migration Strategy" section.
func TestSecurityMigrations_CoreTables_CreatesUsersRefreshTokensJWTKeysSecurityEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := openMigrationDB(t, ctx)

	require.NoError(t, goose.UpTo(db, migrations.Dir, coreVersion), "failed to apply migrations up to version %d", coreVersion)

	t.Run("when core migrations apply the users table should match the spec", func(t *testing.T) {
		info, ok := queryColumn(t, db, "users", "id")
		require.True(t, ok, "expected users.id column to exist")
		assert.Equal(t, "uuid", info.dataType)
		assert.Equal(t, "NO", info.isNullable)

		info, ok = queryColumn(t, db, "users", "email")
		require.True(t, ok, "expected users.email column to exist")
		assert.Equal(t, "character varying", info.dataType)
		assert.Equal(t, "NO", info.isNullable)

		info, ok = queryColumn(t, db, "users", "password_hash")
		require.True(t, ok, "expected users.password_hash column to exist")
		assert.Equal(t, "NO", info.isNullable)

		info, ok = queryColumn(t, db, "users", "role")
		require.True(t, ok, "expected users.role column to exist")
		assert.Equal(t, "NO", info.isNullable)
		assert.Contains(t, info.defaultVal.String, "admin")
		assert.True(t, checkConstraintExists(t, db, "users", "role"), "expected a CHECK constraint on users.role")

		_, ok = queryColumn(t, db, "users", "last_login_at")
		assert.True(t, ok, "expected users.last_login_at column to exist")

		assert.True(t, indexExists(t, db, "idx_users_email"), "expected idx_users_email index to exist")
	})

	t.Run("when core migrations apply the refresh_tokens table should match the spec", func(t *testing.T) {
		info, ok := queryColumn(t, db, "refresh_tokens", "user_id")
		require.True(t, ok, "expected refresh_tokens.user_id column to exist")
		assert.Equal(t, "uuid", info.dataType)
		assert.Equal(t, "NO", info.isNullable)

		info, ok = queryColumn(t, db, "refresh_tokens", "token_hash")
		require.True(t, ok, "expected refresh_tokens.token_hash column to exist")
		assert.Equal(t, "NO", info.isNullable)

		_, ok = queryColumn(t, db, "refresh_tokens", "expires_at")
		assert.True(t, ok, "expected refresh_tokens.expires_at column to exist")

		_, ok = queryColumn(t, db, "refresh_tokens", "revoked_at")
		assert.True(t, ok, "expected refresh_tokens.revoked_at column to exist")

		assert.True(t, indexExists(t, db, "idx_refresh_tokens_token_hash"))
		assert.True(t, indexExists(t, db, "idx_refresh_tokens_user_id"))
		assert.True(t, indexExists(t, db, "idx_refresh_tokens_expires_at"))
	})

	t.Run("when core migrations apply the jwt_signing_keys table should match the spec", func(t *testing.T) {
		info, ok := queryColumn(t, db, "jwt_signing_keys", "key_id")
		require.True(t, ok, "expected jwt_signing_keys.key_id column to exist")
		assert.Equal(t, "character varying", info.dataType)

		info, ok = queryColumn(t, db, "jwt_signing_keys", "public_key")
		require.True(t, ok, "expected jwt_signing_keys.public_key column to exist")
		assert.Equal(t, "text", info.dataType)
		assert.Equal(t, "NO", info.isNullable)

		info, ok = queryColumn(t, db, "jwt_signing_keys", "private_key_secret_arn")
		require.True(t, ok, "expected jwt_signing_keys.private_key_secret_arn column to exist")
		assert.Equal(t, "NO", info.isNullable)

		info, ok = queryColumn(t, db, "jwt_signing_keys", "status")
		require.True(t, ok, "expected jwt_signing_keys.status column to exist")
		assert.Contains(t, info.defaultVal.String, "active")
		assert.True(t, checkConstraintExists(t, db, "jwt_signing_keys", "status"))

		_, ok = queryColumn(t, db, "jwt_signing_keys", "retire_at")
		assert.True(t, ok, "expected jwt_signing_keys.retire_at column to exist")
	})

	t.Run("when core migrations apply the security_events table should match the spec", func(t *testing.T) {
		info, ok := queryColumn(t, db, "security_events", "correlation_id")
		require.True(t, ok, "expected security_events.correlation_id column to exist")
		assert.Equal(t, "uuid", info.dataType)
		assert.Equal(t, "NO", info.isNullable)

		info, ok = queryColumn(t, db, "security_events", "event_type")
		require.True(t, ok, "expected security_events.event_type column to exist")
		assert.Equal(t, "NO", info.isNullable)

		info, ok = queryColumn(t, db, "security_events", "user_id")
		require.True(t, ok, "expected security_events.user_id column to exist")
		assert.Equal(t, "YES", info.isNullable, "user_id must be nullable (ON DELETE SET NULL)")

		info, ok = queryColumn(t, db, "security_events", "severity")
		require.True(t, ok, "expected security_events.severity column to exist")
		assert.True(t, checkConstraintExists(t, db, "security_events", "severity"))

		info, ok = queryColumn(t, db, "security_events", "details")
		require.True(t, ok, "expected security_events.details column to exist")
		assert.Equal(t, "jsonb", info.dataType)
		assert.Equal(t, "NO", info.isNullable)

		_, ok = queryColumn(t, db, "security_events", "ip_address")
		assert.True(t, ok, "expected security_events.ip_address column to exist")
		_, ok = queryColumn(t, db, "security_events", "user_agent")
		assert.True(t, ok, "expected security_events.user_agent column to exist")
		_, ok = queryColumn(t, db, "security_events", "timestamp")
		assert.True(t, ok, "expected security_events.timestamp column to exist")

		assert.True(t, indexExists(t, db, "idx_security_events_correlation_id"))
		assert.True(t, indexExists(t, db, "idx_security_events_user_id"))
		assert.True(t, indexExists(t, db, "idx_security_events_event_type"))
		assert.True(t, indexExists(t, db, "idx_security_events_timestamp"))
	})

	t.Run("when migrations roll back to zero they should drop all four tables", func(t *testing.T) {
		require.NoError(t, goose.DownTo(db, migrations.Dir, 0), "failed to roll back migrations to version 0")

		for _, table := range []string{"security_events", "jwt_signing_keys", "refresh_tokens", "users"} {
			var exists bool
			err := db.QueryRow(
				`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)`,
				table,
			).Scan(&exists)
			require.NoError(t, err)
			assert.False(t, exists, "expected table %s to be dropped after goose down", table)
		}
	})
}
