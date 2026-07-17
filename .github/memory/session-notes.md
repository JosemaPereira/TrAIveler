# Session Notes

Historical summaries of completed development sessions. Committed to git as a record.

**Note**: Older sessions (specs 001-007, sprint infrastructure setup) are compacted to save tokens. Policy: each sprint's implementation sessions are compacted into a `(Compacted)` summary immediately at that sprint's own closure — the same pass that updates the sprint-closure documentation. At any given time, only the sprint currently in progress (not yet closed) has a `(Detailed)` section; the moment it closes, it gets compacted, no rolling window. Compacted sections favor brevity over completeness: PR/issue/Group IDs, decisions with lasting effect, and still-open follow-ups only — narrative process detail is cut, and duplicates of `patterns-discovered.md` entries are referenced by name instead of re-explained. Sprint 1 and Sprint 2 were compacted together on 2026-07-11; Sprint 3 was compacted at its own closure on 2026-07-12 (and re-tightened for conciseness the same day, which also lightly trimmed the Sprint 1/2 sections); Sprint 4 was compacted at its own closure on 2026-07-14.

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

## Sprint 5 Implementation (Detailed)

### Session: G-008-DATABASE — PostgreSQL Pooling & Goose Migrations Config
- **Date**: 2026-07-15
- **Tool**: Claude Code
- **What was accomplished**: Issue #141 (G-008-DATABASE, 008-T009/T010). Verified
  `internal/database/client.go` already satisfied T009's pooling requirement — no new file created,
  avoiding the duplicate-implementation risk the issue itself flagged. For T010, added
  `internal/database/migrations` (`Dir`, `SetDialect()`), replacing three duplicated
  `const migrationsDir = "../../migrations"` + `goose.SetDialect("postgres")` call sites
  (`tests/integration/error_test.go`, `tests/integration/security_migrations_test.go`,
  `internal/example/repository_integration_test.go`) with one `runtime.Caller`-resolved absolute
  path. While verifying T009, found and fixed a real config-wiring gap (see
  `patterns-discovered.md`, "A Config Value Can Be Loaded, Validated, and Logged, Yet Still Never
  Reach the Code It Configures"): `NewClient` ignored `DB_MIN_CONNECTIONS`/`DB_MAX_CONNECTIONS`
  entirely, hardcoding 5/25 despite `config.go` already loading/validating them and `main.go`
  logging them as if applied. `NewClient` now takes `minConns, maxConns int`; all call sites
  (`main.go`, `client_test.go`, `repository_integration_test.go`) updated; new
  `TestNewClient_ConfigurablePoolSize` asserts the pool's real `Pool().Config()` against non-default
  values (2/10) so the assertion can't pass by coincidence. No ticket existed for this gap; fixed in
  the same branch/PR as #141 per explicit user instruction rather than filed separately.
- **Key findings and decisions**: Migration `.sql` files stayed in `backend/migrations/`, not moved
  to the task's literal `pkg/database/migrations/` path — `pkg/` remains dead per established
  convention, and the new shared config package was placed under `internal/database/migrations/`
  for the same reason. Confirmed no mockery regen needed: `database.Client`'s interface shape didn't
  change, only the free-function `NewClient` constructor's parameters. Branch was initially created
  as `feature/142-...` (wrong issue number — that name belongs to the separate, not-yet-started
  G-008-MIGRATIONS ticket) and corrected mid-session: created the correctly-named
  `feature/141-g-008-database-postgresql-pooling-goose-migrations-config` from the same commit and
  deleted the local `feature/142-...` alias (had zero unique commits; remote `feature/142-...`
  untouched).
- **Outcomes**: `go build`/`go vet`/`gofmt`/`golangci-lint` clean; full test suite
  (`internal/database`, `internal/database/migrations`, `internal/example`, `tests/integration`,
  `config`, `cmd`) passes, including Colima-backed testcontainer tests. Not yet committed/pushed/
  PR'd as of this entry.

## 2026-07-16 — Test-suite BDD/TDT alignment + 90% coverage floor (ad-hoc)

**Tool**: Claude Code

**Key Outcomes**
- All 15 frontend test files restructured into the mandated BDD hierarchy
  (`describe('<Component />')` → `describe('when/having …')` → `it('should …')`, AAA with blank
  lines); the legacy `unit:` describe prefix dropped. 136/136 tests green; ESLint + Prettier clean.
- Coverage floor raised 80%→90%: enforced in `frontend/vitest.config.ts` (`coverage.thresholds`,
  actual 98/92.2/100/97.9 — passes; supersedes roadmap `002-T041`'s enforcement gap), documented in
  `docs/testing-guidelines.md` (with an honest note that the last observed backend number, 89.1%
  full-suite for `internal/example`, now sits ~1pt under the new floor), mirrored in `CLAUDE.md` +
  `.github/copilot-instructions.md`.
- 19 Go test files aligned: TDT consolidations (auth/ValidatePassword, observability/
  LogSecurityEvent, ai sanitizer/validator/NewAIClient, errors constructors, config helpers);
  behavioral `t.Run` names ("when … it should …"); 13 helpers normalized to t-first +
  `t.Helper()` (incl. `setupPostgresContainer`, `setupRepositoryTestDB`, `startPostgresContainer`,
  `startAPIServer`, `holdExclusiveTableLock`, `openMigrationDB`); repeated
  `defer pgContainer.Terminate` blocks folded into helpers via `t.Cleanup`; hand-rolled
  `contains`/`findSubstring` in config_test.go replaced with `strings.Contains`.
- Verified for real, not assumed: backend `make lint` clean, `go test -tags=test ./... -short`
  all `ok`, AND the full Colima-backed suite (`go test -tags=test ./...` with `DOCKER_HOST` +
  `TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE`) all `ok`.

**Key Decisions**
- revive `context-as-argument` vs the documented t-first helper convention resolved in favor of the
  docs: new `_test\.go`-scoped exclude-rule in `backend/.golangci.yml` (pattern captured in
  `patterns-discovered.md`, 2026-07-16). Production code keeps ctx-first enforcement.
- `writeAnthropicError`/`writeAnthropicSuccess` deliberately stay t-free — they run inside fake-server
  handler goroutines where `t.Helper()` has no effect.
- `service_test.go`, `ollama_client_test.go`, middleware tests left as flat behavioral test funcs —
  distinct per-test setups make TDT a net readability loss there; the standard's TDT mandate is for
  repetitive/conditional cases.
- Nothing committed/pushed per explicit instruction; ~42 files modified in the working tree.

**Open Follow-Ups**
- Backend `internal/example` coverage (89.1% full-suite) is now marginally under the new 90% floor.
- `.golangci.yml` `run.build-tags` only lists `integration`, not `test` — files behind
  `//go:build test` (ai, example handler/service, cmd/api tests) are invisible to golangci-lint;
  pre-existing gap, worth a deliberate fix.

### Session: G-008-JWT — JWT Generator, Validator (Multi-Key Rotation), Refresher
- **Date**: 2026-07-16
- **Tool**: Claude Code
- **What was accomplished**: Issue #144 (008-T019/T020/T021). New subpackage
  `backend/internal/auth/jwt/`: `Generator` (RS256 access tokens, 24h default, `has_subscription`
  claim + `kid` header), `Validator` (multi-key rotation via per-token `kid` selection; RS256-locked
  parser; issuer+expiry required), `Refresher` (one-time-use refresh-token rotation: validate →
  revoke old → issue new access/refresh pair). Support: `KeyProvider` interface + `StaticKeyProvider`
  + `LoadKeyFromPEM` (PKCS#1/#8, 2048-bit min); `RefreshTokenStore` + `SubscriptionResolver`
  interfaces (impls deferred); opaque 32-byte refresh tokens stored as SHA-256 hashes. 92.4% coverage,
  lint clean, full backend short suite green. Docs updated: `docs/security.md` status note,
  `specs/008-auth-collaboration-ux/tasks.md` T019-T021 checked, roadmap rows Backlog→Done.
- **Key findings and decisions**:
  - **Chose a subpackage `internal/auth/jwt` over the flat-package precedent** (scope note leaned
    flat): cohesive sub-domain with many exported types, matches the original tasks.md path, sidesteps
    the `validator.go` filename clash with the password validator. Cost: `golang-jwt/jwt/v5` imported
    as `gojwt` to avoid the package-name collision.
  - Secrets model follows #143's raw-now decision: private key arrives as raw PEM via config
    (`JWT_SIGNING_KEY`); `LoadKeyFromPEM` is the raw path. ARN-resolving loader + DB-backed
    `RefreshTokenStore` are Spec 004 Phase 3 / Sprint 6-7 — this issue ships the interfaces + logic +
    static/in-memory impls only, so it's fully testable without AWS or a DB.
  - Refresher resolves `has_subscription` fresh at refresh time via `SubscriptionResolver` rather than
    copying the stale claim from the old token.
  - Security-hardening tests included: alg-confusion (`alg:none` and wrong-key signature) rejection,
    retired-key rejection, revoked-token-reuse rejection — all collapse to a uniform
    `authentication_required` domain error (cause wrapped for logs, not leaked).
  - Test timestamp gotcha: parse round-trip tokens with `gojwt.WithTimeFunc` pinned to the generator's
    injected `now`, and compare `time.Time` with `.Equal` not `assert.Equal` (NumericDate decodes to
    Local zone).
- **Outcomes**: JWT token primitives exist and are unit-tested; they feed the auth middleware (#147)
  whose wiring (#148) will remove the `TODO(sprint-5)` Swagger-gating marker in `cmd/api/routes.go`.

### Session: G-008 Sprint-5 Backend Middleware Chain (#146 + #147 + #148)

- **Date**: 2026-07-16
- **Tool**: Claude Code
- **What was accomplished**: Three grouped issues on one branch
  (`feature/008-sprint5-backend-middleware-chain`).
  - **#146 (T024/T025)** — new subpackage `internal/auth/ratelimit`: `Limiter` (progressive delay
    `2^(attempts-5)`s after 5 failures, clamped to the 15-min window per data-model.md:128) over an
    in-memory, concurrency-safe TTL `store` keyed by email. `store.go`/`limiter.go` split matches the
    two task IDs. Consumer is the future login service (T098), not middleware.
  - **#147 (T027 + T029)** — in `internal/middleware`: `Authenticate` (validates `access_token`
    cookie, attaches user id + `has_subscription` to context, 401 `authentication_required`
    otherwise) and `RateLimit(limit, window)` (per-IP fixed window, `X-RateLimit-*` headers,
    429 + `Retry-After`). T026/T028 confirmed already-satisfied (skipped).
  - **#148 (T030)** — `RateLimit` inserted into the chain in `cmd/api/server.go` after Recovery,
    gated on new `config.RateLimitConfig` (`RATE_LIMIT_REQUESTS`/`_WINDOW`, **default disabled**).
  - Full backend short suite green, `-race` clean on both stores, lint clean.
- **Key findings and decisions**:
  - **Two design forks resolved with the user** (both "recommended" options taken): (1) auth
    middleware is shipped **available but NOT globally gated** — no login endpoint exists yet to mint
    a cookie, so gating `/api/v1`+swagger now would break existing tests and lock out staging;
    `cmd/api/routes.go`'s `TODO(sprint-5)` swagger marker intentionally **stays** until Sprint 6.
    (2) Rate limiting is **two mechanisms**: per-IP `RateLimit` middleware (this issue) + email-based
    progressive-delay `Limiter` for login (#146). #146's "consumer is T029" reconciliation note does
    NOT hold — an email-keyed limiter can't be global HTTP middleware; its real consumer is T098.
  - **Auth middleware lives in `internal/middleware`, not `internal/security`** — reuses the existing
    `ctxKeyUserID`/`UserIDFromContext` (`user_context.go`) written for "a future JWT auth middleware".
    Added `ctxKeyHasSubscription`/`HasSubscriptionFromContext` alongside it.
  - **Import-cycle avoidance**: `middleware` can't import `auth/jwt` (jwt→errors→middleware cycle).
    Solved with a middleware-local `AuthClaims` port + `TokenValidator` interface; the real
    `*jwt.Validator`→`AuthClaims` adapter is written in `package main` at Sprint-6 wiring time.
  - **Global rate limit ships disabled** (`Requests: 0`): no spec mandates a global number, and a
    non-zero default would risk throttling the `startAPIServer` integration tests. Per-endpoint
    limits (10/min login etc.) come later as route-level middleware.
- **Process note**: implemented inline and had to correct tests mid-task to TDT + mockery
  (`MockTokenValidator` in `internal/middleware/mocks`, external `middleware_test` package to dodge
  the mock's self-import cycle). Lesson: use the `tdd-developer` agent (reads coding/testing/mock
  docs first) for implementation work to avoid this rework.
- **Outcomes**: Sprint-5 backend middleware chain complete. Remaining open Sprint-5 issues are the
  frontend pair (#149 + #150). Sprint-6 handlers will construct a `jwt.Validator`→`AuthClaims`
  adapter, apply `Authenticate` to protected route groups, and remove the swagger `TODO(sprint-5)`.

### Session: G-008 Sprint-5 Frontend Infra (#149) + Styles (#150)

- **Date**: 2026-07-17 (implementation started 2026-07-16)
- **Tool**: Claude Code
- **What was accomplished**: Both remaining Sprint-5 frontend issues on one branch
  (`feature/008-sprint5-frontend-infra-and-styles`), commit `18c695c`, PR #164 (open). 17 files,
  +458/-71. Mostly reconciliation — most tasks were already partly satisfied by Sprint 2/4 code,
  verified against the real files first.
  - **#150 G-008-FRONTEND-STYLES (T034/T035)**: color/spacing/typography tokens already existed
    (Sprint 2, PR #79). Genuinely-new scope = 3 focus-indicator tokens
    (`--focus-outline-width`/`-offset`/`-color`) added to `frontend/src/styles/tokens.css`;
    `global.css` `:focus-visible` rewired from hardcoded `2px`/`--color-primary`/`2px` onto them.
  - **#149 G-008-FRONTEND-INFRA (T031/T032/T033/T036)**:
    - **T031** (HTTP client): satisfied as-is — **NO Axios added**; the existing fetch-based
      `lib/api-client.ts` already meets it (adding Axios would fragment the HTTP layer). No code.
    - **T032** (error handler): extracted a pure `mapApiError()` into new `lib/error-handler.ts`
      (placed in `lib/`, where error/HTTP utils live — NOT the `.gitkeep`-only `services/`);
      refactored `hooks/useErrorHandler.ts` to consume it, hook keeps the 401→`/login` side effect.
    - **T033** (auth store): satisfied as-is by `stores/auth-store.ts` (Sprint 2) — kept the
      `role`/`subscription_id` shape from `docs/data-model.md`'s User entity (canonical, higher
      authority than the row's literal `full_name`/`has_subscription`, which is the not-yet-landed
      008-T011 alt users scheme, data-model.md:753). No field change.
    - **T036** (router): the real new work — 6 routes added to the existing `routes/index.tsx`
      (NOT a competing `router.tsx`: `App.tsx` already imports `router` from `./routes`, real code
      wins over the literal path). 6 placeholder page components under `routes/`.
      `components/ProtectedRoute.tsx` now genuinely gates on `useIsAuthenticated()` (redirect
      `/login`, `replace`), replacing the unconditional `<Outlet/>` scaffold — single wrapper
      shared with Spec 004 per issue #149.
- **Key findings and decisions**:
  - Implemented inline (not via the `tdd-developer` agent) but followed the coding/testing
    conventions upfront: BDD test hierarchy (`describe('<Component />')`→`describe('when …')`→
    `it('should …')`), Zustand `setState` reset in `beforeEach`, kebab-case filenames. The
    `use-tdd-developer-for-implementation` memory allows this "read docs upfront" alternative; the
    frontend rework risk (no mockery, RTL-only) was low.
  - Followed the "Issue-Body Snippets Are Lowest-Authority" pattern throughout — every literal
    task path/field (`services/api.ts`, `router.tsx`, `full_name`/`has_subscription`) was
    overridden by real code + `data-model.md`, and the override recorded in each roadmap row's
    Notes (the "Roadmap Row Notes are the only channel future issue bodies inherit context
    through" pattern).
  - Status-flip convention: `tasks.md` T031-T036 → `[x]` and roadmap Status Backlog→Done applied
    in the implementing change set (the one that became PR #164), consistent with the Sprint-4
    "flip as part of the closing PR" convention and the JWT #144 precedent.
  - Delegated commit-and-push and open-pr to their subagents (user-invoked); commit footer carries
    `Closes #149`/`Closes #150`.
- **Outcomes**: Sprint-5 frontend track complete (#149 + #150), all 15 Sprint-5 work items now
  implemented. Frontend verified for real: `type-check`/`lint`/`prettier --check`/`build` clean,
  `test:coverage` 152/152 pass (was 136; +16), coverage 98.15/92.45/100/98.1 (>90% floor). PR #164
  open, awaiting review — not merged.
- **Open follow-up flagged (not mine, pre-existing)**: `tasks.md`/roadmap rows 008-T022–T030
  (backend middleware) still `Backlog`/`[ ]` despite PR #163 merged — status drift to reconcile at
  Sprint-5 closure (`sprint-closure-status-drift-check` memory).
