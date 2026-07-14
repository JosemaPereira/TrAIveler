# Session Notes

Historical summaries of completed development sessions. Committed to git as a record.

**Note**: Older sessions (specs 001-007, sprint infrastructure setup) are compacted to save tokens. Policy: each sprint's implementation sessions are compacted into a `(Compacted)` summary immediately at that sprint's own closure — the same pass that updates the sprint-closure documentation. At any given time, only the sprint currently in progress (not yet closed) has a `(Detailed)` section; the moment it closes, it gets compacted, no rolling window. Compacted sections favor brevity over completeness: PR/issue/Group IDs, decisions with lasting effect, and still-open follow-ups only — narrative process detail is cut, and duplicates of `patterns-discovered.md` entries are referenced by name instead of re-explained. Sprint 1 and Sprint 2 were compacted together on 2026-07-11; Sprint 3 was compacted at its own closure on 2026-07-12 (and re-tightened for conciseness the same day, which also lightly trimmed the Sprint 1/2 sections).

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

## Sprint 4 Implementation (Detailed)

### Session: G-OBS-LOGGER — Logger Middleware Structured Log Fields
- **Date**: 2026-07-13
- **Tool**: Claude Code
- **Outcome**: Issue #109 (002-T008/T009), PR #125 merged. `backend/internal/middleware/logger.go` extended
  with `service`/`msg`/`user_id` fields on top of the Sprint 2 baseline (005-T025, PR #71), reconciling
  it with spec 002's `StructuredLogEntry` contract. `logger_test.go` extended to cover the new fields.
- **Note**: this entry was written retroactively during the next ticket's session (G-OBS-HEALTHZ, below) —
  `docs/roadmap.md` rows 002-T008/T009 were left at `Backlog` after PR #125 merged instead of being
  flipped to `Done` immediately. Going forward, flip roadmap status + append this log entry as part of
  the same PR that closes the ticket, not deferred to a later session — see the `commit-and-push`/
  `open-pr` workflow note below.

### Session: G-OBS-HEALTHZ — Healthz Handler Reconciled to HealthCheckResponse Schema
- **Date**: 2026-07-13
- **Tool**: Claude Code
- **Outcome**: Issue #110 (002-T010/T011). `backend/cmd/api/server.go`'s `/healthz` handler reconciled
  from its Sprint 2 schema (`status: "healthy"/"unhealthy"`, `database`, `error`, 503 on ping failure)
  to spec 002's `HealthCheckResponse` (`status: "ok"/"degraded"`, `version`, `uptime_seconds`, always
  `200 OK` per the spec's validation rule — DB ping failure now maps to `status:"degraded"` plus a
  `logger.Warn` instead of a non-2xx response). `version` is a new single-source-of-truth package var
  in `main`, also replacing a hardcoded literal in `main.go`'s startup log. `server_test.go`'s three
  healthz tests updated to the new contract; all other tests in the file unaffected.
- **Key decision**: the "always 200, only the body signals degraded" behavior is a deliberate reversal
  of typical REST health-check convention (503-on-unhealthy) — driven entirely by the literal text of
  `specs/002-nfr-system-constraints/data-model.md`'s `HealthCheckResponse` validation rules, not a
  judgment call. See `patterns-discovered.md` for the generalized lesson.
- **Workflow note**: this session closed out the `docs/roadmap.md` staleness gap left by the previous
  G-OBS-LOGGER session (see above) as part of its own commit sequence, per an explicit user instruction
  to update session-notes/memory/roadmap per ticket rather than batching at sprint close — the pattern
  from Sprint 1-3 was to reconcile the roadmap only at sprint-closure sessions, which is why #109/#125
  slipped for one ticket-cycle. Sprint 4 tickets should each self-close their own roadmap row instead.

### Session: G-ARCH-INTEGRATION-BACKEND — Correlation ID, 5xx Stack Trace, Anthropic Client, Retry-After
- **Date**: 2026-07-13
- **Tool**: Claude Code
- **Outcome**: Issue #117 (005-T110–T113), on branch
  `feature/117-g-arch-integration-backend-backend-integration-patterns-correlation-id-retrybackoff-anthropic-client`
  (branched off `main` post-#127 merge). 005-T110 (correlation ID in error responses) was already
  fully satisfied by PR #74 — verified via existing tests, no code touched. New work: (1) 005-T111 —
  `middleware/logger.go`'s `Logger` now logs a `runtime/debug.Stack()` field for any 5xx response, not
  just panics (`Recovery` already covered those); (2) 005-T112 — new `internal/ai/anthropic.go`,
  `AnthropicClient` implementing `AIClient` on `github.com/anthropics/anthropic-sdk-go` (new direct
  dependency, v1.57.0), 60s timeout, retry/backoff delegated to the SDK's own
  `option.WithMaxRetries`/`option.WithRequestTimeout` rather than a hand-rolled loop (see
  `patterns-discovered.md`, "Prefer an Official SDK's Built-In Retry Over Hand-Rolling One"); a
  429/503 surviving every retry becomes `*ai.ProviderUnavailableError`; (3) 005-T113 —
  `internal/ai/client.go` gained `NewAIClient(cfg config.AIConfig) (AIClient, error)`, a
  provider-agnostic factory switching on `cfg.Provider`, and `TranslateError`, which turns a
  provider-unavailable failure into a `"service_unavailable"` `*errors.DomainError` (new
  `errors.ServiceUnavailable` constructor) carrying a `Retry-After` HTTP header
  (`internal/errors/handler.go`'s `writeErrorResponse`).
- **Key decision**: the user explicitly required, while scoping this issue, that `NewAIClient` NOT be
  built as if Anthropic were the primary/default option with Ollama as a dev-only workaround — both
  provider branches are equally first-class in the switch, matching the existing `AI_PROVIDER`-driven
  split from issue #55 (`sprint2-ai-client-ollama-mvp` in Claude's personal memory has the full
  history). No HTTP route wires `NewAIClient` in yet (`cmd/api/main.go` untouched) — that's deliberately
  out of scope for this issue; a future handler ticket will consume the factory.
- **Mockery question, answered**: user asked whether the new tests warranted (re)generating mockery
  mocks. Confirmed empirically (`make mocks` produces a byte-identical `ai_client_mock.go`, zero diff)
  that no regeneration was needed — the `AIClient` interface itself didn't change. `anthropic.go`
  (the implementation *under test*) is faked at the HTTP layer (`option.WithBaseURL`/`httptest`,
  mirroring `ollama_client_test.go`'s existing pattern) rather than mocked, since you don't mock the
  thing you're testing; `NewAIClient`/`TranslateError` are pure logic needing no mock at all. Reaffirms
  the Sprint 2 "mock any interface a test needs to fake across its own package's dependency boundary"
  rule rather than adding a new one.
- **Documentation pass** (technical-writer, same session): fixed three genuinely stale spots exposed by
  this change — `backend/README.md`'s `DomainError` constructor count/status table (missing
  `ServiceUnavailable`/503), `middleware/logger.go`'s `Logger` doc comment (didn't mention the new
  stack-trace field), and `docs/local-ai-setup.md` (two spots still called `AnthropicClient`/
  `NewAIClient` "future"/"not yet implemented"). Everything else checked was already accurate — see
  the technical-writer's file-by-file confirmation in-session; flagged one genuine pre-existing,
  out-of-scope drift for later: `specs/005-system-architecture/data-model.md`'s `AIClient` interface
  snippet still shows the old `GenerateItinerary(ctx, prompt string)`/`ItineraryChunk` shape instead of
  the real `ItineraryRequest`/`StreamChunk` — predates #117, not caused by it.
- **Workflow note**: `docs/roadmap.md`'s 005-T110–T113 rows are deliberately left at `Backlog` in this
  commit — per the G-OBS-HEALTHZ session's established convention, the Backlog→Done flip (with
  "Closed by PR #NN") happens once the PR exists/merges, not at commit time on the feature branch.
  Whoever runs `open-pr`/closes this ticket next should flip those four rows.

### Session: G-ARCH-INTEGRATION-FRONTEND — X-Request-ID, useErrorHandler, ErrorMessage Correlation ID
- **Date**: 2026-07-13
- **Tool**: Claude Code
- **Outcome**: Issue #118 (005-T114–T116), frontend counterpart to #117, on branch
  `feature/118-g-arch-integration-frontend-frontend-integration-patterns-x-request-id-useerrorhandler-errormessage`
  (branched off `main` post-#128 merge; `origin` already had an auto-created same-name branch with
  zero diff from `main`, so no divergence to reconcile). 005-T114 (`X-Request-ID` header +
  `requestId` extraction in `frontend/src/lib/api-client.ts`) was already fully satisfied by PR #81
  (Sprint 2) — verified via existing tests, no code touched, mirroring how #117 found 005-T110
  already done on the backend side. New work (tdd-developer, strict RED-GREEN): (1) 005-T115 —
  `frontend/src/hooks/useErrorHandler.ts`, the **first hook** in `src/hooks/` (directory didn't
  exist before). Takes `unknown`, returns `{ title, message, requestId?, isRetryable } | null`;
  401 triggers a `useNavigate('/login')` redirect inside `useEffect` (React Router forbids
  navigating during render), 403/404 return non-retryable display info, 500/503/anything-else/
  non-`APIError` falls back to a retryable generic result; uses the existing `isAPIError` helper
  from `query-client.ts` rather than a fresh `instanceof` check. (2) 005-T116 —
  `ErrorMessage.tsx` gained an optional `requestId?: string` prop, rendered as a de-emphasized
  "Reference ID: ..." line (`data-testid="error-request-id"`), gated on truthiness so `''` (a real
  possible value from `APIError.requestId`, not just `undefined`) renders nothing; new `.requestId`
  CSS class in `ErrorMessage.module.css` using existing tokens (`--font-size-sm`,
  `--color-neutral-600`). Neither the hook nor the requestId prop is wired into any real page yet —
  deliberately out of scope, since no route in this codebase makes a live API call yet (auth/trip
  pages are all future-sprint work).
- **Independent verification** (main session, not just the subagent's self-report): re-read every
  changed file, re-ran `npm run lint` / `npm run type-check` / `npm test` myself from a clean
  invocation (136/136 tests across 15 files, zero lint/type errors) both right after
  tdd-developer's pass and again after technical-writer's docstring edit to `ErrorMessage.tsx`, and
  grepped `tokens.css` directly to confirm `--color-neutral-600` is a real token, not a guess.
- **Test-authoring note**: `useErrorHandler.test.tsx` needed a router-context wrapper (`MemoryRouter`)
  plus a `vi.mock('react-router', ...)` with `importActual` to stub just `useNavigate` as a spy —
  no existing repo precedent for testing a hook that calls `useNavigate` (`ErrorBoundary.test.tsx`
  uses `window.location.assign`, a different mechanism; `routes/index.test.tsx` renders a router but
  never asserts navigation calls). This is a new, reusable idiom for the next hook that needs to
  assert on `useNavigate`.
- **Documentation pass** (technical-writer, same session): `frontend/README.md`'s Implementation
  Status callout and Project Structure "Current"/"Target" trees updated — `hooks/` is no longer
  purely aspirational now that `useErrorHandler.ts` is the first real entry; also fixed a stale
  coverage-thresholds sentence that said thresholds would apply "once those directories exist" even
  though `src/components/` and `src/hooks/` both already exist today (the thresholds themselves are
  still unconfigured, tracked separately as roadmap 002-T041). `ErrorMessage.tsx`'s docstring
  enriched to describe the new `requestId` prop's behavior. **Flagged, not fixed**:
  `specs/005-system-architecture/contracts/frontend-patterns.md` Patterns 1/2/4/7 all illustrate
  `ErrorMessage` with a stale children-based API (`<ErrorMessage>{message}</ErrorMessage>`) instead
  of the real prop-based one (`title`/`message`/`requestId`/`onRetry`) — predates this session,
  spans multiple patterns/snippets (not a trivial one-line fix), and is a spec-kit-owned contract
  doc, so left as a flagged known gap rather than hand-edited, per the same reasoning as the existing
  "Issue-Body Snippets Are Lowest-Authority" pattern. Also noted: that same contract file isn't
  cross-linked from `frontend/README.md`'s Related Specifications table (pre-existing gap, unrelated
  to this session).
- **Workflow note**: `docs/roadmap.md`'s 005-T114–T116 rows are deliberately left at `Backlog` in
  this commit, same established convention as #117/#110/#109 — the Backlog→Done flip happens once
  the PR exists/merges. Whoever runs `open-pr`/closes this ticket next should flip those three rows.

### Session: G-ARCH-INTEGRATION-TESTS (#119) + 005-T119 Docs (#120) — Healthz/DB-Timeout Integration Tests, Error-Handling Pattern Docs
- **Date**: 2026-07-13
- **Tool**: Claude Code
- **Outcome**: Found the previous session's workflow note above had gone unactioned — #117/#118's
  PRs (#128, #129) were already merged, but `docs/roadmap.md`'s 005-T110–T116 rows were still
  `Backlog`. Fixed that debt first, standalone (PR #130, branch
  `docs/sprint4-roadmap-status-117-118`), before starting new work. Then implemented both issues
  requested in the same prompt on separate branches/PRs, per this repo's one-issue-per-PR norm:
  - **#119** (005-T117/T118, tdd-developer, branch
    `feature/119-g-arch-integration-tests-healthz-correlation-id-db-timeout-error-response`, PR
    #131): `backend/tests/integration/health_test.go` (2 tests — 200 always, generated/echoed
    `X-Request-ID`) and `error_test.go` (1 test). The DB-timeout scenario needed a *realistic*
    simulated failure per the issue's Review Focus, not a short-circuit — see the new
    `patterns-discovered.md` entry "Simulating a Realistic DB Timeout in an Integration Test" for
    the technique (statement_timeout connection param + a second connection holding an `ACCESS
    EXCLUSIVE` table lock). No RED phase existed (same honest caveat as the swagger_test.go
    precedent) since #117's backend work was already merged — tests went GREEN immediately against
    real Colima-backed Postgres testcontainers.
  - **#120** (005-T119, technical-writer, branch
    `feature/120-005-t119-document-error-handling-patterns`, PR #132): extended
    `backend/README.md`'s error-handling section with 5 concrete, verified-against-source examples
    (domain-error construction, wrap-vs-translate, handler `HandleError` usage, retry/backoff
    cross-reference, correlation-ID propagation), and added a new `frontend/README.md` "Error
    Handling" section (`APIError` → `useErrorHandler` → `ErrorMessage` end-to-end). While in that
    same backend section, found and fixed a genuine stale spot: the `/healthz` table still
    described the pre-#126 503/`unhealthy` contract instead of the real always-200,
    status-in-body one — PR #126 apparently never updated this doc when it changed the behavior.
- **Independent verification** (main session, not just subagents' self-reports): re-ran
  `gofmt`/`go vet`/`golangci-lint`/`go build` myself, then the full non-short suite
  (`go test -tags=test ./...` with Colima's `DOCKER_HOST`/`TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE`
  env vars) — all 8 packages `ok`, the new DB-timeout test bounded at ~6s (proving the timeout
  mechanism genuinely engaged, not an instant pass). Re-read every README diff hunk against the
  actual source files and confirmed the new `backend/README.md#error-handling-internalerrors`
  anchor link matches GitHub's heading-slug convention already used elsewhere in the same file
  (`#ai-client-foundation-internalai` precedent).
- **Workflow note**: unlike #117/#118, this session flipped the *already-merged* 005-T110–T116 rows
  to `Done` right away (PR #130) since that debt existed and the user asked to keep the roadmap
  current — but 005-T117–T119's own rows are deliberately left `Backlog` in PRs #131/#132, same
  established convention: flip after those two PRs merge, not now.
