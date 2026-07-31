// Package integration: itinerary_migrations_test.go verifies the itinerary
// data model migrations (004-T133 to 004-T136, issue #170) — the
// destinations → days → activities chain that hangs off trips.
//
// Like security_migrations_test.go and spec008_migrations_test.go this is a
// migrations-only slice: no Go domain code exists yet for these tables, so the
// schema can only be verified by applying goose migrations against a real
// Postgres testcontainer and inspecting the catalog directly. Helpers are
// shared with those two files rather than duplicated.
//
// The migration set under test (backend/migrations/013-015):
//   - 013 CREATE destinations (004-T133: coordinate CHECKs, country + coordinate indexes)
//   - 014 CREATE days         (004-T134: trip CASCADE, destination SET NULL, UNIQUE(trip_id, day_number))
//   - 015 CREATE activities   (004-T135: day CASCADE, type/sequence_order CHECKs, `version` inline)
//
// 004-T011 asked for a `version` counter on the versioned itinerary entity.
// That entity is Activity, not the never-modeled `itinerary_items` its task
// text names (see specs/004-security-auth-model/tasks.md "Naming correction"),
// so the counter ships inline in 015 the way 010_create_trips_table.sql
// shipped trips.version for 004-T010. The activities.version assertions below
// are 004-T011's acceptance criteria.
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

// preItineraryVersion is the goose version of the last migration that predates
// this set (012, CREATE suggestions). Rolling back to it must drop activities,
// days, and destinations while leaving the Spec 008 tables — trips above all —
// intact.
const preItineraryVersion = 12

// TestItineraryMigrations_Schema applies the full migration directory and
// asserts the destinations, days, and activities columns, defaults, and
// indexes match docs/data-model.md (canonical).
func TestItineraryMigrations_Schema(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := applyAllMigrations(t, ctx)

	t.Run("when the itinerary migrations apply the destinations table should match the spec", func(t *testing.T) {
		columns := []columnExpectation{
			{"destinations", "id", "uuid", "NO"},
			{"destinations", "name", "character varying", "NO"},
			{"destinations", "country", "character varying", "NO"},
			{"destinations", "region", "character varying", "YES"},
			{"destinations", "latitude", "numeric", "NO"},
			{"destinations", "longitude", "numeric", "NO"},
			{"destinations", "created_at", "timestamp without time zone", "NO"},
		}
		for _, c := range columns {
			info, ok := queryColumn(t, db, c.table, c.column)
			require.True(t, ok, "expected %s.%s column to exist", c.table, c.column)
			assert.Equal(t, c.dataType, info.dataType, "%s.%s data type", c.table, c.column)
			assert.Equal(t, c.nullable, info.isNullable, "%s.%s nullability", c.table, c.column)
		}

		assert.True(t, indexExists(t, db, "idx_destinations_country"))
		assert.True(t, indexExists(t, db, "idx_destinations_coordinates"))

		// Destination is deliberately NOT a versioned entity (docs/data-model.md
		// gives the optimistic-locking counter to Trip and Activity only).
		_, ok := queryColumn(t, db, "destinations", "version")
		assert.False(t, ok, "destinations must not carry a version column")
	})

	t.Run("when the itinerary migrations apply the days table should match the spec", func(t *testing.T) {
		columns := []columnExpectation{
			{"days", "id", "uuid", "NO"},
			{"days", "trip_id", "uuid", "NO"},
			{"days", "destination_id", "uuid", "YES"}, // nullable: arrival/transfer days
			{"days", "day_number", "integer", "NO"},
			{"days", "label", "character varying", "YES"},
			{"days", "created_at", "timestamp without time zone", "NO"},
		}
		for _, c := range columns {
			info, ok := queryColumn(t, db, c.table, c.column)
			require.True(t, ok, "expected %s.%s column to exist", c.table, c.column)
			assert.Equal(t, c.dataType, info.dataType, "%s.%s data type", c.table, c.column)
			assert.Equal(t, c.nullable, info.isNullable, "%s.%s nullability", c.table, c.column)
		}

		assert.True(t, indexExists(t, db, "idx_days_trip_id"))
		assert.True(t, indexExists(t, db, "idx_days_destination_id"))
		assert.True(t, indexExists(t, db, "uq_days_trip_day_number"),
			"the UNIQUE(trip_id, day_number) constraint should back a same-named index")

		_, ok := queryColumn(t, db, "days", "version")
		assert.False(t, ok, "days must not carry a version column")
	})

	t.Run("when the itinerary migrations apply the activities table should match the spec", func(t *testing.T) {
		columns := []columnExpectation{
			{"activities", "id", "uuid", "NO"},
			{"activities", "day_id", "uuid", "NO"},
			{"activities", "title", "character varying", "NO"},
			{"activities", "type", "character varying", "NO"},
			{"activities", "sequence_order", "integer", "NO"},
			{"activities", "description", "text", "YES"},
			{"activities", "is_ai_generated", "boolean", "NO"},
			{"activities", "metadata", "jsonb", "YES"},
			{"activities", "version", "bigint", "NO"},
			{"activities", "created_at", "timestamp without time zone", "NO"},
			{"activities", "updated_at", "timestamp without time zone", "NO"},
		}
		for _, c := range columns {
			info, ok := queryColumn(t, db, c.table, c.column)
			require.True(t, ok, "expected %s.%s column to exist", c.table, c.column)
			assert.Equal(t, c.dataType, info.dataType, "%s.%s data type", c.table, c.column)
			assert.Equal(t, c.nullable, info.isNullable, "%s.%s nullability", c.table, c.column)
		}

		info, ok := queryColumn(t, db, "activities", "is_ai_generated")
		require.True(t, ok)
		assert.Contains(t, info.defaultVal.String, "true", "is_ai_generated should default to TRUE")

		// 004-T011: the optimistic-locking counter, folded into this CREATE
		// rather than shipped as a separate ALTER (mirrors trips.version).
		info, ok = queryColumn(t, db, "activities", "version")
		require.True(t, ok)
		assert.Contains(t, info.defaultVal.String, "1", "activities.version should default to 1 (004-T011)")

		assert.True(t, indexExists(t, db, "idx_activities_day_id"))
	})
}

// TestItineraryMigrations_ConstraintsAndForeignKeys applies the full migration
// directory, seeds a minimal valid itinerary graph, and verifies the CHECK
// constraints and the cascade / set-null behaviors that matter for data
// integrity.
func TestItineraryMigrations_ConstraintsAndForeignKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := applyAllMigrations(t, ctx)

	// A stable UUID set for the seeded graph, reused across subtests.
	const (
		adminID  = "11111111-1111-1111-1111-111111111111"
		tripAID  = "22222222-2222-2222-2222-222222222222"
		tripBID  = "33333333-3333-3333-3333-333333333333"
		destID   = "44444444-4444-4444-4444-444444444444"
		dayA1ID  = "55555555-5555-5555-5555-555555555555"
		dayA2ID  = "66666666-6666-6666-6666-666666666666"
		actA1ID  = "77777777-7777-7777-7777-777777777777"
		actA2ID  = "88888888-8888-8888-8888-888888888888"
		strayID  = "99999999-9999-9999-9999-999999999999"
		absentID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	)

	seedGraph := func(t *testing.T) {
		t.Helper()
		_, err := db.Exec(
			`INSERT INTO users (id, email, password_hash, role) VALUES ($1, 'admin@example.com', 'x', 'admin')`,
			adminID,
		)
		require.NoError(t, err, "failed to seed user")

		_, err = db.Exec(
			`INSERT INTO trips (id, creator_id, title, status) VALUES
			 ($1, $3, 'Trip A', 'draft'),
			 ($2, $3, 'Trip B', 'draft')`,
			tripAID, tripBID, adminID,
		)
		require.NoError(t, err, "failed to seed trips")

		_, err = db.Exec(
			`INSERT INTO destinations (id, name, country, region, latitude, longitude)
			 VALUES ($1, 'Tokyo', 'JP', 'Kanto', 35.689487, 139.691711)`,
			destID,
		)
		require.NoError(t, err, "failed to seed destination")

		_, err = db.Exec(
			`INSERT INTO days (id, trip_id, destination_id, day_number, label) VALUES
			 ($1, $3, $4, 1, 'Tokyo Day 1'),
			 ($2, $3, $4, 2, 'Tokyo Day 2')`,
			dayA1ID, dayA2ID, tripAID, destID,
		)
		require.NoError(t, err, "failed to seed days")

		_, err = db.Exec(
			`INSERT INTO activities (id, day_id, title, type, sequence_order) VALUES
			 ($1, $3, 'Senso-ji', 'visit', 1),
			 ($2, $4, 'Tsukiji breakfast', 'food', 1)`,
			actA1ID, actA2ID, dayA1ID, dayA2ID,
		)
		require.NoError(t, err, "failed to seed activities")
	}

	seedGraph(t)

	t.Run("when a destination violates its column constraints it should be rejected", func(t *testing.T) {
		tests := []struct {
			name      string
			country   string
			latitude  float64
			longitude float64
		}{
			{"having a country code that is not two characters", "JPN", 35.0, 139.0},
			{"having a latitude above the valid range", "JP", 90.5, 139.0},
			{"having a latitude below the valid range", "JP", -90.5, 139.0},
			{"having a longitude above the valid range", "JP", 35.0, 180.5},
			{"having a longitude below the valid range", "JP", 35.0, -180.5},
		}
		for _, tc := range tests {
			t.Run(tc.name+" it should be rejected", func(t *testing.T) {
				_, err := db.Exec(
					`INSERT INTO destinations (name, country, latitude, longitude) VALUES ('Bad', $1, $2, $3)`,
					tc.country, tc.latitude, tc.longitude,
				)
				require.Error(t, err, "expected the CHECK constraint to reject this destination")
			})
		}
	})

	t.Run("when a day violates its constraints it should be rejected", func(t *testing.T) {
		t.Run("having a non-positive day number", func(t *testing.T) {
			_, err := db.Exec(
				`INSERT INTO days (trip_id, day_number) VALUES ($1, 0)`, tripAID,
			)
			require.Error(t, err, "expected CHECK (day_number > 0) to reject day_number 0")
		})

		t.Run("having a day number already used in the same trip", func(t *testing.T) {
			_, err := db.Exec(
				`INSERT INTO days (trip_id, day_number) VALUES ($1, 1)`, tripAID,
			)
			require.Error(t, err, "expected UNIQUE(trip_id, day_number) to reject the duplicate")
		})

		t.Run("having an unknown trip", func(t *testing.T) {
			_, err := db.Exec(
				`INSERT INTO days (trip_id, day_number) VALUES ($1, 1)`, absentID,
			)
			require.Error(t, err, "expected the trip_id foreign key to reject an unknown trip")
		})
	})

	t.Run("when the same day number is used in a different trip it should be allowed", func(t *testing.T) {
		_, err := db.Exec(
			`INSERT INTO days (id, trip_id, day_number) VALUES ($1, $2, 1)`, strayID, tripBID,
		)
		require.NoError(t, err, "UNIQUE(trip_id, day_number) must be scoped per trip, not global")
	})

	t.Run("when a day has no destination it should be allowed", func(t *testing.T) {
		var destinationID sql.NullString
		require.NoError(t, db.QueryRow(`SELECT destination_id FROM days WHERE id = $1`, strayID).Scan(&destinationID))
		assert.False(t, destinationID.Valid, "destination_id must be nullable for arrival/transfer days")
	})

	t.Run("when an activity violates its constraints it should be rejected", func(t *testing.T) {
		t.Run("having a type outside the allowed set", func(t *testing.T) {
			_, err := db.Exec(
				`INSERT INTO activities (day_id, title, type, sequence_order) VALUES ($1, 'Bad', 'sightseeing', 1)`,
				dayA1ID,
			)
			require.Error(t, err, "expected the type CHECK constraint to reject an unknown activity type")
		})

		t.Run("having a non-positive sequence order", func(t *testing.T) {
			_, err := db.Exec(
				`INSERT INTO activities (day_id, title, type, sequence_order) VALUES ($1, 'Bad', 'visit', 0)`,
				dayA1ID,
			)
			require.Error(t, err, "expected CHECK (sequence_order > 0) to reject sequence_order 0")
		})

		t.Run("having an unknown day", func(t *testing.T) {
			_, err := db.Exec(
				`INSERT INTO activities (day_id, title, type, sequence_order) VALUES ($1, 'Orphan', 'visit', 1)`,
				absentID,
			)
			require.Error(t, err, "expected the day_id foreign key to reject an unknown day")
		})
	})

	// 004-T011's acceptance criterion, asserted behaviorally rather than only
	// against information_schema: a freshly inserted activity starts at
	// version 1 so optimistic locking has a baseline to compare If-Match with.
	t.Run("when an activity is inserted without a version it should start at 1", func(t *testing.T) {
		var version int64
		require.NoError(t, db.QueryRow(`SELECT version FROM activities WHERE id = $1`, actA1ID).Scan(&version))
		assert.Equal(t, int64(1), version, "activities.version must default to 1 (004-T011)")
	})

	t.Run("when a destination is deleted its days should survive with a NULL destination", func(t *testing.T) {
		_, err := db.Exec(`DELETE FROM destinations WHERE id = $1`, destID)
		require.NoError(t, err, "deleting a destination must be allowed (SET NULL)")

		var dayCount int
		require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM days WHERE trip_id = $1`, tripAID).Scan(&dayCount))
		assert.Equal(t, 2, dayCount, "days must survive their destination's deletion")

		var destinationID sql.NullString
		require.NoError(t, db.QueryRow(`SELECT destination_id FROM days WHERE id = $1`, dayA1ID).Scan(&destinationID))
		assert.False(t, destinationID.Valid, "days.destination_id should be NULL after destination deletion")
	})

	t.Run("when a day is deleted its activities should cascade", func(t *testing.T) {
		_, err := db.Exec(`DELETE FROM days WHERE id = $1`, dayA1ID)
		require.NoError(t, err, "deleting a day must be allowed and cascade to its activities")

		var activityCount int
		require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM activities WHERE id = $1`, actA1ID).Scan(&activityCount))
		assert.Equal(t, 0, activityCount, "activities should cascade-delete with their day")

		require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM activities WHERE id = $1`, actA2ID).Scan(&activityCount))
		assert.Equal(t, 1, activityCount, "activities on other days must be untouched")
	})

	t.Run("when a trip is deleted its days and activities should cascade", func(t *testing.T) {
		_, err := db.Exec(`DELETE FROM trips WHERE id = $1`, tripAID)
		require.NoError(t, err, "deleting a trip must be allowed and cascade to days and activities")

		var dayCount, activityCount int
		require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM days WHERE trip_id = $1`, tripAID).Scan(&dayCount))
		require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM activities WHERE id = $1`, actA2ID).Scan(&activityCount))
		assert.Equal(t, 0, dayCount, "days should cascade-delete with their trip")
		assert.Equal(t, 0, activityCount, "activities should cascade-delete through their day when the trip is deleted")
	})
}

// TestItineraryMigrations_Reversibility verifies the full Up path (entire
// directory, including the trailing examples migration) applies cleanly and
// that DownTo the last pre-itinerary version rolls back activities, days, and
// destinations while leaving trips and the rest of the Spec 008 set intact.
func TestItineraryMigrations_Reversibility(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainer test in short mode")
	}
	ctx := context.Background()
	db := applyAllMigrations(t, ctx)

	itineraryTables := []string{"activities", "days", "destinations"}

	tableExists := func(t *testing.T, table string) bool {
		t.Helper()
		var exists bool
		require.NoError(t, db.QueryRow(
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)`,
			table,
		).Scan(&exists))
		return exists
	}

	t.Run("when the full directory is applied all itinerary tables should exist", func(t *testing.T) {
		for _, table := range itineraryTables {
			assert.True(t, tableExists(t, table), "expected table %s to exist after goose.Up", table)
		}
	})

	t.Run("when rolling back to the pre-itinerary version the itinerary tables should be dropped", func(t *testing.T) {
		require.NoError(t, goose.DownTo(db, migrations.Dir, preItineraryVersion),
			"failed to roll back the itinerary migrations to version %d", preItineraryVersion)

		for _, table := range itineraryTables {
			assert.False(t, tableExists(t, table), "expected itinerary table %s to be dropped after DownTo", table)
		}

		// The parent table the itinerary hangs off must survive the rollback.
		assert.True(t, tableExists(t, "trips"),
			"expected trips to survive rollback to version %d", preItineraryVersion)
	})
}
