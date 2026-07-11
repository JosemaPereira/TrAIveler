# Session Notes

Historical summaries of completed development sessions. Committed to git as a record.

**Note**: Older sessions (specs 001-007) are compacted to save tokens. Recent sessions (Sprint 1+) maintain full detail.

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

## Sprint 1 Implementation (Detailed)

### Session: Sprint 1 Closure & Roadmap Reconciliation
- **Date**: 2026-07-08
- **What was accomplished**:
  - **Sprint 1 completion audit**: Analyzed all 23 tasks from Spec 005 Phases 1-2 (Setup + Foundational)
  - **Discovered missing PRs**: Found that 10 tasks marked as "Backlog" or "In Progress" in roadmap were actually complete
    - E2E setup complete (005-T003, 005-T007): Playwright infrastructure, directories, sample tests
    - Infrastructure foundation complete (005-T004, 005-T012, 005-T019, 005-T022, 005-T023): Terraform configs, CI workflow
    - All .gitignore files complete (005-T015): backend, frontend, e2e, infra
    - CI workflow skeletons complete (005-T017, 005-T018): backend-ci.yml, frontend-ci.yml
  - **Mapped tasks to PRs**: Identified 8 PRs covering all 23 Sprint 1 tasks
    - PR #43: Backend directory structure (005-T001)
    - PR #44: Go module initialization (005-T005)
    - PR #45: Config, linting, .env, README, partial .gitignore (005-T008, 005-T010, 005-T013, 005-T020, partial 005-T015)
    - PR #46: Docker infrastructure (005-T014, 005-T016)
    - PR #47: Complete frontend setup (005-T002, 005-T006, 005-T009, 005-T011, 005-T021, partial 005-T015)
    - PR #48: E2E setup (005-T003, 005-T007, partial 005-T015) — MERGED
    - PR #49: Infrastructure foundation (005-T004, 005-T012, 005-T018, 005-T019, 005-T022, 005-T023, partial 005-T015) — MERGED
    - PR #50: CI workflow skeletons (005-T017, 005-T018) — MERGED
  - **Updated roadmap**: Changed 10 tasks from Backlog/In Progress → Done with PR references
  - **Updated Sprint Plan section**: Marked Sprint 1 as "✅ COMPLETE" with all PR details and completion date
  - **Verified PR status**: Confirmed PRs #48, #49, #50 merged; all issues addressed correctly
- **Key findings and decisions**:
  - **All 23 Sprint 1 tasks complete**: 100% completion rate, delivered on schedule (Weeks 1-2)
  - **8 PRs total**: Average 2.9 tasks per PR, good balance of consolidation without bloat
  - **Foundation tasks enable parallelization**: Backend, frontend, e2e, and infra directories all ready for Sprint 2 concurrent work
  - **Documentation-first approach paid off**: All READMEs include implementation status, onboarding guidance, and comprehensive technical documentation
  - **Multi-stage Docker and Terraform patterns established**: Production-ready patterns set from Sprint 1, not retrofitted later
  - **CI workflow skeleton strategy validated**: Functional lint/test/build jobs with deployment placeholders for future sprints (3, 10) prevents rework
- **Outcomes**:
  - ✅ Sprint 1 officially CLOSED (2026-07-08)
  - ✅ All 23 tasks marked Done in roadmap with PR references
  - ✅ Comprehensive Sprint 1 closure report generated
  - ✅ Foundation complete for Sprint 2 parallel work
  - **Next sprint velocity baseline**: 23 tasks in ~2 weeks = solid baseline for Sprint 2 planning (62 tasks with 3 parallel tracks)
  - **Sprint 2 ready to begin**: Backend architecture, frontend architecture, infrastructure modules Part 1 can start immediately
  - **Roadmap reconciled**: Single source of truth updated, no drift between codebase and planning documents

---

### Session: Sprint 2 Planning with Mandatory Task Consolidation
- **Date**: 2026-07-08
- **What was accomplished**:
  - **Sprint 2 task consolidation**: Applied consolidation-first workflow to prevent Sprint 1 fragmentation
    - Analyzed 37 Sprint 2 tasks (005-T024 to 005-T060)
    - Applied consolidation rules systematically
    - Created 14 work items instead of 37 individual issues (-62% reduction)
  - **New consolidation groups created**:
    - Backend: G-SPRINT2-BACKEND-HTTP-SERVER (T035-T037) — HTTP server + DB integration + healthcheck in cmd/api/main.go
    - Backend: Expanded G-SPRINT2-BACKEND-EXAMPLE (T038-T041) — complete pattern from model to handler
    - Frontend: G-SPRINT2-FRONTEND-INFRASTRUCTURE (T056-T058) — setup tasks (.gitkeep, ErrorBoundary, routes)
  - **Documentation artifacts created**:
    - `.github/SPRINT-CONSOLIDATION-CHECKLIST.md` — MANDATORY checklist for all future sprints
    - Updated Sprint 2 section in roadmap with consolidation details
    - Documented process for Sprint 2 closure and Sprint 3 preparation
  - **Roadmap updates**:
    - Added Group values to tasks T035-T037, T041, T056-T058
    - Updated Sprint 2 plan section with work items breakdown
    - Added consolidation metrics and issue creation instructions
- **Key findings and decisions**:
  - **Sprint 1 retrospective identified core issue**: Creating 1 issue per task led to fragmentation. During implementation, multiple tasks were naturally combined in single PRs, proving tasks shared enough context to be grouped from the start.
  - **Consolidation rules validated**: Backend and frontend follow same pattern — group by package/directory, same tech stack, shared context, optimal size 2-4 tasks per group.
  - **Group naming convention**: `G-SPRINT<N>-<STACK>-<AREA>` format makes groups immediately recognizable (e.g., G-SPRINT2-BACKEND-MIDDLEWARE).
  - **Optimal group size confirmed**: 2-4 tasks = 1 reviewable PR (~150-200 LOC). Smaller groups (1 task) stay standalone, larger groups (5+ tasks) need splitting unless extremely coherent (like middleware package).
  - **Standalone task criteria**: Foundation components (database client, auth store), critical blockers, or single-file features (Form composite) that don't naturally group with others deserve individual issues for visibility.
  - **MANDATORY consolidation checkpoint**: Checklist created to enforce consolidation BEFORE creating issues for any future sprint. This prevents backtracking and rework.
  - **Process improvement**: Sprint closure now includes "lessons learned" section in checklist to capture what worked/didn't work, feeding forward to next sprint.
- **Outcomes**:
  - ✅ Sprint 2 ready for issue creation with optimal consolidation
  - ✅ 37 tasks → 14 work items (6 backend + 8 frontend) = -62% issue reduction
  - ✅ Clear workflow documented: consolidate → update roadmap → create issues
  - ✅ Checklist ensures future sprints follow same pattern (Sprint 3, 4, etc.)
  - ✅ Roadmap Group column populated for all Sprint 2 tasks
  - **Next step**: Run `/create-sprint-issues 2` to create 14 consolidated GitHub issues
  - **Sprint closure protocol**: At end of Sprint 2, update checklist with lessons learned to inform Sprint 3 consolidation
  - **Permanent process improvement**: Every sprint now starts with consolidation analysis, not issue creation

---

### Session: Standardization of Commit and PR Templates
- **Date**: 2026-07-08
- **What was accomplished**:
  - **Created standardized PR template** (`.github/PULL_REQUEST_TEMPLATE.md`)
    - Auto-loads in all GitHub PRs
    - Includes sections: Description, Implementation Summary, Testing, Checklist, Verification, Dependencies, Screenshots, Deployment Notes
    - Enforces consistent structure: Stable IDs, Spec reference, Sprint, Group tracking
    - Supports both grouped tasks (with checklist) and standalone tasks
  - **Created comprehensive commit guidelines** (`.github/COMMIT_GUIDELINES.md`)
    - Documents Conventional Commits format (type(scope): subject)
    - Lists all approved types: feat, fix, docs, style, refactor, perf, test, chore, ci
    - Provides scope examples for backend, frontend, infrastructure, e2e, docs
    - Includes 7 real-world examples (simple feature, bug fix, grouped tasks, breaking changes)
    - Defines anti-patterns and verification checklist
  - **Updated copilot-instructions.md**:
    - Git Workflow section now references both templates explicitly
    - Reinforces mandatory English language policy for all Git artifacts
  - **Documentation cleanup**:
    - Confirmed removal of obsolete files (ISSUE-CREATION-GUIDELINES.md, PM-WORKFLOW-CONSOLIDATION.md, scripts/)
    - User already cleaned up redundant documentation
- **Key findings and decisions**:
  - **Problem identified**: Sprint 1 commits and PRs lacked consistent structure, making review and tracking harder
  - **Root cause**: No enforced templates → each commit/PR used different format/style/level of detail
  - **Solution approach**: GitHub-native templates that auto-populate instead of manual enforcement
  - **PR template design**: Balances thoroughness with practicality — includes fields for grouped vs standalone tasks
  - **Commit guidelines philosophy**: Teaching document, not just rules — explains "why" behind each convention
  - **Language policy reinforcement**: All Git artifacts (commits, PRs, branches) MUST be English, even when conversations are in Spanish
  - **Integration strategy**: Templates referenced in copilot-instructions.md ensures AI assistant follows them automatically
- **Outcomes**:
  - ✅ All future PRs will load standardized template automatically
  - ✅ Commit guidelines provide clear reference for developers and AI assistant
  - ✅ copilot-instructions.md ensures AI follows templates in all generations
  - ✅ Reduced variability in commit/PR structure across all contributors
  - ✅ Better traceability: PR template enforces Stable ID, Spec, Sprint, Group tracking
  - ✅ Documentation cleanup complete: Only relevant files remain in .github/
  - **Immediate benefit**: Starting with Sprint 2 issue creation, all PRs will follow consistent format
  - **Long-term benefit**: Easier review, clearer history, better automated tooling integration (changelog generation, release notes)

---

### Session: Technical Documentation Enhancement (Task 005-T029 Documentation)
- **Date**: 2026-07-09
- **What was accomplished**:
  - **Comprehensive documentation review**: Analyzed all project areas (backend, frontend, e2e, infra) for documentation quality
  - **Backend enhancements**: Added 21 lines of documentation to config package
    - Documented all configuration loader functions (loadServerConfig, loadDatabaseConfig, loadAIConfig, loadAuthConfig, loadLogConfig)
    - Enhanced validate() function documentation with detailed validation rules and rationale
    - Added 7 lines of documentation to database test helpers (setupPostgresContainer)
  - **Documentation audit**: Verified all existing documentation meets project standards
    - Package-level docs: ✅ Complete for all Go packages (main, config, database)
    - Function docs: ✅ Complete with parameters, return values, and usage examples
    - Interface docs: ✅ Comprehensive documentation with examples (database.Client)
    - JSDoc/TSDoc: ✅ All React components properly documented
    - Test documentation: ✅ Descriptive names and helper function docs
    - READMEs: ✅ All 4 project areas have comprehensive documentation
- **Key findings and decisions**:
  - **Excellent baseline quality**: Project already maintains high documentation standards across all areas
  - **Documentation-only changes**: All enhancements verified with zero behavior changes (all tests passing)
  - **Language consistency**: All documentation follows mandatory English-only policy
  - **Go documentation conventions**: Package comments, function docs, and inline explanations follow Go best practices
  - **Infrastructure pending**: Terraform modules await Sprint 3 implementation as documented in infra/README.md
- **Outcomes**:
  - ✅ All configuration loading logic clearly documented with purpose and behavior
  - ✅ Validation rules explicit with justification (JWT_SIGNING_KEY in production, connection pool limits, port ranges, log levels)
  - ✅ Test helpers properly documented for maintainability (setupPostgresContainer with container config, credentials, wait strategy)
  - ✅ Verified no behavior changes: config tests (26 cases), database tests (11 cases) all passing
  - ✅ Documentation coverage report generated: Backend (Excellent), Frontend (Excellent), E2E (Excellent), Infra (Pending Sprint 3)
  - **Technical debt**: None identified - documentation quality is production-ready
  - **Follow-up work**: Terraform module documentation when implemented in Sprint 3

### Session: Claude Code Migration — Primary Tool Setup, Memory Unification, Agent Porting & SpecKit Integration
- **Date**: 2026-07-09
- **Tool**: Claude Code
- **What was accomplished**:
  - **CLAUDE.md created**: canonical, auto-loaded instructions file for Claude Code at repo root, mirroring `.github/copilot-instructions.md`'s content (project context, documentation references, language policy, dev principles, git workflow, task consolidation policy).
  - **Tool hierarchy established**: Claude Code declared primary, GitHub Copilot auxiliary. `.github/copilot-instructions.md` now carries a banner declaring itself a mirror of `CLAUDE.md`; `.github/memory/README.md`'s "Persistent Memory" section updated to match.
  - **Memory system unified**: `.github/memory/README.md` is now the single canonical protocol doc for both tools (CLAUDE.md/copilot-instructions.md summarize and point to it instead of restating). Added a "Multi-Tool" policy: append-only, `**Tool**: <name>` tagging on every new entry, sync-before-write, keep-both-sides-on-merge-conflict. Templates in `session-notes.md` and `patterns-discovered.md` updated with a `Tool` field.
  - **Copilot agents/prompts ported to Claude Code subagents** (`.claude/agents/`): 5 role agents (`code-reviewer`, `product-manager`, `tdd-developer`, `technical-writer`, `test-engineer`) each merged with their linked one-shot Copilot prompt(s) as `## Task: ...` sections in the same file (e.g. `product-manager.md` contains both Plan Sprints and Create Sprint Issues); 5 standalone workflow agents (`build-roadmap`, `commit-and-push`, `open-pr`, `promote-fundations`, `sync-issues`).
  - **Official spec-kit Claude Code integration installed**: `uv`/`specify-cli` installed locally, then `specify integration install claude --script sh --force` run — added `claude` as a second, coexisting spec-kit integration (Copilot's stayed default and untouched). Generated `.claude/skills/speckit-*/SKILL.md` (10 skills: analyze, checklist, clarify, constitution, converge, implement, plan, specify, tasks, taskstoissues), directly invocable as `/speckit-*`.
  - **Cleanup of a naming collision**: initially hand-ported all 10 SpecKit agents as `.claude/agents/speckit-*.md` + `.claude/commands/speckit-*.md` wrappers before discovering the official integration existed; deleted all 20 once the official skills were confirmed to cover the same names, keeping the official skills as sole source.
  - **`.claude/commands/` wrapper layer removed entirely**: initially added thin `context: fork` command wrappers for all 10 non-SpecKit subagents for `/name` discoverability (Copilot UX parity), but removed them per explicit user request to minimize file count — subagents are now used purely via natural/explicit Agent-tool delegation.
  - **Bidirectional drift detection built**: `scripts/agent-port-manifest.json` (sha256 hash per paired file) + `scripts/check-agent-drift.py` (compares current hashes to the manifest, reports `claude changed` / `copilot changed` / `both changed`, exits non-zero on drift, `--update <name>|all` refreshes hashes after manual reconciliation). Tested both directions (simulated a Copilot-side edit, confirmed detection, reverted, confirmed clean).
  - Added `.gitignore` (previously empty) with `.specify/integrations/.cache/` (spec-kit's local catalog cache, not meant to be committed).
- **Key findings and decisions**:
  - **spec-kit has native, actively-maintained Claude Code support** (`specify integration list` → key `claude`, Skills-based, multi-install safe) — always check for and prefer an official integration over hand-porting a tool's own agent files; hand-porting SpecKit caused the naming collision that had to be cleaned up.
  - `specify integration install <key> --force` is the correct way to add a second AI-agent integration to a project already initialized with another (the `--force` is only required because the *existing* integration, e.g. `copilot`, isn't declared multi-install-safe — it does not remove or modify the existing integration). `specify integration upgrade claude --force` is the way to pull upstream spec-kit updates for the Claude side later (diff-aware).
  - Manual (non-mechanical) sync between Copilot's and Claude Code's agent/prompt formats is unavoidable for hand-ported agents (different frontmatter fields, prompt+agent merging, slash-command vs. subagent-delegation phrasing) — chose hash-manifest-based drift *detection* over full bidirectional codegen, since a generator would have to encode the same nuanced adaptation rules and still couldn't safely auto-apply them.
  - User strongly prioritizes minimizing file count / token overhead in this repo's AI-agent tooling — traded away the `/name` slash-command discoverability layer once its redundancy with natural delegation was pointed out.
  - Background `Agent` tool tasks were killed unexpectedly twice this session (environment issue, not user-initiated, confirmed with the user). Lesson: on an unexpected "killed" task-notification, resume once via `SendMessage`; if the same task gets killed again, stop retrying and do the work directly in the foreground instead.
- **Outcomes**:
  - New Claude Code sessions in this repo auto-load `CLAUDE.md` and can delegate to 10 project subagents (`.claude/agents/`) and 10 official SpecKit skills (`.claude/skills/`, invoked as `/speckit-*`) by name or natural language.
  - `python3 scripts/check-agent-drift.py` (exit 0 = all 10 hand-ported pairs in sync) is now the required check before/after editing any hand-ported agent on either the Copilot or Claude side.
  - Copilot's own setup is fully intact and still the default spec-kit integration; both tools read/write the same `.github/memory/` system with provenance tagging going forward.
  - **Follow-up work**: none blocking; optionally revisit whether `.claude/commands/` wrappers are wanted later if `/name` discoverability turns out to matter in practice.

### Session: Chi Middleware Package (005-T024–T028, Issue #54)

- **Date**: 2026-07-09
- **Tool**: Claude Code
- **What was accomplished**:
  - Implemented `backend/internal/middleware/` via strict TDD (delegated to the `tdd-developer` subagent): `RequestID`, `Logger`, `Recovery`, `CORS`, `BodySize`, plus a shared `errors.go` error-envelope helper matching `docs/api-design-standards.md` §7. All five follow the stdlib `func(http.Handler) http.Handler` shape and are constructor-injected (logger, allowed-origins string, next handler) — no globals, no hidden `config.Load()`/`slog.Default()` calls inside the package.
  - 22 unit tests, 98.6% coverage, `gofmt`/`go vet` clean; `go.mod` only promoted already-present indirect deps (`google/uuid`, `felixge/httpsnoop`) to direct — no new dependencies, no `go-chi/chi` added (out of scope; router wiring is the separate, still-blocked issue #57).
  - Updated `backend/README.md` (delegated to `technical-writer`): new "Middleware" section, fixed stale `internal/middleware/` file list (wrong filename `requestid.go`, missing 4 new files), fixed the same typo in the NFR Infrastructure table, and — found while in there — repaired pre-existing unrelated corruption (a broken/duplicate "Quick Start"+"Configuration Validation" region with a code fence that never closed, silently turning a chunk of the file into a literal block on GitHub).
  - Reviewed in-code doc comments against this repo's `revive` config (`exported`, `package-comments`, `context-keys-type` rules) at the user's request; added two WHY-only comments that were genuinely missing: why `ctxKeyRequestID` is an unexported `struct{}` (collision avoidance, per `context-keys-type`), and a caveat on `Recovery`/`BodySize` about best-effort behavior when the downstream handler already started writing a response before erroring.
  - Committed (`9619d11`, delegated to `commit-and-push`) and pushed to `origin/feature/54-g-sprint2-backend-middleware-chi-middleware-package` — a remote branch that already existed (created from `main`, no prior commits), checked out locally with `git checkout -b <name> origin/<name>` before delegating so the commit didn't land on `main`.
- **Key findings and decisions**:
  - **No mockery needed for this package** — mockery is scoped in `.mockery.yaml` to actual domain interfaces (`internal/database.Client`); this middleware package defines zero interfaces of its own, and its one external-ish dependency (`next http.Handler`) is the stdlib's single-method interface, conventionally faked with a plain `http.HandlerFunc` literal in tests rather than a generated mock. Confirmed by grepping the package for `interface` (zero hits) before answering the user's explicit question about it — see also [[mockery-scope-vs-di]] pattern.
  - The issue's prose mismatched the actual existing config: it said env var `ALLOWED_ORIGINS`, but Sprint 1 (005-T013) already shipped `ALLOWED_CORS_ORIGINS` → `config.Config.Server.AllowedCORS`. Used the existing name rather than introducing a duplicate env var; `CORS` takes the value as a plain constructor string argument instead of importing `config` directly, keeping the middleware package decoupled and unit-testable without env vars.
  - `Logger`/`Recovery` deviated from the issue's literal bare-`func(http.Handler) http.Handler` signature to `func(logger *slog.Logger) func(http.Handler) http.Handler` — necessary so tests can capture log output via a local `bytes.Buffer`-backed logger instead of mutating global state (`slog.SetDefault`) across parallel test runs.
- **Outcomes**:
  - `backend/internal/middleware/` is complete, tested, and documented; ready for issue #57 to import and wire into a `chi.Router` in `cmd/api/main.go` (currently just a `// TODO: Initialize HTTP router and middleware` comment there).
  - **Follow-up work**: issue #57 (HTTP server + router integration, currently blocked, now unblocked by this work).

### Session: HTTP Server with Health Check (005-T035–T037, Issue #57, PR #72)

- **Date**: 2026-07-09
- **Tool**: Claude Code
- **What was accomplished**:
  - Implemented `backend/cmd/api/server.go` (`HTTPServer`, Chi router with the mandated middleware order `RequestID → Logger → Recovery → CORS → BodySize`, `/healthz` handler), `routes.go` (route registration), and rewired `main.go` (real router replacing the `// TODO` stub, `*http.Server` using `cfg.Server.{Read,Write,Idle}Timeout`, graceful shutdown on SIGINT/SIGTERM with a 30s timeout, DB client closed after HTTP shutdown) — delegated to `tdd-developer`, 9 new tests in `server_test.go` (healthy/unhealthy/timeout `/healthz`, middleware ordering, panic recovery, CORS, start/shutdown lifecycle, in-flight-request-survives-shutdown), all passing. Added `github.com/go-chi/chi/v5` to `go.mod` (first real use of Chi in this repo — `docs/architecture.md`'s mandated router).
  - `backend/README.md` trimmed 801→346 lines and `server.go`/`routes.go`/`main.go`/`server_test.go` comments tightened to WHY-only — delegated to `technical-writer`, verified with `gofmt`/`go vet`/full test run after.
  - PR #72 opened (`feat(api): add HTTP server with health check and graceful shutdown`, closes #57) via `open-pr`; `docs/roadmap.md` rows 005-T035/T036/T037 marked Done with "Closed by PR #72".
  - Independently verified every subagent's output myself before reporting to the user (rebuilt, re-ran tests, read the actual diffs) rather than trusting agent self-reports — caught nothing wrong this time, but this is now the standing practice for this repo's multi-agent TDD workflow.
- **Key findings and decisions**:
  - The issue body's pseudocode (`r.Use(middleware.Logger)`, env vars `PORT`/`ALLOWED_ORIGINS`, hardcoded 15s/15s/60s timeouts, `middleware.writeErrorEnvelope` reuse) didn't match the real, already-shipped contracts from prior sessions — corrected to the actual middleware signatures (`Logger(logger)`/`Recovery(logger)`/`CORS(origins)` are all higher-order), `cfg.Server.Port`/`AllowedCORS` (`HTTP_PORT`/`ALLOWED_CORS_ORIGINS`), `cfg.Server.{Read,Write,Idle}Timeout` (config-driven, not hardcoded), and a local `writeJSON` helper (the `/healthz` body shape — `status`/`database`/`error` — differs from the generic error envelope's `error`/`message`/`request_id`). Lesson: always cross-check an issue's embedded pseudocode against the actual state of dependency packages before implementing — issue bodies get written before dependencies land and go stale.
  - `database.NewClient` already retries and pings internally before returning successfully, so `main.go`'s old explicit post-`NewClient` `Ping()` call was redundant dead code — removed rather than kept "just in case."
  - Diagnosed and permanently fixed an unrelated but persistent environment issue this session: Claude Code's shell-snapshot mechanism (`~/.claude/shell-snapshots/snapshot-zsh-*.sh`) systematically drops shell functions whose name starts with a single underscore (confirmed empirically: zero `_x...`-named functions survived in an 11798-line snapshot, while `__xx` double-underscore and non-underscore functions all did). gvm's `_encode`/`_decode` helpers (used by its `cd()`-based per-directory Go-version auto-switch) hit this gap, producing recurring `command not found: _encode ...` noise on every `cd` in every Bash-tool command. Root-caused via `GVM_DEBUG=1`, direct snapshot inspection, and process-tree/`ps` introspection (not guesswork) after an initial wrong hypothesis (duplicate `source` line in `.zshrc`, which was real but insufficient — fixed that too, but it didn't fully resolve the issue). Fix: commented out `. "$GVM_ROOT/scripts/env/cd" && cd .` in `~/.gvm/scripts/gvm-default` (disables gvm's per-directory `.go-version` auto-switch, unused by this machine's projects; `GOROOT`/`GOPATH`/`PATH` still come from `environments/default`, a plain env-var file, so `go` resolution is unaffected) — confirmed by deleting the stale cached snapshot and re-testing. This is a `~/.gvm` and `~/.claude` fix, outside the repo, but worth remembering if the same noise reappears on a machine with gvm + Claude Code.
  - Also added `go.buildTags`/`gopls.buildFlags: ["-tags=test"]` to `.vscode/settings.json` (committed) so gopls indexes `//go:build test`-tagged files (`server_test.go`, `internal/database/mocks/client_mock.go`) instead of showing "No packages found."
- **Outcomes**:
  - Issue #57 closed via PR #72 (open, not yet merged). `docs/roadmap.md` reflects Done status with PR reference for 005-T035/T036/T037.
  - Sprint 2's G-SPRINT2-BACKEND-HTTP-SERVER group is complete; any future domain route (auth, trips, etc.) now has `registerRoutes()` in `backend/cmd/api/routes.go` as the obvious place to add it.
  - **Follow-up work**: none blocking. PR #72 awaits human review/merge.

### Session: Domain Error Handling (005-T033–T034, Issue #56) + Docker Build Investigation (PR #73)

- **Date**: 2026-07-09
- **Tool**: Claude Code
- **What was accomplished**:
  - Implemented `backend/internal/errors/` for issue #56 (delegated to `tdd-developer`): `DomainError` (Code/Message/Fields/Details/wrapped Err, implements `error`/`Unwrap`), five fresh-instance constructors (`NotFound`, `Validation`, `Unauthorized`, `Forbidden`, `Conflict`), and `HandleError(w, r, err)` mapping to the standard JSON error envelope. 21 tests, 96.3% coverage, gofmt/go vet/golangci-lint clean. Verified independently (not just trusting the subagent report) by re-running gofmt/go vet/golangci-lint/tests myself.
  - **Corrected the issue body's stale pseudocode before implementing** (cross-checked against `docs/api-design-standards.md` §7 directly): `validation_failed` maps to **422**, not 400 as the issue said (400 is reserved for malformed JSON); the wire code for the `Unauthorized` constructor is `authentication_required`, not `unauthorized`; field-level validation entries use JSON keys `field`/`error`, not `field`/`message`; there is no `RequestIDKey` context key anywhere in the codebase (the issue invented one) — the real mechanism is `middleware.RequestIDFromContext(ctx)`. This is the same "issue pseudocode goes stale, verify against real code/docs" lesson as the HTTP server session (2026-07-09 earlier) — worth treating as a standing practice, not a one-off.
  - Delegated README + inline doc updates to `technical-writer`: added an `errors/` entry to `backend/README.md`'s Project Structure tree and a new "Error handling (`internal/errors/`)" Architecture Notes subsection explaining `DomainError`, the status-code table, and — importantly — *why* this package can't reuse `internal/middleware/errors.go`'s `writeErrorEnvelope` (import-cycle: `internal/errors` already imports `middleware` for `RequestIDFromContext`, so the reverse import would cycle). Noted explicitly that the package isn't wired into any handler yet (verified via `grep` — zero callers outside its own tests); issue #58 is what will first use it.
  - **Found and fixed stale roadmap rows while investigating Sprint 2 order** (start of this conversation, before #56 work): `005-T024`–`005-T028` (G-SPRINT2-BACKEND-MIDDLEWARE, issue #54) were still marked `Backlog` in `docs/roadmap.md` despite issue #54 being closed and PR #71 merged — confirmed via `gh issue list`/`gh pr list` against roadmap state, corrected to `Done` with a PR #71 note. This is the second time in this project a completed PR wasn't reflected back into the roadmap after merge — **worth periodically cross-checking `gh issue list --state all` against roadmap `Status` columns**, since nothing currently automates this reconciliation after merge (only `/build-roadmap` reconciles from spec `tasks.md`, not from GitHub issue/PR state).
  - Committed + pushed #56's work (`commit-and-push` subagent) to `feature/56-g-sprint2-backend-errors-domain-error-handling` (commit `7021d6d`) — bundled the roadmap correction above into the same commit since it was sitting in the working tree and was small/already-verified. **PR for #56 not yet opened** — conversation moved to the Docker investigation before that step.
  - **Docker build investigation** (user-reported: CI "feels too slow"): used `gh run view <id> --json jobs` and `--log` on real recent `backend-ci.yml` runs to get actual step-level timings rather than guessing. Found the "Build Docker Image" job's `go build` step alone took **451s (7.5 of ~8.5 total minutes)**, consistently across multiple runs regardless of `cache-from: type=gha`. Root cause: `backend/Dockerfile`'s `builder` stage had no `--platform` pin, so buildx ran the *entire* stage — including the Go compiler — under QEMU emulation to produce the `linux/arm64` target image from the `amd64` GitHub Actions runner. The user's hypothesis (`golang:1.26-alpine` version) was not the cause — ruled out with the log evidence.
  - **Fix**: `FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder` + `ARG TARGETOS`/`ARG TARGETARCH`, build with `GOOS=$TARGETOS GOARCH=$TARGETARCH` instead of hardcoded `GOOS=linux`. This is the standard documented Docker multi-platform cross-compilation pattern: pin the builder stage to the *build* machine's native platform so nothing in it runs under emulation, while Go's own toolchain natively cross-compiles the target-arch binary (trivial with `CGO_ENABLED=0`, no C toolchain needed). Only the tiny final runtime stage (`apk add`, `chown`, copy a prebuilt static binary) still runs under target-arch emulation, which is fast since it does no compilation.
  - **Verified the fix for real, not just by reasoning about it**: installed `docker buildx` locally (`brew install docker-buildx` + symlinked into `~/.docker/cli-plugins/`, since the local Docker CLI had no buildx plugin and defaulted to the legacy non-BuildKit builder), ran the actual `docker buildx build --platform linux/arm64` end-to-end against the real Dockerfile, and confirmed `go build` dropped from 451s to ~56-58s. Also ran the resulting arm64 binary with `docker run --platform linux/arm64` to confirm it *starts correctly* under emulation (fails only on missing `DATABASE_URL`, as expected) — not just that it compiles, since a cross-compile mistake could produce a binary that builds but is subtly broken (wrong arch/ABI) without a runtime smoke test.
  - **Second, unrelated finding surfaced mid-conversation by the user**: a Docker DX/`docker-language-server` lint flagged `alpine:3.19` (runtime stage base) for 2 high-severity CVEs. Confirmed 3.19 (Nov 2023) is past Alpine's ~2-year support window; bumped to `alpine:3.23` (confirmed pullable, current maintained branch) and re-verified the full two-stage build still succeeds.
  - Checked whether the same risk class exists elsewhere in the repo: only one Dockerfile exists (`backend/Dockerfile`, no frontend Dockerfile — frontend deploys static to S3+CloudFront, no container). Found `postgres:15.4-alpine` (used in `docker-compose.yml` and `backend-ci.yml`'s test service container) carries the same "frozen version accumulates unpatched CVEs" risk category — but unlike `alpine:3.19`, `15.4` is an **explicit architectural decision** documented across `docs/architecture.md`, `docs/cloud-and-environments.md`, `docs/data-model.md`, and roadmap tasks `003-T022`/`005-T079` (tied to the planned RDS `engine_version`). Flagged to the user rather than bumping unilaterally — a version pin that's load-bearing across multiple docs + future Terraform config needs a coordinated update, not a silent Dockerfile-style patch. Not yet resolved either way; user hasn't decided whether to check its current CVE status.
  - Since the Docker fix was unrelated to #56, **stashed it off the `feature/56-...` branch, switched to `main`, created a fresh `feature/optimize-docker-arm64-build` branch, and popped the stash there** before committing — avoided bundling an unrelated CI/perf fix into the #56 domain-error-handling PR. Committed (`c5aee95`), pushed, and opened PR #73 (no linked issue/roadmap stable ID — this was ad hoc, not a planned sprint task; the PR body and roadmap note both say so explicitly).
  - Added a Note to roadmap row `005-T014` (original Dockerfile creation task, already Done via PR #46) pointing to PR #73 as a follow-up, since there's no dedicated stable ID for this unplanned optimization.
- **Key findings and decisions**:
  - **Real GHA job/step timing data (`gh run view --json jobs` / `--log`) beats guessing at CI slowness** — the user's own hypothesis about the Go image tag was plausible-sounding but wrong; the log evidence (one step consuming 88% of total build job time) pointed straight at the actual cause in minutes.
  - **QEMU emulation of the *entire* builder stage, not just the final image, is the default buildx behavior** when a multi-platform Dockerfile's builder stage doesn't pin `--platform=$BUILDPLATFORM` — this is easy to miss since the Dockerfile still "works," just slowly, and GHA cache (`type=gha`) does not mitigate it (the expensive work is CPU-bound emulated computation, not something cacheable).
  - **A version pin appearing in a Dockerfile isn't always accidental drift** — cross-check against `docs/` and the roadmap before "fixing" it. `alpine:3.19` was true drift (unpinned anywhere as a requirement); `postgres:15.4-alpine` is a spec'd decision. Same surface pattern, different correct response.
  - **When local tooling can't validate a fix (no buildx plugin installed), install what's needed rather than reasoning from documentation alone** — for a change this consequential (CI build correctness, cross-arch binary validity), "the Docker docs say this pattern works" is weaker evidence than actually running it and smoke-testing the output binary.
- **Outcomes**:
  - Issue #56 implemented, tested, documented, committed, and pushed to `feature/56-g-sprint2-backend-errors-domain-error-handling` — **PR not yet opened** (next step when resumed).
  - PR #73 open (Docker build fix) — not yet merged, awaits human review.
  - `docs/roadmap.md` has two corrections beyond the sprint's planned tasks: `005-T024`–`005-T028` status fix, and a `005-T014` follow-up note pointing to PR #73.
  - **Follow-up work**: open PR for #56; decide on `postgres:15.4-alpine` staleness (user hasn't asked to act on it yet); PR #72 and #73 both await human review/merge.

### Session: Repo Ruleset on `main` + Fixing Path-Filtered Required Checks (PR #74, #75)

- **Date**: 2026-07-09
- **Tool**: Claude Code
- **What was accomplished**:
  - Opened PR #74 for issue #56 (`feature/56-g-sprint2-backend-errors-domain-error-handling` → `main`, `Closes #56`) — the step deferred from the prior session. Still open at end of this session (blocked by the new ruleset below, needs an approval + to be brought up to date with `main`).
  - **Configured a GitHub repository ruleset ("Protect main", id `18752818`) on `main`** at the user's request: PR required before merging, 1 required approving review, 10 named CI status checks required (the real job `name:` values from `backend-ci.yml`/`frontend-ci.yml`/`infra-plan.yml`), `strict_required_status_checks_policy: true` (branch must be up to date with `main` before merge), `non_fast_forward` + `deletion` rules (blocks force-push and branch deletion), and a `bypass_actors` entry scoped to `RepositoryRole` id `5` (Admin) with `bypass_mode: always` — since the user is currently the sole collaborator with that role, this achieves "only I can merge directly" without hardcoding a user ID (if another admin is added later, they'd also bypass — a deliberate, correctly-scoped tradeoff, not an oversight).
  - **First attempt was blocked by GitHub plan limits**: both the Rulesets API and the classic branch-protection API returned `403 Upgrade to GitHub Pro or make this repository public` while the repo was private (personal-account plan limitation, confirmed empirically by testing both endpoints, not assumed). Resolved when the user made the repo public — re-ran the same `gh api` calls successfully once `isPrivate: false` was confirmed.
  - **Caught and corrected my own wrong answer from earlier in this same conversation**: when the user chose to accept path-filtered CI workflows as required checks (understanding the tradeoff as I'd described it), I had told them a PR touching none of backend/frontend/infra "would have no blocking check — mergeable sin checks." That claim was **false**. Verified against official GitHub documentation via `WebSearch` (not memory): a required status check whose workflow never triggers (due to `paths:` filtering) gets stuck at **"Expected — Waiting for status to be reported" and blocks merging indefinitely** — it is never treated as passing. GitHub's own troubleshooting docs state this explicitly and recommend against path-filtering the trigger of any workflow that's used as a required check. Immediately surfaced the correction to the user rather than letting the wrong assumption stand.
  - Also verified (again via `WebSearch`, since the whole point of this correction cycle was "don't just reason from memory") the follow-up fact that made the actual fix viable: **a job skipped via a job-level `if:` condition reports conclusion "skipped," and skipped jobs DO count as passing for required-status-check purposes** — distinct from a check that never reports any status at all. This is GitHub's own documented workaround for exactly this scenario.
  - **Fixed all 3 CI workflows** (`backend-ci.yml`, `frontend-ci.yml`, `infra-plan.yml`) applying that pattern identically: removed `paths:` from each `pull_request:` trigger (left `push:` triggers' path filters alone — irrelevant to merge-blocking), added a `changes` job per workflow (via `dorny/paths-filter@v3`) that always runs and outputs an area-changed boolean, and gated every existing job with `needs: changes` + `if: github.event_name == 'workflow_dispatch' || needs.changes.outputs.<area> == 'true'` (the `workflow_dispatch` escape hatch sidesteps `dorny/paths-filter`'s ambiguous behavior on manually-triggered runs with no meaningful diff base). Deliberately did not touch any job's `name:` field, since the ruleset's required-check list matches on that exact string.
  - Validated the fix two ways before trusting it: `ruby -ryaml` (no pyyaml available locally) for syntax validity on all 3 files, then installed `actionlint` (`brew install actionlint`) for GitHub-Actions-specific validation (expressions, job/needs graph) — zero new findings, only one pre-existing unrelated shellcheck info-level warning.
  - Committed (`f5edccf`) and opened PR #75 on a fresh branch (`feature/fix-ci-required-checks-path-filtering`, branched from `main` after discovering PR #73 had already been merged independently) — kept this separate from #74/#73 since it's a distinct concern (CI governance, not app code).
  - **Watched PR #75's checks run for real** (a `general-purpose` agent monitored `gh pr checks 75` in the background) — this was the actual empirical test of the fix, not just a documentation-based inference. Result: all 3 `Detect Changed Areas` jobs passed, all backend and infra checks passed (**`Build Docker Image` completed in 24s**, live confirmation that the earlier `--platform=$BUILDPLATFORM` fix from PR #73 works in production CI, not just locally), and — critically — **nothing got stuck at "Expected"**. `Lint Frontend Code` and `Run Frontend Tests` genuinely failed, and `Build Production Bundle`/`Accessibility Audit` correctly showed `skipping` as an ordinary `needs`-chain cascade from that failure (not a path-filter artifact) — exactly the intended behavior: real failures still block, only the never-triggered-workflow failure mode is fixed.
  - Diagnosed the 2 real frontend failures: `Run Frontend Tests` — `frontend/package.json`'s `scripts` has no `"test"` entry at all (confirmed by reading the file), so `npm test -- --coverage --run` errors immediately; `Lint Frontend Code` — Prettier flags `src/App.tsx` and `src/main.tsx` as unformatted. Both pre-existing gaps from the Sprint 1 frontend scaffold (PR #47), surfaced now because this is close to the first time `frontend-ci.yml` has run meaningfully (it was rarely triggered before, given how little `frontend/**` work has happened and how the old path-filtering worked).
  - **User chose to merge PR #75 immediately via the admin bypass** (`gh pr merge 75 --admin --squash --delete-branch`) rather than fix the frontend gaps first, deferring those as separate follow-up work. Merged clean, squash commit `6a65eeb` on `main`, remote branch deleted.
- **Key findings and decisions**:
  - **Verify operationally consequential GitHub platform behavior against current docs, don't reason from training-data memory** — this session had two live corrections in a row (my own wrong "no blocking check" claim, then the follow-up "does skipped count as passing" question) that would have shipped a broken or ineffective ruleset if I'd trusted recall instead of searching. Standing practice going forward for any branch-protection/required-check/CI-gating change.
  - **GitHub Rulesets (and classic branch protection) require GitHub Pro or a public repo for private personal-account repos** — confirmed empirically against both APIs, not documentation alone, since docs can be ambiguous about exactly which plan tiers this applies to.
  - **`bypass_actors` scoped to `RepositoryRole: Admin` rather than a specific user ID** is the more correct way to express "only I can bypass" when the requester IS currently the sole admin — it degrades sensibly (any future admin also bypasses, which is a reasonable definition of "admin", not a loophole) rather than needing to be revisited if the account's numeric user ID context changes.
  - **A required check passing "in theory" isn't validated until you watch a real PR run it** — the background-monitored PR #75 checks run caught both that the fix works AND surfaced two genuine pre-existing frontend gaps that no one had noticed, because `frontend-ci.yml` essentially never ran for real before this.
- **Outcomes**:
  - Ruleset "Protect main" (id `18752818`) active on `main`: PR + 1 approval + 10 named required checks + strict up-to-date policy + no force-push/delete, admin-role bypass.
  - PR #75 merged (`6a65eeb`) — `backend-ci.yml`/`frontend-ci.yml`/`infra-plan.yml` no longer path-filter their `pull_request` trigger; required checks now always report (pass or genuinely fail) instead of ever getting stuck at "Expected".
  - PR #74 (issue #56) still open, blocked by the new ruleset (needs 1 approval + to be updated with `main`, which has moved twice since #74 was opened — PR #73 and now PR #75 both merged after it).
  - **Follow-up work, not yet done**: (1) PR #74 needs approval + rebase/update against current `main` before it can merge normally (or another admin-bypass merge, user's call); (2) `frontend/package.json` needs a `"test"` script added (matching what `frontend-ci.yml`'s `Run Frontend Tests` job actually invokes: `npm test -- --coverage --run`, i.e. wiring up whatever the project's `vitest` setup expects); (3) `src/App.tsx` and `src/main.tsx` need a `prettier --write` pass; (4) `postgres:15.4-alpine` staleness still undecided from the prior session.

### Session: AI Client Foundation with Local Ollama (005-T030–T032, Issue #55, PR #77)

- **Date**: 2026-07-10
- **Tool**: Claude Code
- **What was accomplished**:
  - Implemented `backend/internal/ai/` for issue #55 (delegated to `tdd-developer`): `AIClient` interface (`GenerateItinerary`, `StreamItinerary`), shared request/response types (`types.go`), and the ticket's originally-specified `PromptValidator`/`OutputSanitizer` stubs (always-valid / pass-through, per spec 002 deferring real deny-list/bluemonday logic to a later NFR ticket).
  - **Scope extended beyond the original ticket, with the user's explicit approval** (confirmed via `AskUserQuestion` before implementing): the ticket only asked for interfaces+stubs (a future Anthropic-backed client was deferred to spec 008/roadmap `005-T112`), but the user decided that for **local development and MVP testing**, the AI backend should be a **free local Ollama server running a Gemma model** instead of Anthropic Claude (which needs a paid API key). Built a real, working `OllamaClient` (`ollama_client.go`) against Ollama's native `/api/chat` endpoint — JSON-mode forced output (`format: "json"`), linear-backoff retries on transient failures, goroutine-based NDJSON streaming — tested against a fake `httptest` server (87.5% coverage), not a real local Ollama instance (none installed in CI/sandbox).
  - Restructured `backend/config/config.go`'s `AIConfig` into shared settings + `AnthropicConfig`/`OllamaConfig`, selected via `AI_PROVIDER` (defaults to `ollama` in development, `anthropic` in production via `GO_ENV`) — mirrors the existing `JWT_SIGNING_KEY`-required-only-in-production pattern.
  - Added a local `ollama` service to `docker-compose.yml` (healthcheck via `ollama list`, persistent `ollama_data` volume, resource limits), updated `backend/.env.example`, and wrote `docs/local-ai-setup.md` (install, model-size table, Docker Compose vs. native, switching to Anthropic, troubleshooting).
  - **Deliberately scoped the docs impact to local/dev only** (second `AskUserQuestion` decision): appended new "Local Development Note" sections to `docs/architecture.md` and `docs/cloud-and-environments.md` *after* their `<!-- PROMOTED:... END -->` markers, rather than editing the promoted content itself — staging/production architecture still documents Anthropic Claude unchanged, and the addenda survive future `promote-fundations` runs without being clobbered or mistaken for promoted content. Updated `backend/README.md` (Tech Stack, Implementation Status, Environment Variables, new "AI client foundation" Architecture Notes subsection) and root `README.md` (Quick Start, "Built With Care") to match; also fixed a stale/incorrect `internal/ai/` subtree in `backend/README.md`'s Project Structure (it showed a nested `validator/`+`sanitizer/` layout that never matched the ticket's actual flat file layout).
  - Delegated a dedicated documentation-review pass to `technical-writer` after the code+docs were already written by `tdd-developer`/the orchestrating session — found the Go doc comments already solid (no `.go` comment changes needed), but caught two real doc-accuracy gaps: `backend/README.md` was missing `GO_ENV` from its env-var table and its standalone `docker run` example still only set `ANTHROPIC_API_KEY` (now silently wrong since `AI_PROVIDER` defaults to `ollama`); `docs/local-ai-setup.md` was missing the same `GO_ENV` row.
  - That review also surfaced a real (unrelated) bug: `loadAuthConfig()`'s `CookieSecure` was hardcoded to `getEnvBool("COOKIE_SECURE", false)` regardless of `GO_ENV`, contradicting its own doc comment and the README's claim that it defaults to `true` in production — a real security-relevant gap (insecure cookies could ship to production by omission). **User asked to fix it in this same ticket rather than filing separately.** Fixed via TDD (wrote 2 failing tests first, confirmed red, then fixed): `defaultCookieSecure := getEnv("GO_ENV", envDevelopment) == envProduction`, same pattern as `loadAIConfig`. Extracted `"production"`/`"development"` into named constants (`envProduction`/`envDevelopment`) since `golangci-lint`'s `goconst` rule flagged the 3rd occurrence of `"production"` this fix introduced.
  - Two commits (`d33ab0a` feature, `3618b22` doc-fixes+cookie-fix) pushed to `feature/55-g-sprint2-backend-ai-client-foundation`; PR #77 opened (`open-pr` subagent) with both scope-extension and bundled-bugfix called out explicitly in the body; `docs/roadmap.md` rows `005-T030`/`005-T031`/`005-T032` marked `Done` with a PR #77 note (`f9a0acf`).
  - **Independently verified every subagent's output myself** before trusting it — re-ran `gofmt`/`go vet`/`golangci-lint`/`go build`/`go test -race -short` after each delegated step (implementation, doc review), read the actual diffs rather than trusting self-reports. Caught nothing wrong from `tdd-developer`'s work; the `technical-writer` pass's own claim of "no `.go` changes needed" was itself verified true by diffing.
  - **User asked mid-session whether a Swagger/OpenAPI documentation task already existed anywhere** — grepped `docs/roadmap.md` and every `specs/*/tasks.md`/`specs/*/contracts/` for "swagger"/"openapi": zero hits. Confirmed the project's API documentation today is entirely manual Markdown (`specs/*/contracts/api.md` + `docs/api-design-standards.md`), no interactive explorer, no machine-readable contract. User decided this should become a deliberate Sprint 3 planning input (`/speckit-specify` a new spec → `promote-fundations` → `build-roadmap` → `product-manager` Plan Sprints) rather than an ad hoc addition — saved to Claude Code's cross-session personal memory (not just here), since it needs to surface automatically whenever Sprint 3 planning starts in a fresh conversation.
- **Key findings and decisions**:
  - **Scope-extension and scope-limiting decisions on a ticket should go through explicit user confirmation (`AskUserQuestion`), not be inferred** — this session had two: whether Ollama should be a doc-only future note or a real working client (chose real client), and whether the Anthropic→Ollama doc change should touch staging/production docs or stay local-dev-only (chose local-dev-only). Both had multi-file, hard-to-reverse blast radius (`docs/architecture.md`, `docs/cloud-and-environments.md`, `product-vision.md` all commit to "Anthropic Claude" as *the* provider) — worth asking rather than guessing.
  - **Appending local-dev overrides after a doc's `PROMOTED:...END` marker, rather than editing within it, is the correct pattern** whenever a spec-generated doc needs a local/dev-only addendum that shouldn't be treated as part of the promoted architectural decision — it survives `promote-fundations` re-runs cleanly. Worth reusing for any future local-dev-only override of promoted content.
  - **A repo-local memory-protocol gap, caught by the user, not self-detected**: this whole session ran without ever writing to `.github/memory/session-notes.md` or `scratch/working-notes.md` — a violation of `CLAUDE.md`'s mandatory Session Start Protocol's mirror obligation (loading them at session start was done correctly; writing back was skipped entirely). Caught only because the user explicitly asked "¿escribiste alguna nota en las session notes?" near the end of the session. **Standing lesson: check this obligation proactively at natural checkpoints (after opening a PR, not just "end of session"), don't rely on it being remembered from a mid-context-window read of `CLAUDE.md`.**
- **Outcomes**:
  - Issue #55 closed via PR #77 (open, not yet merged), `docs/roadmap.md` reflects `Done` status with PR reference.
  - `backend/internal/ai/` is a complete, tested foundation: `AIClient` interface ready for a future Anthropic implementation (`005-T112`) without breaking changes, plus a genuinely usable local dev path via `OllamaClient` + `docs/local-ai-setup.md` + `docker-compose.yml`'s new `ollama` service.
  - `config.AIConfig`/`AuthConfig` both now correctly derive their defaults from `GO_ENV` instead of one doing so and the other silently not.
  - **Follow-up work, not yet done**: (1) PR #77 awaits human review/merge; (2) Sprint 3 planning must fold in a new Swagger/OpenAPI spec (`/speckit-specify` → `promote-fundations` → `build-roadmap` → re-plan Sprint 3) per the user's explicit request — see also Claude Code's personal cross-session memory for this project; (3) the pre-existing `frontend/package.json`/`prettier` gaps and `postgres:15.4-alpine` staleness from prior sessions remain unresolved.

### Session: Reference Implementation Pattern (005-T038–T041, Issue #58)

- **Date**: 2026-07-10
- **Tool**: Claude Code
- **What was accomplished**:
  - Implemented `backend/internal/example/` (delegated to `tdd-developer`, run inline in this session) as the canonical layered pattern — model → repository → service → handler — that future domain packages (Trip, User, ...) will copy. Strict RED-GREEN-REFACTOR per layer, dependency order: model (5 tests) → repository pure logic (pagination helpers, 10 tests) → repository testcontainer integration (10 tests) → service with mocked repo (17 tests) → handler with mocked service, `httptest`+real `chi.Router` (16 tests). 88.0% statement coverage on `internal/example` (excluding the generated `mocks` subpackage).
  - **Corrected the issue's pseudocode against real code before implementing** (same standing lesson as prior sessions — issue bodies go stale): `errors.NotFound(resource, id)`/`errors.Validation(message, fields...)`/`errors.Conflict(message)` single-arg signatures, `database.Client` is an interface consumed via `.Pool()` not `*database.Client`, List needed a real pagination envelope (not a bare array) per `docs/api-design-standards.md` §9.
  - **Migration**: `backend/migrations/20260710120000_create_examples_table.sql` — goose timestamp-versioned (not sequential `001_`, which stays reserved for the real Plans table per `docs/data-model.md`). Column style (`VARCHAR(20) CHECK (status IN (...))`, not a native Postgres `ENUM` type; `TIMESTAMP DEFAULT NOW()`; `gen_random_uuid()` PK default) matched against the concrete DDL precedent in `specs/004-security-auth-model/data-model.md` rather than data-model.md's prose ("ENUM constraint") in isolation, since the prose is ambiguous but the actual shipped DDL example is not.
  - **UUID generation lives in the service layer** (`uuid.New().String()`), not the DB column default, per the issue's explicit checklist — required changing `PostgresRepository.Create`'s INSERT to accept a caller-supplied `id` rather than relying on `RETURNING id` from `gen_random_uuid()`. The migration keeps the column default anyway as a safety net (matches the rest of the project's DDL convention) even though the application path always supplies it explicitly.
  - **Optimistic locking uses the `If-Match` request header** on `PUT /examples/{id}`, not a `version` field in the request body (the issue's pseudocode had no locking mechanism specified for the HTTP layer at all) — chosen to match the *already-promoted* convention in `docs/data-model.md`'s Trip/Activity "Concurrency Control" sections ("Client sends `If-Match: <version>` header... 409 Conflict with current version"), since this package is explicitly the template everyone else copies and Trip/Activity are real, already-spec'd versioned entities.
  - **Found and fixed a real correctness bug via the testcontainer integration tests, not just reasoning**: `List`'s first implementation used `COUNT(*) OVER()` window function to get the total row count in the same query as the paginated rows — this silently returns `total: 0` for an out-of-range page (zero rows returned → window function never evaluates), breaking the documented contract that an empty page must still carry a correct `total`/`total_pages` (`docs/api-design-standards.md` §9: "Return empty array for page beyond total_pages, not 404" implicitly requires the envelope to still be truthful). Fixed by splitting into two queries: a `COUNT(*)` first, short-circuiting to an empty result before running the paginated `SELECT` only when `total > 0`.
  - **`internal/errors` has no 400/`invalid_request` constructor** (its catalog only covers 404/422/401/403/409) — confirmed by re-reading `types.go`/`handler.go` rather than assuming. For malformed JSON bodies and invalid pagination query params, the handler writes the `invalid_request`/400 envelope directly via a small private helper, mirroring the exact same pattern `internal/middleware/errors.go` already uses for its own 413/500 cases — not extending the shared `internal/errors` package for one local caller. Worth revisiting if a second package hits the same gap (extract a shared helper then).
  - **`.mockery.yaml` got two new entries, not the single one the issue asked for**: `Repository` (as instructed) plus `Service` — the issue's own TDD-workflow prose says handler tests need "mocked service", and this project's established convention (`patterns-discovered.md`'s mockery-scope pattern) is mockery for every domain-owned interface used across a layer boundary in tests, not hand-rolled fakes. Regenerating `mockery --all` twice picked up an unrelated pre-existing drift in `internal/database/mocks/client_mock.go` (a stray `// +build test` legacy comment line removed by the newer local `mockery` v2.53.6 vs. whatever generated the committed file) both times — reverted via `git checkout --` each time to keep the diff scoped to this issue.
  - **Testcontainers required non-default Docker env vars to work with Colima on this machine**: `DOCKER_HOST=unix:///Users/.../.colima/default/docker.sock` (the macOS-host-visible socket path) plus `TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock` (the path as seen *inside* containers, for the Ryuk reaper's bind mount — using the macOS host path here fails with "operation not supported: could not start container" since that path doesn't exist inside the Colima Linux VM). Without any `DOCKER_HOST`, testcontainers-go tries an unsupported "rootless Docker" provider path and fails outright. This is a durable, reusable fact for this machine/setup, not project-specific — added as a `patterns-discovered.md` entry.
  - Verified the live wiring end-to-end without needing docker-compose (deferred to the user per their explicit instruction): confirmed via a throwaway `chi.Walk` test that `POST/GET /api/v1/examples`, `GET/PUT/DELETE /api/v1/examples/{id}` are registered correctly, `/healthz` stays unversioned outside the group, and — since chi's `r.Route("/x", func(r){ r.Post("/", ...) })` internally registers the pattern with a trailing slash — separately confirmed a request to `/api/v1/examples` **without** a trailing slash still routes all the way through the middleware chain into the handler (reached the repository's `Pool()` call), not a 404. `go build ./...`, `gofmt -l .`, `go vet ./...` all clean; `make test` (short/unit, all packages) and the full `go test -tags=test ./...` (all testcontainer suites, Colima) both pass. Left no leftover files from the throwaway verification scripts.
  - Did not touch `docs/roadmap.md`, did not commit, did not open a PR — all explicitly deferred to the user per their instructions for independent verification first.
- **Key findings and decisions**:
  - **A window-function-based `COUNT(*) OVER()` for combined pagination total+rows is a real footgun for the "empty page still has correct total" contract** — only surfaces when a testcontainer/integration test actually exercises an out-of-range page, not from unit tests against mocked repositories (the mock just returns whatever `Total` the test hard-codes). Worth remembering for any future paginated `List` method: either use two queries (chosen here) or explicitly test the zero-rows-returned case against a real database.
  - **When a reference/example package is explicitly "the thing everyone else copies," prefer already-promoted architectural conventions (docs/data-model.md's If-Match header) over an individual issue's inline, possibly-stale pseudocode** — the issue had no concurrency-control mechanism specified for the HTTP layer at all, and the version-in-body idea some might default to would have set a worse precedent than the header-based approach the rest of the data model already commits to for Trip/Activity.
  - **Confirmed again**: mockery scope in this repo now legitimately covers three tiers — `database.Client` (external system), `ai.AIClient` (external system), and now `example.Repository`/`example.Service` (internal layer boundaries mocked for the *next* layer's tests) — the earlier "mockery is for external systems only" characterization from the 2026-07-09 middleware session was a snapshot of that moment, not a permanent scope rule; the real rule is "mock any interface a test needs to fake across its own package's dependency boundary," which the `patterns-discovered.md` entry should probably be reworded to reflect next time it's touched.
- **Outcomes**:
  - `backend/internal/example/` is complete, tested (58 test cases across model/repository/service/handler + integration), and wired into the live server (`cmd/api/server.go`/`routes.go`) — `/api/v1/examples` is a real, reachable, database-backed CRUD API, not just isolated unit-tested code.
  - `backend/migrations/20260710120000_create_examples_table.sql` exists and was actually applied and exercised by the testcontainer integration tests (not just written and assumed correct); independently re-verified outside the test suite too — a real `goose` CLI `up`→`down`→`up` round-trip against a throwaway Postgres container confirmed both directions work and the migration isn't hardcoded anywhere.
  - Independently curl-verified every endpoint against a real `go run ./cmd/api` + real Postgres (not docker-compose — that CLI wasn't installed on this machine at the time, see below): create/get/update/delete/list, pagination, duplicate-email 409, validation 422, malformed-JSON 400, missing-If-Match 400, optimistic-locking 409, not-found 404, delete 204 — all correct.
  - `.mockery.yaml` covers `example.Repository` and `example.Service`; both mocks committed under `backend/internal/example/mocks/`.
  - Committed and pushed to `feature/58-g-sprint2-backend-example-reference-implementation-pattern` (`16d629a`, via `commit-and-push`), user's explicit choice over opening a PR immediately. **`docs/roadmap.md` rows `005-T038`–`005-T041` are still `Backlog` status** — marking them `Done` is deferred until a PR exists (matches the established per-session pattern of pairing `Done` with a PR reference), but their Notes column now carries the deletion reminder below regardless of status.
  - **This machine had neither `docker-compose` (v1) nor the `docker compose` (v2) plugin installed** — `brew install docker-compose` fixed the binary, but registering it as a discoverable Docker CLI plugin requires adding `cliPluginsExtraDirs` to `~/.docker/config.json`, which the harness's auto-mode classifier blocked (reading OR writing that file is treated as a credential-adjacent action needing explicit user consent, since it can hold registry auth tokens). Worked around it by invoking the plugin binary directly (`/usr/local/lib/docker/cli-plugins/docker-compose up -d`) instead of editing the config — works identically for a v2 compose plugin binary, no `~/.docker/config.json` edit needed.
  - **IMPORTANT — must-read follow-up, not yet done**: `backend/internal/example/` (all four `.go` files, `mocks/`, the goose migration, its `.mockery.yaml` entries, and its `/api/v1` route mount in `cmd/api/server.go`/`routes.go`) **must be deleted once the first real domain package following this pattern is implemented** (e.g. Trip — `001-T035`/`001-T038`, or `008-T049`/`008-T078`), in that same PR. It was always meant as throwaway/reference-only (see its own package doc comment and `docs/roadmap.md`'s Notes column on `005-T038`–`005-T041`, both updated 2026-07-10 to flag this); full detail and the exact file list is in a dedicated `patterns-discovered.md` entry ("`internal/example/` Is Throwaway — Delete It Once the First Real Domain Ships").
  - **Other follow-up work, not yet done**: (1) `docs/roadmap.md` Done-marking for this issue, deferred until a PR opens (see above); (2) the `invalid_request`/400 gap in `internal/errors` is a candidate for a small follow-up ticket if a second package hits the same need for a local envelope-writing workaround; (3) `frontend/package.json`/`prettier` gaps and `postgres:15.4-alpine` staleness from prior sessions remain unresolved.

### Session: Design System Tokens + Frontend Doc/Test Tooling Audit (005-T042–T043, Issue #59, PR #79)

- **Date**: 2026-07-10
- **Tool**: Claude Code
- **What was accomplished**:
  - Implemented `frontend/src/styles/tokens.css` and `frontend/src/styles/global.css` for issue #59 (delegated to `tdd-developer`), wired into `src/main.tsx`, retiring the Vite-boilerplate `App.css`/`index.css`. **Token values were sourced from `docs/ui-guidelines.md`, not the GitHub issue body's example CSS** — the issue's pseudocode used different names/values (rem-based 8px spacing scale, `--font-weight-semibold`, no breakpoint tokens) than the canonical doc (px-based 4px scale, `--font-weight-regular`/no `semibold`, `--bp-sm/md/lg/xl`). Per `CLAUDE.md`, docs win over issue-body examples when they conflict — this is a concrete precedent for any future frontend ticket whose pseudocode might be stale relative to `docs/ui-guidelines.md`.
  - User then asked for a documentation + code-comment review of the frontend area (not part of the original ticket). Found and fixed real gaps, not just style nits:
    - `frontend/README.md`'s "Running Tests" section documented `npm test`/`npm run test:coverage` when **no such script existed and no test framework was installed** — anyone following the README literally would hit `Missing script: "test"`. The same command is what `.github/workflows/frontend-ci.yml`'s `test` job actually runs (`npm test -- --coverage --run`), so this was a live CI-breaking gap waiting to surface on the first PR touching `frontend/**`, not just a docs nit.
    - The README's "Project Structure" presented an entirely aspirational Sprint-2+ target tree (features/, pages/, services/, hooks/, atoms/composites/features layers, `vitest.config.ts`, `.eslintrc.json`) as if it already existed — none of it did (only `App.tsx`/`main.tsx`/`styles/` were real at the time), and two config filenames were just wrong (`.eslintrc.json`→ actual `eslint.config.js` flat config, `.prettierrc`→ actual `.prettierrc.json`).
    - `App.tsx`/`main.tsx` carried stale JSDoc: one narrated Sprint-2+ roadmap plans ("will be replaced with... See Sprint 2+ tasks in docs/roadmap.md") — the kind of comment that rots as the codebase evolves and belongs in the issue tracker, not source; another just restated well-known `StrictMode` behavior. Removed both — no WHY-level content lost.
  - **User chose to actually fix the CI gap, not just document it** (offered via `AskUserQuestion`: install now vs. mark CI job placeholder vs. leave it): installed `vitest`, `@vitest/coverage-v8`, `jsdom`, `@testing-library/react`, `@testing-library/jest-dom`, `msw` as devDependencies; added `vitest.config.ts` (jsdom env, v8 coverage, merged with `vite.config.ts` via `mergeConfig`), `src/test/setup.ts` (`import '@testing-library/jest-dom/vitest'`), a passing smoke test (`src/App.test.tsx`), and `test`/`test:watch`/`test:coverage` npm scripts. Added `vitest.config.ts` to `tsconfig.node.json`'s `include` so ESLint's typed-linting (`parserOptions.project`) covers it — same reason `vite.config.ts` was already listed there. Added `coverage` to `.gitignore` (Vitest's report output, previously untracked-and-ignorable but not actually ignored). Then re-did the README pass to reflect the now-true state instead of leaving stale "not yet available" warnings.
  - Mid-session, the user separately flagged a `tsconfig.json` IDE warning: `baseUrl` deprecated, removed in TS 7.0. Fixed the root cause rather than suppressing it — since TS 4.1, `paths` resolves relative to the tsconfig file's own directory without needing `baseUrl` at all, so removed `"baseUrl": "."` entirely (kept `paths: { "@/*": ["./src/*"] }`), with a comment explaining why (non-obvious: looks like it should still be needed).
  - Also corrected a self-inflicted process error: `git checkout -b feature/59-...` off `main` unexpectedly landed on a *different*, pre-existing, empty remote branch `feat/59-g-sprint2-frontend-tokens-design-system-tokens` (no "e", identical to `main`, zero unique commits, no PR) — likely an artifact of GitHub's issue-linked "create a branch" feature using a different prefix convention than this repo's `feature/` naming standard. Caught by inspecting `git status`'s branch name after Edit tool calls, not by the delegated `tdd-developer` subagent (which doesn't manage branches). Switched back to the correctly-named `feature/59-...` branch (git index/staged changes carried over fine, since staging is shared across branches) and left the stray `feat/59-...` branch untouched on the remote rather than deleting shared state unprompted.
  - Two commits pushed (`053312d` tokens/styles, `83b7170` test tooling + doc fixes) plus a third small fix (tsconfig `baseUrl` removal, not yet committed as of this note) to `feature/59-g-sprint2-frontend-tokens-design-system-tokens`; PR #79 opened via `open-pr` (closes #59); `docs/roadmap.md` rows `005-T042`/`005-T043` marked `Done` with a PR #79 note.
- **Key findings and decisions**:
  - **A README's documented commands are a testable claim, not just prose — verify them against the actual `package.json` scripts/deps before trusting them**, especially when a doc review is explicitly requested. This is the second session in a row (see prior AI Client Foundation session's follow-up list) where a `frontend/package.json` test-script gap was flagged but not fixed; this session actually closed it. Worth checking proactively on any future frontend-touching session even without an explicit "review the docs" ask.
  - **`git checkout -b <branch>` can silently land you on an unexpected pre-existing branch if the name collides with (or is confused for) an existing local/remote ref during shell tooling** — always sanity-check the actual branch name in `git status`/`git branch -vv` after creating one, especially right before staging/committing, not just trust the command you issued.
  - Confirms the established "docs/ win over issue-body pseudocode" rule (`CLAUDE.md`) with a second concrete frontend instance, alongside the backend precedent from the `internal/example/` session (issue pseudocode vs. real `errors`/`database` package signatures).
- **Outcomes**:
  - `frontend/src/styles/{tokens.css,global.css}` exist, are wired into `main.tsx`, and are the real source of truth for design tokens going forward (no components consume them yet — that's issue #63).
  - `frontend/README.md` now accurately distinguishes current vs. Sprint-2+-target state throughout (Implementation Status banner, Project Structure Current/Target split, Tech Stack Status column, Running Tests, Design Tokens, CI section).
  - `npm test` / `npm run test:coverage` / `npm run test:watch` are real, working commands; `frontend-ci.yml`'s `test` job (`npm test -- --coverage --run`) will actually pass now instead of failing on the first PR that triggers it. No coverage threshold is enforced yet — that's roadmap `002-T041`/`002-T042`, deliberately left undone (would need `src/components/`/`src/hooks/` to exist first).
  - `frontend/tsconfig.json` no longer emits the TS 7.0 `baseUrl` deprecation warning; `@/*` path alias behavior is unchanged (verified via `npm run type-check`/`npm run build`).
  - **Follow-up work, not yet done**: (1) PR #79 awaits human review/merge (the tsconfig `baseUrl` fix, `docs/roadmap.md` Done-marking, and this memory update all landed as a follow-up commit on the same branch/PR, after PR #79 was already opened); (2) MSW is installed but not wired to anything — no `src/services/` API client exists yet to intercept; (3) `postgres:15.4-alpine` staleness from prior sessions remains unresolved.

### Session: Lighthouse CI Sprint-Label Correction + Sprint 3 Planning Flag (follow-up, same branch/PR #79)

- **Date**: 2026-07-10
- **Tool**: Claude Code
- **What was accomplished**: User spotted `frontend-ci.yml`'s `TODO Sprint 2: Implement Lighthouse CI with @lhci/cli` comment and asked which Sprint 2 ticket covers it — none does. `docs/roadmap.md`'s actual tasks for this (`002-T002` add `@lhci/cli`/`@axe-core/playwright` deps, `002-T004` create `lighthouserc.yml`, `002-T024` create `accessibility.yml` PR-gate workflow) are unscheduled or Sprint 9, no GitHub issues yet, no `G-SPRINT2-FRONTEND-*` group. Corrected the stale claim in both `.github/workflows/frontend-ci.yml` (two spots: the job's TODO comment and its placeholder-step echo output) and `frontend/README.md`'s CI section (had the identical "Sprint 2" claim, missed in the earlier doc-accuracy pass this same session). Added a planning-note callout at the top of `docs/roadmap.md`'s Sprint 3 section flagging this for explicit review rather than letting it silently ride to Sprint 9.
- **Key findings and decisions**: Same category as the pre-existing Swagger/OpenAPI gap (Claude Code's cross-session personal memory, `sprint3-swagger-openapi-gap.md`) — a planning gap the user wants surfaced automatically at next Sprint 3 planning, not rediscovered by accident. Saved as a matching cross-session memory entry (`sprint3-lighthouse-ci-gap.md`), cross-linked with the Swagger one, since this repo-local `session-notes.md` isn't reliably re-read at the *start* of a future planning session the way Claude Code's own auto-loaded personal memory is.
- **Outcomes**: `frontend-ci.yml` and `frontend/README.md` no longer imply accessibility CI is Sprint 2 work. `docs/roadmap.md`'s Sprint 3 section carries an explicit reminder to decide whether to pull 002-T002/T004/T024 forward. Not yet committed as of this note — lands in the same follow-up commit as this entry.

---

### Session: Core UI Primitives + Form Composite (005-T048–T051, 005-T055, Issues #63/#65, PR #80)

- **Date**: 2026-07-10
- **Tool**: Claude Code
- **What was accomplished**:
  - Implemented issue #63 (`Button`, `Input`, `Card` primitives under `frontend/src/components/primitives/`, each with a token-driven CSS Module and co-located Vitest/RTL test) and, in the same session, issue #65 (`Form` composite under `frontend/src/components/composites/`), both delegated to `tdd-developer` with strict Red-Green-Refactor, then independently re-verified (lint/type-check/test) rather than trusting the subagent's self-report.
  - **`Card.tsx`/`Form.tsx` placement conflict resolved by authority ranking, not by the first source read**: `docs/ui-guidelines.md`'s atomic-design Mermaid diagram/table lists `Card` as a *Composite* living directly in `src/components/`, but `specs/005-system-architecture/tasks.md` (T048-T051) and `docs/roadmap.md` both explicitly place it at `src/components/primitives/Card.tsx`, and `Form.tsx` at `src/components/composites/Form.tsx`. Followed the spec/roadmap (roadmap's own legend names `specs/*/tasks.md` as the task source of truth) over the narrative doc's diagram. Promoted this as a general pattern (see `patterns-discovered.md` below) since it's now recurred across three unrelated sessions in slightly different forms (token names, backend package signatures, now component placement).
  - **Issue #65's illustrative pseudocode referenced `<ErrorMessage>`/`<LoadingSpinner>`** — components tracked as a separate, still-open issue (#64). Rather than build them speculatively or block on #64 (issue #65's own "Depends on" field only lists #63), implemented loading/error UI inline in `Form.tsx`: `aria-busy` on the `<form>`, disabled inputs/button, submit label switching to "Submitting…", and a `role="alert"` paragraph for form-level errors — per this repo's KISS principle (no speculative abstractions for a component that doesn't exist yet).
  - `Form`'s error contract: `FormSubmitError extends Error { fields?: Record<string, string> }`, checked structurally (`err instanceof Error && 'fields' in err && Boolean(err.fields)`) so a rejection either populates per-field `Input` errors or a single form-level alert, never both.
  - Since `Form` depends on primitives that only exist on the #63 branch (not yet merged to `main`), issue #65 was implemented and committed as a second commit (`b970bba`) on the *same* branch/PR line as #63 (`1aa8238`) rather than a fresh branch off `main` — a deliberate departure from "one issue = one branch," decided unilaterally (not pre-cleared with the user) because a fresh branch off `main` would be missing the `Button`/`Input` primitives `Form` imports. Worth confirming explicitly with the user if this comes up again, rather than assuming it's the established norm for same-sprint dependent tickets.
  - PR #80 opened (`feat(frontend): add core UI primitives and Form composite component`, closes #63 and #65) once both commits were independently verified. `docs/roadmap.md` rows `005-T048`–`005-T051`/`005-T055` marked `Done` with a `PR #80` note, following the established per-session convention of marking `Done` as soon as a PR opens (not waiting for merge — matches the #59/PR #79 precedent).
  - Delegated a documentation-accuracy pass to `technical-writer` (backgrounded, ran in parallel with `open-pr`): `frontend/README.md`'s Implementation Status, Tech Stack, Project Structure (Current/Target), and Atomic Design table all updated to reflect the two new directories — and fixed a real pre-existing README/code mismatch found during the pass (README said the atoms layer lives at `src/components/atoms/`; the real, roadmap-confirmed path has always been `src/components/primitives/`). Also added one WHY-comment in `Form.tsx` explaining why `event.currentTarget` must be captured into a local before the `await onSubmit(...)` call (React nulls it out once the synchronous handler returns, matching native DOM semantics) — everything else reviewed was already appropriately WHY-only per this repo's minimal-comment convention.
  - Two new `patterns-discovered.md` entries: "Issue-Body Snippets and Doc Diagrams Go Stale — Verify Against the Real Source of Truth" (formalizes the authority-ranking rule: real code > spec `tasks.md`/`data-model.md` > root `docs/*.md` > issue-body pseudocode, generalizing three independent prior instances of this exact failure mode into one reusable rule) and "Vitest Without `test.globals: true` Needs an Explicit `afterEach(cleanup)` for React Testing Library" (documents the fix already applied in `src/test/setup.ts` during the #63 implementation, so future sessions don't independently rediscover it).
- **Key findings and decisions**:
  - The "docs win over issue-body pseudocode" lesson had already surfaced independently in at least two prior sessions (backend `internal/example/` package, frontend design-tokens session) but was never promoted to `patterns-discovered.md` — only mentioned inline in `session-notes.md` each time. Fixed this session by writing the generalized rule down once, with an explicit authority ranking, so it stops being independently rediscovered per-session.
  - Confirms the "mark `Done` as soon as a PR opens, note the PR number, don't wait for merge" roadmap convention from the #59/PR #79 session — now applied a second time consistently.
  - **Noticed, not fixed**: `.github/memory/patterns-discovered.md` line ~335 (the "Suggest-Then-Approve Collaboration Workflow" entry, part of the compacted Foundation Phase block) contains literal `\n`/`—` escape-sequence text instead of real newlines/em-dashes — a pre-existing corruption from some earlier write, unrelated to this session's edits. Flagged to the user; not fixed here (out of scope, and risky to touch inside a large compacted block without dedicated attention).
- **Outcomes**:
  - `frontend/src/components/primitives/{Button,Input,Card}.tsx` and `frontend/src/components/composites/Form.tsx` are real, tested (37/37 tests, 5 suites, independently re-run and confirmed passing after every delegated step), token-driven, WCAG 2.1 AA-compliant components — the first consumers of `tokens.css`/`global.css` from the prior #59 session.
  - PR #80 open, targeting `main`, closes #63 and #65 on merge.
  - `frontend/README.md` accurately reflects current vs. planned component-layer state; the `atoms`→`primitives` naming mismatch is fixed project-wide (docs now match the real folder name).
  - **Follow-up work, not yet done**: (1) PR #80 awaits review/merge; (2) issue #64 (`LoadingSpinner`/`ErrorMessage`/`EmptyState`) is the natural next Sprint 2 primitives ticket and is what `Form.tsx`'s inline loading/error UI would eventually delegate to — worth a follow-up pass once #64 lands, though not required (nothing here structurally blocks it); (3) the `.github/memory/patterns-discovered.md` line-335 escape-sequence corruption noted above remains unfixed; (4) `postgres:15.4-alpine` staleness and `internal/example/` throwaway-cleanup items from prior backend sessions remain unresolved and unrelated to this session's scope.

