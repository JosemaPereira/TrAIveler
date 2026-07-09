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

