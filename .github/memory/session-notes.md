# Session Notes

Historical summaries of completed development sessions. Committed to git as a record.

**Note**: Older sessions (specs 001-007, sprint infrastructure setup) are compacted to save tokens. Policy: each sprint's implementation sessions are compacted into a `(Compacted)` summary immediately at that sprint's own closure — the same pass that updates the sprint-closure documentation. At any given time, only the sprint currently in progress (not yet closed) has a `(Detailed)` section; the moment it closes, it gets compacted, no rolling window. Compacted sections favor brevity over completeness: PR/issue/Group IDs, decisions with lasting effect, and still-open follow-ups only — narrative process detail is cut, and duplicates of `patterns-discovered.md` entries are referenced by name instead of re-explained. Sprint 1 and Sprint 2 were compacted together on 2026-07-11; Sprint 3 was compacted at its own closure on 2026-07-12 (and re-tightened for conciseness the same day, which also lightly trimmed the Sprint 1/2 sections); Sprint 4 was compacted at its own closure on 2026-07-14; Sprint 5 was compacted at its own closure on 2026-07-17 (264 lines → 97; the pre-compaction full text is kept, uncommitted, at `scratch/session-notes-pre-sprint5-compaction-2026-07-17.md`); Sprint 6 was compacted at its own closure on 2026-08-01 (489 lines → ~130; the pre-compaction full text is kept, uncommitted, at `scratch/session-notes-pre-sprint6-compaction-2026-08-01.md`). No sprint currently has a `(Detailed)` section — Sprint 7 opens the next one.

## Template

### Session: <name>
- **Date**: <YYYY-MM-DD>
- **Tool**: <Claude Code | GitHub Copilot>
- **What was accomplished**: <features built/fixed>
- **Key findings and decisions**: <important learnings, trade-offs>
- **Outcomes**: <what now works in the application>

---

## Foundation Phase (Compacted)

### Sessions: Specs 001-007 Foundation Work
- **Date Range**: 2026-07-02 to 2026-07-06
- **Specs Created**: 001 (Product Vision), 002 (NFRs), 003 (Cloud/IaC), 004 (Security/Auth), 005 (Architecture), 006 (Data Model), 007 (API Standards)
- **Key Outcomes**:
  - **Product Vision**: 4 personas, subscription model, suggest-then-approve collaboration, visible payment stub, Anthropic Claude AI
  - **NFRs**: 38 measurable requirements across 8 quality attributes, prompt injection deny-list, output sanitization (bluemonday)
  - **Cloud Strategy**: AWS us-east-1, ECS Fargate (not Lambda), 2-environment topology (staging $200/mo active, production $300-400/mo dormant), Terraform IaC, OIDC CI/CD
  - **Security Model**: JWT RS256 with multi-key rotation, bcrypt cost 12, optimistic locking (version numbers), 30-day CloudWatch logs
  - **Architecture**: Go 1.24+ backend (Chi, pgx/v5, goose, slog), React 19 frontend (Vite, TanStack Query v5, Zustand, Router v7), Playwright E2E
  - **Data Model**: 16 entities with 101+ validation tags ([DB], [Logic], [API]), 18 invariants, 26 performance indexes, forward-only state transitions, GDPR-aware
  - **API Standards**: 15-section comprehensive guide covering resource naming, URL structure, versioning, HTTP methods, standardized error format (11 codes), pagination (offset-based), filtering (8 operators), sorting, rate limiting (100/min), 7 endpoint patterns
  - **Documentation Promoted**: 8 docs/ files created (product-vision.md, nfrs.md, security.md, cloud-and-environments.md, architecture.md, data-model.md, api-design-standards.md, plus others)
  - **Constitution**: Updated to v1.3.0 with mandated AWS/Terraform/ECS stack, prompt injection prevention (NON-NEGOTIABLE), output sanitization (NON-NEGOTIABLE)
  - **Roadmap**: Built consolidated 745-task roadmap across 8 specs (539 foundation + 206 feature spec 008), 49 task groups, 54.7% parallelizable
  - **Reports**: Created PROMOTION-REPORT.md and ROADMAP-RECONCILIATION-REPORT.md for audit trail
  - **Memory System**: Established session-notes.md, patterns-discovered.md, working-notes.md structure with mandatory Session Start Protocol
- **Key Decisions**:
  - ECS Fargate over Lambda (AI workloads need unlimited execution time, no payload limits, persistent connections)
  - Pattern-based prompt deny-list over LLM validation (zero AI cost per rejection, deterministic blocking)
  - Foundation promotion workflow (specs → constitution/docs/ → copilot-instructions.md) prevents context drift
  - Task grouping reduces 127 tasks → 81 issues (59% grouped)
  - TDD mandate for all security tasks (RED-GREEN-REFACTOR)
  - Roadmap reconciliation is idempotent (ADD/UPDATE/REMOVE with human-field preservation)

---

## Sprint Planning Phase (Compacted)

### Sessions: Sprint Infrastructure Setup
- **Date**: 2026-07-06
- **Key Outcomes**:
  - **10-Sprint MVP Plan**: 390 of 745 tasks assigned, 20 weeks timeline, 2-3 developers
  - **GitHub Infrastructure**: 39 labels created (11 epic, 10 sprint, 3 priority, 4 type, 2 status, 1 group, 8 spec)
  - **Sprint 1 Issues**: Created 23 issues (#13-35) with proper labels and project assignment
  - **Dependency Tracking**: Automated script establishing 19 blocking relationships across 4 foundation issues
  - **Consolidation Policy**: Integrated consolidation-first workflow into PM agent, plan-sprints prompt, and guidelines
- **Key Decisions**:
  - Simplified workflow: one task = one issue (no parent issues), GitHub Projects handles grouping
  - Dependency tracking automated via cross-reference comments (mandatory for all sprints)
  - Sprint 2 pre-consolidated: 37 tasks → 19 work items (-49% reduction) before issue creation
  - Template-based scripts for sprint-specific dependency tracking

---

## Sprint 1 Implementation (Compacted)

### Sessions: Sprint 1 Closure, Sprint 2 Consolidation Planning, Git Templates, Doc Audit & Claude Code Migration
- **Date Range**: 2026-07-08 to 2026-07-09
- **Key Outcomes**:
  - Sprint 1 closure: 23/23 tasks done across 8 PRs (#43-#50). Established ~23-tasks/sprint velocity baseline used to scope Sprint 2.
  - Sprint 2 pre-consolidated: 37 tasks (005-T024–T060) → 14 work items (-62%). `.github/SPRINT-CONSOLIDATION-CHECKLIST.md` created as the enforced pre-issue-creation checkpoint going forward.
  - `.github/PULL_REQUEST_TEMPLATE.md` and `.github/COMMIT_GUIDELINES.md` created (Conventional Commits, Stable-ID/Spec/Sprint/Group fields) to fix Sprint 1's inconsistent commit/PR structure.
  - Doc audit (005-T029): backend `config`/`database` packages documented; all 4 project areas already met documentation standards.
  - Claude Code established as primary tool: `CLAUDE.md` is now canonical (`copilot-instructions.md` a mirror); `.github/memory/README.md` unified cross-tool memory protocol. 10 Copilot agents/prompts ported to `.claude/agents/`. Official spec-kit Claude integration installed; a mistaken hand-ported duplicate (`.claude/agents/speckit-*.md` + `.claude/commands/`) was found and deleted once the official integration covered the same ground.
- **Key Decisions**:
  - Consolidation-first became **mandatory starting Sprint 2** (one-task-per-issue caused backlog fragmentation). Group naming `G-SPRINT<N>-<STACK>-<AREA>`, optimal size 2-4 tasks; foundation/critical-blocker/single-file work stays standalone.
  - **Always check for an official multi-agent-tool integration before hand-porting a toolkit's own agent config** (`patterns-discovered.md`) — why the SpecKit hand-port was reverted.
  - Bidirectional drift detection (`scripts/check-agent-drift.py`) adopted to reconcile Claude/Copilot hand-ported agent pairs.
  - Killed background `Agent`-tool tasks: resume once via `SendMessage`; if killed again, fall back to foreground.

---

## Sprint 2 Implementation (Compacted)

### Sessions: Backend Foundation (Middleware, HTTP Server, Errors, AI Client, Reference Pattern), CI/Docker Fixes & Repo Ruleset, Frontend Foundation (Tokens, Primitives, App Shell, Auth Store)
- **Date Range**: 2026-07-09 to 2026-07-10 (closed 2026-07-11 — 37 tasks / 14 work items, 10+ PRs)
- **Key Outcomes**:
  - G-SPRINT2-BACKEND-MIDDLEWARE (005-T024–T028, #54, PR #71): `RequestID`/`Logger`/`Recovery`/`CORS`/`BodySize`, constructor-injected. Established "mockery is for external-system interfaces only" rule (`patterns-discovered.md`).
  - G-SPRINT2-BACKEND-HTTP-SERVER (005-T035–T037, #57, PR #72): real Chi router, mandated middleware order, `/healthz`, graceful shutdown.
  - G-SPRINT2-BACKEND-ERRORS (005-T033–T034, #56, PR #74): `DomainError` + `HandleError`. Same session fixed `backend/Dockerfile`'s QEMU-emulated builder (PR #73, `--platform=$BUILDPLATFORM`, 8.5min→~56s build) — formalized in `patterns-discovered.md` ("Native Cross-Compilation..."). `alpine:3.19`→`3.23` bumped (CVE); `postgres:15.4-alpine` deliberately left alone (architectural pin, not drift).
  - Repo governance (PR #75): "Protect main" ruleset (10 required checks). Fixed a real gap: `paths:`-filtered required-check workflows stick at "Expected" forever — fixed via an always-running gate job, formalized as "Required Status Checks Must Always Run" (`patterns-discovered.md`).
  - G-SPRINT2-BACKEND-AI-CLIENT (005-T030–T032, #55, PR #77): `AIClient` interface + stubs, plus an approved scope extension — real local-dev `OllamaClient` (default in dev), `docker-compose.yml` `ollama` service, `docs/local-ai-setup.md`. Surfaced the Swagger/OpenAPI planning gap for Sprint 3.
  - G-SPRINT2-BACKEND-EXAMPLE (005-T038–T041, #58, PR #78): `internal/example/` canonical model→repository→service→handler reference pattern, `If-Match` optimistic locking, a two-query pagination fix (`patterns-discovered.md`). **Throwaway — must be deleted once the first real domain package ships**; still pending.
  - G-SPRINT2-FRONTEND-TOKENS (005-T042–T043, #59, PR #79): `tokens.css`/`global.css` from `docs/ui-guidelines.md`; installed the real Vitest/RTL/MSW toolchain (was missing entirely). Flagged the Lighthouse/a11y-CI scheduling gap for Sprint 3.
  - G-SPRINT2-FRONTEND-PRIMITIVES+FORM (005-T048–T051/T055, #63/#65, PR #80): `Button`/`Input`/`Card`/`Form`. Formalized "issue-body pseudocode is lowest authority" as a general rule (`patterns-discovered.md`: real code > `tasks.md` > root docs > issue body).
  - G-SPRINT2-FRONTEND-APP-SHELL (005-T044/T045/T047/T056-T060, #60/#62/#66, PR #81): `api-client.ts`/`query-client.ts`, `ErrorBoundary` + Router v7, `App.tsx` wiring. Removed a hardcoded `VITE_API_BASE_URL` fallback masking a CI wiring gap.
  - G-SPRINT2-FRONTEND-AUTH-STORE+PRIMITIVES (005-T046/T052-T054, #61/#64, PR #82): Zustand `auth-store.ts` (`'admin'|'partner'` roles per `docs/data-model.md`), `LoadingSpinner`/`ErrorMessage`/`EmptyState`. Closed out all remaining Sprint 2 frontend items.
- **Key Decisions**:
  - "Issue-body pseudocode/diagrams go stale — verify against real code/docs" recurred all sprint; promoted to a general authority-ranking rule in `patterns-discovered.md`.
  - **Independently re-verify every subagent's output** (rebuild, re-run tests/lint, read the diff) rather than trusting self-reports — caught real issues multiple times this sprint.
  - Operationally consequential platform behavior (GitHub required-checks, Docker/QEMU) must be verified against current docs or by actually running it, not recalled.
  - Scope-extension decisions on a ticket (e.g. Ollama-as-real-client) go through explicit `AskUserQuestion` confirmation, not inference.
  - **Known open items carried past Sprint 2 close**: `postgres:15.4-alpine` staleness decision pending; `internal/example/` deletion pending (blocked on first real domain package).

---

## Sprint 3 Implementation (Compacted)

### Sessions: Terraform Modules (VPC, RDS, ALB, CloudFront, Secrets, ECS, Root Wiring, CI/CD+OIDC), Terraform State Bootstrap, Accessibility CI Gate
- **Date Range**: 2026-07-11 to 2026-07-12 (closed 2026-07-12 — 54 tasks / 11 work items, 10 PRs)
- **Key Outcomes** (`infra/modules/` unless noted; PR = merged, no live `terraform apply`/`aws` calls made against real AWS in any of them):
  - G-SPRINT3-INFRA-VPC (005-T061–T068, #84, PR #95): `vpc/` — subnets across 2 AZs, IGW, conditional NAT, tagging.
  - G-SPRINT3-INFRA-RDS (005-T078–T084, #86, PR #98): `rds/` — Postgres 15.4, ECS-only ingress, owns the `db_credentials` secret (auto via `random_password`).
  - G-SPRINT3-INFRA-ALB (005-T085–T092, #87, PR #99): `alb/` — HTTPS→target-group, HTTP redirect-only.
  - G-SPRINT3-INFRA-CLOUDFRONT (005-T093–T098, #88, PR #100): `cloudfront/` — private S3 + OAI, SPA 404→`/index.html`. Same PR added `docs/README.md` and established the "two-tier docs update" pattern (`patterns-discovered.md`, "What 'Keep docs/ Updated' Actually Means for an Infra Module PR").
  - G-SPRINT3-INFRA-SECRETS (005-T099–T101, #89, PR #102): `secrets/` — only `ai_api_key`/`jwt_signing_key` (not `db_credentials`, owned by RDS above — deliberate split, avoids two sources of truth).
  - G-SPRINT3-INFRA-ECS (005-T069–T077, #85, PR #105): `ecs/` — Fargate, ARM64 task def, autoscaling 70%, ALB-only ingress (added additive `alb_security_group_id` output to the ALB module for this).
  - G-SPRINT3-INFRA-ROOT-WIRING (005-T102–T106, #90, PR #106): root `main.tf` wires all 6 modules + staging/production `.tfvars`. Fixed a real bug: `alb/variables.tf` had an unused `ecs_security_group_id` that would have created a circular module dependency — removed.
  - G-SPRINT3-INFRA-CICD + 005-T109 OIDC (005-T107–T109, #91/#92, PR #107): `infra-plan.yml` real init/validate/fmt/plan; `infra-apply.yml` (staging auto-apply on merge to `main`) added but every AWS-touching step gates on the nonexistent `AWS_ROLE_ARN` secret (skip, not fail); OIDC federation setup documented in `infra/README.md` only, no `aws iam` calls.
  - 003-T009 Terraform state bootstrap (#96, PR #104): idempotent `infra/scripts/bootstrap-state.sh` (S3 + DynamoDB lock table), test-verified against a mocked `aws` CLI, never run for real.
  - G-SPRINT3-A11Y-CI (002-T002/T004/T022/T024, #93, PR #103): `@axe-core/playwright`+`@lhci/cli`, `lighthouserc.yml`, `checkPageA11y` helper, `accessibility.yml` workflow. Two CI bugs fixed and already captured in `patterns-discovered.md`: `interaction-to-next-paint` → `total-blocking-time` (INP can't be measured in single-navigation `lhci autorun`), and a required-check stuck at "Expected" from a job-name mismatch (must exactly match the ruleset string).
- **Key Decisions / still open**:
  - AWS-cost-avoidance constraint held all sprint (nothing executed against real AWS) and **remains in force** — no real `terraform apply`/`aws` resource creation until the user says infra is ready.
  - Open follow-up: 002-T023 (tag E2E specs `@accessibility`, Sprint 9) still needed before `accessibility.yml` can become a required check (`sprint3-lighthouse-ci-gap` memory).
  - Carried from Sprint 2, still open: `postgres:15.4-alpine` staleness watch; `internal/example/` deletion (blocked on first real domain package).

---

## Sprint 4 Implementation (Compacted)

### Sessions: Observability Core (Logger/Healthz), Backend/Frontend Integration Patterns, Integration Tests, Spec 009 Swagger/OpenAPI MVP, Spec 005 Validation & Polish
- **Date Range**: 2026-07-13 to 2026-07-14 (closed 2026-07-14 — 43 tasks / 15 work items, 10 PRs,
  15/15 issues #109-#123 closed)
- **Key Outcomes**:
  - G-OBS-LOGGER (002-T008/T009, #109, PR #125): `Logger` middleware extended with `service`/`msg`/
    `user_id` fields on the Sprint 2 baseline (005-T025, PR #71), matching `StructuredLogEntry`.
  - G-OBS-HEALTHZ (002-T010/T011, #110, PR #126): `/healthz` reconciled to `HealthCheckResponse`
    (`status: "ok"/"degraded"`, `version`, `uptime_seconds`, always `200 OK` — DB failure maps to
    `degraded` + `logger.Warn`, not a non-2xx). See `patterns-discovered.md`, "A Spec's Data-Model
    Validation Rule Can Override Normal REST Convention — Read It Literally."
  - G-SPEC009-SETUP through G-SPEC009-INTERACTIVE (009-T001-T017, #111-#116, one consolidated PR
    #127): full Swagger/OpenAPI MVP slice — `swag`/`swaggo` tooling, `cmd/api/docs.go` annotation
    skeleton, RED-phase contract tests, `internal/example/handler.go` annotated, `/swagger/*` mounted
    (top-level URL, shared `/api/v1` Chi middleware group, `TODO(sprint-5)` marker for future JWT
    gating per `specs/009-api-documentation/research.md`), `/swagger/index.html` verified pointing at
    the live `/swagger/doc.json` (not stale), README "Try the API" section added.
  - G-ARCH-INTEGRATION-BACKEND (005-T110-T113, #117, PR #128): 005-T110 (correlation ID) was already
    satisfied by PR #74, no code touched. New: 5xx stack-trace logging in `Logger`; `internal/ai/
    anthropic.go` `AnthropicClient` (`anthropic-sdk-go` v1.57.0, SDK-native retry/timeout — see
    "Prefer an Official SDK's Built-In Retry Over Hand-Rolling One"); `NewAIClient(cfg)` provider-
    agnostic factory (Anthropic and Ollama both first-class, per explicit user requirement — no route
    wires it in yet) + `TranslateError` → `ServiceUnavailable` `DomainError` with `Retry-After` header.
    Confirmed empirically that no mockery regen was needed (interface unchanged).
  - G-ARCH-INTEGRATION-FRONTEND (005-T114-T116, #118, PR #129): 005-T114 (`X-Request-ID`) already
    satisfied by PR #81. New: `frontend/src/hooks/useErrorHandler.ts` (first hook in `src/hooks/`;
    401→redirect, 403/404 non-retryable, 500/503/other retryable); `ErrorMessage.tsx` gained an
    optional `requestId` prop. New reusable test idiom: `MemoryRouter` + `vi.mock('react-router')` to
    spy on `useNavigate`. Neither wired into a real page yet (no live-API route exists pre-Sprint 8).
  - G-ARCH-INTEGRATION-TESTS (005-T117/T118, #119, PR #131): `health_test.go`/`error_test.go`; DB-
    timeout scenario used a realistic simulated failure — see "Simulating a Realistic DB Timeout in
    an Integration Test."
  - 005-T119 docs (#120, PR #132): backend/frontend README error-handling sections with verified
    examples; fixed a stale pre-#126 `/healthz` table left behind in `backend/README.md`.
  - G-ARCH-POLISH-VALIDATION (005-T124-T130, #122, PR #133): first `validation-results.md` in the
    repo — all 7 checks passed for real (build/healthz/DB-pool/CORS/graceful-shutdown; frontend 0
    console errors; terraform init/validate/fmt clean, `plan` deliberately skipped per the standing
    AWS-cost-avoidance policy; lint 0 errors; `internal/example` coverage 89.1% full suite vs. 64.1%
    CI-short; 136/136 frontend tests; ad-hoc axe-core scan 0 violations). Deliberately did not author
    a permanent `accessibility.spec.ts` — 002-T023/Sprint 9 still owns that.
  - G-ARCH-POLISH-DOCS (005-T120-T123, #121) and 005-T131 (#123), both PR #134: architecture docs
    polish and `specs/005-system-architecture/tasks.md` marked complete, closing out Spec 005
    end-to-end. No dedicated session-log entry exists for these two (PR-only reconstructed history,
    same as several Sprint 3 infra work items).
  - PR #130: standalone roadmap-status-drift fix (005-T110-T116 Backlog→Done), closed no issue.
- **Key Decisions**:
  - `/healthz` always-200 behavior is a deliberate, spec-literal reversal of typical REST convention
    (`patterns-discovered.md`).
  - `AnthropicClient`/`OllamaClient` are equally first-class in `NewAIClient` — never frame Anthropic
    as primary and Ollama as a dev-only fallback (see `sprint2-ai-client-ollama-mvp` personal memory).
  - Swagger UI shares the `/api/v1` Chi middleware group specifically so Sprint 5's JWT middleware
    will cover it too — `TODO(sprint-5)` marks the removal point once that lands.
  - **Process gap found and fixed mid-sprint**: the "flip roadmap Backlog→Done as part of the closing
    PR" convention (adopted this sprint per explicit user instruction, replacing the old
    sprint-closure-only reconciliation) initially slipped for #109/#110/#117/#118 before being caught
    and corrected (PR #130). Every ticket from that point on self-closed its own roadmap row.
  - Flagged, not fixed (pre-existing, out of scope): `specs/005-system-architecture/data-model.md`'s
    `AIClient` snippet still shows the old `GenerateItinerary`/`ItineraryChunk` shape; `specs/005-
    system-architecture/contracts/frontend-patterns.md` still shows `ErrorMessage`'s old children-based
    API instead of the real prop-based one (spec-kit-owned contract docs, same reasoning as the
    "Issue-Body Snippets Are Lowest-Authority" pattern).
  - Carried forward unresolved (see Sprint 5 entry in `docs/roadmap.md`'s `## Sprint Plan`):
    AWS-cost-avoidance constraint still in force; `internal/example/` deletion still waits on Sprint 8
    (Trip); `postgres:15.4-alpine` staleness watch; 002-T023 (Sprint 9); 009-T018-T026 (Spec 009
    swagger-drift CI gate + polish, deferred to Sprint 5).

---

## Sprint 5 Implementation (Compacted)

### Sessions: Spec 008 Setup/Database/Migrations, Config Loader, JWT, Password, RateLimit, Security Middleware Chain, Frontend Infra/Styles, Test-Suite BDD/TDT Alignment + CI Consolidation, Sprint 5 Closure
- **Date Range**: 2026-07-14 to 2026-07-17 (closed 2026-07-17 — 49/50 tasks, 15 work items
  #136–#150, PRs #151–#164; 004-T011 the one genuine miss, see below)
- **Key Outcomes**:
  - **#143 008-T018 config loader** (PR #161): **"raw now, ARN later"** secrets model, confirmed with
    the user. `backend/config/config.go` already loaded DATABASE_URL/JWT_SIGNING_KEY/ANTHROPIC_API_KEY
    as raw values — that IS T018's deliverable. App-side `*_SECRET_ARN` resolution deferred to
    004-T070/T071/T072 (Sprint 6+); raw values cover both local dev and the ECS `valueFrom` path.
  - **#141 G-008-DATABASE** (008-T009/T010): T009 already satisfied — no new file. T010 added
    `internal/database/migrations` (`Dir`/`SetDialect`), replacing 3 duplicated call sites. Fixed a
    real gap found while verifying: `NewClient` ignored `DB_MIN/MAX_CONNECTIONS` (see
    `patterns-discovered.md`, "A Config Value Can Be Loaded, Validated, and Logged, Yet Still Never
    Reach the Code It Configures").
  - **#142 G-008-MIGRATIONS** (PR #160): 7 tables. **One `users` table** — Spec 008's delta shipped as
    `005_alter_users_add_auth_fields.sql` on top of Spec 004's `001_create_users_table.sql`, resolving
    the duplicate-`users` risk flagged at planning. Spec 008's tables landed at repo-global numbers
    **006–012**, not the 001–007 its task text names.
  - **#144 G-008-JWT** (PR #162, T019–T021): subpackage `internal/auth/jwt/` — `Generator` (RS256, 24h,
    `has_subscription` claim + `kid` header), `Validator` (multi-key rotation via per-token `kid`;
    RS256-locked), `Refresher` (one-time-use rotation; resolves `has_subscription` fresh, not from the
    stale claim). `golang-jwt/jwt/v5` imported as `gojwt` to dodge the package-name collision.
    `RefreshTokenStore`/`SubscriptionResolver` interfaces only — impls deferred to Sprint 6-7.
    Hardened against alg-confusion, retired-key, and revoked-token reuse; all collapse to a uniform
    `authentication_required` error.
  - **#145 G-008-PASSWORD** (T022/T023): flat `internal/auth/password.go` + `validator.go` — there is
    **no `internal/auth/password/` subpackage**; this is the *same* deliverable as 004-T012.
  - **#146 G-008-RATELIMIT** (T024/T025, PR #163): `internal/auth/ratelimit` — email-keyed progressive
    delay (`2^(attempts-5)`s after 5 failures, clamped to the 15-min window). Real consumer is
    **008-T098 (login, Sprint 6), NOT T029**.
  - **#147 G-008-SECURITY-PKG** (T026–T029, PR #163): T026 already satisfied by
    `observability.LogSecurityEvent` (004-T014), T028 by `middleware.RequestID` (Sprint 2) — both
    no-new-code. T027 `Authenticate` + T029 per-IP `RateLimit` landed in **`internal/middleware/`, not
    `internal/security/`** (repo precedent beat the literal path; reuses `ctxKeyUserID`). Import cycle
    (`middleware`→`jwt`→`errors`→`middleware`) broken with a local `AuthClaims` port +
    `TokenValidator` interface — see "Mockery Mock Self-Import → External Test Package + Interface
    Ports Break Cycles".
  - **#148 008-T030** (PR #163): chain wired in **`cmd/api/server.go`** (not `main.go`); RateLimit
    **off by default** (`RATE_LIMIT_REQUESTS=0`) — no spec mandates a global number and a non-zero
    default would throttle the `startAPIServer` integration tests.
  - **#149/#150 frontend** (PR #164, T031–T036): **no Axios** — the existing fetch-based
    `lib/api-client.ts` meets T031; pure `mapApiError()` extracted to `lib/error-handler.ts` (not the
    `.gitkeep`-only `services/`); `stores/auth-store.ts` kept as-is (`role`/`subscription_id` per
    `docs/data-model.md`, which outranks the row's literal `full_name`/`has_subscription`); 6 routes
    added to the existing `routes/index.tsx` (not a competing `router.tsx`) and `ProtectedRoute` now
    really gates on `useIsAuthenticated()`; 3 focus-indicator tokens added to `tokens.css` with
    `global.css` `:focus-visible` rewired onto them.
  - **Ad-hoc PR #159**: 15 frontend test files restructured to the mandated BDD hierarchy; 19 Go test
    files aligned to TDT + t-first helpers; coverage floor raised 80%→90% (enforced in
    `vitest.config.ts`). CI consolidated 14→9 per-PR jobs; `accessibility.yml` **deleted and folded
    into `frontend-ci.yml`** (job name "Accessibility Audit" preserved — it is pinned in the ruleset).
    See "revive's `context-as-argument` Conflicts with the t-First Test-Helper Convention".
  - **Ad-hoc PR #158**: documentation and test-standards reconciliation with as-built state.
- **Key Decisions**:
  - **Auth middleware shipped available but NOT globally gated**: no login endpoint mints a cookie
    yet, so gating `/api/v1`+swagger now would break tests and lock out staging. `cmd/api/routes.go`'s
    `TODO(sprint-5)` swagger marker **deliberately stays** until Sprint 6.
  - **Two rate limiters is deliberate, not duplication**: email-keyed progressive delay (login) vs.
    per-IP HTTP throttle (middleware). An email-keyed limiter cannot be global HTTP middleware.
  - `pkg/` stays dead by convention — real homes are `internal/database/`, `config/`,
    `backend/migrations/`.
  - **Closure found 36 rows of status drift** (004-T001–T005/T010/T012–T014, 008-T001–T018/T022–T030
    still `Backlog` despite merged PRs) — far wider than the 11 flagged mid-sprint. Every row was
    grep/ls-verified against real code before flipping. Reinforces the `sprint-closure-status-drift-check`
    discipline: the per-PR "flip as part of the closing PR" convention did **not** hold in practice.
  - **"Issue-Body Snippets Are Lowest-Authority" applies to prose hand-off summaries too** — two claims
    in the closure hand-off (a `password/` subpackage; logger/middleware in `internal/security/`) did
    not survive verification against the tree.
  - Path overrides recorded per-row in roadmap Notes (the only channel future issue bodies inherit
    context through): `e2e/tests/` not `e2e/specs/`; `infra/modules/secrets/` not
    `infra/terraform/modules/secrets/`; `eslint.config.js`/`.prettierrc.json` not
    `.eslintrc.json`/`.prettierrc`.
  - Process lesson (recurred): use the `tdd-developer` agent for implementation — inline work needed
    mid-task rework to TDT + mockery conventions (`use-tdd-developer-for-implementation` memory).
- **Verification at closure (re-run, not trusted)**: backend `go build ./...` clean, `go test -short
  ./...` exit 0 (11 pkgs), `golangci-lint` (v1.64.8) exit 0; frontend `type-check`/`lint` clean,
  `test:coverage` 152/152 across 17 files (98.15/92.45/100/98.1), `build` succeeds.
- **Open follow-ups carried into Sprint 6+**:
  1. **004-T011 is genuinely incomplete** — its blocker never resolved: `itinerary_items` exists in no
     migration and no planned task creates it; **#138 was closed with this task unfinished**. Schedule
     the CREATE in Sprint 6+ with `version` folded in inline (as `010_create_trips_table.sql` did),
     not a standalone ALTER. Sibling 004-T010 *is* satisfied (`trips.version` shipped inline).
  2. **`TODO(sprint-5)` in `cmd/api/routes.go`** — Sprint 6 builds the `jwt.Validator`→`AuthClaims`
     adapter in `package main`, applies `Authenticate` to protected groups, removes the marker.
  3. ~~Ruleset "Protect main" (id 18752818) required checks 10 → 6~~ — **NOT a follow-up; verified
     already applied on 2026-07-17.** The ruleset requires exactly the 6 post-#159 checks and all 6
     match the real workflow `name:` values. This was carried forward from `scratch/working-notes.md`
     and re-stated at closure without re-verification, then surfaced to the user twice as blocking.
     **Lesson: verify an inherited "still pending" note against the live system before repeating it**
     — "Issue-Body Snippets Are Lowest-Authority" applies to our own stale notes too. Clearing
     `working-notes.md` at each sprint closure exists precisely to stop this class of zombie item.
  4. Deferred to Sprint 6/7: DB-backed `RefreshTokenStore`, `SubscriptionResolver`, Secrets Manager
     ARN loader (004-T070/T071/T072).
  5. `internal/security/doc.go`'s package comment is **stale** — it promises a logger/middleware that
     were deliberately built in `observability`/`middleware` instead.
  6. `.golangci.yml` `run.build-tags` lists only `integration`, not `test` — files behind
     `//go:build test` are invisible to the linter (pre-existing).
  7. **`BCRYPT_COST` is loaded but never consumed** — same class as the `DB_MIN/MAX_CONNECTIONS` gap
     above, but harmless today because the env value coincides with the hardcoded 12. Documented in
     `backend/README.md`/`docs/security.md`/issue #143; code deliberately left alone.
  8. **Postgres version skew between test and runtime**: testcontainers use `postgres:16-alpine`
     while docker-compose/CI pin `postgres:15.4-alpine` (the deliberate RDS match). Pre-existing;
     needs a deliberate decision, not a doc fix.
  9. `ValidatePassword` has a runes-vs-bytes edge at the 72-char bound (known; documented in
     `docs/security.md`).
  10. Long-standing, still open: 002-T023 a11y gate still `--pass-with-no-tests` (Sprint 9);
     `internal/example/` deletion waits on Sprint 8 (Trip); `postgres:15.4-alpine` staleness watch;
     AWS-cost-avoidance constraint still in force; 009-T018–T026 (swagger-drift CI gate) deferred.

---

## Sprint 6 Implementation (Compacted)

### Sessions: Sprint 6 Planning; Auth Data/Service/HTTP Layers (Models, Payment Stub, Repos, Services, Register/Login/Logout); Config Audit; Auth Activation (Gate + Refresh); Itinerary Migrations; Frontend Auth Plumbing + Session Bootstrap; Register/Login UI; Swagger-Drift + Roadmap-Status-Drift CI Gates
- **Date Range**: 2026-07-17 to 2026-07-31 (closed 2026-08-01 — 58 tasks / 18 work items #167–#184
  plus #195, PRs #186–#202)
- **Key Outcomes**:
  - **Planning** (#167–#184, 2026-07-17): Sprint 6 reshaped from a projected "data layer +
    repositories" (31 tasks) into the full auth vertical + carried-forward remediation (58 tasks /
    18 work items) after tree verification. Pulled ~27 auth tasks forward from Sprints 7-8 so auth
    activates end-to-end rather than sitting inert (repeat of the Sprint 5 "available but not wired"
    anti-pattern would have compounded otherwise). 14 rows marked SUPERSEDED (001-T009–T015
    duplicated already-shipped Sprint 5 config/DB/migrations and Spec 008's own data layer; several
    008-T0XX duplicated Sprint 2 primitives). 004-T011's orphaned migration folded into #170. Scope
    choices made via `AskUserQuestion` ("Vertical auth completo", "Superseded, sin código"). No
    AWS/terraform touched all sprint.
  - **#167+#169 G-008-AUTH-MODELS/SUB-REPO** (PR #186): `auth.User`/Register/Login models,
    `subscription.Subscription` (matches migration 007 — **no period/version columns**;
    register-response `current_period_*` fields deferred to a future ALTER migration),
    `subscription.Repository`/`Resolver`. #169 was pulled into the same branch mid-implementation
    because `RegisterResponse` structurally needs `Subscription`.
  - **#168+#175 G-008-PAYMENT/AUTH-REPO**: `payment.PaymentProvider` + `StubPaymentProvider`
    (always-succeeds, `[DEMO]`-logged). `auth.UserRepository`/`RefreshTokenRepository` with
    **optimistic locking via the `version` column** on `UpdateUser` (stale-version → Conflict,
    unknown-id → NotFound, disambiguated by a follow-up existence check).
  - **#176 G-008-AUTH-SERVICE**: `subscription.Service.CreateSubscription` (charge-then-persist).
    `auth.Service.Register` — **FK-safe ordering**: CreateUser first, subscription second, then
    UpdateUser `has_subscription`. `auth.Service.Login` — **rate-limit check before bcrypt**, and an
    unknown email is treated identically to a wrong password (**no enumeration**). **Incident**: a
    dependency PR showed MERGED on GitHub but `main` had been silently reset behind its merge commit,
    so #176 started on a `main` missing its own dependencies — caught before any #176 code landed,
    fixed via `git merge --ff-only` (pure fast-forward, no history rewrite). Promoted to personal
    memory (`verify-origin-main-has-merged-deps`): a PR marked MERGED does not guarantee `main`
    actually advanced to it.
  - **#177+#178 Auth HTTP Surface** (PR #189): register/login/logout live under `/api/v1/auth/*`,
    HTTP-only Secure SameSite=Strict cookies. New `jwt.Issuer` (mints initial pairs) alongside
    `jwt.Refresher` (rotates only). Composition root `cmd/api/auth.go` generates a **dev-ephemeral
    RSA-2048 key** when `JWT_SIGNING_KEY` is unset (still required in prod). Only register gets
    per-IP rate-limit middleware; login's throttle is the service's email-keyed limiter.
    `REFRESH_TOKEN_EXPIRATION` default corrected 7d→30d to match `docs/security.md`.
  - **#174 Config Audit** (stacked on #189's branch): closed the long-tracked `BCRYPT_COST`
    loaded-but-unconsumed gap — `HashPassword` now takes a `cost`, and a prod floor clamps
    `BCRYPT_COST`<12 up to 12. Every `AuthConfig` field now reaches a real consumer. Fixed
    `internal/security/doc.go`, which had promised JWT/rate-limit/logging implementations that
    actually shipped in `middleware`/`observability`.
  - **#179 Auth Activation** (branched off `main` after confirming #177/#178's and #174's PRs had
    really landed): the `Authenticate` gate is mounted and the `TODO(sprint-5)` marker is gone —
    **auth goes live**. `POST /auth/refresh` (008-T148/T149) was pulled forward so the `Refresher`
    built here has a real caller from day one. Routing fixed per *Chi Panics on a Duplicate Routed
    Pattern* (`patterns-discovered.md`) — `RegisterPublicRoutes`/`RegisterProtectedRoutes` split.
    `securityDefinitions` corrected from a fictitious `Authorization` header scheme to `CookieAuth`.
    A uniform-401 gap (missing-cookie vs. invalid-refresh had different `message` strings) was fixed
    and pinned by a body-equality test (*A "These Two Responses Are Indistinguishable" Comment Needs
    a Test Comparing Them*). `HTTPServer.extraProtectedRoutes` — a test-only hook left in prod
    code — flagged for deletion the moment a real subscription-gated route lands.
  - **#179 follow-up, T150** (2026-07-30, PR #191): added a `token_expired` error code
    (**snake_case**, not the spec's literal `TOKEN_EXPIRED` — the sprint's 4th such casing
    correction) so the SPA can distinguish "no session" from "expired session" on a 401 — a
    deliberate, narrow, safe break of the uniform-401 property (only a validly-signed-but-expired
    token can reach that branch). T160, the frontend consumer, was homed on #180 instead of being
    built here.
  - **#170 Itinerary Migrations** (closes 004-T011): `013–015_create_{destinations,days,
    activities}.sql`, `activities.version` inline (mirrors `trips.version`). **004-T011's two-sprint
    blocker was a false premise** — `itinerary_items` was never in the canonical model (see *A
    Long-Blocked Task Can Be Blocked on a Premise That Was Never True*, `patterns-discovered.md`).
    `docs/data-model.md` gained an As-Built Schema Diagram documenting two divergences from the
    promoted target diagram (no `users.subscription_id`; a Day has zero-or-one Destination, not
    exactly one).
  - **#171+#180 Frontend Auth Plumbing** (PR #194): `Label` primitive; `authApi`/
    `use{Register,Login,Logout}` hooks; the 401→refresh→retry interceptor (**single-flight**, so
    concurrent expiries can't race the one-time-use refresh token), reaching the store through a
    **handler registry, not an import**, navigating with **`window.location`, not the router** (see
    *A Client-Level Interceptor That Must Touch the Store Needs a Handler Registry, Not an Import*).
    **MSW wired for the first time**, the switch `docs/testing-guidelines.md` had deferred until a
    real feature called the backend. Corrected the auth store's `User` type to the real wire shape
    (`role`/`subscription_id` were never actually serialized — dropped; `full_name`/`has_subscription`
    added). A scope audit found and fixed two live defects: `mapApiError` let 409/422/429 fall
    through to a generic retryable error, and the rate-limit countdown wasn't honoring the
    `Retry-After` **header** (register's 429 ships with no body `details`).
  - **#195 `GET /auth/me`** (001-T021, PR #196): reads **from the DB, not token claims**
    (subscription status can go stale within a token's lifetime); **401, not 404**, on a deleted
    account; **exempt from session-expiry teardown** except when the 401 is itself an unrescued
    `token_expired`. Unblocked 008-T120/T121 (#184).
  - **#182+#183 Register/Login UI** (PR #198): real `/register`/`/login` pages replacing
    placeholders. Closed an open-redirect gap in `?redirect=` handling (`resolveLoginRedirect`
    allow-lists a single leading `/`). Client-side password validation deliberately mirrors the
    backend validator rule-for-rule.
  - **#172+#181 `swagger-drift` CI Gate** (PR #200): `backend-ci.yml` gained a swagger-drift check,
    proven both ways with a real throwaway PR (#201, closed same session) that intentionally
    drifted docs and failed CI as designed. 009-T019 (branch-protection automation) stays **Backlog,
    marked Blocked** — GitHub 403s rulesets/branch-protection APIs on a private personal-account
    repo (re-confirmed; 002-T050/#173 was about to hit the identical wall).
  - **#173 Lint/Postgres/roadmap-status-drift** (PR #202): `test` build-tag added to golangci-lint's
    scope (14 real findings fixed). New `internal/testdb.PostgresImage` constant unifies 5
    independent testcontainer call sites on `postgres:15.4-alpine` (the deliberate RDS-matching
    pin). New `scripts/check-roadmap-status-drift.py` + CI step, needing `issues: read` added to the
    workflow's `permissions:` block. **Post-PR fix**: the gate's own first CI run flagged its own
    just-Done rows against a still-OPEN #173 (an issue only closes at merge) — the symmetric
    "open-but-Done" rule was dropped entirely, keeping only "closed-but-not-{Done, Superseded,
    Blocked}" per the ticket's literal scope (*A New CI Gate Must Be Test-Run Against `main`'s
    Actual State, Not Just a Clean Diff*, `patterns-discovered.md`).
- **Key Decisions**:
  - Every subagent-implemented ticket this sprint was independently re-verified by the parent
    session (rebuild/vet/lint/full-suite re-run against real Postgres, diffs read in full) before
    being trusted — caught real gaps multiple times (a missing test, stale docs, the #176 dangling-
    merge incident above).
  - `technical-writer` review proved genuinely load-bearing, not ceremonial: caught a real defect
    (`Number('') === 0` making an empty `Retry-After` header inject `retry_after_seconds: 0`) and
    several overstated comments across the sprint.
  - Roadmap-row Notes remained the load-bearing channel for scope decisions reaching not-yet-created
    tickets (superseded rows, casing corrections, path deviations) — *Roadmap Row Notes Are the Only
    Channel Future Issue Bodies Inherit Context Through* was reused repeatedly this sprint.
- **Post-closure follow-up** (2026-08-01, issue #192, PR #204, 3 commits): fixed all three items
  flagged above at #179's close — `/auth/refresh` now returns 503 (not 500) on a DB outage via new
  `errors.ServiceUnavailableFromDB`; `handleRegister`/`handleLogin` gained `@Failure 500`;
  `/swagger/*` is now auth-gated only when `Config.IsProduction()` (fixes the cold-start 401).
  `backend/docs/` regenerated via `make swagger` in the same commits. **Gap**: issue #192 itself
  said item B should "route through `technical-writer` per the API-doc policy" — `gh pr view 204`
  shows no reviews/comments, so that review appears to have been skipped.
- **`@Failure 429` gap fixed** (2026-08-01, same day as closure): `handleRegister` was missing
  `@Failure 429` despite `middleware.RateLimit(10/min)` making it a genuine possible response —
  found during #195's session but never folded into #192's scope (#192 only covered `@Failure
  500`). Annotation added, `backend/docs/` regenerated via `make swagger`, build/vet clean.
  Uncommitted pending `technical-writer` review (see below).
- **Open follow-ups carried into Sprint 7+**:
  1. `HTTPServer.extraProtectedRoutes` test-only hook — delete the moment a real subscription-gated
     route lands.
  2. Register-response `current_period_start`/`current_period_end` fields need a future ALTER
     migration on `subscriptions` (migration 007 has no period columns).
  3. 009-T019 / 002-T050 both permanently Blocked on the same GitHub-plan wall (rulesets/
     branch-protection require Pro or a public repo) — not fixable by more engineering.
  4. Long-standing, still open: 002-T023 a11y gate `--pass-with-no-tests` (Sprint 9);
     `internal/example/` deletion waits on Sprint 8 (Trip); AWS-cost-avoidance constraint still in
     force.

---

## Sprint 7 Implementation (Compacted)

### Sessions: Sprint 7 Re-planning; GuestRoute + Badge Primitives; Repo-Wide Comment-Quality Sweep (Backend, Frontend, E2E, Infra)
- **Date Range**: 2026-08-01 to 2026-08-02 (closed 2026-08-02 — 5 work items, issues #218–#222, PRs
  #223–#229)
- **Key Outcomes**:
  - **Re-planning pass** (PR #223, 2026-08-01): the drafted 28-task Sprint 7 (Spec 001 Phase 2 +
    Spec 008 Phase 3 Part 2) was found 26/28 tasks already Done/Superseded (Sprint 2/5/6 leakage) or
    correctly re-scoped away this pass. The 6 Trip-dependent frontend tasks (`useTrips`, `TripCard`,
    `TripDashboard`, etc.) plus their 4 backend-stub counterparts (008-T048–T051) moved to Sprint 8 —
    no Trip backend exists yet to call, and building against a mock would be discarded the same
    sprint the real endpoint lands (repeats the "available but not wired" anti-pattern this project
    already rejected once in Sprint 6). A cross-spec duplicate-work risk was flagged for Sprint 8's
    own planning, not resolved here: Spec 008's `internal/collaboration/` Trip stub vs. Spec 001
    Phase 3's real AI-backed `internal/trip/` Trip backend, and Spec 008's `TripCard` vs. Spec 001's
    own `TripCard`. 001-T030 (stub Checkout page) was deliberately left unscheduled — the Sprint 6
    registration flow never grew a caller for a standalone checkout page (payment is collected
    inline via a demo-token checkbox), and the row likely overlaps 008-T186 (Post-MVP); flagged as an
    open product question, not decided unilaterally. Real remaining spec-sourced scope came out to
    one 2-task work item (`GuestRoute` + `Badge`); user confirmed accepting a deliberately light
    sprint rather than pulling Sprint 8 scope forward. User separately requested a repo-wide code
    comment-quality audit (not derived from any spec), scoped as 4 standalone work items per
    `.github/SPRINT-CONSOLIDATION-CHECKLIST.md`'s explicit "different tech stacks stay separate"
    rule (Go, React/TypeScript, Playwright/TypeScript, Terraform/HCL).
  - **#218 G-SPRINT7-FRONTEND-SHELL-GAPS** (PR #224): `frontend/src/components/GuestRoute.tsx`
    (redirects authenticated users away from `/login`/`/register` to `/dashboard`, mirroring
    `ProtectedRoute`'s existing `useIsAuthenticated()` gate) + `Badge` primitive
    (`components/primitives/Badge.tsx`, variant-based, design-token-driven). Closed both halves of
    001-T027 and 001-T028; the roadmap's own note flags this pre-empts 008-T083 (Phase 4 US2), which
    should be marked Superseded when reached.
  - **#219 G-SPRINT7-COMMENT-AUDIT-BACKEND** (PR #225): swept `backend/` (Go) — corrected stale
    package `doc.go` comments (`auth/jwt`, `concurrency`, `observability` had drifted from what they
    actually implement), documented three loaded-but-unconsumed config fields, updated
    `backend/README.md`.
  - **#220 G-SPRINT7-COMMENT-AUDIT-FRONTEND** (PR #228): swept all 78 `.ts`/`.tsx` files in
    `frontend/src/` — already exceptionally clean (no restating/overlong comments found). One defect
    surfaced: `lib/auth.ts` and `lib/authContext.tsx` are dead Spec-004 scaffolding placeholders
    (their functionality shipped elsewhere as `stores/auth-store.ts` and `lib/api-client.ts`'s
    refresh interceptor) with zero remaining imports — flagged rather than deleted (out of scope for
    a comment-only pass), tracked at closure as its own issue, #231 (no gating condition, unlike
    #210's Sprint-8-triggered cleanup).
  - **#221 G-SPRINT7-COMMENT-AUDIT-E2E** (PR #227): swept `e2e/tests/` and `e2e/fixtures/` — only
    `e2e/tests/sample.spec.ts` had real content (four restating comments removed, header tightened);
    the rest of the tree is `.gitkeep`-only placeholders, correctly left untouched.
  - **#222 G-SPRINT7-COMMENT-AUDIT-INFRA** (PR #229): swept `infra/` (Terraform/HCL) — most of the 23
    `.tf`/`.tfvars` files were already clean; `backend.tf`/`versions.tf` had restating comments
    removed (also fixing an S3-state-key example that described the local-backend path format
    instead of the real one); `modules/ecs/{main,outputs}.tf` had stale "wiring not done yet"
    comments corrected now that root-module wiring (005-T102/#90) actually shipped;
    `modules/rds/outputs.tf` and `modules/secrets/outputs.tf` had comments falsely claiming their
    outputs are consumed downstream — verified via grep that neither actually is, corrected to point
    at the real open follow-up (003-T033 / G-INFRA-IAM-MODULE, Backlog).
- **Key Decisions**:
  - Consolidation was applied in both directions this sprint: the two genuinely-open spec-sourced
    tasks (`GuestRoute`, `Badge`) were merged into one work item per the checklist's normal rules,
    while the four comment-audit tasks were deliberately kept as four *separate* work items because
    they span four different tech stacks — the checklist's own "DON'T consolidate" example.
  - The comment-quality rubric applied consistently across all four sweeps: remove or rewrite
    comments that restate what the code already shows, are excessively long relative to what they
    explain, or add nothing beyond the obvious — keep only comments explaining a non-obvious WHY
    (hidden constraint, subtle invariant, workaround). All four passes were read-only/no-behavior-
    change by design; the backend and infra sweeps additionally corrected factually stale claims
    they encountered along the way (drifted `doc.go` comments, drifted wiring-status comments),
    which is a natural side effect of actually reading the code closely, not scope creep.
- **Sprint 7 closure** (2026-08-02): one narrative-note drift found and fixed — #222's roadmap Notes
  cell still said "PR open" after PR #229 had already merged (the automated
  `check-roadmap-status-drift.py` gate only checks the Status column, not free-text Notes, so this
  class of drift needs a manual look at closure, not just the script). `gh issue list --state open`
  re-confirmed exactly 3 items (#209, #210, #212), none targeting Sprint 7. Two items carried
  forward, neither blocking closure: 001-T030 still awaits a product decision (see re-planning notes
  above); the #220 dead-code finding got its own tracking issue, #231.
