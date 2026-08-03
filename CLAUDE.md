# CLAUDE.md

Instructions for Claude Code working in this repository. This file is loaded automatically at the
start of every session.

## Project Context

- Create a web tool to plan trips to any destination in the world. The tool offers suggestions for
  places, excursions, gastronomy, tips, and recommendations for exploring the most attractive and
  hidden treasures of a city or country within a specific timeframe.
- Stack: Backend: Golang 1.24 or higher, Frontend: React 19 or higher
- Actual Phase: Initial setup of the project, including the creation of the repository, initial
  commit, and basic project structure.

## Documentation References

Read the following files before generating code, tests, or UI for this project:

### Product & Vision

- **docs/product-vision.md** — product identity, core problem, value proposition, target personas,
  MVP scope, explicit out-of-scope, and roles/permissions model.

### Requirements & Constraints

- **docs/functional-requirements.md** — normative requirements ("The system shall…"); defines what
  the application must do and what is out of scope for the MVP.
- **docs/nfrs.md** — measurable non-functional requirements with validation methods covering
  performance, scalability, availability, accessibility (WCAG 2.1 AA), security, maintainability,
  privacy, and observability.

### Architecture & Infrastructure

- **docs/architecture.md** — system component boundaries, integration rules, scalability
  constraints, security boundaries, and observability strategy. Defines how backend (Go on ECS
  Fargate), frontend (React on S3+CloudFront), database (PostgreSQL RDS), and external dependencies
  (Anthropic AI) interact.
- **docs/cloud-and-environments.md** — cloud provider (AWS us-east-1), environment topology
  (staging active, production dormant), IaC approach (Terraform), compute platform (ECS Fargate,
  not Lambda), CI/CD (GitHub Actions with OIDC), secrets management (AWS Secrets Manager), and cost
  strategy ($200 staging, $300-400 production).
- **docs/data-model.md** — complete catalog of 16 core entities with full attribute specifications,
  three-layer validation rules ([DB], [Logic], [API]), performance-critical indexes, forward-only
  state transitions, and cascade behavior.

### Security & Authorization

- **docs/security.md** — authentication model (JWT RS256 with multi-key rotation), authorization
  roles (admin/partner), subscription limits, password security (bcrypt cost 12), session
  management, optimistic locking for concurrency, secrets management (AWS Secrets Manager), PII
  handling (GDPR-aware), input validation (prompt injection prevention), output sanitization,
  dependency security, security logging & monitoring, and testing requirements.

### Development Standards

- **docs/api-design-standards.md** — API design conventions: resource naming, URL structure,
  versioning (/api/v1), HTTP methods, request/response formats (snake_case JSON, ISO 8601
  timestamps), standardized error format, HTTP status codes, pagination, filtering, sorting, and
  rate limiting.
- **docs/coding-guidelines.md** — formatting rules, import organization, naming conventions, and
  KISS/DRY principles for Go (backend) and React/TypeScript (frontend).
- **docs/testing-guidelines.md** — three-layer testing strategy (unit, integration, E2E), folder
  structure, naming conventions, and coverage targets (90% business logic, 90% shared components).
- **docs/ui-guidelines.md** — design tokens (color, spacing, typography), component layers (Atomic
  Design), responsive breakpoints, accessibility rules (WCAG 2.1 AA), and loading/error/empty state
  requirements.

## Language Policy (MANDATORY)

- Conversation with the developer may be in English or Spanish; respond in whichever language the
  developer uses.
- ALL generated artifacts MUST be in English, without exception. This includes: source code,
  identifiers, comments, docstrings, commit messages, branch names, documentation (docs/, README,
  specs), test names, and any file content committed to the repository.
- Never mix languages inside an artifact. If the developer writes a request in Spanish, still
  produce the code/docs in English.

## Development Principles

- Test-Driven Development: Red-Green-Refactor
- Incremental, small, and testable changes
- Validation before commit: tests pass, no lint errors, no formatting issues (frontend:
  `npm run format:check` before every commit — CI's Prettier gate catches unformatted files late;
  do not rely on it to be the first check)

## API Documentation Enforcement (MANDATORY)

Whenever a backend endpoint is added, **or an existing endpoint's request/response shape or
behavior changes**:

- Add/update its `swag` doc-comment annotations per `specs/009-api-documentation/contracts/api.md`'s
  shape (mirrors `internal/auth/handler.go`, the canonical reference implementation — see
  `backend/README.md`).
- Regenerate `backend/docs/` via `make swagger` and commit the regenerated artifact alongside the
  code change (it's a generated file — never hand-edit it).
- Delegate to the `technical-writer` subagent to review the annotation change as part of the same
  piece of work — a new or changed endpoint is not done until `technical-writer` has confirmed the
  annotations are complete/accurate and the regenerated docs are committed.

This is a manual-but-mandatory step until the deferred `swagger-drift` CI gate (Sprint 5, spec 009
US3) lands and enforces it automatically — do not treat "CI will catch it later" as a reason to
skip this now.

## Git Workflow

- **Commit messages**: Follow `.github/COMMIT_GUIDELINES.md` (Conventional Commits, imperative
  mood, English only)
- **Pull requests**: Use `.github/PULL_REQUEST_TEMPLATE.md`, including task checklist,
  dependencies, and verification steps
- **Branch naming**: `feature/descriptive-name` (English, lowercase, hyphens)
- **Protection**: Never commit directly to main
- **Versioning**: Semantic Versioning (SemVer) 2.0.0

## Local Agents & Workflows

Two different mechanisms are in play here, deliberately:

**SpecKit (`speckit-*`) — official integration.** This project uses GitHub's
[spec-kit](https://github.com/github/spec-kit) CLI (`specify`), with Claude Code as its sole
integration:

```bash
specify integration install claude --script sh
```

Installed integrations are tracked in `.specify/integration.json`.

This generated `.claude/skills/speckit-*/SKILL.md` (nine core spec-kit skills plus
`speckit-taskstoissues`, a project-level extension) — invoke them directly as `/speckit-specify`,
`/speckit-clarify`, `/speckit-plan`, `/speckit-tasks`, `/speckit-analyze`/`/speckit-checklist`,
`/speckit-implement`, `/speckit-converge`/`/speckit-taskstoissues`. These are maintained upstream by
spec-kit, not by us — to update them when spec-kit releases a new version, run `specify integration
upgrade claude --force` (diff-aware; review the diff before committing). Do not hand-edit files
under `.claude/skills/speckit-*/` or hand-port speckit agents/commands elsewhere — that previously
caused a naming collision with these official skills and was removed.

**Everything else (`.claude/agents/`) — this project's custom subagents, since spec-kit doesn't
cover them.** Delegate to these by name (via the Agent tool) — either naturally ("implement X with
TDD" auto-matches `tdd-developer`, since Claude delegates based on each subagent's `description`) or
explicitly ("use the `commit-and-push` subagent") — instead of re-deriving their instructions
inline. Each one is a self-contained persona, with any linked one-shot workflow merged into the
same file as a `## Task: ...` section, to avoid duplicating large persona text across files.

- **Role agents**: `code-reviewer`, `product-manager` (Plan Sprints, Create Sprint Issues),
  `tdd-developer` (Implement Feature), `technical-writer` (Update Documentation), `test-engineer`
  (Create UI Tests, Run UI Tests).
- **Standalone workflows**: `build-roadmap`, `commit-and-push`, `open-pr`, `promote-fundations`,
  `sync-issues`.

There are deliberately no `.claude/commands/` wrapper files for these — that would just be a thin
duplicate of each subagent's `name`/`description` with no functional benefit over delegating by
name, and one more file to keep in sync per agent. If a `.claude/commands/<name>.md` collides with
one of these subagent names later, remove the command, not the subagent.

## Shared Project Memory (MANDATORY)

This repository has **one** memory system, persisted under `.github/memory/`. The full protocol —
including how to avoid conflicts and stale entries across concurrent sessions — is canonical in
**`.github/memory/README.md`**. Read it once per session and follow it; do not rely on the summary
below if it ever seems to diverge, the README wins.

Claude Code's obligations:

- **Session start**: before starting substantive work, read `.github/memory/session-notes.md`,
  `.github/memory/patterns-discovered.md`, and `.github/memory/scratch/working-notes.md`, and
  briefly confirm what was loaded.
- **While working**: capture in-progress notes in `.github/memory/scratch/working-notes.md`
  (gitignored, not committed).
- **When a reusable pattern emerges**: append it to `.github/memory/patterns-discovered.md` using
  its template, tagged with today's date and `**Tool**: Claude Code`.
- **At the end of a significant session**: append a summary to `.github/memory/session-notes.md`
  using its template, tagged with today's date and `**Tool**: Claude Code`.
- **Append-only**: never edit or delete another session's entry — including historical entries
  tagged `GitHub Copilot` from before this project consolidated on Claude Code as its sole AI tool
  (issue #212).

Note: Claude Code also has its own persistent memory (`~/.claude/projects/.../memory/`) for
cross-project user/feedback context — that is separate from this repo-local, tool-agnostic memory
and should not duplicate it.

## Task Consolidation Policy

Before creating GitHub issues, consolidate atomic tasks into one issue when they share the same
spec, area/module, tech stack, and have no blocking dependencies (see
`.github/SPRINT-CONSOLIDATION-CHECKLIST.md` and the `docs/roadmap.md` Group column for details).
Keep different tech stacks, critical blockers, or cross-spec work as separate issues.
