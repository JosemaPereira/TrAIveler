// Package integration: spec008_migrations_test.go verifies the Spec 008
// core data model migrations (008-T011 to 008-T017, plus the folded-in
// 004-T010 trips.version column, issue #142). Like security_migrations_test.go
// this is a migrations-only slice: no Go domain code exists yet for the Spec
// 008 tables, so the only way to verify the schema is by applying goose
// migrations against a real Postgres testcontainer and inspecting
// information_schema / pg_indexes / catalog behavior directly.
//
// The migration set under test (backend/migrations/005-012):
//   - 005 ALTER users        (008-T011: full_name, has_subscription,
//     failed_login_attempts, last_failed_login_at, email_verified, version)
//   - 006 CREATE plans        (unplanned but required: subscriptions.plan_id FK)
//   - 007 CREATE subscriptions (008-T012)
//   - 008 CREATE password_reset_tokens (008-T013)
//   - 009 ALTER security_events (008-T014: add email column + index)
//   - 010 CREATE trips        (008-T015 + folds 004-T010 version column)
//   - 011 CREATE collaborators (008-T016)
//   - 012 CREATE suggestions  (008-T017)
//
// Helpers (openMigrationDB, queryColumn, indexExists, checkConstraintExists,
// startPostgresContainer) are shared with security_migrations_test.go /
// swagger_test.go in this same package and are intentionally reused here
// rather than duplicated.
package integration

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pressly/goose/v3"

	"github.com/JosemaPereira/TrAIveler/backend/internal/database/migrations"
)

// columnExpectation is a table-driven-test row describing one column's
// expected shape in information_schema.columns.
type columnExpectation struct {
	table    string
	column   string
	dataType string // information_schema data_type, "" to skip the type check
	nullable string // "YES" or "NO", "" to skip the nullability check
}

// applyAllMigrations opens a fresh testcontainer DB and applies the ENTIRE
// backend/migrations directory (001-012 plus the trailing timestamp-versioned
// examples migration) via goose.Up. Applying the whole directory — not just
// UpTo a fixed version — is the regression guard the #138 session established:
// a migration that references a not-yet-created table would stop goose.Up and
// block every later migration, including the examples one whose timestamp
// version sorts last.
func applyAllMigrations(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	db := openMigrationDB(t, ctx)
	require.NoError(t, goose.Up(db, migrations.Dir), "goose.Up must apply the entire migrations directory cleanly")
	return db
}

// TestSpec008Migrations_Schema applies the full migration directory and
// asserts every new column, constraint, index, and the plans seed row match
// docs/data-model.md (canonical) + specs/008-auth-collaboration-ux/data-model.md
// (extension fields).
func TestSpec008Migrations_Schema(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := applyAllMigrations(t, ctx)

	t.Run("when Spec 008 migrations apply the users table should gain the new auth fields", func(t *testing.T) {
		columns := []columnExpectation{
			{"users", "full_name", "character varying", "YES"},
			{"users", "has_subscription", "boolean", "NO"},
			{"users", "failed_login_attempts", "integer", "NO"},
			{"users", "last_failed_login_at", "timestamp without time zone", "YES"},
			{"users", "email_verified", "boolean", "NO"},
			{"users", "version", "bigint", "NO"},
		}
		for _, c := range columns {
			info, ok := queryColumn(t, db, c.table, c.column)
			require.True(t, ok, "expected %s.%s column to exist", c.table, c.column)
			assert.Equal(t, c.dataType, info.dataType, "%s.%s data type", c.table, c.column)
			assert.Equal(t, c.nullable, info.isNullable, "%s.%s nullability", c.table, c.column)
		}

		info, ok := queryColumn(t, db, "users", "has_subscription")
		require.True(t, ok)
		assert.Contains(t, info.defaultVal.String, "false", "has_subscription should default to FALSE")

		info, ok = queryColumn(t, db, "users", "version")
		require.True(t, ok)
		assert.Contains(t, info.defaultVal.String, "1", "version should default to 1")

		assert.True(t, indexExists(t, db, "idx_users_failed_login_attempts"),
			"expected idx_users_failed_login_attempts index to exist")
	})

	t.Run("when Spec 008 migrations apply the plans table should exist with its seed row", func(t *testing.T) {
		columns := []columnExpectation{
			{"plans", "id", "uuid", "NO"},
			{"plans", "name", "character varying", "NO"},
			{"plans", "max_admin_users", "integer", "NO"},
			{"plans", "max_partner_users", "integer", "NO"},
			{"plans", "created_at", "timestamp without time zone", "NO"},
		}
		for _, c := range columns {
			info, ok := queryColumn(t, db, c.table, c.column)
			require.True(t, ok, "expected %s.%s column to exist", c.table, c.column)
			assert.Equal(t, c.dataType, info.dataType, "%s.%s data type", c.table, c.column)
			assert.Equal(t, c.nullable, info.isNullable, "%s.%s nullability", c.table, c.column)
		}

		assert.True(t, checkConstraintExists(t, db, "plans", "max_admin_users"),
			"expected a CHECK constraint on plans.max_admin_users")
		assert.True(t, checkConstraintExists(t, db, "plans", "max_partner_users"),
			"expected a CHECK constraint on plans.max_partner_users")

		var name string
		var maxAdmin, maxPartner int
		err := db.QueryRow(
			`SELECT name, max_admin_users, max_partner_users FROM plans WHERE id = $1`,
			"00000000-0000-0000-0000-000000000001",
		).Scan(&name, &maxAdmin, &maxPartner)
		require.NoError(t, err, "expected the documented basic plan seed row to exist")
		assert.Equal(t, "basic", name)
		assert.Equal(t, 1, maxAdmin)
		assert.Equal(t, 1, maxPartner)
	})

	t.Run("when Spec 008 migrations apply the subscriptions table should match the spec", func(t *testing.T) {
		columns := []columnExpectation{
			{"subscriptions", "id", "uuid", "NO"},
			{"subscriptions", "user_id", "uuid", "NO"},
			{"subscriptions", "plan_id", "uuid", "NO"},
			{"subscriptions", "status", "character varying", "NO"},
			{"subscriptions", "stub_payment_ref", "character varying", "YES"},
			{"subscriptions", "grace_period_ends_at", "timestamp without time zone", "YES"},
			{"subscriptions", "cancelled_at", "timestamp without time zone", "YES"},
			{"subscriptions", "created_at", "timestamp without time zone", "NO"},
		}
		for _, c := range columns {
			info, ok := queryColumn(t, db, c.table, c.column)
			require.True(t, ok, "expected %s.%s column to exist", c.table, c.column)
			assert.Equal(t, c.dataType, info.dataType, "%s.%s data type", c.table, c.column)
			assert.Equal(t, c.nullable, info.isNullable, "%s.%s nullability", c.table, c.column)
		}

		assert.True(t, checkConstraintExists(t, db, "subscriptions", "status"),
			"expected a CHECK constraint on subscriptions.status")
		assert.True(t, indexExists(t, db, "idx_subscriptions_user_id"))
		assert.True(t, indexExists(t, db, "idx_subscriptions_status"))
	})

	t.Run("when Spec 008 migrations apply the password_reset_tokens table should match the spec", func(t *testing.T) {
		columns := []columnExpectation{
			{"password_reset_tokens", "id", "uuid", "NO"},
			{"password_reset_tokens", "user_id", "uuid", "NO"},
			{"password_reset_tokens", "token_hash", "text", "NO"},
			{"password_reset_tokens", "expires_at", "timestamp without time zone", "NO"},
			{"password_reset_tokens", "used_at", "timestamp without time zone", "YES"},
			{"password_reset_tokens", "created_at", "timestamp without time zone", "NO"},
		}
		for _, c := range columns {
			info, ok := queryColumn(t, db, c.table, c.column)
			require.True(t, ok, "expected %s.%s column to exist", c.table, c.column)
			assert.Equal(t, c.dataType, info.dataType, "%s.%s data type", c.table, c.column)
			assert.Equal(t, c.nullable, info.isNullable, "%s.%s nullability", c.table, c.column)
		}

		assert.True(t, checkConstraintExists(t, db, "password_reset_tokens", "expires_at"),
			"expected a CHECK constraint enforcing expires_at > created_at")
		assert.True(t, checkConstraintExists(t, db, "password_reset_tokens", "used_at"),
			"expected a CHECK constraint enforcing used_at >= created_at")
		assert.True(t, indexExists(t, db, "idx_password_reset_tokens_user_id"))
		assert.True(t, indexExists(t, db, "idx_password_reset_tokens_token_hash"))
		assert.True(t, indexExists(t, db, "idx_password_reset_tokens_expires_at"))
	})

	t.Run("when Spec 008 migrations apply the security_events table should gain the email column", func(t *testing.T) {
		info, ok := queryColumn(t, db, "security_events", "email")
		require.True(t, ok, "expected security_events.email column to exist")
		assert.Equal(t, "character varying", info.dataType)
		assert.Equal(t, "YES", info.isNullable, "security_events.email must be nullable")

		assert.True(t, indexExists(t, db, "idx_security_events_email"),
			"expected idx_security_events_email index to exist")

		// The pre-existing timestamp column must NOT have been renamed to
		// created_at (that would break internal/observability.LogSecurityEvent).
		_, ok = queryColumn(t, db, "security_events", "timestamp")
		assert.True(t, ok, "security_events.timestamp must be preserved (not renamed)")
	})

	t.Run("when Spec 008 migrations apply the trips table should match the spec plus archived", func(t *testing.T) {
		columns := []columnExpectation{
			{"trips", "id", "uuid", "NO"},
			{"trips", "creator_id", "uuid", "NO"},
			{"trips", "title", "character varying", "NO"},
			{"trips", "description", "text", "YES"},
			{"trips", "status", "character varying", "NO"},
			{"trips", "archived", "boolean", "NO"},
			{"trips", "version", "bigint", "NO"},
			{"trips", "created_at", "timestamp without time zone", "NO"},
			{"trips", "updated_at", "timestamp without time zone", "NO"},
		}
		for _, c := range columns {
			info, ok := queryColumn(t, db, c.table, c.column)
			require.True(t, ok, "expected %s.%s column to exist", c.table, c.column)
			assert.Equal(t, c.dataType, info.dataType, "%s.%s data type", c.table, c.column)
			assert.Equal(t, c.nullable, info.isNullable, "%s.%s nullability", c.table, c.column)
		}

		info, ok := queryColumn(t, db, "trips", "status")
		require.True(t, ok)
		assert.Contains(t, info.defaultVal.String, "draft", "trips.status should default to 'draft'")

		info, ok = queryColumn(t, db, "trips", "version")
		require.True(t, ok)
		assert.Contains(t, info.defaultVal.String, "1", "trips.version should default to 1 (004-T010)")

		assert.True(t, checkConstraintExists(t, db, "trips", "status"),
			"expected a CHECK constraint on trips.status")
		assert.True(t, indexExists(t, db, "idx_trips_creator_id"))
		assert.True(t, indexExists(t, db, "idx_trips_status"))
		assert.True(t, indexExists(t, db, "idx_trips_creator_id_archived"))
	})

	t.Run("when Spec 008 migrations apply the collaborators table should match the spec", func(t *testing.T) {
		columns := []columnExpectation{
			{"collaborators", "id", "uuid", "NO"},
			{"collaborators", "trip_id", "uuid", "NO"},
			{"collaborators", "user_id", "uuid", "YES"}, // nullable: invite-by-email before registration
			{"collaborators", "email", "character varying", "NO"},
			{"collaborators", "status", "character varying", "NO"},
			{"collaborators", "invited_at", "timestamp without time zone", "NO"},
			{"collaborators", "accepted_at", "timestamp without time zone", "YES"},
		}
		for _, c := range columns {
			info, ok := queryColumn(t, db, c.table, c.column)
			require.True(t, ok, "expected %s.%s column to exist", c.table, c.column)
			assert.Equal(t, c.dataType, info.dataType, "%s.%s data type", c.table, c.column)
			assert.Equal(t, c.nullable, info.isNullable, "%s.%s nullability", c.table, c.column)
		}

		assert.True(t, checkConstraintExists(t, db, "collaborators", "status"),
			"expected a CHECK constraint on collaborators.status")
		assert.True(t, indexExists(t, db, "idx_collaborators_trip_id"))
		assert.True(t, indexExists(t, db, "idx_collaborators_user_id"))
		assert.True(t, indexExists(t, db, "idx_collaborators_status"))
	})

	t.Run("when Spec 008 migrations apply the suggestions table should match the spec", func(t *testing.T) {
		columns := []columnExpectation{
			{"suggestions", "id", "uuid", "NO"},
			{"suggestions", "trip_id", "uuid", "NO"},
			{"suggestions", "author_id", "uuid", "NO"},
			{"suggestions", "target_type", "character varying", "NO"},
			{"suggestions", "target_id", "uuid", "NO"},
			{"suggestions", "suggested_action", "character varying", "NO"},
			{"suggestions", "payload", "jsonb", "NO"},
			{"suggestions", "status", "character varying", "NO"},
			{"suggestions", "reviewed_by", "uuid", "YES"},
			{"suggestions", "reviewed_at", "timestamp without time zone", "YES"},
			{"suggestions", "created_at", "timestamp without time zone", "NO"},
		}
		for _, c := range columns {
			info, ok := queryColumn(t, db, c.table, c.column)
			require.True(t, ok, "expected %s.%s column to exist", c.table, c.column)
			assert.Equal(t, c.dataType, info.dataType, "%s.%s data type", c.table, c.column)
			assert.Equal(t, c.nullable, info.isNullable, "%s.%s nullability", c.table, c.column)
		}

		assert.True(t, checkConstraintExists(t, db, "suggestions", "target_type"),
			"expected a CHECK constraint on suggestions.target_type")
		assert.True(t, checkConstraintExists(t, db, "suggestions", "suggested_action"),
			"expected a CHECK constraint on suggestions.suggested_action")
		assert.True(t, checkConstraintExists(t, db, "suggestions", "status"),
			"expected a CHECK constraint on suggestions.status")
		assert.True(t, indexExists(t, db, "idx_suggestions_trip_id"))
		assert.True(t, indexExists(t, db, "idx_suggestions_author_id"))
		assert.True(t, indexExists(t, db, "idx_suggestions_status"))
	})
}

// TestSpec008Migrations_ForeignKeyBehaviors applies the full migration
// directory, seeds a minimal valid object graph, and verifies the cascade /
// restrict / set-null behaviors that matter for data integrity.
func TestSpec008Migrations_ForeignKeyBehaviors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := applyAllMigrations(t, ctx)

	// A stable UUID set for the seeded graph, reused across subtests.
	const (
		adminID    = "11111111-1111-1111-1111-111111111111"
		partnerID  = "22222222-2222-2222-2222-222222222222"
		reviewerID = "33333333-3333-3333-3333-333333333333"
		planID     = "00000000-0000-0000-0000-000000000001" // seeded basic plan
		subID      = "44444444-4444-4444-4444-444444444444"
		tripID     = "55555555-5555-5555-5555-555555555555"
		collabID   = "66666666-6666-6666-6666-666666666666"
		suggID     = "77777777-7777-7777-7777-777777777777"
	)

	seedGraph := func(t *testing.T) {
		t.Helper()
		_, err := db.Exec(
			`INSERT INTO users (id, email, password_hash, role) VALUES
			 ($1, 'admin@example.com', 'x', 'admin'),
			 ($2, 'partner@example.com', 'x', 'partner'),
			 ($3, 'reviewer@example.com', 'x', 'admin')`,
			adminID, partnerID, reviewerID,
		)
		require.NoError(t, err, "failed to seed users")

		_, err = db.Exec(
			`INSERT INTO subscriptions (id, user_id, plan_id, status) VALUES ($1, $2, $3, 'active')`,
			subID, adminID, planID,
		)
		require.NoError(t, err, "failed to seed subscription")

		_, err = db.Exec(
			`INSERT INTO trips (id, creator_id, title, status) VALUES ($1, $2, 'Trip', 'draft')`,
			tripID, adminID,
		)
		require.NoError(t, err, "failed to seed trip")

		_, err = db.Exec(
			`INSERT INTO collaborators (id, trip_id, user_id, email, status) VALUES ($1, $2, $3, 'partner@example.com', 'accepted')`,
			collabID, tripID, partnerID,
		)
		require.NoError(t, err, "failed to seed collaborator")

		_, err = db.Exec(
			`INSERT INTO suggestions (id, trip_id, author_id, target_type, target_id, suggested_action, payload, status, reviewed_by)
			 VALUES ($1, $2, $3, 'day', $4, 'add', '{}'::jsonb, 'pending', $5)`,
			suggID, tripID, partnerID, tripID, reviewerID,
		)
		require.NoError(t, err, "failed to seed suggestion")
	}

	seedGraph(t)

	t.Run("when a plan still has a subscription it should not be deletable (RESTRICT)", func(t *testing.T) {
		_, err := db.Exec(`DELETE FROM plans WHERE id = $1`, planID)
		require.Error(t, err, "deleting a plan referenced by a subscription must be blocked by RESTRICT")
	})

	t.Run("when a user still owns a trip it should not be deletable (RESTRICT)", func(t *testing.T) {
		_, err := db.Exec(`DELETE FROM users WHERE id = $1`, adminID)
		require.Error(t, err, "deleting a user who created a trip must be blocked by RESTRICT")
	})

	t.Run("when the reviewer is deleted the suggestion reviewed_by should be set to NULL", func(t *testing.T) {
		_, err := db.Exec(`DELETE FROM users WHERE id = $1`, reviewerID)
		require.NoError(t, err, "deleting the reviewer must be allowed (SET NULL)")

		var reviewedBy sql.NullString
		err = db.QueryRow(`SELECT reviewed_by FROM suggestions WHERE id = $1`, suggID).Scan(&reviewedBy)
		require.NoError(t, err)
		assert.False(t, reviewedBy.Valid, "suggestions.reviewed_by should be NULL after reviewer deletion")
	})

	t.Run("when a trip is deleted its collaborators and suggestions should cascade", func(t *testing.T) {
		_, err := db.Exec(`DELETE FROM trips WHERE id = $1`, tripID)
		require.NoError(t, err, "deleting a trip must be allowed and cascade to children")

		var collabCount, suggCount int
		require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM collaborators WHERE trip_id = $1`, tripID).Scan(&collabCount))
		require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM suggestions WHERE trip_id = $1`, tripID).Scan(&suggCount))
		assert.Equal(t, 0, collabCount, "collaborators should cascade-delete with their trip")
		assert.Equal(t, 0, suggCount, "suggestions should cascade-delete with their trip")
	})
}

// TestSpec008Migrations_Reversibility verifies the full Up path (entire
// directory, including the trailing examples migration) applies cleanly and
// that DownTo the last Spec 004 version rolls back every Spec 008 table while
// leaving the Spec 004 core tables intact.
func TestSpec008Migrations_Reversibility(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := applyAllMigrations(t, ctx)

	spec008Tables := []string{
		"suggestions", "collaborators", "trips",
		"password_reset_tokens", "subscriptions", "plans",
	}
	spec004Tables := []string{"users", "refresh_tokens", "jwt_signing_keys", "security_events"}

	tableExists := func(t *testing.T, table string) bool {
		t.Helper()
		var exists bool
		require.NoError(t, db.QueryRow(
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)`,
			table,
		).Scan(&exists))
		return exists
	}

	t.Run("when the full directory is applied all Spec 008 tables should exist", func(t *testing.T) {
		for _, table := range spec008Tables {
			assert.True(t, tableExists(t, table), "expected table %s to exist after goose.Up", table)
		}
	})

	t.Run("when rolling back to the Spec 004 version the Spec 008 tables should be dropped", func(t *testing.T) {
		require.NoError(t, goose.DownTo(db, migrations.Dir, coreVersion),
			"failed to roll back Spec 008 migrations to version %d", coreVersion)

		for _, table := range spec008Tables {
			assert.False(t, tableExists(t, table), "expected Spec 008 table %s to be dropped after DownTo", table)
		}

		// The Spec 004 core tables and the new users columns' base table must
		// survive the rollback.
		for _, table := range spec004Tables {
			assert.True(t, tableExists(t, table), "expected Spec 004 table %s to survive rollback to version %d", table, coreVersion)
		}

		// The users ALTER (version 5) is rolled back, so its added columns
		// should be gone while the base table remains.
		_, ok := queryColumn(t, db, "users", "failed_login_attempts")
		assert.False(t, ok, "users.failed_login_attempts should be dropped after rolling back the ALTER migration")
	})
}
