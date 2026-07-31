-- +goose Up
-- destinations (004-T133): geographic location shared across trips, referenced
-- by days. Authority: docs/data-model.md §Destination.
--   - No `version`: Destination is reference data, not a versioned entity
--     (Invariant 7 scopes optimistic locking to Trip and Activity).
--   - No UNIQUE on (name, country, lat, lng): deliberately unenforced, so
--     minor coordinate variations of the same place can coexist.
--   - idx_destinations_coordinates is a plain composite B-tree, not a spatial
--     index — the data model conditions that on PostGIS, which is not enabled.
CREATE TABLE destinations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    country VARCHAR(2) NOT NULL CHECK (LENGTH(country) = 2),
    region VARCHAR(100),
    latitude DECIMAL(9, 6) NOT NULL CHECK (latitude BETWEEN -90 AND 90),
    longitude DECIMAL(9, 6) NOT NULL CHECK (longitude BETWEEN -180 AND 180),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_destinations_country ON destinations(country);
CREATE INDEX idx_destinations_coordinates ON destinations(latitude, longitude);

-- +goose Down
DROP INDEX IF EXISTS idx_destinations_coordinates;
DROP INDEX IF EXISTS idx_destinations_country;
DROP TABLE destinations;
