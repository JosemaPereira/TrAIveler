-- +goose Up
-- users extension (008-T011): ADD columns only; users exists since 001.
-- Authority: docs/data-model.md + specs/008-auth-collaboration-ux/data-model.md.
-- Omitted: email_verification_token_hash (unused in MVP); subscription_id FK
-- (redundant with subscriptions.user_id UNIQUE + has_subscription).
ALTER TABLE users
    ADD COLUMN full_name VARCHAR(255),
    ADD COLUMN has_subscription BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN failed_login_attempts INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN last_failed_login_at TIMESTAMP,
    ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN version BIGINT NOT NULL DEFAULT 1;

CREATE INDEX idx_users_failed_login_attempts ON users(failed_login_attempts, last_failed_login_at);

-- +goose Down
DROP INDEX IF EXISTS idx_users_failed_login_attempts;
ALTER TABLE users
    DROP COLUMN version,
    DROP COLUMN email_verified,
    DROP COLUMN last_failed_login_at,
    DROP COLUMN failed_login_attempts,
    DROP COLUMN has_subscription,
    DROP COLUMN full_name;
