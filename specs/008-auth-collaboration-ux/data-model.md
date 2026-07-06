# Data Model: Authentication & Collaboration UX

**Feature**: Authentication & Collaboration UX  
**Date**: 2026-07-06  
**Phase**: 1 (Design)

## Overview

This document defines the data model entities, fields, relationships, and validation rules for authentication, session management, subscription lifecycle, and collaboration workflows. It extends the foundational entities in docs/data-model.md with new fields, new entities, and detailed validation rules.

**Scope**: This data model supports Paid User (subscription-based) and Free User (no-payment collaborator) tiers with single-collaboration limit for Free Users.

---

## Entity Catalog

### New Entities

#### PasswordResetToken

Represents a single-use token for password reset requests. Tokens are cryptographically random, hashed before storage, and marked as used after redemption.

**Attributes**:

| Attribute | Type | Constraints | Description |
|-----------|------|-------------|-------------|
| `id` | UUID | PK, NOT NULL | Primary key |
| `user_id` | UUID | FK(users.id), NOT NULL, ON DELETE CASCADE | User requesting password reset |
| `token_hash` | TEXT | NOT NULL, UNIQUE | SHA-256 hash of reset token |
| `expires_at` | TIMESTAMP | NOT NULL | Token expiration (created_at + 1 hour) |
| `used_at` | TIMESTAMP | NULL | Timestamp when token was redeemed (NULL = unused) |
| `created_at` | TIMESTAMP | NOT NULL, DEFAULT NOW() | Token creation time |

**Validation Rules**:
- **[DB]** `expires_at > created_at`: Expiration must be after creation
- **[DB]** `used_at IS NULL OR used_at >= created_at`: Used timestamp must be after creation if set
- **[Logic]** Token generation: 32 bytes cryptographically random (crypto.randombytes), base64url encoded
- **[Logic]** Token hashing: SHA-256 before storage (never store plaintext)
- **[Logic]** Token validation: Requires `expires_at > NOW() AND used_at IS NULL`
- **[API]** One active reset token per user at a time (previous tokens invalidated on new request)

**Indexes**:
- `idx_password_reset_tokens_user_id` on `user_id` (lookup active tokens for user)
- `idx_password_reset_tokens_token_hash` on `token_hash` (validation lookup)
- `idx_password_reset_tokens_expires_at` on `expires_at` (cleanup expired tokens)

**State Transitions**:
```
[created] → [used] (on successful password reset)
[created] → [expired] (after 1 hour, cleanup job)
```

**Business Rules**:
- BR-001: Only one active (unexpired, unused) token per user
- BR-002: Token must be used within 1 hour of creation
- BR-003: Used tokens cannot be reused (replay prevention)
- BR-004: Previous unused tokens invalidated when new token requested

---

#### SecurityEvent

Represents structured security event logs for authentication and authorization actions. Enables incident investigation, brute-force detection, and audit trails.

**Attributes**:

| Attribute | Type | Constraints | Description |
|-----------|------|-------------|-------------|
| `id` | UUID | PK, NOT NULL | Primary key |
| `correlation_id` | UUID | NOT NULL, INDEX | Request ID for tracing multi-step flows |
| `event_type` | TEXT | NOT NULL, CHECK(...) | Event type (see enum below) |
| `user_id` | UUID | FK(users.id), NULL, ON DELETE SET NULL | User involved (NULL for registration/failed login) |
| `email` | TEXT | NULL | Email for failed login attempts (when user_id unknown) |
| `severity` | TEXT | NOT NULL, CHECK(severity IN ('info', 'warning', 'error')) | Event severity |
| `ip_address` | TEXT | NULL | Client IP address |
| `user_agent` | TEXT | NULL | Client User-Agent header |
| `details` | JSONB | NULL | Additional event-specific data (no sensitive info) |
| `created_at` | TIMESTAMP | NOT NULL, DEFAULT NOW() | Event timestamp |

**Event Type Enum**:
```
'auth_registration'          # User registration completed
'auth_login_success'         # Successful login
'auth_login_failure'         # Failed login (incorrect password, nonexistent user)
'auth_password_reset_request'  # Password reset email sent
'auth_password_reset_complete' # Password successfully reset via token
'auth_password_change'       # Password changed (authenticated user)
'auth_token_refresh'         # Access token refreshed via refresh token
'auth_logout'                # User logged out (refresh token revoked)
'auth_session_expired'       # Session expired (access token not refreshed)
```

**Validation Rules**:
- **[DB]** `event_type` must be one of the enum values above
- **[DB]** `severity` must be 'info', 'warning', or 'error'
- **[Logic]** `user_id` OR `email` required (at least one must be set)
- **[Logic]** `details` JSONB must NOT contain: password, password_hash, token, refresh_token, jwt
- **[API]** Write-only (no read API; accessed via CloudWatch Logs)

**Indexes**:
- `idx_security_events_correlation_id` on `correlation_id` (trace requests)
- `idx_security_events_user_id` on `user_id` (user activity audit)
- `idx_security_events_email` on `email` (failed login pattern detection)
- `idx_security_events_created_at` on `created_at` (time-range queries)
- `idx_security_events_event_type` on `event_type` (event type filtering)

**Business Rules**:
- BR-001: All authentication/authorization actions must be logged
- BR-002: No sensitive data (passwords, tokens, PII beyond user_id/email) in `details` field
- BR-003: Correlation ID must match request ID from incoming request header

---

### Modified Entities

#### User (extension of docs/data-model.md)

**New Fields**:

| Attribute | Type | Constraints | Description |
|-----------|------|-------------|-------------|
| `failed_login_attempts` | INTEGER | NOT NULL, DEFAULT 0 | Failed login count for progressive delay rate limiting |
| `last_failed_login_at` | TIMESTAMP | NULL | Timestamp of most recent failed login |
| `email_verified` | BOOLEAN | NOT NULL, DEFAULT FALSE | Email verification status (for future MFA) |
| `email_verification_token_hash` | TEXT | NULL, UNIQUE | SHA-256 hash of email verification token (for future) |

**Modified Validation Rules**:
- **[Logic]** Rate limiting: If `failed_login_attempts >= 5` within 15 minutes, apply progressive delay: `2^(failed_login_attempts - 5)` seconds
- **[Logic]** Reset `failed_login_attempts = 0` on successful login
- **[Logic]** Increment `failed_login_attempts` on failed login, set `last_failed_login_at = NOW()`

**Indexes** (new):
- `idx_users_email` on `email` (already exists, used for login lookup)
- `idx_users_failed_login_attempts` on `failed_login_attempts, last_failed_login_at` (rate limiting queries)

---

#### Subscription (extension of docs/data-model.md)

**New Fields**:

| Attribute | Type | Constraints | Description |
|-----------|------|-------------|-------------|
| `grace_period_ends_at` | TIMESTAMP | NULL | End of 30-day grace period (set on cancellation) |
| `cancelled_at` | TIMESTAMP | NULL | Timestamp when subscription was cancelled |

**Modified Validation Rules**:
- **[Logic]** On cancellation: Set `status='cancelled'`, `cancelled_at=NOW()`, `grace_period_ends_at=NOW() + INTERVAL '30 days'`, `User.has_subscription=false`
- **[Logic]** During grace period (`NOW() BETWEEN cancelled_at AND grace_period_ends_at`): Owned trips visible but read-only (403 on PUT/DELETE)
- **[Logic]** After grace period (`NOW() > grace_period_ends_at`): Scheduled job sets `Trip.archived=true` for all owned trips
- **[Logic]** On resubscription: Set `status='active'`, `cancelled_at=NULL`, `grace_period_ends_at=NULL`, `User.has_subscription=true`, `Trip.archived=false` for owned trips

**State Transitions** (extended):
```
[active] → [cancelled] (user cancels subscription)
[cancelled] → [active] (user resubscribes during grace period)
[cancelled] → [expired] (grace period ends, archival triggered)
[expired] → [active] (user resubscribes after archival)
```

**Business Rules**:
- BR-001: Grace period is exactly 30 days from cancellation
- BR-002: During grace period, owned trips are read-only (collaborations as Free User continue)
- BR-003: After grace period, owned trips archived (not deleted)
- BR-004: Resubscription restores archived trips (`archived=false`)

---

#### Trip (extension of docs/data-model.md)

**New Fields**:

| Attribute | Type | Constraints | Description |
|-----------|------|-------------|-------------|
| `archived` | BOOLEAN | NOT NULL, DEFAULT FALSE | True if trip archived due to subscription lapse |

**Modified Validation Rules**:
- **[Logic]** Archived trips: Read-only for creator, inaccessible for collaborators (404 response)
- **[Logic]** Archival trigger: Scheduled job sets `archived=true` for trips where `creator_id IN (SELECT user_id FROM subscriptions WHERE grace_period_ends_at < NOW())`
- **[Logic]** Unarchive: On resubscription, `UPDATE trips SET archived=false WHERE creator_id = ? AND archived=true`
- **[API]** Archived trips: Excluded from list endpoints by default (optional `include_archived=true` query param for creator only)

**Indexes** (new):
- `idx_trips_creator_id_archived` on `creator_id, archived` (archival queries)

---

#### Collaborator (extension of docs/data-model.md)

**Modified Validation Rules**:
- **[Logic]** Free User single-collaboration limit: `SELECT COUNT(*) FROM collaborators WHERE user_id = ? AND status != 'rejected'` must be <= 1 for Free Users
- **[Logic]** Before accepting invitation: Check collaboration limit, if limit reached, display "Leave current trip first"
- **[Logic]** Paid Users: No collaboration limit check
- **[Logic]** On Free User upgrade: Leave Collaborator records unchanged (collaboration continues)

**Business Rules**:
- BR-001: Free Users can collaborate on at most 1 trip at a time
- BR-002: Paid Users have no collaboration limit
- BR-003: Upgrading from Free to Paid preserves existing collaborations

---

## Entity Relationships

```
User 1 ───> * PasswordResetToken (one-to-many, cascade delete)
User 1 ───> * SecurityEvent (one-to-many, set null on delete)
User 1 ───> 0..1 Subscription (one-to-optional-one)
Subscription 1 ───> * Plan (many-to-one, FK constraint)
User 1 ───> * Trip (one-to-many as creator)
User 1 ───> * Collaborator (one-to-many as collaborator, max 1 for Free Users)
Trip 1 ───> * Collaborator (one-to-many)
Trip 1 ───> * Suggestion (one-to-many)
```

---

## Database Migration Strategy

**Migration Files** (in order):
1. `001_create_users.sql`: Create users table (already exists from spec 006)
2. `002_create_subscriptions.sql`: Create subscriptions, plans tables (already exists)
3. `003_create_password_reset_tokens.sql`: **NEW** - Create password_reset_tokens table
4. `004_create_security_events.sql`: **NEW** - Create security_events table
5. `005_add_user_rate_limiting_fields.sql`: **NEW** - Add failed_login_attempts, last_failed_login_at to users
6. `006_add_user_email_verification_fields.sql`: **NEW** - Add email_verified, email_verification_token_hash to users
7. `007_add_subscription_grace_period_fields.sql`: **NEW** - Add grace_period_ends_at, cancelled_at to subscriptions
8. `008_add_trip_archived_field.sql`: **NEW** - Add archived to trips

**Forward-Only Strategy**: All migrations are additive (ADD COLUMN, CREATE INDEX) or backward-compatible. No DROP operations in production until dependent code removed.

---

## Validation Rule Tags

- **[DB]**: Database-level constraint (CHECK, NOT NULL, UNIQUE, FK)
- **[Logic]**: Business logic validation in application code (before database write)
- **[API]**: API-level validation (before business logic, e.g., request schema validation)

---

## GDPR Compliance

**Right to Erasure (User Deletion)**:
- PasswordResetToken: CASCADE delete when User deleted
- SecurityEvent: SET NULL for user_id when User deleted (preserve audit trail with anonymized user_id)
- Subscription: CASCADE delete (or archive with user_id anonymized)
- Collaborator: CASCADE delete (or replace with "[Deleted User]" display)

**Right to Access**:
- User can request all SecurityEvent records where `user_id = ?`
- User can request all PasswordResetToken records (show only metadata: created_at, used_at, expires_at; never token_hash)

---

## Performance Considerations

**High-Frequency Queries**:
1. Login validation: `SELECT * FROM users WHERE email = ?` (indexed)
2. Rate limiting check: `SELECT failed_login_attempts, last_failed_login_at FROM users WHERE email = ?` (indexed)
3. Password reset token validation: `SELECT * FROM password_reset_tokens WHERE token_hash = ? AND expires_at > NOW() AND used_at IS NULL` (indexed)
4. Free User collaboration limit: `SELECT COUNT(*) FROM collaborators WHERE user_id = ? AND status != 'rejected'` (indexed)
5. Grace period check: `SELECT grace_period_ends_at FROM subscriptions WHERE user_id = ?` (indexed)

**Optimization Strategies**:
- In-memory cache for rate limiting (Redis with 15-minute TTL, fallback to DB)
- Scheduled job for trip archival (runs nightly, not on-demand during user request)
- SecurityEvent writes are async (buffered in-memory, batch insert every 10 seconds or 100 events)

---

## Summary

**New Entities**: PasswordResetToken, SecurityEvent  
**Modified Entities**: User (+4 fields), Subscription (+2 fields), Trip (+1 field), Collaborator (validation rules)  
**Total Fields Added**: 7 new fields across existing entities  
**Total Indexes Added**: 8 new indexes for performance  
**Migration Files**: 6 new migrations (003-008)

All entities support the 6 user stories in spec.md with explicit validation rules for authentication security, subscription lifecycle management, and collaboration constraints.
