-- +goose Up
-- users table: Spec 004 security/auth data model (004-T006), per
-- specs/004-security-auth-model/data-model.md's "Migration Strategy" section.
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'admin' CHECK (role IN ('admin', 'partner')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

-- +goose Down
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE users;
