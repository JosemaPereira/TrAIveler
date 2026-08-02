-- Local-dev fixture data for manual browser testing.
--
-- NOT a goose migration on purpose: this is throwaway fixture data, not schema,
-- and must never run against a CI testcontainer DB or a real (staging/production)
-- database. Run it by hand (`make seed` from backend/, or `psql "$DATABASE_URL" -f
-- scripts/seed_local_dev.sql`) against your local Postgres only.
--
-- Idempotent: every INSERT is ON CONFLICT DO NOTHING against a real unique
-- constraint, so re-running this script after a fresh `goose up` is always safe.
--
-- Creates one paying user (active subscription) and one free user, plus a Trip
-- owned by the paying user with the free user invited as an accepted
-- collaborator -- the "paid user invites a free companion" scenario.
--
-- Login credentials (password validation: 8-72 chars, upper+lower+digit --
-- these bcrypt hashes at cost 12 were generated with the exact same
-- golang.org/x/crypto/bcrypt call backend/internal/auth/password.go uses):
--   paying@traiveler.local  / PayingUser123
--   free@traiveler.local    / FreeUser123
--
-- Caveat: as of Sprint 8 planning, DashboardPage/TripDetailPage are still
-- hardcoded placeholders (no useTrips/GET-/trips wiring yet), so the seeded
-- Trip/Collaborator rows will NOT render in the browser until that ships --
-- they exist now so no second seeding pass is needed once it does. What IS
-- visible today: logging in as either user, and the "Subscriber" badge in
-- Navigation, which is driven by users.has_subscription.

-- Paying user: admin role, active subscription.
INSERT INTO users (
    id, email, password_hash, role, full_name, has_subscription, email_verified
) VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'paying@traiveler.local',
    '$2a$12$IOK7LUu.NFxX1yeZAe/W9uom3xh7iEs/JRo1Gduy5vOoxXUOAv7Ka',
    'admin',
    'Paying Traveler',
    TRUE,
    TRUE
) ON CONFLICT (email) DO NOTHING;

-- Free user: partner role, no subscription of their own -- gets invited as a
-- collaborator on the paying user's trip below.
INSERT INTO users (
    id, email, password_hash, role, full_name, has_subscription, email_verified
) VALUES (
    'a0000000-0000-0000-0000-000000000002',
    'free@traiveler.local',
    '$2a$12$GC./fmWUBv5k3W/7SBQkSuLkXdXBEqVeNC8A6jv1D5VEk/O4dM1QC',
    'partner',
    'Free Companion',
    FALSE,
    TRUE
) ON CONFLICT (email) DO NOTHING;

-- Active subscription for the paying user, on the single seeded 'basic' plan
-- (migrations/006_create_plans_table.sql).
INSERT INTO subscriptions (
    id, user_id, plan_id, status, stub_payment_ref,
    current_period_start, current_period_end
) VALUES (
    'b0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'active',
    'seed-local-dev-stub-payment',
    NOW(),
    NOW() + INTERVAL '30 days'
) ON CONFLICT (user_id) DO NOTHING;

-- Sample trip owned by the paying user (Sprint 8 backend/UI not wired to real
-- data yet -- see the caveat above).
INSERT INTO trips (
    id, creator_id, title, description, status, archived
) VALUES (
    'c0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001',
    'Sample Trip - Paris Getaway',
    'Seed trip for local browser testing.',
    'draft',
    FALSE
) ON CONFLICT (id) DO NOTHING;

-- Free user invited to (and already accepted onto) the paying user's trip.
INSERT INTO collaborators (
    id, trip_id, user_id, email, status, accepted_at
) VALUES (
    'd0000000-0000-0000-0000-000000000001',
    'c0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000002',
    'free@traiveler.local',
    'accepted',
    NOW()
) ON CONFLICT (trip_id, user_id) DO NOTHING;
