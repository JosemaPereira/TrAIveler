-- +goose Up
-- security_events extension (008-T014): ADD email only (failed-login detection
-- when user_id is unknown); table exists since 004.
-- Authority: specs/008-auth-collaboration-ux/data-model.md SecurityEvent.
-- Not applied: renaming `timestamp` to created_at (breaks
-- observability.LogSecurityEvent, PR #155); spec 008's 9-value event_type CHECK
-- (excludes emitted authz_denied/validation_*/concurrency_conflict) — event_type
-- stays free VARCHAR.
ALTER TABLE security_events
    ADD COLUMN email VARCHAR(255);

CREATE INDEX idx_security_events_email ON security_events(email);

-- +goose Down
DROP INDEX IF EXISTS idx_security_events_email;
ALTER TABLE security_events
    DROP COLUMN email;
