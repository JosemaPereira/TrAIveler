-- +goose Up
-- subscriptions (008-T012): links a user to a plan.
-- Authority: docs/data-model.md + specs/008-auth-collaboration-ux/data-model.md.
-- grace_period_ends_at + status 'expired' from spec 008; current_period_start/
-- end and version omitted (in no design doc).
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL CHECK (status IN ('stub_pending', 'active', 'cancelled', 'expired')),
    stub_payment_ref VARCHAR(100),
    grace_period_ends_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- One subscription per user (docs/data-model.md).
CREATE UNIQUE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);

-- +goose Down
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_user_id;
DROP TABLE subscriptions;
