# Patterns Discovered

Recurring code, testing, and debugging patterns that improve reliability and speed.
Append-only knowledge base, written in English.

**Note**: Compacted 2026-07-12 (Sprint 3 closure) — merged overlapping entries (Mockery tool usage +
Mockery scope; Colima install + Colima env-var gotcha; the two "required check stuck at Expected"
causes; the DI pattern + its payment-provider example), and cut narrative/step-by-step detail and
long code examples down to the minimum needed to reuse each lesson. No decision-relevant fact was
dropped — only prose and illustrative code were trimmed. Supersedes the 2026-07-11 cleanup note.
Reviewed again 2026-08-01 (Sprint 6 closure, first pass with a full pre-review archive kept at
`scratch/patterns-discovered-pre-sprint6-compaction-2026-08-01.md`): read every entry (~50) against
the Sprint 5-6 additions (mockery, testcontainers, CI-gate, auth-session clusters) looking for true
duplicates. Found none — the 2026-07-12 pass already merged the pairs that existed at the time, and
every entry added since documents a distinct, independently-triggerable lesson even where several
cluster around the same feature (e.g. the three mockery entries answer three different questions:
whether to mock, what `--all` actually generates, and how to break an import cycle from a mock). Only
action taken: light prose tightening on the wordiest entry (*A New CI Gate Must Be Test-Run Against
`main`'s Actual State*) with no fact removed. This establishes the same archive-before-touching
convention `session-notes.md` already follows for future passes.

## Index

Scan this first and jump to the entries that touch your task — the full file is long and most of it
will not apply. Grep the exact title to jump. Entries are append-only; add new ones at the end of
the file **and** add a line here.

### Go backend — code

- *Dependency Injection with Interface-First Design (Go)* — constructor-injected deps, no globals.
- *Two-Query Pagination, Not `COUNT(*) OVER()`* — the window-function version is a footgun.
- *Prefer an Official SDK's Built-In Retry Over Hand-Rolling One* — SDK-native retry/timeout wins.
- *A Config Value Can Be Loaded, Validated, and Logged, Yet Still Never Reach the Code It Configures*
  — "loaded" ≠ "forwarded"; grep the consuming signature. Bit us twice (DB pool, `BCRYPT_COST`).
- *`internal/example/` Is Throwaway — Delete on First Real Domain* — still pending (Sprint 8, Trip).
- *Goose Migrations Must Not Reference Tables From a Later, Not-Yet-Landed Migration Set*.
- *Chi Panics on a Duplicate Routed Pattern — Split Auth Groups by Full Path, Not Nested `Route()`* —
  a startup panic, not a compile error; bit the public-vs-gated `/api/v1` split.
- *A Long-Blocked Task Can Be Blocked on a Premise That Was Never True* — verify the blocker exists
  in the canonical model before waiting another sprint for it (004-T011/`itinerary_items`).

### Go backend — testing

- *Colima for testcontainers-go (Free Runtime + Required Env Vars)* — needs **both** env vars.
- *Layered Testing Strategy with Optional Testcontainers (Go)* — `testing.Short()` gating.
- *Mockery: Generation + Scope* — mocks are for **external-system interfaces only**.
- *`mockery --all` Regenerates a Mock for Every Interface in a Configured Package — Prune Internal
  Ports* — `--all` ignores the per-package `interfaces:` filter; `rm` the auto-generated internal-port
  mocks after `make mocks`.
- *Mockery Mock Self-Import → External Test Package + Interface Ports Break Cycles*.
- *revive's `context-as-argument` Conflicts with the t-First Test-Helper Convention* — docs win;
  linter carve-out for `_test.go`.
- *Simulating a Realistic DB Timeout in an Integration Test*.
- *A "These Two Responses Are Indistinguishable" Comment Needs a Test Comparing Them* — per-path
  assertions can't catch divergence between paths; share one constant.
- *A Test Can Encode a Contract the Backend Never Had* — a mock is not evidence; it is a claim,
  and an unbuilt endpoint can never contradict it.

### Frontend

- *Vitest Needs Explicit `afterEach(cleanup)` Without `test.globals: true`*.
- *Remove `baseUrl` When Only Used to Support `paths` (tsconfig)*.
- *A Client-Level Interceptor That Must Touch the Store Needs a Handler Registry, Not an Import* —
  the store already imports the client; invert with `setXHandler()` wired at the composition root.
- *MSW and a Stubbed Global `fetch` Cannot Share a Test File* — the stub bypasses the interceptor.
- *An Endpoint Where a 401 Is an Answer, Not a Failure* — session probes and credential
  endpoints need a carve-out from any global 401 teardown.

### Infra, CI & GitHub

- *Required Status Check Stuck at "Expected" — Two Root Causes* — `paths:` filters and job-name
  mismatch. Job names are pinned in the ruleset — renaming one silently breaks the gate.
- *Native Cross-Compilation to Avoid QEMU Emulation (Docker Multi-Platform)* — 8.5min → ~56s.
- *A Version Pin Can Be Deliberate Architecture, Not Drift* — e.g. `postgres:15.4-alpine` = RDS match.
- *A Hardcoded Env-Var Fallback Can Hide a CI Wiring Gap*.
- *Lighthouse CI Can't Assert Real INP in a Standard `autorun` — Use Total Blocking Time*.
- *GitHub Rulesets/Branch Protection Require Public Repo or Pro (Personal Accounts)*.
- *ECS Fargate for AI Workloads (Compute Platform Choice)*.
- *Hash-Manifest Drift Detection for Hand-Ported Config Pairs* — `scripts/check-agent-drift.py`.
- *Prefer the Official Multi-Agent Integration Over Hand-Porting* — why the SpecKit hand-port died.
- *Claude Code's Shell Snapshot Drops Single-Underscore Shell Functions*.
- *A New CI Gate Must Be Test-Run Against `main`'s Actual State, Not Just a Clean Diff* — a status
  gate can be internally correct yet ship permanently red.
- *`permissions:` on a Workflow Silently Denies `gh` Calls a Local Session Would Allow* — least-
  privilege scoping and a new `gh` subcommand are easy to add in different sessions.
- *A Whole-Document Drift Gate Fails PRs on Pre-Existing Drift They Didn't Cause* — why
  `roadmap-status-drift` was removed; scope a gate to the PR's own diff or don't automate it.

### Authority, docs & process

Read these before trusting any spec/issue text.

- *Issue-Body Snippets Are Lowest-Authority — Verify Against Real Source* — the ranking that keeps
  recurring: real code > `tasks.md` > root docs > issue body. Applies to prose hand-offs too.
- *Roadmap Row Notes Are the Only Channel Future Issue Bodies Inherit Context Through* — a decision
  recorded only in docs/PRs/closed issues will **not** reach tickets that don't exist yet.
- *A Spec's Data-Model Validation Rule Can Override Normal REST Convention — Read It Literally* —
  e.g. `/healthz` always-200.
- *Append Local-Dev-Only Doc Overrides After `PROMOTED:...END`, Don't Edit Inside It*.
- *Treat README Command Blocks as Executable Claims*.
- *What "Keep docs/ Updated" Means for an Infra Module PR* — the two-tier docs rule.
- *README Documentation Consistency*, *Mermaid Diagrams for Documentation*,
  *GitHub Documentation URL Formatting*.
- *Mermaid Straight Edges: Flowcharts Have `curve`, `erDiagram` Does Not*.
- *Targeted (Grep-First) Memory Loading in Subagent Prompts*.

### Planning & workflow

- *Consolidation-First Issue Creation* — mandatory since Sprint 2; groups of 2-4 tasks.
- *Simplified GitHub Issue Workflow (Flat Issues, No Sub-Issues)*.
- *Idempotent Roadmap Reconciliation*, *Existing Sprints as Refinement Evidence (Idempotency)*.
- *Foundation Promotion Workflow*, *Documentation-Centric SpecKit Workflow*.
- *NFR Observability Primitives Must Precede Feature Work*.

### Product & security domain

- *Suggest-Then-Approve Collaboration Workflow*.
- *Server-Side Prompt Validation Deny-List (AI Security)* — pattern-based, not LLM-based.

---

## Pattern Template

### Pattern Name
- **Discovered**: <YYYY-MM-DD> — **Tool**: <Claude Code | GitHub Copilot>
- **Context**: <where this applies>
- **Problem**: <what issue repeatedly occurs>
- **Solution**: <recommended approach, terse>
- **Related**: <file paths>

---

### Colima for testcontainers-go (Free Runtime + Required Env Vars)
- **Context**: Backend integration tests using testcontainers-go on macOS/Linux without paid Docker Desktop.
- **Problem**: Docker Desktop needs a paid license; testcontainers-go also fails under Colima with just
  `docker context` set — `DOCKER_HOST` alone breaks Ryuk's reaper container, which needs a different
  socket path as seen *inside* a container.
- **Solution**: `brew install colima`, `colima start --cpu 2 --memory 4`. Then export both
  `DOCKER_HOST=unix:///Users/<you>/.colima/default/docker.sock` (host socket, used by the Docker API
  client) and `TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock` (in-container path, used only
  by Ryuk's bind mount).
- **Related**: `backend/README.md`, `backend/internal/database/client_test.go`

---

### Layered Testing Strategy with Optional Testcontainers (Go)
- **Context**: Backend test organization/CI, any Go project using testcontainers.
- **Problem**: Testcontainers need Docker/Colima locally, adding friction to fast TDD loops; CI
  shouldn't need Docker-in-Docker either.
- **Solution**: Use `testing.Short()` to make container tests skippable. `make test` = unit-only,
  `-short`, no Docker (fast default). `make test-all` = full suite incl. testcontainers, needs Colima.
  `make test-coverage` = CI target, unit-only + `-race`.
- **Related**: `backend/Makefile`, `backend/TESTING.md`

---

### Mockery: Generation + Scope
- **Context**: Backend — any package with interfaces needing mocks for tests.
- **Problem**: Hand-written mocks drift from interface changes; conversely, reaching for mockery on
  stdlib-typed dependencies (a `*slog.Logger`, plain config) adds an unneeded abstraction layer.
- **Solution**: Generate with vektra/mockery (`with-expecter: true`, output to
  `<pkg>/mocks/<interface>_mock.go`, tagged `//go:build test`, committed to git). Scope it to real
  domain interfaces — external-system wrappers (`database.Client`) AND internal layer-boundary
  interfaces (`Repository`/`Service`) — but skip stdlib types or single-method interfaces trivially
  satisfied by a real instance; grep the package for `interface` first.
- **Example**: `mockDB := dbmocks.NewMockClient(t); mockDB.EXPECT().Ping(mock.Anything).Return(nil).Once()`
- **Related**: `backend/.mockery.yaml`, `docs/mock-standards.md`

---

### Dependency Injection with Interface-First Design (Go)
- **Context**: Backend — all internal packages needing testability/swappable implementations.
- **Problem**: Direct dependencies on concrete types block mocking and swapping implementations.
- **Solution**: Define a small interface (1-5 methods) in the consumer package, implement it with a
  private concrete type, have the factory return the interface. Example: `PaymentProvider`
  (`CreateSubscription`/`CancelSubscription`/`GetSubscription`) with a `StubProvider` for MVP, swappable
  for a real provider later with no domain-logic changes.
- **Related**: `backend/internal/database/client.go`, `backend/internal/subscription/payment/`

---

### Suggest-Then-Approve Collaboration Workflow
- **Context**: Backend `internal/suggestion/`, frontend `features/suggestions/`.
- **Problem**: Partners need to contribute to a shared itinerary without directly modifying it.
- **Solution**: Partners create `Suggestion` records (`pending`); only admin can `approve` (applies
  change) or `reject` (no change). Never hard-deleted — kept for audit history.
- **Related**: `specs/001-product-vision-scope/data-model.md` (Suggestion entity)

---

### Server-Side Prompt Validation Deny-List (AI Security)
- **Context**: Backend `internal/ai/validator/`, any endpoint forwarding user input to an LLM.
- **Problem**: Adversarial prompts can override the system prompt or extract secrets; the LLM's own
  refusal behavior isn't sufficient.
- **Solution**: `PromptValidator` loads a versioned deny-list from `backend/config/prompt-rules.yml`
  (id/description/pattern/match_type/enabled), validates before forwarding. On match: `400` + user-safe
  message + `request_id`, plus a `WARN` log with the matched rule ID (never shown to client). System
  prompt itself is a server-side secret, never logged.
- **Related**: `backend/internal/ai/validator/prompt_validator.go`, `backend/config/prompt-rules.yml`

---

### Documentation-Centric SpecKit Workflow
- **Context**: SpecKit features that produce docs/standards rather than code.
- **Problem**: Doc-only features tempt a shortcut straight to writing prose, risking ambiguous
  acceptance criteria.
- **Solution**: Run docs through the same specify → clarify → plan → tasks pipeline as code — user
  stories, research decisions, dependency-ordered tasks.
- **Related**: `specs/007-api-design-standards/`

---

### Idempotent Roadmap Reconciliation
- **Context**: `docs/roadmap.md` maintenance, triggered by `/build-roadmap`.
- **Problem**: Specs evolve; the roadmap must stay in sync without destroying human-curated fields
  (sprint, priority, issue links, notes).
- **Solution**: Reconcile, don't regenerate — diff source `tasks.md` against the roadmap keyed by a
  stable ID (`<spec>-<task-id>`). ADD new rows (default Priority=TBD/Status=Backlog); UPDATE only
  source-derived fields; REMOVE archives if it has an Issue link or Status>Backlog, else hard-deletes.
  Repeated runs converge without duplication.
- **Related**: `docs/roadmap.md`

---

### Existing Sprints as Refinement Evidence (Idempotency)
- **Context**: Sprint replanning — running `/plan-sprints` again.
- **Problem**: Re-planning a sprint whose issues already exist causes duplicate issues/URL conflicts.
- **Solution**: If a sprint's rows already have an Issue URL, treat that as evidence it's already
  planned — skip it entirely (don't touch Sprint/Priority/Group/consolidation). Apply full planning
  only to sprints with no Issue URLs yet.
- **Related**: `docs/roadmap.md`

---

### NFR Observability Primitives Must Precede Feature Work
- **Context**: Cross-cutting infra, especially when an NFR spec is authored after feature specs.
- **Problem**: Observability primitives (request-ID middleware, structured logging, `/healthz`) are
  foundational, but an NFR spec's own P2 user-story priority can bury their implementation behind
  unrelated P1 feature work.
- **Solution**: Extract the *implementation* of such primitives into the Foundational phase regardless
  of the owning user story's priority; leave only *validation tests* in the original P2 story.
- **Related**: `specs/002-nfr-system-constraints/tasks.md`

---

### Foundation Promotion Workflow
- **Context**: Run after foundation specs stabilize, or whenever one is refined.
- **Problem**: SpecKit only reads the constitution + current spec — architectural/NFR/security/cloud
  decisions living only in `specs/NNN-*/spec.md` are invisible to future planning.
- **Solution**: Run `/promote-foundations` to route durable decisions into the constitution
  (`.specify/memory/constitution.md`) and `docs/*.md` (wired into Copilot/Claude instructions), using
  idempotent `<!-- PROMOTED:name START/END -->` markers.
- **Related**: `.specify/memory/constitution.md`, `docs/*.md`

---

### ECS Fargate for AI Workloads (Compute Platform Choice)
- **Context**: AWS compute platform choice for AI-integrated services.
- **Problem**: Lambda is the default serverless choice, but multi-turn AI conversations can exceed its
  15-min timeout and 10MB payload limit; repeated AI calls benefit from persistent connections.
- **Solution**: Use ECS Fargate (unlimited execution time, no payload limit, persistent HTTP
  connections) with ARM64 Graviton2 for ~20% cost savings, no code changes. Trade-off: no scale-to-zero.
- **Related**: `docs/cloud-and-environments.md`

---

### Mermaid Diagrams for Documentation
- **Context**: `docs/*.md` diagrams.
- **Problem**: ASCII box diagrams render poorly and are hard to maintain.
- **Solution**: Use Mermaid for real multi-component/branching diagrams; keep plain arrows (→) for
  simple linear sequences. Always validate Mermaid syntax before committing.
- **Related**: `docs/architecture.md`, `docs/security.md`

---

### README Documentation Consistency
- **Context**: Area READMEs (`backend/`, `frontend/`, `e2e/`, `infra/`).
- **Problem**: Inconsistent structure across areas makes navigation hard.
- **Solution**: Standard section order: Title/tagline → breadcrumbs → Responsibility → Tech Stack →
  Project Structure → Prerequisites → Env Vars → Setup → Dev commands → Related Docs. Root README
  links all area READMEs.
- **Related**: `README.md`, area READMEs

---

### Simplified GitHub Issue Workflow (Flat Issues, No Sub-Issues)
- **Context**: Every sprint's GitHub issue structure.
- **Problem**: Native parent/sub-issue tasklists add creation-order overhead and duplicate progress
  tracking.
- **Solution**: Flat issues only; use epic labels + a `Group` field + Project views for grouping
  instead of native hierarchy.
- **Related**: `docs/roadmap.md` (Group column)

---

### Consolidation-First Issue Creation
- **Context**: Sprint planning, before creating GitHub issues.
- **Problem**: One issue per atomic task fragments the backlog (e.g. a middleware file set as 5
  separate issues/PRs).
- **Solution**: Group tasks sharing spec + area/module + tech stack + no blocking deps into one
  `Group` (e.g. `G-SPRINT2-BACKEND-MIDDLEWARE`), one issue per group with a checklist. Consolidate:
  same spec/module/stack, 2-4 tasks. Don't: mixed stacks, critical blockers, cross-spec. Mandatory
  step in PM sprint planning, not optional cleanup.
- **Related**: `.claude/agents/product-manager.md`, `.github/SPRINT-CONSOLIDATION-CHECKLIST.md`

---

### GitHub Documentation URL Formatting
- **Context**: Referencing docs/specs in GitHub issue bodies.
- **Problem**: `https://github.com/OWNER/REPO/docs/x.md` 404s — missing `/blob/main/`.
- **Solution**: Always use full blob URLs: `https://github.com/{owner}/{repo}/blob/main/{path}`;
  verify the link renders before creating the issue.

---

### Prefer the Official Multi-Agent Integration Over Hand-Porting
- **Discovered**: 2026-07-09
- **Context**: Adding a new AI coding agent to a project already using a toolkit (e.g. spec-kit) set
  up for a different one.
- **Problem**: Hand-porting an existing agent's config collides with the toolkit's own native
  integration if/when one exists.
- **Solution**: Check for first-class support first (`specify integration list` shows availability +
  multi-install-safety; `specify integration install <key> --force` adds alongside an existing one).
  Only hand-port what the official integration doesn't cover.
- **Related**: `.specify/integration.json`, `CLAUDE.md`

---

### Hash-Manifest Drift Detection for Hand-Ported Config Pairs
- **Discovered**: 2026-07-09
- **Context**: File pairs that must stay equivalent but can't be codegen'd (Copilot
  `.agent.md`/`.prompt.md` vs Claude `.claude/agents/*.md`).
- **Problem**: An edit to one side silently drifts from the other with no signal.
- **Solution**: Maintain a JSON manifest of sha256 hashes per file as of last confirmed sync; a script
  recomputes and reports unchanged / one-side-changed / both-changed (non-zero exit on drift). Script
  only detects — a human/agent ports the change and refreshes the manifest (`--update <name>`).
- **Related**: `scripts/check-agent-drift.py`, `scripts/agent-port-manifest.json`

---

### Native Cross-Compilation to Avoid QEMU Emulation (Docker Multi-Platform)
- **Context**: Multi-stage Dockerfiles targeting `linux/arm64` (ECS Graviton2) built on an amd64 CI
  runner.
- **Problem**: `docker buildx build --platform linux/arm64` with no `--platform` pin on the builder
  stage runs the *entire* build stage under QEMU emulation — succeeds, just slow (measured: 451s of
  an 8.5-min job for one `go build` step), and GHA cache doesn't help since the cost is CPU-bound.
- **Solution**: Pin the builder stage to the build machine's native platform
  (`FROM --platform=$BUILDPLATFORM ... AS builder`), cross-compile with `ARG TARGETOS`/`ARG TARGETARCH`
  (Go cross-compiles natively, no emulation, esp. with `CGO_ENABLED=0`). Only the runtime stage targets
  the real platform. Verify with a real timed `docker buildx build` + `docker run --platform <target>`
  smoke test, not just docs.
- **Related**: `backend/Dockerfile`, `.github/workflows/backend-ci.yml`

---

### Required Status Check Stuck at "Expected" — Two Root Causes
- **Context**: Repos with branch-protection/rulesets requiring named CI checks, area-scoped workflows.
- **Problem**: A required check that never reports shows "Expected — Waiting for status to be
  reported" and blocks merging **indefinitely** — a permanent hang, not a timeout. Two distinct causes
  seen in this repo:
  1. **Trigger `paths:` filter** — filtering a workflow's `pull_request:` trigger by path means a PR
     that doesn't touch that path never triggers the workflow, so the required check never runs.
  2. **Job `name:` mismatch** — branch-protection `required_status_checks` matches the exact job
     **name** string, not the workflow filename or job id; a job named differently than the
     pre-provisioned ruleset context (e.g. `Accessibility Audit (Lighthouse + axe-core)` vs. required
     `Accessibility Audit`) hangs forever even though the workflow runs fine.
- **Solution**: Never filter the *trigger* — add a `changes` job (`dorny/paths-filter@v3`) that always
  runs, and gate real jobs with `needs: changes` + `if: needs.changes.outputs.<area> == 'true'` (plus
  `|| github.event_name == 'workflow_dispatch'`); a job skipped via job-level `if:` still counts as a
  passing required check. Separately, before trusting a new workflow to back an existing required
  check, pull the live ruleset
  (`gh api repos/<o>/<r>/rulesets/<id> --jq '...required_status_checks[].context'`) and grep workflow
  `name:` fields for an exact match — fix the job's name to match the ruleset, not vice versa.
- **Related**: `.github/workflows/*.yml`

---

### Two-Query Pagination, Not `COUNT(*) OVER()`
- **Context**: Any `List` repository method returning a page + total count.
- **Problem**: `COUNT(*) OVER()` only produces a value on rows actually returned — an out-of-range page
  (`LIMIT`/`OFFSET` past the end) returns zero rows, so `total` silently reads 0 even though matching
  rows exist. A mocked-repo unit test can't catch this; only a real DB with real out-of-range SQL
  surfaces it.
- **Solution**: Run `SELECT COUNT(*)` first; only run the paginated query if `total > 0` (else
  short-circuit to empty). Add an integration test requesting a page past the end of a small seeded
  dataset.
- **Related**: `backend/internal/example/repository.go`, `docs/api-design-standards.md` §9

---

### `internal/example/` Is Throwaway — Delete on First Real Domain
- **Context**: `backend/internal/example/` (full layered reference pattern + its route mount).
- **Problem**: Exists only to demonstrate the pattern; risks lingering alongside real domains as dead
  code plus a confusing live endpoint.
- **Solution**: The first time a real domain (e.g. Trip) ships using this pattern, delete
  `internal/example/` entirely (all files, mocks, migration + down-migration, `.mockery.yaml` entries,
  route mount) in the same PR.
- **Related**: `docs/roadmap.md` rows `005-T038`-`T041`

---

### Goose Migrations Must Not Reference Tables From a Later, Not-Yet-Landed Migration Set
- **Discovered**: 2026-07-15 — **Tool**: Claude Code
- **Context**: `backend/migrations/` (single flat goose directory, shared across specs); adding a
  migration that `ALTER TABLE`s a table owned by a different, not-yet-implemented spec/issue (here,
  004-T010/T011 adding `version` to `trips`/`itinerary_items`, which are created by Spec 008's own
  not-yet-landed migration set, issue #142).
- **Problem**: `goose.Up`/`.UpTo` applies every pending migration in one directory strictly in version
  order and **stops at the first failure**. A single migration referencing a table that doesn't exist
  yet doesn't just fail itself — it blocks every migration that sorts after it from ever applying,
  including unrelated pre-existing ones. Confirmed by actually running the full suite: adding
  `005_add_version_to_trips.sql` before any `trips`-creating migration existed broke all of
  `internal/example`'s integration tests and `tests/integration/error_test.go`, because the
  pre-existing `20260710120000_create_examples_table.sql` (which sorts after 005 by goose's version
  ordering) never got applied.
- **Solution**: Before adding a migration that touches a table outside the current issue's own scope,
  grep `backend/migrations/` for a migration that actually creates that table. If none exists, defer
  the dependent migration (don't add the file yet) rather than adding it "for completeness" — the SQL
  is usually already finalized in the spec's `data-model.md`, so nothing is lost by waiting; land it
  in the same PR as (or after) the migration that creates the base table.
- **Related**: `backend/migrations/001-004_*.sql` (#138), `specs/004-security-auth-model/data-model.md`
  Migration Strategy Phase 4, `backend/tests/integration/security_migrations_test.go`, issue #142
  (G-008-MIGRATIONS, will create `trips`/`itinerary_items`)

---

### Treat README Command Blocks as Executable Claims
- **Context**: Any README with example commands (`frontend/README.md` vs `package.json`).
- **Problem**: A documented command (`npm test`) can be stale/wrong if the underlying script doesn't
  exist — readers hit `Missing script` errors, and CI may be running the same broken command.
- **Solution**: Cross-check every documented command against real `package.json` scripts/deps (or Go
  equivalent) and against any CI workflow using the same command — don't just proofread for prose
  quality.
- **Related**: `frontend/README.md`, `frontend/package.json`

---

### Remove `baseUrl` When Only Used to Support `paths` (tsconfig)
- **Context**: Any `tsconfig.json` using `paths` for import aliases.
- **Problem**: `baseUrl` is deprecated (removed in TS 7.0); `ignoreDeprecations` only suppresses the
  warning.
- **Solution**: Since TS 4.1, `paths` resolves relative to `tsconfig.json`'s own directory — delete
  `baseUrl` if it exists only to support `paths`. Verify with `type-check`/`build`.
- **Related**: `frontend/tsconfig.json`

---

### Issue-Body Snippets Are Lowest-Authority — Verify Against Real Source
- **Context**: Any ticket whose issue body has illustrative pseudocode, a "Files Created" path, or
  file placement also described by a narrative doc.
- **Problem**: Issue bodies are written once and never updated — token names, package signatures, and
  file paths in them can silently go stale (real cases: nonexistent CSS tokens, wrong component folder
  vs. `docs/ui-guidelines.md`, a `__tests__/` path violating the unit/integration test split).
- **Solution**: Rank sources by authority: (1) actual current code, (2) `specs/<NNN>/tasks.md` +
  `data-model.md`, (3) root `docs/*.md`, (4) issue-body illustrative code — lowest, treat as a sketch
  of intent only. When a narrative doc conflicts with tiers 1-2, follow the higher tier and flag the
  doc as stale (separate task, not silent).
- **Related**: `docs/testing-guidelines.md`, `docs/roadmap.md`

---

### Vitest Needs Explicit `afterEach(cleanup)` Without `test.globals: true`
- **Context**: Any Vitest + React Testing Library component test; `frontend/vitest.config.ts`.
- **Problem**: RTL auto-cleanup relies on detecting a global `afterEach` — with `test.globals` unset
  (deliberate here), nothing is detected, so DOM nodes accumulate across tests in the same file
  (invisible until a file has 2+ tests).
- **Solution**: Register cleanup once in the shared setup file:
  `afterEach(() => cleanup())` after importing from `vitest`/`@testing-library/react`.
- **Related**: `frontend/src/test/setup.ts`

---

### A Hardcoded Env-Var Fallback Can Hide a CI Wiring Gap
- **Context**: Any `import.meta.env.VITE_*` read at runtime (`frontend/src/lib/api-client.ts`).
- **Problem**: A fallback to a hardcoded URL when `VITE_API_BASE_URL` was unset masked that
  `frontend-ci.yml`'s **test** job (unlike its build job) never injected the var at all.
- **Solution**: Remove "should never really be needed" fallbacks and see what breaks; if tests were
  relying on it, add a committed `.env.<mode>` file instead, and make the missing-var case throw
  loudly.
- **Related**: `frontend/src/lib/api-client.ts`, `frontend/.env.test`

---

### Targeted (Grep-First) Memory Loading in Subagent Prompts
- **Context**: Subagent instructions that mandate loading this repo's shared memory before work.
- **Problem**: "Read both memory files in full" burns most of a subagent's context budget before it
  touches code, and gets worse every sprint as the files grow.
- **Solution**: Always read `scratch/working-notes.md` in full (small, non-growing); grep the two
  large files for task-derived keywords and read only matches; fall back to a full read only if grep
  yields nothing on a foundational/cross-cutting area. State which entries were used.
- **Related**: `.claude/agents/tdd-developer.md`

---

### Claude Code's Shell Snapshot Drops Single-Underscore Shell Functions
- **Discovered**: 2026-07-09
- **Context**: Machine-level — any dev machine running Claude Code's Bash tool with `gvm` or similar
  shell-function-based tooling.
- **Problem**: `command not found: _encode` noise on every `cd` — Claude Code's shell-snapshot
  mechanism systematically drops functions whose name starts with a single underscore (confirmed
  empirically), which broke gvm's per-directory `.go-version` auto-switch.
- **Solution**: If the auto-switch isn't actually needed, disable it (comment out gvm's `cd`-hook line
  in `~/.gvm/scripts/gvm-default`) and delete the stale cached snapshot. Worth checking for on any
  machine combining gvm (or similar) + Claude Code.
- **Related**: `~/.gvm/scripts/gvm-default`

---

### Append Local-Dev-Only Doc Overrides After `PROMOTED:...END`, Don't Edit Inside It
- **Context**: Any `docs/*.md` with `<!-- PROMOTED:name START/END -->` markers, when a change is
  genuinely local/dev-only (e.g. adding a free local Ollama AI client alongside the promoted
  Anthropic-as-provider decision).
- **Problem**: Editing inside a `PROMOTED` block would misrepresent a dev-only addition as a
  staging/production architecture change, and risks being clobbered by a future `/promote-foundations`
  run.
- **Solution**: Append a new section (e.g. "Local Development Note") *after* the `PROMOTED:...END`
  marker; the promoted content stays untouched and the addendum survives future promotion re-runs.
- **Related**: `docs/architecture.md`, `docs/cloud-and-environments.md`, `docs/local-ai-setup.md`

---

### GitHub Rulesets/Branch Protection Require Public Repo or Pro (Personal Accounts)
- **Discovered**: 2026-07-09
- **Context**: Personal (non-org) GitHub repos wanting rulesets or classic branch protection.
- **Problem**: Both APIs return `403 Upgrade to GitHub Pro or make this repository public` on a
  private personal-account repo — confirmed empirically, not just from docs.
- **Solution**: Upgrade to Pro or make the repo public before configuring rulesets/branch protection;
  re-run the same `gh api` calls once public.

---

### A Version Pin Can Be Deliberate Architecture, Not Drift
- **Discovered**: 2026-07-09
- **Context**: Any Dockerfile/compose version pin that looks stale (`alpine:3.19`,
  `postgres:15.4-alpine`).
- **Problem**: Not every old-looking pin is accidental — `alpine:3.19` was genuine drift (safely
  bumped to 3.23), but `postgres:15.4-alpine` is a deliberate decision tied to the planned RDS
  `engine_version`, documented across `docs/architecture.md`/`cloud-and-environments.md`/
  `data-model.md` and roadmap tasks `003-T022`/`005-T079`.
- **Solution**: Before bumping "for hygiene," grep `docs/*.md` + roadmap for the exact version string.
  If referenced as a deliberate decision, flag to the user for a coordinated Terraform+docs update
  rather than a silent patch. `postgres:15.4-alpine` — still unresolved as of Sprint 2 close, not yet
  re-checked for CVEs.
- **Related**: `docker-compose.yml`, `docs/architecture.md`

---

### What "Keep docs/ Updated" Means for an Infra Module PR
- **Discovered**: 2026-07-12
- **Context**: Any Terraform module PR under `infra/modules/`.
- **Problem**: The standing instruction "keep /docs up to date on every infra PR" naively reads as
  "edit `architecture.md`/`cloud-and-environments.md` every time" — but those describe agreed target
  architecture at a high level and hadn't needed edits through VPC/RDS/ALB/CloudFront, since the built
  modules already matched them.
- **Solution**: Two tiers, treated differently every infra PR: (1) **implementation-status trackers**
  (`infra/README.md`, `docs/roadmap.md` rows) — update every time, flip Backlog→Done only after merge
  with a PR reference; (2) **target-architecture docs** (`architecture.md`/`cloud-and-environments.md`,
  inside `PROMOTED:...` blocks) — only touch if something is factually wrong (check `git log -p`
  first), never just because "not yet built."
- **Related**: `infra/README.md`, `docs/roadmap.md`

---

### Lighthouse CI Can't Assert Real INP in a Standard `autorun` — Use Total Blocking Time
- **Discovered**: 2026-07-12
- **Context**: `lighthouserc.yml` / any Lighthouse CI config asserting Core Web Vitals.
- **Problem**: Asserting `interaction-to-next-paint` directly always fails (`auditRan: found 0`) under
  a standard single-navigation `lhci autorun` — that audit only supports `timespan` mode with a real
  recorded interaction, and structurally returns `notApplicable` otherwise. Confirmed by reading the
  Lighthouse audit source, not assumed.
- **Solution**: Assert `total-blocking-time` instead — Google's documented lab-mode proxy for input
  responsiveness, same 200ms "good" threshold as the NFR's INP target, so only the audit id changes.
  Before trusting any LHCI audit id, check its `supportedModes` or just run `lhci autorun` locally and
  read real output.
- **Related**: `lighthouserc.yml`, `frontend/README.md`

---

### A Spec's Data-Model Validation Rule Can Override Normal REST Convention — Read It Literally
- **Discovered**: 2026-07-13 — **Tool**: Claude Code
- **Context**: Reconciling `backend/cmd/api/server.go`'s `/healthz` handler with spec 002's
  `HealthCheckResponse` (issue #110, 002-T010/T011).
- **Problem**: The typical REST health-check convention is 503 on an unhealthy dependency; the
  existing handler followed that convention. But `specs/002-nfr-system-constraints/data-model.md`'s
  `HealthCheckResponse` validation rules explicitly state the HTTP status is *always* `200 OK`, even
  when the body's `status` field is `"degraded"` — because load balancers key routing off the HTTP
  code, and a transient DB blip shouldn't trigger ECS/ALB to cycle the task. Assuming "unhealthy → 5xx"
  from convention alone would have shipped a spec violation that also looked correct to a casual
  reviewer.
- **Solution**: When a spec's data-model doc states an explicit validation rule for an
  entity/response, treat it as literal and higher-authority than general REST/HTTP convention — grep
  the entity's own "Validation rules" bullets before assuming standard behavior. Here it meant moving
  the failure signal out of the HTTP status entirely and into a body field, with a `logger.Warn` added
  so the failure isn't silently lost to operators now that the HTTP status can't carry it.
- **Related**: `backend/cmd/api/server.go`, `specs/002-nfr-system-constraints/data-model.md`

---

### Prefer an Official SDK's Built-In Retry Over Hand-Rolling One
- **Discovered**: 2026-07-13 — **Tool**: Claude Code
- **Context**: Wrapping any well-maintained official client SDK (e.g. `github.com/anthropics/
  anthropic-sdk-go`) in this codebase's own client-wrapper pattern (issue #117, `internal/ai/anthropic.go`).
- **Problem**: The existing precedent in this codebase, `OllamaClient` (`ollama_client.go`), hand-rolls
  its own linear-backoff retry loop — reasonable there since it talks to Ollama's bare HTTP API with
  no SDK. Copying that same hand-rolled loop for `AnthropicClient` would have meant reimplementing
  (and re-testing) exponential backoff and Retry-After-aware retry logic the official SDK already
  ships and maintains.
- **Solution**: Before hand-rolling retry/backoff/timeout around any newly-added official SDK, check
  whether the SDK exposes idiomatic construction-time options for it first (here,
  `option.WithMaxRetries`/`option.WithRequestTimeout` on `anthropic.NewClient`, confirmed by reading
  the SDK's own `internal/requestconfig` source — it already does exponential backoff and honors a
  response's `Retry-After` header). Only hand-roll when wrapping a bare HTTP API with no SDK (still
  correct for `OllamaClient`) or when the SDK's built-in behavior doesn't match the required contract.
- **Related**: `backend/internal/ai/anthropic.go`, `backend/internal/ai/ollama_client.go`

---

### Simulating a Realistic DB Timeout in an Integration Test
- **Discovered**: 2026-07-13 — **Tool**: Claude Code
- **Context**: Black-box integration tests (real compiled binary + real Postgres testcontainer,
  `backend/tests/integration/`) needing to exercise a genuine DB-timeout failure path (issue #119,
  `error_test.go`), where the issue's own Review Focus explicitly required "a realistic simulated
  failure ... not an artificial short-circuit that wouldn't occur in production."
- **Problem**: Mocking the repository or short-circuiting with a canceled `context.Context` would
  prove the Go-level error-wrapping logic works, but not that a *real* Postgres timeout produces the
  same code path — and a query against a healthy, idle test database never naturally times out on
  its own.
- **Solution**: (1) Append `&statement_timeout=<ms>` to the app's own `DATABASE_URL` connection
  string — pgx's `ParseConfig` forwards unrecognized connection-string keys into the connection's
  startup `RuntimeParams`, so Postgres applies it like `SET statement_timeout = ...` to every pooled
  connection the app opens, not just the first. (2) From a *second*, independent connection, open a
  transaction and run `LOCK TABLE <table> IN ACCESS EXCLUSIVE MODE` without committing/rolling back —
  Postgres's most restrictive lock, conflicting with even a plain `SELECT`. (3) Issue the request
  under test while that lock is held: `statement_timeout` counts time spent waiting on a lock, so
  Postgres itself cancels the app's blocked query (SQLSTATE `57014`) once the timeout elapses — a
  genuine driver error, not a fabricated one. Release the lock (rollback) once the assertions are
  done. Keep the timeout short (hundreds of ms) so the test stays fast, and use a bounded
  client-side HTTP timeout on the request so a wiring mistake fails fast instead of hanging.
- **Related**: `backend/tests/integration/error_test.go`, `backend/internal/example/repository.go`

---

### A Config Value Can Be Loaded, Validated, and Logged, Yet Still Never Reach the Code It Configures
- **Discovered**: 2026-07-15 — **Tool**: Claude Code
- **Context**: `backend/internal/database/client.go`'s `NewClient`, called from `cmd/api/main.go`
  (issue #141, G-008-DATABASE).
- **Problem**: `config/config.go` loaded and validated `DB_MIN_CONNECTIONS`/`DB_MAX_CONNECTIONS`
  (with its own min-cannot-exceed-max check), and `main.go` logged
  `cfg.Database.MinConnections`/`MaxConnections` right after connecting — every visible signal said
  the values were live. But `NewClient` only accepted the connection URL and hardcoded
  `MinConns=5`/`MaxConns=25` internally, so the env vars had zero effect and the log line actively
  lied about the real pool size. Every existing test passed regardless, since none exercised a
  non-default value. Only surfaced by cross-checking `backend/README.md`'s claim ("pool sizing is
  driven by `DB_MAX_CONNECTIONS`/`DB_MIN_CONNECTIONS`") against `NewClient`'s actual parameter list.
- **Solution**: "Loaded and validated" is not "forwarded" — when a doc says a config value drives
  behavior, grep the consuming function's real signature, not just where the env var is parsed.
  Fixed by adding `minConns, maxConns int32` params to `NewClient` and updating the one real caller
  plus every test call site; added a test asserting the pool's actual `Pool().Config()` reflects
  non-default values (2/10), so the assertion can't pass by coincidence against the old hardcoded
  5/25.
- **Related**: `backend/internal/database/client.go`, `backend/cmd/api/main.go`, `backend/config/config.go`

---

### Roadmap Row Notes Are the Only Channel Future Issue Bodies Inherit Context Through
- **Discovered**: 2026-07-15 — **Tool**: Claude Code
- **Context**: Mid-sprint reconciliation after implementation decisions invalidate spec/tasks.md
  literal text (flat `internal/auth` vs `internal/auth/password/`, dead `pkg/`, no Axios,
  pre-existing `users` table, `LogSecurityEvent` superset).
- **Problem**: A decision documented only in docs/ addenda, closed-PR descriptions, or comments on
  already-created issues does NOT reach tickets that don't exist yet — future issue bodies are
  drafted from `docs/roadmap.md` rows, so a stale row Description (e.g. "calls axios",
  "pkg/secrets/manager.go", "create users table") reproduces the drift in every future sprint's
  issue batch, forcing the same reconciliation again.
- **Solution**: Propagate on three surfaces, each with a different audience: (1) docs/ —
  post-`PROMOTED:...END` addenda for humans/agents reading target design; (2) `docs/roadmap.md` —
  a dated callout in the sprint-plan block PLUS per-row Notes on every affected future row (this is
  the load-bearing one for future issues); (3) already-created open issues — a dated "Scope
  reconciliation (YYYY-MM-DD)" section appended to the body (original text preserved), stating only
  code-verified facts. Never rely on surface (1) or (3) alone to reach not-yet-created tickets.
- **Related**: `docs/roadmap.md` (Sprint 5 "Implementation decisions locked in mid-sprint"
  callout + 008-T059/T064/T150/T160/T192/T193/T206, 004-T065/T070, 001-T033, 002-T023, 008-T040
  row Notes), issues #142–#150

---

### revive's `context-as-argument` Conflicts with the t-First Test-Helper Convention
- **Discovered**: 2026-07-16 — **Tool**: Claude Code
- **Context**: Any Go test helper taking both `*testing.T` and `context.Context`, with revive's
  `context-as-argument` rule enabled (it is, repo-wide, in `backend/.golangci.yml`).
- **Problem**: `docs/coding-guidelines.md` ("Function Parameters" / "Helper Parameters") mandates
  `t *testing.T` as the FIRST parameter, before ctx (`setupTestDB(t *testing.T, ctx context.Context)`),
  but revive's `context-as-argument` demands ctx first — so a guideline-compliant helper fails
  `make lint`. Historically helpers were written `(ctx, t)` to appease the linter, silently violating
  the documented convention.
- **Solution**: Keep the documented convention (t first) and carve out the linter instead: a
  `.golangci.yml` `issues.exclude-rules` entry with `path: _test\.go`, `text: "context-as-argument"`,
  `linters: [revive]`. Production code keeps ctx-first enforcement untouched.
- **Related**: `backend/.golangci.yml`, `docs/coding-guidelines.md`,
  `backend/internal/database/client_test.go`, `backend/tests/integration/swagger_test.go`

---

### Mockery Mock Self-Import → External Test Package + Interface Ports Break Cycles
- **Discovered**: 2026-07-16 — **Tool**: Claude Code
- **Context**: Backend — mocking an interface whose method signatures reference types from the
  interface's own package (e.g. `middleware.TokenValidator` returning `middleware.AuthClaims`).
- **Problem**: (1) The generated mock in `<pkg>/mocks/` imports its source `<pkg>` for those types,
  so an *internal* `package <pkg>` test that uses the mock forms a cycle (`pkg`→`mocks`→`pkg`).
  (2) Separately, a middleware wanting to consume `auth/jwt` couldn't: `jwt`→`errors`→`middleware`
  already, so `middleware`→`jwt` is a cycle.
- **Solution**: (1) Put mock-using tests in the *external* `package <pkg>_test` (same dir is fine
  alongside internal `package <pkg>` tests) — mirrors `internal/example/service_test.go`. (2) Define
  a local port: a minimal interface + plain struct (`TokenValidator`/`AuthClaims`) in the consumer
  package, and write the concrete-type→port adapter at the composition root (`package main`), not in
  the library. Keeps the library decoupled and testable without the heavy dependency.
- **Related**: `backend/internal/middleware/auth.go`, `backend/internal/middleware/auth_test.go`,
  `backend/internal/middleware/mocks/token_validator_mock.go`, `docs/mock-standards.md`

---

## `mockery --all` Regenerates a Mock for Every Interface in a Configured Package — Prune Internal Ports

**Date**: 2026-07-23 · **Tool**: Claude Code

- **Context**: `backend/Makefile`'s `make mocks` runs `mockery --config .mockery.yaml --all`. The
  `--all` flag makes mockery generate a mock for **every** interface found in each package listed
  under `packages:` — the per-package `interfaces:` list is *not* a filter when `--all` is set. So
  adding any new interface to an already-configured package (e.g. `internal/auth`) silently produces
  a new `<interface>_mock.go` on the next `make mocks`, including for small consumer-owned ports you
  intended to hand-fake.
- **Symptom**: After adding `auth.TokenIssuer` (1-method port) and with the pre-existing unexported
  `auth.subscriptionCreator` port present, `make mocks` emitted `token_issuer_mock.go` and
  `subscription_creator_mock.go` that were never meant to be committed (the tests hand-fake those, per
  `docs/mock-standards.md`).
- **Rule**: Keep generated mocks only for the seams you actually mock — external-system interfaces
  (DB/AI/payment repos) and the handler↔service seam (mirroring `example.Service`). After `make
  mocks`, **`rm` the auto-generated mocks for internal single-method ports** and hand-fake them. There
  is no mock-drift CI gate (unlike `swagger-drift`), so pruning is safe; `git status internal/*/mocks`
  is the check. Watch for `AD`/`MM` index states from repeated `make mocks` runs — `git add -A` before
  committing to reconcile.
- **Related**: `backend/.mockery.yaml`, `backend/internal/auth/mocks/`, *Mockery: Generation + Scope*
  (above), `docs/mock-standards.md`.

---

### Chi Panics on a Duplicate Routed Pattern — Split Auth Groups by Full Path, Not Nested `Route()`

- **Discovered**: 2026-07-23 — **Tool**: Claude Code
- **Context**: `backend/cmd/api/routes.go` (008-T208, issue #179) — splitting one `/api/v1` group into
  a public group (register/login/refresh) and an `Authenticate`-gated group (examples, logout), with
  `/swagger/*` gated too.
- **Problem**: The obvious shape — two sibling `r.Group(...)`, each calling `r.Route("/auth", ...)` for
  its own endpoints — **panics at startup**. `chi.Mux.Route()` delegates to `Mount()`, which panics
  when the same pattern is already registered on that routing tree; `Group()` shares the parent's
  tree, so two `/auth` subtrees collide. The same applies to calling `r.Route("/api/v1", ...)` twice
  at the top level. This is a startup panic, not a compile or lint error, so it only surfaces when
  something actually constructs the router.
- **Solution**: When one URL prefix has to span two different middleware groups, register **full
  paths** (`r.Post("/auth/login", ...)`) instead of a nested `Route()` subtree — distinct patterns
  don't collide. Here that meant splitting `auth.Handler.RegisterRoutes` into `RegisterPublicRoutes` /
  `RegisterProtectedRoutes`. Keep one `Route("/api/v1", ...)` holding two nested `Group`s. Assert it
  with a real `assert.NotPanics` test that builds the server — reasoning about it is not enough, and a
  route-table mistake otherwise reaches production as a crash on boot.
- **Related**: `backend/cmd/api/routes.go`, `backend/cmd/api/routes_test.go`,
  `backend/internal/auth/handler.go`

---

### A "These Two Responses Are Indistinguishable" Comment Needs a Test Comparing Them

- **Discovered**: 2026-07-23 — **Tool**: Claude Code
- **Context**: `POST /api/v1/auth/refresh` (008-T148) — a missing refresh cookie is rejected by the
  handler, while an invalid/expired/revoked one is rejected deeper, by `jwt.Refresher`.
- **Problem**: Both paths returned `401 authentication_required`, and a code comment asserted they were
  indistinguishable — but they were built from **two separate string literals** in two packages, so the
  `message` fields differed (`"Session expired. Please log in again."` vs `"invalid or expired refresh
  token"`). Every existing test asserted only the status code and the `error` code, so all of them
  passed while the anti-enumeration property the uniform envelope exists to provide was quietly broken.
  Found by review, not by the suite.
- **Solution**: Any time two code paths must produce an identical response, (1) give them **one shared
  constant** rather than two literals — export it across the package boundary if needed (precedent:
  `jwt.HashRefreshToken` is exported so logout hashes the same way), and (2) write a test that issues
  both requests and asserts the **bodies are equal**, not just that each matches an expected code. A
  per-path assertion can never catch divergence between paths.
- **Related**: `backend/internal/auth/jwt/refresher.go` (`RefreshFailureMessage`),
  `backend/internal/auth/handler.go`, `backend/internal/auth/handler_test.go`

---

### A Long-Blocked Task Can Be Blocked on a Premise That Was Never True

- **Discovered**: 2026-07-30 — **Tool**: Claude Code
- **Context**: 004-T011 (issue #138/#170) sat Backlog across two sprints "blocked on #142 landing
  the `itinerary_items` table."
- **Problem**: The blocker was never going to resolve. `itinerary_items` is not an entity in
  `docs/data-model.md` at all — it appears only in spec 004's own `research.md`/`data-model.md` and
  in T011/T062. The canonical itinerary is Trip → Day → **Activity**, and Activity is what carries
  the optimistic-locking counter (Invariant 7). Each sprint closure re-verified that the table was
  still missing and re-scheduled the wait, instead of asking whether the table should exist.
- **Solution**: When a task has been deferred more than once on "waiting for X", grep the canonical
  root docs for X itself before re-deferring. If X is absent there, the dependency is fictional —
  re-map the task onto the real entity and close it out. Building the fictional table would have
  fabricated schema no document specifies and split the itinerary across two entities.
- **Related**: `backend/migrations/015_create_activities_table.sql`, `docs/roadmap.md` row 004-T011,
  `specs/004-security-auth-model/tasks.md` ("Naming correction" note)

---

### Mermaid Straight Edges: Flowcharts Have `curve`, `erDiagram` Does Not

- **Discovered**: 2026-07-30 — **Tool**: Claude Code
- **Context**: `docs/*.md` — 8 `graph TD/TB/LR` flowcharts plus 2 `erDiagram`s; request to render
  connector lines straight/orthogonal instead of curved for readability.
- **Problem**: Mermaid curves edges by default (`basis`). The fix is per-diagram-type, and applying
  the wrong directive silently does nothing.
- **Solution**: For **flowcharts**, prepend `%%{init: {'flowchart': {'curve': 'linear'}}}%%` as the
  first line of the block (`step`/`stepAfter`/`stepBefore` give hard right angles instead). For
  **`erDiagram` there is no `curve` option at all** — its documented config is only sizing/spacing/
  color; orthogonal ER edges need the ELK layout engine, an optional plugin GitHub's renderer does
  not load, so ER lines stay curved. Verify the directive actually took effect by rendering and
  diffing the SVG path data: `curve: linear` yields only `L` commands, the default yields `C`
  (cubic bezier). Validate every block by rendering it — `@mermaid-js/mermaid-cli` with
  `PUPPETEER_SKIP_DOWNLOAD=true` + `PUPPETEER_EXECUTABLE_PATH` pointed at the Chromium that
  Playwright already installed for E2E avoids a second browser download.
- **Related**: `docs/architecture.md`, `docs/cloud-and-environments.md`, `docs/security.md`,
  `docs/testing-guidelines.md`, `docs/ui-guidelines.md`, `docs/data-model.md`

---

### A Client-Level Interceptor That Must Touch the Store Needs a Handler Registry, Not an Import

- **Discovered**: 2026-07-31 — **Tool**: Claude Code
- **Context**: `frontend/src/lib/api-client.ts` gaining the 401→refresh→retry interceptor (008-T160,
  issue #180), whose failure path has to clear the Zustand auth store and redirect to `/login`.
- **Problem**: The obvious implementation — `api-client.ts` importing `useAuthStore` — is a cycle:
  `stores/auth-store.ts` already imports `api-client` for its own `refreshSession()`. ESM tolerates
  such cycles only while every use is deferred inside a function, so it "works" until someone adds a
  module-scope read and it breaks with an undefined binding at import time. The same module also has
  no router instance to navigate with (it is plain module code, not a component).
- **Solution**: Invert the dependency at the *client*: export
  `setSessionExpiredHandler(handler | null)` with a `null` default (a no-op, so the client stays
  usable and testable standalone), and put the real behavior in a feature module
  (`features/auth/session-expiry.ts`) that the composition root (`main.tsx`) installs before first
  render. Navigate with `window.location.assign`, not the router — no router instance is reachable,
  and a full document load also guarantees no stale authenticated state survives in any tree, store,
  or in-flight query. Bonus: a later ticket can swap in a router-aware handler without reopening the
  client. Tests set/unset the handler in `afterEach`, and one test drives a real 401 through
  `apiFetch` end-to-end rather than adding a test-only export to production code.
- **Related**: `frontend/src/lib/api-client.ts`, `frontend/src/features/auth/session-expiry.ts`,
  `frontend/src/main.tsx`, `frontend/src/stores/auth-store.ts`

---

### MSW and a Stubbed Global `fetch` Cannot Share a Test File

- **Discovered**: 2026-07-31 — **Tool**: Claude Code
- **Context**: Wiring MSW for the first time (`src/test/msw/`, lifecycle in `src/test/setup.ts`) — the
  switch `docs/testing-guidelines.md` §Layer 2 said to make "once a real feature lands that calls the
  backend" — alongside the pre-existing `lib/api-client.test.ts`, which stubs global `fetch`.
- **Problem**: `vi.stubGlobal('fetch', …)` replaces the very function MSW's Node interceptor patches,
  so an MSW handler in the same file is silently never consulted. The failure mode is not an error:
  the stub answers, the handler's call counter stays 0, and an assertion about a multi-request
  sequence quietly tests nothing.
- **Solution**: Split by style, not by subject. Keep the client's own unit tests on direct `fetch`
  stubs and put anything exercising a real request *sequence* (401→refresh→retry) in a separate
  MSW-backed file. Run the server with `onUnhandledRequest: 'error'` so a forgotten handler fails
  loudly instead of escaping to the network, `server.resetHandlers()` in the shared `afterEach`
  (alongside the existing RTL `cleanup()`), and deliberately register **no** default handler for
  endpoints that should only ever be reached indirectly — a stray call then fails the test instead of
  being answered. Derive handler URLs from the same `VITE_API_BASE_URL` the client reads, so handler
  and client paths cannot drift.
- **Related**: `frontend/src/test/msw/server.ts`, `frontend/src/test/msw/handlers.ts`,
  `frontend/src/test/setup.ts`, `frontend/src/lib/api-client.refresh.test.ts`,
  `frontend/src/lib/api-client.test.ts`, `docs/testing-guidelines.md`

---

### An Endpoint Where a 401 Is an Answer, Not a Failure

- **Discovered**: 2026-07-31 — **Tool**: Claude Code
- **Context**: `frontend/src/lib/api-client.ts` after 008-T160 gave it a global rule — any 401 that a
  token refresh cannot rescue tears down the session and redirects to `/login`. Then `GET /auth/me`
  (001-T021) landed as the session-bootstrap probe.
- **Problem**: A global "401 means the session died" rule is wrong for endpoints whose whole purpose
  is to *ask* about authentication. `/auth/me` answers 401 for "you are not signed in" — the normal
  negative result — so calling it on app boot from a public page would have bounced every anonymous
  visitor to `/login` via the app's own startup sequence. The same already applied to `/auth/login`
  and `/auth/register`, where 401 means "wrong credentials" and the user is looking at the page the
  redirect targets. None of this shows up in tests that exercise one endpoint at a time.
- **Solution**: Keep an explicit allowlist of endpoints where a 401 is data rather than failure
  (`EXPECTED_401_ENDPOINTS` — the name matters: an earlier `CREDENTIAL_ENDPOINTS` stopped describing
  the set once the probe joined it). Two things to get right: (1) the carve-out must **not** disable
  silent recovery — a `token_expired` 401 on the probe should still refresh and retry, and only a
  *failed* refresh escalates, because at that point the caller demonstrably held a token we issued;
  (2) write the exemption's boundary as a test, since the difference between "no session" and
  "expired session" on the very same URL is invisible in the code path otherwise.
- **Related**: `frontend/src/lib/api-client.ts`, `frontend/src/lib/api-client.refresh.test.ts`,
  `frontend/src/stores/auth-store.ts`, `backend/internal/auth/handler.go` (`handleCurrentUser`)

---

### A Test Can Encode a Contract the Backend Never Had

- **Discovered**: 2026-07-31 — **Tool**: Claude Code
- **Context**: `frontend/src/stores/auth-store.ts`'s `refreshSession()` called `GET /auth/me` from
  Sprint 5 onward. The endpoint did not exist until 001-T021, two sprints later.
- **Problem**: The action was written as `api.get<User>('/auth/me')` — a bare user — while the real
  endpoint, once built, answered `{user}` like register and login. Its unit test mocked `api.get` to
  resolve a bare user and passed for two sprints. It was not testing the backend; it was asserting
  the frontend agreed with itself. **An endpoint that does not exist can never contradict a mock**, so
  the wrong shape was frozen in by the only artifact that looked like evidence.
- **Solution**: When client code calls an endpoint that is not built yet, treat the response shape as
  an **assumption to be re-verified at wiring time**, not a settled contract — record it in the
  roadmap row's Notes, and re-derive it from the Go struct's `json` tags the moment the endpoint
  lands. The failure is silent in the good direction (correcting the code turns the test red
  immediately, which is how this surfaced), so the discipline is simply to *look*, not to add
  machinery. Related smell: an integration test that "passes" without exercising what its name claims
  — the same session had `TestAuthMe_AfterLogout_Returns401` quietly re-testing the anonymous case,
  because its `http.Client` had no cookie jar and sent nothing.
- **Related**: `frontend/src/stores/auth-store.ts`, `frontend/src/stores/auth-store.test.ts`,
  `backend/internal/auth/models.go`, `backend/tests/integration/auth_me_test.go`

---

### A New CI Gate Must Be Test-Run Against `main`'s Actual State, Not Just a Clean Diff

- **Discovered**: 2026-07-31 — **Tool**: Claude Code
- **Context**: Building `scripts/check-roadmap-status-drift.py` (002-T049, issue #173), a gate
  comparing every `docs/roadmap.md` row's Status against its linked GitHub issue's live state.
- **Problem**: The script's unit tests were all green and its own PR's diff was clean, but running it
  for real against the live repo (not a scratch fixture) found an immediate failure: 009-T019 is
  deliberately left `Backlog` with its parent issue #172 closed — a documented, permanent platform
  blocker (see the ruleset-403 pattern below), not oversight. A gate that is internally correct can
  still be **wrong about what "correct" means for the data it will actually run against** — shipping
  it as-is would have made a brand-new required check red on `main` from the moment it merged, for a
  row nobody was ever going to "fix".
- **Solution**: Before trusting a new drift/lint/status gate as done, run it for real against the
  current `main`/live state it will actually gate, not just fixture data or the PR's own diff. Here
  that meant adding a second exemption class beyond `Superseded`: rows whose Notes cell starts with
  `**Blocked**` (this repo's existing convention for "attempted, genuinely can't proceed") are also
  skipped, mirroring how 002-T050 will hit the identical wall on its own row. Caught by the parent
  session's independent re-verification pass — the implementing subagent's own local run used its
  personal `gh` auth and never surfaced the token-permission half of this (see the next entry).
- **Second failure, same class, only visible on the real PR**: even after the `**Blocked**` fix, the
  *first real CI run on the PR that introduced the gate* still failed — it flagged its own
  002-T047–T049 rows as `Status=Done` while issue #173 was still `OPEN` (issues only close at merge,
  but this repo's convention is to flip a row to `Done` in the *same* PR that closes its issue). The
  symmetric "OPEN-but-Done" rule — added for parity with "CLOSED-but-not-Done", not because the
  ticket asked for it — would therefore fail on *every* future closing PR. A local run against `main`
  can't surface this, since `main` never contains a `Done` row for a still-open issue, only a live PR
  does. **Lesson: a self-referential gate (rules the adding PR itself can trip) must be evaluated
  against that PR's own effect, not just pre-existing history.** Fix: drop the OPEN-but-Done rule —
  the ticket's literal ask ("cross-checking Status against *closed* GitHub issues") only ever required
  the CLOSED-but-not-Done direction.
- **Related**: *GitHub Rulesets/Branch Protection Require Public Repo or Pro (Personal Accounts)*,
  `scripts/check-roadmap-status-drift.py`, `scripts/check-agent-drift.py`

---

### `permissions:` on a Workflow Silently Denies `gh` Calls a Local Session Would Allow

- **Discovered**: 2026-07-31 — **Tool**: Claude Code
- **Context**: Same `roadmap-status-drift` CI step (002-T049) calling `gh issue list` inside
  `backend-ci.yml`'s `lint-test` job.
- **Problem**: `backend-ci.yml` sets an explicit least-privilege `permissions:` block
  (`contents: read`, `pull-requests: read` — added for `dorny/paths-filter`). Any scope not listed
  defaults to no access for the job's `GITHUB_TOKEN`, so `gh issue list` would 403 in real Actions
  runs even though it worked perfectly when the implementing subagent tested it locally — a local
  `gh` session authenticates with the developer's own PAT, which has full repo access regardless of
  what the workflow YAML grants. Passing tests plus a clean local script run gave false confidence;
  nothing before actually merging would have caught the CI-only failure.
- **Solution**: Every time a new step adds a `gh`/`GITHUB_TOKEN`-authenticated API call to a workflow
  that already has an explicit `permissions:` block, re-check that block for the specific scope the
  new call needs (here, `issues: read`) — don't assume "it ran locally" proves the CI token can do
  it. Workflows with no `permissions:` block at all don't have this failure mode (the default token
  is broad), which is exactly why it's easy to miss on a repo where some workflows are scoped and
  others aren't.
- **Related**: `.github/workflows/backend-ci.yml`, *Least-Privilege Token* comment above its
  `permissions:` block, `scripts/check-roadmap-status-drift.py` (this step was itself removed
  2026-08-02 — see *A Whole-Document Drift Gate Fails PRs on Pre-Existing Drift They Didn't Cause*
  below)

---

### A Whole-Document Drift Gate Fails PRs on Pre-Existing Drift They Didn't Cause

- **Discovered**: 2026-08-02 — **Tool**: Claude Code
- **Context**: `roadmap-status-drift` CI step (002-T049, added Sprint 6) — scans every row of
  `docs/roadmap.md` against live GitHub issue state on every PR, regardless of what that PR touches.
- **Problem**: PR #246 (issues #235/001-T035-T036, an unrelated backend service change) failed this
  required check on `main`'s pre-existing drift from a *different* merged PR (#244, issue #238,
  rows 001-T041–T045 never flipped Backlog→Done). A whole-repo/whole-document consistency gate — as
  opposed to a gate scoped to files the PR itself changed (`swagger-drift`, which only compares
  `backend/docs/` against what the PR's own `swag` annotations would regenerate) — has no way to
  distinguish "this PR introduced drift" from "drift already existed on `main` and nobody's PR
  happened to touch it yet." Every future PR inherits every not-yet-fixed row as a blocker, so the
  gate's failure rate tracks the total backlog of unfixed drift, not the quality of the PR under
  review — the opposite of what a required check should do. Removed entirely at the user's explicit
  request rather than patched, since scoping it to "only rows touched by this diff" would have
  required either a Notes-column diff (fragile) or per-row git blame (a much bigger rebuild than the
  gate's original scope).
- **Solution**: Before adding a required CI check that validates a document/artifact's *entire*
  current state rather than just the diff a PR introduces, consider whether drift can accumulate on
  `main` between runs of whatever manually fixes it (here: sprint closure). If so, either scope the
  check to the PR's own changed lines, or accept that periodic manual reconciliation (this repo's
  `sprint-closure-status-drift-check`, `.github/memory/session-notes.md`) is the right mechanism and
  skip automating it as a merge-blocking gate at all — automating it only pays off if the underlying
  artifact can't drift *between* check runs, which a hand-maintained Markdown table full of narrative
  Notes columns generally can.
- **Related**: `.github/workflows/backend-ci.yml`, `docs/roadmap.md` row 002-T049,
  `sprint-closure-status-drift-check` (personal memory)
