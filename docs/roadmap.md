# Project Roadmap

> Generated and reconciled by /build-roadmap. Source of truth for task existence,
> titles, and dependencies is `specs/*/tasks.md`. Priority, Status, Phase, Issue, and
> Notes are human-owned and preserved across runs. Do not hand-edit the stable IDs.

**Last reconciled**: 2026-07-06 (updated with specs 006, 007, and 008)  
**Sprint planning**: 2026-07-06 (MVP: 10 sprints, 390 tasks assigned to Sprints 1-10; Post-MVP: 355 tasks unassigned)

## Legend

- **Group**: shared value (e.g. `G-SETUP-1`) = tasks handled by ONE issue (checklist inside); empty = standalone (1 task = 1 issue). Human-owned — set when grouping is desired.
- **Sprint**: sprint number (1-10 for MVP) or milestone label. Human-owned — assigned during sprint planning. See **Sprint Plan** section below for details.
- **Priority**: P1 (critical) | P2 | P3 | TBD
- **Status**: Backlog | Ready | In Progress | In Review | Done
- **Parallel**: yes = task carries `[P]` flag in source (can run concurrently with peers)
- **Issue**: link to the tracker issue once created (empty = not yet created). Grouped tasks share the same URL.

**Note**: Sprint assignments shown in phase headers (e.g., "→ Sprint 1") and key task Sprint columns. See [Sprint Plan](#sprint-plan) section for full sprint breakdown with goals, deliverables, and risks.

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

#### Phase 2 — Foundational (Backend Data Layer) → **Sprint 6**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 001-T009 | Create 8 database migration SQL files (users, plans, subscriptions, trips, destinations, days+activities, collaborators, suggestions+conversation) | | 6 | P1 | Backlog | 001-T001 | no | | |
| 001-T010 | Implement config struct with env var loading and fail-fast validation | G-BACKEND-CONFIG | 6 | P1 | Backlog | 001-T001 | no | | |
| 001-T011 | Implement pgxpool connection initialization with context-aware open/close | G-BACKEND-CONFIG | 6 | P1 | Backlog | 001-T001 | yes | | |
| 001-T012 | Define `PaymentProvider` interface with CreateSubscription, CancelSubscription, GetSubscription | G-BACKEND-CONFIG | 6 | P1 | Backlog | 001-T001 | yes | | |
| 001-T013 | Implement `StubProvider` satisfying `PaymentProvider` (always succeeds, logs [STUB]) | G-BACKEND-CONFIG | 6 | P1 | Backlog | 001-T012 | no | | |
| 001-T014 | Implement User repository (Create, FindByEmail, FindByID, UpdateSubscription) | G-BACKEND-AUTH-REPOS | 6 | P1 | Backlog | 001-T009 | yes | | |
| 001-T015 | Implement Plan and Subscription repositories (FindPlanByName, CreateSubscription, FindSubscriptionByUser) | G-BACKEND-AUTH-REPOS | 6 | P1 | Backlog | 001-T009 | yes | | |

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

#### Phase 2 — Foundational (Frontend Shell & Auth) → **Sprint 7**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 001-T024 | Create CSS custom property design tokens (colors, spacing, typography, border-radius) | | 7 | P1 | Backlog | 001-T001 | yes | | |
| 001-T025 | Create typed API client base (`apiFetch` wrapper with credentials: include, JSON parsing) | G-FRONTEND-INFRA | 7 | P1 | Backlog | 001-T001 | yes | | |
| 001-T026 | Create auth Zustand store (user, setUser, clearUser; persist to sessionStorage) | G-FRONTEND-INFRA | 7 | P1 | Backlog | 001-T001 | yes | | |
| 001-T027 | Create `ProtectedRoute` (redirect to /login) and `GuestRoute` (redirect to /dashboard) | G-FRONTEND-INFRA | 7 | P1 | Backlog | 001-T001 | yes | | |
| 001-T028 | Create primitive Button, Input, Label, Badge components using design tokens | | 7 | P1 | Backlog | 001-T024 | yes | | |
| 001-T029 | Implement Register page (email + password form, calls POST /auth/register, redirects to checkout) | G-FRONTEND-AUTH-PAGES | 7 | P1 | Backlog | 001-T024, 001-T025 | no | | |
| 001-T030 | Implement stub Checkout page (plan summary, POST /subscription/checkout + /confirm, redirect dashboard) | G-FRONTEND-AUTH-PAGES | 7 | P1 | Backlog | 001-T028, 001-T029 | no | | |
| 001-T031 | Implement Login page (calls POST /auth/login, sets user in store, redirects to dashboard) | G-FRONTEND-AUTH-PAGES | 7 | P1 | Backlog | 001-T028, 001-T030 | no | | |
| 001-T032 | Wire React Router v7 with all routes (/, /login, /register, /subscribe, /dashboard, /trips/:id, /generate) | | 7 | P1 | Backlog | 001-T029, 001-T030, 001-T031 | no | | |

#### Phase 3 — User Story 1: First-Time Traveler Plans a Trip (Priority: P1) 🎯 MVP → **Sprint 8**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 001-T033 | Implement Trip + Destination + Day + Activity repository (CreateTrip, UpsertDay, UpsertActivity) | G-US1-REPOS | 8 | P1 | Backlog | 001-T009 | yes | | |
| 001-T034 | Implement ConversationSession + ConversationMessage repository (CreateSession, AppendMessage, ListMessages) | G-US1-REPOS | 8 | P1 | Backlog | 001-T009 | yes | | |
| 001-T035 | Implement Trip service (Create, List, Get, Update, Delete; admin-only writes) | | 8 | P1 | Backlog | 001-T033 | no | | |
| 001-T036 | Implement Conversation service (SendMessage, GetHistory; detects itinerary_ready) | | 8 | P1 | Backlog | 001-T034 | no | | |
| 001-T037 | Implement Itinerary service (build Claude prompt, streaming, parse tool-use response, persist to DB) | | 8 | P1 | Backlog | 001-T035 | no | | |
| 001-T038 | Implement Trip HTTP handlers (GET/POST/PUT/DELETE /trips; RequireRole admin on writes) | G-US1-HANDLERS | 8 | P1 | Backlog | 001-T035 | no | | |
| 001-T039 | Implement Conversation HTTP handlers (POST/GET /trips/:id/conversation; SSE streaming) | G-US1-HANDLERS | 8 | P1 | Backlog | 001-T036 | no | | |
| 001-T040 | Register trip and conversation routes in Chi router | | 8 | P1 | Backlog | 001-T038, 001-T039 | no | | |
| 001-T041 | Create `TripCard` composite (destination names, duration, status badge; loading/empty states) | G-US1-COMPONENTS | 8 | P1 | Backlog | 001-T028 | yes | | |
| 001-T042 | Create `ActivityItem` composite (Lucide type icon, title, description, AI-generated indicator) | G-US1-COMPONENTS | 8 | P1 | Backlog | 001-T028 | yes | | |
| 001-T043 | Create `DaySection` composite (day number, label, ordered ActivityItem list; empty state) | G-US1-COMPONENTS | 8 | P1 | Backlog | 001-T028 | yes | | |
| 001-T044 | Create `ConversationPanel` feature (message thread, text input, SSE stream rendering, loading indicator) | G-US1-COMPONENTS | 8 | P1 | Backlog | 001-T028 | yes | | |
| 001-T045 | Create `ItineraryView` feature (scrollable DaySection list; loading skeleton, error, empty states) | G-US1-COMPONENTS | 8 | P1 | Backlog | 001-T028 | yes | | |
| 001-T046 | Implement trips and conversation API service functions with TanStack Query hooks | | 8 | P1 | Backlog | 001-T025 | no | | |
| 001-T047 | Implement Generate page (new trip form → ConversationPanel → ItineraryView on itinerary_ready) | G-US1-PAGES | 8 | P1 | Backlog | 001-T044, 001-T045, 001-T046 | no | | |
| 001-T048 | Implement Trip detail page (ItineraryView, action bar: edit title, delete trip) | G-US1-PAGES | 8 | P1 | Backlog | 001-T045, 001-T046 | no | | |
| 001-T049 | Implement Dashboard page (TripCard grid, "New Trip" CTA, empty state illustration) | G-US1-PAGES | 8 | P1 | Backlog | 001-T041, 001-T046 | no | | |
| 001-T050 | Add Playwright E2E spec: register → subscribe → generate itinerary via conversation → verify Day 1 | | 10 | P1 | Backlog | 001-T047, 001-T049, 001-T040 | no | | |

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

#### Phase 2 — Foundational: Observability Core (NFR-OBS-001–003) → **Sprint 4**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 002-T006 | Implement `RequestID` Chi middleware (generate UUID v4 if absent; propagate; set X-Request-ID on response) | G-OBS-REQUESTID | 4 | P1 | Backlog | - | no | | |
| 002-T007 | Write unit tests for `RequestID` middleware (generates UUID, propagates existing, sets response header) | G-OBS-REQUESTID | 4 | P1 | Backlog | 002-T006 | no | | |
| 002-T008 | Implement `Logger` Chi middleware using `log/slog` JSON handler (emit StructuredLogEntry per request) | G-OBS-LOGGER | 4 | P1 | Backlog | - | yes | | |
| 002-T009 | Write unit tests for `Logger` middleware (all fields present, duration_ms ≥ 0, user_id absent on unauth) | G-OBS-LOGGER | 4 | P1 | Backlog | 002-T008 | no | | |
| 002-T010 | Implement `GET /healthz` handler returning HealthCheckResponse JSON (status, version, uptime_seconds) | G-OBS-HEALTHZ | 4 | P1 | Backlog | - | yes | | |
| 002-T011 | Write unit tests for `/healthz` handler (200 OK, schema valid, status is "ok", uptime ≥ 0) | G-OBS-HEALTHZ | 4 | P1 | Backlog | 002-T010 | no | | |
| 002-T012 | Register `/healthz` and wire `RequestID` → `Logger` middleware chain globally in Chi router | | 4 | P1 | Backlog | 002-T006, 002-T008, 002-T010, 001-T023 | no | | |

#### Phase 3 — User Story 1: Engineering Team Verifies Performance Under Load (Priority: P1) 🎯 → **Sprint 9**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 002-T013 | Create k6 baseline load test (ramp to 500 VUs, sustain 10 min; assert p95 ≤ 500 ms and error rate < 1%) | G-PERF-K6-TESTS | 9 | P1 | Backlog | 002-T012 | yes | | |
| 002-T014 | Create k6 API latency scenario (constant 100 RPS; assert p95 ≤ 500 ms per non-AI endpoint) | G-PERF-K6-TESTS | 9 | P1 | Backlog | 002-T012 | yes | | |
| 002-T015 | Create `load-test.yml` GitHub Actions workflow (manual `workflow_dispatch`; runs k6 against staging URL) | | 9 | P1 | Backlog | 002-T013 | no | | |
| 002-T016 | Write integration test asserting `/healthz` responds in ≤ 100 ms for 100 sequential calls | | 9 | P1 | Backlog | 002-T010 | no | | |

#### Phase 4 — User Story 2: Accessibility Reviewer Confirms WCAG 2.1 AA (Priority: P1) → **Sprint 9**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 002-T017 | Implement `PrivacyPolicyPage` static component (WCAG-compliant heading, section structure, landmarks) | G-A11Y-PRIVACY-PAGE | 9 | P1 | Backlog | 002-T002 | yes | | |
| 002-T018 | Write Vitest unit test for `PrivacyPolicyPage` (renders heading, ≥ 3 sections, no dangerouslySetInnerHTML) | G-A11Y-PRIVACY-PAGE | 9 | P1 | Backlog | 002-T017 | yes | | |
| 002-T019 | Implement `PrivacyPolicyLink` atom component (accessible `<a>` linking to /privacy-policy) | G-A11Y-PRIVACY-LINK | 9 | P1 | Backlog | 002-T002 | yes | | |
| 002-T020 | Write Vitest unit test for `PrivacyPolicyLink` (correct href, accessible text present) | G-A11Y-PRIVACY-LINK | 9 | P1 | Backlog | 002-T019 | yes | | |
| 002-T021 | Add `/privacy-policy` route to React Router; add `PrivacyPolicyLink` to registration form footer | | 9 | P1 | Backlog | 002-T017, 002-T019 | no | | |
| 002-T022 | Create accessibility E2E helper `checkPageA11y(page)` wrapping `@axe-core/playwright` | | 9 | P1 | Backlog | 002-T002 | yes | | |
| 002-T023 | Add `checkPageA11y(page)` call to every existing Playwright E2E spec; tag with `@accessibility` | | 9 | P1 | Backlog | 002-T022 | no | | |
| 002-T024 | Create `accessibility.yml` GitHub Actions workflow (axe-core Playwright run + lhci autorun; PR gate) | | 9 | P1 | Backlog | 002-T022, 002-T004 | no | | |

#### Phase 5 — User Story 3: Security Reviewer Confirms Security Posture (Priority: P1) → **Sprint 9**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 002-T025 | Implement `PromptValidator` (load rules from YAML; Validate() with substring + regex modes; panic on invalid regex) | G-SEC-PROMPT-VALIDATOR | 9 | P1 | Backlog | 002-T003 | yes | | |
| 002-T026 | Write unit tests for `PromptValidator` (clean prompt passes; 5 seed rules trigger match; disabled rule skipped) | G-SEC-PROMPT-VALIDATOR | 9 | P1 | Backlog | 002-T025 | no | | |
| 002-T027 | Implement `OutputSanitizer` using `bluemonday.UGCPolicy()` (strip HTML/script before storage) | G-SEC-OUTPUT-SANITIZER | 9 | P1 | Backlog | 002-T001 | yes | | |
| 002-T028 | Write unit tests for `OutputSanitizer` (script stripped, event handlers stripped, plain text preserved) | G-SEC-OUTPUT-SANITIZER | 9 | P1 | Backlog | 002-T027 | no | | |
| 002-T029 | Wire `PromptValidator` into itinerary generation handler (return 400 PromptRejectionResponse on match) | | 9 | P1 | Backlog | 002-T025, 001-T038 | no | | |
| 002-T030 | Wire `OutputSanitizer` into itinerary service before every DB INSERT of AI-generated content | | 9 | P1 | Backlog | 002-T027, 001-T037 | no | | |
| 002-T031 | Write integration test for prompt rejection (10 injection payloads → assert 400 + request_id matches header) | | 9 | P1 | Backlog | 002-T029 | yes | | |
| 002-T032 | Implement `DELETE /users/me` auth endpoint (cascade delete all PII; respond 204) | G-SEC-USER-DELETION | 9 | P1 | Backlog | 001-T021 | yes | | |
| 002-T033 | Write integration test for user deletion (register → create trip → DELETE → assert 204 + zero DB rows) | G-SEC-USER-DELETION | 9 | P1 | Backlog | 002-T032 | no | | |
| 002-T034 | Extend `backend-lint.yml` with `gosec -severity high` and `govulncheck` steps (gate on non-zero exit) | G-SEC-CI-GATES | 9 | P1 | Backlog | 001-T076 | yes | | |
| 002-T035 | Extend `frontend-lint.yml` with `npm audit --audit-level=high` step (gate on critical/high findings) | G-SEC-CI-GATES | 9 | P1 | Backlog | 001-T077 | yes | | |
| 002-T036 | Add `gitleaks detect` secret-scanning step to `backend-lint.yml` (every PR and push to main) | | 9 | P1 | Backlog | 001-T076, 002-T005 | no | | |

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

#### Phase 3 — User Story 1: Infrastructure Provisioning & Environment Setup (Priority: P1) 🎯 MVP → **Sprint 10**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 003-T014 | Create VPC module: provisions VPC with CIDR from tfvars, 2 public subnets (ALB), 2 private subnets (ECS, RDS), internet gateway, route tables | G-INFRA-VPC-MODULE | 10 | P1 | Backlog | 003-T009 | yes | | |
| 003-T015 | Create VPC security groups: `alb-sg`, `ecs-sg`, `rds-sg` | G-INFRA-VPC-MODULE | 10 | P1 | Backlog | 003-T009 | yes | | |
| 003-T016 | Create VPC NAT resource: conditional NAT instance (staging) or NAT Gateway (production) | G-INFRA-VPC-MODULE | 10 | P1 | Backlog | 003-T009 | yes | | |
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
| 003-T035 | Wire VPC module in root `main.tf`: call `modules/vpc` with staging/production-specific CIDR blocks and NAT type | | 10 | P1 | Backlog | 003-T014 | no | | |
| 003-T036 | Wire ALB module in root `main.tf`: call `modules/alb` with VPC outputs (public subnets, security group) | | 10 | P1 | Backlog | 003-T019 | no | | |
| 003-T037 | Wire RDS module in root `main.tf`: call `modules/rds` with VPC outputs and environment-specific instance class, Multi-AZ flag | | 10 | P1 | Backlog | 003-T022 | no | | |
| 003-T038 | Wire ECS module in root `main.tf`: call `modules/ecs` with VPC outputs, ALB target group ARN, environment-specific task sizing | | 10 | P1 | Backlog | 003-T025 | no | | |
| 003-T039 | Wire S3+CloudFront module in root `main.tf`: call `modules/s3-cloudfront` with environment-specific bucket name and CloudFront price class | | 10 | P1 | Backlog | 003-T029 | no | | |
| 003-T040 | Wire IAM module in root `main.tf`: call `modules/iam` with resource ARNs (ECR, Secrets Manager, S3) from other modules | | 10 | P1 | Backlog | 003-T032 | no | | |
| 003-T041 | Define root module variables: `environment`, `aws_region`, `vpc_cidr`, `availability_zones`, ECS task sizing, RDS config, NAT type, cost budget, tags | | 10 | P1 | Backlog | 003-T035 | no | | |
| 003-T042 | Define root module outputs: all outputs from modules (VPC, ALB, RDS, ECS, S3+CloudFront, IAM) | | 10 | P1 | Backlog | 003-T035 | no | | |
| 003-T043 | Populate `staging.tfvars` with staging-specific values: vpc_cidr, ECS task sizing, RDS instance class, NAT instance, cost budget | | 10 | P1 | Backlog | 003-T041 | yes | | |
| 003-T044 | Populate `production.tfvars` with production-specific values: vpc_cidr, ECS task sizing, RDS Multi-AZ, NAT Gateway, cost budget | | 10 | P1 | Backlog | 003-T041 | yes | | |

#### Phase 4 — User Story 2: Secrets & Configuration Management (Priority: P1) → **Sprint 10**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 003-T045 | Create secrets initialization script: uses AWS CLI to create secrets in Secrets Manager with naming pattern `${environment}/${service}/${secret_name}` | G-INFRA-SECRETS | 10 | P1 | Backlog | 003-T009 | yes | | |
| 003-T046 | Create Terraform data sources for Secrets Manager: reference existing secrets for database URL, Anthropic API key, JWT secret | G-INFRA-SECRETS | 10 | P1 | Backlog | 003-T009 | yes | | |
| 003-T047 | Update ECS task definition in T025 to reference secret ARNs from T046 in `secrets` block | | 10 | P1 | Backlog | 003-T025, 003-T046 | no | | |

#### Phase 5 — User Story 3: CI/CD Pipeline & Deployment Promotion (Priority: P1) → **Sprint 10**

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

#### Phase 3 — User Story 1: Backend Engineer Implements Secure API Endpoint (Priority: P1) 🎯 MVP → **Sprint 9**

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

### Spec 005 — System Architecture and Technology Stack &nbsp; `specs/005-system-architecture/tasks.md`

> Cross-spec note: Architecture implementation establishes foundations for all subsequent feature work.
> Backend patterns align with 001/002 application code. Frontend integrates design tokens and accessibility
> requirements from 002. Infrastructure modules depend on 003 cloud strategy decisions. Integration patterns
> unify error handling across all layers.

#### Phase 1 — Setup (Project Initialization) → **Sprint 1**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 005-T001 | Create backend directory structure per plan.md: backend/{cmd/api,internal/{middleware,database,ai,errors},pkg,config,migrations,tests/{integration,fixtures}} | | 1 | P1 | Done | - | no | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/13 | PR #43 - Backend directory structure with .gitkeep files |
| 005-T002 | Create frontend directory structure per plan.md: frontend/src/{components/{primitives,composites},features,hooks,lib,stores,styles,routes} | G-ARCH-SETUP-DIRS | | P1 | Backlog | - | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/20 | |
| 005-T003 | Create e2e directory structure: e2e/{tests,fixtures,playwright.config.ts} | G-ARCH-SETUP-DIRS | | P1 | Backlog | - | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/21 | |
| 005-T004 | Create infrastructure directory structure: infra/{modules/{vpc,ecs,rds,alb,cloudfront,secrets},environments} | G-ARCH-SETUP-DIRS | | P1 | Backlog | - | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/22 | |
| 005-T005 | Initialize Go module in backend/go.mod with Go 1.24+ and core dependencies (Chi, pgx/v5, goose/v3, google/uuid, log/slog) | | 1 | P1 | Done | 005-T001 | no | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/14 | PR #44 - Go 1.25.7 with core dependencies |
| 005-T006 | Initialize React project in frontend/ with Vite, TypeScript strict mode, and core dependencies (TanStack Query v5, Zustand, React Router v7, Lucide React) | G-ARCH-SETUP-INIT | | P1 | Backlog | 005-T002 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/23 | |
| 005-T007 | Initialize E2E project in e2e/ with Playwright and axe-core dependencies | G-ARCH-SETUP-INIT | | P1 | Backlog | 005-T003 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/24 | |
| 005-T008 | Create backend/.env.example with required environment variables (DATABASE_URL, HTTP_PORT, LOG_LEVEL, ANTHROPIC_API_KEY) | G-ARCH-SETUP-CONFIG | 1 | P1 | Done | 005-T001 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/25 | PR #45 - Complete template with 25+ config options |
| 005-T009 | Create frontend/.env.example with VITE_API_BASE_URL variable | G-ARCH-SETUP-CONFIG | | P1 | Backlog | 005-T002 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/26 | |
| 005-T010 | Create backend/README.md with quickstart instructions, directory structure explanation, and development workflow | G-ARCH-SETUP-DOCS | 1 | P1 | Done | 005-T001 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/27 | PR #45 - Enhanced with config validation examples |
| 005-T011 | Create frontend/README.md with development server instructions, component guidelines, and testing commands | G-ARCH-SETUP-DOCS | | P1 | Backlog | 005-T002 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/28 | |
| 005-T012 | Create infra/README.md with Terraform initialization instructions and environment deployment guide | G-ARCH-SETUP-DOCS | | P1 | Backlog | 005-T004 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/29 | |

#### Phase 2 — Foundational (Blocking Prerequisites) → **Sprint 1**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 005-T013 | Create backend/config/config.go with configuration struct and environment variable loading using os.Getenv with validation | | 1 | P1 | Done | 005-T001 | no | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/15 | PR #45 - Includes tests with 100% coverage |
| 005-T014 | Create backend/Dockerfile with multi-stage build (builder stage with Go 1.24+, runtime stage with minimal Alpine) | G-ARCH-FOUNDATIONAL-DOCKER | | P1 | Backlog | 005-T001 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/30 | |
| 005-T015 | Create .gitignore files for backend/ (exclude vendor/, .env, binary), frontend/ (exclude node_modules/, dist/, .env), and infra/ (exclude .terraform/, *.tfstate) | G-ARCH-FOUNDATIONAL-DOCKER | | P1 | In Progress | 005-T001 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/31 | Partial: backend/.gitignore done (PR #45) |
| 005-T016 | Create docker-compose.yml for local development with PostgreSQL 15.4 service and backend service configuration | G-ARCH-FOUNDATIONAL-DOCKER | | P1 | Backlog | 005-T001 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/32 | |
| 005-T017 | Create .github/workflows/backend-ci.yml skeleton (lint, test, build jobs without full implementation) | G-ARCH-FOUNDATIONAL-CI | | P1 | Backlog | 005-T001 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/33 | |
| 005-T018 | Create .github/workflows/frontend-ci.yml skeleton (lint, test, build, accessibility jobs without full implementation) | G-ARCH-FOUNDATIONAL-CI | | P1 | Backlog | 005-T002 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/34 | |
| 005-T019 | Create .github/workflows/infra-plan.yml skeleton (validate, format check, plan jobs) | G-ARCH-FOUNDATIONAL-CI | | P1 | Backlog | 005-T004 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/35 | |
| 005-T020 | Configure golangci-lint in backend/.golangci.yml with required linters (errcheck, govet, staticcheck, revive, gosec) | | 1 | P1 | Done | 005-T001 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/16 | PR #45 - 15 linters enabled with test exclusions |
| 005-T021 | Configure ESLint and Prettier in frontend/ with TypeScript strict mode rules and no-any enforcement | | | P1 | Backlog | 005-T002 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/17 | |
| 005-T022 | Create infra/backend.tf with S3 backend configuration for remote state (bucket: traveler-terraform-state, DynamoDB table: traveler-terraform-locks) | G-ARCH-FOUNDATIONAL-TERRAFORM | | P1 | Backlog | 005-T004 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/18 | |
| 005-T023 | Create infra/versions.tf with Terraform >= 1.5 and AWS provider ~> 5.0 version constraints | | | P1 | Backlog | 005-T004 | yes | https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/19 | |

#### Phase 3 — User Story 1: Backend Service Architecture (Priority: P1) 🎯 MVP → **Sprint 2**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 005-T024 | Create backend/internal/middleware/request_id.go implementing UUID v4 generation, context injection, and X-Request-ID response header | G-SPRINT2-BACKEND-MIDDLEWARE | 2 | P1 | Backlog | 005-T013 | yes | | |
| 005-T025 | Create backend/internal/middleware/logger.go using log/slog for structured JSON logging with method, path, status, duration, correlation ID | G-SPRINT2-BACKEND-MIDDLEWARE | 2 | P1 | Backlog | 005-T013 | yes | | |
| 005-T026 | Create backend/internal/middleware/recovery.go implementing panic recovery with stack trace logging and 500 response | G-SPRINT2-BACKEND-MIDDLEWARE | 2 | P1 | Backlog | 005-T013 | yes | | |
| 005-T027 | Create backend/internal/middleware/cors.go with configurable allowed origins from environment variable | G-SPRINT2-BACKEND-MIDDLEWARE | 2 | P1 | Backlog | 005-T013 | yes | | |
| 005-T028 | Create backend/internal/middleware/body_size.go limiting request body to 10 MB with 413 response on violation | G-SPRINT2-BACKEND-MIDDLEWARE | 2 | P1 | Backlog | 005-T013 | yes | | |
| 005-T029 | Create backend/internal/database/client.go implementing pgx connection pool with min 5, max 25 connections, health check (Ping), and graceful closure | | 2 | P1 | Backlog | 005-T013 | yes | | |
| 005-T030 | Create backend/internal/ai/client.go defining AIClient interface with GenerateItinerary and StreamItinerary methods | G-SPRINT2-BACKEND-AI | 2 | P1 | Backlog | 005-T013 | yes | | |
| 005-T031 | Create backend/internal/ai/validator.go implementing prompt validation stub (to be enhanced with injection detection rules later) | G-SPRINT2-BACKEND-AI | 2 | P1 | Backlog | 005-T013 | yes | | |
| 005-T032 | Create backend/internal/ai/sanitizer.go implementing output sanitization stub (HTML/script stripping to be enhanced later) | G-SPRINT2-BACKEND-AI | 2 | P1 | Backlog | 005-T013 | yes | | |
| 005-T033 | Create backend/internal/errors/handler.go implementing domain error-to-HTTP status mapping (404, 400, 401, 403, 409, 500) with structured JSON responses | G-SPRINT2-BACKEND-ERRORS | 2 | P1 | Backlog | 005-T013 | yes | | |
| 005-T034 | Create backend/internal/errors/types.go defining domain error types (ErrNotFound, ErrValidation, ErrUnauthorized, ErrForbidden, ErrConflict) | G-SPRINT2-BACKEND-ERRORS | 2 | P1 | Backlog | 005-T013 | yes | | |
| 005-T035 | Create backend/cmd/api/main.go implementing HTTPServer with Chi router, middleware chain registration (RequestID → Logger → Recovery → CORS → BodySize), health check endpoint, and graceful shutdown | | 2 | P1 | Backlog | 005-T024, 005-T029 | no | | |
| 005-T036 | Integrate backend/internal/database/client.go initialization in main.go with configuration from config package and connection pool lifecycle management | | 2 | P1 | Backlog | 005-T035 | no | | |
| 005-T037 | Add /healthz endpoint to main.go verifying database Ping() succeeds before returning 200 OK | | 2 | P1 | Backlog | 005-T036 | no | | |
| 005-T038 | Create backend/internal/example/model.go with sample domain model struct demonstrating naming conventions and field tags | G-SPRINT2-BACKEND-EXAMPLE | 2 | P1 | Backlog | 005-T013 | yes | | |
| 005-T039 | Create backend/internal/example/repository.go implementing repository interface pattern with Create, FindByID, Update, Delete, List methods using pgx connection pool | G-SPRINT2-BACKEND-EXAMPLE | 2 | P1 | Backlog | 005-T029 | yes | | |
| 005-T040 | Create backend/internal/example/service.go implementing service interface pattern with business logic, repository dependency injection, and domain error returns | G-SPRINT2-BACKEND-EXAMPLE | 2 | P1 | Backlog | 005-T039 | yes | | |
| 005-T041 | Create backend/internal/example/handler.go implementing HTTP handler calling service layer, using errors.HandleError for error responses, and demonstrating context value extraction (requestID, userID) | | 2 | P1 | Backlog | 005-T040 | no | | |

#### Phase 4 — User Story 2: Frontend Application Structure (Priority: P1) 🎯 MVP → **Sprint 2**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 005-T042 | Create frontend/src/styles/tokens.css defining CSS custom properties for colors (primary, surface, text), spacing (sm, md, lg), typography (font-size-base, font-size-lg, font-weight-bold), border-radius (radius-md), and shadows (shadow-sm, shadow-md) | G-SPRINT2-FRONTEND-TOKENS | 2 | P1 | Backlog | 005-T002 | yes | | |
| 005-T043 | Create frontend/src/styles/global.css importing tokens.css and setting base styles (font-family, box-sizing, CSS reset) | G-SPRINT2-FRONTEND-TOKENS | 2 | P1 | Backlog | 005-T042 | yes | | |
| 005-T044 | Create frontend/src/lib/api-client.ts implementing fetch wrapper with base URL from env var, Content-Type and X-Request-ID headers, credentials include, and APIError class (status, message, requestId fields) | G-SPRINT2-FRONTEND-API-CONFIG | 2 | P1 | Backlog | 005-T002 | yes | | |
| 005-T045 | Create frontend/src/lib/query-client.ts configuring TanStack Query defaults (staleTime: 5 min, retry: 1, refetchOnWindowFocus: false) | G-SPRINT2-FRONTEND-API-CONFIG | 2 | P1 | Backlog | 005-T002 | yes | | |
| 005-T046 | Create frontend/src/stores/auth-store.ts implementing Zustand store with isAuthenticated, user, login, logout, refreshSession actions (no persistence for MVP - session storage can be added later) | | 2 | P1 | Backlog | 005-T002 | yes | | |
| 005-T047 | Create frontend/src/App.tsx wrapping application with QueryClientProvider from lib/query-client.ts and importing global.css | G-SPRINT2-FRONTEND-APP-SHELL | 2 | P1 | Backlog | 005-T043, 005-T045 | no | | |
| 005-T048 | Create frontend/src/components/primitives/Button.tsx with variant prop (primary, secondary, danger), size prop (sm, md, lg), CSS Module styling using design tokens, and aria-label support | G-SPRINT2-FRONTEND-PRIMITIVES-CORE | 2 | P1 | Backlog | 005-T042 | yes | | |
| 005-T049 | Create frontend/src/components/primitives/Button.module.css referencing var(--color-primary), var(--space-md), var(--radius-md) from tokens.css | G-SPRINT2-FRONTEND-PRIMITIVES-CORE | 2 | P1 | Backlog | 005-T042 | yes | | |
| 005-T050 | Create frontend/src/components/primitives/Input.tsx with label, id, name, required props, CSS Module styling, and associated label for accessibility | G-SPRINT2-FRONTEND-PRIMITIVES-CORE | 2 | P1 | Backlog | 005-T042 | yes | | |
| 005-T051 | Create frontend/src/components/primitives/Card.tsx with children prop and CSS Module styling using var(--color-surface), var(--shadow-sm), var(--space-lg) | G-SPRINT2-FRONTEND-PRIMITIVES-CORE | 2 | P1 | Backlog | 005-T042 | yes | | |
| 005-T052 | Create frontend/src/components/primitives/LoadingSpinner.tsx with aria-label prop for screen readers | G-SPRINT2-FRONTEND-PRIMITIVES-STATE | 2 | P1 | Backlog | 005-T042 | yes | | |
| 005-T053 | Create frontend/src/components/primitives/ErrorMessage.tsx displaying error with retry button (optional onClick prop) | G-SPRINT2-FRONTEND-PRIMITIVES-STATE | 2 | P1 | Backlog | 005-T042 | yes | | |
| 005-T054 | Create frontend/src/components/primitives/EmptyState.tsx with message and optional action button | G-SPRINT2-FRONTEND-PRIMITIVES-STATE | 2 | P1 | Backlog | 005-T042 | yes | | |
| 005-T055 | Create frontend/src/components/composites/Form.tsx composing Button and Input primitives, handling onSubmit with loading state, error display, and validation error mapping | | 2 | P1 | Backlog | 005-T048, 005-T050 | no | | |
| 005-T056 | Create frontend/src/features/.gitkeep as placeholder (actual features will be added in subsequent specs) | | 2 | P1 | Backlog | 005-T002 | yes | | |
| 005-T057 | Create frontend/src/components/ErrorBoundary.tsx implementing React.Component error boundary with fallback UI showing error message and "Go Home" action | | 2 | P1 | Backlog | 005-T002 | yes | | |
| 005-T058 | Create frontend/src/routes/index.tsx defining React Router v7 routes configuration (root route returning simple "TrAIveler" heading as placeholder) | | 2 | P1 | Backlog | 005-T002 | no | | |
| 005-T059 | Update frontend/src/App.tsx to include RouterProvider with routes from routes/index.tsx | G-SPRINT2-FRONTEND-APP-SHELL | 2 | P1 | Backlog | 005-T047, 005-T058 | no | | |
| 005-T060 | Wrap App.tsx with ErrorBoundary component | G-SPRINT2-FRONTEND-APP-SHELL | 2 | P1 | Backlog | 005-T057, 005-T059 | no | | |

#### Phase 5 — User Story 3: Infrastructure as Code Foundations (Priority: P1) 🎯 MVP → **Sprints 2-3**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 005-T061 | Create infra/modules/vpc/main.tf defining aws_vpc resource with CIDR from var.vpc_cidr, enable_dns_hostnames, enable_dns_support | G-ARCH-INFRA-VPC | | P1 | Backlog | 005-T022 | yes | | |
| 005-T062 | Add aws_subnet.public resources in infra/modules/vpc/main.tf creating 2 public subnets across 2 AZs with cidrsubnet() and count | G-ARCH-INFRA-VPC | | P1 | Backlog | 005-T022 | yes | | |
| 005-T063 | Add aws_subnet.private resources in infra/modules/vpc/main.tf creating 2 private subnets across 2 AZs | G-ARCH-INFRA-VPC | | P1 | Backlog | 005-T022 | yes | | |
| 005-T064 | Add aws_internet_gateway and aws_route_table resources for public subnet routing in infra/modules/vpc/main.tf | G-ARCH-INFRA-VPC | | P1 | Backlog | 005-T022 | yes | | |
| 005-T065 | Add conditional NAT resource (aws_instance for staging, aws_nat_gateway for production) based on var.enable_nat_gateway in infra/modules/vpc/main.tf | G-ARCH-INFRA-VPC | | P1 | Backlog | 005-T022 | yes | | |
| 005-T066 | Create infra/modules/vpc/variables.tf defining environment, vpc_cidr, availability_zones, enable_nat_gateway variables | | | P1 | Backlog | 005-T061 | no | | |
| 005-T067 | Create infra/modules/vpc/outputs.tf exporting vpc_id, public_subnet_ids, private_subnet_ids | | | P1 | Backlog | 005-T066 | no | | |
| 005-T068 | Add resource tagging in infra/modules/vpc/main.tf with Environment, ManagedBy, Project tags per FR-023 | | | P1 | Backlog | 005-T061 | no | | |
| 005-T069 | Create infra/modules/ecs/main.tf defining aws_ecs_cluster resource | G-ARCH-INFRA-ECS | | P1 | Backlog | 005-T022 | yes | | |
| 005-T070 | Add aws_ecs_task_definition in infra/modules/ecs/main.tf with ARM64 architecture, task_cpu and task_memory from variables, container definition with ECR image URL, environment variables, CloudWatch log configuration | G-ARCH-INFRA-ECS | | P1 | Backlog | 005-T022 | yes | | |
| 005-T071 | Add aws_ecs_service in infra/modules/ecs/main.tf with desired_count, launch_type FARGATE, network_configuration using private subnets, load_balancer attachment to ALB target group, health_check_grace_period_seconds 60 | G-ARCH-INFRA-ECS | | P1 | Backlog | 005-T022 | yes | | |
| 005-T072 | Add aws_appautoscaling_target and aws_appautoscaling_policy in infra/modules/ecs/main.tf for CPU-based target tracking at 70% with min_tasks and max_tasks from variables | G-ARCH-INFRA-ECS | | P1 | Backlog | 005-T022 | yes | | |
| 005-T073 | Add aws_cloudwatch_log_group in infra/modules/ecs/main.tf with retention_in_days from variable | G-ARCH-INFRA-ECS | | P1 | Backlog | 005-T022 | yes | | |
| 005-T074 | Add aws_security_group for ECS tasks in infra/modules/ecs/main.tf allowing ingress from ALB security group on port 8080 | G-ARCH-INFRA-ECS | | P1 | Backlog | 005-T022 | yes | | |
| 005-T075 | Create infra/modules/ecs/variables.tf defining environment, vpc_id, private_subnet_ids, task_cpu, task_memory, min_tasks, max_tasks, log_retention_days, ecr_repository_url variables | | | P1 | Backlog | 005-T069 | no | | |
| 005-T076 | Create infra/modules/ecs/outputs.tf exporting cluster_name, service_name, task_definition_family, log_group_name | | | P1 | Backlog | 005-T075 | no | | |
| 005-T077 | Add resource tagging in infra/modules/ecs/main.tf with Environment, ManagedBy, Project tags | | | P1 | Backlog | 005-T069 | no | | |
| 005-T078 | Create infra/modules/rds/main.tf defining aws_db_subnet_group using private subnets | G-ARCH-INFRA-RDS | | P1 | Backlog | 005-T022 | yes | | |
| 005-T079 | Add aws_db_instance in infra/modules/rds/main.tf with engine postgresql, engine_version 15.4, instance_class from variable, allocated_storage 20, multi_az from variable, backup_retention_period from variable, backup_window 03:00-04:00, maintenance_window sun:04:00-sun:05:00 | G-ARCH-INFRA-RDS | | P1 | Backlog | 005-T022 | yes | | |
| 005-T080 | Add aws_security_group for RDS in infra/modules/rds/main.tf allowing ingress from ECS security group on port 5432 | G-ARCH-INFRA-RDS | | P1 | Backlog | 005-T022 | yes | | |
| 005-T081 | Add aws_secretsmanager_secret and aws_secretsmanager_secret_version in infra/modules/rds/main.tf for database credentials (username, password generated with random_password) | G-ARCH-INFRA-RDS | | P1 | Backlog | 005-T022 | yes | | |
| 005-T082 | Create infra/modules/rds/variables.tf defining environment, vpc_id, private_subnet_ids, instance_class, multi_az, backup_retention_days, db_name variables | | | P1 | Backlog | 005-T078 | no | | |
| 005-T083 | Create infra/modules/rds/outputs.tf exporting db_endpoint, db_name, db_secret_arn (marked sensitive) | | | P1 | Backlog | 005-T082 | no | | |
| 005-T084 | Add resource tagging in infra/modules/rds/main.tf with Environment, ManagedBy, Project tags | | | P1 | Backlog | 005-T078 | no | | |
| 005-T085 | Create infra/modules/alb/main.tf defining aws_lb resource with load_balancer_type application, subnets from public_subnet_ids, security_groups | G-ARCH-INFRA-ALB | | P1 | Backlog | 005-T022 | yes | | |
| 005-T086 | Add aws_lb_target_group in infra/modules/alb/main.tf with target_type ip, port 8080, protocol HTTP, health_check path /healthz, interval 30, timeout 5, healthy_threshold 2, unhealthy_threshold 3 | G-ARCH-INFRA-ALB | | P1 | Backlog | 005-T022 | yes | | |
| 005-T087 | Add aws_lb_listener for HTTPS (port 443) in infra/modules/alb/main.tf with ssl_policy, certificate_arn from variable, default_action forward to target group | G-ARCH-INFRA-ALB | | P1 | Backlog | 005-T022 | yes | | |
| 005-T088 | Add aws_lb_listener for HTTP (port 80) in infra/modules/alb/main.tf with redirect action to HTTPS | G-ARCH-INFRA-ALB | | P1 | Backlog | 005-T022 | yes | | |
| 005-T089 | Add aws_security_group for ALB in infra/modules/alb/main.tf allowing ingress on ports 80 and 443 from 0.0.0.0/0 | G-ARCH-INFRA-ALB | | P1 | Backlog | 005-T022 | yes | | |
| 005-T090 | Create infra/modules/alb/variables.tf defining environment, vpc_id, public_subnet_ids, certificate_arn, ecs_security_group_id variables | | | P1 | Backlog | 005-T085 | no | | |
| 005-T091 | Create infra/modules/alb/outputs.tf exporting alb_dns_name, alb_zone_id, target_group_arn | | | P1 | Backlog | 005-T090 | no | | |
| 005-T092 | Add resource tagging in infra/modules/alb/main.tf with Environment, ManagedBy, Project tags | | | P1 | Backlog | 005-T085 | no | | |
| 005-T093 | Create infra/modules/cloudfront/main.tf defining aws_s3_bucket for frontend builds with private ACL | G-ARCH-INFRA-CLOUDFRONT | | P1 | Backlog | 005-T022 | yes | | |
| 005-T094 | Add aws_cloudfront_origin_access_identity and bucket policy in infra/modules/cloudfront/main.tf granting OAI read access to S3 | G-ARCH-INFRA-CLOUDFRONT | | P1 | Backlog | 005-T022 | yes | | |
| 005-T095 | Add aws_cloudfront_distribution in infra/modules/cloudfront/main.tf with S3 origin, default_cache_behavior (viewer_protocol_policy redirect-to-https, allowed_methods GET HEAD OPTIONS), custom_error_response for SPA routing (404 → /index.html), aliases from domain_name variable, viewer_certificate with certificate_arn | G-ARCH-INFRA-CLOUDFRONT | | P1 | Backlog | 005-T022 | yes | | |
| 005-T096 | Create infra/modules/cloudfront/variables.tf defining environment, domain_name, certificate_arn variables | | | P1 | Backlog | 005-T093 | no | | |
| 005-T097 | Create infra/modules/cloudfront/outputs.tf exporting s3_bucket_name, cloudfront_distribution_id, cloudfront_domain_name | | | P1 | Backlog | 005-T096 | no | | |
| 005-T098 | Add resource tagging in infra/modules/cloudfront/main.tf with Environment, ManagedBy, Project tags | | | P1 | Backlog | 005-T093 | no | | |
| 005-T099 | Create infra/modules/secrets/main.tf defining aws_secretsmanager_secret resources for db_credentials, ai_api_key, jwt_signing_key (secrets created empty, values populated manually post-apply) | G-ARCH-INFRA-SECRETS | | P1 | Backlog | 005-T022 | yes | | |
| 005-T100 | Create infra/modules/secrets/variables.tf defining environment variable | | | P1 | Backlog | 005-T099 | no | | |
| 005-T101 | Create infra/modules/secrets/outputs.tf exporting db_secret_arn, ai_api_key_secret_arn, jwt_signing_key_secret_arn (all marked sensitive) | | | P1 | Backlog | 005-T100 | no | | |
| 005-T102 | Create infra/main.tf calling vpc, ecs, rds, alb, cloudfront, secrets modules with dependency injection via module outputs | | | P1 | Backlog | 005-T067, 005-T076, 005-T083, 005-T091, 005-T097, 005-T101 | no | | |
| 005-T103 | Create infra/variables.tf defining all root-level variables (environment, vpc_cidr, availability_zones, enable_nat_gateway, task_cpu, task_memory, min_tasks, max_tasks, log_retention_days, instance_class, multi_az, backup_retention_days, backend_domain, frontend_domain) | | | P1 | Backlog | 005-T102 | no | | |
| 005-T104 | Create infra/outputs.tf exporting CI/CD-relevant outputs (ecr_repository_url, ecs_cluster_name, ecs_service_name, s3_bucket_name, cloudfront_distribution_id) | | | P1 | Backlog | 005-T103 | no | | |
| 005-T105 | Create infra/environments/staging.tfvars with cost-optimized configuration (vpc_cidr 10.0.0.0/16, enable_nat_gateway false, task_cpu 256, task_memory 512, min_tasks 1, max_tasks 2, log_retention_days 7, instance_class db.t4g.micro, multi_az false, backup_retention_days 1) | | | P1 | Backlog | 005-T103 | yes | | |
| 005-T106 | Create infra/environments/production.tfvars with production configuration (vpc_cidr 10.1.0.0/16, enable_nat_gateway true, task_cpu 1024, task_memory 2048, min_tasks 2, max_tasks 20, log_retention_days 30, instance_class db.t4g.small, multi_az true, backup_retention_days 30) | | | P1 | Backlog | 005-T103 | yes | | |
| 005-T107 | Update .github/workflows/infra-plan.yml implementing terraform init, terraform validate, terraform fmt -check, terraform plan for both staging and production .tfvars, and PR comment with plan output | | | P1 | Backlog | 005-T019, 005-T104 | no | | |
| 005-T108 | Create .github/workflows/infra-apply.yml implementing terraform apply -auto-approve for staging on main merge, with output export to GitHub Secrets for backend-ci.yml and frontend-ci.yml | | | P1 | Backlog | 005-T107 | no | | |
| 005-T109 | Configure OIDC federation in AWS IAM (manual step documented in infra/README.md) creating IAM role with trust policy for GitHub Actions and permissions for Terraform operations | | | P1 | Backlog | 005-T108 | no | | |

#### Phase 6 — User Story 4: Integration and Error Handling Patterns (Priority: P2) → **Sprint 4**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 005-T110 | Enhance backend/internal/errors/handler.go to extract correlation ID from context and include in all error responses | G-ARCH-INTEGRATION-BACKEND | | P2 | Backlog | 005-T033 | yes | | |
| 005-T111 | Update backend/internal/middleware/logger.go to log all 5xx errors with stack trace and correlation ID for observability | G-ARCH-INTEGRATION-BACKEND | | P2 | Backlog | 005-T025 | yes | | |
| 005-T112 | Create backend/internal/ai/anthropic.go implementing Anthropic SDK wrapper with exponential backoff retry (max 3 attempts), timeout configuration (60s non-streaming), and 429/503 error handling | G-ARCH-INTEGRATION-BACKEND | | P2 | Backlog | 005-T030 | yes | | |
| 005-T113 | Update backend/internal/ai/client.go to use anthropic.go implementation and return service unavailable errors with Retry-After header on AI provider outages | | | P2 | Backlog | 005-T112 | no | | |
| 005-T114 | Enhance frontend/src/lib/api-client.ts to include X-Request-ID in all requests using crypto.randomUUID() and extract requestId from error responses | G-ARCH-INTEGRATION-FRONTEND | | P2 | Backlog | 005-T044 | yes | | |
| 005-T115 | Create frontend/src/hooks/useErrorHandler.ts custom hook handling common error scenarios (401 → redirect to login, 403 → show permission error, 404 → show not found, 500/503 → show retry) | G-ARCH-INTEGRATION-FRONTEND | | P2 | Backlog | 005-T044 | yes | | |
| 005-T116 | Update frontend/src/components/primitives/ErrorMessage.tsx to display correlation ID when available in error responses | | | P2 | Backlog | 005-T053, 005-T115 | no | | |
| 005-T117 | Create integration test scenario in backend/tests/integration/health_test.go verifying health check returns 200 OK with correlation ID header | G-ARCH-INTEGRATION-TESTS | | P2 | Backlog | 005-T037 | yes | | |
| 005-T118 | Create integration test scenario in backend/tests/integration/error_test.go verifying database timeout returns 500 with correlation ID and structured error response | G-ARCH-INTEGRATION-TESTS | | P2 | Backlog | 005-T033 | yes | | |
| 005-T119 | Document error handling patterns in backend/README.md and frontend/README.md with examples of domain error creation, error wrapping, and client error handling | | | P2 | Backlog | 005-T010, 005-T011, 005-T116 | no | | |

#### Phase 7 — Polish & Cross-Cutting Concerns (Priority: P2-P3) → **Sprint 4**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 005-T120 | Update docs/architecture.md with final backend/frontend/infrastructure architecture diagrams and component descriptions from specs/005-system-architecture/ | G-ARCH-POLISH-DOCS | | P2 | Backlog | 005-T109 | yes | | |
| 005-T121 | Update docs/coding-guidelines.md with Go import order (stdlib → external → internal), TypeScript import order (react → external → @/ → relative), and file naming conventions | G-ARCH-POLISH-DOCS | | P2 | Backlog | 005-T109 | yes | | |
| 005-T122 | Update docs/testing-guidelines.md with backend testing patterns (service mocks, repository integration tests), frontend testing patterns (React Testing Library, MSW, axe-core) | G-ARCH-POLISH-DOCS | | P2 | Backlog | 005-T109 | yes | | |
| 005-T123 | Update project root README.md with architecture overview, getting started instructions (Docker Compose local dev), and links to backend/frontend/infra READMEs | G-ARCH-POLISH-DOCS | | P2 | Backlog | 005-T109 | yes | | |
| 005-T124 | Run all quickstart validation scenarios from specs/005-system-architecture/quickstart.md (Scenario 1: Backend skeleton, Scenario 2: Frontend scaffold, Scenario 3: Infrastructure plan, Scenario 4: E2E integration, Scenario 5: Accessibility) and document results | G-ARCH-POLISH-VALIDATION | | P3 | Backlog | 005-T060, 005-T109 | no | | |
| 005-T125 | Validate golangci-lint passes with zero errors on backend code | G-ARCH-POLISH-VALIDATION | | P3 | Backlog | 005-T020, 005-T041 | yes | | |
| 005-T126 | Validate ESLint and Prettier pass with zero errors on frontend code | G-ARCH-POLISH-VALIDATION | | P3 | Backlog | 005-T021, 005-T060 | yes | | |
| 005-T127 | Validate terraform fmt check passes on all infra files | G-ARCH-POLISH-VALIDATION | | P3 | Backlog | 005-T109 | yes | | |
| 005-T128 | Run backend unit tests and verify ≥ 80% coverage on example domain service and repository | G-ARCH-POLISH-VALIDATION | | P3 | Backlog | 005-T041 | yes | | |
| 005-T129 | Run frontend component tests and verify primitives (Button, Input, Card) render correctly with design tokens | G-ARCH-POLISH-VALIDATION | | P3 | Backlog | 005-T060 | yes | | |
| 005-T130 | Run axe-core accessibility audit on frontend and verify zero WCAG 2.1 AA violations | G-ARCH-POLISH-VALIDATION | | P3 | Backlog | 005-T060 | yes | | |
| 005-T131 | Update specs/005-system-architecture/tasks.md marking all tasks complete and adding completion notes | | | P3 | Backlog | 005-T124 | no | | |

---

### Spec 006 — Core Domain and Data Model Foundations &nbsp; `specs/006-core-domain-model/tasks.md`

> Cross-spec note: Documentation feature completing the canonical domain model reference. Gap-fills
> docs/data-model.md with missing entities (ConversationSession, ConversationMessage), validation layers,
> indexes, state transitions, and cascade behavior. Provides definitive reference for all feature work.

#### Phase 1 — Setup & Validation Infrastructure

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 006-T001 | Verify specs/006-core-domain-model/ directory structure exists with spec.md, plan.md, research.md, data-model.md, quickstart.md, checklists/requirements.md | | | P1 | Backlog | - | no | | |
| 006-T002 | Review research.md gap analysis and confirm all 6 gaps identified (missing entities, incomplete Destination, validation layers, indexes, state transitions, cascade behavior) | | | P1 | Backlog | - | yes | | |
| 006-T003 | Validate plan.md constitution check passed with no violations | | | P1 | Backlog | - | yes | | |

#### Phase 2 — User Story 1: Core Entity Reference (Priority: P1) 🎯 MVP

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 006-T004 | Verify all 16 entities documented in specs/006-core-domain-model/data-model.md (User, RefreshToken, JWTSigningKey, SecurityEvent, Plan, Subscription, Trip, Day, Activity, Collaborator, Suggestion, Destination, TravelStyle, TripTravelStyle, ConversationSession, ConversationMessage) | G-DOC-ENTITY-VERIFICATION | | P1 | Backlog | 006-T001 | yes | | |
| 006-T005 | Verify ConversationSession entity has complete definition (id, trip_id, started_at, completed_at, status, total_tokens, ai_provider, created_at attributes documented) | G-DOC-ENTITY-VERIFICATION | | P1 | Backlog | 006-T001 | yes | | |
| 006-T006 | Verify ConversationMessage entity has complete definition (id, session_id, role, content, token_count, timestamp attributes documented) | G-DOC-ENTITY-VERIFICATION | | P1 | Backlog | 006-T001 | yes | | |
| 006-T007 | Verify Destination entity includes geographic attributes (id, name, country, region, latitude, longitude, created_at attributes documented with coordinate constraints) | G-DOC-ENTITY-VERIFICATION | | P1 | Backlog | 006-T001 | yes | | |
| 006-T008 | Confirm all 16 entities have data types specified for each attribute (UUID, VARCHAR with length, TEXT, INT, BIGINT, DECIMAL, TIMESTAMP, ENUM, BOOLEAN, JSONB) | | | P1 | Backlog | 006-T004 | no | | |
| 006-T009 | Confirm all 16 entities have nullability specified for each attribute (NOT NULL or NULLABLE) | | | P1 | Backlog | 006-T004 | no | | |
| 006-T010 | Confirm all 16 entities have constraint documentation (PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK constraints) | | | P1 | Backlog | 006-T004 | no | | |
| 006-T011 | Validate spec.md acceptance scenarios 1-3 for US1 are met (User, Trip, Collaborator entities findable with complete definitions) | | | P1 | Backlog | 006-T008 | no | | |

#### Phase 3 — User Story 2: Relationship and Constraint Understanding (Priority: P1)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 006-T012 | Verify ERD in specs/006-core-domain-model/data-model.md includes all 16 entities with relationship lines showing cardinality (1:1, 1:many, many:many) | | | P1 | Backlog | 006-T004 | yes | | |
| 006-T013 | Verify all foreign key attributes specify target table and column (e.g., `user_id` (UUID, FK → users.id)) | | | P1 | Backlog | 006-T004 | yes | | |
| 006-T014 | Verify cascade behavior documented for all foreign keys in Cascade Behavior sections (User, Trip, Day, Activity, Collaborator, Suggestion, ConversationSession, ConversationMessage, Subscription, RefreshToken, SecurityEvent, TripTravelStyle) | G-DOC-CASCADE-BEHAVIOR | | P1 | Backlog | 006-T012 | no | | |
| 006-T015 | Confirm Trip cascade behavior documents Days → CASCADE, Activities → CASCADE (via Day), Collaborator → CASCADE, Suggestion → CASCADE, ConversationSession → CASCADE | G-DOC-CASCADE-BEHAVIOR | | P1 | Backlog | 006-T012 | yes | | |
| 006-T016 | Confirm User cascade behavior documents RefreshToken → CASCADE, SecurityEvent → SET NULL, Trip → RESTRICT, Collaborator → CASCADE, Suggestion → CASCADE | G-DOC-CASCADE-BEHAVIOR | | P1 | Backlog | 006-T012 | yes | | |
| 006-T017 | Confirm unique constraints documented (User.email, RefreshToken.token_hash, Day (trip_id, day_number), Collaborator (trip_id, user_id), TripTravelStyle composite PK) | | | P1 | Backlog | 006-T014 | no | | |
| 006-T018 | Validate spec.md acceptance scenarios 1-3 for US2 are met (Trip deletion cascades, User deletion cascade rules, Collaborator UNIQUE constraint documented) | | | P1 | Backlog | 006-T017 | no | | |

#### Phase 4 — User Story 3: Business Rule Enforcement (Priority: P1)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 006-T019 | Verify all 16 entities have Validation Rules section with [DB], [Logic], [API] layer tags in specs/006-core-domain-model/data-model.md | G-DOC-VALIDATION-LAYERS | | P1 | Backlog | 006-T011 | yes | | |
| 006-T020 | Verify User entity has validation rules tagged (email uniqueness [DB], password complexity [Logic], email format [API]) | G-DOC-VALIDATION-LAYERS | | P1 | Backlog | 006-T011 | yes | | |
| 006-T021 | Verify Trip entity has optimistic locking rules documented (version field, If-Match header requirement, 409 Conflict response) | G-DOC-VALIDATION-LAYERS | | P1 | Backlog | 006-T011 | yes | | |
| 006-T022 | Verify Subscription entity has plan limit enforcement documented (Basic plan: 1 admin, 1 partner enforced [Logic]) | G-DOC-VALIDATION-LAYERS | | P1 | Backlog | 006-T011 | yes | | |
| 006-T023 | Verify Invariants and Business Rules section documents 26 global rules (authentication, authorization, concurrency, data protection, plan limits, GDPR, state transition immutability) | | | P1 | Backlog | 006-T019 | no | | |
| 006-T024 | Confirm password handling rules documented (bcrypt cost 12+, never return password_hash in API responses, never log passwords) | | | P1 | Backlog | 006-T023 | yes | | |
| 006-T025 | Confirm token handling rules documented (JWT RS256 with active keys, access token 24h expiry, refresh token 30d expiry, revocation on logout) | | | P1 | Backlog | 006-T023 | yes | | |
| 006-T026 | Validate spec.md acceptance scenarios 1-3 for US3 are met (authentication rules, trip editing rules, subscription limit rules findable) | | | P1 | Backlog | 006-T025 | no | | |

#### Phase 5 — User Story 4: State Transition Clarity (Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 006-T027 | Verify User entity State Transitions section documents Active → Deleted as forward-only with immutable audit trail rationale in specs/006-core-domain-model/data-model.md | G-DOC-STATE-TRANSITIONS | | P2 | Backlog | 006-T026 | yes | | |
| 006-T028 | Verify Subscription entity State Transitions section documents stub_pending → active → cancelled as forward-only (no reactivation path) | G-DOC-STATE-TRANSITIONS | | P2 | Backlog | 006-T026 | yes | | |
| 006-T029 | Verify Trip entity State Transitions section documents draft → published as forward-only (no unpublish operation) | G-DOC-STATE-TRANSITIONS | | P2 | Backlog | 006-T026 | yes | | |
| 006-T030 | Verify Suggestion entity State Transitions section documents pending → approved/rejected as forward-only (resubmission creates new record) | G-DOC-STATE-TRANSITIONS | | P2 | Backlog | 006-T026 | yes | | |
| 006-T031 | Verify ConversationSession entity State Transitions section documents in_progress → completed/abandoned as forward-only | G-DOC-STATE-TRANSITIONS | | P2 | Backlog | 006-T026 | yes | | |
| 006-T032 | Confirm all stateful entities include rationale explaining immutable audit trail requirement (prevents data tampering, preserves decision history) | | | P2 | Backlog | 006-T027 | no | | |
| 006-T033 | Validate spec.md acceptance scenarios 1-3 for US4 are met (trip publishing, subscription management, suggestion workflow transitions documented as forward-only) | | | P2 | Backlog | 006-T032 | no | | |

#### Phase 6 — User Story 5: Concurrency and Versioning Strategy (Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 006-T034 | Verify Trip entity has Concurrency Control section documenting version field, If-Match header, WHERE clause check, 409 Conflict response format in specs/006-core-domain-model/data-model.md | G-DOC-CONCURRENCY | | P2 | Backlog | 006-T033 | yes | | |
| 006-T035 | Verify Activity entity has Concurrency Control section documenting same optimistic locking mechanism as Trip | G-DOC-CONCURRENCY | | P2 | Backlog | 006-T033 | yes | | |
| 006-T036 | Confirm Trip entity documents version increment strategy (SET version = version + 1 on successful update, atomic increment) | | | P2 | Backlog | 006-T034 | no | | |
| 006-T037 | Confirm Concurrency invariants #6-8 documented (optimistic locking enforced, version numbers increment atomically, 409 Conflict with current version in response) | | | P2 | Backlog | 006-T036 | no | | |
| 006-T038 | Validate spec.md acceptance scenarios 1-3 for US5 are met (trip updates with If-Match, activity updates with version increment, conflict handling with 409 response documented) | | | P2 | Backlog | 006-T037 | no | | |

#### Phase 7 — Performance-Critical Indexes

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 006-T039 | Verify User entity indexes documented (idx_users_email UNIQUE expression index for case-insensitive lookup, idx_users_subscription_id) in specs/006-core-domain-model/data-model.md | G-DOC-INDEXES | | P2 | Backlog | 006-T038 | yes | | |
| 006-T040 | Verify Trip entity indexes documented (idx_trips_creator_id for user's trip list, idx_trips_status for published/draft filtering) | G-DOC-INDEXES | | P2 | Backlog | 006-T038 | yes | | |
| 006-T041 | Verify Day entity indexes documented (idx_days_trip_id for trip detail queries, idx_days_destination_id for destination usage) | G-DOC-INDEXES | | P2 | Backlog | 006-T038 | yes | | |
| 006-T042 | Verify Activity entity indexes documented (idx_activities_day_id for day detail queries ordered by sequence) | G-DOC-INDEXES | | P2 | Backlog | 006-T038 | yes | | |
| 006-T043 | Verify Collaborator entity indexes documented (idx_collaborators_trip_id, idx_collaborators_user_id for collaboration queries) | G-DOC-INDEXES | | P2 | Backlog | 006-T038 | yes | | |
| 006-T044 | Verify Suggestion entity indexes documented (idx_suggestions_trip_id, idx_suggestions_author_id, idx_suggestions_status for suggestion management) | G-DOC-INDEXES | | P2 | Backlog | 006-T038 | yes | | |
| 006-T045 | Verify Subscription entity indexes documented (idx_subscriptions_user_id UNIQUE for one subscription per user, idx_subscriptions_status) | G-DOC-INDEXES | | P2 | Backlog | 006-T038 | yes | | |
| 006-T046 | Verify ConversationSession entity indexes documented (idx_conversation_sessions_trip_id, idx_conversation_sessions_status) | G-DOC-INDEXES | | P2 | Backlog | 006-T038 | yes | | |
| 006-T047 | Verify ConversationMessage entity indexes documented (idx_conversation_messages_session_id for message history chronological ordering) | G-DOC-INDEXES | | P2 | Backlog | 006-T038 | yes | | |
| 006-T048 | Verify Destination entity indexes documented (idx_destinations_country, optional spatial index for coordinates if PostGIS enabled) | G-DOC-INDEXES | | P2 | Backlog | 006-T038 | yes | | |

#### Phase 8 — Developer Reference Guide

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 006-T049 | Verify specs/006-core-domain-model/quickstart.md contains 12 validation scenarios covering all 5 user stories | G-DOC-QUICKSTART | | P2 | Backlog | 006-T048 | yes | | |
| 006-T050 | Verify Scenario 1 (three-layer validation) references User entity validation rules from data-model.md | G-DOC-QUICKSTART | | P2 | Backlog | 006-T048 | yes | | |
| 006-T051 | Verify Scenario 2 (optimistic locking) references Trip concurrency control from data-model.md | G-DOC-QUICKSTART | | P2 | Backlog | 006-T048 | yes | | |
| 006-T052 | Verify Scenario 3 (cascade deletes) references Trip cascade behavior from data-model.md | G-DOC-QUICKSTART | | P2 | Backlog | 006-T048 | yes | | |
| 006-T053 | Verify Scenario 4 (forward-only transitions) references Subscription state transitions from data-model.md | G-DOC-QUICKSTART | | P2 | Backlog | 006-T048 | yes | | |
| 006-T054 | Verify Scenario 5 (suggest-then-approve) references Collaborator and Suggestion entities from data-model.md | G-DOC-QUICKSTART | | P2 | Backlog | 006-T048 | yes | | |
| 006-T055 | Verify Scenario 6 (plan limits) references Plan and Collaborator business rules from data-model.md | G-DOC-QUICKSTART | | P2 | Backlog | 006-T048 | yes | | |
| 006-T056 | Verify Scenario 7 (GDPR deletion) references User cascade behavior and invariants #16-17 from data-model.md | G-DOC-QUICKSTART | | P2 | Backlog | 006-T048 | yes | | |
| 006-T057 | Verify Scenario 8 (geographic data) references Destination entity with coordinates from data-model.md | G-DOC-QUICKSTART | | P2 | Backlog | 006-T048 | yes | | |
| 006-T058 | Verify Scenario 9 (activity sequencing) references Activity sequence_order from data-model.md | G-DOC-QUICKSTART | | P2 | Backlog | 006-T048 | yes | | |
| 006-T059 | Verify Scenario 10 (token tracking) references ConversationSession and ConversationMessage from data-model.md | G-DOC-QUICKSTART | | P2 | Backlog | 006-T048 | yes | | |
| 006-T060 | Verify Scenario 11 (security audit trail) references SecurityEvent cascade behavior from data-model.md | G-DOC-QUICKSTART | | P2 | Backlog | 006-T048 | yes | | |
| 006-T061 | Verify Scenario 12 (JWT rotation) references JWTSigningKey rotation strategy from data-model.md | G-DOC-QUICKSTART | | P2 | Backlog | 006-T048 | yes | | |

#### Phase 9 — Completeness Validation & Success Criteria

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 006-T062 | Validate SC-001: All 16 entities locatable without questions (entity catalog complete, organized alphabetically in data-model.md) | G-DOC-SUCCESS-CRITERIA | | P1 | Backlog | 006-T061 | yes | | |
| 006-T063 | Validate SC-002: Zero ambiguity issues (all attributes have types, nullability, constraints; validation layers tagged) | G-DOC-SUCCESS-CRITERIA | | P1 | Backlog | 006-T061 | yes | | |
| 006-T064 | Validate SC-003: 100% entity relationship coverage (ERD includes all entities, all FK relationships documented with cascade behavior) | G-DOC-SUCCESS-CRITERIA | | P1 | Backlog | 006-T061 | yes | | |
| 006-T065 | Validate SC-004: 100% business rule coverage (all 16 entities have Business Rules section, 26 global invariants documented) | G-DOC-SUCCESS-CRITERIA | | P1 | Backlog | 006-T061 | yes | | |
| 006-T066 | Validate SC-005: Migration sequence documented (16 migrations listed in order in data-model.md Database Migrations section) | G-DOC-SUCCESS-CRITERIA | | P1 | Backlog | 006-T061 | yes | | |
| 006-T067 | Validate SC-006: Cascade behavior prevents FK violations (all foreign keys document CASCADE, SET NULL, or RESTRICT) | G-DOC-SUCCESS-CRITERIA | | P1 | Backlog | 006-T061 | yes | | |
| 006-T068 | Validate SC-007: Concurrency-sensitive operations documented (Trip and Activity have Concurrency Control sections with optimistic locking) | G-DOC-SUCCESS-CRITERIA | | P1 | Backlog | 006-T061 | yes | | |
| 006-T069 | Validate SC-008: Security-sensitive field handling documented (password_hash bcrypt rules, token storage rules, private key AWS Secrets Manager ARN, PII removal rules) | G-DOC-SUCCESS-CRITERIA | | P1 | Backlog | 006-T061 | yes | | |
| 006-T070 | Update specs/006-core-domain-model/checklists/requirements.md with final validation status (all 16 checklist items passing) | | | P1 | Backlog | 006-T062 | no | | |

#### Phase 10 — Documentation Promotion & Handoff

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 006-T071 | Review specs/006-core-domain-model/data-model.md for accuracy and completeness (all gaps from research.md filled) | | | P1 | Backlog | 006-T070 | no | | |
| 006-T072 | Update docs/data-model.md with promoted content from specs/006-core-domain-model/data-model.md (preserve PROMOTED markers, update promotion date to 2026-07-06) | | | P1 | Backlog | 006-T071 | no | | |
| 006-T073 | Add promotion metadata to docs/data-model.md header (<!-- Generated from specs/006-core-domain-model/data-model.md, Last promoted: 2026-07-06 -->) | | | P1 | Backlog | 006-T072 | no | | |
| 006-T074 | Update .github/memory/session-notes.md with session summary (document gaps filled, entities added, validation strategy) | | | P1 | Backlog | 006-T073 | yes | | |
| 006-T075 | Update .github/memory/patterns-discovered.md if new reusable patterns identified (three-layer validation taxonomy, forward-only state transition pattern) | | | P1 | Backlog | 006-T073 | yes | | |
| 006-T076 | Create PR with title "feat(docs): complete core domain model with gap-filled entity catalog" targeting main branch | | | P1 | Backlog | 006-T074 | no | | |
| 006-T077 | Add PR description summarizing 6 gaps filled, 16 entities documented, 8 success criteria met, and link to specs/006-core-domain-model/spec.md | | | P1 | Backlog | 006-T076 | no | | |

---

### Spec 007 — API Design Standards and Conventions &nbsp; `specs/007-api-design-standards/tasks.md`

> Cross-spec note: Documentation feature establishing API conventions for all backend endpoints. Promotes
> standards document to docs/ as authoritative reference. Integration tests validate endpoint compliance.
> Extends PR template with standards compliance checklist. No code implementation — pure standards definition.

#### Phase 1 — Setup (Documentation Foundation)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 007-T001 | Verify specs/007-api-design-standards/ structure is complete (plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md) | | | P1 | Backlog | - | no | | |
| 007-T002 | Create backend/tests/integration/ directory if not exists | | | P1 | Backlog | - | yes | | |
| 007-T003 | Read existing .github/pull_request_template.md to understand current structure | | | P1 | Backlog | - | yes | | |

#### Phase 2 — Foundational (Standards Document Promotion) ⚠️ CRITICAL BLOCKER

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 007-T004 | Promote specs/007-api-design-standards/contracts/api-design-standards.md to docs/api-design-standards.md (authoritative location) | | | P1 | Backlog | - | no | | |
| 007-T005 | Verify docs/api-design-standards.md includes all 15 sections with examples (resource naming, URL structure, versioning, request/response format, error format, status codes, pagination, filtering, sorting, rate limiting, auth headers, timestamps, endpoint patterns, compliance, references) | | | P1 | Backlog | 007-T004 | no | | |
| 007-T006 | Update .github/copilot-instructions.md Documentation References section to include docs/api-design-standards.md with description | | | P1 | Backlog | - | yes | | |
| 007-T007 | Audit existing endpoints in specs/001-product-vision-scope/contracts/api.md against docs/api-design-standards.md | | | P1 | Backlog | 007-T005 | yes | | |
| 007-T008 | Audit existing endpoints in specs/004-security-auth-model/contracts/api.md against docs/api-design-standards.md | | | P1 | Backlog | 007-T005 | yes | | |
| 007-T009 | Document any deviations found in audits (create specs/007-api-design-standards/audit-report.md with list of compliant vs. non-compliant patterns) | | | P1 | Backlog | 007-T007, 007-T008 | no | | |

#### Phase 3 — User Story 1: Backend Developer Creates Consistent Endpoints (Priority: P1) 🎯 MVP

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 007-T010 | Verify docs/api-design-standards.md Table of Contents has anchor links to all 15 sections | G-API-US1-ACCESSIBILITY | | P1 | Backlog | 007-T005 | yes | | |
| 007-T011 | Verify each standard section includes ✅ DO and ❌ DON'T examples with code snippets | G-API-US1-ACCESSIBILITY | | P1 | Backlog | 007-T005 | yes | | |
| 007-T012 | Verify error codes catalog in docs/api-design-standards.md lists all 11 machine-readable codes (invalid_request, validation_failed, authentication_required, forbidden, not_found, conflict, rate_limit_exceeded, internal_error, etc.) | G-API-US1-ACCESSIBILITY | | P1 | Backlog | 007-T005 | yes | | |
| 007-T013 | Verify endpoint patterns section includes 7 reusable templates (List Resources, Get Single, Create, Update Full, Update Partial, Delete, Action on Resource) | | | P1 | Backlog | 007-T005 | no | | |
| 007-T014 | Create docs/api-design-standards.md quick reference card section at top (1-page summary of all conventions for printing/bookmarking) | | | P1 | Backlog | 007-T013 | no | | |
| 007-T015 | Verify resource naming section documents plural nouns, lowercase, hyphenated compounds with 5+ examples | G-API-US1-NAMING | | P1 | Backlog | 007-T005 | yes | | |
| 007-T016 | Verify URL nesting section documents max 2 levels with nested vs. top-level endpoint guidance | G-API-US1-NAMING | | P1 | Backlog | 007-T005 | yes | | |
| 007-T017 | Add decision tree diagram to docs/api-design-standards.md: "Should this resource be nested or top-level?" | | | P1 | Backlog | 007-T015, 007-T016 | no | | |
| 007-T018 | Verify request format section documents snake_case fields, ISO 8601 timestamps, required vs. optional fields | G-API-US1-FORMAT | | P1 | Backlog | 007-T005 | yes | | |
| 007-T019 | Verify response format section documents flat JSON for singles, envelope for lists with pagination metadata | G-API-US1-FORMAT | | P1 | Backlog | 007-T005 | yes | | |
| 007-T020 | Add side-by-side comparison in docs/api-design-standards.md: correct vs. incorrect field naming examples | | | P1 | Backlog | 007-T018, 007-T019 | no | | |
| 007-T021 | Verify specs/007-api-design-standards/quickstart.md includes 10 validation scenarios covering all standards | | | P1 | Backlog | 007-T005 | yes | | |
| 007-T022 | Add "How to validate your endpoint" checklist to docs/api-design-standards.md referencing quickstart.md scenarios | | | P1 | Backlog | 007-T021 | no | | |

#### Phase 4 — User Story 2: Frontend Developer Integrates Predictably (Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 007-T023 | Verify error format section in docs/api-design-standards.md includes full TypeScript interface for error response structure | G-API-US2-ERROR-HANDLING | | P2 | Backlog | 007-T022 | yes | | |
| 007-T024 | Add frontend integration example to docs/api-design-standards.md showing generic error handler using error codes | G-API-US2-ERROR-HANDLING | | P2 | Backlog | 007-T022 | yes | | |
| 007-T025 | Document in docs/api-design-standards.md which error codes map to which user-facing messages (UX guidance) | | | P2 | Backlog | 007-T023, 007-T024 | no | | |
| 007-T026 | Verify pagination section includes complete response envelope structure with all metadata fields | G-API-US2-PAGINATION | | P2 | Backlog | 007-T022 | yes | | |
| 007-T027 | Add frontend integration example to docs/api-design-standards.md showing generic pagination component using response metadata | G-API-US2-PAGINATION | | P2 | Backlog | 007-T022 | yes | | |
| 007-T028 | Document edge cases in docs/api-design-standards.md: empty lists, page beyond total_pages, per_page validation | | | P2 | Backlog | 007-T026, 007-T027 | no | | |
| 007-T029 | Verify filtering section documents all operators ([eq], [ne], [gt], [gte], [lt], [lte], [in], [like]) with URLSearchParams examples | G-API-US2-FILTERING-SORTING | | P2 | Backlog | 007-T022 | yes | | |
| 007-T030 | Verify sorting section documents minus prefix for descending, comma-separated multi-field with priority order | G-API-US2-FILTERING-SORTING | | P2 | Backlog | 007-T022 | yes | | |
| 007-T031 | Add frontend integration example to docs/api-design-standards.md showing generic query builder constructing filter/sort params | | | P2 | Backlog | 007-T029, 007-T030 | no | | |
| 007-T032 | Verify rate limiting section documents all response headers (X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset) | | | P2 | Backlog | 007-T022 | yes | | |
| 007-T033 | Add frontend integration example to docs/api-design-standards.md showing how to read rate limit headers and implement proactive throttling | | | P2 | Backlog | 007-T032 | no | | |

#### Phase 5 — User Story 3: Technical Lead Reviews API Changes (Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 007-T034 | Read current .github/pull_request_template.md structure | | | P2 | Backlog | 007-T033 | no | | |
| 007-T035 | Add "API Standards Compliance" section to .github/pull_request_template.md with checklist (if PR touches backend/ or adds/modifies API endpoints) | | | P2 | Backlog | 007-T034 | no | | |
| 007-T036 | Create API standards compliance checklist in .github/pull_request_template.md covering 10 categories (resource naming, URL structure, versioning, request format, response format, error format, status codes, pagination, filtering, rate limiting) | | | P2 | Backlog | 007-T035 | no | | |
| 007-T037 | Create docs/api-review-checklist.md with detailed verification steps for each standard (what to look for, common violations, how to verify) | | | P2 | Backlog | 007-T033 | yes | | |
| 007-T038 | Add "For Code Reviewers" section to docs/api-design-standards.md with quick verification tips | | | P2 | Backlog | 007-T033 | yes | | |
| 007-T039 | Update .github/pull_request_template.md to link to docs/api-review-checklist.md in API Standards Compliance section | | | P2 | Backlog | 007-T036, 007-T037 | no | | |
| 007-T040 | Verify docs/api-design-standards.md Exceptions section documents process: document why, what alternative, get technical lead approval | | | P2 | Backlog | 007-T033 | yes | | |
| 007-T041 | Add exception template to docs/api-design-standards.md (required fields: endpoint, standard violated, reason, alternative approach, approver) | | | P2 | Backlog | 007-T040 | no | | |

#### Phase 6 — User Story 4: API Consumer Learns the System (Priority: P3)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 007-T042 | Create docs/api-getting-started.md for external developers (assumes no internal context) with introduction to API philosophy | G-API-US4-EXTERNAL-DOCS | | P3 | Backlog | 007-T041 | yes | | |
| 007-T043 | Add "Key Conventions at a Glance" section to docs/api-getting-started.md (10-item list of most important patterns) | G-API-US4-EXTERNAL-DOCS | | P3 | Backlog | 007-T041 | yes | | |
| 007-T044 | Add 3 complete endpoint examples to docs/api-getting-started.md demonstrating all major conventions (List with pagination, Create with validation error, Get single with success) | | | P3 | Backlog | 007-T042, 007-T043 | no | | |
| 007-T045 | Add "If you know this... then you know that" section to docs/api-getting-started.md (pattern transfer examples) | G-API-US4-PATTERN-RECOGNITION | | P3 | Backlog | 007-T044 | yes | | |
| 007-T046 | Document consistency guarantees in docs/api-getting-started.md: all lists paginate identically, all errors structured identically, all timestamps formatted identically | G-API-US4-PATTERN-RECOGNITION | | P3 | Backlog | 007-T044 | yes | | |
| 007-T047 | Add FAQ section to docs/api-getting-started.md addressing common API consumer questions (How do I handle errors? How do I paginate? How do I filter?) | | | P3 | Backlog | 007-T045, 007-T046 | no | | |
| 007-T048 | Update docs/api-design-standards.md to include "For API Consumers" callouts highlighting patterns that benefit external developers | | | P3 | Backlog | 007-T047 | no | | |
| 007-T049 | Link docs/api-getting-started.md from docs/api-design-standards.md and README.md (make discoverable) | | | P3 | Backlog | 007-T048 | no | | |

#### Phase 7 — Polish & Enforcement (Automated Validation)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 007-T050 | Create backend/tests/integration/api_standards_test.go file structure | | | P2 | Backlog | 007-T022 | yes | | |
| 007-T051 | Implement test helper functions in backend/tests/integration/api_standards_test.go: assertErrorFormat(), assertPaginationFormat(), assertStatusCode(), assertRateLimitHeaders() | G-API-POLISH-TEST-HELPERS | | P2 | Backlog | 007-T050 | yes | | |
| 007-T052 | Write integration test in backend/tests/integration/api_standards_test.go: TestErrorResponseFormat validates error structure with all required fields (error, message, request_id) | G-API-POLISH-INTEGRATION-TESTS | | P2 | Backlog | 007-T051 | yes | | |
| 007-T053 | Write integration test in backend/tests/integration/api_standards_test.go: TestPaginationFormat validates list responses have data envelope and pagination metadata | G-API-POLISH-INTEGRATION-TESTS | | P2 | Backlog | 007-T051 | yes | | |
| 007-T054 | Write integration test in backend/tests/integration/api_standards_test.go: TestStatusCodeSemantics validates GET returns 200, POST returns 201, DELETE returns 204 | G-API-POLISH-INTEGRATION-TESTS | | P2 | Backlog | 007-T051 | yes | | |
| 007-T055 | Write integration test in backend/tests/integration/api_standards_test.go: TestValidationErrorIncludesFields validates 422 responses include fields array with field-level errors | G-API-POLISH-INTEGRATION-TESTS | | P2 | Backlog | 007-T051 | yes | | |
| 007-T056 | Write integration test in backend/tests/integration/api_standards_test.go: TestRateLimitHeaders validates all responses include X-RateLimit-* headers | G-API-POLISH-INTEGRATION-TESTS | | P2 | Backlog | 007-T051 | yes | | |
| 007-T057 | Write integration test in backend/tests/integration/api_standards_test.go: TestTimestampFormat validates timestamps use ISO 8601 with UTC (Z suffix) | G-API-POLISH-INTEGRATION-TESTS | | P2 | Backlog | 007-T051 | yes | | |
| 007-T058 | Write integration test in backend/tests/integration/api_standards_test.go: TestFieldNaming validates response fields use snake_case (not camelCase or PascalCase) | G-API-POLISH-INTEGRATION-TESTS | | P2 | Backlog | 007-T051 | yes | | |
| 007-T059 | Write integration test in backend/tests/integration/api_standards_test.go: TestRequestIDCorrelation validates error response request_id matches X-Request-ID header | G-API-POLISH-INTEGRATION-TESTS | | P2 | Backlog | 007-T051 | yes | | |
| 007-T060 | Research golangci-lint custom linter options for URL structure validation | G-API-POLISH-LINTER | | P2 | Backlog | 007-T022 | yes | | |
| 007-T061 | Document linter integration plan in specs/007-api-design-standards/audit-report.md (which checks can be automated, which remain manual) | G-API-POLISH-LINTER | | P2 | Backlog | 007-T060 | yes | | |
| 007-T062 | Add comment to backend/.golangci.yml noting future custom linter rules for API standards (placeholder for future automation) | G-API-POLISH-LINTER | | P2 | Backlog | 007-T061 | yes | | |
| 007-T063 | Update README.md to link to docs/api-design-standards.md in Documentation section | G-API-POLISH-DOCUMENTATION | | P2 | Backlog | 007-T022 | yes | | |
| 007-T064 | Update docs/coding-guidelines.md to reference docs/api-design-standards.md for API-specific conventions | G-API-POLISH-DOCUMENTATION | | P2 | Backlog | 007-T022 | yes | | |
| 007-T065 | Update specs/007-api-design-standards/audit-report.md with final compliance status of existing endpoints (list compliant, non-compliant, grandfathered exceptions) | G-API-POLISH-DOCUMENTATION | | P2 | Backlog | 007-T009 | yes | | |
| 007-T066 | Run all 10 validation scenarios from specs/007-api-design-standards/quickstart.md against staging environment | G-API-POLISH-VALIDATION | | P3 | Backlog | 007-T065 | yes | | |
| 007-T067 | Document validation results in specs/007-api-design-standards/audit-report.md (which scenarios pass, which need fixes) | G-API-POLISH-VALIDATION | | P3 | Backlog | 007-T066 | yes | | |
| 007-T068 | Run backend/tests/integration/api_standards_test.go test suite and verify all tests pass (or document expected failures for non-compliant endpoints) | G-API-POLISH-VALIDATION | | P3 | Backlog | 007-T059 | yes | | |
| 007-T069 | Review .specify/memory/constitution.md to determine if API standards should be elevated to non-negotiable principles | | | P3 | Backlog | 007-T068 | yes | | |
| 007-T070 | If constitution update warranted, document proposal in specs/007-api-design-standards/constitution-amendment-proposal.md (rationale, proposed changes, impact) | | | P3 | Backlog | 007-T069 | no | | |

---

## Feature Phase

### Spec 008 — Authentication & Collaboration User Experience &nbsp; `specs/008-auth-collaboration-ux/tasks.md`

> Cross-spec note: Authentication feature implements security model from spec 004, uses architecture patterns from spec 005, follows API standards from spec 007, and extends data model from spec 006.

#### Phase 1 — Setup (Shared Infrastructure)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 008-T001 | Create backend project structure: backend/cmd/api/, backend/internal/{auth,subscription,collaboration,security}/, backend/pkg/{database,config}/, backend/tests/{unit,integration,contract}/ | | | P1 | Backlog | 005-T001 | no | | |
| 008-T002 | Initialize Go module with Chi v5, pgx/v5, goose, golang-jwt/jwt v5, bcrypt dependencies in backend/go.mod | G-008-SETUP | | P1 | Backlog | 008-T001 | yes | | |
| 008-T003 | Create frontend project structure: frontend/src/{features,components,stores,styles,services,routes}/, frontend/tests/{unit,integration,e2e}/ | G-008-SETUP | | P1 | Backlog | 005-T002 | yes | | |
| 008-T004 | Initialize Vite React TypeScript project with TanStack Query v5, Zustand, React Router v7, Vitest, Playwright in frontend/package.json | G-008-SETUP | | P1 | Backlog | 008-T003 | yes | | |
| 008-T005 | Configure backend linting: golangci-lint.yml with errcheck, govet, staticcheck, revive, gosec in backend/.golangci.yml | G-008-SETUP | | P1 | Backlog | 008-T001 | yes | | |
| 008-T006 | Configure frontend linting: ESLint + Prettier with TypeScript strict mode in frontend/.eslintrc.json and frontend/.prettierrc | G-008-SETUP | | P1 | Backlog | 008-T003 | yes | | |
| 008-T007 | Create E2E test structure: e2e/specs/{auth,collaboration,accessibility}/ directories | G-008-SETUP | | P1 | Backlog | 005-T003 | yes | | |
| 008-T008 | Create infrastructure directory: infra/terraform/modules/secrets/ for JWT key rotation | G-008-SETUP | | P1 | Backlog | 005-T004 | yes | | |

#### Phase 2 — Foundational (Blocking Prerequisites) → **Sprint 5**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 008-T009 | Setup PostgreSQL connection pooling with pgx/v5 in backend/pkg/database/connection.go | | 5 | P1 | Backlog | 008-T001 | no | | |
| 008-T010 | Configure goose migrations framework in backend/pkg/database/migrations/ directory | G-008-DATABASE | | P1 | Backlog | 008-T009 | yes | | |
| 008-T011 | Create migration 001_create_users.sql: users table with id, email (unique), password_hash, full_name, has_subscription (boolean), failed_login_attempts (integer default 0), last_failed_login_at (timestamp), email_verified (boolean default false), created_at, updated_at, version (integer for optimistic locking) | G-008-MIGRATIONS | | P1 | Backlog | 008-T010 | yes | | |
| 008-T012 | Create migration 002_create_subscriptions.sql: subscriptions table with id, user_id (FK), plan_id (FK), status, current_period_start, current_period_end, grace_period_ends_at (nullable), cancelled_at (nullable), created_at, updated_at, version | G-008-MIGRATIONS | | P1 | Backlog | 008-T010 | yes | | |
| 008-T013 | Create migration 003_create_password_reset_tokens.sql: password_reset_tokens table with id, user_id (FK CASCADE), token_hash (unique), expires_at, used_at (nullable), created_at; indexes on user_id, token_hash, expires_at | G-008-MIGRATIONS | | P1 | Backlog | 008-T010 | yes | | |
| 008-T014 | Create migration 004_create_security_events.sql: security_events table with id, correlation_id (index), event_type (check constraint), user_id (FK SET NULL), email, severity, ip_address, user_agent, details (jsonb), created_at; indexes on correlation_id, user_id, email, created_at, event_type | G-008-MIGRATIONS | | P1 | Backlog | 008-T010 | yes | | |
| 008-T015 | Create migration 005_create_trips.sql: trips table with id, creator_id (FK), destination, start_date, end_date, archived (boolean default false), created_at, updated_at, version; index on creator_id, archived | G-008-MIGRATIONS | | P1 | Backlog | 008-T010 | yes | | |
| 008-T016 | Create migration 006_create_collaborators.sql: collaborators table with id, trip_id (FK CASCADE), user_id (FK CASCADE), email, status (pending/accepted/rejected), invited_at, accepted_at, created_at; indexes on trip_id, user_id, status | G-008-MIGRATIONS | | P1 | Backlog | 008-T010 | yes | | |
| 008-T017 | Create migration 007_create_suggestions.sql: suggestions table with id, trip_id (FK CASCADE), collaborator_id (FK CASCADE), suggestion_type, content, details (jsonb), status (pending/approved/rejected), approved_at, rejected_at, created_at | G-008-MIGRATIONS | | P1 | Backlog | 008-T010 | yes | | |
| 008-T018 | Create environment config loader in backend/pkg/config/config.go: load DATABASE_URL, JWT_SIGNING_KEY_SECRET_ARN, ANTHROPIC_API_KEY_SECRET_ARN from environment | | | P1 | Backlog | 008-T001 | no | | |
| 008-T019 | Implement JWT generator with RS256 in backend/internal/auth/jwt/generator.go: GenerateAccessToken(userID, hasSubscription) returns signed JWT with 24h expiration | G-008-JWT | | P1 | Backlog | 008-T018 | yes | | |
| 008-T020 | Implement JWT validator in backend/internal/auth/jwt/validator.go: ValidateToken(token) returns claims, supports multi-key validation for zero-downtime rotation | G-008-JWT | | P1 | Backlog | 008-T018 | yes | | |
| 008-T021 | Implement JWT refresher in backend/internal/auth/jwt/refresher.go: RefreshToken(refreshToken) validates, revokes old token, issues new access+refresh tokens | G-008-JWT | | P1 | Backlog | 008-T020 | yes | | |
| 008-T022 | Implement bcrypt password hasher in backend/internal/auth/password/hasher.go: HashPassword(password) with cost 12, ComparePassword(hash, password) for validation | G-008-PASSWORD | | P1 | Backlog | 008-T018 | yes | | |
| 008-T023 | Implement password validator in backend/internal/auth/password/validator.go: ValidatePassword(password) checks 8-72 chars, uppercase, lowercase, digit | G-008-PASSWORD | | P1 | Backlog | 008-T018 | yes | | |
| 008-T024 | Implement progressive delay rate limiter in backend/internal/auth/ratelimit/limiter.go: CheckRateLimit(email) tracks failed attempts, returns delay seconds (exponential backoff after 5 failures) | G-008-RATELIMIT | | P1 | Backlog | 008-T018 | yes | | |
| 008-T025 | Implement rate limiter storage in backend/internal/auth/ratelimit/store.go: in-memory map with 15-minute TTL for failed attempt tracking (user_id, count, first_attempt_at) | G-008-RATELIMIT | | P1 | Backlog | 008-T024 | yes | | |
| 008-T026 | Implement security event logger in backend/internal/security/logger.go: LogSecurityEvent(correlationID, eventType, userID, email, severity, ipAddress, userAgent, details) writes to CloudWatch Logs as structured JSON | | | P1 | Backlog | 008-T018 | yes | | |
| 008-T027 | Implement auth middleware in backend/internal/security/middleware.go: ValidateJWTCookie() extracts token from cookie, validates with JWT validator, attaches user context to request | G-008-MIDDLEWARE | | P1 | Backlog | 008-T020 | yes | | |
| 008-T028 | Implement request ID middleware in backend/internal/security/middleware.go: GenerateRequestID() creates correlation ID, adds to context and response header X-Request-ID | G-008-MIDDLEWARE | | P1 | Backlog | 008-T001 | yes | | |
| 008-T029 | Implement rate limit middleware in backend/internal/security/middleware.go: RateLimitMiddleware() checks X-RateLimit headers, returns 429 with Retry-After if exceeded | G-008-MIDDLEWARE | | P1 | Backlog | 008-T024 | yes | | |
| 008-T030 | Setup Chi router with middleware chain in backend/cmd/api/main.go: request ID → logging → CORS → rate limit → recovery | | | P1 | Backlog | 008-T027, 008-T028, 008-T029 | no | | |
| 008-T031 | Create Axios instance in frontend/src/services/api.ts: base URL, withCredentials=true, request/response interceptors (correlation ID, 401 refresh, error mapping) | G-008-FRONTEND-API | | P1 | Backlog | 008-T003 | yes | | |
| 008-T032 | Create error handler utility in frontend/src/services/errorHandler.ts: mapApiError(error) converts API errors to user-friendly messages per FR-019 | G-008-FRONTEND-API | | P1 | Backlog | 008-T031 | yes | | |
| 008-T033 | Create auth store in frontend/src/stores/authStore.ts: Zustand store with user state (id, email, full_name, has_subscription), isAuthenticated boolean, login/logout/setUser actions | | | P1 | Backlog | 008-T003 | yes | | |
| 008-T034 | Create CSS design tokens in frontend/src/styles/tokens.css: CSS custom properties for colors (primary, secondary, error, warning with WCAG AA contrast), spacing scale (4/8/12/16/24/32/48px), typography (font sizes, weights, line heights), focus indicators (outline-width, outline-color, outline-offset) | | | P1 | Backlog | 008-T003 | yes | | |
| 008-T035 | Create global styles in frontend/src/styles/global.css: CSS reset, base typography, box-sizing border-box, accessible focus styles using tokens | | | P1 | Backlog | 008-T034 | yes | | |
| 008-T036 | Create React Router configuration in frontend/src/routes/router.tsx: routes for /register, /login, /dashboard, /trips/:id, /settings, /password-reset, with protected route wrapper checking authStore.isAuthenticated | | | P1 | Backlog | 008-T033 | no | | |

#### Phase 3 — User Story 1: Paid User Registration & First Trip Creation (Priority: P1) 🎯 MVP → **Sprints 6-7**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 008-T037 | Create User model in backend/internal/auth/models.go: User struct with ID, Email, PasswordHash, FullName, HasSubscription, FailedLoginAttempts, LastFailedLoginAt, EmailVerified, CreatedAt, UpdatedAt, Version | G-008-US1-MODELS | 6 | P1 | Backlog | 008-T001 | yes | | |
| 008-T038 | Create RegisterRequest/RegisterResponse models in backend/internal/auth/models.go: RegisterRequest{Email, Password, FullName, PaymentMethodToken}, RegisterResponse{User, Subscription} | G-008-US1-MODELS | | P1 | Backlog | 008-T037 | yes | | |
| 008-T039 | Create Subscription model in backend/internal/subscription/models.go: Subscription struct with ID, UserID, PlanID, Status, CurrentPeriodStart, CurrentPeriodEnd, GracePeriodEndsAt, CancelledAt, CreatedAt, UpdatedAt, Version | G-008-US1-MODELS | | P1 | Backlog | 008-T001 | yes | | |
| 008-T040 | Create User repository in backend/internal/auth/repository.go: CreateUser(user), GetUserByEmail(email), UpdateUser(user) with optimistic locking check | | | P1 | Backlog | 008-T037, 008-T011 | no | | |
| 008-T041 | Create Subscription repository in backend/internal/subscription/repository.go: CreateSubscription(sub), GetSubscriptionByUserID(userID), UpdateSubscription(sub), CancelSubscription(id) | G-008-US1-REPOS | | P1 | Backlog | 008-T039, 008-T012 | yes | | |
| 008-T042 | Create stub payment provider interface in backend/internal/subscription/payment/provider.go: PaymentProvider interface with ProcessPayment(token, planID) method | G-008-US1-REPOS | | P1 | Backlog | 008-T001 | yes | | |
| 008-T043 | Implement stub payment provider in backend/internal/subscription/payment/stub.go: StubPaymentProvider always returns success, logs to console "[DEMO] Payment processed: [token]" | G-008-US1-REPOS | | P1 | Backlog | 008-T042 | yes | | |
| 008-T044 | Implement auth service Register method in backend/internal/auth/service.go: validates email/password, checks email uniqueness (409 if exists), hashes password with bcrypt cost 12, creates User, if paymentMethodToken provided calls subscription service, sets HasSubscription=true, logs auth_registration security event | | | P1 | Backlog | 008-T040, 008-T022 | no | | |
| 008-T045 | Implement subscription service CreateSubscription in backend/internal/subscription/service.go: calls payment provider ProcessPayment, creates Subscription with status='active', current_period_end=now+30days, returns subscription | | | P1 | Backlog | 008-T041, 008-T043 | no | | |
| 008-T046 | Implement POST /auth/register handler in backend/internal/auth/handler.go: validates request body (email format, password strength), calls authService.Register, generates JWT tokens (access 24h, refresh 30d), sets HTTP-only Secure SameSite=Strict cookies, returns 201 with user+subscription JSON | | | P1 | Backlog | 008-T044, 008-T019 | no | | |
| 008-T047 | Register POST /api/v1/auth/register route in backend/cmd/api/main.go: attach register handler to Chi router with rate limit middleware (10/min per IP) | | | P1 | Backlog | 008-T046, 008-T030 | no | | |
| 008-T048 | Create Trip model in backend/internal/collaboration/models.go: Trip struct with ID, CreatorID, Destination, StartDate, EndDate, Archived, CreatedAt, UpdatedAt, Version (note: full AI generation logic deferred to future spec, stub returns hardcoded itinerary) | G-008-US1-TRIP | | P1 | Backlog | 008-T001 | yes | | |
| 008-T049 | Create Trip repository in backend/internal/collaboration/repository.go: CreateTrip(trip), GetTripsByCreatorID(creatorID), GetTripByID(id), UpdateTrip(trip), DeleteTrip(id) | G-008-US1-TRIP | | P1 | Backlog | 008-T048, 008-T015 | yes | | |
| 008-T050 | Implement POST /trips handler in backend/internal/collaboration/handler.go: validates auth, extracts user from context, validates request (destination, dates), creates Trip with CreatorID=user.ID, returns 201 with trip JSON (AI stub: returns trip with hardcoded 3-day Paris itinerary) | | | P1 | Backlog | 008-T049, 008-T027 | no | | |
| 008-T051 | Register POST /api/v1/trips route in backend/cmd/api/main.go: attach trip creation handler with auth middleware (requires valid JWT) | | | P1 | Backlog | 008-T050, 008-T030 | no | | |
| 008-T052 | Create Button primitive in frontend/src/components/primitives/Button.tsx: <Button> with variant (primary/secondary/danger), size (small/medium/large), disabled, loading props; uses design tokens; keyboard accessible | G-008-US1-PRIMITIVES | | P1 | Backlog | 008-T034 | yes | | |
| 008-T053 | Create Input primitive in frontend/src/components/primitives/Input.tsx: <Input> with type, label, error, disabled props; uses design tokens; associated label for accessibility | G-008-US1-PRIMITIVES | | P1 | Backlog | 008-T034 | yes | | |
| 008-T054 | Create Label primitive in frontend/src/components/primitives/Label.tsx: <Label> with htmlFor, required indicator; uses design tokens | G-008-US1-PRIMITIVES | | P1 | Backlog | 008-T034 | yes | | |
| 008-T055 | Create ErrorMessage primitive in frontend/src/components/primitives/ErrorMessage.tsx: <ErrorMessage> displays validation/API errors with icon; uses error color token; aria-live for screen readers | G-008-US1-PRIMITIVES | | P1 | Backlog | 008-T034 | yes | | |
| 008-T056 | Create Form composite in frontend/src/components/composites/Form.tsx: <Form> with onSubmit, children; prevents default, handles loading state, disables submit during loading | | | P1 | Backlog | 008-T052, 008-T053 | no | | |
| 008-T057 | Create LoadingSpinner feature in frontend/src/components/features/LoadingSpinner.tsx: <LoadingSpinner> with message prop; uses primary color token; aria-busy for screen readers | G-008-US1-COMPONENTS | | P1 | Backlog | 008-T034 | yes | | |
| 008-T058 | Create useRegister hook in frontend/src/features/auth/hooks/useRegister.ts: TanStack Query mutation for POST /auth/register, updates authStore on success, handles errors via errorHandler | | | P1 | Backlog | 008-T031, 008-T033 | no | | |
| 008-T059 | Create authApi service in frontend/src/features/auth/services/authApi.ts: register(email, password, fullName, paymentMethodToken) calls axios POST /api/v1/auth/register with withCredentials | G-008-US1-API | | P1 | Backlog | 008-T031 | yes | | |
| 008-T060 | Create RegisterForm component in frontend/src/features/auth/components/RegisterForm.tsx: form with email, password, fullName inputs, "Continue to Payment" checkbox, "Create Free Account" button; uses useRegister hook, validates client-side, shows loading/error states | | | P1 | Backlog | 008-T056, 008-T058 | no | | |
| 008-T061 | Create RegisterPage in frontend/src/features/auth/pages/RegisterPage.tsx: renders RegisterForm, heading "Create Your Account", links to /login; redirects to /dashboard on success | | | P1 | Backlog | 008-T060, 008-T036 | no | | |
| 008-T062 | Create useTrips hook in frontend/src/features/trips/hooks/useTrips.ts: TanStack Query query for GET /trips, returns user's owned trips and collaborations | G-008-US1-TRIP-HOOKS | | P1 | Backlog | 008-T031 | yes | | |
| 008-T063 | Create useCreateTrip hook in frontend/src/features/trips/hooks/useCreateTrip.ts: TanStack Query mutation for POST /trips, invalidates trips query on success | G-008-US1-TRIP-HOOKS | | P1 | Backlog | 008-T031 | yes | | |
| 008-T064 | Create tripsApi service in frontend/src/features/trips/services/tripsApi.ts: getTrips(), createTrip(destination, startDate, endDate) calls axios | G-008-US1-API | | P1 | Backlog | 008-T031 | yes | | |
| 008-T065 | Create TripCard component in frontend/src/features/trips/components/TripCard.tsx: displays trip destination, dates, "View Details" link; uses Card composite; keyboard accessible | | | P1 | Backlog | 008-T052, 008-T034 | no | | |
| 008-T066 | Create Card composite in frontend/src/components/composites/Card.tsx: <Card> with heading, children; uses design tokens for border, padding, shadow | G-008-US1-COMPOSITES | | P1 | Backlog | 008-T034 | yes | | |
| 008-T067 | Create TripDashboard component in frontend/src/features/trips/components/TripDashboard.tsx: renders "My Trips" section (if hasSubscription), "Shared with Me" section, "Create Trip" button (enabled if hasSubscription, disabled with tooltip if Free User), uses useTrips hook, displays LoadingSpinner/ErrorMessage/EmptyState | | | P1 | Backlog | 008-T062, 008-T065, 008-T057 | no | | |
| 008-T068 | Create EmptyState feature in frontend/src/components/features/EmptyState.tsx: <EmptyState> with message, illustration icon, action button; uses design tokens | G-008-US1-COMPONENTS | | P1 | Backlog | 008-T034 | yes | | |
| 008-T069 | Create DashboardPage in frontend/src/features/trips/pages/DashboardPage.tsx: renders TripDashboard, heading "My Trips", protected route (requires auth) | | | P1 | Backlog | 008-T067, 008-T036 | no | | |
| 008-T070 | Create TripDetailPage in frontend/src/features/trips/pages/TripDetailPage.tsx: displays trip destination, dates, day-by-day itinerary (stub: hardcoded Paris 3-day plan), "Edit Trip" button (if owner), "Delete Trip" button (if owner), uses useTripDetail hook (fetch GET /trips/:id) | | | P1 | Backlog | 008-T062, 008-T036 | no | | |

_Checkpoint: Paid User can now register, create subscription, and generate first trip (US1 complete and independently testable)_

#### Phase 4 — User Story 2: Free User Registration & Accepting Collaboration Invite (Priority: P1)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 008-T071 | Create Collaborator model in backend/internal/collaboration/models.go: Collaborator struct with ID, TripID, UserID, Email, Status (pending/accepted/rejected), InvitedAt, AcceptedAt, CreatedAt | G-008-US2-MODELS | | P1 | Backlog | 008-T048 | yes | | |
| 008-T072 | Create Suggestion model in backend/internal/collaboration/models.go: Suggestion struct with ID, TripID, CollaboratorID, SuggestionType, Content, Details (jsonb), Status (pending/approved/rejected), ApprovedAt, RejectedAt, CreatedAt | G-008-US2-MODELS | | P1 | Backlog | 008-T048 | yes | | |
| 008-T073 | Create Collaborator repository in backend/internal/collaboration/repository.go: CreateCollaborator(collab), GetCollaboratorsByTripID(tripID), GetCollaboratorsByUserID(userID), UpdateCollaboratorStatus(id, status), DeleteCollaborator(id), CountActiveCollaborations(userID) | | | P1 | Backlog | 008-T071, 008-T016 | no | | |
| 008-T074 | Implement POST /trips/:trip_id/collaborators handler in backend/internal/collaboration/handler.go: validates auth, checks user is trip creator OR existing collaborator with invite permission, validates email format, checks basic plan limit (1 collaborator max), creates Collaborator with status='pending', logs invitation to console (email stub), returns 201 with collaborator JSON | | | P1 | Backlog | 008-T073, 008-T027 | no | | |
| 008-T075 | Register POST /api/v1/trips/:trip_id/collaborators route in backend/cmd/api/main.go: attach invite handler with auth middleware | | | P1 | Backlog | 008-T074, 008-T030 | no | | |
| 008-T076 | Implement PATCH /collaborators/:id/accept handler in backend/internal/collaboration/handler.go: validates auth, checks user owns collaborator record, if user.HasSubscription=false checks CountActiveCollaborations(userID) <= 0 (returns 400 "Leave current trip first" if > 0), updates status='accepted', sets AcceptedAt=now, returns 200 with collaborator JSON | | | P1 | Backlog | 008-T073, 008-T027 | no | | |
| 008-T077 | Register PATCH /api/v1/collaborators/:id/accept route in backend/cmd/api/main.go: attach accept handler with auth middleware | | | P1 | Backlog | 008-T076, 008-T030 | no | | |
| 008-T078 | Implement DELETE /trips/:trip_id/collaborators/leave handler in backend/internal/collaboration/handler.go: validates auth, checks user is collaborator (not creator), deletes Collaborator record, returns 204 No Content | G-008-US2-HANDLERS | | P1 | Backlog | 008-T073, 008-T027 | yes | | |
| 008-T079 | Register DELETE /api/v1/trips/:trip_id/collaborators/leave route in backend/cmd/api/main.go: attach leave handler with auth middleware | G-008-US2-HANDLERS | | P1 | Backlog | 008-T078, 008-T030 | yes | | |
| 008-T080 | Modify POST /auth/register handler in backend/internal/auth/handler.go: if no paymentMethodToken provided, skip subscription creation, set User.HasSubscription=false, log auth_registration with has_subscription=false detail | | | P1 | Backlog | 008-T046 | no | | |
| 008-T081 | Implement GET /invitations handler in backend/internal/collaboration/handler.go: validates auth, returns Collaborator records where UserID=user.ID AND status='pending', includes trip details (destination, creator name) | G-008-US2-HANDLERS | | P1 | Backlog | 008-T073, 008-T027 | yes | | |
| 008-T082 | Register GET /api/v1/invitations route in backend/cmd/api/main.go: attach invitations list handler with auth middleware | G-008-US2-HANDLERS | | P1 | Backlog | 008-T081, 008-T030 | yes | | |
| 008-T083 | Create Badge primitive in frontend/src/components/primitives/Badge.tsx: <Badge> with variant (free/paid/pending); uses design tokens; displays "Free User" or "Paid User" text | G-008-US2-PRIMITIVES | | P1 | Backlog | 008-T034 | yes | | |
| 008-T084 | Create Modal composite in frontend/src/components/composites/Modal.tsx: <Modal> with isOpen, onClose, title, children; uses design tokens; keyboard trap, ESC to close, focus management | G-008-US2-COMPOSITES | | P1 | Backlog | 008-T034 | yes | | |
| 008-T085 | Create Banner composite in frontend/src/components/composites/Banner.tsx: <Banner> with type (info/warning/error/success), message, action button; uses design tokens; dismissible | G-008-US2-COMPOSITES | | P1 | Backlog | 008-T034 | yes | | |
| 008-T086 | Create useInvitations hook in frontend/src/features/collaboration/hooks/useInvitations.ts: TanStack Query query for GET /invitations, returns pending invitations | G-008-US2-HOOKS | | P1 | Backlog | 008-T031 | yes | | |
| 008-T087 | Create useAcceptInvitation hook in frontend/src/features/collaboration/hooks/useAcceptInvitation.ts: TanStack Query mutation for PATCH /collaborators/:id/accept, invalidates invitations and trips queries on success | G-008-US2-HOOKS | | P1 | Backlog | 008-T031 | yes | | |
| 008-T088 | Create useLeaveTrip hook in frontend/src/features/collaboration/hooks/useLeaveTrip.ts: TanStack Query mutation for DELETE /trips/:trip_id/collaborators/leave, invalidates trips query, shows success confirmation | G-008-US2-HOOKS | | P1 | Backlog | 008-T031 | yes | | |
| 008-T089 | Create collaborationApi service in frontend/src/features/collaboration/services/collaborationApi.ts: getInvitations(), acceptInvitation(id), leaveTrip(tripId), inviteCollaborator(tripId, email) | G-008-US2-API | | P1 | Backlog | 008-T031 | yes | | |
| 008-T090 | Create InvitationCard component in frontend/src/features/collaboration/components/InvitationCard.tsx: displays trip destination, creator, invited date, "Accept" button, "Reject" button; uses Card, Badge; shows error if collaboration limit reached with "Leave Current Trip" button | | | P1 | Backlog | 008-T066, 008-T083 | no | | |
| 008-T091 | Create InvitationsPage in frontend/src/features/collaboration/pages/InvitationsPage.tsx: renders list of InvitationCard, uses useInvitations hook, shows EmptyState if no invitations, LoadingSpinner during fetch | | | P1 | Backlog | 008-T090, 008-T057, 008-T068, 008-T036 | no | | |
| 008-T092 | Modify RegisterPage in frontend/src/features/auth/pages/RegisterPage.tsx: add "Sign up as Free User" option alongside "Continue to Payment", conditionally omit paymentMethodToken from request body | | | P1 | Backlog | 008-T061 | no | | |
| 008-T093 | Modify TripDashboard in frontend/src/features/trips/components/TripDashboard.tsx: add "Shared with Me" section, display collaborator badge (Free User/Paid User), disable "Create Trip" button for Free Users with tooltip "Upgrade to create your own trips" | | | P1 | Backlog | 008-T067 | no | | |
| 008-T094 | Create InviteCollaborator component in frontend/src/features/collaboration/components/InviteCollaborator.tsx: form with email input, "Send Invite" button, displays "Basic plan: 1 collaborator maximum" note, uses inviteCollaborator mutation | | | P1 | Backlog | 008-T056, 008-T089 | no | | |
| 008-T095 | Modify TripDetailPage in frontend/src/features/trips/pages/TripDetailPage.tsx: if user is creator, show "Invite Collaborator" button → opens InviteCollaborator modal; if user is collaborator, show badge and "Leave Trip" button | | | P1 | Backlog | 008-T070, 008-T094 | no | | |
| 008-T096 | Create LeaveTripButton component in frontend/src/features/collaboration/components/LeaveTripButton.tsx: button with confirmation modal "Are you sure you want to leave this trip?", uses useLeaveTrip hook, redirects to /dashboard on success | | | P1 | Backlog | 008-T084, 008-T088 | no | | |

_Checkpoint: Free Users can register, accept invitations, enforce single-collaboration limit, leave trips (US2 complete and independently testable)_

#### Phase 5 — User Story 3: Returning User Login & Trip Management (Priority: P1) → **Sprint 8**

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 008-T097 | Create LoginRequest/LoginResponse models in backend/internal/auth/models.go: LoginRequest{Email, Password}, LoginResponse{User} | G-008-US3-MODELS | 8 | P1 | Backlog | 008-T037 | yes | | |
| 008-T098 | Implement auth service Login method in backend/internal/auth/service.go: validates email format, retrieves user by email (returns 401 "Invalid credentials" if not found, no enumeration), checks rate limit via rate limiter (returns 429 with retry_after if exceeded), compares password hash with bcrypt (returns 401 if mismatch, increments failed attempts), resets failed_login_attempts=0 on success, logs auth_login_success or auth_login_failure security event | | | P1 | Backlog | 008-T040, 008-T022, 008-T024 | no | | |
| 008-T099 | Implement POST /auth/login handler in backend/internal/auth/handler.go: validates request body, calls authService.Login, generates JWT tokens (access 24h, refresh 30d), sets HTTP-only cookies, returns 200 with user JSON | | | P1 | Backlog | 008-T098, 008-T019 | no | | |
| 008-T100 | Register POST /api/v1/auth/login route in backend/cmd/api/main.go: attach login handler with rate limit middleware (progressive delay after 5 failures) | | | P1 | Backlog | 008-T099, 008-T030 | no | | |
| 008-T101 | Implement POST /auth/logout handler in backend/internal/auth/handler.go: validates auth, revokes refresh token (mark as revoked in RefreshToken table), clears access_token and refresh_token cookies, logs auth_logout security event, returns 204 No Content | G-008-US3-HANDLERS | | P1 | Backlog | 008-T027 | yes | | |
| 008-T102 | Register POST /api/v1/auth/logout route in backend/cmd/api/main.go: attach logout handler with auth middleware | G-008-US3-HANDLERS | | P1 | Backlog | 008-T101, 008-T030 | yes | | |
| 008-T103 | Implement GET /trips handler in backend/internal/collaboration/handler.go: validates auth, returns trips where CreatorID=user.ID (owned trips) OR user in Collaborator records (shared trips), excludes archived trips by default (optional ?include_archived=true for creator only), returns 200 with trips array JSON | G-008-US3-HANDLERS | | P1 | Backlog | 008-T049, 008-T027 | yes | | |
| 008-T104 | Register GET /api/v1/trips route in backend/cmd/api/main.go: attach trips list handler with auth middleware | G-008-US3-HANDLERS | | P1 | Backlog | 008-T103, 008-T030 | yes | | |
| 008-T105 | Implement GET /trips/:id handler in backend/internal/collaboration/handler.go: validates auth, retrieves trip, checks user is creator OR collaborator (returns 404 if neither, prevents ID enumeration), returns 200 with trip JSON including full itinerary | G-008-US3-HANDLERS | | P1 | Backlog | 008-T049, 008-T027 | yes | | |
| 008-T106 | Register GET /api/v1/trips/:id route in backend/cmd/api/main.go: attach trip detail handler with auth middleware | G-008-US3-HANDLERS | | P1 | Backlog | 008-T105, 008-T030 | yes | | |
| 008-T107 | Implement PUT /trips/:id handler in backend/internal/collaboration/handler.go: validates auth, checks user is creator (returns 403 if not), checks subscription active OR in grace period (returns 403 "Renew subscription" if grace period expired), validates request body, updates trip with optimistic locking check (returns 409 if version mismatch), returns 200 with updated trip JSON | G-008-US3-HANDLERS | | P1 | Backlog | 008-T049, 008-T027 | yes | | |
| 008-T108 | Register PUT /api/v1/trips/:id route in backend/cmd/api/main.go: attach trip update handler with auth middleware | G-008-US3-HANDLERS | | P1 | Backlog | 008-T107, 008-T030 | yes | | |
| 008-T109 | Implement DELETE /trips/:id handler in backend/internal/collaboration/handler.go: validates auth, checks user is creator (returns 403 if not), checks subscription active (returns 403 if grace period or cancelled), deletes trip (CASCADE deletes collaborators and suggestions), returns 204 No Content | G-008-US3-HANDLERS | | P1 | Backlog | 008-T049, 008-T027 | yes | | |
| 008-T110 | Register DELETE /api/v1/trips/:id route in backend/cmd/api/main.go: attach trip delete handler with auth middleware | G-008-US3-HANDLERS | | P1 | Backlog | 008-T109, 008-T030 | yes | | |
| 008-T111 | Create useLogin hook in frontend/src/features/auth/hooks/useLogin.ts: TanStack Query mutation for POST /auth/login, updates authStore on success, handles 401 "Invalid credentials" and 429 "Rate limit" errors, displays retry_after countdown | G-008-US3-HOOKS | | P1 | Backlog | 008-T031, 008-T033 | yes | | |
| 008-T112 | Create useLogout hook in frontend/src/features/auth/hooks/useLogout.ts: TanStack Query mutation for POST /auth/logout, clears authStore on success, redirects to /login | G-008-US3-HOOKS | | P1 | Backlog | 008-T031, 008-T033 | yes | | |
| 008-T113 | Create LoginForm component in frontend/src/features/auth/components/LoginForm.tsx: form with email, password inputs, "Log In" button, "Forgot Password?" link; uses useLogin hook, validates client-side, shows loading/error states, displays rate limit countdown if 429 error | | | P1 | Backlog | 008-T056, 008-T111 | no | | |
| 008-T114 | Create LoginPage in frontend/src/features/auth/pages/LoginPage.tsx: renders LoginForm, heading "Welcome Back", link to /register; redirects to /dashboard on success, preserves ?redirect query param | | | P1 | Backlog | 008-T113, 008-T036 | no | | |
| 008-T115 | Create useTripDetail hook in frontend/src/features/trips/hooks/useTripDetail.ts: TanStack Query query for GET /trips/:id, returns trip with full itinerary | G-008-US3-TRIP-HOOKS | | P1 | Backlog | 008-T062 | yes | | |
| 008-T116 | Create useUpdateTrip hook in frontend/src/features/trips/hooks/useUpdateTrip.ts: TanStack Query mutation for PUT /trips/:id, invalidates trip and trips queries on success | G-008-US3-TRIP-HOOKS | | P1 | Backlog | 008-T062 | yes | | |
| 008-T117 | Create useDeleteTrip hook in frontend/src/features/trips/hooks/useDeleteTrip.ts: TanStack Query mutation for DELETE /trips/:id, invalidates trips query, redirects to /dashboard on success | G-008-US3-TRIP-HOOKS | | P1 | Backlog | 008-T062 | yes | | |
| 008-T118 | Modify TripDetailPage in frontend/src/features/trips/pages/TripDetailPage.tsx: if user is creator and hasSubscription=true, show "Edit Trip" button → opens edit modal; show "Delete Trip" button → opens confirmation modal; if user is creator and in grace period, show "Renew Subscription" banner with disabled edit button | | | P1 | Backlog | 008-T070, 008-T115 | no | | |
| 008-T119 | Create DeleteTripModal component in frontend/src/features/trips/components/DeleteTripModal.tsx: confirmation modal "Are you sure you want to delete this trip?", uses useDeleteTrip hook | G-008-US3-COMPONENTS | | P1 | Backlog | 008-T084, 008-T117 | yes | | |
| 008-T120 | Modify authStore in frontend/src/stores/authStore.ts: add logout action that clears user state, add setUser action that updates user from API response | | | P1 | Backlog | 008-T033 | no | | |
| 008-T121 | Modify protected route wrapper in frontend/src/routes/router.tsx: check authStore.isAuthenticated, redirect to /login?redirect=<current-path> if not authenticated, restore redirect after login | | | P1 | Backlog | 008-T036 | no | | |
| 008-T122 | Create Navigation component in frontend/src/components/features/Navigation.tsx: displays user name, has_subscription badge, "Log Out" button, uses useLogout hook, responsive mobile menu | | | P1 | Backlog | 008-T083, 008-T112 | no | | |

_Checkpoint: Users can login, view dashboard, manage trips, logout (US3 complete and independently testable)_

#### Phase 6 — User Story 4: Collaboration: Invite & Manage Suggestions (Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 008-T123 | Create Suggestion repository in backend/internal/collaboration/repository.go: CreateSuggestion(suggestion), GetSuggestionsByTripID(tripID), GetSuggestionByID(id), UpdateSuggestionStatus(id, status), ApplySuggestionToTrip(suggestionID) (stub: logs to console "Applied suggestion [id] to trip") | G-008-US4-REPOS | | P2 | Backlog | 008-T072, 008-T017 | yes | | |
| 008-T124 | Implement POST /trips/:trip_id/suggestions handler in backend/internal/collaboration/handler.go: validates auth, checks user is collaborator on trip (not creator, returns 403 if creator), validates request body (suggestion_type, content, details), creates Suggestion with status='pending', returns 201 with suggestion JSON | | | P2 | Backlog | 008-T123, 008-T027 | no | | |
| 008-T125 | Register POST /api/v1/trips/:trip_id/suggestions route in backend/cmd/api/main.go: attach create suggestion handler with auth middleware | | | P2 | Backlog | 008-T124, 008-T030 | no | | |
| 008-T126 | Implement GET /trips/:trip_id/suggestions handler in backend/internal/collaboration/handler.go: validates auth, checks user is creator OR collaborator, returns suggestions for trip filtered by status (optional ?status=pending query param), returns 200 with suggestions array JSON | G-008-US4-HANDLERS | | P2 | Backlog | 008-T123, 008-T027 | yes | | |
| 008-T127 | Register GET /api/v1/trips/:trip_id/suggestions route in backend/cmd/api/main.go: attach list suggestions handler with auth middleware | G-008-US4-HANDLERS | | P2 | Backlog | 008-T126, 008-T030 | yes | | |
| 008-T128 | Implement PATCH /suggestions/:id/approve handler in backend/internal/collaboration/handler.go: validates auth, retrieves suggestion, checks user is trip creator (returns 403 if not), checks suggestion status='pending' (returns 409 if already processed), updates status='approved', sets ApprovedAt=now, calls ApplySuggestionToTrip (stub implementation), returns 200 with suggestion JSON and message "Suggestion approved" | | | P2 | Backlog | 008-T123, 008-T027 | no | | |
| 008-T129 | Register PATCH /api/v1/suggestions/:id/approve route in backend/cmd/api/main.go: attach approve suggestion handler with auth middleware | | | P2 | Backlog | 008-T128, 008-T030 | no | | |
| 008-T130 | Implement PATCH /suggestions/:id/reject handler in backend/internal/collaboration/handler.go: validates auth, checks user is trip creator, checks suggestion status='pending' (returns 409 if already processed), updates status='rejected', sets RejectedAt=now, returns 200 with suggestion JSON and message "Suggestion rejected" | G-008-US4-HANDLERS | | P2 | Backlog | 008-T123, 008-T027 | yes | | |
| 008-T131 | Register PATCH /api/v1/suggestions/:id/reject route in backend/cmd/api/main.go: attach reject suggestion handler with auth middleware | G-008-US4-HANDLERS | | P2 | Backlog | 008-T130, 008-T030 | yes | | |
| 008-T132 | Create useSuggestions hook in frontend/src/features/collaboration/hooks/useSuggestions.ts: TanStack Query query for GET /trips/:trip_id/suggestions, returns trip suggestions grouped by status | G-008-US4-HOOKS | | P2 | Backlog | 008-T031 | yes | | |
| 008-T133 | Create useCreateSuggestion hook in frontend/src/features/collaboration/hooks/useCreateSuggestion.ts: TanStack Query mutation for POST /trips/:trip_id/suggestions, invalidates suggestions query on success | G-008-US4-HOOKS | | P2 | Backlog | 008-T031 | yes | | |
| 008-T134 | Create useApproveSuggestion hook in frontend/src/features/collaboration/hooks/useApproveSuggestion.ts: TanStack Query mutation for PATCH /suggestions/:id/approve, invalidates suggestions and trip queries on success | G-008-US4-HOOKS | | P2 | Backlog | 008-T031 | yes | | |
| 008-T135 | Create useRejectSuggestion hook in frontend/src/features/collaboration/hooks/useRejectSuggestion.ts: TanStack Query mutation for PATCH /suggestions/:id/reject, invalidates suggestions query on success | G-008-US4-HOOKS | | P2 | Backlog | 008-T031 | yes | | |
| 008-T136 | Create SuggestionCard component in frontend/src/features/collaboration/components/SuggestionCard.tsx: displays suggestion type, content, details, collaborator name, status badge (pending/approved/rejected), timestamp; if user is creator and status='pending', show "Approve" and "Reject" buttons; uses Card, Badge | | | P2 | Backlog | 008-T066, 008-T083 | no | | |
| 008-T137 | Create SuggestionsList component in frontend/src/features/collaboration/components/SuggestionsList.tsx: renders list of SuggestionCard, grouped by status, uses useSuggestions hook, shows EmptyState if no suggestions | | | P2 | Backlog | 008-T136, 008-T132, 008-T068 | no | | |
| 008-T138 | Modify TripDetailPage in frontend/src/features/trips/pages/TripDetailPage.tsx: if user is collaborator (not creator), replace "Edit" buttons with "Suggest Change" button next to itinerary items; if user is creator, add "Review Suggestions" section displaying pending suggestions count badge, click opens SuggestionsList modal | | | P2 | Backlog | 008-T118, 008-T137 | no | | |
| 008-T139 | Create SuggestChangeModal component in frontend/src/features/collaboration/components/SuggestChangeModal.tsx: form with suggestion_type select, content textarea, details JSONB editor (simple key-value inputs), "Submit Suggestion" button, uses useCreateSuggestion hook, shows success message "Suggestion submitted. Waiting for creator approval." | | | P2 | Backlog | 008-T084, 008-T133 | no | | |

_Checkpoint: Collaborators can suggest changes, creators can approve/reject, suggestions workflow complete (US4 complete and independently testable)_

#### Phase 7 — User Story 5: Password Management & Account Security (Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 008-T140 | Create PasswordResetToken repository in backend/internal/auth/repository.go: CreatePasswordResetToken(token), GetPasswordResetTokenByHash(hash), MarkTokenAsUsed(id), InvalidatePreviousTokens(userID), CleanupExpiredTokens() | G-008-US5-REPOS | | P2 | Backlog | 008-T013 | yes | | |
| 008-T141 | Create RefreshToken repository in backend/internal/auth/repository.go: CreateRefreshToken(token), GetRefreshTokenByHash(hash), RevokeRefreshToken(id), RevokeAllRefreshTokens(userID), RevokeAllRefreshTokensExceptCurrent(userID, currentTokenID) | G-008-US5-REPOS | | P2 | Backlog | 008-T001 | yes | | |
| 008-T142 | Implement POST /auth/password-reset-request handler in backend/internal/auth/handler.go: validates email format, retrieves user by email (if not found, still returns 200 "Link sent" to prevent enumeration, but logs security event with user_exists=false), invalidates previous reset tokens for user, generates 32-byte random token, hashes with SHA-256, creates PasswordResetToken with expires_at=now+1h, logs to console "[DEMO] Password reset link: http://localhost:3000/password-reset/complete?token=[token]", logs auth_password_reset_request security event, returns 200 with generic message "If an account with that email exists, a password reset link has been sent." | | | P2 | Backlog | 008-T140, 008-T040, 008-T026 | no | | |
| 008-T143 | Register POST /api/v1/auth/password-reset-request route in backend/cmd/api/main.go: attach password reset request handler with rate limit middleware (10/hour per IP) | | | P2 | Backlog | 008-T142, 008-T030 | no | | |
| 008-T144 | Implement POST /auth/password-reset-complete handler in backend/internal/auth/handler.go: validates request body (token, new_password), hashes token with SHA-256, retrieves PasswordResetToken by token_hash, validates expires_at>now AND used_at IS NULL (returns 400 "Invalid or expired token" if validation fails), retrieves user, validates new password format, hashes new password with bcrypt cost 12, updates user.password_hash, marks token as used (set used_at=now), revokes all refresh tokens for user (force re-login on all devices), logs auth_password_reset_complete security event, returns 200 with message "Password reset successfully" | | | P2 | Backlog | 008-T140, 008-T022, 008-T141, 008-T026 | no | | |
| 008-T145 | Register POST /api/v1/auth/password-reset-complete route in backend/cmd/api/main.go: attach password reset complete handler with rate limit middleware | | | P2 | Backlog | 008-T144, 008-T030 | no | | |
| 008-T146 | Implement POST /auth/password-change handler in backend/internal/auth/handler.go: validates auth, validates request body (current_password, new_password, log_out_all_other_devices boolean), retrieves user, compares current_password hash with bcrypt (returns 401 if mismatch), validates new password format, hashes new password with bcrypt cost 12, updates user.password_hash, if log_out_all_other_devices=true revokes all refresh tokens except current token, if log_out_all_other_devices=false no token revocation, logs auth_password_change security event, returns 200 with message "Password changed successfully" | G-008-US5-HANDLERS | | P2 | Backlog | 008-T027, 008-T022, 008-T141, 008-T026 | yes | | |
| 008-T147 | Register POST /api/v1/auth/password-change route in backend/cmd/api/main.go: attach password change handler with auth middleware | G-008-US5-HANDLERS | | P2 | Backlog | 008-T146, 008-T030 | yes | | |
| 008-T148 | Implement POST /auth/refresh handler in backend/internal/auth/handler.go: extracts refresh_token from cookie, validates refresh token via JWT validator, checks token not revoked in RefreshToken table (returns 401 "Invalid refresh token" if revoked or expired), generates new access token (24h), optionally rotates refresh token (generate new 30d, revoke old one), sets new cookies, logs auth_token_refresh security event, returns 200 with message "Token refreshed" | G-008-US5-HANDLERS | | P2 | Backlog | 008-T020, 008-T141, 008-T026 | yes | | |
| 008-T149 | Register POST /api/v1/auth/refresh route in backend/cmd/api/main.go: attach token refresh handler (no auth middleware, uses refresh token from cookie) | G-008-US5-HANDLERS | | P2 | Backlog | 008-T148, 008-T030 | yes | | |
| 008-T150 | Modify auth middleware in backend/internal/security/middleware.go: on 401 Unauthorized (expired access token), return 401 with error="TOKEN_EXPIRED"; frontend will detect and call /auth/refresh | | | P2 | Backlog | 008-T027 | no | | |
| 008-T151 | Create usePasswordResetRequest hook in frontend/src/features/auth/hooks/usePasswordResetRequest.ts: TanStack Query mutation for POST /auth/password-reset-request, shows success message "Link sent" | G-008-US5-HOOKS | | P2 | Backlog | 008-T031 | yes | | |
| 008-T152 | Create usePasswordResetComplete hook in frontend/src/features/auth/hooks/usePasswordResetComplete.ts: TanStack Query mutation for POST /auth/password-reset-complete, redirects to /login on success with success banner "Password reset successfully" | G-008-US5-HOOKS | | P2 | Backlog | 008-T031 | yes | | |
| 008-T153 | Create usePasswordChange hook in frontend/src/features/auth/hooks/usePasswordChange.ts: TanStack Query mutation for POST /auth/password-change, shows success message, optionally logs out other devices | G-008-US5-HOOKS | | P2 | Backlog | 008-T031 | yes | | |
| 008-T154 | Create PasswordResetRequestForm component in frontend/src/features/auth/components/PasswordResetRequestForm.tsx: form with email input, "Send Reset Link" button, uses usePasswordResetRequest hook, shows generic success message | | | P2 | Backlog | 008-T056, 008-T151 | no | | |
| 008-T155 | Create PasswordResetRequestPage in frontend/src/features/auth/pages/PasswordResetRequestPage.tsx: renders PasswordResetRequestForm, heading "Reset Your Password", link back to /login | | | P2 | Backlog | 008-T154, 008-T036 | no | | |
| 008-T156 | Create PasswordResetCompleteForm component in frontend/src/features/auth/components/PasswordResetCompleteForm.tsx: form with new_password, confirm_password inputs, "Reset Password" button, extracts token from URL query param, uses usePasswordResetComplete hook, shows error if token invalid/expired | G-008-US5-COMPONENTS | | P2 | Backlog | 008-T056, 008-T152 | yes | | |
| 008-T157 | Create PasswordResetCompletePage in frontend/src/features/auth/pages/PasswordResetCompletePage.tsx: renders PasswordResetCompleteForm, heading "Set New Password" | G-008-US5-COMPONENTS | | P2 | Backlog | 008-T156, 008-T036 | yes | | |
| 008-T158 | Create PasswordChangeForm component in frontend/src/features/auth/components/PasswordChangeForm.tsx: form with current_password, new_password, confirm_password inputs, "Log out all other devices" checkbox, "Change Password" button, uses usePasswordChange hook, shows success message | G-008-US5-COMPONENTS | | P2 | Backlog | 008-T056, 008-T153 | yes | | |
| 008-T159 | Create AccountSettingsPage in frontend/src/features/auth/pages/AccountSettingsPage.tsx: renders PasswordChangeForm, heading "Account Settings", protected route | G-008-US5-COMPONENTS | | P2 | Backlog | 008-T158, 008-T036 | yes | | |
| 008-T160 | Modify axios interceptor in frontend/src/services/api.ts: on 401 response with error="TOKEN_EXPIRED", call POST /auth/refresh, retry original request with new access token; if refresh fails (401 "Invalid refresh token"), clear authStore, redirect to /login?redirect=<original-path> with message "Session expired. Please log in again." | | | P2 | Backlog | 008-T031, 008-T033 | no | | |

_Checkpoint: Password reset and password change flows complete, session management with automatic refresh (US5 complete and independently testable)_

#### Phase 8 — User Story 6: UI Component Library & Design System Basics (Priority: P3)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 008-T161 | Create Checkbox primitive in frontend/src/components/primitives/Checkbox.tsx: <Checkbox> with label, checked, onChange, disabled props; uses design tokens; keyboard accessible; associated label | G-008-US6-PRIMITIVES | | P3 | Backlog | 008-T034 | yes | | |
| 008-T162 | Create Select primitive in frontend/src/components/primitives/Select.tsx: <Select> with options, label, error, disabled props; uses design tokens; keyboard accessible; ARIA attributes | G-008-US6-PRIMITIVES | | P3 | Backlog | 008-T034 | yes | | |
| 008-T163 | Create Textarea primitive in frontend/src/components/primitives/Textarea.tsx: <Textarea> with label, error, disabled, maxLength props; uses design tokens; character counter | G-008-US6-PRIMITIVES | | P3 | Backlog | 008-T034 | yes | | |
| 008-T164 | Create Link primitive in frontend/src/components/primitives/Link.tsx: <Link> wraps React Router Link; uses design tokens; keyboard accessible; visible focus indicator | G-008-US6-PRIMITIVES | | P3 | Backlog | 008-T034 | yes | | |
| 008-T165 | Create Toast composite in frontend/src/components/composites/Toast.tsx: <Toast> notification with type (success/error/info/warning), message, auto-dismiss; uses design tokens; aria-live for screen readers | G-008-US6-COMPOSITES | | P3 | Backlog | 008-T034 | yes | | |
| 008-T166 | Create Tooltip composite in frontend/src/components/composites/Tooltip.tsx: <Tooltip> with trigger, content; uses design tokens; keyboard accessible (on focus); ARIA attributes | G-008-US6-COMPOSITES | | P3 | Backlog | 008-T034 | yes | | |
| 008-T167 | Create ErrorBoundary feature in frontend/src/components/features/ErrorBoundary.tsx: React error boundary component, catches runtime errors, displays user-friendly "Something went wrong" message with "Reload Page" button, logs error to console (no stack trace to user) | | | P3 | Backlog | 008-T034 | no | | |
| 008-T168 | Audit all forms for accessibility: verify all <Input>/<Select>/<Checkbox> have associated <Label> with htmlFor, verify error messages use ErrorMessage component with aria-live, verify focus indicators visible on all interactive elements | G-008-US6-AUDITS | | P3 | Backlog | 008-T053, 008-T162, 008-T161 | yes | | |
| 008-T169 | Audit all interactive elements for keyboard accessibility: verify all buttons/links/form elements reachable via Tab, verify Enter/Space activate buttons, verify ESC closes modals, verify focus trap in modals | G-008-US6-AUDITS | | P3 | Backlog | 008-T052, 008-T084 | yes | | |
| 008-T170 | Audit all text for color contrast: verify all text meets WCAG AA 4.5:1 contrast ratio against background, adjust design tokens if needed, use browser DevTools Accessibility panel | G-008-US6-AUDITS | | P3 | Backlog | 008-T034 | yes | | |
| 008-T171 | Audit all async actions for loading states: verify all mutations (register, login, create trip, submit suggestion, etc.) show <LoadingSpinner> during processing, disable submit button during loading to prevent double submission | | | P3 | Backlog | 008-T057 | no | | |
| 008-T172 | Audit all error scenarios for user-friendly messages: verify all API errors mapped via errorHandler.ts to clear messages, verify no raw error messages or stack traces displayed, verify actionable next steps provided (e.g., "Try again" button) | | | P3 | Backlog | 008-T032 | no | | |
| 008-T173 | Audit all empty states for EmptyState component: verify "No trips yet" → "Create Your First Trip" button, "No invitations" → "Invitations appear here" message, "No suggestions" → "Suggestions appear here" message | | | P3 | Backlog | 008-T068 | no | | |
| 008-T174 | Create documentation for design system in frontend/src/styles/README.md: document all design tokens (colors, spacing, typography, focus), document Atomic Design component layering (Primitives → Composites → Features), document accessibility guidelines (keyboard nav, focus indicators, WCAG AA), provide usage examples for each token | G-008-US6-DOCS | | P3 | Backlog | 008-T034 | yes | | |

_Checkpoint: Design system complete, all components accessible, consistent visual identity (US6 complete and independently testable)_

#### Phase 9 — Subscription Lifecycle Management (Additional Priority: P2)

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 008-T175 | Implement DELETE /subscriptions/:id handler in backend/internal/subscription/handler.go: validates auth, checks user owns subscription, updates Subscription set status='cancelled', cancelled_at=now, grace_period_ends_at=now+30days, updates User.has_subscription=false, returns 200 with subscription JSON and message "Subscription cancelled. Read-only access until [grace_period_ends_at]. Renew anytime." | G-008-SUBSCRIPTION | | P2 | Backlog | 008-T041, 008-T027 | yes | | |
| 008-T176 | Register DELETE /api/v1/subscriptions/:id route in backend/cmd/api/main.go: attach cancel subscription handler with auth middleware | G-008-SUBSCRIPTION | | P2 | Backlog | 008-T175, 008-T030 | yes | | |
| 008-T177 | Implement POST /subscriptions/:id/renew handler in backend/internal/subscription/handler.go: validates auth, checks user owns subscription, checks subscription status='cancelled', calls payment provider ProcessPayment (stub succeeds), updates Subscription set status='active', cancelled_at=null, grace_period_ends_at=null, updates User.has_subscription=true, updates all owned trips set archived=false, returns 200 with subscription JSON and message "Subscription renewed successfully" | G-008-SUBSCRIPTION | | P2 | Backlog | 008-T041, 008-T027, 008-T043 | yes | | |
| 008-T178 | Register POST /api/v1/subscriptions/:id/renew route in backend/cmd/api/main.go: attach renew subscription handler with auth middleware | G-008-SUBSCRIPTION | | P2 | Backlog | 008-T177, 008-T030 | yes | | |
| 008-T179 | Create scheduled job for trip archival in backend/cmd/api/archival_job.go: runs daily (cron), queries subscriptions where grace_period_ends_at < now AND status='cancelled', updates all trips where creator_id IN (affected users) set archived=true, logs to console "Archived [count] trips for [count] users" | | | P2 | Backlog | 008-T041, 008-T049 | no | | |
| 008-T180 | Modify POST /subscriptions handler in backend/internal/subscription/handler.go: if user already has subscription with status='cancelled', allow resubscription by reusing existing subscription record (set status='active', clear cancellation fields), if user has active subscription return 409 "Subscription already active" | | | P2 | Backlog | 008-T045 | no | | |
| 008-T181 | Create useCancelSubscription hook in frontend/src/features/subscription/hooks/useCancelSubscription.ts: TanStack Query mutation for DELETE /subscriptions/:id, invalidates subscription and user queries on success | G-008-SUBSCRIPTION-HOOKS | | P2 | Backlog | 008-T031 | yes | | |
| 008-T182 | Create useRenewSubscription hook in frontend/src/features/subscription/hooks/useRenewSubscription.ts: TanStack Query mutation for POST /subscriptions/:id/renew, invalidates subscription, user, and trips queries on success | G-008-SUBSCRIPTION-HOOKS | | P2 | Backlog | 008-T031 | yes | | |
| 008-T183 | Create useUpgrade hook in frontend/src/features/subscription/hooks/useUpgrade.ts: TanStack Query mutation for POST /subscriptions (upgrade Free User to Paid User), updates authStore.user.has_subscription=true on success, invalidates trips query to show restored/enabled trip creation | G-008-SUBSCRIPTION-HOOKS | | P2 | Backlog | 008-T031 | yes | | |
| 008-T184 | Create GracePeriodBanner component in frontend/src/features/subscription/components/GracePeriodBanner.tsx: Banner with type='warning', message "Your subscription has expired. Renew to continue editing your trips. Read-only access until [grace_period_ends_at].", "Renew Subscription" button, uses useRenewSubscription hook | | | P2 | Backlog | 008-T085, 008-T182 | no | | |
| 008-T185 | Create UpgradePrompt component in frontend/src/features/subscription/components/UpgradePrompt.tsx: Banner with type='info', message "Upgrade to create your own trips", "Subscribe Now" button opens payment flow, uses useUpgrade hook | | | P2 | Backlog | 008-T085, 008-T183 | no | | |
| 008-T186 | Create SubscriptionCheckout component in frontend/src/features/subscription/components/SubscriptionCheckout.tsx: form with plan selection (Monthly $9.99), payment stub labeled "[DEMO] Subscription Checkout", "Complete Subscription" button (always succeeds), uses useUpgrade or POST /subscriptions mutation depending on context (register vs upgrade) | | | P2 | Backlog | 008-T056, 008-T183 | no | | |
| 008-T187 | Modify TripDashboard in frontend/src/features/trips/components/TripDashboard.tsx: if user.has_subscription=false AND subscription.status='cancelled' AND grace_period_ends_at NOT NULL, display GracePeriodBanner at top; if user.has_subscription=false AND no subscription, display UpgradePrompt | | | P2 | Backlog | 008-T093, 008-T184, 008-T185 | no | | |
| 008-T188 | Modify TripDetailPage in frontend/src/features/trips/pages/TripDetailPage.tsx: if user is creator AND in grace period (has_subscription=false, grace_period_ends_at>now), disable "Edit Trip" and "Delete Trip" buttons with tooltip "Renew subscription to edit", show GracePeriodBanner | | | P2 | Backlog | 008-T118, 008-T184 | no | | |
| 008-T189 | Create UpgradePage in frontend/src/features/subscription/pages/UpgradePage.tsx: renders SubscriptionCheckout, heading "Upgrade to Paid User", displays plan benefits (create unlimited trips, invite collaborators, AI itinerary generation), uses useUpgrade hook, redirects to /dashboard on success | | | P2 | Backlog | 008-T186, 008-T036 | no | | |
| 008-T190 | Modify Navigation component in frontend/src/components/features/Navigation.tsx: if user.has_subscription=false AND no subscription, show "Upgrade" button in nav; if in grace period, show "Renew" button | | | P2 | Backlog | 008-T122 | no | | |

_Checkpoint: Subscription lifecycle complete with grace period, archival, and restoration (additional feature complete and independently testable)_

#### Phase 10 — Polish & Cross-Cutting Concerns

| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 008-T191 | Add comprehensive error logging to backend: wrap all errors with context using fmt.Errorf, log to CloudWatch Logs with correlation ID, structured JSON format | G-008-POLISH-LOGGING | | P2 | Backlog | 008-T026 | yes | | |
| 008-T192 | Add request logging middleware in backend/internal/security/middleware.go: log all requests with method, path, status, duration, correlation ID to CloudWatch | G-008-POLISH-LOGGING | | P2 | Backlog | 008-T028 | yes | | |
| 008-T193 | Add security headers middleware in backend/internal/security/middleware.go: set Content-Security-Policy, X-Frame-Options, X-Content-Type-Options, Strict-Transport-Security headers | G-008-POLISH-SECURITY | | P2 | Backlog | 008-T027 | yes | | |
| 008-T194 | Run quickstart.md validation: execute all 6 scenarios (Paid User registration, Free User registration, Login, Password reset, Collaboration, Free User upgrade), verify all acceptance criteria met, document any deviations | G-008-POLISH-VALIDATION | | P2 | Backlog | 008-T190 | yes | | |
| 008-T195 | Backend code cleanup: remove unused imports, remove commented code, run gofmt on all files, run golangci-lint and fix all errors | G-008-POLISH-CLEANUP | | P2 | Backlog | 008-T005 | yes | | |
| 008-T196 | Frontend code cleanup: remove unused imports, remove commented code, run Prettier on all files, run ESLint and fix all errors | G-008-POLISH-CLEANUP | | P2 | Backlog | 008-T006 | yes | | |
| 008-T197 | Update README in backend/: document environment variables, database setup, migration commands, how to run server locally, API endpoint documentation | G-008-POLISH-DOCS | | P2 | Backlog | 008-T001 | yes | | |
| 008-T198 | Update README in frontend/: document environment variables, how to run dev server, how to run tests (Vitest, Playwright), component library structure | G-008-POLISH-DOCS | | P2 | Backlog | 008-T003 | yes | | |
| 008-T199 | Create E2E test for US1 in e2e/specs/auth/paid-user-registration.spec.ts: Playwright test navigates to /register, fills form with payment, submits, verifies redirect to /dashboard, creates trip, verifies trip in list | G-008-POLISH-E2E | | P2 | Backlog | 008-T007 | yes | | |
| 008-T200 | Create E2E test for US2 in e2e/specs/auth/free-user-registration.spec.ts: Playwright test for Free User registration, accept invitation, verify single-collaboration limit enforced | G-008-POLISH-E2E | | P2 | Backlog | 008-T007 | yes | | |
| 008-T201 | Create E2E test for US3 in e2e/specs/auth/login-and-dashboard.spec.ts: Playwright test for login with valid credentials, verify dashboard loads, verify rate limiting after 6 failed attempts | G-008-POLISH-E2E | | P2 | Backlog | 008-T007 | yes | | |
| 008-T202 | Create E2E test for US4 in e2e/specs/collaboration/invite-and-suggestions.spec.ts: Playwright test for invite collaborator, submit suggestion, approve suggestion, verify trip updated | G-008-POLISH-E2E | | P2 | Backlog | 008-T007 | yes | | |
| 008-T203 | Create E2E test for US5 in e2e/specs/auth/password-management.spec.ts: Playwright test for password reset request, complete reset via link, login with new password, change password in settings | G-008-POLISH-E2E | | P2 | Backlog | 008-T007 | yes | | |
| 008-T204 | Create E2E accessibility test in e2e/specs/accessibility/wcag-compliance.spec.ts: Playwright test with axe-core integration, verify all pages pass WCAG AA checks for color contrast, keyboard navigation, ARIA attributes | G-008-POLISH-E2E | | P2 | Backlog | 008-T007 | yes | | |
| 008-T205 | Security audit: review all authentication endpoints for vulnerabilities (SQL injection, XSS, CSRF), verify secrets in environment variables, verify no sensitive data in logs, verify rate limiting on all public endpoints | | | P2 | Backlog | 008-T193 | no | | |
| 008-T206 | Performance optimization: add indexes to frequently queried columns (email, user_id, trip_id, status), optimize N+1 queries with joins/batch loading, add caching headers to static assets | | | P2 | Backlog | 008-T011 | no | | |

_Checkpoint: All polish tasks complete, authentication & collaboration UX feature fully implemented and validated_

---

## Critical Path

The minimum sequential chain to reach a fully functional, security-hardened, demo-able MVP:

**REVISED PATH** — spec 005 architecture foundations now GATE all feature implementation:

```
005-T001   Create project structure (backend/, frontend/, e2e/, infra/)
  └─ 005-T013   Create backend config.go with env var loading
       └─ 005-T014–023   Foundational (Docker, CI skeletons, gitignore, linting config)
            └─ [PARALLEL ARCHITECTURE TRACKS]:
                 ├─ 005-T024–041   Backend Architecture (middleware → database → AI client → HTTPServer → domain patterns)
                 ├─ 005-T042–060   Frontend Architecture (design tokens → primitives → composites → routing)
                 └─ 005-T061–109   Infrastructure Architecture (VPC → ECS → RDS → ALB → CloudFront → Secrets → CI/CD)
                      └─ 005-T110–119   Integration Patterns (error correlation, retry logic, observability)
                           └─ [FEATURE IMPLEMENTATION BEGINS]:
                                ├─ 001-T009   Create 8 DB migration files
                                │    └─ 001-T014   Implement User repository
                                │         └─ 001-T016   Implement auth service
                                │              └─ 001-T021   Implement auth HTTP handlers
                                │                   └─ 001-T023   Scaffold Chi router (using 005 patterns)
                                │                        ├─ 002-T012   Register /healthz + wire RequestID/Logger
                                │                        └─ 001-T033   Implement Trip + Day + Activity repository
                                │                             └─ 001-T035   Implement Trip service
                                │                                  └─ 001-T037   Implement Itinerary service (Claude streaming)
                                │                                       ├─ 002-T029   Wire PromptValidator
                                │                                       └─ 001-T038   Implement Trip HTTP handlers
                                │                                            └─ 001-T040   Register trip/conversation routes
                                │                                                 └─ 001-T046   Implement TanStack Query hooks
                                │                                                      └─ 001-T047   Implement Generate page
                                │                                                           └─ 001-T050   E2E: full itinerary generation
                                │
                                ├─ 003-T001   Provision infra (execute Terraform modules from 005)
                                │    └─ 003-T009   Bootstrap Terraform remote state
                                │         └─ 003-T045–047   Initialize secrets + wire into ECS
                                │              └─ 003-T052   Deploy backend to ECS Fargate
                                │                   └─ 003-T053   Deploy frontend to S3+CloudFront
                                │
                                └─ 004-T001–075   Security implementation (JWT, RBAC, validation, logging)
```

**Key Critical Path Changes:**

1. **Architecture-First Gate**: Spec 005 Phase 2 (T013–T023, 11 tasks) is now a HARD BLOCKER. Nothing from 001-004 can begin until foundational configs, Docker, CI skeletons, and linting are in place.

2. **Parallel Architecture Foundation** (after 005-T023):
   - Backend: 18 tasks (005-T024–041) — middleware, database pooling, AI client interface, error handling
   - Frontend: 19 tasks (005-T042–060) — design tokens, primitives, composites, routing
   - Infrastructure: 49 tasks (005-T061–109) — Terraform modules for all AWS resources
   - **Total**: 86 tasks can run in 3 parallel tracks (3 teams, ~2 sprints)

3. **Integration Convergence**: 005-T110–119 (10 tasks) requires all 3 architecture tracks complete. Unifies error correlation, retry logic, and observability across all layers.

4. **Feature Implementation Unlocked**: After 005-T119, feature work from specs 001-004 can proceed using established patterns:
   - Backend domains follow `internal/example/` patterns (model, repository, service, handler)
   - Frontend components follow Atomic Design (primitives → composites → features)
   - Infrastructure provisioning executes pre-built Terraform modules
   - Security integrates with established middleware/validation patterns

**Secondary critical paths** (all now depend on 005 completion):
- **Infrastructure provisioning**: `005-T109 → 003-T001 → 003-T009 → ... → 003-T053` (deploy to AWS)
- **Privacy / GDPR**: `005-T041 → 001-T021 → 002-T032 → 002-T033` (user deletion)
- **Collaboration**: `005-T041 → 001-T050 → 001-T063 → ... → 001-T075` (partner workflow)
- **Accessibility CI gate**: `005-T018 → 002-T002 → 002-T022 → 002-T024` (axe-core + LHCI)
- **Security CI gates**: `005-T017 → 001-T076 → 002-T034 → 002-T036 → 003-T067` (backend)
- **Coverage gates**: `005-T017 → 001-T078 → 002-T044` (backend) and `005-T018 → 001-T079 → 002-T043` (frontend)
- **Infrastructure observability**: `005-T109 → 003-T054 → 003-T055–064` (CloudWatch dashboards)

**Estimated Timeline** (assuming 3-person team):
- **Sprint 1**: Spec 005 Phase 1-2 (setup + foundational, 23 tasks) — 1-2 weeks
- **Sprint 2-3**: Spec 005 Phase 3-5 (parallel architecture, 86 tasks across 3 tracks) — 3-4 weeks
- **Sprint 4**: Spec 005 Phase 6-7 (integration + polish, 22 tasks) — 1 week
- **Sprint 5+**: Feature implementation (specs 001-004, 331 tasks) — 8-12 weeks

**Total time to first deployable MVP**: ~15-20 weeks with architecture-first approach

---

## Sprint Plan

> **Planning Assumptions**: 2-3 full-stack developers, 2-week sprints, ~20-30 tasks/sprint velocity, foundation-first approach.
> **MVP Timeline**: 10 sprints (20 weeks) to deployable MVP covering P1 features only.
> **Post-MVP**: Sprints 11+ for P2-P3 features (collaboration, password reset, design system, observability).
> 
> **GitHub Workflow**: Uses native sub-issues (parent with `- [ ] #N` tasklists), GitHub Projects for organization, and epic labels (`epic:name`) instead of separate epic issues. See [ISSUE-CREATION-GUIDELINES.md](../.github/ISSUE-CREATION-GUIDELINES.md) for details.

### 🏗️ Sprint 1: Architecture Foundation (Weeks 1-2)

**Epic Label**: `epic:architecture-foundation`

**Goal**: Establish project structure, Docker, CI skeletons, and configuration patterns that gate all feature work.

**Scope**: Spec 005 Phases 1-2 (Setup + Foundational)

| Work Items | Task Count | Key Deliverables |
|------------|------------|------------------|
| Project structure setup | 12 | backend/, frontend/, e2e/, infra/ directories with .gitignore |
| Foundational blocking prerequisites | 11 | Docker, linting configs, CI skeletons, config loaders |

**Total**: 23 tasks  
**Risks**: Team onboarding delays, tooling setup issues  
**Dependencies**: None (starting point)

---

### ⚡ Sprint 2: Backend & Frontend Architecture (Weeks 3-4)

**Epic Labels**: `epic:architecture-backend`, `epic:architecture-frontend`, `epic:architecture-infra`

**Goal**: Implement backend and frontend architectural patterns in parallel.

**Scope**: Spec 005 Phase 3-4 (Backend + Frontend Architecture) + Phase 5 Part 1

| Work Items | Task Count | Assignee | Key Deliverables |
|------------|------------|----------|------------------|
| Backend architecture patterns | 18 | Backend Dev | Middleware chain, DB pooling, AI client interface, HTTP server, domain patterns |
| Frontend architecture patterns | 19 | Frontend Dev | Design tokens, primitives, composites, routing patterns |
| Infrastructure modules (Part 1) | 25 | Shared | Terraform modules for VPC, security groups, NAT |

**Total**: 62 tasks  
**Risks**: Parallel track synchronization, pattern adoption  
**Dependencies**: Sprint 1 complete

---

### ☁️ Sprint 3: Infrastructure Architecture (Weeks 5-6)

**Epic Label**: `epic:architecture-infra` (continuation)

**Goal**: Complete Terraform modules for all AWS resources and CI/CD patterns.

**Scope**: Spec 005 Phase 5 Part 2 (Infrastructure completion)

| Work Items | Task Count | Key Deliverables |
|------------|------------|------------------|
| Infrastructure modules (Part 2) | 24 | Terraform modules for ECS, RDS, ALB, CloudFront, IAM, Secrets Manager |

**Total**: 24 tasks  
**Risks**: AWS account limits, Terraform state management  
**Dependencies**: Sprint 2 complete

---

### 🔗 Sprint 4: Integration & Observability (Weeks 7-8)

**Epic Label**: `epic:integration-observability`

**Goal**: Unify architecture layers with error correlation, retry logic, and observability primitives.

**Scope**: Spec 005 Phases 6-7 + Spec 002 Phase 2

| Work Items | Task Count | Key Deliverables |
|------------|------------|------------------|
| Integration patterns | 10 | Error correlation across layers, retry strategies with exponential backoff |
| Architecture polish | 12 | Documentation, example domain, validation |
| Observability core | 7 | RequestID → Logger → /healthz middleware chain |

**Total**: 29 tasks  
**Risks**: Cross-layer integration complexity  
**Dependencies**: Sprint 3 complete

---

### 🔐 Sprint 5: Authentication & Security Foundation (Weeks 9-10)

**Epic Label**: `epic:auth-security`

**Goal**: Implement secure authentication, JWT handling, and security middleware.

**Scope**: Spec 004 Phase 2 + Spec 008 Phase 2

| Work Items | Task Count | Key Deliverables |
|------------|------------|------------------|
| Security foundations | 36 | JWT RS256 multi-key rotation, refresh tokens, RBAC middleware, security event logging |
| Auth infrastructure | 28 | bcrypt password hashing (cost 12), rate limiting, validation, Chi router middleware chain |

**Total**: 64 tasks  
**Risks**: JWT multi-key rotation complexity, rate limiter tuning  
**Dependencies**: Sprint 4 complete

---

### 📝 Sprint 6: Core Data Layer & Repositories (Weeks 11-12)

**Epic Label**: `epic:data-layer`

**Goal**: Implement database migrations, repositories, and core domain models.

**Scope**: Spec 001 Phase 2 (Data Layer) + Spec 008 Phase 3 Part 1

| Work Items | Task Count | Key Deliverables |
|------------|------------|------------------|
| Database migrations | 7 | 8 migration files (users, plans, subscriptions, trips, days, activities, collaborators, suggestions) |
| Backend repositories | 15 | User, Subscription, Trip repositories with optimistic locking; Payment provider stub |

**Total**: 22 tasks  
**Risks**: Migration ordering, foreign key constraints  
**Dependencies**: Sprint 5 complete

---

### 🎨 Sprint 7: Frontend Shell & Components (Weeks 13-14)

**Epic Label**: `epic:frontend-shell`

**Goal**: Build React app shell, authentication pages, and component library primitives.

**Scope**: Spec 001 Phase 2 (Frontend) + Spec 008 Phase 3 Part 2

| Work Items | Task Count | Key Deliverables |
|------------|------------|------------------|
| Frontend shell | 9 | Design tokens, API client, Zustand auth store, routing, primitives (Button, Input, Label) |
| Authentication UI | 19 | Register/Login/Dashboard pages, TripCard, TripDashboard, TripDetail with TanStack Query hooks |

**Total**: 28 tasks  
**Risks**: Component library scope creep, TanStack Query learning curve  
**Dependencies**: Sprint 6 complete

---

### 🚀 Sprint 8: MVP User Stories (Trip Generation) (Weeks 15-16)

**Epic Label**: `epic:trip-generation`

**Goal**: Implement core trip generation flow with AI integration (stubbed for MVP).

**Scope**: Spec 001 Phase 3 + Spec 008 Phase 5

| Work Items | Task Count | Key Deliverables |
|------------|------------|------------------|
| Trip generation backend | 18 | Trip/Conversation/Itinerary services, AI Claude streaming (stub: 3-day Paris itinerary), HTTP handlers |
| Login & trip management | 26 | Login/logout flow, trips list/detail/update/delete handlers, trip CRUD UI pages |

**Total**: 44 tasks  
**Risks**: SSE streaming complexity, AI stub maintainability  
**Dependencies**: Sprint 7 complete

---

### 🔒 Sprint 9: Security Hardening & NFR Gates (Weeks 17-18)

**Epic Label**: `epic:security-hardening`

**Goal**: Wire security validators and implement NFR CI gates (accessibility, performance, security).

**Scope**: Spec 002 Phases 3-5 + Spec 004 Phase 3

| Work Items | Task Count | Key Deliverables |
|------------|------------|------------------|
| Performance validation | 4 | k6 load tests (500 VUs, p95 ≤ 500ms), API latency scenarios |
| Accessibility gates | 8 | axe-core + LHCI, privacy policy page, @accessibility E2E tag |
| Security validation | 12 | Prompt injection validator (YAML deny-list), output sanitizer (bluemonday), GDPR user deletion, CI gates (gosec, gitleaks) |
| Security backend implementation | 32 | Wire PromptValidator into handlers, OutputSanitizer into services, security logging, multi-key JWT storage |

**Total**: 56 tasks  
**Risks**: Load test threshold tuning, CI gate false positives  
**Dependencies**: Sprint 8 complete

---

### ☁️ Sprint 10: Infrastructure Deployment & MVP (Weeks 19-20)

**Epic Label**: `epic:deployment`

**Goal**: Deploy to AWS staging, validate end-to-end flow, achieve deployable MVP.

**Scope**: Spec 003 Phases 3-5 + E2E validation

| Work Items | Task Count | Key Deliverables |
|------------|------------|------------------|
| AWS infrastructure provisioning | 23 | Terraform apply (VPC, ALB, RDS, ECS, CloudFront), wire all modules in root main.tf |
| Secrets & configuration | 6 | Populate AWS Secrets Manager (DB credentials, Anthropic API key, JWT signing keys), wire into ECS task definitions |
| CI/CD pipeline deployment | 14 | GitHub Actions workflows with OIDC, backend ECR build/push, ECS deployment, frontend S3+CloudFront deploy, smoke tests |
| E2E validation | 1 | Full user journey: register → subscribe → generate itinerary → verify Day 1 activities |

**Total**: 44 tasks  
**Risks**: AWS service limits, IAM permission issues, Terraform state conflicts, ECS deployment failures  
**Dependencies**: Sprint 9 complete

---

### 📊 Sprint Summary Statistics

| Metric | Value |
|--------|-------|
| **MVP sprints** | 10 sprints (20 weeks) |
| **Total MVP tasks** | 390 tasks (52% of project) |
| **Average velocity** | 39 tasks/sprint |
| **Parallelization rate** | 54% of tasks can run concurrently |
| **GitHub Issues (estimated)** | ~270 parent + sub-issues (vs 745 if ungrouped) |
| **Epic labels** | 11 epics across all sprints (see each sprint header) |
| **Team size assumption** | 2-3 full-stack developers |

**GitHub Organization**:
- **Epic tracking**: Filter by `epic:architecture-foundation`, `epic:auth-security`, etc. in GitHub Projects
- **Issue hierarchy**: Parent issues with `- [ ] #N` tasklists auto-create sub-issue relationships
- **Sprint tracking**: Filter by `sprint:1`, `sprint:2`, etc. in GitHub Projects Sprint Board view
- **Workflow**: Create sub-issues first (to get numbers), then parent with tasklist references

### 🎯 Post-MVP Roadmap (Sprints 11+)

**Sprints 11-12: Collaboration Features (P2)**
- Spec 008 Phase 4: Free user registration + collaboration acceptance (26 tasks)
- Spec 008 Phase 6: Suggestion workflow (17 tasks)
- Spec 001 Phase 6: Group collaboration advanced (13 tasks)
- **Total**: ~56 tasks

**Sprints 13-14: Enhanced UX & Security (P2)**
- Spec 008 Phase 7: Password reset + account security (21 tasks)
- Spec 008 Phase 9: Subscription lifecycle (16 tasks)
- Spec 004 Phases 4-5: Frontend security + QA (35 tasks)
- **Total**: ~72 tasks

**Sprints 15+: Polish & Observability (P2-P3)**
- Spec 002 Phases 6-8: Observability validation + maintainability (8 tasks)
- Spec 003 Phases 6-8: Cost monitoring + infra observability (36 tasks)
- Spec 006-007: Documentation + API standards validation (125 tasks)
- Spec 008 Phases 8-10: Design system + polish (40 tasks)
- **Total**: ~209 tasks

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

**Run date**: 2026-07-06

### Specs Discovered

| Spec | Folder | Has tasks.md | Status |
|------|--------|-------------|--------|
| 001 | `specs/001-product-vision-scope/` | ✅ yes | Active |
| 002 | `specs/002-nfr-system-constraints/` | ✅ yes | Active |
| 003 | `specs/003-cloud-env-strategy/` | ✅ yes | Active |
| 004 | `specs/004-security-auth-model/` | ✅ yes | Active |
| 005 | `specs/005-system-architecture/` | ✅ yes | Active |
| 006 | `specs/006-core-domain-model/` | ✅ yes | Active |
| 007 | `specs/007-api-design-standards/` | ✅ yes | **NEW** |

### Change Counts

| Operation | Count |
|-----------|-------|
| **ADD** | 70 (spec 007: all 70 tasks) |
| **UPDATE** | 0 |
| **UNCHANGED** | 462 (001: 81 tasks, 002: 46 tasks, 003: 72 tasks, 004: 132 tasks, 005: 131 tasks, 006: 77 tasks) |
| **REMOVE** | 0 |
| **ARCHIVED** | 0 |
| **Structural** | Added 12 new groups for spec 007 tasks |

**Total tasks in roadmap**: 532 (was 462, added 70)  
**Grouped tasks**: 370 (69.5% of total, forming 107 work items)  
**Standalone tasks**: 162 (30.5% of total)

### New Grouping Summary (Spec 007)

| Group ID | Tasks | Description |
|----------|-------|-------------|
| G-API-US1-ACCESSIBILITY | 3 | Standards doc accessibility verification (007-T010–012) |
| G-API-US1-NAMING | 2 | Resource naming & URL structure verification (007-T015–016) |
| G-API-US1-FORMAT | 2 | Request/response format verification (007-T018–019) |
| G-API-US2-ERROR-HANDLING | 2 | Frontend error handling examples (007-T023–024) |
| G-API-US2-PAGINATION | 2 | Frontend pagination examples (007-T026–027) |
| G-API-US2-FILTERING-SORTING | 2 | Frontend filtering/sorting examples (007-T029–030) |
| G-API-US4-EXTERNAL-DOCS | 2 | External API consumer documentation (007-T042–043) |
| G-API-US4-PATTERN-RECOGNITION | 2 | Pattern transfer documentation (007-T045–046) |
| G-API-POLISH-TEST-HELPERS | 1 | Integration test helper functions (007-T051) |
| G-API-POLISH-INTEGRATION-TESTS | 8 | Standards validation integration tests (007-T052–059) |
| G-API-POLISH-LINTER | 3 | Linter research and configuration (007-T060–062) |
| G-API-POLISH-DOCUMENTATION | 3 | Final documentation updates (007-T063–065) |
| G-API-POLISH-VALIDATION | 3 | Final validation execution (007-T066–068) |

### Cross-Spec Dependencies (Spec 007)

Spec 007 establishes API design conventions that all backend endpoints must follow:

| Type | Description |
|------|-------------|
| **Standards foundation** | 007-T004–009 (Phase 2 Foundational) promotes API standards doc and audits existing endpoints — BLOCKS all user story work in spec 007 |
| **Backend endpoint compliance** | 001/004 API endpoints must follow standards defined in docs/api-design-standards.md (resource naming, error format, pagination) |
| **Frontend API client patterns** | 001 frontend API client code will use predictable patterns documented in 007-T023–033 (error handling, pagination, filtering) |
| **Integration test validation** | 007-T050–059 integration tests will validate 001/004 endpoints comply with standards (error format, status codes, timestamps, field naming) |
| **PR template extension** | 007-T035–036 extends .github/pull_request_template.md with API standards compliance checklist for all backend PR reviews |
| **Documentation reference** | 007-T063–064 links standards doc from README and coding-guidelines.md for discoverability |

**No blocking dependencies** — spec 007 is pure documentation/standards definition that runs in parallel with feature implementation.

### Human Attention Required

| Item | Detail |
|------|--------|
| **New spec review** | Spec 007 (API Design Standards and Conventions) added with 70 tasks. Review grouping and priorities. |
| **Priority review** | Spec 007 priorities: P1 (Phases 1-3, 22 tasks — MVP standards doc), P2 (Phases 4-5 + most of Phase 7, 32 tasks — frontend examples + code review tools + tests), P3 (Phase 6 + Phase 7 validation, 16 tasks — external consumer docs + final validation). Phase 2 is CRITICAL BLOCKER for all other spec 007 work. |
| **Status review** | All 70 new tasks default to `Backlog`. Mark Phase 1-2 tasks `Ready` to begin standards promotion workflow. |
| **Issue column** | All 70 new spec 007 `Issue` fields are empty. **46 grouped tasks will form 13 issues** (with checklists); **24 standalone tasks will form 24 issues**. Total NEW issues: **37 issues** when `/sync-issues` runs. |
| **Sprint column** | All spec 007 `Sprint` fields are empty. Populate during sprint planning. Recommend: Phase 1-2 (setup + foundational promotion) in first sprint, Phases 3-5 (user stories 1-3) in sprint 2, Phases 6-7 (user story 4 + polish) in sprint 3. |
| **Parallel execution** | 38 of 70 tasks (54%) marked parallelizable in source — can run concurrently. User Stories 1-4 can proceed in parallel after Phase 2 completes. Integration tests (Phase 7) can run parallel to US2-US4. |
| **Documentation-only feature** | Spec 007 is pure documentation/standards definition — no code implementation. Focuses on promoting standards doc to docs/, adding PR checklist, creating external consumer guide, and writing validation integration tests. |
| **Foundation promotion complete** | Per conversation summary, spec 007 Phase 2 foundational tasks (007-T004–006) are ALREADY COMPLETE (docs/api-design-standards.md created, copilot-instructions.md updated). Mark these 3 tasks `Done` and update Status column. |

### Critical Path Changes

- **API standards now available**: Spec 007 Phase 2 (Foundational, tasks 007-T004–009) has been COMPLETED per conversation summary. The authoritative API standards document (docs/api-design-standards.md) is now the project-wide reference for all endpoint design.
- **Backend endpoint compliance**: All backend endpoints defined in specs 001 and 004 must now follow conventions in docs/api-design-standards.md (resource naming, URL structure, request/response formats, error handling, pagination, filtering, sorting, rate limiting).
- **No blocking impact on feature work**: Spec 007 is documentation-only and runs in parallel with implementation. Backend developers reference the standards doc when designing endpoints. Code reviewers use the PR checklist (007-T035–036) to verify compliance.
- **Integration test enforcement**: Once spec 007 Phase 7 completes (007-T050–059), automated integration tests will validate endpoint compliance, reducing manual review burden.
- **Revised critical path** (no change to feature implementation sequence):
  ```
  005-T001 (architecture setup)
    → 005-T013–023 (foundational BLOCKER)
      → [PARALLEL]:
         - 005-T024–041 (backend architecture)
         - 005-T042–060 (frontend architecture)
         - 005-T061–109 (infrastructure architecture)
      → 005-T110–119 (integration)
        → [Feature implementation from 001-004 begins here]
           - Backend endpoints follow docs/api-design-standards.md conventions
           - 007-T050–059 integration tests validate compliance
  ```

### Next Steps

1. **Review spec 007 grouping**: The 13 groups bundle 46 tasks into cohesive work items following user story boundaries (US1 standards accessibility, US2 frontend examples, US4 external docs, Polish integration tests). If any grouping doesn't align with team ownership, clear the `Group` value to split into standalone items.
2. **Mark completed tasks**: Per conversation summary, spec 007 Phase 2 foundational tasks (007-T004, 007-T005, 007-T006) are ALREADY COMPLETE. Update their `Status` column to `Done` and record completion date in `Notes`.
3. **Sprint planning**: Populate `Sprint` column for spec 007. Recommend sprint breakdown:
   - **Sprint 1** (DONE): Phase 1-2 (T001–T009) — setup + standards doc promotion ✅ COMPLETE
   - **Sprint 2**: Phase 3 (T010–T022) — User Story 1 (backend developer standards access, 13 tasks)
   - **Sprint 3**: Phase 4-5 (T023–T041) — User Story 2-3 (frontend integration + code review tooling, 19 tasks)
   - **Sprint 4**: Phase 6-7 (T042–T070) — User Story 4 + Polish (external docs + automated tests, 29 tasks)
4. **Documentation-first decision**: Spec 007 establishes standards that inform 001/004 endpoint design. Recommend completing 007 Phase 3 (US1, 13 tasks) BEFORE implementing backend endpoints in 001 to ensure consistency from day one.
5. **Run `/sync-issues`**: This will create **37 new GitHub issues** for spec 007 (13 grouped + 24 standalone). Grouped issues will have checklists with member tasks. Use labels: `documentation`, `standards`, `P1`/`P2`/`P3`.
6. **Coordinate cross-spec work**: Spec 007 standards inform 001/004 endpoint design. Ensure backend teams reference docs/api-design-standards.md when defining handler routes, error responses, and pagination logic. Frontend teams use 007-T023–033 patterns for generic API client code.
7. **Re-run `/build-roadmap`**: When spec 007 `tasks.md` is modified, a new spec is added, or existing specs are updated — the command preserves all `Group`, `Sprint`, `Priority`, `Status`, and `Issue` fields.

---

**Total project task count**: **532 tasks** across 7 foundation specs  
**Grouped work items**: **107 issues** (combining 370 tasks)  
**Standalone work items**: **162 issues**  
**Total GitHub issues when synced**: **269 issues**

---

_End of reconciliation report. All human-owned fields preserved. Ready for sprint planning and `/sync-issues` execution._
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
