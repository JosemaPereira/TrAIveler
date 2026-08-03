// Package trip persists Trip, Destination, Day, and Activity — the core
// itinerary domain (docs/data-model.md §Trip/§Destination/§Day/§Activity).
// Persistence lives in repository.go (001-T033); business rules (e.g. "only
// admin creates trips") live in Service (001-T035, service.go). The package
// still has no HTTP layer of its own — that is a separate, not-yet-built
// ticket (001-T038/T039).
package trip

import "time"

// Status values accepted by Trip.Status, mirroring the CHECK constraint on
// the trips table (migrations/010_create_trips_table.sql).
const (
	TripStatusDraft     = "draft"
	TripStatusPublished = "published"
)

// Type values accepted by Activity.Type, mirroring the CHECK constraint on
// the activities table (migrations/015_create_activities_table.sql).
const (
	ActivityTypeVisit     = "visit"
	ActivityTypeFood      = "food"
	ActivityTypeLogistics = "logistics"
	ActivityTypeTransfer  = "transfer"
)

// Trip is the top-level container for a planned journey (docs/data-model.md
// §Trip). Description is nullable in the database; a nil pointer means no
// description was set. Version is the optimistic-locking counter enforced by
// UpsertActivity for Activity and, in Service, for Trip's own
// admin-modification updates.
type Trip struct {
	ID          string    `json:"id" db:"id"`
	CreatorID   string    `json:"creator_id" db:"creator_id"`
	Title       string    `json:"title" db:"title"`
	Description *string   `json:"description,omitempty" db:"description"`
	Status      string    `json:"status" db:"status"`
	Archived    bool      `json:"archived" db:"archived"`
	Version     int       `json:"version" db:"version"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Destination is a geographic location shared across trips (docs/data-model.md
// §Destination), referenced by Day. Region is nullable. Destination rows are
// deliberately not deduplicated (see migrations/013_create_destinations_table.sql):
// minor coordinate variations of the same place may coexist.
type Destination struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Country   string    `json:"country" db:"country"`
	Region    *string   `json:"region,omitempty" db:"region"`
	Latitude  float64   `json:"latitude" db:"latitude"`
	Longitude float64   `json:"longitude" db:"longitude"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Day is a single calendar day within a Trip (docs/data-model.md §Day).
// DestinationID is nullable for arrival/transfer days; Label is an optional
// human-readable name (e.g. "Tokyo Day 1"). Day has no version column — it is
// not a versioned entity (Invariant 7 scopes optimistic locking to Trip and
// Activity).
type Day struct {
	ID            string    `json:"id" db:"id"`
	TripID        string    `json:"trip_id" db:"trip_id"`
	DestinationID *string   `json:"destination_id,omitempty" db:"destination_id"`
	DayNumber     int       `json:"day_number" db:"day_number"`
	Label         *string   `json:"label,omitempty" db:"label"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// Activity is a specific event assigned to a Day (docs/data-model.md
// §Activity): a visit, meal, logistics step, or transfer. Description is
// nullable; Metadata is raw JSONB (pricing, booking URLs, ...) left
// unparsed by this package. Version is the optimistic-locking counter
// checked by UpsertActivity when updating an existing row.
type Activity struct {
	ID            string    `json:"id" db:"id"`
	DayID         string    `json:"day_id" db:"day_id"`
	Title         string    `json:"title" db:"title"`
	Type          string    `json:"type" db:"type"`
	SequenceOrder int       `json:"sequence_order" db:"sequence_order"`
	Description   *string   `json:"description,omitempty" db:"description"`
	IsAIGenerated bool      `json:"is_ai_generated" db:"is_ai_generated"`
	Metadata      []byte    `json:"metadata,omitempty" db:"metadata"`
	Version       int       `json:"version" db:"version"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}
