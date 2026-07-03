# Project Roadmap

> Generated and reconciled by /build-roadmap. Source of truth for task existence,
> titles, and dependencies is `specs/*/tasks.md`. Priority, Status, Phase, Issue, and
> Notes are human-owned and preserved across runs. Do not hand-edit the stable IDs.

**Last reconciled**: 2026-07-02

## Legend

- **Priority**: P1 (critical) | P2 | P3 | TBD
- **Status**: Backlog | Ready | In Progress | In Review | Done
- **Parallel**: yes = task carries `[P]` flag in source (can run concurrently with peers)
- **Issue**: link to the tracker issue once created (empty = not yet created)

---

## Foundation Phase

<!-- Ordered by cross-spec dependency: vision/scope → NFRs → cloud/IaC → security → architecture → domain -->

### Spec 001 — Product Vision and Scope &nbsp; `specs/001-product-vision-scope/tasks.md`

#### Phase 1 — Setup

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 001-T001 | Create monorepo root directory structure (`backend/`, `frontend/`, `e2e/`, `.github/workflows/`, `docs/`) | P1 | Backlog | - | no | | |
| 001-T002 | Initialize Go 1.24 module with backend dependencies (chi, pgx/v5, jwt/v5, anthropic-sdk-go, goose/v3, testify) | P1 | Backlog | 001-T001 | no | | |
| 001-T003 | Initialize React 19 + TypeScript Vite project with frontend dependencies | P1 | Backlog | 001-T001 | yes | | |
| 001-T004 | Configure `golangci-lint` with required linters (errcheck, govet, staticcheck, revive, gosec) | P1 | Backlog | 001-T001 | yes | | |
| 001-T005 | Configure Prettier and ESLint strict mode with TypeScript plugin | P1 | Backlog | 001-T001 | yes | | |
| 001-T006 | Create `docker-compose.yml` with PostgreSQL 16 service, named volume, and health check | P1 | Backlog | 001-T001 | no | | |
| 001-T007 | Create `.env.example` with all required environment variables | P1 | Backlog | 001-T001 | yes | | |
| 001-T008 | Create `Makefile` with lint, test, migrate-up, migrate-down, build, dev targets | P1 | Backlog | 001-T001 | no | | |

#### Phase 2 — Foundational (Backend Data Layer)

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 001-T009 | Create 8 database migration SQL files (users, plans, subscriptions, trips, destinations, days+activities, collaborators, suggestions+conversation) | P1 | Backlog | 001-T001 | no | | |
| 001-T010 | Implement config struct with env var loading and fail-fast validation | P1 | Backlog | 001-T001 | no | | |
| 001-T011 | Implement pgxpool connection initialization with context-aware open/close | P1 | Backlog | 001-T001 | yes | | |
| 001-T012 | Define `PaymentProvider` interface with CreateSubscription, CancelSubscription, GetSubscription | P1 | Backlog | 001-T001 | yes | | |
| 001-T013 | Implement `StubProvider` satisfying `PaymentProvider` (always succeeds, logs [STUB]) | P1 | Backlog | 001-T012 | no | | |
| 001-T014 | Implement User repository (Create, FindByEmail, FindByID, UpdateSubscription) | P1 | Backlog | 001-T009 | yes | | |
| 001-T015 | Implement Plan and Subscription repositories (FindPlanByName, CreateSubscription, FindSubscriptionByUser) | P1 | Backlog | 001-T009 | yes | | |

#### Phase 2 — Foundational (Backend Service & Handler Layer)

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 001-T016 | Implement auth service (Register with bcrypt, Login with JWT, Me) | P1 | Backlog | 001-T014 | no | | |
| 001-T017 | Implement subscription service (ListPlans, InitiateCheckout, ConfirmCheckout, GetCurrent) | P1 | Backlog | 001-T015 | no | | |
| 001-T018 | Implement JWT auth middleware (validate HTTP-only cookie, attach User to context, 401 on failure) | P1 | Backlog | 001-T001 | yes | | |
| 001-T019 | Implement role-guard middleware factory (`RequireRole("admin")` / `RequireRole("partner")`) | P1 | Backlog | 001-T001 | yes | | |
| 001-T020 | Implement JSON response helpers (Success, Error) | P1 | Backlog | 001-T001 | yes | | |
| 001-T021 | Implement auth HTTP handlers (POST /register, /login, /logout, GET /me) | P1 | Backlog | 001-T016 | no | | |
| 001-T022 | Implement subscription HTTP handlers (GET /plans, POST /checkout, /confirm, GET /current) | P1 | Backlog | 001-T017 | no | | |
| 001-T023 | Scaffold Chi router: mount auth and subscription routes, apply CORS, request-ID, and logging middleware | P1 | Backlog | 001-T021, 001-T022 | no | | |

#### Phase 2 — Foundational (Frontend Shell & Auth)

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 001-T024 | Create CSS custom property design tokens (colors, spacing, typography, border-radius) | P1 | Backlog | 001-T001 | yes | | |
| 001-T025 | Create typed API client base (`apiFetch` wrapper with credentials: include, JSON parsing) | P1 | Backlog | 001-T001 | yes | | |
| 001-T026 | Create auth Zustand store (user, setUser, clearUser; persist to sessionStorage) | P1 | Backlog | 001-T001 | yes | | |
| 001-T027 | Create `ProtectedRoute` (redirect to /login) and `GuestRoute` (redirect to /dashboard) | P1 | Backlog | 001-T001 | yes | | |
| 001-T028 | Create primitive Button, Input, Label, Badge components using design tokens | P1 | Backlog | 001-T024 | yes | | |
| 001-T029 | Implement Register page (email + password form, calls POST /auth/register, redirects to checkout) | P1 | Backlog | 001-T024, 001-T025 | no | | |
| 001-T030 | Implement stub Checkout page (plan summary, POST /subscription/checkout + /confirm, redirect dashboard) | P1 | Backlog | 001-T028, 001-T029 | no | | |
| 001-T031 | Implement Login page (calls POST /auth/login, sets user in store, redirects to dashboard) | P1 | Backlog | 001-T028, 001-T030 | no | | |
| 001-T032 | Wire React Router v7 with all routes (/, /login, /register, /subscribe, /dashboard, /trips/:id, /generate) | P1 | Backlog | 001-T029, 001-T030, 001-T031 | no | | |

#### Phase 3 — User Story 1: First-Time Traveler Plans a Trip (Priority: P1) 🎯 MVP

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 001-T033 | Implement Trip + Destination + Day + Activity repository (CreateTrip, UpsertDay, UpsertActivity) | P1 | Backlog | 001-T009 | yes | | |
| 001-T034 | Implement ConversationSession + ConversationMessage repository (CreateSession, AppendMessage, ListMessages) | P1 | Backlog | 001-T009 | yes | | |
| 001-T035 | Implement Trip service (Create, List, Get, Update, Delete; admin-only writes) | P1 | Backlog | 001-T033 | no | | |
| 001-T036 | Implement Conversation service (SendMessage, GetHistory; detects itinerary_ready) | P1 | Backlog | 001-T034 | no | | |
| 001-T037 | Implement Itinerary service (build Claude prompt, streaming, parse tool-use response, persist to DB) | P1 | Backlog | 001-T035 | no | | |
| 001-T038 | Implement Trip HTTP handlers (GET/POST/PUT/DELETE /trips; RequireRole admin on writes) | P1 | Backlog | 001-T035 | no | | |
| 001-T039 | Implement Conversation HTTP handlers (POST/GET /trips/:id/conversation; SSE streaming) | P1 | Backlog | 001-T036 | no | | |
| 001-T040 | Register trip and conversation routes in Chi router | P1 | Backlog | 001-T038, 001-T039 | no | | |
| 001-T041 | Create `TripCard` composite (destination names, duration, status badge; loading/empty states) | P1 | Backlog | 001-T028 | yes | | |
| 001-T042 | Create `ActivityItem` composite (Lucide type icon, title, description, AI-generated indicator) | P1 | Backlog | 001-T028 | yes | | |
| 001-T043 | Create `DaySection` composite (day number, label, ordered ActivityItem list; empty state) | P1 | Backlog | 001-T028 | yes | | |
| 001-T044 | Create `ConversationPanel` feature (message thread, text input, SSE stream rendering, loading indicator) | P1 | Backlog | 001-T028 | yes | | |
| 001-T045 | Create `ItineraryView` feature (scrollable DaySection list; loading skeleton, error, empty states) | P1 | Backlog | 001-T028 | yes | | |
| 001-T046 | Implement trips and conversation API service functions with TanStack Query hooks | P1 | Backlog | 001-T025 | no | | |
| 001-T047 | Implement Generate page (new trip form → ConversationPanel → ItineraryView on itinerary_ready) | P1 | Backlog | 001-T044, 001-T045, 001-T046 | no | | |
| 001-T048 | Implement Trip detail page (ItineraryView, action bar: edit title, delete trip) | P1 | Backlog | 001-T045, 001-T046 | no | | |
| 001-T049 | Implement Dashboard page (TripCard grid, "New Trip" CTA, empty state illustration) | P1 | Backlog | 001-T041, 001-T046 | no | | |
| 001-T050 | Add Playwright E2E spec: register → subscribe → generate itinerary via conversation → verify Day 1 | P1 | Backlog | 001-T047, 001-T049, 001-T040 | no | | |

#### Phase 4 — User Story 2: Experienced Traveler, Off-the-Beaten-Path (Priority: P2)

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 001-T051 | Create `TravelStyleSelector` composite (multi-select chip group; keyboard accessible) | P2 | Backlog | 001-T028 | yes | | |
| 001-T052 | Integrate `TravelStyleSelector` into Generate page; pass selected styles on trip creation | P2 | Backlog | 001-T051 | no | | |
| 001-T053 | Update Trip service to persist TripTravelStyle join records on create/update | P2 | Backlog | 001-T035 | no | | |
| 001-T054 | Update Itinerary service: inject travel styles into Claude prompt; non-mainstream recs per day | P2 | Backlog | 001-T037 | no | | |
| 001-T055 | Update Trip HTTP handlers to accept travel_styles[] in POST/PUT; return in GET | P2 | Backlog | 001-T038 | no | | |
| 001-T056 | Update ItineraryView to display travel style badges at trip header level | P2 | Backlog | 001-T045 | yes | | |
| 001-T057 | Add E2E spec: gastronomy style → verify ≥ 1 food activity per day | P2 | Backlog | 001-T050 | no | | |

#### Phase 5 — User Story 3: Planner Enriches Existing Plan (Priority: P2)

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 001-T058 | Update Itinerary service: detect anchor mode from enumerated place lists; inject anchors into Claude | P2 | Backlog | 001-T037 | no | | |
| 001-T059 | Update Conversation service: return extracted anchor places for user confirmation | P2 | Backlog | 001-T036 | no | | |
| 001-T060 | Update ConversationPanel to render anchor places confirmation card | P2 | Backlog | 001-T044 | yes | | |
| 001-T061 | Update Generate page to handle anchor confirmation step before final generation | P2 | Backlog | 001-T047 | no | | |
| 001-T062 | Add E2E spec: provide 5 anchor places → confirm → assert all 5 appear in itinerary | P2 | Backlog | 001-T050 | no | | |

#### Phase 6 — User Story 4: Group Collaboration on Shared Itinerary (Priority: P3)

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 001-T063 | Extend Trip repository with collaborator ops (CreateCollaborator, AcceptCollaborator, CountActive) | P3 | Backlog | 001-T033 | yes | | |
| 001-T064 | Implement Suggestion repository (CreateSuggestion, ListByTrip, UpdateStatus) | P3 | Backlog | 001-T009 | yes | | |
| 001-T065 | Extend Trip service: InviteCollaborator (enforce basic plan limit), AcceptInvitation, Remove | P3 | Backlog | 001-T035, 001-T063 | no | | |
| 001-T066 | Implement Suggestion service (Submit, Approve applies Activity update, Reject preserves history) | P3 | Backlog | 001-T064 | no | | |
| 001-T067 | Implement Collaborator HTTP handlers (POST/DELETE /trips/:id/collaborators; admin only) | P3 | Backlog | 001-T065 | no | | |
| 001-T068 | Implement Suggestion HTTP handlers (GET/POST /suggestions, PATCH approve/reject) | P3 | Backlog | 001-T066 | no | | |
| 001-T069 | Register collaborator and suggestion routes in Chi router | P3 | Backlog | 001-T067, 001-T068 | no | | |
| 001-T070 | Create `SuggestionBubble` composite (author, content, status badge, approve/reject buttons) | P3 | Backlog | 001-T028 | yes | | |
| 001-T071 | Create `SuggestionQueue` feature (SuggestionBubble list, status filter, loading/empty states) | P3 | Backlog | 001-T070 | yes | | |
| 001-T072 | Create `CollaboratorInvite` composite (email input, invite button, partner badge, plan limit warning) | P3 | Backlog | 001-T028 | yes | | |
| 001-T073 | Implement suggestion/collaborator API service functions with TanStack Query hooks | P3 | Backlog | 001-T046 | no | | |
| 001-T074 | Integrate SuggestionQueue and CollaboratorInvite into Trip detail page | P3 | Backlog | 001-T071, 001-T072, 001-T073 | no | | |
| 001-T075 | Add E2E spec: admin invites partner → suggestion submitted → admin approves/rejects flow | P3 | Backlog | 001-T050 | no | | |

#### Phase 7 — Polish & Cross-Cutting Concerns (Priority: P2)

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 001-T076 | Add backend lint CI workflow (golangci-lint on push and PR) | P2 | Backlog | 001-T001 | yes | | |
| 001-T077 | Add frontend lint + type-check CI workflow (ESLint + tsc --noEmit) | P2 | Backlog | 001-T001 | yes | | |
| 001-T078 | Add backend test CI workflow (go test -race against PostgreSQL service container) | P2 | Backlog | 001-T001 | yes | | |
| 001-T079 | Add frontend test CI workflow (Vitest) | P2 | Backlog | 001-T001 | yes | | |
| 001-T080 | Create production-ready backend Dockerfile (multi-stage, non-root user, migrations on start) | P2 | Backlog | 001-T023 | no | | |
| 001-T081 | Write README.md (overview, prerequisites, setup commands, architecture reference, workflow) | P2 | Backlog | 001-T001 | no | | |

---

### Spec 002 — Non-Functional Requirements and System Constraints &nbsp; `specs/002-nfr-system-constraints/tasks.md`

> Cross-spec note: Tasks in phases 5 and 7 extend CI workflow files created by 001-T076–T079 and
> depend on application code from spec 001. See the Dependencies section in `specs/002-nfr-system-constraints/tasks.md`.

#### Phase 1 — Setup (NFR Dependencies and Config)

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 002-T001 | Add `github.com/microcosm-cc/bluemonday` HTML sanitisation dependency to Go module | P1 | Backlog | - | no | | |
| 002-T002 | Add `@axe-core/playwright` and `@lhci/cli` as frontend dev dependencies | P1 | Backlog | - | yes | | |
| 002-T003 | Create `backend/config/prompt-rules.yml` with 5 seed deny-list rules (instruction-override, role-switching, jailbreak-prefix, etc.) | P1 | Backlog | - | yes | | |
| 002-T004 | Create `lighthouserc.yml` with LHCI assertion thresholds (a11y ≥ 0.9, LCP ≤ 2500 ms, CLS ≤ 0.1, INP ≤ 200 ms) | P1 | Backlog | - | yes | | |
| 002-T005 | Create `.gitleaks.toml` secret-scanning configuration (scan all committed files, exclude test fixtures) | P1 | Backlog | - | yes | | |

#### Phase 2 — Foundational: Observability Core (NFR-OBS-001–003)

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 002-T006 | Implement `RequestID` Chi middleware (generate UUID v4 if absent; propagate; set X-Request-ID on response) | P1 | Backlog | - | no | | |
| 002-T007 | Write unit tests for `RequestID` middleware (generates UUID, propagates existing, sets response header) | P1 | Backlog | 002-T006 | no | | |
| 002-T008 | Implement `Logger` Chi middleware using `log/slog` JSON handler (emit StructuredLogEntry per request) | P1 | Backlog | - | yes | | |
| 002-T009 | Write unit tests for `Logger` middleware (all fields present, duration_ms ≥ 0, user_id absent on unauth) | P1 | Backlog | 002-T008 | no | | |
| 002-T010 | Implement `GET /healthz` handler returning HealthCheckResponse JSON (status, version, uptime_seconds) | P1 | Backlog | - | yes | | |
| 002-T011 | Write unit tests for `/healthz` handler (200 OK, schema valid, status is "ok", uptime ≥ 0) | P1 | Backlog | 002-T010 | no | | |
| 002-T012 | Register `/healthz` and wire `RequestID` → `Logger` middleware chain globally in Chi router | P1 | Backlog | 002-T006, 002-T008, 002-T010, 001-T023 | no | | |

#### Phase 3 — User Story 1: Engineering Team Verifies Performance Under Load (Priority: P1) 🎯

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 002-T013 | Create k6 baseline load test (ramp to 500 VUs, sustain 10 min; assert p95 ≤ 500 ms and error rate < 1%) | P1 | Backlog | 002-T012 | yes | | |
| 002-T014 | Create k6 API latency scenario (constant 100 RPS; assert p95 ≤ 500 ms per non-AI endpoint) | P1 | Backlog | 002-T012 | yes | | |
| 002-T015 | Create `load-test.yml` GitHub Actions workflow (manual `workflow_dispatch`; runs k6 against staging URL) | P1 | Backlog | 002-T013 | no | | |
| 002-T016 | Write integration test asserting `/healthz` responds in ≤ 100 ms for 100 sequential calls | P1 | Backlog | 002-T010 | no | | |

#### Phase 4 — User Story 2: Accessibility Reviewer Confirms WCAG 2.1 AA (Priority: P1)

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 002-T017 | Implement `PrivacyPolicyPage` static component (WCAG-compliant heading, section structure, landmarks) | P1 | Backlog | 002-T002 | yes | | |
| 002-T018 | Write Vitest unit test for `PrivacyPolicyPage` (renders heading, ≥ 3 sections, no dangerouslySetInnerHTML) | P1 | Backlog | 002-T017 | yes | | |
| 002-T019 | Implement `PrivacyPolicyLink` atom component (accessible `<a>` linking to /privacy-policy) | P1 | Backlog | 002-T002 | yes | | |
| 002-T020 | Write Vitest unit test for `PrivacyPolicyLink` (correct href, accessible text present) | P1 | Backlog | 002-T019 | yes | | |
| 002-T021 | Add `/privacy-policy` route to React Router; add `PrivacyPolicyLink` to registration form footer | P1 | Backlog | 002-T017, 002-T019 | no | | |
| 002-T022 | Create accessibility E2E helper `checkPageA11y(page)` wrapping `@axe-core/playwright` | P1 | Backlog | 002-T002 | yes | | |
| 002-T023 | Add `checkPageA11y(page)` call to every existing Playwright E2E spec; tag with `@accessibility` | P1 | Backlog | 002-T022 | no | | |
| 002-T024 | Create `accessibility.yml` GitHub Actions workflow (axe-core Playwright run + lhci autorun; PR gate) | P1 | Backlog | 002-T022, 002-T004 | no | | |

#### Phase 5 — User Story 3: Security Reviewer Confirms Security Posture (Priority: P1)

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 002-T025 | Implement `PromptValidator` (load rules from YAML; Validate() with substring + regex modes; panic on invalid regex) | P1 | Backlog | 002-T003 | yes | | |
| 002-T026 | Write unit tests for `PromptValidator` (clean prompt passes; 5 seed rules trigger match; disabled rule skipped) | P1 | Backlog | 002-T025 | no | | |
| 002-T027 | Implement `OutputSanitizer` using `bluemonday.UGCPolicy()` (strip HTML/script before storage) | P1 | Backlog | 002-T001 | yes | | |
| 002-T028 | Write unit tests for `OutputSanitizer` (script stripped, event handlers stripped, plain text preserved) | P1 | Backlog | 002-T027 | no | | |
| 002-T029 | Wire `PromptValidator` into itinerary generation handler (return 400 PromptRejectionResponse on match) | P1 | Backlog | 002-T025, 001-T038 | no | | |
| 002-T030 | Wire `OutputSanitizer` into itinerary service before every DB INSERT of AI-generated content | P1 | Backlog | 002-T027, 001-T037 | no | | |
| 002-T031 | Write integration test for prompt rejection (10 injection payloads → assert 400 + request_id matches header) | P1 | Backlog | 002-T029 | yes | | |
| 002-T032 | Implement `DELETE /users/me` auth endpoint (cascade delete all PII; respond 204) | P1 | Backlog | 001-T021 | yes | | |
| 002-T033 | Write integration test for user deletion (register → create trip → DELETE → assert 204 + zero DB rows) | P1 | Backlog | 002-T032 | no | | |
| 002-T034 | Extend `backend-lint.yml` with `gosec -severity high` and `govulncheck` steps (gate on non-zero exit) | P1 | Backlog | 001-T076 | yes | | |
| 002-T035 | Extend `frontend-lint.yml` with `npm audit --audit-level=high` step (gate on critical/high findings) | P1 | Backlog | 001-T077 | yes | | |
| 002-T036 | Add `gitleaks detect` secret-scanning step to `backend-lint.yml` (every PR and push to main) | P1 | Backlog | 001-T076, 002-T005 | no | | |

#### Phase 6 — User Story 4: On-Call Engineer Diagnoses a Production Issue (Priority: P2)

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 002-T037 | Write integration test for `Logger` middleware (parse 5 requests' JSON log lines; assert all StructuredLogEntry fields) | P2 | Backlog | 002-T008 | yes | | |
| 002-T038 | Write integration test for `RequestID` middleware (no-header → UUID generated; preset header → echoed) | P2 | Backlog | 002-T006 | yes | | |
| 002-T039 | Write Playwright E2E test for `/healthz` (X-Request-ID header is UUID; response matches HealthCheckResponse schema) | P2 | Backlog | 002-T010 | no | | |
| 002-T040 | Create `backend/config/alerts.yml` defining 5xx error-rate alerting rule (> 1% over 5-minute window) | P2 | Backlog | 002-T008 | yes | | |

#### Phase 7 — User Story 5: Engineering Team Confirms Maintainability Standards (Priority: P2)

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 002-T041 | Add `@vitest/coverage-v8` and configure coverage thresholds ≥ 80% for `src/components/**` and `src/hooks/**` | P2 | Backlog | 002-T002 | yes | | |
| 002-T042 | Add `npm run test:coverage` script running `vitest run --coverage` (fails if thresholds not met) | P2 | Backlog | 002-T041 | yes | | |
| 002-T043 | Extend `frontend-test.yml` to run `npm run test:coverage`; upload coverage report as artefact | P2 | Backlog | 001-T079, 002-T042 | no | | |
| 002-T044 | Extend `backend-test.yml` to add `-coverprofile` flag and awk assertion for ≥ 80% total coverage | P2 | Backlog | 001-T078 | no | | |

#### Phase 8 — Polish & Cross-Cutting Concerns (Priority: P3)

| ID | Task | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|----------|--------|------------|----------|-------|-------|
| 002-T045 | Create OWASP ZAP baseline scan shell script (`scripts/owasp-zap-scan.sh`; exit 1 on FAIL-level finding) | P3 | Backlog | - | yes | | |
| 002-T046 | Create `backend/config/backup-policy.yml` (RPO ≤ 24 h, daily pg_dump schedule, 7-day retention, restore command) | P3 | Backlog | - | no | | |

---

## Feature Phase

_No feature specs defined yet. Add specs 003+ here as they are created._

---

## Critical Path

The minimum sequential chain to reach a fully functional, security-hardened, demo-able MVP:

```
001-T001   Create monorepo root structure
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
```

**Secondary critical paths** (branch from 001-T021):
- Privacy / GDPR: `001-T021 → 002-T032 → 002-T033`
- Collaboration: `001-T050 → 001-T063 → 001-T065 → 001-T067 → 001-T069 → 001-T075`
- Accessibility CI gate: `002-T002 → 002-T022 → 002-T023 → 002-T024`
- Security CI gates: `001-T076 → 002-T034 → 002-T036` and `001-T077 → 002-T035`
- Coverage gates: `001-T078 → 002-T044` and `001-T079 → 002-T043`

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

**Run date**: 2026-07-02

### Specs Discovered

| Spec | Folder | Has tasks.md | Status |
|------|--------|-------------|--------|
| 001 | `specs/001-product-vision-scope/` | ✅ yes | Active |
| 002 | `specs/002-nfr-system-constraints/` | ✅ yes | Active |

### Change Counts

| Operation | Count |
|-----------|-------|
| **ADD** | 127 (001: 81 tasks, 002: 46 tasks) |
| **UPDATE** | 0 |
| **UNCHANGED** | 0 |
| **REMOVE** | 0 |
| **ARCHIVED** | 0 |

**Total tasks in roadmap**: 127

### Human Attention Required

| Item | Detail |
|------|--------|
| Priority review | All priorities are defaults inherited from spec/phase. Review and override where needed — especially 001 Phase 7 (Polish, currently P2) if CI gates should be P1. |
| Status review | All tasks default to `Backlog`. Mark tasks `Ready` or `In Progress` as work begins. |
| Issue column | All 127 `Issue` fields are empty. Run the issue-creation command when ready to create tracker issues. Only rows with an empty `Issue` field should be created. |
| 002 Phase 5 cross-deps | 002-T029 and 002-T030 depend on 001-T038 and 001-T037 respectively. These can only be implemented after the corresponding 001 itinerary tasks complete. |
| 002 Phase 7 CI extensions | 002-T043 and 002-T044 modify CI workflow files created by 001-T079 and 001-T078. Coordinate with whoever implements those 001 tasks. |

### Next Steps

1. Review and adjust Priority/Status fields as needed — these are human-owned.
2. Run the issue-creation command to generate tracker issues from rows with empty `Issue` fields.
3. Begin implementation following the critical path: `001-T001 → 001-T009 → ... → 001-T050`.
4. Re-run `/build-roadmap` whenever a new spec's `tasks.md` is added or an existing one is modified — the command is idempotent and will add/update/archive as needed.
