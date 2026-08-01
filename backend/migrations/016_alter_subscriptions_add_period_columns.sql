-- +goose Up
-- Adds the billing-period columns deliberately deferred by
-- 007_create_subscriptions_table.sql ("current_period_start/end and version
-- omitted (in no design doc)"). specs/008-auth-collaboration-ux/contracts/api.md
-- documents both fields on the register response and elsewhere; this closes
-- that gap (issue #207 Sub-item 2).
--
-- Nullable: a 'stub_pending' subscription (payment not yet confirmed) has no
-- billing period yet. An 'active' subscription gets current_period_start=now
-- and current_period_end=now+30 days at creation time (subscription.Service).
ALTER TABLE subscriptions
    ADD COLUMN current_period_start TIMESTAMP,
    ADD COLUMN current_period_end TIMESTAMP;

-- +goose Down
ALTER TABLE subscriptions
    DROP COLUMN current_period_end,
    DROP COLUMN current_period_start;
