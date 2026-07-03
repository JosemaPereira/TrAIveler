# Quickstart Validation Guide: TrAIveler MVP

**Date**: 2026-07-02 | **Plan**: [plan.md](plan.md) | **Contracts**: [contracts/api.md](contracts/api.md)

This guide documents the runnable validation scenarios that prove each MVP capability works
end-to-end. It is not an implementation guide — see [data-model.md](data-model.md) for schema
details and [contracts/api.md](contracts/api.md) for full endpoint specifications.

---

## Prerequisites

| Requirement | Notes |
|-------------|-------|
| Go 1.24+ | `go version` |
| Node.js 20+ | `node --version` |
| Docker + Docker Compose | For PostgreSQL local instance |
| Anthropic API key | Set in `.env` as `ANTHROPIC_API_KEY` |
| `.env` file populated | Copy `.env.example` and fill values |

### Environment Setup

```bash
# 1. Start the database
docker compose up -d postgres

# 2. Run migrations
cd backend && go run ./cmd/migrate up

# 3. Start backend
go run ./cmd/server

# 4. Start frontend (separate terminal)
cd frontend && npm install && npm run dev
```

Backend default: `http://localhost:8080`  
Frontend default: `http://localhost:5173`

---

## Scenario 1 — Register, Mock Checkout, and Receive Basic Plan

**Validates**: FR-008, FR-013, SC — subscription onboarding

```
1. Open http://localhost:5173/register
2. Enter email + password → Submit
3. System redirects to the stub checkout screen
4. Click "Subscribe — Basic Plan" → stub always succeeds
5. System redirects to dashboard
```

**Expected outcomes**:
- User is authenticated (cookie set)
- `GET /api/v1/auth/me` returns `{ role: "admin", subscription: { status: "active", plan: "basic" } }`
- `GET /api/v1/subscription/current` returns `{ status: "active" }`
- Console or server log shows `[STUB] Charging user <id>: ...`

---

## Scenario 2 — Generate an Itinerary via Conversational Flow

**Validates**: FR-001 (User Story 1, P1), FR-012, SC-001, SC-005

```
1. From dashboard, click "New Trip"
2. In the conversation panel, type:
   "I have 10 days in Japan. I arrive in Tokyo on day 1 and depart from Osaka on day 10.
    I'm an anime fan; my partner loves metal music. Propose a trip for both of us."
3. AI responds with a clarifying question, e.g.:
   "Do you want to explore together the whole time or go your separate ways sometimes?"
4. Reply: "We'd like mostly joint activities but with some independent time each day."
5. AI generates and streams the itinerary.
```

**Expected outcomes**:
- At least one follow-up AI question before itinerary generation (FR-012)
- Day-by-day plan is visible: 10 days, Day 1 = Tokyo arrival logistics (type: `logistics`)
- At least one activity tagged to anime (e.g., Akihabara) and one to music (e.g., live venue)
- At least one `type: food` activity per day
- Day 9–10 includes `type: transfer` from Tokyo to Osaka region
- `GET /api/v1/trips/:id/conversation` returns full message history
- `GET /api/v1/trips/:id` returns populated days and activities

---

## Scenario 3 — Enrich an Existing Plan (Free-Form Input)

**Validates**: FR-002 (User Story 3, P2), SC-004

```
1. Create a new trip
2. In the conversation panel, type:
   "I want to visit: Sagrada Família, Park Güell, Gothic Quarter, Barceloneta beach,
    and Camp Nou. I have 4 days in Barcelona."
3. AI generates an itinerary anchored to these 5 places.
```

**Expected outcomes**:
- All 5 places appear as activities in the generated itinerary
- None of the 5 is omitted or replaced (SC-004 — 0% omission rate)
- Additional complementary activities fill the remaining time slots
- Food suggestions appear on each day

---

## Scenario 4 — Off-the-Beaten-Path Recommendations

**Validates**: FR-006 (User Story 2, P2), SC-002

```
1. Create a new trip for a well-known destination (e.g., "Paris, 5 days")
2. Select travel style: "gastronomy"
3. Generate itinerary
4. Review Day 1–5
```

**Expected outcomes**:
- At least one activity per day is NOT from the standard top-10 tourist list
  (e.g., not just Eiffel Tower, Louvre, Notre-Dame)
- Food activities feature local neighborhood bistros, not only tourist restaurants
- `travel_styles` field on the trip contains `["gastronomy"]`

---

## Scenario 5 — Invite Partner and Suggest/Approve Workflow

**Validates**: FR-009, FR-011, SC-003, User Story 4

```
Setup: Two registered accounts — Admin (user A) and Partner (user B, role: partner)

1. User A: Open trip → Settings → Invite Collaborator → enter user B's email → Submit
2. User B: Log in → accept invitation
3. User B: Navigate to the trip → open suggestions panel
4. User B: Submit suggestion:
   Target: Activity "Dinner at tourist restaurant"
   Content: "Replace with ramen tasting at Fuunji in Shinjuku — locals' favourite"
5. User A: Open suggestions panel → suggestion shows "pending" status → click Approve
6. Verify itinerary is updated with the suggestion content
7. User A: Submit a second suggestion from User B → click Reject
8. Verify itinerary is unchanged; rejected suggestion appears in history
```

**Expected outcomes**:
- Step 1 enforces plan limit: attempting to invite a second partner returns `409 PLAN_LIMIT_EXCEEDED`
- Step 4: `GET /api/v1/trips/:id/suggestions` shows suggestion with `status: "pending"` to both users
- Step 5: suggestion transitions to `status: "approved"`; itinerary activity is updated
- Step 7: suggestion transitions to `status: "rejected"`; itinerary is unchanged
- Both suggestions remain in history (FR-011 — never deleted)
- If User B attempts direct `PUT /api/v1/trips/:id/days/:day_id/activities/:activity_id` → `403 FORBIDDEN`

---

## Scenario 6 — Role Permission Boundaries

**Validates**: FR-009

```
Using a partner-role account that is an accepted collaborator on a trip:

1. Attempt POST /api/v1/trips (create trip)       → expect 403 FORBIDDEN
2. Attempt DELETE /api/v1/trips/:id               → expect 403 FORBIDDEN
3. Attempt PATCH .../suggestions/:id/approve      → expect 403 FORBIDDEN
4. Attempt PUT .../activities/:id (direct edit)   → expect 403 FORBIDDEN
5. POST .../suggestions (valid suggestion)         → expect 201 Created
6. GET /api/v1/trips/:id (view trip)              → expect 200 OK
```

---

## Scenario 7 — Scale Baseline Check

**Validates**: SC-007 (< 50 concurrent users, single instance)

Using a load testing tool (e.g., `k6` or `hey`):

```bash
# 50 concurrent GET /api/v1/trips requests over 30 seconds
hey -c 50 -z 30s -H "Cookie: auth_token=<valid_token>" http://localhost:8080/api/v1/trips
```

**Expected outcomes**:
- All responses return `200 OK`
- No 5xx errors
- p95 response time < 200 ms (non-AI endpoints)

---

## Test Execution Reference

```bash
# Backend unit + integration tests
cd backend && go test ./... -race -count=1

# Frontend unit + integration tests
cd frontend && npm test

# E2E (requires running stack)
cd e2e && npx playwright test

# Lint gates
cd backend && golangci-lint run ./...
cd frontend && npm run lint
```

All gates must pass before a pull request can be merged (Constitution III).
