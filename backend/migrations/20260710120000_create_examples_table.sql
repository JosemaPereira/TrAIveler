-- +goose Up
-- Reference table for internal/example, the canonical layered-pattern demo.
-- Timestamp-versioned (not sequential 001-016) because those numbers are
-- reserved for the real foundational domain tables (docs/data-model.md
-- "Database Migrations"); this table is explicitly throwaway/reference-only.
-- Column style (VARCHAR + CHECK for enum-like fields, TIMESTAMP DEFAULT
-- NOW(), gen_random_uuid() PK) mirrors the concrete DDL precedent in
-- specs/004-security-auth-model/data-model.md rather than a native Postgres
-- ENUM type, for easier future ALTERs.
CREATE TABLE examples (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 1
);

CREATE INDEX idx_examples_status ON examples(status);

-- +goose Down
DROP TABLE examples;
