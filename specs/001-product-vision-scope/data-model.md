# Data Model: Product Vision and Scope

**Date**: 2026-07-02 | **Plan**: [plan.md](plan.md) | **Spec**: [spec.md](spec.md)

All entities derived from the Key Entities section of the spec. PostgreSQL 16 schema.

---

## Entity Map

```
Plan ──< Subscription >── User ──< Trip ──< Day ──< Activity
                           │          │
                           │          ├──< Collaborator >── User
                           │          ├──< Suggestion
                           │          ├──< TripTravelStyle >── TravelStyle
                           │          └──< ConversationSession ──< ConversationMessage
                           └── (role: admin | partner)
```

---

## Entities

### Plan

Subscription tier governing account limits. Only the `basic` plan exists in the POC.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `gen_random_uuid()` | |
| `name` | `varchar(64)` | NOT NULL, UNIQUE | e.g., `basic` |
| `max_admin_users` | `int` | NOT NULL, DEFAULT 1 | Seats per subscription |
| `max_partner_users` | `int` | NOT NULL, DEFAULT 1 | Collaborator seats |
| `created_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |

**Validation rules**: `max_admin_users >= 1`, `max_partner_users >= 0`.

---

### Subscription

Links a user account to a plan. Tracks mock payment reference for future real integration.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `gen_random_uuid()` | |
| `user_id` | `uuid` | NOT NULL, FK → `users.id` ON DELETE CASCADE | The subscribing admin |
| `plan_id` | `uuid` | NOT NULL, FK → `plans.id` | |
| `status` | `varchar(32)` | NOT NULL, DEFAULT `active` | `active` \| `cancelled` \| `stub_pending` |
| `stub_payment_ref` | `varchar(128)` | NULLABLE | Provider transaction ID from stub |
| `created_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |
| `updated_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |

**State transitions**: `stub_pending → active` (on stub checkout confirmation), `active → cancelled`.

---

### User

Authenticated account. Role determines permission boundary across the product.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `gen_random_uuid()` | |
| `email` | `varchar(320)` | NOT NULL, UNIQUE | Max RFC 5321 address length |
| `password_hash` | `text` | NOT NULL | bcrypt, min cost 12 |
| `role` | `varchar(16)` | NOT NULL | `admin` \| `partner` |
| `subscription_id` | `uuid` | NULLABLE, FK → `subscriptions.id` | NULL until checkout complete |
| `created_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |
| `updated_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |

**Validation rules**: `email` must pass RFC 5322 format validation at application layer.
**Security**: `password_hash` is never returned in any API response.

---

### Trip

Top-level container for a planned journey. Owned by one admin user.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `gen_random_uuid()` | |
| `creator_id` | `uuid` | NOT NULL, FK → `users.id` | Must have role `admin` |
| `title` | `varchar(256)` | NOT NULL | User-facing name |
| `description` | `text` | NULLABLE | Optional context |
| `status` | `varchar(16)` | NOT NULL, DEFAULT `draft` | `draft` \| `published` |
| `created_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |
| `updated_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |

**State transitions**: `draft → published`. Only admin can publish.

---

### Destination

A location (city or country) within a Trip, ordered by sequence.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `gen_random_uuid()` | |
| `trip_id` | `uuid` | NOT NULL, FK → `trips.id` ON DELETE CASCADE | |
| `name` | `varchar(256)` | NOT NULL | City or country name |
| `sequence_order` | `int` | NOT NULL | Display order within trip |
| `duration_days` | `int` | NOT NULL, CHECK > 0 | Time allocated at destination |
| `created_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |

---

### Day

A single calendar day within a Trip, belonging to a Destination.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `gen_random_uuid()` | |
| `trip_id` | `uuid` | NOT NULL, FK → `trips.id` ON DELETE CASCADE | |
| `destination_id` | `uuid` | NULLABLE, FK → `destinations.id` | NULL = arrival/transfer day |
| `day_number` | `int` | NOT NULL, CHECK > 0 | 1-indexed position in trip |
| `label` | `varchar(128)` | NULLABLE | e.g., "Arrival Day", "Tokyo Day 1" |
| `created_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |

**Constraint**: UNIQUE(`trip_id`, `day_number`).

---

### Activity

A specific event assigned to a Day. Covers all content types in one table discriminated by `type`.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `gen_random_uuid()` | |
| `day_id` | `uuid` | NOT NULL, FK → `days.id` ON DELETE CASCADE | |
| `title` | `varchar(256)` | NOT NULL | |
| `type` | `varchar(16)` | NOT NULL | `visit` \| `food` \| `logistics` \| `transfer` |
| `sequence_order` | `int` | NOT NULL | Display order within day |
| `description` | `text` | NULLABLE | Detail from AI or user |
| `is_ai_generated` | `boolean` | NOT NULL, DEFAULT true | False if manually added |
| `metadata` | `jsonb` | NULLABLE | Extensible bag for future fields |
| `created_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |
| `updated_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |

---

### TravelStyle

Lookup table for available travel styles.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `gen_random_uuid()` | |
| `slug` | `varchar(64)` | NOT NULL, UNIQUE | e.g., `gastronomy`, `sports` |
| `label` | `varchar(128)` | NOT NULL | Human-readable label |

**Seed values**: `gastronomy`, `sports`, `technology`, `museums-and-art`, `film-audiovisual`.

---

### TripTravelStyle

Many-to-many join between Trip and TravelStyle (a trip may carry multiple styles per traveler profile).

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `trip_id` | `uuid` | NOT NULL, FK → `trips.id` ON DELETE CASCADE | |
| `travel_style_id` | `uuid` | NOT NULL, FK → `travel_styles.id` | |
| `assigned_by` | `uuid` | NOT NULL, FK → `users.id` | Which traveler added this style |

**PK**: (`trip_id`, `travel_style_id`, `assigned_by`).

---

### Collaborator

Explicit partner invite granting view and suggestion rights on a Trip.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `gen_random_uuid()` | |
| `trip_id` | `uuid` | NOT NULL, FK → `trips.id` ON DELETE CASCADE | |
| `user_id` | `uuid` | NOT NULL, FK → `users.id` | Must have role `partner` |
| `invited_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |
| `accepted_at` | `timestamptz` | NULLABLE | NULL = invitation pending |

**Constraint**: UNIQUE(`trip_id`, `user_id`).
**Business rule**: A basic plan allows at most 1 active Collaborator per Trip; enforced in service layer.

---

### Suggestion

A proposed modification submitted by a partner collaborator; requires admin approval.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `gen_random_uuid()` | |
| `trip_id` | `uuid` | NOT NULL, FK → `trips.id` ON DELETE CASCADE | |
| `author_id` | `uuid` | NOT NULL, FK → `users.id` | Must be a partner collaborator on the trip |
| `target_type` | `varchar(16)` | NOT NULL | `trip` \| `day` \| `activity` |
| `target_id` | `uuid` | NULLABLE | ID of the targeted entity; NULL for trip-level |
| `content` | `text` | NOT NULL | Free-form description of the proposed change |
| `status` | `varchar(16)` | NOT NULL, DEFAULT `pending` | `pending` \| `approved` \| `rejected` |
| `resolved_by` | `uuid` | NULLABLE, FK → `users.id` | Admin who approved/rejected |
| `resolved_at` | `timestamptz` | NULLABLE | |
| `created_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |

**Retention**: Records are NEVER hard-deleted; status history is preserved for the lifetime of
the trip (FR-011).
**State transitions**: `pending → approved`, `pending → rejected`. Terminal states are final.

---

### ConversationSession

Tracks the multi-turn AI conversation that leads to a Trip's itinerary generation (FR-012).

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `gen_random_uuid()` | |
| `trip_id` | `uuid` | NOT NULL, FK → `trips.id` ON DELETE CASCADE | |
| `user_id` | `uuid` | NOT NULL, FK → `users.id` | Initiating admin |
| `status` | `varchar(16)` | NOT NULL, DEFAULT `active` | `active` \| `completed` \| `abandoned` |
| `created_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |
| `updated_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |

---

### ConversationMessage

Individual turn in a ConversationSession (user message or AI response/question).

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `gen_random_uuid()` | |
| `session_id` | `uuid` | NOT NULL, FK → `conversation_sessions.id` ON DELETE CASCADE | |
| `role` | `varchar(16)` | NOT NULL | `user` \| `assistant` |
| `content` | `text` | NOT NULL | Message body |
| `sequence_order` | `int` | NOT NULL | Monotonically increasing per session |
| `created_at` | `timestamptz` | NOT NULL, DEFAULT `now()` | |

**Constraint**: UNIQUE(`session_id`, `sequence_order`).

---

## Indexes

```sql
-- Frequent lookups
CREATE INDEX idx_trips_creator_id      ON trips(creator_id);
CREATE INDEX idx_days_trip_id          ON days(trip_id);
CREATE INDEX idx_activities_day_id     ON activities(day_id);
CREATE INDEX idx_suggestions_trip_id   ON suggestions(trip_id);
CREATE INDEX idx_suggestions_status    ON suggestions(trip_id, status);
CREATE INDEX idx_collaborators_user_id ON collaborators(user_id);
CREATE INDEX idx_conversation_messages_session ON conversation_messages(session_id, sequence_order);
```

---

## Edge Case Handling (Data Layer)

| Edge Case | Resolution |
|-----------|-----------|
| Admin modifies same section as pending suggestion | Suggestion preserved as `pending`; service layer flags it as `potentially_stale` via a computed field (not stored) — resolved by admin on review |
| Admin deletes a Day/Activity with pending suggestions | `target_id` FK is nullable; orphaned suggestions retain their content and status but `target_id` becomes unresolvable — the UI surfaces them as "target deleted" |
| Partner attempts direct write without suggestion flow | Enforced in service layer: any non-suggestion mutation by a `partner`-role user returns `403 Forbidden` |
| Subscription limit exceeded (> 1 collaborator) | Enforced in `subscription/service.go` before INSERT into `collaborators`; returns `409 Conflict` |
