-- +goose Up
-- collaborators (008-T016): partner invite granting view + suggestion rights.
-- Authority: docs/data-model.md + specs/008-auth-collaboration-ux/data-model.md
-- + spec.md US3 (invite-by-email). Decisions (issue #142):
--   - user_id NULLABLE: US3 inserts a 'pending' invite before the invitee
--     registers, so no user row exists yet.
--   - email + status added for invite-by-email + the Free-User 1-collab limit.
--   - both UNIQUEs: (trip_id, user_id) is canonical but NULL user_ids don't
--     collide in Postgres, so (trip_id, email) closes the duplicate-invite gap.
CREATE TABLE collaborators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    invited_at TIMESTAMP NOT NULL DEFAULT NOW(),
    accepted_at TIMESTAMP,
    CONSTRAINT uq_collaborators_trip_user UNIQUE (trip_id, user_id),
    CONSTRAINT uq_collaborators_trip_email UNIQUE (trip_id, email)
);

CREATE INDEX idx_collaborators_trip_id ON collaborators(trip_id);
CREATE INDEX idx_collaborators_user_id ON collaborators(user_id);
CREATE INDEX idx_collaborators_status ON collaborators(status);

-- +goose Down
DROP INDEX IF EXISTS idx_collaborators_status;
DROP INDEX IF EXISTS idx_collaborators_user_id;
DROP INDEX IF EXISTS idx_collaborators_trip_id;
DROP TABLE collaborators;
