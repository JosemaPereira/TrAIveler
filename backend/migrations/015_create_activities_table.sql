-- +goose Up
-- activities (004-T135): event assigned to a day, covering every content type
-- (visit, food, logistics, transfer). Authority: docs/data-model.md §Activity
-- and Invariant 7.
--   - Folds 004-T011's optimistic-locking `version` into this CREATE (no
--     separate ALTER), as 010_create_trips_table.sql did for 004-T010.
--     Activity — not the never-modeled `itinerary_items` of T011's task text —
--     is the versioned itinerary entity; see the "Naming correction" note in
--     specs/004-security-auth-model/tasks.md.
--   - sequence_order is positive but NOT unique within a day: the data model
--     leaves reordering to the service layer.
--   - metadata JSONB extends the row (pricing, booking URLs) without a migration.
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    day_id UUID NOT NULL REFERENCES days(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('visit', 'food', 'logistics', 'transfer')),
    sequence_order INT NOT NULL CHECK (sequence_order > 0),
    description TEXT,
    is_ai_generated BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activities_day_id ON activities(day_id);

-- +goose Down
DROP INDEX IF EXISTS idx_activities_day_id;
DROP TABLE activities;
