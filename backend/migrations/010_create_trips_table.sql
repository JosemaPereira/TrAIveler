-- +goose Up
-- trips (008-T015): top-level container for a planned journey. Folds 004-T010's
-- optimistic-locking `version` into this CREATE (no separate ALTER).
-- Authority: docs/data-model.md + specs/008-auth-collaboration-ux/data-model.md.
-- `archived` from spec 008; destination/start_date/end_date omitted (the
-- Destination entity landed separately in 013). 004-T011's sibling `version`
-- column is no longer deferred: it ships inline in 015_create_activities_table.sql
-- (Activity, not the never-modeled `itinerary_items`, is the versioned
-- itinerary entity).
CREATE TABLE trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
    archived BOOLEAN NOT NULL DEFAULT FALSE,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trips_creator_id ON trips(creator_id);
CREATE INDEX idx_trips_status ON trips(status);
CREATE INDEX idx_trips_creator_id_archived ON trips(creator_id, archived);

-- +goose Down
DROP INDEX IF EXISTS idx_trips_creator_id_archived;
DROP INDEX IF EXISTS idx_trips_status;
DROP INDEX IF EXISTS idx_trips_creator_id;
DROP TABLE trips;
