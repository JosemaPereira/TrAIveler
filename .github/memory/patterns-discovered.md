# Patterns Discovered

Recurring code, testing, and debugging patterns that improve reliability and speed.
Append-only knowledge base, written in English.

**Note**: Compacted 2026-07-12 (Sprint 3 closure) — merged overlapping entries (Mockery tool usage +
Mockery scope; Colima install + Colima env-var gotcha; the two "required check stuck at Expected"
causes; the DI pattern + its payment-provider example), and cut narrative/step-by-step detail and
long code examples down to the minimum needed to reuse each lesson. No decision-relevant fact was
dropped — only prose and illustrative code were trimmed. Supersedes the 2026-07-11 cleanup note.

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
