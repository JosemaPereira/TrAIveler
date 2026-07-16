-- +goose Up
-- suggestions (008-T017): partner-proposed trip change, pending admin approval.
-- Authority: docs/data-model.md Suggestion + spec.md US3 (suggest-then-approve
-- flow); issue sketch's shape is wrong, unused.
--   - target_id has NO FK: target day/activity may be deleted; suggestion kept
--     for audit history.
--   - reviewed_by ON DELETE SET NULL: keep the review record if the admin is
--     deleted.
CREATE TABLE suggestions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type VARCHAR(20) NOT NULL CHECK (target_type IN ('day', 'activity')),
    target_id UUID NOT NULL,
    suggested_action VARCHAR(20) NOT NULL CHECK (suggested_action IN ('add', 'edit', 'delete')),
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_suggestions_trip_id ON suggestions(trip_id);
CREATE INDEX idx_suggestions_author_id ON suggestions(author_id);
CREATE INDEX idx_suggestions_status ON suggestions(status);

-- +goose Down
DROP INDEX IF EXISTS idx_suggestions_status;
DROP INDEX IF EXISTS idx_suggestions_author_id;
DROP INDEX IF EXISTS idx_suggestions_trip_id;
DROP TABLE suggestions;
