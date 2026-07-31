# REST API Contract

**Date**: 2026-07-02 | **Plan**: [plan.md](../plan.md) | **Data Model**: [data-model.md](../data-model.md)

## Conventions

- Base path: `/api/v1`
- All request and response bodies: `application/json`
- Authentication: JWT in HTTP-only cookie `auth_token` (set on login, cleared on logout)
- Error shape (all error responses):
  ```json
  { "error": { "code": "TRIP_NOT_FOUND", "message": "human-readable detail" } }
  ```
- Timestamps: ISO 8601 (`2026-07-02T15:04:05Z`)
- IDs: UUID v4 strings

---

## Auth

### `POST /api/v1/auth/register`

Register a new admin account and begin stub checkout to assign the basic plan.

**Request**
```json
{
  "email": "user@example.com",
  "password": "min8chars"
}
```

**Response `201`**
```json
{
  "user": { "id": "uuid", "email": "user@example.com", "role": "admin" },
  "checkout_url": "/subscribe/checkout"
}
```

**Errors**: `400 VALIDATION_ERROR`, `409 EMAIL_ALREADY_EXISTS`

---

### `POST /api/v1/auth/login`

Authenticate and set `auth_token` cookie.

**Request**
```json
{ "email": "user@example.com", "password": "..." }
```

**Response `200`**
```json
{ "user": { "id": "uuid", "email": "user@example.com", "role": "admin" } }
```

**Errors**: `400 VALIDATION_ERROR`, `401 INVALID_CREDENTIALS`

---

### `POST /api/v1/auth/logout`

Clear `auth_token` cookie.

**Response `204`** (no body)

---

### `GET /api/v1/auth/me`

Return the currently authenticated user.

> **As-built note (2026-07-31, 001-T021)**: this endpoint now exists
> (`auth.Handler.handleCurrentUser`), and the shipped shape differs from the example below. The
> generated contract in `backend/docs/` is the authority; do not "fix" the code to match this
> section. Differences, all deliberate:
>
> | This document | Actually shipped |
> |---------------|------------------|
> | cookie `auth_token` | `access_token` (HTTP-only, Secure, SameSite=Strict) |
> | `user.role` | omitted — `auth.User.Role` carries `json:"-"`, so no client can make an authorization decision from this body |
> | `user.subscription: { status, plan }` | flattened to a `has_subscription` boolean, read live from the `users` row rather than from the access token's claim |
> | (absent) | `full_name`, `created_at` |
> | `401 UNAUTHENTICATED` | `401 authentication_required` — same snake_case correction already recorded in [`specs/008-auth-collaboration-ux/contracts/api.md`](../../008-auth-collaboration-ux/contracts/api.md); `docs/api-design-standards.md` §7 is the authority |
>
> A token whose subject no longer exists returns that same `401`, not a `404`: the session is dead,
> not the resource. An expired-but-validly-signed token is the one 401 that instead carries
> `token_expired`, the frontend's silent-refresh trigger (008-T150).

**Response `200`**
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "role": "admin",
    "subscription": { "status": "active", "plan": "basic" }
  }
}
```

**Errors**: `401 UNAUTHENTICATED`

---

## Subscriptions (Mock Checkout)

### `GET /api/v1/subscription/plans`

List available plans.

**Response `200`**
```json
{
  "plans": [
    {
      "id": "uuid",
      "name": "basic",
      "label": "Basic Plan",
      "max_admin_users": 1,
      "max_partner_users": 1
    }
  ]
}
```

---

### `POST /api/v1/subscription/checkout`

Initiate a stub checkout session. Always succeeds in the POC.

**Request**
```json
{ "plan_id": "uuid" }
```

**Response `200`**
```json
{
  "checkout_session_id": "stub-session-uuid",
  "status": "stub_pending",
  "confirm_url": "/api/v1/subscription/confirm"
}
```

**Note**: In a real integration, this would redirect to the payment provider.

---

### `POST /api/v1/subscription/confirm`

Confirm the stub checkout. Assigns the plan and sets subscription to `active`.

**Request**
```json
{ "checkout_session_id": "stub-session-uuid" }
```

**Response `200`**
```json
{
  "subscription": {
    "id": "uuid",
    "plan": "basic",
    "status": "active",
    "stub_payment_ref": "stub-tx-1751500000"
  }
}
```

**Errors**: `400 INVALID_SESSION`, `409 SUBSCRIPTION_ALREADY_ACTIVE`

---

### `GET /api/v1/subscription/current`

Return the current user's subscription status.

**Response `200`**
```json
{
  "subscription": { "id": "uuid", "plan": "basic", "status": "active" }
}
```

**Errors**: `401 UNAUTHENTICATED`, `404 NO_SUBSCRIPTION`

---

## Trips

All trip endpoints require authentication. Write operations (POST/PUT/DELETE) require `role: admin`.

### `GET /api/v1/trips`

List trips visible to the authenticated user (owned or invited-to).

**Response `200`**
```json
{
  "trips": [
    {
      "id": "uuid",
      "title": "Japan 2027",
      "status": "draft",
      "role": "admin",
      "destinations": ["Tokyo", "Osaka"],
      "created_at": "2026-07-02T15:04:05Z"
    }
  ]
}
```

---

### `POST /api/v1/trips`

Create a new trip.

**Request**
```json
{
  "title": "Japan 2027",
  "description": "optional context",
  "travel_styles": ["gastronomy", "film-audiovisual"]
}
```

**Response `201`**
```json
{ "trip": { "id": "uuid", "title": "Japan 2027", "status": "draft" } }
```

**Errors**: `401 UNAUTHENTICATED`, `403 FORBIDDEN` (partner role), `402 SUBSCRIPTION_REQUIRED`

---

### `GET /api/v1/trips/:id`

Get full trip details including days, activities, and collaborators.

**Response `200`**
```json
{
  "trip": {
    "id": "uuid",
    "title": "Japan 2027",
    "status": "draft",
    "travel_styles": ["gastronomy"],
    "destinations": [{ "id": "uuid", "name": "Tokyo", "sequence_order": 1, "duration_days": 5 }],
    "days": [
      {
        "id": "uuid",
        "day_number": 1,
        "label": "Arrival Day",
        "activities": [
          {
            "id": "uuid",
            "title": "Land at Narita, transfer to Shinjuku",
            "type": "logistics",
            "sequence_order": 1,
            "is_ai_generated": true
          }
        ]
      }
    ],
    "collaborators": [{ "user_id": "uuid", "email": "partner@example.com", "accepted_at": "..." }]
  }
}
```

**Errors**: `401 UNAUTHENTICATED`, `403 FORBIDDEN`, `404 TRIP_NOT_FOUND`

---

### `PUT /api/v1/trips/:id`

Update trip metadata (admin only).

**Request**
```json
{ "title": "Japan 2027 — Updated", "status": "published" }
```

**Response `200`**
```json
{ "trip": { "id": "uuid", "title": "Japan 2027 — Updated", "status": "published" } }
```

**Errors**: `401`, `403 FORBIDDEN`, `404 TRIP_NOT_FOUND`

---

### `DELETE /api/v1/trips/:id`

Delete a trip and all related data (admin only, cascades to days/activities/suggestions).

**Response `204`** (no body)

**Errors**: `401`, `403 FORBIDDEN`, `404 TRIP_NOT_FOUND`

---

## Itinerary Generation (Conversational)

### `POST /api/v1/trips/:id/conversation`

Send a user turn in the multi-turn generation conversation. The AI may respond with a clarifying
question or generate the final itinerary.

**Request**
```json
{ "message": "I have 10 days in Japan, arriving in Tokyo and leaving from Osaka. I'm an anime fan, my partner loves metal music." }
```

**Response `200`**
```json
{
  "session_id": "uuid",
  "role": "assistant",
  "message": "Do you want to plan joint activities or sometimes go your separate ways?",
  "itinerary_ready": false
}
```

When the AI considers the conversation complete:
```json
{
  "session_id": "uuid",
  "role": "assistant",
  "message": "Your itinerary is ready!",
  "itinerary_ready": true
}
```

**Errors**: `401`, `403`, `404 TRIP_NOT_FOUND`, `503 AI_UNAVAILABLE`

---

### `GET /api/v1/trips/:id/conversation`

Retrieve the full conversation history for a trip.

**Response `200`**
```json
{
  "session_id": "uuid",
  "status": "active",
  "messages": [
    { "role": "user", "content": "...", "sequence_order": 1, "created_at": "..." },
    { "role": "assistant", "content": "...", "sequence_order": 2, "created_at": "..." }
  ]
}
```

---

## Days and Activities

### `PUT /api/v1/trips/:id/days/:day_id`

Update a day's label (admin only).

**Request**
```json
{ "label": "Tokyo — anime district day" }
```

**Response `200`**
```json
{ "day": { "id": "uuid", "day_number": 3, "label": "Tokyo — anime district day" } }
```

---

### `POST /api/v1/trips/:id/days/:day_id/activities`

Add an activity to a day (admin only).

**Request**
```json
{
  "title": "Visit Akihabara Electric Town",
  "type": "visit",
  "sequence_order": 3,
  "description": "Electronics and anime merchandise district"
}
```

**Response `201`**
```json
{ "activity": { "id": "uuid", "title": "Visit Akihabara Electric Town", "type": "visit", "is_ai_generated": false } }
```

---

### `PUT /api/v1/trips/:id/days/:day_id/activities/:activity_id`

Update an activity (admin only).

**Response `200`** — updated activity object

---

### `DELETE /api/v1/trips/:id/days/:day_id/activities/:activity_id`

Delete an activity (admin only).

**Response `204`** (no body)

---

## Collaborators

### `POST /api/v1/trips/:id/collaborators`

Invite a partner to collaborate on a trip (admin only). Enforces basic plan limit (max 1 partner).

**Request**
```json
{ "email": "partner@example.com" }
```

**Response `201`**
```json
{
  "collaborator": {
    "trip_id": "uuid",
    "user_id": "uuid",
    "email": "partner@example.com",
    "invited_at": "2026-07-02T15:04:05Z",
    "accepted_at": null
  }
}
```

**Errors**: `401`, `403 FORBIDDEN`, `404 TRIP_NOT_FOUND`, `404 USER_NOT_FOUND`,
`409 COLLABORATOR_ALREADY_INVITED`, `409 PLAN_LIMIT_EXCEEDED`

---

### `DELETE /api/v1/trips/:id/collaborators/:user_id`

Remove a partner collaborator (admin only).

**Response `204`** (no body)

---

## Suggestions

### `GET /api/v1/trips/:id/suggestions`

List all suggestions for a trip (visible to all users on the trip). Supports optional
`?status=pending|approved|rejected` filter.

**Response `200`**
```json
{
  "suggestions": [
    {
      "id": "uuid",
      "author": { "id": "uuid", "email": "partner@example.com" },
      "target_type": "activity",
      "target_id": "uuid",
      "content": "Replace dinner with a ramen tasting at Fuunji in Shinjuku",
      "status": "pending",
      "created_at": "2026-07-02T15:04:05Z",
      "resolved_at": null
    }
  ]
}
```

---

### `POST /api/v1/trips/:id/suggestions`

Submit a suggestion (partner only; user must be an accepted collaborator on the trip).

**Request**
```json
{
  "target_type": "activity",
  "target_id": "uuid",
  "content": "Replace dinner with a ramen tasting at Fuunji in Shinjuku"
}
```

**Response `201`**
```json
{ "suggestion": { "id": "uuid", "status": "pending", "created_at": "..." } }
```

**Errors**: `401`, `403 FORBIDDEN` (admin role or non-collaborator), `404 TRIP_NOT_FOUND`

---

### `PATCH /api/v1/trips/:id/suggestions/:suggestion_id/approve`

Approve a pending suggestion and apply its content to the itinerary (admin only).

**Response `200`**
```json
{ "suggestion": { "id": "uuid", "status": "approved", "resolved_at": "..." } }
```

**Errors**: `401`, `403 FORBIDDEN`, `404 SUGGESTION_NOT_FOUND`, `409 ALREADY_RESOLVED`

---

### `PATCH /api/v1/trips/:id/suggestions/:suggestion_id/reject`

Reject a pending suggestion without modifying the itinerary (admin only). History preserved.

**Response `200`**
```json
{ "suggestion": { "id": "uuid", "status": "rejected", "resolved_at": "..." } }
```

**Errors**: `401`, `403 FORBIDDEN`, `404 SUGGESTION_NOT_FOUND`, `409 ALREADY_RESOLVED`
