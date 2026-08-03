package trip

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JosemaPereira/TrAIveler/backend/internal/database"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// Repository defines data access for Trip, Destination, Day, and Activity
// (001-T033). PostgresRepository is the only implementation; tests use the
// generated mock (mocks/repository_mock.go). Method names follow
// specs/001-product-vision-scope/tasks.md's T033 literal list rather than
// this codebase's usual short-form convention (Create/Find/...), since the
// package already carries multiple entities and the longer names disambiguate
// which one each method targets.
type Repository interface {
	// CreateTrip inserts trip, which must already have ID/CreatorID/Title/
	// Status populated by the caller. Populates CreatedAt/UpdatedAt/Version
	// from the database defaults.
	CreateTrip(ctx context.Context, trip *Trip) error
	// ListTripsByUser returns every trip owned by userID, most recently
	// created first. Returns an empty (non-nil) slice, not an error, when the
	// user owns no trips.
	ListTripsByUser(ctx context.Context, userID string) ([]*Trip, error)
	// FindTripByID returns the trip with the given ID, or a domain NotFound
	// error if no row matches.
	FindTripByID(ctx context.Context, id string) (*Trip, error)
	// UpdateTrip writes the mutable trip fields (title, description, status)
	// for trip.ID using optimistic locking on trip.Version: the UPDATE
	// matches on id AND version and increments version atomically, with
	// RETURNING feeding the new version/updated_at back into trip. A
	// zero-row result is disambiguated by a follow-up existence check
	// (mirrors auth.PostgresUserRepository.UpdateUser/resolveUpdateMiss): a
	// present id means the version was stale (domain Conflict), an absent
	// id means the trip is gone (domain NotFound).
	UpdateTrip(ctx context.Context, trip *Trip) error
	// DeleteTrip removes the trip with the given ID (and, via ON DELETE
	// CASCADE, every Day/Activity/ConversationSession under it). Deleting an
	// ID that doesn't exist is reported as a domain NotFound.
	DeleteTrip(ctx context.Context, id string) error
	// CreateDestination inserts destination, populating ID/CreatedAt from the
	// database defaults. Deliberately a plain insert with no dedup/find-or-
	// create logic: destinations are reference data with no uniqueness
	// constraint (migrations/013_create_destinations_table.sql), so minor
	// coordinate variations of the same place are allowed to coexist.
	CreateDestination(ctx context.Context, destination *Destination) error
	// UpsertDay inserts a new Day, or updates the existing row in place when
	// (TripID, DayNumber) already exists (the table's UNIQUE constraint). Day
	// has no version column, so there is no optimistic-locking check here.
	UpsertDay(ctx context.Context, day *Day) error
	// UpsertActivity inserts a new Activity when activity.ID is empty
	// (version defaults to 1 at the database level), or applies a
	// version-checked update when activity.ID is set: the update only
	// succeeds when both ID and the caller-supplied Version match, and a
	// stale Version is reported as a domain Conflict rather than silently
	// overwriting a concurrent change.
	UpsertActivity(ctx context.Context, activity *Activity) error
	// DeleteActivity removes the activity with the given ID. Deleting an ID
	// that doesn't exist is reported as a domain NotFound.
	DeleteActivity(ctx context.Context, id string) error
}

// PostgresRepository implements Repository using the shared pgx pool exposed
// by database.Client.
type PostgresRepository struct {
	db database.Client
}

// NewPostgresRepository builds a Repository backed by PostgreSQL.
func NewPostgresRepository(db database.Client) Repository {
	return &PostgresRepository{db: db}
}

// CreateTrip inserts trip; see the Repository interface doc comment.
func (r *PostgresRepository) CreateTrip(ctx context.Context, trip *Trip) error {
	const query = `
		INSERT INTO trips (id, creator_id, title, description, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING archived, version, created_at, updated_at`

	err := r.db.Pool().QueryRow(ctx, query,
		trip.ID, trip.CreatorID, trip.Title, trip.Description, trip.Status,
	).Scan(&trip.Archived, &trip.Version, &trip.CreatedAt, &trip.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create trip: %w", err)
	}

	return nil
}

// ListTripsByUser returns userID's trips; see the Repository interface doc
// comment. No pagination: MVP scope has no requirement for it on this query.
func (r *PostgresRepository) ListTripsByUser(ctx context.Context, userID string) ([]*Trip, error) {
	const query = `
		SELECT id, creator_id, title, description, status, archived, version, created_at, updated_at
		FROM trips
		WHERE creator_id = $1
		ORDER BY created_at DESC, id DESC`

	rows, err := r.db.Pool().Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list trips by user: %w", err)
	}
	defer rows.Close()

	trips := []*Trip{}
	for rows.Next() {
		var tr Trip
		if err := rows.Scan(
			&tr.ID, &tr.CreatorID, &tr.Title, &tr.Description, &tr.Status,
			&tr.Archived, &tr.Version, &tr.CreatedAt, &tr.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan trip row: %w", err)
		}
		trips = append(trips, &tr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list trips by user: %w", err)
	}

	return trips, nil
}

// FindTripByID returns the trip with the given id; see the Repository
// interface doc comment.
func (r *PostgresRepository) FindTripByID(ctx context.Context, id string) (*Trip, error) {
	const query = `
		SELECT id, creator_id, title, description, status, archived, version, created_at, updated_at
		FROM trips
		WHERE id = $1`

	var tr Trip
	err := r.db.Pool().QueryRow(ctx, query, id).Scan(
		&tr.ID, &tr.CreatorID, &tr.Title, &tr.Description, &tr.Status,
		&tr.Archived, &tr.Version, &tr.CreatedAt, &tr.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("trip", id)
		}
		return nil, fmt.Errorf("find trip: %w", err)
	}

	return &tr, nil
}

// UpdateTrip writes trip's mutable fields; see the Repository interface doc
// comment.
func (r *PostgresRepository) UpdateTrip(ctx context.Context, trip *Trip) error {
	const query = `
		UPDATE trips
		SET title = $1, description = $2, status = $3, updated_at = NOW(), version = version + 1
		WHERE id = $4 AND version = $5
		RETURNING version, updated_at`

	err := r.db.Pool().QueryRow(ctx, query,
		trip.Title, trip.Description, trip.Status, trip.ID, trip.Version,
	).Scan(&trip.Version, &trip.UpdatedAt)
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return r.resolveTripUpdateMiss(ctx, trip.ID)
	}
	return fmt.Errorf("update trip: %w", err)
}

// resolveTripUpdateMiss disambiguates a zero-row optimistic update on trips:
// an existing id means the version was stale (Conflict); a missing id means
// NotFound.
func (r *PostgresRepository) resolveTripUpdateMiss(ctx context.Context, id string) error {
	var exists bool
	err := r.db.Pool().QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM trips WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check trip existence: %w", err)
	}
	if exists {
		return domainerrors.Conflict(fmt.Sprintf("trip %q was modified concurrently", id))
	}
	return domainerrors.NotFound("trip", id)
}

// DeleteTrip removes the trip with the given id; see the Repository
// interface doc comment.
func (r *PostgresRepository) DeleteTrip(ctx context.Context, id string) error {
	const query = `DELETE FROM trips WHERE id = $1`

	tag, err := r.db.Pool().Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete trip: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.NotFound("trip", id)
	}

	return nil
}

// CreateDestination inserts destination; see the Repository interface doc
// comment.
func (r *PostgresRepository) CreateDestination(ctx context.Context, destination *Destination) error {
	const query = `
		INSERT INTO destinations (name, country, region, latitude, longitude)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err := r.db.Pool().QueryRow(ctx, query,
		destination.Name, destination.Country, destination.Region, destination.Latitude, destination.Longitude,
	).Scan(&destination.ID, &destination.CreatedAt)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}

	return nil
}

// UpsertDay inserts or updates day; see the Repository interface doc
// comment. The ON CONFLICT target is the table's UNIQUE(trip_id,
// day_number) constraint (migrations/014_create_days_table.sql).
func (r *PostgresRepository) UpsertDay(ctx context.Context, day *Day) error {
	const query = `
		INSERT INTO days (trip_id, destination_id, day_number, label)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (trip_id, day_number) DO UPDATE
		SET destination_id = EXCLUDED.destination_id, label = EXCLUDED.label
		RETURNING id, created_at`

	err := r.db.Pool().QueryRow(ctx, query,
		day.TripID, day.DestinationID, day.DayNumber, day.Label,
	).Scan(&day.ID, &day.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert day: %w", err)
	}

	return nil
}

// UpsertActivity inserts or version-checked-updates activity; see the
// Repository interface doc comment.
func (r *PostgresRepository) UpsertActivity(ctx context.Context, activity *Activity) error {
	if activity.ID == "" {
		return r.insertActivity(ctx, activity)
	}
	return r.updateActivityWithVersionCheck(ctx, activity)
}

// insertActivity creates a new activity row; version defaults to 1 at the
// database level (migrations/015_create_activities_table.sql).
func (r *PostgresRepository) insertActivity(ctx context.Context, activity *Activity) error {
	const query = `
		INSERT INTO activities (day_id, title, type, sequence_order, description, is_ai_generated, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, version, created_at, updated_at`

	err := r.db.Pool().QueryRow(ctx, query,
		activity.DayID, activity.Title, activity.Type, activity.SequenceOrder,
		activity.Description, activity.IsAIGenerated, activity.Metadata,
	).Scan(&activity.ID, &activity.Version, &activity.CreatedAt, &activity.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert activity: %w", err)
	}

	return nil
}

// updateActivityWithVersionCheck applies optimistic locking: it only updates
// the row when both id and the caller-supplied activity.Version match,
// incrementing the version on success. Zero rows affected (pgx.ErrNoRows on
// the RETURNING clause) is reported as a domain Conflict, the same
// optimistic-locking convention used across this codebase's other versioned
// repositories.
func (r *PostgresRepository) updateActivityWithVersionCheck(ctx context.Context, activity *Activity) error {
	const query = `
		UPDATE activities
		SET title = $1, type = $2, sequence_order = $3, description = $4,
		    is_ai_generated = $5, metadata = $6, updated_at = NOW(), version = version + 1
		WHERE id = $7 AND version = $8
		RETURNING updated_at, version`

	err := r.db.Pool().QueryRow(ctx, query,
		activity.Title, activity.Type, activity.SequenceOrder, activity.Description,
		activity.IsAIGenerated, activity.Metadata, activity.ID, activity.Version,
	).Scan(&activity.UpdatedAt, &activity.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainerrors.Conflict("activity was modified by another request")
		}
		return fmt.Errorf("update activity: %w", err)
	}

	return nil
}

// DeleteActivity removes the activity with the given id; see the Repository
// interface doc comment.
func (r *PostgresRepository) DeleteActivity(ctx context.Context, id string) error {
	const query = `DELETE FROM activities WHERE id = $1`

	tag, err := r.db.Pool().Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete activity: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.NotFound("activity", id)
	}

	return nil
}
