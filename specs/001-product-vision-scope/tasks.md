# Tasks: Product Vision and Scope — TrAIveler MVP

**Input**: Design documents from `/specs/001-product-vision-scope/`

**Prerequisites**: [plan.md](plan.md) · [spec.md](spec.md) · [research.md](research.md) · [data-model.md](data-model.md) · [contracts/api.md](contracts/api.md)

**Tests**: Not explicitly requested in spec — no test tasks generated.
Test code is expected to be written alongside implementation per TDD (Constitution I).

**Format**: `- [ ] [ID] [P?] [Story?] Description — file path`
- `[P]` = parallelizable (no incomplete dependencies, different files)
- `[US#]` = user story label (Phase 3+ only)

---

## Phase 1: Setup

**Purpose**: Monorepo initialization, toolchain configuration, and dev environment scaffold.
No user story work can begin before this phase is complete.

- [ ] T001 Create monorepo root directory structure: `backend/`, `frontend/`, `e2e/`, `.github/workflows/`, `docs/` — repository root
- [ ] T002 Initialize Go 1.24 module with all backend dependencies (`chi`, `pgx/v5`, `jwt/v5`, `anthropic-sdk-go`, `goose/v3`, `testify`) — `backend/go.mod`
- [ ] T003 [P] Initialize React 19 + TypeScript Vite project with all frontend dependencies (`react-router`, `@tanstack/react-query`, `zustand`, `lucide-react`, `vitest`, `@testing-library/react`, `msw`, `playwright`) — `frontend/package.json`
- [ ] T004 [P] Configure `golangci-lint` with required linters (`errcheck`, `govet`, `staticcheck`, `revive`, `gosec`) — `backend/.golangci.yml`
- [ ] T005 [P] Configure Prettier and ESLint strict mode with TypeScript plugin — `frontend/.prettierrc`, `frontend/.eslintrc.json`, `frontend/tsconfig.json`
- [ ] T006 Create `docker-compose.yml` with PostgreSQL 16 service, named volume, and health check — `docker-compose.yml`
- [ ] T007 [P] Create `.env.example` listing all required environment variables (`DATABASE_URL`, `ANTHROPIC_API_KEY`, `JWT_SECRET`, `JWT_TTL_SECONDS`, `PORT`, `FRONTEND_ORIGIN`) — `.env.example`
- [ ] T008 Create `Makefile` with targets: `lint`, `test`, `migrate-up`, `migrate-down`, `build`, `dev` — `Makefile`

---

## Phase 2: Foundational

**Purpose**: Authentication, subscription (mock checkout), database schema, shared middleware, and
frontend shell. **All user story phases block on this phase completing.**

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Backend — Data Layer

- [ ] T009 Create all 8 database migration SQL files in sequential order (users, plans, subscriptions, trips, destinations, days+activities, collaborators, suggestions+conversation) — `backend/migrations/001_*.sql` … `backend/migrations/008_*.sql`
- [ ] T010 Implement config struct with env var loading and validation (fail-fast on missing required vars) — `backend/config/config.go`
- [ ] T011 [P] Implement pgxpool connection initialization with context-aware open/close — `backend/internal/database/db.go`
- [ ] T012 [P] Define `PaymentProvider` interface with `CreateSubscription`, `CancelSubscription`, `GetSubscription` methods and all request/response types — `backend/internal/subscription/payment/provider.go`
- [ ] T013 Implement `StubProvider` satisfying `PaymentProvider`: always returns `status: succeeded`, logs `[STUB]` to stdout, generates deterministic transaction ID — `backend/internal/subscription/payment/stub.go`
- [ ] T014 [P] Implement User repository: `Create`, `FindByEmail`, `FindByID`, `UpdateSubscription` — `backend/internal/auth/repository.go`
- [ ] T015 [P] Implement Plan and Subscription repositories: `FindPlanByName`, `CreateSubscription`, `FindSubscriptionByUser`, `UpdateSubscriptionStatus` — `backend/internal/subscription/repository.go`

### Backend — Service & Handler Layer

- [ ] T016 Implement auth service: `Register` (bcrypt min-cost 12), `Login` (compare hash, issue JWT), `Me` — `backend/internal/auth/service.go`
- [ ] T017 Implement subscription service: `ListPlans`, `InitiateCheckout` (via stub), `ConfirmCheckout` (assigns plan, enforces no duplicate subscription), `GetCurrent` — `backend/internal/subscription/service.go`
- [ ] T018 [P] Implement JWT auth middleware: validate HTTP-only cookie `auth_token`, attach `User` to request context, return `401` on failure — `backend/pkg/middleware/auth.go`
- [ ] T019 [P] Implement role-guard middleware factory: `RequireRole("admin")` / `RequireRole("partner")`, return `403` on mismatch — `backend/pkg/middleware/role.go`
- [ ] T020 [P] Implement JSON response helpers: `Success(w, status, payload)`, `Error(w, status, code, message)` — `backend/pkg/response/response.go`
- [ ] T021 Implement auth HTTP handlers: `POST /register`, `POST /login` (set cookie), `POST /logout` (clear cookie), `GET /me` — `backend/internal/auth/handler.go`
- [ ] T022 Implement subscription HTTP handlers: `GET /plans`, `POST /checkout`, `POST /confirm`, `GET /current` — `backend/internal/subscription/handler.go`
- [ ] T023 Scaffold Chi router: mount auth and subscription route groups, apply CORS, request-ID, and logging middleware, wire `config` and DB pool — `backend/cmd/server/main.go`

### Frontend — Shell & Auth

- [ ] T024 [P] Create CSS custom property design tokens (colors, spacing, typography, border-radius) following `docs/ui-guidelines.md` — `frontend/src/styles/tokens.css`
- [ ] T025 [P] Create typed API client base: `apiFetch` wrapper with `credentials: include`, JSON parsing, typed error shape — `frontend/src/services/api.ts`
- [ ] T026 [P] Create auth Zustand store: `user`, `setUser`, `clearUser` actions; persist to `sessionStorage` — `frontend/src/features/auth/store.ts`
- [ ] T027 [P] Create `ProtectedRoute` (redirect to `/login` if no session) and `GuestRoute` (redirect to `/dashboard` if authenticated) components — `frontend/src/components/ProtectedRoute.tsx`
- [ ] T028 [P] Create primitive `Button`, `Input`, `Label`, `Badge` components using design tokens — `frontend/src/components/primitives/`
- [ ] T029 Implement Register page: email + password form with validation, calls `POST /auth/register`, redirects to stub checkout — `frontend/src/pages/RegisterPage.tsx`
- [ ] T030 Implement stub Checkout page: plan summary display, "Subscribe" CTA that calls `POST /subscription/checkout` then `POST /subscription/confirm`, redirects to dashboard — `frontend/src/pages/SubscribePage.tsx`
- [ ] T031 Implement Login page: email + password form, calls `POST /auth/login`, sets user in store, redirects to dashboard — `frontend/src/pages/LoginPage.tsx`
- [ ] T032 Wire React Router v7 with all routes (`/`, `/login`, `/register`, `/subscribe`, `/dashboard`, `/trips/:id`, `/generate`) and layout wrappers — `frontend/src/App.tsx`

**Checkpoint**: Auth + subscription complete — all user stories can now begin in parallel.

---

## Phase 3: User Story 1 — First-Time Traveler Plans a Trip from Zero (Priority: P1) 🎯 MVP

**Goal**: Authenticated admin user submits a free-form travel request, the AI conducts a
multi-turn conversation to clarify preferences, then generates a complete day-by-day itinerary
with visits, food, logistics, and inter-city transfers.

**Independent Test**: Quickstart Scenario 1 + Scenario 2 — register → subscribe → create trip →
converse with AI → verify day-by-day itinerary with at least one activity of each type.

### Backend

- [ ] T033 [P] [US1] Implement Trip + Destination + Day + Activity repository: `CreateTrip`, `ListTripsByUser`, `FindTripByID`, `DeleteTrip`, `UpsertDay`, `UpsertActivity`, `DeleteActivity` — `backend/internal/trip/repository.go`
- [ ] T034 [P] [US1] Implement ConversationSession + ConversationMessage repository: `CreateSession`, `AppendMessage`, `GetSessionByTrip`, `ListMessages` — `backend/internal/conversation/repository.go`
- [ ] T035 [US1] Implement Trip service: `Create`, `List`, `Get`, `Update`, `Delete`; enforce creator must be `admin` role — `backend/internal/trip/service.go`
- [ ] T036 [US1] Implement Conversation service: `SendMessage` (appends user turn, calls itinerary service, appends AI turn), `GetHistory`; detect when itinerary is ready and set `itinerary_ready: true` — `backend/internal/conversation/service.go`
- [ ] T037 [US1] Implement Itinerary service: build Claude system prompt with trip context, send multi-turn message history via streaming, parse structured tool-use response into `[]Day` with `[]Activity`, persist to DB — `backend/internal/itinerary/service.go`
- [ ] T038 [US1] Implement Trip HTTP handlers: `GET /trips`, `POST /trips`, `GET /trips/:id`, `PUT /trips/:id`, `DELETE /trips/:id`; apply `RequireRole("admin")` on write ops — `backend/internal/trip/handler.go`
- [ ] T039 [US1] Implement Conversation HTTP handlers: `POST /trips/:id/conversation`, `GET /trips/:id/conversation`; stream AI response chunks via `text/event-stream` — `backend/internal/conversation/handler.go`
- [ ] T040 [US1] Register trip and conversation routes in Chi router — `backend/cmd/server/main.go`

### Frontend

- [ ] T041 [P] [US1] Create `TripCard` composite: destination names, duration, status badge, created date; handles loading skeleton and empty state — `frontend/src/components/composites/TripCard.tsx`
- [ ] T042 [P] [US1] Create `ActivityItem` composite: type icon (Lucide), title, description, AI-generated indicator; handles loading and empty states — `frontend/src/components/composites/ActivityItem.tsx`
- [ ] T043 [P] [US1] Create `DaySection` composite: day number, label, ordered activity list using `ActivityItem`; handles empty day state — `frontend/src/components/composites/DaySection.tsx`
- [ ] T044 [P] [US1] Create `ConversationPanel` feature: message thread (user + AI bubbles), text input, send button, SSE stream rendering, loading indicator while AI responds — `frontend/src/components/features/ConversationPanel.tsx`
- [ ] T045 [P] [US1] Create `ItineraryView` feature: scrollable list of `DaySection` components; handles loading skeleton, error banner, and empty "not yet generated" state — `frontend/src/components/features/ItineraryView.tsx`
- [ ] T046 [US1] Implement trips and conversation API service functions with TanStack Query hooks (`useTrips`, `useTrip`, `useConversation`, `useSendMessage`) — `frontend/src/services/trips.ts`, `frontend/src/services/conversation.ts`
- [ ] T047 [US1] Implement Generate page: new trip form (title, destination) → `ConversationPanel` for multi-turn flow → transitions to `ItineraryView` when `itinerary_ready: true` — `frontend/src/pages/GeneratePage.tsx`
- [ ] T048 [US1] Implement Trip detail page: `ItineraryView` for existing trip, action bar (edit title, delete trip) — `frontend/src/pages/TripPage.tsx`
- [ ] T049 [US1] Implement Dashboard page: `TripCard` grid via `useTrips`, "New Trip" CTA, empty state illustration — `frontend/src/pages/DashboardPage.tsx`
- [ ] T050 [US1] Add Playwright E2E spec covering: register → subscribe → generate itinerary via conversation → verify Day 1 visible — `e2e/generate-itinerary.spec.ts`

**Checkpoint**: User Story 1 complete — demo-able MVP. AI generates a full itinerary from a
conversation. All subsequent stories extend this foundation.

---

## Phase 4: User Story 2 — Experienced Traveler Seeks an Off-the-Beaten-Path Perspective (Priority: P2)

**Goal**: User selects one or more travel styles; the AI tailors content to those styles and
surfaces non-obvious recommendations alongside mainstream attractions.

**Independent Test**: Quickstart Scenario 3 + Scenario 4 — create trip with `gastronomy` style →
verify food activities are prominent; create same trip without style → compare outputs.

- [ ] T051 [P] [US2] Create `TravelStyleSelector` composite: multi-select chip group for all 5 seed styles, controlled component with accessible keyboard support — `frontend/src/components/composites/TravelStyleSelector.tsx`
- [ ] T052 [US2] Integrate `TravelStyleSelector` into Generate page new-trip form; pass selected styles on trip creation — `frontend/src/pages/GeneratePage.tsx`
- [ ] T053 [US2] Update Trip service to persist `TripTravelStyle` join records on create/update — `backend/internal/trip/service.go`
- [ ] T054 [US2] Update Itinerary service: inject selected travel styles into Claude system prompt; instruct Claude to include at least one non-mainstream recommendation per day per active style — `backend/internal/itinerary/service.go`
- [ ] T055 [US2] Update Trip HTTP handlers to accept `travel_styles[]` in `POST /trips` and `PUT /trips/:id` and return them in `GET /trips/:id` — `backend/internal/trip/handler.go`
- [ ] T056 [P] [US2] Update `ItineraryView` to display travel style badges at the trip header level — `frontend/src/components/features/ItineraryView.tsx`
- [ ] T057 [US2] Add Playwright E2E spec: create trip with `gastronomy` style → verify at least one food activity per day — `e2e/travel-style-personalization.spec.ts`

**Checkpoint**: User Story 2 complete — personalized, style-driven itineraries demonstrable.

---

## Phase 5: User Story 3 — Planner Provides an Existing Plan for Enrichment (Priority: P2)

**Goal**: User provides a free-form list of places they want to visit; the AI preserves all of
them as itinerary anchors and fills the remaining time with complementary suggestions.

**Independent Test**: Quickstart Scenario 3 — provide 5 named places → verify all 5 appear in
output with 0% omission rate.

- [ ] T058 [US3] Update Itinerary service: detect "anchor mode" when conversation input contains an enumerated list of places; extract place names; inject them as mandatory anchors in Claude system prompt with explicit "do not omit any anchor place" instruction — `backend/internal/itinerary/service.go`
- [ ] T059 [US3] Update Conversation service: when anchor places are detected, return extracted list in AI response message for user confirmation before final generation — `backend/internal/conversation/service.go`
- [ ] T060 [P] [US3] Update `ConversationPanel` to render an "anchored places" confirmation card when API response includes `anchor_places[]` — `frontend/src/components/features/ConversationPanel.tsx`
- [ ] T061 [US3] Update Generate page to handle anchor confirmation step: show extracted places, allow user to proceed or amend before triggering generation — `frontend/src/pages/GeneratePage.tsx`
- [ ] T062 [US3] Add Playwright E2E spec: provide list of 5 places → confirm anchors → generate → assert all 5 appear in itinerary — `e2e/enrich-existing-plan.spec.ts`

**Checkpoint**: User Story 3 complete — Detail Planner persona fully served.

---

## Phase 6: User Story 4 — Group Collaborates on a Shared Itinerary (Priority: P3)

**Goal**: Admin invites one partner collaborator; partner views the trip and submits suggestions;
admin approves or rejects each suggestion; history is permanently preserved.

**Independent Test**: Quickstart Scenario 5 + Scenario 6 — two accounts; admin invites partner →
partner submits suggestion → admin approves → itinerary updated; admin rejects second suggestion →
itinerary unchanged; both suggestions visible in history.

### Backend

- [ ] T063 [P] [US4] Extend Trip repository with collaborator operations: `CreateCollaborator`, `AcceptCollaborator`, `RemoveCollaborator`, `ListCollaboratorsByTrip`, `CountActiveCollaborators` — `backend/internal/trip/repository.go`
- [ ] T064 [P] [US4] Implement Suggestion repository: `CreateSuggestion`, `ListByTrip`, `ListByStatus`, `UpdateStatus` (to approved/rejected), `FindByID` — `backend/internal/suggestion/repository.go`
- [ ] T065 [US4] Extend Trip service with collaborator logic: `InviteCollaborator` (enforce basic plan limit of 1 via `CountActiveCollaborators`; return `409` if exceeded), `AcceptInvitation`, `RemoveCollaborator` — `backend/internal/trip/service.go`
- [ ] T066 [US4] Implement Suggestion service: `Submit` (only accepted partner collaborators), `Approve` (admin only; applies content as Activity update), `Reject` (admin only; history preserved, no itinerary change) — `backend/internal/suggestion/service.go`
- [ ] T067 [US4] Implement Collaborator HTTP handlers: `POST /trips/:id/collaborators`, `DELETE /trips/:id/collaborators/:user_id`; apply `RequireRole("admin")` — `backend/internal/trip/handler.go`
- [ ] T068 [US4] Implement Suggestion HTTP handlers: `GET /trips/:id/suggestions`, `POST /trips/:id/suggestions`, `PATCH /trips/:id/suggestions/:id/approve`, `PATCH /trips/:id/suggestions/:id/reject` — `backend/internal/suggestion/handler.go`
- [ ] T069 [US4] Register collaborator and suggestion routes in Chi router — `backend/cmd/server/main.go`

### Frontend

- [ ] T070 [P] [US4] Create `SuggestionBubble` composite: author email, target label, content text, status badge (`pending`/`approved`/`rejected`), approve/reject action buttons (admin only) — `frontend/src/components/composites/SuggestionBubble.tsx`
- [ ] T071 [P] [US4] Create `SuggestionQueue` feature: list of `SuggestionBubble` components filtered by status; empty state for no suggestions; loading skeleton — `frontend/src/components/features/SuggestionQueue.tsx`
- [ ] T072 [P] [US4] Create `CollaboratorInvite` composite: email input, invite button, current partner badge with remove option; plan limit warning when limit reached — `frontend/src/components/composites/CollaboratorInvite.tsx`
- [ ] T073 [US4] Implement suggestion and collaborator API service functions with TanStack Query hooks (`useSuggestions`, `useSubmitSuggestion`, `useApproveSuggestion`, `useRejectSuggestion`, `useInviteCollaborator`) — `frontend/src/services/suggestions.ts`, `frontend/src/services/collaborators.ts`
- [ ] T074 [US4] Integrate `SuggestionQueue` and `CollaboratorInvite` into Trip detail page; show `CollaboratorInvite` to admin only; show `SuggestionQueue` to all trip users — `frontend/src/pages/TripPage.tsx`
- [ ] T075 [US4] Add Playwright E2E spec: admin invites partner → partner submits suggestion → admin approves → assert itinerary updated; admin rejects second suggestion → assert history preserved — `e2e/suggestion-approval-flow.spec.ts`

**Checkpoint**: User Story 4 complete — full collaboration flow demonstrable end-to-end.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: CI pipelines, containerization, documentation, and final production-readiness gates.

- [ ] T076 [P] Add backend lint CI workflow: runs `golangci-lint run ./...` on push and PR — `.github/workflows/backend-lint.yml`
- [ ] T077 [P] Add frontend lint + type-check CI workflow: runs `npm run lint` and `tsc --noEmit` on push and PR — `.github/workflows/frontend-lint.yml`
- [ ] T078 [P] Add backend test CI workflow: runs `go test ./... -race -count=1` against a PostgreSQL service container — `.github/workflows/backend-test.yml`
- [ ] T079 [P] Add frontend test CI workflow: runs `npm test -- --run` (Vitest) on push and PR — `.github/workflows/frontend-test.yml`
- [ ] T080 Create production-ready `Dockerfile` for the backend (multi-stage build, non-root user, runs migrations on start) — `backend/Dockerfile`
- [ ] T081 Write `README.md` with: project overview, prerequisites, local setup commands, architecture diagram reference, and development workflow — `README.md`

---

## Dependencies

```
Phase 1 (T001–T008)
    └── Phase 2 (T009–T032)
            ├── Phase 3: US1 (T033–T050)  🎯 MVP — can start immediately after Phase 2
            │       ├── Phase 4: US2 (T051–T057)  extends itinerary service from US1
            │       ├── Phase 5: US3 (T058–T062)  extends conversation+itinerary service from US1
            │       └── Phase 6: US4 (T063–T075)  requires a trip to collaborate on (US1)
            └── Phase 7 (T076–T081)         can begin after Phase 1; polish tasks are independent
```

**US2, US3, US4 can be implemented in parallel** once Phase 3 (US1) is complete, as they extend
different services (itinerary, conversation, suggestion) and different frontend components.

---

## Parallel Execution Examples

### During Phase 2 (Foundational)
| Stream A | Stream B | Stream C |
|----------|----------|----------|
| T009 → T010 → T011 | T012 → T013 | T024 → T025 → T026 |
| T014 → T016 → T021 | T015 → T017 → T022 | T027 → T028 → T029 |
| T018 → T019 → T020 | T023 | T030 → T031 → T032 |

### During Phase 3 (US1)
| Stream A — Backend | Stream B — Components | Stream C — Pages |
|--------------------|-----------------------|-----------------|
| T033 → T035 → T038 | T041 | T046 → T047 |
| T034 → T036 → T039 | T042 → T043 | T048 |
| T037 → T040 | T044 → T045 | T049 → T050 |

### After US1 Complete — Phase 4/5/6 in parallel
| Stream A — US2 | Stream B — US3 | Stream C — US4 |
|----------------|----------------|----------------|
| T051 → T052 | T058 → T059 | T063 → T065 → T067 |
| T053 → T054 → T055 | T060 → T061 | T064 → T066 → T068 |
| T056 → T057 | T062 | T069 → T070 → T073 → T074 → T075 |

---

## Implementation Strategy

**MVP Scope (minimum demonstrable product)**: Complete **Phase 1 + Phase 2 + Phase 3 (US1)** only.
This delivers: register, mock checkout, generate an AI-powered itinerary via multi-turn
conversation — the core product value proposition. All 4 quickstart validation scenarios
for US1 must pass.

**Increment 2**: Add Phase 4 (US2) — travel style personalization. Minimal surface area change;
extends itinerary service only.

**Increment 3**: Add Phase 5 (US3) — existing plan enrichment. Minimal change; extends conversation
and itinerary services.

**Increment 4**: Add Phase 6 (US4) — collaboration. New domain (`suggestion`), new backend handlers,
new frontend components. Largest increment; can be deferred safely after MVP demo.

**Increment 5**: Phase 7 polish — CI, Docker, README. Can be started in parallel with any increment
once Phase 1 is done.
