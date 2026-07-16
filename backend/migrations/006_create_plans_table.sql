-- +goose Up
-- plans (docs/data-model.md Plan): subscription tier governing account limits.
-- Not a Spec 008 task but MUST precede subscriptions (007), whose plan_id FKs
-- here — goose.Up fails on a migration referencing a not-yet-created table
-- (patterns-discovered.md "Goose Migrations Must Not Reference ... Later ... Set").
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    max_admin_users INTEGER NOT NULL DEFAULT 1 CHECK (max_admin_users > 0),
    max_partner_users INTEGER NOT NULL DEFAULT 1 CHECK (max_partner_users >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed the only MVP plan (docs/data-model.md "Seed Data"); fixed UUID for
-- deterministic reference from app code and tests.
INSERT INTO plans (id, name, max_admin_users, max_partner_users)
VALUES ('00000000-0000-0000-0000-000000000001', 'basic', 1, 1);

-- +goose Down
DROP TABLE plans;
