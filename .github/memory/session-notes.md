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
