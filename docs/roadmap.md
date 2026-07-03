# Project Roadmap

> Generated and reconciled by /build-roadmap. Source of truth for task existence,
> titles, and dependencies is `specs/*/tasks.md`. Priority, Status, Phase, Issue, and
> Notes are human-owned and preserved across runs. Do not hand-edit the stable IDs.

**Last reconciled**: 2026-07-03 (updated with spec 004)

## Legend

- **Group**: shared value (e.g. `G-SETUP-1`) = tasks handled by ONE issue (checklist inside); empty = standalone (1 task = 1 issue). Human-owned — set when grouping is desired.
- **Sprint**: sprint number or milestone label. Human-owned — leave blank until sprint planning.
- **Priority**: P1 (critical) | P2 | P3 | TBD
- **Status**: Backlog | Ready | In Progress | In Review | Done
- **Parallel**: yes = task carries `[P]` flag in source (can run concurrently with peers)
- **Issue**: link to the tracker issue once created (empty = not yet created). Grouped tasks share the same URL.

---

## Foundation Phase

<!-- Ordered by cross-spec dependency: vision/scope → NFRs → cloud/IaC → security → architecture → domain -->

### Spec 001 — Product Vision and Scope &nbsp; `specs/001-product-vision-scope/tasks.md`

#### Phase 1 — Setup

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 001-T001 | Create monorepo root directory structure (`backend/`, `frontend/`, `e2e/`, `.github/workflows/`, `docs/`) | | | P1 | Backlog | - | no | | |
| 001-T002 | Initialize Go 1.24 module with backend dependencies (chi, pgx/v5, jwt/v5, anthropic-sdk-go, goose/v3, testify) | G-SETUP-INIT | | P1 | Backlog | 001-T001 | no | | |
| 001-T003 | Initialize React 19 + TypeScript Vite project with frontend dependencies | G-SETUP-INIT | | P1 | Backlog | 001-T001 | yes | | |
| 001-T004 | Configure `golangci-lint` with required linters (errcheck, govet, staticcheck, revive, gosec) | G-SETUP-INIT | | P1 | Backlog | 001-T001 | yes | | |
| 001-T005 | Configure Prettier and ESLint strict mode with TypeScript plugin | G-SETUP-INIT | | P1 | Backlog | 001-T001 | yes | | |
| 001-T006 | Create `docker-compose.yml` with PostgreSQL 16 service, named volume, and health check | G-SETUP-TOOLS | | P1 | Backlog | 001-T001 | no | | |
| 001-T007 | Create `.env.example` with all required environment variables | G-SETUP-TOOLS | | P1 | Backlog | 001-T001 | yes | | |
| 001-T008 | Create `Makefile` with lint, test, migrate-up, migrate-down, build, dev targets | G-SETUP-TOOLS | | P1 | Backlog | 001-T001 | no | | |

#### Phase 2 — Foundational (Backend Data Layer)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 001-T009 | Create 8 database migration SQL files (users, plans, subscriptions, trips, destinations, days+activities, collaborators, suggestions+conversation) | | | P1 | Backlog | 001-T001 | no | | |
| 001-T010 | Implement config struct with env var loading and fail-fast validation | G-BACKEND-CONFIG | | P1 | Backlog | 001-T001 | no | | |
| 001-T011 | Implement pgxpool connection initialization with context-aware open/close | G-BACKEND-CONFIG | | P1 | Backlog | 001-T001 | yes | | |
| 001-T012 | Define `PaymentProvider` interface with CreateSubscription, CancelSubscription, GetSubscription | G-BACKEND-CONFIG | | P1 | Backlog | 001-T001 | yes | | |
| 001-T013 | Implement `StubProvider` satisfying `PaymentProvider` (always succeeds, logs [STUB]) | G-BACKEND-CONFIG | | P1 | Backlog | 001-T012 | no | | |
| 001-T014 | Implement User repository (Create, FindByEmail, FindByID, UpdateSubscription) | G-BACKEND-AUTH-REPOS | | P1 | Backlog | 001-T009 | yes | | |
| 001-T015 | Implement Plan and Subscription repositories (FindPlanByName, CreateSubscription, FindSubscriptionByUser) | G-BACKEND-AUTH-REPOS | | P1 | Backlog | 001-T009 | yes | | |

#### Phase 2 — Foundational (Backend Service & Handler Layer)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 001-T016 | Implement auth service (Register with bcrypt, Login with JWT, Me) | | | P1 | Backlog | 001-T014 | no | | |
| 001-T017 | Implement subscription service (ListPlans, InitiateCheckout, ConfirmCheckout, GetCurrent) | | | P1 | Backlog | 001-T015 | no | | |
| 001-T018 | Implement JWT auth middleware (validate HTTP-only cookie, attach User to context, 401 on failure) | G-BACKEND-MIDDLEWARE | | P1 | Backlog | 001-T001 | yes | | |
| 001-T019 | Implement role-guard middleware factory (`RequireRole("admin")` / `RequireRole("partner")`) | G-BACKEND-MIDDLEWARE | | P1 | Backlog | 001-T001 | yes | | |
| 001-T020 | Implement JSON response helpers (Success, Error) | G-BACKEND-MIDDLEWARE | | P1 | Backlog | 001-T001 | yes | | |
| 001-T021 | Implement auth HTTP handlers (POST /register, /login, /logout, GET /me) | | | P1 | Backlog | 001-T016 | no | | |
| 001-T022 | Implement subscription HTTP handlers (GET /plans, POST /checkout, /confirm, GET /current) | | | P1 | Backlog | 001-T017 | no | | |
| 001-T023 | Scaffold Chi router: mount auth and subscription routes, apply CORS, request-ID, and logging middleware | | | P1 | Backlog | 001-T021, 001-T022 | no | | |

#### Phase 2 — Foundational (Frontend Shell & Auth)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 001-T024 | Create CSS custom property design tokens (colors, spacing, typography, border-radius) | | | P1 | Backlog | 001-T001 | yes | | |
| 001-T025 | Create typed API client base (`apiFetch` wrapper with credentials: include, JSON parsing) | G-FRONTEND-INFRA | | P1 | Backlog | 001-T001 | yes | | |
| 001-T026 | Create auth Zustand store (user, setUser, clearUser; persist to sessionStorage) | G-FRONTEND-INFRA | | P1 | Backlog | 001-T001 | yes | | |
| 001-T027 | Create `ProtectedRoute` (redirect to /login) and `GuestRoute` (redirect to /dashboard) | G-FRONTEND-INFRA | | P1 | Backlog | 001-T001 | yes | | |
| 001-T028 | Create primitive Button, Input, Label, Badge components using design tokens | | | P1 | Backlog | 001-T024 | yes | | |
| 001-T029 | Implement Register page (email + password form, calls POST /auth/register, redirects to checkout) | G-FRONTEND-AUTH-PAGES | | P1 | Backlog | 001-T024, 001-T025 | no | | |
| 001-T030 | Implement stub Checkout page (plan summary, POST /subscription/checkout + /confirm, redirect dashboard) | G-FRONTEND-AUTH-PAGES | | P1 | Backlog | 001-T028, 001-T029 | no | | |
| 001-T031 | Implement Login page (calls POST /auth/login, sets user in store, redirects to dashboard) | G-FRONTEND-AUTH-PAGES | | P1 | Backlog | 001-T028, 001-T030 | no | | |
| 001-T032 | Wire React Router v7 with all routes (/, /login, /register, /subscribe, /dashboard, /trips/:id, /generate) | | | P1 | Backlog | 001-T029, 001-T030, 001-T031 | no | | |

#### Phase 3 — User Story 1: First-Time Traveler Plans a Trip (Priority: P1) 🎯 MVP

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 001-T033 | Implement Trip + Destination + Day + Activity repository (CreateTrip, UpsertDay, UpsertActivity) | G-US1-REPOS | | P1 | Backlog | 001-T009 | yes | | |
| 001-T034 | Implement ConversationSession + ConversationMessage repository (CreateSession, AppendMessage, ListMessages) | G-US1-REPOS | | P1 | Backlog | 001-T009 | yes | | |
| 001-T035 | Implement Trip service (Create, List, Get, Update, Delete; admin-only writes) | | | P1 | Backlog | 001-T033 | no | | |
| 001-T036 | Implement Conversation service (SendMessage, GetHistory; detects itinerary_ready) | | | P1 | Backlog | 001-T034 | no | | |
| 001-T037 | Implement Itinerary service (build Claude prompt, streaming, parse tool-use response, persist to DB) | | | P1 | Backlog | 001-T035 | no | | |
| 001-T038 | Implement Trip HTTP handlers (GET/POST/PUT/DELETE /trips; RequireRole admin on writes) | G-US1-HANDLERS | | P1 | Backlog | 001-T035 | no | | |
| 001-T039 | Implement Conversation HTTP handlers (POST/GET /trips/:id/conversation; SSE streaming) | G-US1-HANDLERS | | P1 | Backlog | 001-T036 | no | | |
| 001-T040 | Register trip and conversation routes in Chi router | | | P1 | Backlog | 001-T038, 001-T039 | no | | |
| 001-T041 | Create `TripCard` composite (destination names, duration, status badge; loading/empty states) | G-US1-COMPONENTS | | P1 | Backlog | 001-T028 | yes | | |
| 001-T042 | Create `ActivityItem` composite (Lucide type icon, title, description, AI-generated indicator) | G-US1-COMPONENTS | | P1 | Backlog | 001-T028 | yes | | |
| 001-T043 | Create `DaySection` composite (day number, label, ordered ActivityItem list; empty state) | G-US1-COMPONENTS | | P1 | Backlog | 001-T028 | yes | | |
| 001-T044 | Create `ConversationPanel` feature (message thread, text input, SSE stream rendering, loading indicator) | G-US1-COMPONENTS | | P1 | Backlog | 001-T028 | yes | | |
| 001-T045 | Create `ItineraryView` feature (scrollable DaySection list; loading skeleton, error, empty states) | G-US1-COMPONENTS | | P1 | Backlog | 001-T028 | yes | | |
| 001-T046 | Implement trips and conversation API service functions with TanStack Query hooks | | | P1 | Backlog | 001-T025 | no | | |
| 001-T047 | Implement Generate page (new trip form → ConversationPanel → ItineraryView on itinerary_ready) | G-US1-PAGES | | P1 | Backlog | 001-T044, 001-T045, 001-T046 | no | | |
| 001-T048 | Implement Trip detail page (ItineraryView, action bar: edit title, delete trip) | G-US1-PAGES | | P1 | Backlog | 001-T045, 001-T046 | no | | |
| 001-T049 | Implement Dashboard page (TripCard grid, "New Trip" CTA, empty state illustration) | G-US1-PAGES | | P1 | Backlog | 001-T041, 001-T046 | no | | |
| 001-T050 | Add Playwright E2E spec: register → subscribe → generate itinerary via conversation → verify Day 1 | | | P1 | Backlog | 001-T047, 001-T049, 001-T040 | no | | |

#### Phase 4 — User Story 2: Experienced Traveler, Off-the-Beaten-Path (Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 001-T051 | Create `TravelStyleSelector` composite (multi-select chip group; keyboard accessible) | | | P2 | Backlog | 001-T028 | yes | | |
| 001-T052 | Integrate `TravelStyleSelector` into Generate page; pass selected styles on trip creation | | | P2 | Backlog | 001-T051 | no | | |
| 001-T053 | Update Trip service to persist TripTravelStyle join records on create/update | G-US2-BACKEND | | P2 | Backlog | 001-T035 | no | | |
| 001-T054 | Update Itinerary service: inject travel styles into Claude prompt; non-mainstream recs per day | G-US2-BACKEND | | P2 | Backlog | 001-T037 | no | | |
| 001-T055 | Update Trip HTTP handlers to accept travel_styles[] in POST/PUT; return in GET | G-US2-BACKEND | | P2 | Backlog | 001-T038 | no | | |
| 001-T056 | Update ItineraryView to display travel style badges at trip header level | | | P2 | Backlog | 001-T045 | yes | | |
| 001-T057 | Add E2E spec: gastronomy style → verify ≥ 1 food activity per day | | | P2 | Backlog | 001-T050 | no | | |

#### Phase 5 — User Story 3: Planner Enriches Existing Plan (Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 001-T058 | Update Itinerary service: detect anchor mode from enumerated place lists; inject anchors into Claude | G-US3-BACKEND | | P2 | Backlog | 001-T037 | no | | |
| 001-T059 | Update Conversation service: return extracted anchor places for user confirmation | G-US3-BACKEND | | P2 | Backlog | 001-T036 | no | | |
| 001-T060 | Update ConversationPanel to render anchor places confirmation card | | | P2 | Backlog | 001-T044 | yes | | |
| 001-T061 | Update Generate page to handle anchor confirmation step before final generation | | | P2 | Backlog | 001-T047 | no | | |
| 001-T062 | Add E2E spec: provide 5 anchor places → confirm → assert all 5 appear in itinerary | | | P2 | Backlog | 001-T050 | no | | |

#### Phase 6 — User Story 4: Group Collaboration on Shared Itinerary (Priority: P3)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 001-T063 | Extend Trip repository with collaborator ops (CreateCollaborator, AcceptCollaborator, CountActive) | G-US4-REPOS | | P3 | Backlog | 001-T033 | yes | | |
| 001-T064 | Implement Suggestion repository (CreateSuggestion, ListByTrip, UpdateStatus) | G-US4-REPOS | | P3 | Backlog | 001-T009 | yes | | |
| 001-T065 | Extend Trip service: InviteCollaborator (enforce basic plan limit), AcceptInvitation, Remove | | | P3 | Backlog | 001-T035, 001-T063 | no | | |
| 001-T066 | Implement Suggestion service (Submit, Approve applies Activity update, Reject preserves history) | | | P3 | Backlog | 001-T064 | no | | |
| 001-T067 | Implement Collaborator HTTP handlers (POST/DELETE /trips/:id/collaborators; admin only) | G-US4-HANDLERS | | P3 | Backlog | 001-T065 | no | | |
| 001-T068 | Implement Suggestion HTTP handlers (GET/POST /suggestions, PATCH approve/reject) | G-US4-HANDLERS | | P3 | Backlog | 001-T066 | no | | |
| 001-T069 | Register collaborator and suggestion routes in Chi router | | | P3 | Backlog | 001-T067, 001-T068 | no | | |
| 001-T070 | Create `SuggestionBubble` composite (author, content, status badge, approve/reject buttons) | G-US4-COMPONENTS | | P3 | Backlog | 001-T028 | yes | | |
| 001-T071 | Create `SuggestionQueue` feature (SuggestionBubble list, status filter, loading/empty states) | G-US4-COMPONENTS | | P3 | Backlog | 001-T070 | yes | | |
| 001-T072 | Create `CollaboratorInvite` composite (email input, invite button, partner badge, plan limit warning) | G-US4-COMPONENTS | | P3 | Backlog | 001-T028 | yes | | |
| 001-T073 | Implement suggestion/collaborator API service functions with TanStack Query hooks | | | P3 | Backlog | 001-T046 | no | | |
| 001-T074 | Integrate SuggestionQueue and CollaboratorInvite into Trip detail page | | | P3 | Backlog | 001-T071, 001-T072, 001-T073 | no | | |
| 001-T075 | Add E2E spec: admin invites partner → suggestion submitted → admin approves/rejects flow | | | P3 | Backlog | 001-T050 | no | | |

#### Phase 7 — Polish & Cross-Cutting Concerns (Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 001-T076 | Add backend lint CI workflow (golangci-lint on push and PR) | G-CI-WORKFLOWS | | P2 | Backlog | 001-T001 | yes | | |
| 001-T077 | Add frontend lint + type-check CI workflow (ESLint + tsc --noEmit) | G-CI-WORKFLOWS | | P2 | Backlog | 001-T001 | yes | | |
| 001-T078 | Add backend test CI workflow (go test -race against PostgreSQL service container) | G-CI-WORKFLOWS | | P2 | Backlog | 001-T001 | yes | | |
| 001-T079 | Add frontend test CI workflow (Vitest) | G-CI-WORKFLOWS | | P2 | Backlog | 001-T001 | yes | | |
| 001-T080 | Create production-ready backend Dockerfile (multi-stage, non-root user, migrations on start) | | | P2 | Backlog | 001-T023 | no | | |
| 001-T081 | Write README.md (overview, prerequisites, setup commands, architecture reference, workflow) | | | P2 | Backlog | 001-T001 | no | | |

---

### Spec 002 — Non-Functional Requirements and System Constraints &nbsp; `specs/002-nfr-system-constraints/tasks.md`

> Cross-spec note: Tasks in phases 5 and 7 extend CI workflow files created by 001-T076–T079 and
> depend on application code from spec 001. See the Dependencies section in `specs/002-nfr-system-constraints/tasks.md`.

#### Phase 1 — Setup (NFR Dependencies and Config)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 002-T001 | Add `github.com/microcosm-cc/bluemonday` HTML sanitisation dependency to Go module | | | P1 | Backlog | - | no | | |
| 002-T002 | Add `@axe-core/playwright` and `@lhci/cli` as frontend dev dependencies | G-NFR-CONFIG | | P1 | Backlog | - | yes | | |
| 002-T003 | Create `backend/config/prompt-rules.yml` with 5 seed deny-list rules (instruction-override, role-switching, jailbreak-prefix, etc.) | G-NFR-CONFIG | | P1 | Backlog | - | yes | | |
| 002-T004 | Create `lighthouserc.yml` with LHCI assertion thresholds (a11y ≥ 0.9, LCP ≤ 2500 ms, CLS ≤ 0.1, INP ≤ 200 ms) | G-NFR-CONFIG | | P1 | Backlog | - | yes | | |
| 002-T005 | Create `.gitleaks.toml` secret-scanning configuration (scan all committed files, exclude test fixtures) | G-NFR-CONFIG | | P1 | Backlog | - | yes | | |

#### Phase 2 — Foundational: Observability Core (NFR-OBS-001–003)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 002-T006 | Implement `RequestID` Chi middleware (generate UUID v4 if absent; propagate; set X-Request-ID on response) | G-OBS-REQUESTID | | P1 | Backlog | - | no | | |
| 002-T007 | Write unit tests for `RequestID` middleware (generates UUID, propagates existing, sets response header) | G-OBS-REQUESTID | | P1 | Backlog | 002-T006 | no | | |
| 002-T008 | Implement `Logger` Chi middleware using `log/slog` JSON handler (emit StructuredLogEntry per request) | G-OBS-LOGGER | | P1 | Backlog | - | yes | | |
| 002-T009 | Write unit tests for `Logger` middleware (all fields present, duration_ms ≥ 0, user_id absent on unauth) | G-OBS-LOGGER | | P1 | Backlog | 002-T008 | no | | |
| 002-T010 | Implement `GET /healthz` handler returning HealthCheckResponse JSON (status, version, uptime_seconds) | G-OBS-HEALTHZ | | P1 | Backlog | - | yes | | |
| 002-T011 | Write unit tests for `/healthz` handler (200 OK, schema valid, status is "ok", uptime ≥ 0) | G-OBS-HEALTHZ | | P1 | Backlog | 002-T010 | no | | |
| 002-T012 | Register `/healthz` and wire `RequestID` → `Logger` middleware chain globally in Chi router | | | P1 | Backlog | 002-T006, 002-T008, 002-T010, 001-T023 | no | | |

#### Phase 3 — User Story 1: Engineering Team Verifies Performance Under Load (Priority: P1) 🎯

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 002-T013 | Create k6 baseline load test (ramp to 500 VUs, sustain 10 min; assert p95 ≤ 500 ms and error rate < 1%) | G-PERF-K6-TESTS | | P1 | Backlog | 002-T012 | yes | | |
| 002-T014 | Create k6 API latency scenario (constant 100 RPS; assert p95 ≤ 500 ms per non-AI endpoint) | G-PERF-K6-TESTS | | P1 | Backlog | 002-T012 | yes | | |
| 002-T015 | Create `load-test.yml` GitHub Actions workflow (manual `workflow_dispatch`; runs k6 against staging URL) | | | P1 | Backlog | 002-T013 | no | | |
| 002-T016 | Write integration test asserting `/healthz` responds in ≤ 100 ms for 100 sequential calls | | | P1 | Backlog | 002-T010 | no | | |

#### Phase 4 — User Story 2: Accessibility Reviewer Confirms WCAG 2.1 AA (Priority: P1)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 002-T017 | Implement `PrivacyPolicyPage` static component (WCAG-compliant heading, section structure, landmarks) | G-A11Y-PRIVACY-PAGE | | P1 | Backlog | 002-T002 | yes | | |
| 002-T018 | Write Vitest unit test for `PrivacyPolicyPage` (renders heading, ≥ 3 sections, no dangerouslySetInnerHTML) | G-A11Y-PRIVACY-PAGE | | P1 | Backlog | 002-T017 | yes | | |
| 002-T019 | Implement `PrivacyPolicyLink` atom component (accessible `<a>` linking to /privacy-policy) | G-A11Y-PRIVACY-LINK | | P1 | Backlog | 002-T002 | yes | | |
| 002-T020 | Write Vitest unit test for `PrivacyPolicyLink` (correct href, accessible text present) | G-A11Y-PRIVACY-LINK | | P1 | Backlog | 002-T019 | yes | | |
| 002-T021 | Add `/privacy-policy` route to React Router; add `PrivacyPolicyLink` to registration form footer | | | P1 | Backlog | 002-T017, 002-T019 | no | | |
| 002-T022 | Create accessibility E2E helper `checkPageA11y(page)` wrapping `@axe-core/playwright` | | | P1 | Backlog | 002-T002 | yes | | |
| 002-T023 | Add `checkPageA11y(page)` call to every existing Playwright E2E spec; tag with `@accessibility` | | | P1 | Backlog | 002-T022 | no | | |
| 002-T024 | Create `accessibility.yml` GitHub Actions workflow (axe-core Playwright run + lhci autorun; PR gate) | | | P1 | Backlog | 002-T022, 002-T004 | no | | |

#### Phase 5 — User Story 3: Security Reviewer Confirms Security Posture (Priority: P1)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 002-T025 | Implement `PromptValidator` (load rules from YAML; Validate() with substring + regex modes; panic on invalid regex) | G-SEC-PROMPT-VALIDATOR | | P1 | Backlog | 002-T003 | yes | | |
| 002-T026 | Write unit tests for `PromptValidator` (clean prompt passes; 5 seed rules trigger match; disabled rule skipped) | G-SEC-PROMPT-VALIDATOR | | P1 | Backlog | 002-T025 | no | | |
| 002-T027 | Implement `OutputSanitizer` using `bluemonday.UGCPolicy()` (strip HTML/script before storage) | G-SEC-OUTPUT-SANITIZER | | P1 | Backlog | 002-T001 | yes | | |
| 002-T028 | Write unit tests for `OutputSanitizer` (script stripped, event handlers stripped, plain text preserved) | G-SEC-OUTPUT-SANITIZER | | P1 | Backlog | 002-T027 | no | | |
| 002-T029 | Wire `PromptValidator` into itinerary generation handler (return 400 PromptRejectionResponse on match) | | | P1 | Backlog | 002-T025, 001-T038 | no | | |
| 002-T030 | Wire `OutputSanitizer` into itinerary service before every DB INSERT of AI-generated content | | | P1 | Backlog | 002-T027, 001-T037 | no | | |
| 002-T031 | Write integration test for prompt rejection (10 injection payloads → assert 400 + request_id matches header) | | | P1 | Backlog | 002-T029 | yes | | |
| 002-T032 | Implement `DELETE /users/me` auth endpoint (cascade delete all PII; respond 204) | G-SEC-USER-DELETION | | P1 | Backlog | 001-T021 | yes | | |
| 002-T033 | Write integration test for user deletion (register → create trip → DELETE → assert 204 + zero DB rows) | G-SEC-USER-DELETION | | P1 | Backlog | 002-T032 | no | | |
| 002-T034 | Extend `backend-lint.yml` with `gosec -severity high` and `govulncheck` steps (gate on non-zero exit) | G-SEC-CI-GATES | | P1 | Backlog | 001-T076 | yes | | |
| 002-T035 | Extend `frontend-lint.yml` with `npm audit --audit-level=high` step (gate on critical/high findings) | G-SEC-CI-GATES | | P1 | Backlog | 001-T077 | yes | | |
| 002-T036 | Add `gitleaks detect` secret-scanning step to `backend-lint.yml` (every PR and push to main) | | | P1 | Backlog | 001-T076, 002-T005 | no | | |

#### Phase 6 — User Story 4: On-Call Engineer Diagnoses a Production Issue (Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 002-T037 | Write integration test for `Logger` middleware (parse 5 requests' JSON log lines; assert all StructuredLogEntry fields) | G-OBS-VALIDATION | | P2 | Backlog | 002-T008 | yes | | |
| 002-T038 | Write integration test for `RequestID` middleware (no-header → UUID generated; preset header → echoed) | G-OBS-VALIDATION | | P2 | Backlog | 002-T006 | yes | | |
| 002-T039 | Write Playwright E2E test for `/healthz` (X-Request-ID header is UUID; response matches HealthCheckResponse schema) | G-OBS-VALIDATION | | P2 | Backlog | 002-T010 | no | | |
| 002-T040 | Create `backend/config/alerts.yml` defining 5xx error-rate alerting rule (> 1% over 5-minute window) | | | P2 | Backlog | 002-T008 | yes | | |

#### Phase 7 — User Story 5: Engineering Team Confirms Maintainability Standards (Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 002-T041 | Add `@vitest/coverage-v8` and configure coverage thresholds ≥ 80% for `src/components/**` and `src/hooks/**` | G-COVERAGE-FRONTEND | | P2 | Backlog | 002-T002 | yes | | |
| 002-T042 | Add `npm run test:coverage` script running `vitest run --coverage` (fails if thresholds not met) | G-COVERAGE-FRONTEND | | P2 | Backlog | 002-T041 | yes | | |
| 002-T043 | Extend `frontend-test.yml` to run `npm run test:coverage`; upload coverage report as artefact | | | P2 | Backlog | 001-T079, 002-T042 | no | | |
| 002-T044 | Extend `backend-test.yml` to add `-coverprofile` flag and awk assertion for ≥ 80% total coverage | | | P2 | Backlog | 001-T078 | no | | |

#### Phase 8 — Polish & Cross-Cutting Concerns (Priority: P3)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 002-T045 | Create OWASP ZAP baseline scan shell script (`scripts/owasp-zap-scan.sh`; exit 1 on FAIL-level finding) | G-NFR-POLISH | | P3 | Backlog | - | yes | | |
| 002-T046 | Create `backend/config/backup-policy.yml` (RPO ≤ 24 h, daily pg_dump schedule, 7-day retention, restore command) | G-NFR-POLISH | | P3 | Backlog | - | no | | |

---

### Spec 003 — Cloud & Environments Strategy &nbsp; `specs/003-cloud-env-strategy/tasks.md`

> Cross-spec note: This spec provisions infrastructure for application code defined in specs 001-002.
> Tasks T067 and T070 extend CI workflow files created by 001-T076 and 001-T080.

#### Phase 1 — Setup

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 003-T001 | Create top-level `infra/` directory with subdirectories: `terraform/`, `docker/`, `iam/`, `scripts/` | | | P1 | Backlog | - | no | | |
| 003-T002 | Create Terraform root module structure: `main.tf`, `variables.tf`, `outputs.tf`, `terraform.tf` | G-INFRA-SETUP | | P1 | Backlog | 003-T001 | yes | | |
| 003-T003 | Create Terraform workspace-specific variable files: `staging.tfvars`, `production.tfvars` | G-INFRA-SETUP | | P1 | Backlog | 003-T001 | yes | | |
| 003-T004 | Create module directories: `modules/vpc/`, `modules/ecs/`, `modules/rds/`, `modules/alb/`, `modules/s3-cloudfront/`, `modules/iam/` | G-INFRA-SETUP | | P1 | Backlog | 003-T001 | yes | | |
| 003-T005 | Install Terraform 1.5+ locally and verify version: `terraform version` | G-INFRA-TOOLS | | P1 | Backlog | 003-T001 | yes | | |
| 003-T006 | Install AWS CLI 2.x and configure default region (us-east-1) | G-INFRA-TOOLS | | P1 | Backlog | 003-T001 | yes | | |
| 003-T007 | Install Docker 24.x for backend container builds | G-INFRA-TOOLS | | P1 | Backlog | 003-T001 | yes | | |
| 003-T008 | Add `infra/terraform/.gitignore` excluding: `*.tfstate`, `*.tfstate.backup`, `.terraform/`, `*.tfplan`, `*.tfvars` | G-INFRA-SETUP | | P1 | Backlog | 003-T001 | yes | | |

#### Phase 2 — Foundational (Terraform State & OIDC Setup)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 003-T009 | Create Terraform state bootstrap script: provisions S3 bucket (`trAIveler-terraform-state`) with versioning and encryption, DynamoDB table (`trAIveler-terraform-locks`) | | | P1 | Backlog | 003-T001 | no | | |
| 003-T010 | Create Terraform backend configuration using S3 bucket and DynamoDB table from T009 | | | P1 | Backlog | 003-T009 | no | | |
| 003-T011 | Create AWS provider configuration with default tags (Project, Environment, ManagedBy, CostCenter); region: us-east-1 | G-INFRA-TF-CONFIG | | P1 | Backlog | 003-T001 | yes | | |
| 003-T012 | Create GitHub Actions OIDC trust policy JSON templates for staging and production environments | G-INFRA-TF-CONFIG | | P1 | Backlog | 003-T001 | yes | | |
| 003-T013 | Create IAM OIDC identity provider Terraform module: configures GitHub OIDC provider in AWS, creates IAM roles for staging and production | | | P1 | Backlog | 003-T012 | no | | |

#### Phase 3 — User Story 1: Infrastructure Provisioning & Environment Setup (Priority: P1) 🎯 MVP

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 003-T014 | Create VPC module: provisions VPC with CIDR from tfvars, 2 public subnets (ALB), 2 private subnets (ECS, RDS), internet gateway, route tables | G-INFRA-VPC-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T015 | Create VPC security groups: `alb-sg`, `ecs-sg`, `rds-sg` | G-INFRA-VPC-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T016 | Create VPC NAT resource: conditional NAT instance (staging) or NAT Gateway (production) | G-INFRA-VPC-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T017 | Define VPC module variables: `environment`, `vpc_cidr`, `availability_zones`, `nat_gateway_type` | G-INFRA-VPC-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T018 | Define VPC module outputs: `vpc_id`, `public_subnet_ids`, `private_subnet_ids`, security group IDs | G-INFRA-VPC-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T019 | Create ALB module: provisions Application Load Balancer in public subnets, HTTP/HTTPS listeners, target group for ECS | G-INFRA-ALB-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T020 | Define ALB module variables: `environment`, `vpc_id`, `public_subnet_ids`, `alb_security_group_id`, `certificate_arn` | G-INFRA-ALB-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T021 | Define ALB module outputs: `alb_arn`, `alb_dns_name`, `target_group_arn` | G-INFRA-ALB-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T022 | Create RDS module: provisions PostgreSQL 15.4 instance with environment-specific configuration | G-INFRA-RDS-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T023 | Define RDS module variables: `environment`, `instance_identifier`, `instance_class`, `vpc_id`, `private_subnet_ids`, security group, Multi-AZ, backup retention, database name, credentials | G-INFRA-RDS-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T024 | Define RDS module outputs: `endpoint`, `instance_id`, `database_name` | G-INFRA-RDS-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T025 | Create ECS module: provisions ECS cluster, task definition (Go backend container), ECS service with ALB integration, auto-scaling | G-INFRA-ECS-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T026 | Create ECS auto-scaling configuration: target tracking scaling policy (CPU 70%), cooldown periods, min/max capacity | G-INFRA-ECS-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T027 | Define ECS module variables: `environment`, `cluster_name`, `vpc_id`, `private_subnet_ids`, ALB target group ARN, security group, task sizing, capacity, ECR image URI, secret ARNs | G-INFRA-ECS-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T028 | Define ECS module outputs: `ecs_cluster_arn`, `ecs_service_name`, `task_definition_arn`, `service_arn` | G-INFRA-ECS-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T029 | Create S3+CloudFront module: provisions S3 bucket for frontend static assets with website hosting, CloudFront distribution | G-INFRA-CDN-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T030 | Define S3+CloudFront module variables: `environment`, `bucket_name`, `cloudfront_price_class` | G-INFRA-CDN-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T031 | Define S3+CloudFront module outputs: `s3_bucket_name`, `cloudfront_distribution_id`, `cloudfront_domain_name` | G-INFRA-CDN-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T032 | Create IAM module: provisions ECS task execution role, ECS task role, GitHub Actions roles | G-INFRA-IAM-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T033 | Define IAM module variables: `environment`, `ecr_repository_arn`, `secrets_manager_arns`, `s3_bucket_arns` | G-INFRA-IAM-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T034 | Define IAM module outputs: `ecs_task_execution_role_arn`, `ecs_task_role_arn`, `github_actions_role_arn` | G-INFRA-IAM-MODULE | | P1 | Backlog | 003-T009 | yes | | |
| 003-T035 | Wire VPC module in root `main.tf`: call `modules/vpc` with staging/production-specific CIDR blocks and NAT type | | | P1 | Backlog | 003-T014 | no | | |
| 003-T036 | Wire ALB module in root `main.tf`: call `modules/alb` with VPC outputs (public subnets, security group) | | | P1 | Backlog | 003-T019 | no | | |
| 003-T037 | Wire RDS module in root `main.tf`: call `modules/rds` with VPC outputs and environment-specific instance class, Multi-AZ flag | | | P1 | Backlog | 003-T022 | no | | |
| 003-T038 | Wire ECS module in root `main.tf`: call `modules/ecs` with VPC outputs, ALB target group ARN, environment-specific task sizing | | | P1 | Backlog | 003-T025 | no | | |
| 003-T039 | Wire S3+CloudFront module in root `main.tf`: call `modules/s3-cloudfront` with environment-specific bucket name and CloudFront price class | | | P1 | Backlog | 003-T029 | no | | |
| 003-T040 | Wire IAM module in root `main.tf`: call `modules/iam` with resource ARNs (ECR, Secrets Manager, S3) from other modules | | | P1 | Backlog | 003-T032 | no | | |
| 003-T041 | Define root module variables: `environment`, `aws_region`, `vpc_cidr`, `availability_zones`, ECS task sizing, RDS config, NAT type, cost budget, tags | | | P1 | Backlog | 003-T035 | no | | |
| 003-T042 | Define root module outputs: all outputs from modules (VPC, ALB, RDS, ECS, S3+CloudFront, IAM) | | | P1 | Backlog | 003-T035 | no | | |
| 003-T043 | Populate `staging.tfvars` with staging-specific values: vpc_cidr, ECS task sizing, RDS instance class, NAT instance, cost budget | | | P1 | Backlog | 003-T041 | yes | | |
| 003-T044 | Populate `production.tfvars` with production-specific values: vpc_cidr, ECS task sizing, RDS Multi-AZ, NAT Gateway, cost budget | | | P1 | Backlog | 003-T041 | yes | | |

#### Phase 4 — User Story 2: Secrets & Configuration Management (Priority: P1)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 003-T045 | Create secrets initialization script: uses AWS CLI to create secrets in Secrets Manager with naming pattern `${environment}/${service}/${secret_name}` | G-INFRA-SECRETS | | P1 | Backlog | 003-T009 | yes | | |
| 003-T046 | Create Terraform data sources for Secrets Manager: reference existing secrets for database URL, Anthropic API key, JWT secret | G-INFRA-SECRETS | | P1 | Backlog | 003-T009 | yes | | |
| 003-T047 | Update ECS task definition in T025 to reference secret ARNs from T046 in `secrets` block | | | P1 | Backlog | 003-T025, 003-T046 | no | | |

#### Phase 5 — User Story 3: CI/CD Pipeline & Deployment Promotion (Priority: P1)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 003-T048 | Create `terraform-plan.yml` GitHub Actions workflow: triggers on PR, validates Terraform, posts plan output | G-INFRA-TF-WORKFLOWS | | P1 | Backlog | 003-T010 | yes | | |
| 003-T049 | Create `terraform-apply.yml` GitHub Actions workflow: auto-deploy staging on main merge, manual production deployment | G-INFRA-TF-WORKFLOWS | | P1 | Backlog | 003-T010 | yes | | |
| 003-T050 | Create multi-stage Dockerfile for Go backend: stage 1 builds Go binary targeting `linux/arm64`, stage 2 uses alpine base | G-INFRA-DOCKER | | P1 | Backlog | 003-T001 | yes | | |
| 003-T051 | Create `.dockerignore` for backend: excludes `*.md`, `tests/`, `.git/`, `.env*` | G-INFRA-DOCKER | | P1 | Backlog | 003-T001 | yes | | |
| 003-T052 | Create `backend-deploy.yml` GitHub Actions workflow: builds Docker image for linux/arm64, pushes to ECR, updates ECS service | | | P1 | Backlog | 003-T050, 003-T013 | no | | |
| 003-T053 | Create `frontend-deploy.yml` GitHub Actions workflow: builds React production bundle, syncs to S3, invalidates CloudFront | | | P1 | Backlog | 003-T013 | yes | | |
| 003-T054 | Update ECS service configuration in T025 to enable deployment circuit breaker with automatic rollback | | | P1 | Backlog | 003-T025 | no | | |

#### Phase 6 — User Story 4: Cost Monitoring & Budget Controls (Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 003-T055 | Create AWS Budget Terraform resource for staging: monthly budget $200, alerts at 80% and 100%, SNS topic | G-INFRA-COST-MONITORING | | P2 | Backlog | 003-T009 | yes | | |
| 003-T056 | Create AWS Budget Terraform resource for production: monthly budget $400, alerts at 80% and 100%, SNS topic | G-INFRA-COST-MONITORING | | P2 | Backlog | 003-T009 | yes | | |
| 003-T057 | Verify all Terraform modules use default tags from provider: validate tags propagate to all resources | G-INFRA-COST-MONITORING | | P2 | Backlog | 003-T014 | no | | |
| 003-T058 | Create AWS Config rule to enforce required tags: rule triggers on resource creation, fails if missing Project, Environment, ManagedBy tags | | | P2 | Backlog | 003-T009 | yes | | |

#### Phase 7 — User Story 5: Infrastructure Observability & Health Monitoring (Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 003-T059 | Create CloudWatch log groups for ECS tasks: `/ecs/trAIveler-backend-staging` (retention 7 days), `/ecs/trAIveler-backend-production` (retention 30 days) | G-INFRA-OBSERVABILITY | | P2 | Backlog | 003-T009 | yes | | |
| 003-T060 | Create CloudWatch dashboard for staging: ECS task count, CPU/memory utilization, ALB request count, RDS metrics | G-INFRA-OBSERVABILITY | | P2 | Backlog | 003-T009 | yes | | |
| 003-T061 | Create CloudWatch dashboard for production: same widgets as staging but for production resources | G-INFRA-OBSERVABILITY | | P2 | Backlog | 003-T009 | yes | | |
| 003-T062 | Create CloudWatch metric alarm for ECS high CPU: triggers when staging ECS service CPU > 80% for 5 minutes, sends SNS notification | G-INFRA-OBSERVABILITY | | P2 | Backlog | 003-T009 | yes | | |
| 003-T063 | Create CloudWatch metric alarm for RDS high connections: triggers when staging RDS connections > 80 for 5 minutes, sends SNS notification | G-INFRA-OBSERVABILITY | | P2 | Backlog | 003-T009 | yes | | |
| 003-T064 | Create CloudWatch metric alarm for ALB unhealthy targets: triggers when staging ALB has 0 healthy targets for 2 minutes, sends SNS notification | G-INFRA-OBSERVABILITY | | P2 | Backlog | 003-T009 | yes | | |

#### Phase 8 — Polish & Cross-Cutting Concerns (Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 003-T065 | Add `tflint` configuration file: enables AWS plugin, sets minimum Terraform version 1.5.0 | G-INFRA-POLISH | | P2 | Backlog | 003-T001 | yes | | |
| 003-T066 | Add `tfsec` configuration file: enables all HIGH severity checks, ignores known false positives | G-INFRA-POLISH | | P2 | Backlog | 003-T001 | yes | | |
| 003-T067 | Extend `backend-lint.yml` workflow (from spec 001-T076) to include Docker linting: runs `hadolint infra/docker/backend/Dockerfile` | | | P2 | Backlog | 001-T076, 003-T050 | no | | Extends 001 workflow |
| 003-T068 | Create infrastructure documentation README: overview of Terraform modules, quick start guide, links to contracts and research | G-INFRA-POLISH | | P2 | Backlog | 003-T001 | yes | | |
| 003-T069 | Create Terraform module documentation: auto-generate module docs using `terraform-docs` for each module | G-INFRA-POLISH | | P2 | Backlog | 003-T014 | yes | | |
| 003-T070 | Move backend Dockerfile from `backend/Dockerfile` (if exists from spec 001-T002) to `infra/docker/backend/Dockerfile` | | | P2 | Backlog | 001-T080, 003-T050 | no | | Relocates 001 artifact |
| 003-T071 | Create `.editorconfig` for Terraform files: indent 2 spaces, trim trailing whitespace | G-INFRA-POLISH | | P2 | Backlog | 003-T001 | yes | | |
| 003-T072 | Add pre-commit hook configuration: runs `terraform fmt` on staged `.tf` files | | | P2 | Backlog | 003-T001 | yes | | |

---

### Spec 004 — Security & Authentication/Authorization Model &nbsp; `specs/004-security-auth-model/tasks.md`

> Cross-spec note: Security foundation tasks integrate with application auth (001), NFR validation (002),
> and cloud infrastructure (003). JWT signing keys stored in AWS Secrets Manager (003-T045–047). CloudWatch
> alarms/metrics integrate with 003 observability resources (003-T059–064). Prompt validation extends 002
> security requirements.

#### Phase 1 — Setup (Shared Infrastructure)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 004-T001 | Create backend security package structure: `backend/internal/auth/`, `backend/internal/authorization/`, `backend/internal/validation/`, `backend/internal/concurrency/`, `backend/internal/observability/` | | | P1 | Backlog | - | no | | |
| 004-T002 | Create frontend security structure: `frontend/src/lib/auth.ts`, `frontend/src/lib/authContext.tsx`, `frontend/src/hooks/`, `frontend/src/components/ProtectedRoute.tsx` | G-SEC-FRONTEND-STRUCTURE | | P1 | Backlog | - | yes | | |
| 004-T003 | Add backend dependencies: `go get github.com/golang-jwt/jwt/v5`, `go get golang.org/x/crypto/bcrypt`, `go get github.com/microcosm-cc/bluemonday` | G-SEC-BACKEND-DEPS | | P1 | Backlog | - | yes | | |
| 004-T004 | Add frontend dependencies: `npm install @tanstack/react-query zustand` (if not already present) | G-SEC-FRONTEND-DEPS | | P1 | Backlog | - | yes | | |
| 004-T005 | Create test directory structure: `backend/tests/integration/`, `backend/tests/security/`, `e2e/tests/` | G-SEC-TEST-STRUCTURE | | P1 | Backlog | - | yes | | |

#### Phase 2 — Foundational (Blocking Prerequisites)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 004-T006 | Create `users` table migration: `backend/migrations/001_create_users_table.sql` with columns per data-model.md | G-SEC-DB-MIGRATIONS | | P1 | Backlog | - | yes | | |
| 004-T007 | Create `refresh_tokens` table migration: `backend/migrations/002_create_refresh_tokens_table.sql` | G-SEC-DB-MIGRATIONS | | P1 | Backlog | - | yes | | |
| 004-T008 | Create `jwt_signing_keys` table migration: `backend/migrations/003_create_jwt_signing_keys_table.sql` | G-SEC-DB-MIGRATIONS | | P1 | Backlog | - | yes | | |
| 004-T009 | Create `security_events` table migration: `backend/migrations/004_create_security_events_table.sql` | G-SEC-DB-MIGRATIONS | | P1 | Backlog | - | yes | | |
| 004-T010 | Create `trips.version` column migration: `backend/migrations/005_add_version_to_trips.sql` | G-SEC-DB-MIGRATIONS | | P1 | Backlog | - | yes | | |
| 004-T011 | Create `itinerary_items.version` column migration: `backend/migrations/006_add_version_to_itinerary_items.sql` | G-SEC-DB-MIGRATIONS | | P1 | Backlog | - | yes | | |
| 004-T012 | Implement password hashing utility in `backend/internal/auth/password.go` (bcrypt cost 12) | G-SEC-CORE-UTILITIES | | P1 | Backlog | - | yes | | |
| 004-T013 | Implement correlation ID generator in `backend/internal/observability/correlation.go` | G-SEC-CORE-UTILITIES | | P1 | Backlog | - | yes | | |
| 004-T014 | Implement structured logger in `backend/internal/observability/logger.go` with CloudWatch JSON output | G-SEC-CORE-UTILITIES | | P1 | Backlog | - | yes | | |

#### Phase 3 — User Story 1: Backend Engineer Implements Secure API Endpoint (Priority: P1) 🎯 MVP

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 004-T015 | Write unit test for JWT generation in `backend/internal/auth/jwt_test.go` | G-SEC-JWT-TESTS | | P1 | Backlog | 004-T014 | yes | | RED phase |
| 004-T016 | Write unit test for JWT validation in `backend/internal/auth/jwt_test.go` | G-SEC-JWT-TESTS | | P1 | Backlog | 004-T014 | yes | | RED phase |
| 004-T017 | Write unit test for multi-key JWT rotation in `backend/internal/auth/jwt_test.go` | G-SEC-JWT-TESTS | | P1 | Backlog | 004-T014 | yes | | RED phase |
| 004-T018 | Implement JWT token generation in `backend/internal/auth/jwt.go` (RS256, 24h exp) | G-SEC-JWT-IMPL | | P1 | Backlog | 004-T015 | no | | GREEN phase |
| 004-T019 | Implement JWT token validation in `backend/internal/auth/jwt.go` (multi-key support) | G-SEC-JWT-IMPL | | P1 | Backlog | 004-T016 | no | | GREEN phase |
| 004-T020 | Implement refresh token generation in `backend/internal/auth/jwt.go` (32-byte random) | G-SEC-JWT-IMPL | | P1 | Backlog | 004-T015 | yes | | GREEN phase |
| 004-T021 | Implement key fetching from DB in `backend/internal/auth/jwt.go` (5-min cache) | G-SEC-JWT-IMPL | | P1 | Backlog | 004-T017 | yes | | GREEN phase |
| 004-T022 | Write integration test for auth middleware in `backend/tests/integration/auth_test.go` | G-SEC-AUTH-MIDDLEWARE-TESTS | | P1 | Backlog | 004-T018 | yes | | RED phase |
| 004-T023 | Write integration test for expired token in `backend/tests/integration/auth_test.go` | G-SEC-AUTH-MIDDLEWARE-TESTS | | P1 | Backlog | 004-T018 | yes | | RED phase |
| 004-T024 | Implement authentication middleware in `backend/internal/auth/middleware.go` | | | P1 | Backlog | 004-T022 | no | | GREEN phase |
| 004-T025 | Register authentication middleware in `backend/cmd/server/main.go` (all routes except public) | | | P1 | Backlog | 004-T024 | no | | GREEN phase |
| 004-T026 | Write integration test for registration in `backend/tests/integration/auth_test.go` (bcrypt, 409 Conflict) | G-SEC-AUTH-HANDLERS-TESTS | | P1 | Backlog | 004-T024 | yes | | RED phase |
| 004-T027 | Write integration test for login in `backend/tests/integration/auth_test.go` (cookies, last_login_at) | G-SEC-AUTH-HANDLERS-TESTS | | P1 | Backlog | 004-T024 | yes | | RED phase |
| 004-T028 | Write integration test for refresh in `backend/tests/integration/auth_test.go` (new access token) | G-SEC-AUTH-HANDLERS-TESTS | | P1 | Backlog | 004-T024 | yes | | RED phase |
| 004-T029 | Write integration test for logout in `backend/tests/integration/auth_test.go` (cookies cleared, token revoked) | G-SEC-AUTH-HANDLERS-TESTS | | P1 | Backlog | 004-T024 | yes | | RED phase |
| 004-T030 | Implement registration handler in `backend/internal/auth/handler.go` (JSON schema, email unique, bcrypt) | G-SEC-AUTH-HANDLERS-IMPL | | P1 | Backlog | 004-T026 | no | | GREEN phase |
| 004-T031 | Implement login handler in `backend/internal/auth/handler.go` (SHA-256 refresh token, cookies) | G-SEC-AUTH-HANDLERS-IMPL | | P1 | Backlog | 004-T027 | no | | GREEN phase |
| 004-T032 | Implement refresh handler in `backend/internal/auth/handler.go` (SHA-256 lookup, new access token) | G-SEC-AUTH-HANDLERS-IMPL | | P1 | Backlog | 004-T028 | no | | GREEN phase |
| 004-T033 | Implement logout handler in `backend/internal/auth/handler.go` (revoke refresh token, clear cookies) | G-SEC-AUTH-HANDLERS-IMPL | | P1 | Backlog | 004-T029 | no | | GREEN phase |
| 004-T034 | Implement password change handler in `backend/internal/auth/handler.go` (invalidate_all_sessions option) | G-SEC-AUTH-HANDLERS-IMPL | | P1 | Backlog | 004-T030 | yes | | GREEN phase |
| 004-T035 | Implement account deletion handler in `backend/internal/auth/handler.go` (202 Accepted, +30 days PII removal) | G-SEC-AUTH-HANDLERS-IMPL | | P1 | Backlog | 004-T030 | yes | | GREEN phase |
| 004-T036 | Write unit test for RBAC in `backend/internal/authorization/rbac_test.go` (admin/partner permissions) | G-SEC-RBAC-TESTS | | P1 | Backlog | 004-T030 | yes | | RED phase |
| 004-T037 | Write integration test for authorization middleware in `backend/tests/integration/rbac_test.go` (403 Forbidden) | G-SEC-RBAC-TESTS | | P1 | Backlog | 004-T030 | yes | | RED phase |
| 004-T038 | Implement RBAC permission checker in `backend/internal/authorization/rbac.go` (FR-010 through FR-021 rules) | G-SEC-RBAC-IMPL | | P1 | Backlog | 004-T036 | no | | GREEN phase |
| 004-T039 | Implement authorization middleware in `backend/internal/authorization/middleware.go` (RequireRole) | G-SEC-RBAC-IMPL | | P1 | Backlog | 004-T037 | no | | GREEN phase |
| 004-T040 | Apply authorization middleware to admin routes in `backend/cmd/server/main.go` (POST/PUT/DELETE trips) | | | P1 | Backlog | 004-T039 | no | | GREEN phase |
| 004-T041 | Write security test for SQL injection in `backend/tests/security/injection_test.go` (OWASP patterns) | G-SEC-VALIDATION-TESTS | | P1 | Backlog | 004-T030 | yes | | RED phase |
| 004-T042 | Write security test for XSS in `backend/tests/security/injection_test.go` (script, onclick, iframe) | G-SEC-VALIDATION-TESTS | | P1 | Backlog | 004-T030 | yes | | RED phase |
| 004-T043 | Write security test for path traversal in `backend/tests/security/injection_test.go` (../ patterns) | G-SEC-VALIDATION-TESTS | | P1 | Backlog | 004-T030 | yes | | RED phase |
| 004-T044 | Implement SQL injection detection in `backend/internal/validation/injection.go` (regex patterns) | G-SEC-VALIDATION-IMPL | | P1 | Backlog | 004-T041 | no | | GREEN phase |
| 004-T045 | Implement XSS detection in `backend/internal/validation/injection.go` (regex patterns) | G-SEC-VALIDATION-IMPL | | P1 | Backlog | 004-T042 | no | | GREEN phase |
| 004-T046 | Implement path traversal detection in `backend/internal/validation/injection.go` (../, absolute paths) | G-SEC-VALIDATION-IMPL | | P1 | Backlog | 004-T043 | no | | GREEN phase |
| 004-T047 | Implement JSON schema validator in `backend/internal/validation/schema.go` (predefined schemas) | G-SEC-VALIDATION-IMPL | | P1 | Backlog | 004-T044 | yes | | GREEN phase |
| 004-T048 | Create validation middleware in `backend/internal/validation/middleware.go` (400 on failure) | | | P1 | Backlog | 004-T044 | no | | GREEN phase |
| 004-T049 | Register validation middleware globally in `backend/cmd/server/main.go` (before auth middleware) | | | P1 | Backlog | 004-T048 | no | | GREEN phase |
| 004-T050 | Write security test for prompt injection in `backend/tests/security/llm_prompts_test.go` (instruction override) | G-SEC-PROMPT-TESTS | | P1 | Backlog | 004-T048 | yes | | RED phase |
| 004-T051 | Write security test for off-topic prompts in `backend/tests/security/llm_prompts_test.go` (finance, medical, code) | G-SEC-PROMPT-TESTS | | P1 | Backlog | 004-T048 | yes | | RED phase |
| 004-T052 | Implement prompt validation in `backend/internal/validation/prompt.go` (regex patterns, correlation ID) | G-SEC-PROMPT-IMPL | | P1 | Backlog | 004-T050 | no | | GREEN phase |
| 004-T053 | Create prompt validation endpoint in `backend/internal/validation/handler.go` (internal, 400 if rejected) | G-SEC-PROMPT-IMPL | | P1 | Backlog | 004-T051 | no | | GREEN phase |
| 004-T054 | Write unit test for HTML sanitization in `backend/internal/validation/sanitize_test.go` (script stripped) | G-SEC-SANITIZATION-TESTS | | P1 | Backlog | 004-T048 | yes | | RED phase |
| 004-T055 | Write unit test for dangerous URLs in `backend/internal/validation/sanitize_test.go` (javascript: removed) | G-SEC-SANITIZATION-TESTS | | P1 | Backlog | 004-T048 | yes | | RED phase |
| 004-T056 | Implement AI output sanitizer in `backend/internal/validation/sanitize.go` (bluemonday strict policy) | G-SEC-SANITIZATION-IMPL | | P1 | Backlog | 004-T054 | no | | GREEN phase |
| 004-T057 | Apply sanitization in AI integration layer before DB persist/client return | | | P1 | Backlog | 004-T056 | no | | GREEN phase |
| 004-T058 | Write integration test for optimistic locking in `backend/tests/integration/concurrency_test.go` (409 Conflict) | G-SEC-CONCURRENCY-TESTS | | P1 | Backlog | 004-T056 | yes | | RED phase |
| 004-T059 | Implement version checker in `backend/internal/concurrency/versioning.go` (CheckVersion) | G-SEC-CONCURRENCY-IMPL | | P1 | Backlog | 004-T058 | no | | GREEN phase |
| 004-T060 | Implement version updater in `backend/internal/concurrency/versioning.go` (IncrementVersion) | G-SEC-CONCURRENCY-IMPL | | P1 | Backlog | 004-T058 | no | | GREEN phase |
| 004-T061 | Create versioning middleware in `backend/internal/concurrency/middleware.go` (If-Match header) | G-SEC-CONCURRENCY-IMPL | | P1 | Backlog | 004-T058 | no | | GREEN phase |
| 004-T062 | Apply versioning middleware to trip routes in `backend/cmd/server/main.go` (PUT/DELETE trips, items) | | | P1 | Backlog | 004-T061 | no | | GREEN phase |
| 004-T063 | Write integration test for security logging in `backend/tests/integration/observability_test.go` (CloudWatch structure) | G-SEC-LOGGING-TESTS | | P1 | Backlog | 004-T061 | yes | | RED phase |
| 004-T064 | Write integration test for CloudWatch metrics in `backend/tests/integration/observability_test.go` (metric increments) | G-SEC-LOGGING-TESTS | | P1 | Backlog | 004-T061 | yes | | RED phase |
| 004-T065 | Implement CloudWatch logger in `backend/internal/observability/logger.go` (LogSecurityEvent, JSON) | G-SEC-LOGGING-IMPL | | P1 | Backlog | 004-T063 | no | | GREEN phase |
| 004-T066 | Implement CloudWatch metrics emitter in `backend/internal/observability/metrics.go` (EmitMetric) | G-SEC-LOGGING-IMPL | | P1 | Backlog | 004-T064 | no | | GREEN phase |
| 004-T067 | Integrate logging into all handlers (auth, authz, validation) per FR-047–049 | | | P1 | Backlog | 004-T065 | no | | GREEN phase |
| 004-T068 | Integrate metrics into all handlers (call EmitMetric after LogSecurityEvent) | | | P1 | Backlog | 004-T066 | no | | GREEN phase |
| 004-T069 | Write unit test for secrets retrieval in `backend/internal/observability/secrets_test.go` (5-min cache, no fallback) | G-SEC-SECRETS-TESTS | | P1 | Backlog | 004-T066 | yes | | RED phase |
| 004-T070 | Implement AWS Secrets Manager client in `backend/pkg/secrets/manager.go` (IAM role auth, cache) | G-SEC-SECRETS-IMPL | | P1 | Backlog | 004-T069, 003-T045 | no | | GREEN phase, needs 003 |
| 004-T071 | Integrate secrets manager in JWT key loader (fetch private keys from Secrets Manager by ARN) | | | P1 | Backlog | 004-T070 | no | | GREEN phase |
| 004-T072 | Add secret refresh timer in `backend/cmd/server/main.go` (5-min goroutine) | | | P1 | Backlog | 004-T071 | no | | GREEN phase |
| 004-T073 | Create CloudWatch Logs retention policy via Terraform (30 days) for `/traivelr/staging/security` | | | P1 | Backlog | 003-T059 | no | | Infra integration |
| 004-T074 | Create CloudWatch Alarm for auth failures via Terraform (>100/min for 5 min, SNS notify) | | | P1 | Backlog | 003-T062 | yes | | Infra integration |
| 004-T075 | Create CloudWatch Alarm for prompt injection via Terraform (>10/min for 5 min, SNS notify) | | | P1 | Backlog | 003-T062 | yes | | Infra integration |

#### Phase 4 — User Story 2: Frontend Engineer Implements Secure UI Component (Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 004-T076 | Write unit test for auth context in `frontend/src/lib/authContext.test.tsx` (useAuth hook) | G-SEC-FRONTEND-AUTH-TESTS | | P2 | Backlog | 004-T030 | yes | | RED phase |
| 004-T077 | Write unit test for token refresh in `frontend/src/lib/auth.test.ts` (401 redirect) | G-SEC-FRONTEND-AUTH-TESTS | | P2 | Backlog | 004-T030 | yes | | RED phase |
| 004-T078 | Implement auth context in `frontend/src/lib/authContext.tsx` (user state, login/logout functions) | G-SEC-FRONTEND-AUTH-IMPL | | P2 | Backlog | 004-T076 | no | | GREEN phase |
| 004-T079 | Implement auth API client in `frontend/src/lib/auth.ts` (credentials=include, error handling) | G-SEC-FRONTEND-AUTH-IMPL | | P2 | Backlog | 004-T077 | no | | GREEN phase |
| 004-T080 | Implement token refresh interceptor in `frontend/src/lib/auth.ts` (retry once, redirect on failure) | G-SEC-FRONTEND-AUTH-IMPL | | P2 | Backlog | 004-T077 | no | | GREEN phase |
| 004-T081 | Write unit test for role hook in `frontend/src/hooks/useRole.test.ts` (hasRole returns boolean) | G-SEC-FRONTEND-ROLE-TESTS | | P2 | Backlog | 004-T078 | yes | | RED phase |
| 004-T082 | Write integration test for protected route in `frontend/tests/integration/roleUI.test.ts` (redirect on unauth) | G-SEC-FRONTEND-ROLE-TESTS | | P2 | Backlog | 004-T078 | yes | | RED phase |
| 004-T083 | Implement role hook in `frontend/src/hooks/useRole.ts` (useRole, hasRole, requireRole) | G-SEC-FRONTEND-ROLE-IMPL | | P2 | Backlog | 004-T081 | no | | GREEN phase |
| 004-T084 | Implement protected route guard in `frontend/src/components/ProtectedRoute.tsx` (requireRole prop) | G-SEC-FRONTEND-ROLE-IMPL | | P2 | Backlog | 004-T082 | no | | GREEN phase |
| 004-T085 | Apply protected routes in `frontend/src/App.tsx` (wrap /dashboard, /trips, /admin/*) | | | P2 | Backlog | 004-T084 | no | | GREEN phase |
| 004-T086 | Implement role-based UI toggling (show Edit button only if hasRole admin) | | | P2 | Backlog | 004-T083 | yes | | GREEN phase |
| 004-T087 | Write unit test for secure rendering in `frontend/src/lib/secureRender.test.ts` (script stripped) | G-SEC-FRONTEND-RENDER-TESTS | | P2 | Backlog | 004-T084 | yes | | RED phase |
| 004-T088 | Write integration test for AI content display in `frontend/tests/integration/authFlow.test.ts` (no script exec) | G-SEC-FRONTEND-RENDER-TESTS | | P2 | Backlog | 004-T084 | yes | | RED phase |
| 004-T089 | Implement secure rendering utility in `frontend/src/lib/secureRender.ts` (sanitizeAIContent, SafeAIContent) | G-SEC-FRONTEND-RENDER-IMPL | | P2 | Backlog | 004-T087 | no | | GREEN phase |
| 004-T090 | Implement secure content hook in `frontend/src/hooks/useSecureContent.ts` (wraps sanitizeAIContent) | G-SEC-FRONTEND-RENDER-IMPL | | P2 | Backlog | 004-T087 | no | | GREEN phase |
| 004-T091 | Apply secure rendering to AI content in trip/itinerary components | | | P2 | Backlog | 004-T089 | no | | GREEN phase |
| 004-T092 | Write unit test for error handling in `frontend/src/lib/auth.test.ts` (401 message, 500 correlation ID) | G-SEC-FRONTEND-ERROR-TESTS | | P2 | Backlog | 004-T089 | yes | | RED phase |
| 004-T093 | Write integration test for session expiration in `frontend/tests/integration/authFlow.test.ts` (expired message) | G-SEC-FRONTEND-ERROR-TESTS | | P2 | Backlog | 004-T089 | yes | | RED phase |
| 004-T094 | Implement error formatter in `frontend/src/lib/errors.ts` (formatSecurityError, no sensitive data) | G-SEC-FRONTEND-ERROR-IMPL | | P2 | Backlog | 004-T092 | no | | GREEN phase |
| 004-T095 | Implement session expiration handler in `frontend/src/lib/auth.ts` (toast, redirect after 5s) | G-SEC-FRONTEND-ERROR-IMPL | | P2 | Backlog | 004-T093 | no | | GREEN phase |
| 004-T096 | Apply error handling to forms (generic 401 message, no enumeration) | | | P2 | Backlog | 004-T094 | no | | GREEN phase |
| 004-T097 | Implement password change form in `frontend/src/pages/Settings.tsx` ("log out all devices" checkbox) | G-SEC-FRONTEND-PASSWORD | | P2 | Backlog | 004-T034 | yes | | GREEN phase |
| 004-T098 | Write integration test for password change in `frontend/tests/integration/authFlow.test.ts` (checkbox behavior) | G-SEC-FRONTEND-PASSWORD | | P2 | Backlog | 004-T097 | no | | RED phase |

#### Phase 5 — User Story 3: QA Engineer Validates Security Controls (Priority: P3)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 004-T099 | Create OWASP Top 10 test vectors in `backend/tests/security/owasp_vectors_test.go` (50+ SQL injection variants) | G-SEC-QA-OWASP | | P3 | Backlog | 004-T044 | yes | | |
| 004-T100 | Create XSS test vectors in `backend/tests/security/owasp_vectors_test.go` (50+ XSS variants) | G-SEC-QA-OWASP | | P3 | Backlog | 004-T045 | yes | | |
| 004-T101 | Create path traversal test vectors in `backend/tests/security/owasp_vectors_test.go` (../, symlinks) | G-SEC-QA-OWASP | | P3 | Backlog | 004-T046 | yes | | |
| 004-T102 | Create OWASP LLM Top 10 test vectors in `backend/tests/security/llm_prompts_test.go` (20+ prompt injection patterns) | G-SEC-QA-LLM | | P3 | Backlog | 004-T052 | yes | | |
| 004-T103 | Create off-topic prompt tests in `backend/tests/security/llm_prompts_test.go` (finance, medical, code, SQL) | G-SEC-QA-LLM | | P3 | Backlog | 004-T052 | yes | | |
| 004-T104 | Create role-based authorization test matrix in `backend/tests/integration/rbac_test.go` (11 operation × role combinations) | | | P3 | Backlog | 004-T038 | no | | |
| 004-T105 | Test subscription plan limits in `backend/tests/integration/rbac_test.go` (2nd partner invite → 403) | | | P3 | Backlog | 004-T104 | yes | | |
| 004-T106 | Create logging validation test suite in `backend/tests/integration/observability_test.go` (all event types) | | | P3 | Backlog | 004-T065 | no | | |
| 004-T107 | Test CloudWatch metrics emission in `backend/tests/integration/observability_test.go` (metric increments) | | | P3 | Backlog | 004-T066 | yes | | |
| 004-T108 | Create E2E authentication test in `e2e/tests/auth.spec.ts` (register → login → protected → logout) | G-SEC-QA-E2E | | P3 | Backlog | 004-T078 | yes | | |
| 004-T109 | Create E2E authorization test in `e2e/tests/authorization.spec.ts` (admin vs partner trip creation) | G-SEC-QA-E2E | | P3 | Backlog | 004-T078 | yes | | |
| 004-T110 | Create E2E security test in `e2e/tests/security.spec.ts` (XSS rejection, prompt injection, safe rendering) | G-SEC-QA-E2E | | P3 | Backlog | 004-T078 | yes | | |
| 004-T111 | Validate quickstart Scenario 1 in `backend/tests/integration/quickstart_test.go` (register, login, 401 unauth) | G-SEC-QA-QUICKSTART | | P3 | Backlog | 004-T030 | yes | | |
| 004-T112 | Validate quickstart Scenario 2 (token refresh) in automated test | G-SEC-QA-QUICKSTART | | P3 | Backlog | 004-T032 | yes | | |
| 004-T113 | Validate quickstart Scenario 3 (RBAC partner → 403) in automated test | G-SEC-QA-QUICKSTART | | P3 | Backlog | 004-T038 | yes | | |
| 004-T114 | Validate quickstart Scenario 4 (optimistic locking 409) in automated test | G-SEC-QA-QUICKSTART | | P3 | Backlog | 004-T059 | yes | | |
| 004-T115 | Validate quickstart Scenario 5 (prompt injection rejection) in automated test | G-SEC-QA-QUICKSTART | | P3 | Backlog | 004-T052 | yes | | |
| 004-T116 | Validate quickstart Scenario 6 (password change session invalidation) in automated test | G-SEC-QA-QUICKSTART | | P3 | Backlog | 004-T034 | yes | | |
| 004-T117 | Validate quickstart Scenario 7 (JWT multi-key rotation) in automated test | G-SEC-QA-QUICKSTART | | P3 | Backlog | 004-T021 | yes | | |

#### Phase 6 — Polish & Cross-Cutting Concerns (Priority: P2-P3)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 004-T118 | Update backend README: security architecture, JWT key rotation, CloudWatch alarm response | G-SEC-DOCS | | P2 | Backlog | 004-T067 | yes | | |
| 004-T119 | Update frontend README: auth context usage, secure rendering guidelines, role-based UI patterns | G-SEC-DOCS | | P2 | Backlog | 004-T089 | yes | | |
| 004-T120 | Create security runbook in `docs/security-operations.md` (JWT rotation, alarm response, GDPR deletion) | G-SEC-DOCS | | P2 | Backlog | 004-T067 | yes | | |
| 004-T121 | Refactor authentication handlers (extract common validation, DRY principle) | G-SEC-REFACTOR | | P2 | Backlog | 004-T030 | yes | | |
| 004-T122 | Refactor authorization middleware (permission rules → config map) | G-SEC-REFACTOR | | P2 | Backlog | 004-T038 | yes | | |
| 004-T123 | Frontend code review (no dangerouslySetInnerHTML with AI content, ESLint rule) | G-SEC-HARDENING | | P2 | Backlog | 004-T089 | yes | | |
| 004-T124 | Add security headers middleware in `backend/internal/observability/middleware.go` (HSTS, CSP, etc.) | G-SEC-HARDENING | | P2 | Backlog | 004-T024 | yes | | |
| 004-T125 | Configure CORS in `backend/cmd/server/main.go` (staging origin, credentials=true) | G-SEC-HARDENING | | P2 | Backlog | 004-T024 | yes | | |
| 004-T126 | Add gitleaks pre-commit hook in `.git/hooks/pre-commit` (exit on secrets detected) | | | P2 | Backlog | 002-T036 | yes | | |
| 004-T127 | Run full quickstart.md validation manually (all 9 scenarios with curl) | G-SEC-VALIDATION | | P3 | Backlog | 004-T111 | yes | | Manual QA |
| 004-T128 | Run OWASP ZAP baseline scan on staging (no high/medium findings) | G-SEC-VALIDATION | | P3 | Backlog | 003-T052 | yes | | Manual QA |
| 004-T129 | Run `npm audit --audit-level=high` on frontend (zero high-severity CVEs) | G-SEC-VALIDATION | | P3 | Backlog | 002-T035 | yes | | Manual QA |
| 004-T130 | Run `govulncheck ./...` on backend (zero critical/high CVEs) | G-SEC-VALIDATION | | P3 | Backlog | 002-T034 | yes | | Manual QA |
| 004-T131 | Verify CloudWatch Logs retention (query AWS CLI, confirm 30 days) | G-SEC-VALIDATION | | P3 | Backlog | 004-T073 | yes | | Manual QA |
| 004-T132 | Verify CloudWatch Alarms configured (auth-failure-rate, prompt-injection-rate) | G-SEC-VALIDATION | | P3 | Backlog | 004-T074 | yes | | Manual QA |

---

## Feature Phase

_No feature specs defined yet. Add specs 004+ here as they are created._

---

## Critical Path

The minimum sequential chain to reach a fully functional, security-hardened, demo-able MVP:

```
001-T001   Create monorepo root structure
  ├─ 003-T001   Create infra/ directory structure (parallel foundational work)
  │    └─ 003-T009   Bootstrap Terraform remote state
  │         └─ 003-T010   Configure Terraform backend
  │              └─ 003-T014–034   Build Terraform modules (VPC, ECS, RDS, ALB, S3+CF, IAM)
  │                   └─ 003-T035–044   Wire modules in root + populate tfvars
  │                        └─ 003-T045–047   Initialize secrets + wire into ECS
  │                             └─ 003-T048–054   Create CI/CD workflows (Terraform, backend, frontend)
  │                                  
  └─ 001-T009   Create 8 DB migration files
       └─ 001-T014   Implement User repository
            └─ 001-T016   Implement auth service
                 └─ 001-T021   Implement auth HTTP handlers
                      └─ 001-T023   Scaffold Chi router (all routes + middleware)
                           ├─ 002-T012   Register /healthz + wire RequestID/Logger middleware
                           └─ 001-T033   Implement Trip + Day + Activity repository
                                └─ 001-T035   Implement Trip service
                                     └─ 001-T037   Implement Itinerary service (Claude streaming)
                                          ├─ 002-T029   Wire PromptValidator into generation handler
                                          └─ 001-T038   Implement Trip HTTP handlers
                                               └─ 001-T040   Register trip/conversation routes
                                                    └─ 001-T046   Implement TanStack Query hooks
                                                         └─ 001-T047   Implement Generate page
                                                              └─ 001-T050   E2E: full itinerary generation
                                                                   └─ 003-T052   Deploy backend to ECS Fargate
                                                                        └─ 003-T053   Deploy frontend to S3+CloudFront
```

**Secondary critical paths** (branch from 001-T021):
- **Infrastructure provisioning**: `003-T001 → 003-T009 → 003-T010 → ... → 003-T054` (enables deployment)
- **Privacy / GDPR**: `001-T021 → 002-T032 → 002-T033`
- **Collaboration**: `001-T050 → 001-T063 → 001-T065 → 001-T067 → 001-T069 → 001-T075`
- **Accessibility CI gate**: `002-T002 → 002-T022 → 002-T023 → 002-T024`
- **Security CI gates**: `001-T076 → 002-T034 → 002-T036 → 003-T067` and `001-T077 → 002-T035`
- **Coverage gates**: `001-T078 → 002-T044` and `001-T079 → 002-T043`
- **Infrastructure observability**: `003-T054 → 003-T055–064` (cost monitoring and CloudWatch dashboards)

---

## Specs Without Tasks Yet

_None — all discovered spec folders have a `tasks.md`._

---

## Archived / Removed

_No tasks archived on this run (first run; all tasks are ADD operations)._

| ID | Task | Reason | Issue |
|----|------|--------|-------|
| — | — | — | — |

---

## Reconciliation Report

**Run date**: 2026-07-03

### Specs Discovered

| Spec | Folder | Has tasks.md | Status |
|------|--------|-------------|--------|
| 001 | `specs/001-product-vision-scope/` | ✅ yes | Active |
| 002 | `specs/002-nfr-system-constraints/` | ✅ yes | Active |
| 003 | `specs/003-cloud-env-strategy/` | ✅ yes | Active |
| 004 | `specs/004-security-auth-model/` | ✅ yes | **NEW** |

### Change Counts

| Operation | Count |
|-----------|-------|
| **ADD** | 132 (spec 004: all 132 tasks) |
| **UPDATE** | 0 |
| **UNCHANGED** | 199 (001: 81 tasks, 002: 46 tasks, 003: 72 tasks) |
| **REMOVE** | 0 |
| **ARCHIVED** | 0 |
| **Structural** | Added 29 new groups for spec 004 tasks |

**Total tasks in roadmap**: 331 (was 199, added 132)  
**Grouped tasks**: 234 (70.7% of total, forming 71 work items)  
**Standalone tasks**: 97 (29.3% of total)

### New Grouping Summary (Spec 004)

| Group ID | Tasks | Description |
|----------|-------|-------------|
| G-SEC-FRONTEND-STRUCTURE | 1 | Frontend security directory structure (004-T002) |
| G-SEC-BACKEND-DEPS | 1 | Backend security dependencies (004-T003) |
| G-SEC-FRONTEND-DEPS | 1 | Frontend security dependencies (004-T004) |
| G-SEC-TEST-STRUCTURE | 1 | Security test directories (004-T005) |
| G-SEC-DB-MIGRATIONS | 6 | Security-related database migrations (004-T006–011) |
| G-SEC-CORE-UTILITIES | 3 | Core security utilities: password, correlation ID, logger (004-T012–014) |
| G-SEC-JWT-TESTS | 3 | JWT generation/validation/rotation tests (004-T015–017) |
| G-SEC-JWT-IMPL | 4 | JWT implementation (004-T018–021) |
| G-SEC-AUTH-MIDDLEWARE-TESTS | 2 | Auth middleware integration tests (004-T022–023) |
| G-SEC-AUTH-HANDLERS-TESTS | 4 | Auth handlers integration tests (004-T026–029) |
| G-SEC-AUTH-HANDLERS-IMPL | 6 | Auth handlers implementation (004-T030–035) |
| G-SEC-RBAC-TESTS | 2 | RBAC permission tests (004-T036–037) |
| G-SEC-RBAC-IMPL | 2 | RBAC implementation (004-T038–039) |
| G-SEC-VALIDATION-TESTS | 3 | Input validation security tests (004-T041–043) |
| G-SEC-VALIDATION-IMPL | 4 | Input validation implementation (004-T044–047) |
| G-SEC-PROMPT-TESTS | 2 | Prompt injection tests (004-T050–051) |
| G-SEC-PROMPT-IMPL | 2 | Prompt validation implementation (004-T052–053) |
| G-SEC-SANITIZATION-TESTS | 2 | Output sanitization tests (004-T054–055) |
| G-SEC-SANITIZATION-IMPL | 1 | Output sanitization implementation (004-T056) |
| G-SEC-CONCURRENCY-TESTS | 1 | Optimistic locking tests (004-T058) |
| G-SEC-CONCURRENCY-IMPL | 3 | Optimistic locking implementation (004-T059–061) |
| G-SEC-LOGGING-TESTS | 2 | Security logging/metrics tests (004-T063–064) |
| G-SEC-LOGGING-IMPL | 2 | CloudWatch logging/metrics implementation (004-T065–066) |
| G-SEC-SECRETS-TESTS | 1 | Secrets Manager tests (004-T069) |
| G-SEC-SECRETS-IMPL | 1 | AWS Secrets Manager client (004-T070) |
| G-SEC-FRONTEND-AUTH-TESTS | 2 | Frontend auth context tests (004-T076–077) |
| G-SEC-FRONTEND-AUTH-IMPL | 3 | Frontend auth context implementation (004-T078–080) |
| G-SEC-FRONTEND-ROLE-TESTS | 2 | Frontend role hook tests (004-T081–082) |
| G-SEC-FRONTEND-ROLE-IMPL | 2 | Frontend role-based UI implementation (004-T083–084) |
| G-SEC-FRONTEND-RENDER-TESTS | 2 | Frontend secure rendering tests (004-T087–088) |
| G-SEC-FRONTEND-RENDER-IMPL | 2 | Frontend secure rendering implementation (004-T089–090) |
| G-SEC-FRONTEND-ERROR-TESTS | 2 | Frontend error handling tests (004-T092–093) |
| G-SEC-FRONTEND-ERROR-IMPL | 2 | Frontend error handling implementation (004-T094–095) |
| G-SEC-FRONTEND-PASSWORD | 2 | Frontend password change form (004-T097–098) |
| G-SEC-QA-OWASP | 3 | OWASP Top 10 test vectors (004-T099–101) |
| G-SEC-QA-LLM | 2 | OWASP LLM Top 10 test vectors (004-T102–103) |
| G-SEC-QA-E2E | 3 | Security E2E tests (004-T108–110) |
| G-SEC-QA-QUICKSTART | 7 | Automated quickstart scenario validation (004-T111–117) |
| G-SEC-DOCS | 3 | Security documentation (004-T118–120) |
| G-SEC-REFACTOR | 2 | Code cleanup and refactoring (004-T121–122) |
| G-SEC-HARDENING | 3 | Additional security hardening (004-T123–125) |
| G-SEC-VALIDATION | 6 | Final validation tasks (004-T127–132) |

### Cross-Spec Dependencies (Spec 004)

| Task | Depends on | Type | Description |
|------|------------|------|-------------|
| 004-T070 | 003-T045 | Requires | AWS Secrets Manager client needs secrets initialized by 003 |
| 004-T073 | 003-T059 | Extends | Security CloudWatch log group with 30-day retention policy |
| 004-T074 | 003-T062 | Extends | Auth failure CloudWatch alarm (integrates with 003 observability) |
| 004-T075 | 003-T062 | Extends | Prompt injection CloudWatch alarm (integrates with 003 observability) |
| 004-T126 | 002-T036 | Extends | Adds gitleaks pre-commit hook (complements 002 CI gate) |
| 004-T129 | 002-T035 | Validates | Confirms frontend npm audit passes (002 CI gate) |
| 004-T130 | 002-T034 | Validates | Confirms backend govulncheck passes (002 CI gate) |
| 004-T057 | 001-T037 | Integration | Apply sanitization in AI integration layer (app code location TBD) |

### Human Attention Required

| Item | Detail |
|------|--------|
| **New spec review** | Spec 004 (Security & Authentication/Authorization Model) added with 132 tasks. Review grouping and priorities. |
| **Priority review** | All spec 004 priorities default to P1 (Phase 3 US1), P2 (Phase 4 US2, Phase 6 Polish), P3 (Phase 5 US3). Review and override if security work should have different sprint priorities. |
| **Status review** | All 132 new tasks default to `Backlog`. Mark tasks `Ready` or `In Progress` as security implementation begins. |
| **Issue column** | All 132 new spec 004 `Issue` fields are empty. **90 grouped tasks will form 29 issues** (with checklists); **42 standalone tasks will form 42 issues**. Total NEW issues: **71 issues** when `/sync-issues` runs. |
| **Sprint column** | All spec 004 `Sprint` fields are empty. Populate during sprint planning. Security work is foundational — consider prioritizing Phase 2–3 (foundational + backend US1) early. |
| **TDD mandate** | Spec 004 follows strict RED-GREEN-REFACTOR TDD workflow. All test tasks (RED phase) MUST be completed BEFORE implementation tasks (GREEN phase). This is non-negotiable per constitution. |
| **004 cross-spec deps** | 8 tasks integrate with specs 001, 002, 003. Coordinate: 004-T070 needs 003-T045 (Secrets Manager); 004-T073–075 extend 003 CloudWatch infrastructure; 004-T057 integrates with 001 AI layer. |
| **Security blocking** | Spec 004 Phase 2 (foundational tasks 004-T006–014) BLOCKS all security user stories. Must complete migrations and core utilities before auth/validation implementation can begin. |

### Critical Path Changes

- **New parallel path**: Security foundation (004-T001 → 004-T014) runs in parallel with application development (001), NFR validation (002), and infrastructure provisioning (003).
- **Security integration points**:
  - 004-T070 BLOCKS 004-T071–072 (JWT key rotation): requires 003-T045 (Secrets Manager setup)
  - 004-T073–075 BLOCKS final deployment: CloudWatch alarms required for production readiness
  - 004-T057 integrates with 001-T037 (AI itinerary service): sanitization must be applied before DB/client return
- **TDD critical path**: Within spec 004, test tasks (RED phase) strictly BLOCK implementation tasks (GREEN phase). Example: 004-T015–017 (JWT tests) BLOCK 004-T018–021 (JWT implementation).
- **Auth dependency cascade**: 004-T018–021 (JWT core) BLOCKS 004-T022–025 (auth middleware) BLOCKS 004-T026–035 (auth handlers) BLOCKS all frontend auth work (004-T076–098).

### Next Steps

1. **Review spec 004 grouping**: The 29 groups bundle 90 tasks into cohesive work items following TDD workflow. If any grouping doesn't align with team velocity, clear the `Group` value to split into standalone items.
2. **Sprint planning**: Populate `Sprint` column for spec 004. Recommend: Phase 2 (foundational) in first security sprint, Phase 3 (backend US1) in sprints 2-3, Phase 4 (frontend US2) in sprint 4, Phase 5-6 (QA + Polish) in final sprint.
3. **Coordinate cross-spec work**: 004-T070 depends on 003-T045; 004-T073–075 extend 003-T059–062; 004-T057 integrates with 001-T037. Ensure teams coordinate on these integration points.
4. **Run `/sync-issues`**: This will create **71 new GitHub issues** for spec 004 (29 grouped + 42 standalone). Grouped issues will have checklists with member tasks. Use labels: `security`, `tdd`, `P1`/`P2`/`P3`.
5. **Begin security implementation**: Follow the critical path: `004-T001 → 004-T006–014 (foundational) → 004-T015–075 (backend US1) → 004-T076–098 (frontend US2) → 004-T099–117 (QA US3) → 004-T118–132 (polish)`. Remember TDD mandate: RED phase (tests) MUST come before GREEN phase (implementation).
6. **Re-run `/build-roadmap`**: When spec 004 `tasks.md` is modified or a new spec is added — the command preserves all `Group`, `Sprint`, `Priority`, `Status`, and `Issue` fields.

---

**Total project task count**: **331 tasks** across 4 foundation specs  
**Grouped work items**: **71 issues** (combining 234 tasks)  
**Standalone work items**: **97 issues**  
**Total GitHub issues when synced**: **168 issues**

---

_End of reconciliation report. All human-owned fields preserved. Ready for sprint planning and `/sync-issues` execution._
