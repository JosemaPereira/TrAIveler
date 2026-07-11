# Session Notes

Historical summaries of completed development sessions. Committed to git as a record.

**Note**: Older sessions (specs 001-007, sprint infrastructure setup) are compacted to save tokens. Policy: each sprint's implementation sessions are compacted into a `(Compacted)` summary immediately at that sprint's own closure — the same pass that updates the sprint-closure documentation. At any given time, only the sprint currently in progress (not yet closed) has a `(Detailed)` section; the moment it closes, it gets compacted, no rolling window. Sprint 1 and Sprint 2 (both closed) were compacted together on 2026-07-11 to bring the file in line with this policy after it changed; Sprint 3's detailed section will be compacted at Sprint 3's own closure, and so on each sprint thereafter.

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
  - **Sprint 1 closure audit**: All 23 tasks (Spec 005 Phases 1-2, Setup + Foundational) confirmed complete across 8 PRs (#43-#50; #48/#49/#50 merged). 10 stale roadmap rows corrected Backlog/In Progress → Done with PR references; `## Sprint Plan`'s Sprint 1 block marked "✅ COMPLETE". 100% completion rate in ~2 weeks — established the ~23-tasks/sprint velocity baseline used to scope Sprint 2 (62 tasks, 3 parallel tracks).
  - **Sprint 2 consolidation planning**: 37 tasks (005-T024–T060) grouped into 14 work items (-62% issue reduction) via new `Group` values (e.g. `G-SPRINT2-BACKEND-HTTP-SERVER` for T035-T037, expanded `G-SPRINT2-BACKEND-EXAMPLE` for T038-T041, `G-SPRINT2-FRONTEND-INFRASTRUCTURE` for T056-T058). `.github/SPRINT-CONSOLIDATION-CHECKLIST.md` created as the enforced pre-issue-creation checkpoint for all future sprints.
  - **Git templates standardized**: `.github/PULL_REQUEST_TEMPLATE.md` (Description/Implementation Summary/Testing/Checklist/Verification/Dependencies, Stable-ID/Spec/Sprint/Group fields, grouped-vs-standalone support) and `.github/COMMIT_GUIDELINES.md` (Conventional Commits, approved types, 7 worked examples, anti-patterns) created to fix Sprint 1's inconsistent commit/PR structure; `copilot-instructions.md` updated to reference both.
  - **Documentation audit (005-T029)**: Added 28 lines of documentation to backend `config`/`database` packages (loader functions, `validate()` rationale, `setupPostgresContainer` helper); confirmed all 4 project areas (backend/frontend/e2e/infra) already met documentation standards with zero behavior changes (config: 26 tests, database: 11 tests, all passing).
  - **Claude Code established as primary tool**: `CLAUDE.md` created as the canonical instructions file (`copilot-instructions.md` demoted to a mirror); `.github/memory/README.md` unified as the single cross-tool memory protocol with `**Tool**:` provenance tagging on every new entry. 10 Copilot agents/prompts hand-ported to `.claude/agents/` subagents (5 role agents merging their linked one-shot prompt as a `## Task:` section; 5 standalone workflows: `build-roadmap`, `commit-and-push`, `open-pr`, `promote-fundations`, `sync-issues`). Official spec-kit Claude integration installed (`specify integration install claude --script sh --force`), generating 10 `.claude/skills/speckit-*/SKILL.md` skills. A mistaken hand-ported `.claude/agents/speckit-*.md` + `.claude/commands/speckit-*.md` duplication (20 files) was found and deleted once the official integration was confirmed to cover the same ground; the entire `.claude/commands/` wrapper layer was also removed, per explicit user request to minimize file count.
- **Key Decisions**:
  - One-task-per-issue was Sprint 1's approach but caused backlog fragmentation (PRs naturally combined multiple tasks anyway); consolidation-first became **mandatory starting Sprint 2**. Group naming fixed as `G-SPRINT<N>-<STACK>-<AREA>`, optimal group size 2-4 tasks (~150-200 LOC, one reviewable PR); foundation components/critical blockers/single-file features stay standalone.
  - PR/commit templates are GitHub-native (auto-populating) rather than manually enforced, guaranteeing Stable-ID/Spec/Sprint/Group traceability on every future PR.
  - **Always check for an official multi-agent-tool integration before hand-porting a toolkit's own agent config** (see `patterns-discovered.md` "Prefer the official multi-agent integration..." entry) — this is why the SpecKit hand-port was reverted; standing practice for adding any future AI tool to this repo. `specify integration install <key> --force` adds a second integration without touching the existing one.
  - Bidirectional drift detection (`scripts/agent-port-manifest.json` + `scripts/check-agent-drift.py`, see matching `patterns-discovered.md` entry) adopted to reconcile Claude Code's and Copilot's hand-ported agent/prompt pairs, since the two formats differ too much for safe full codegen.
  - Background `Agent`-tool tasks that get unexpectedly killed should be resumed once via `SendMessage`; if killed again, fall back to foreground work rather than retrying indefinitely.

---

## Sprint 2 Implementation (Compacted)

### Sessions: Backend Foundation (Middleware, HTTP Server, Errors, AI Client, Reference Pattern), CI/Docker Fixes & Repo Ruleset, Frontend Foundation (Tokens, Primitives, App Shell, Auth Store)
- **Date Range**: 2026-07-09 to 2026-07-10 (Sprint 2 closed 2026-07-11 — all 37 tasks / 14 work items done across 10+ PRs; see `docs/roadmap.md`'s Sprint 2 entry in `## Sprint Plan`)
- **Key Outcomes**:
  - **G-SPRINT2-BACKEND-MIDDLEWARE** (005-T024–T028, issue #54, PR #71): `backend/internal/middleware/` — `RequestID`/`Logger`/`Recovery`/`CORS`/`BodySize`, all constructor-injected (no globals), 22 tests/98.6% coverage. Established the "mockery is for external-system interfaces only, plain constructor injection is enough for stdlib-typed deps" rule (`patterns-discovered.md`).
  - **G-SPRINT2-BACKEND-HTTP-SERVER** (005-T035–T037, issue #57, PR #72): real Chi router (`cmd/api/server.go`/`routes.go`) replacing the `main.go` TODO stub, mandated middleware order, `/healthz`, graceful shutdown; first real use of `go-chi/chi/v5`. Also fixed a persistent local dev annoyance (gvm's per-directory `.go-version` auto-switch dropping shell functions in Claude Code's snapshot mechanism — machine-level `~/.gvm` fix, not project code) and added gopls `-tags=test` build flags.
  - **G-SPRINT2-BACKEND-ERRORS** (005-T033–T034, issue #56, PR #74): `backend/internal/errors/` — `DomainError` + 5 constructors + `HandleError`, 21 tests/96.3% coverage. Alongside it, a real CI perf investigation (not part of any ticket) found `backend/Dockerfile`'s builder stage running fully under QEMU emulation (451s of an 8.5min build); fixed via `FROM --platform=$BUILDPLATFORM` + `ARG TARGETOS/TARGETARCH` (PR #73, dropped `go build` to ~56s) — formalized in `patterns-discovered.md` ("Native Cross-Compilation..."). Also bumped `alpine:3.19`→`3.23` (real CVE drift) but left `postgres:15.4-alpine` alone (a documented architectural pin, not drift — still open, tracked in working-notes.md).
  - **Repo governance** (PR #75): configured a "Protect main" ruleset (PR + 1 approval + 10 named required checks + strict up-to-date + no force-push/delete, Admin-role bypass) — required discovering GitHub Rulesets need a public repo or Pro plan for personal accounts. Found and fixed a real correctness gap along the way: `paths:`-filtered `pull_request` triggers on required-check workflows leave GitHub's check stuck at "Expected" forever on non-matching PRs; fixed via an always-running `dorny/paths-filter` gate job per workflow (`backend-ci.yml`/`frontend-ci.yml`/`infra-plan.yml`) — formalized in `patterns-discovered.md` ("Required Status Checks Must Always Run"). Live-verified via PR #75, which also surfaced two real pre-existing frontend gaps (missing `test` script, unformatted files).
  - **G-SPRINT2-BACKEND-AI-CLIENT** (005-T030–T032, issue #55, PR #77): `backend/internal/ai/` `AIClient` interface + stubs, plus a user-approved scope extension — a real local-dev `OllamaClient` (`AI_PROVIDER=ollama` by default in dev) with `docker-compose.yml`'s `ollama` service and `docs/local-ai-setup.md`. Local-dev-only doc additions appended *after* the `PROMOTED:...END` markers in `docs/architecture.md`/`docs/cloud-and-environments.md` rather than editing promoted content (reusable pattern, not yet in `patterns-discovered.md` — flagged in the compaction report). Also fixed an unrelated real bug: `CookieSecure` wasn't deriving from `GO_ENV` like `AIConfig` did. Surfaced the still-open Swagger/OpenAPI planning gap for Sprint 3 (saved to cross-session memory).
  - **G-SPRINT2-BACKEND-EXAMPLE** (005-T038–T041, issue #58, PR #78): `backend/internal/example/` — the canonical model→repository→service→handler reference pattern (58 tests, 88% coverage), `If-Match`-header optimistic locking, a real two-query pagination fix for a `COUNT(*) OVER()` empty-page bug (formalized in `patterns-discovered.md`). **This package is throwaway and must be deleted once the first real domain package (Trip, etc.) ships** — full deletion checklist in its own `patterns-discovered.md` entry; still pending as of Sprint 2 close.
  - **G-SPRINT2-FRONTEND-TOKENS** (005-T042–T043, issue #59, PR #79): `tokens.css`/`global.css` sourced from `docs/ui-guidelines.md` (not stale issue pseudocode). Same session fixed a live CI gap (`frontend/package.json` had no `test` script at all) by installing the real Vitest/RTL/MSW toolchain, and removed the deprecated `tsconfig.json` `baseUrl`. Follow-up commit corrected a stale "Sprint 2" Lighthouse-CI claim in `frontend-ci.yml`/README and flagged the Lighthouse/a11y-CI scheduling gap for Sprint 3 planning (cross-session memory).
  - **G-SPRINT2-FRONTEND-PRIMITIVES+FORM** (005-T048–T051/T055, issues #63/#65, PR #80): `Button`/`Input`/`Card` primitives + `Form` composite, first entities driven purely by the tokens above. Formalized the recurring "issue-body pseudocode is lowest authority" lesson into a general rule in `patterns-discovered.md` (real code > spec `tasks.md` > root `docs/*.md` > issue body), after it had independently resurfaced 3 times by this point.
  - **G-SPRINT2-FRONTEND-APP-SHELL** (005-T044/T045/T047/T056-T060, issues #60/#62/#66, PR #81): `api-client.ts`/`query-client.ts`, `ErrorBoundary` + React Router v7 config, `App.tsx` wiring `ErrorBoundary > QueryClientProvider > RouterProvider`. Removed a hardcoded `VITE_API_BASE_URL` fallback that was silently masking a CI test-job wiring gap (fixed with a committed `frontend/.env.test`) — formalized in `patterns-discovered.md`. Live headless-Chromium smoke test confirmed zero console errors through the full provider stack.
  - **G-SPRINT2-FRONTEND-AUTH-STORE+PRIMITIVES** (005-T046/T052-T054, issues #61/#64, PR #82): Zustand `auth-store.ts` (role enum corrected to the real `'admin'|'partner'` from `docs/data-model.md`, not the issue's stale sketch) plus `LoadingSpinner`/`ErrorMessage`/`EmptyState`. Bundled a `tdd-developer`/`implement-feature` self-optimization (targeted grep-based memory loading instead of full-file reads, now in `patterns-discovered.md`) into the same PR at the user's request. This closed out all remaining `005-T042–T060` frontend-architecture items.
- **Key Decisions**:
  - **"Issue-body pseudocode/diagrams go stale — verify against real code/docs" recurred independently across nearly every session this sprint** (token names, backend package signatures, component file placement, test-file placement, domain-entity shape) and was eventually promoted to a single general rule in `patterns-discovered.md` with an explicit authority ranking — treat any new instance as confirming that rule, not as a fresh one-off.
  - **Independently re-verify every subagent's output** (rebuild, re-run tests/lint, read the actual diff) rather than trusting self-reports — this standing practice caught real issues multiple times this sprint (stale doc claims, wrong test-file location, a hardcoded fallback masking a CI gap) and was extended explicitly to documentation-review subagent passes, not just `tdd-developer`.
  - **Operationally consequential GitHub/Docker platform behavior must be verified against current docs or by actually running it, not recalled from training data** — both the required-status-check semantics and the QEMU/cross-compile fix were confirmed empirically (WebSearch against current GitHub docs; a real local `docker buildx` run) after an initial wrong assumption in the ruleset session.
  - Scope-extension/scope-limiting decisions on a ticket (e.g. Ollama-as-real-client vs. doc-only stub) should go through explicit `AskUserQuestion` confirmation rather than being inferred, given multi-file/hard-to-reverse blast radius.
  - **Known open items carried past Sprint 2 close** (see `scratch/working-notes.md` and `patterns-discovered.md` for full detail): `postgres:15.4-alpine` version-staleness decision still pending; `internal/example/` deletion still pending (blocked on the first real domain package); a pre-existing escape-sequence text corruption (literal `\n`/`—`) in `patterns-discovered.md`'s "Suggest-Then-Approve Collaboration Workflow" entry was found and fixed in a concurrent edit during this same session (`patterns-discovered.md` was also simplified/deduplicated in that pass — see that file's own top note).
