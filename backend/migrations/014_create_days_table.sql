-- +goose Up
-- days (004-T134): single calendar day within a trip, optionally tied to a
-- destination. Authority: docs/data-model.md §Day.
--   - destination_id NULLABLE + SET NULL: arrival/transfer days have no
--     destination, and removing one must not delete the day.
--   - UNIQUE(trip_id, day_number) is scoped per trip, so two trips may both
--     have a day 1. Day numbers must be positive but need not be contiguous.
--   - No `version`: Day is not a versioned entity (Invariant 7).
CREATE TABLE days (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    destination_id UUID REFERENCES destinations(id) ON DELETE SET NULL,
    day_number INT NOT NULL CHECK (day_number > 0),
    label VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_days_trip_day_number UNIQUE (trip_id, day_number)
);

CREATE INDEX idx_days_trip_id ON days(trip_id);
CREATE INDEX idx_days_destination_id ON days(destination_id);

-- +goose Down
DROP INDEX IF EXISTS idx_days_destination_id;
DROP INDEX IF EXISTS idx_days_trip_id;
DROP TABLE days;
