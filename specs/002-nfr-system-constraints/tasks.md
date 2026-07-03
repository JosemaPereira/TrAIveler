# Tasks: Non-Functional Requirements and System Constraints — TrAIveler NFR Infrastructure

**Input**: Design documents from `specs/002-nfr-system-constraints/`

**Prerequisites**: [plan.md](plan.md) · [spec.md](spec.md) · [research.md](research.md) · [data-model.md](data-model.md) · [contracts/api.md](contracts/api.md)

**Tests**: Not explicitly requested in spec — no test tasks generated for user stories.
Test code is expected alongside implementation per TDD (Constitution I). The middleware,
validator, and sanitiser implementations each include a paired test file in the task list.

**Cross-spec awareness**: Tasks below are additive to `specs/001-product-vision-scope/tasks.md`.
Tasks that extend 001 CI workflow files (T076–T079) are marked with `[extends 001-T0XX]`.
The observability primitives in Phase 2 slot into the router wired in 001-T023.

**Format**: `- [ ] [ID] [P?] [Story?] Description — file path`
- `[P]` = parallelizable (no incomplete dependencies, different files)
- `[US#]` = user story label (Phase 3+ only)

---

## Phase 1: Setup

**Purpose**: Add NFR-specific dependencies, tool configurations, and seed config files.
None of these tasks modify existing code — they only add new files or update manifests.

- [ ] T001 Add `github.com/microcosm-cc/bluemonday` dependency to Go module — `backend/go.mod`
- [ ] T002 [P] Add `@axe-core/playwright` and `@lhci/cli` as dev dependencies — `frontend/package.json`
- [ ] T003 [P] Create `backend/config/prompt-rules.yml` with 5 seed deny-list rules covering: instruction-override, system-prompt-extraction, role-switching, jailbreak-prefix, and non-travel domain directive patterns — `backend/config/prompt-rules.yml`
- [ ] T004 [P] Create `lighthouserc.yml` at repo root with LHCI assertion thresholds: accessibility ≥ 0.9, LCP ≤ 2500 ms, CLS ≤ 0.1, INP ≤ 200 ms — `lighthouserc.yml`
- [ ] T005 [P] Create `.gitleaks.toml` at repo root configuring gitleaks to scan all committed files and exclude test fixture paths — `.gitleaks.toml`

---

## Phase 2: Foundational — Observability Core

**Purpose**: Implement the request-ID middleware, structured-log middleware, and `/healthz`
endpoint. These primitives are required by every subsequent NFR validation task and must
be wired into the Chi router before any load-test, security, or observability tests can run.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T006 Implement `RequestID` Chi middleware: generate UUID v4 if `X-Request-ID` header is absent; inject into request context; set header on response — `backend/internal/middleware/requestid.go`
- [ ] T007 Write unit tests for `RequestID` middleware: (1) generates UUID when header absent, (2) propagates existing UUID unchanged, (3) sets `X-Request-ID` on response — `backend/internal/middleware/requestid_test.go`
- [ ] T008 [P] Implement `Logger` Chi middleware: start timer on request entry; extract correlation ID from context; emit `slog.NewJSONHandler` JSON log line with all `StructuredLogEntry` fields on response completion — `backend/internal/middleware/logger.go`
- [ ] T009 Write unit tests for `Logger` middleware: (1) emits all required JSON fields, (2) `duration_ms` is non-negative, (3) `user_id` absent on unauthenticated request — `backend/internal/middleware/logger_test.go`
- [ ] T010 [P] Implement `GET /healthz` handler returning `HealthCheckResponse` JSON (`status`, `version`, `uptime_seconds`); handler must respond in ≤ 100 ms — `backend/pkg/health/handler.go`
- [ ] T011 Write unit tests for `/healthz` handler: (1) returns 200 OK, (2) response body matches `HealthCheckResponse` schema, (3) `status` is `"ok"`, (4) `uptime_seconds` ≥ 0 — `backend/pkg/health/handler_test.go`
- [ ] T012 Register `/healthz` route and apply `RequestID` → `Logger` middleware chain globally in the Chi router; wire `startTime` for `uptime_seconds` into handler via closure — `backend/cmd/server/main.go`

**Checkpoint**: `/healthz` returns `200 OK` with valid JSON; every request produces a structured JSON log line on stdout; `X-Request-ID` appears in all responses.

---

## Phase 3: User Story 1 — Engineering Team Verifies Performance Under Load (Priority: P1) 🎯

**Goal**: A load-test suite that can be triggered manually before each release asserts that the
system sustains 500 VUs for 10 minutes with p95 API latency ≤ 500 ms and error rate < 1%.

**Independent Test**: `k6 run tests/load/k6/baseline.js --env BASE_URL=<staging>` completes with
all threshold assertions passing (`✓`) and no failing checks.

- [ ] T013 [P] [US1] Create k6 baseline load test: ramp to 500 VUs over 2 min, sustain 10 min, ramp down 2 min; assert `http_req_duration{p(95)} < 500ms` and `http_req_failed rate < 0.01`; target non-AI endpoints (`/healthz`, `/api/v1/trips`) — `tests/load/k6/baseline.js`
- [ ] T014 [P] [US1] Create k6 API latency scenario: constant 100 RPS against non-AI endpoints; assert p95 ≤ 500 ms per endpoint — `tests/load/k6/scenarios/api_latency.js`
- [ ] T015 [US1] Create `load-test.yml` GitHub Actions workflow with `workflow_dispatch` manual trigger; runs `k6 run` against `${{ vars.STAGING_URL }}`; fails workflow if any k6 threshold assertion fails — `.github/workflows/load-test.yml`
- [ ] T016 [US1] Write integration test using `net/http/httptest` asserting `/healthz` responds in ≤ 100 ms for 100 sequential calls (validates NFR-OBS-002 timing requirement) — `backend/pkg/health/handler_test.go`

**Checkpoint**: `k6 run tests/load/k6/baseline.js` passes all threshold assertions on staging.

---

## Phase 4: User Story 2 — Accessibility Reviewer Confirms WCAG 2.1 AA Compliance (Priority: P1)

**Goal**: Every PR that touches frontend code automatically runs axe-core and Lighthouse CI scans;
violations or score < 90 block the merge. The registration flow has a privacy policy link (NFR-PRIV-003).

**Independent Test**: `npx playwright test --grep @accessibility` passes with zero axe-core
violations on all application pages; `npx lhci autorun` passes all thresholds in `lighthouserc.yml`.

- [ ] T017 [P] [US2] Implement `PrivacyPolicyPage` static page: heading, section structure (what data is collected, how it is used, right to deletion contact), WCAG-compliant landmarks — `frontend/src/pages/PrivacyPolicyPage.tsx`
- [ ] T018 [P] [US2] Write Vitest unit test for `PrivacyPolicyPage`: (1) renders heading, (2) renders at least three content sections, (3) no `dangerouslySetInnerHTML` usage — `frontend/src/pages/PrivacyPolicyPage.test.tsx`
- [ ] T019 [P] [US2] Implement `PrivacyPolicyLink` atom: accessible `<a>` element linking to `/privacy-policy`; text "Privacy Policy"; opens in same tab — `frontend/src/components/atoms/PrivacyPolicyLink.tsx`
- [ ] T020 [P] [US2] Write Vitest unit test for `PrivacyPolicyLink`: (1) renders with correct `href`, (2) accessible text present — `frontend/src/components/atoms/PrivacyPolicyLink.test.tsx`
- [ ] T021 [US2] Add `/privacy-policy` route to React Router and render `PrivacyPolicyPage`; add `PrivacyPolicyLink` to the registration form footer — `frontend/src/App.tsx`, `frontend/src/pages/RegisterPage.tsx`
- [ ] T022 [P] [US2] Create accessibility E2E helper wrapping `@axe-core/playwright`; exposes `checkPageA11y(page)` which runs the scan and throws on any WCAG 2.1 AA violation — `frontend/tests/helpers/a11y.ts`
- [ ] T023 [US2] Add `checkPageA11y(page)` call to every existing Playwright E2E spec; tag each accessibility assertion with `@accessibility` — `e2e/*.spec.ts`
- [ ] T024 [US2] Create `accessibility.yml` GitHub Actions workflow: (1) install deps, build frontend, serve; (2) run `playwright test --grep @accessibility`; (3) run `lhci autorun --config=lighthouserc.yml`; blocks PR on any failure — `.github/workflows/accessibility.yml`

**Checkpoint**: `npx playwright test --grep @accessibility` passes with zero violations; `npx lhci autorun` passes; privacy policy page renders and is linked from registration.

---

## Phase 5: User Story 3 — Security Reviewer Confirms Security Posture (Priority: P1)

**Goal**: The CI pipeline blocks merges on any secret, high-severity SAST finding, or critical
dependency CVE. The AI prompt-validation layer rejects injection attempts before they reach the
provider. AI output is sanitised before storage and rendering.

**Independent Test**: (1) `gosec -severity high ./...` exits with zero findings; (2) `gitleaks detect` exits with zero leaks; (3) unit test suite for `PromptValidator` and `OutputSanitizer` passes; (4) integration test for prompt-rejection endpoint returns 400 for all injection payloads in the test matrix.

- [ ] T025 [P] [US3] Implement `PromptValidator`: load rules from `backend/config/prompt-rules.yml` at startup; expose `Validate(prompt string) (ruleID string, matched bool)` supporting `substring` and `regex` match types; panic with descriptive message on invalid regex at startup — `backend/internal/ai/validator/prompt_validator.go`
- [ ] T026 [US3] Write unit tests for `PromptValidator`: (1) clean travel prompt returns `matched=false`, (2) each of the 5 seed rules triggers `matched=true`, (3) invalid regex causes startup panic, (4) disabled rule is not evaluated — `backend/internal/ai/validator/prompt_validator_test.go`
- [ ] T027 [P] [US3] Implement `OutputSanitizer`: apply `bluemonday.UGCPolicy()` to AI response text before any DB INSERT or API response serialisation; expose `Sanitize(raw string) string` — `backend/internal/ai/sanitizer/output_sanitizer.go`
- [ ] T028 [US3] Write unit tests for `OutputSanitizer`: (1) `<script>alert(1)</script>` stripped, (2) `onerror=` event handler stripped, (3) plain text preserved unchanged, (4) safe markdown-like formatting (bold text, lists) preserved — `backend/internal/ai/sanitizer/output_sanitizer_test.go`
- [ ] T029 [US3] Wire `PromptValidator` into the itinerary generation handler: call `Validate()` on the user's input before forwarding to the AI provider; return `PromptRejectionResponse` (400) and emit a `WARN` slog entry with `"error": "rule matched: <ruleID>"` on match — `backend/internal/itinerary/handler.go`
- [ ] T030 [US3] Wire `OutputSanitizer` into the itinerary service: call `Sanitize()` on AI response content before every DB INSERT of itinerary text fields — `backend/internal/itinerary/service.go`
- [ ] T031 [P] [US3] Write integration test for prompt rejection: use `httptest` to POST a set of 10 known injection payloads to the generation endpoint; assert each returns 400 with `"error": "invalid_prompt"` and that `request_id` in the response matches `X-Request-ID` header — `backend/internal/ai/validator/prompt_validator_integration_test.go`
- [ ] T032 [P] [US3] Implement `DELETE /users/me` auth endpoint: delete the authenticated user's account and all associated PII (cascade DELETE via FK); respond 204 on success — `backend/internal/auth/handler.go`
- [ ] T033 [US3] Write integration test for user deletion (NFR-PRIV-001): register user → create trip → DELETE /users/me → assert 204; assert subsequent GET /me returns 401; assert DB query returns zero rows for user and all associated trips, itinerary days, activities — `backend/internal/auth/user_deletion_test.go`
- [ ] T034 [P] [US3] Extend `backend-lint.yml` (extends 001-T076) to add `gosec -severity high ./...` step and `govulncheck ./...` step; both gate on non-zero exit — `.github/workflows/backend-lint.yml`
- [ ] T035 [P] [US3] Extend `frontend-lint.yml` (extends 001-T077) to add `npm audit --audit-level=high` step; gate on any critical/high finding — `.github/workflows/frontend-lint.yml`
- [ ] T036 [US3] Add `gitleaks detect --source . --no-git` step to `backend-lint.yml` (runs on every PR and push to `main`); gate on any secret finding; add `.gitleaks.toml` allowlist for test fixture files — `.github/workflows/backend-lint.yml`

**Checkpoint**: PR with a hardcoded secret is blocked by gitleaks; PR with a prompt injection payload is rejected 400; unit test suites for `PromptValidator` and `OutputSanitizer` all pass.

---

## Phase 6: User Story 4 — On-Call Engineer Diagnoses a Production Issue (Priority: P2)

**Goal**: Integration tests formally verify the structured log format and correlation-ID contract
so that alerting rules and runbooks can safely reference the log field names.

**Independent Test**: `go test ./internal/middleware/...` passes all integration assertions for log
structure, header presence, and field-name stability.

- [ ] T037 [P] [US4] Write integration test for `Logger` middleware: use `httptest` to make 5 requests; capture stdout; parse each JSON log line; assert all `StructuredLogEntry` fields are present with correct types — `backend/internal/middleware/logger_test.go`
- [ ] T038 [P] [US4] Write integration test for `RequestID` middleware: make a request with no `X-Request-ID` header; assert response header contains a valid UUID v4; make a request with a preset `X-Request-ID`; assert response echoes the same value — `backend/internal/middleware/requestid_test.go`
- [ ] T039 [US4] Write Playwright E2E test: navigate to the app; make a `fetch()` call to `/healthz`; assert `X-Request-ID` response header is present and is a valid UUID; assert response body matches `HealthCheckResponse` schema — `e2e/health-check.spec.ts`
- [ ] T040 [P] [US4] Create `backend/config/alerts.yml` defining the error-rate alerting rule (5xx rate > 1% over 5-minute window); document the slog field reference (`status`) used by the rule — `backend/config/alerts.yml`

**Checkpoint**: All middleware integration tests pass; Playwright health-check spec passes.

---

## Phase 7: User Story 5 — Engineering Team Confirms Maintainability Standards (Priority: P2)

**Goal**: The CI pipeline enforces coverage floors (≥ 80%) on every PR; coverage drops block merges.

**Independent Test**: Create a branch that deletes a test file; push; assert `backend-test.yml` and `frontend-test.yml` workflows fail with a coverage gate message.

- [ ] T041 [P] [US5] Add `@vitest/coverage-v8` dev dependency and configure coverage thresholds in `vitest.config.ts`: `lines: 80`, `functions: 80` for `src/components/**` and `src/hooks/**`; `reporter: ['text', 'lcov']` — `frontend/package.json`, `frontend/vitest.config.ts`
- [ ] T042 [P] [US5] Add `npm run test:coverage` script to `package.json`; script runs `vitest run --coverage` and fails if thresholds are not met — `frontend/package.json`
- [ ] T043 [US5] Extend `frontend-test.yml` (extends 001-T079) to run `npm run test:coverage` instead of `npm test -- --run`; upload coverage report as a workflow artefact — `.github/workflows/frontend-test.yml`
- [ ] T044 [US5] Extend `backend-test.yml` (extends 001-T078) to add `-coverprofile=coverage.out` flag to `go test ./internal/...`; add `go tool cover -func=coverage.out` step; add `awk` assertion that total coverage is ≥ 80% (exit 1 if below) — `.github/workflows/backend-test.yml`

**Checkpoint**: Deleting a Go test file and opening a PR causes `backend-test.yml` to fail at the coverage gate; same for a frontend test file.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final integration of all NFR gates; OWASP ZAP scan documentation; backup strategy
configuration reference.

- [ ] T045 [P] Create `scripts/owasp-zap-scan.sh` shell script that runs OWASP ZAP baseline scan (`docker run zap ...`) against a provided `BASE_URL`; outputs results to `reports/zap-baseline.html`; exit 1 on any FAIL-level finding — `scripts/owasp-zap-scan.sh`
- [ ] T046 Create `backend/config/backup-policy.yml` documenting RPO ≤ 24 h, daily pg_dump schedule reference, retention period (7 daily snapshots), and restore validation command — `backend/config/backup-policy.yml`

---

## Dependencies

```
Phase 1 (T001–T005)
    └── Phase 2 (T006–T012)  ← must complete before any NFR validation
            ├── Phase 3: US1 Performance (T013–T016)   P1 — load tests reference /healthz
            ├── Phase 4: US2 Accessibility (T017–T024) P1 — can start after Phase 1 (frontend only)
            ├── Phase 5: US3 Security (T025–T036)      P1 — prompt validator wired after T012
            │       └── T029, T030 depend on 001-T037 (itinerary service + handler)
            │           T032, T033 depend on 001-T021 (auth handler + user entity)
            ├── Phase 6: US4 Observability (T037–T040) P2 — validates Phase 2 implementation
            └── Phase 7: US5 Maintainability (T041–T044) P2 — extends 001-T078 and 001-T079

Phase 8 (T045–T046) — independent; can be done any time after Phase 1

001 cross-dependencies (MUST be completed before listed tasks):
  001-T021 (auth handler) → 002-T032, 002-T033
  001-T037 (itinerary service) → 002-T029, 002-T030
  001-T076 (backend-lint.yml created) → 002-T034, 002-T036
  001-T077 (frontend-lint.yml created) → 002-T035
  001-T078 (backend-test.yml created) → 002-T044
  001-T079 (frontend-test.yml created) → 002-T043
  001-T023 (Chi router wired) → 002-T012 (register /healthz and middleware)
```

---

## Parallel Execution Examples

### During Phase 2 (Foundational)

| Stream A — Request ID | Stream B — Logger | Stream C — Health |
|----------------------|-------------------|-------------------|
| T006 → T007          | T008 → T009       | T010 → T011       |
| T012 (wires all three, runs last) | | |

### During Phase 3 + 4 + 5 in parallel (after Phase 2)

| Stream A — Performance (US1) | Stream B — Accessibility (US2) | Stream C — Security (US3) |
|-----------------------------|-------------------------------|--------------------------|
| T013                        | T017, T018, T019, T020        | T025 → T026              |
| T014                        | T021 → T022                   | T027 → T028              |
| T015                        | T023 → T024                   | T029, T030 (after 001-T037) |
| T016                        |                               | T031                     |
|                             |                               | T032 → T033 (after 001-T021) |
|                             |                               | T034, T035, T036          |

### During Phase 6 + 7 in parallel (after Phase 2)

| Stream A — Observability (US4) | Stream B — Maintainability (US5) |
|-------------------------------|----------------------------------|
| T037                          | T041, T042                        |
| T038                          | T043 (after 001-T079)             |
| T039                          | T044 (after 001-T078)             |
| T040                          |                                   |

---

## Implementation Strategy

**MVP Scope (minimum viable NFR baseline)**: Complete **Phase 1 + Phase 2 + Phase 5 (US3 security
core: T025–T031)** only. This delivers: structured logging, correlation IDs, health endpoint,
prompt injection defence, and AI output sanitisation — the minimum required to safely ship any
AI-powered feature from spec 001.

**Increment 2**: Add Phase 3 (US1 load tests) + Phase 4 (US2 accessibility) + Phase 7 (US5
coverage gates). These complete the full set of P1 NFRs and establish the CI quality gates.

**Increment 3**: Add Phase 6 (US4 observability validation) + security CI gates (T034–T036) +
user deletion (T032–T033). This closes the full NFR backlog and prepares for production readiness.

**Increment 4**: Phase 8 polish — OWASP ZAP scan and backup policy documentation.
