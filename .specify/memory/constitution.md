<!--
SYNC IMPACT REPORT
==================
Version change: 1.1.0 → 1.2.0
Modified principles: none
Added sections: Team Structure (new section before Team Collaboration)
Modified sections: Team Collaboration → Pull Requests (review routing by role added)
Removed sections: none
Templates requiring updates:
  ✅ .specify/templates/plan-template.md — Constitution Check section is generic; no update required
  ✅ .specify/templates/spec-template.md — no impact; specs ownership already implicit in PM role
  ✅ .specify/templates/tasks-template.md — no new task categories required
Follow-up TODOs: none
-->

# TrAIveler Constitution

## Core Principles

### I. Test-First Development (NON-NEGOTIABLE)

TDD is mandatory across the entire codebase. The Red-Green-Refactor cycle MUST be followed without
exception: write a failing test, make it pass with the minimum code required, then refactor.

- Tests MUST be written before implementation code for every new unit of behaviour.
- A failing test MUST NOT be committed unless the branch is intentionally in the Red phase and that
  fact is documented in the PR description.
- Three testing layers are required: unit, integration, and E2E (see `docs/testing-guidelines.md`).
- Coverage floors: 80% of business logic (backend), 80% of shared components and hooks (frontend),
  all public API endpoints (integration), all primary user flows (E2E).
- Tests MUST be deterministic: no real clocks, random values, or external network calls in unit or
  integration tests.
- Test names MUST be in English and MUST describe expected behaviour, not implementation details.

### II. Simplicity — KISS & DRY

The simplest correct solution MUST always be preferred. Complexity must be explicitly justified.

- Speculative abstractions are forbidden: do not build for requirements that do not yet exist (YAGNI).
- Shared logic MUST only be extracted into a reusable function, hook, or package when the identical
  pattern appears in three or more places. Two similar-looking pieces of code do not warrant
  abstraction.
- Shared constants (strings, numbers, config keys) MUST be declared once and referenced everywhere.
- A function or component that cannot be described in one sentence MUST be simplified or split.

### III. Code Quality & Consistency

All code MUST pass automated quality gates before a pull request can be merged. No exceptions.

- **Go**: code MUST be formatted with `gofmt`; `golangci-lint` MUST report zero errors
  (`errcheck`, `govet`, `staticcheck`, `revive`, `gosec` are required linters). Errors MUST never be
  silently ignored; wrap them with `fmt.Errorf("context: %w", err)`. Do not use `log.Fatal` or
  `os.Exit` outside of `main`.
- **React/TypeScript**: code MUST be formatted with Prettier; ESLint MUST report zero errors.
  TypeScript strict mode MUST be enabled. `any` is forbidden; use explicit types or `unknown` with
  type guards.
- Import order conventions defined in `docs/coding-guidelines.md` MUST be followed in both Go and
  TypeScript files.
- Dead code and commented-out code MUST be removed before merging.
- All identifiers, comments, documentation, commit messages, and file names MUST be in English.

### IV. Accessible & Token-Driven UI

Every user-facing component MUST be accessible and MUST use the design system tokens. Visual
consistency and inclusivity are non-negotiable.

- WCAG 2.1 AA compliance is required: 4.5:1 minimum contrast ratio, keyboard operability, visible
  focus indicators, meaningful `alt` attributes, associated labels for all form fields.
- Components MUST reference CSS custom property tokens from `src/styles/tokens.css`. Hard-coded
  colour, spacing, or typography values are forbidden.
- The Atomic Design layering (Primitives → Composites → Features) MUST be respected.
- Every data-dependent component MUST explicitly handle Loading, Error, and Empty states.
- Design MUST be mobile-first; touch targets MUST be at least 44 × 44 px on mobile.

### V. Secure Configuration

Secrets and environment-specific values MUST be read from environment variables. Hard-coding
credentials, API keys, or environment-specific URLs anywhere in the codebase is forbidden and
constitutes a blocking defect.

- Use `.env` files (gitignored) locally; use the deployment platform's secret management in
  production.
- Raw error messages or stack traces MUST NOT be surfaced to end users; display clear, actionable
  messages only.
- Dependencies MUST be kept up to date; known vulnerabilities MUST be resolved before shipping.

## Technology Stack

- **Backend**: Go 1.24 or higher, structured following the domain-based layout in
  `docs/coding-guidelines.md` (`cmd/`, `internal/<domain>/`, `pkg/`, `config/`).
- **Frontend**: React 19 or higher with TypeScript (strict mode), Vitest, React Testing Library, MSW,
  Playwright for E2E.
- **API**: RESTful HTTP; JSON request/response bodies.
- **Styling**: CSS Modules + CSS custom properties (no CSS-in-JS runtime or utility-class framework
  unless explicitly adopted by constitution amendment).
- **Icons**: single icon library project-wide (Lucide React is the default).
- **MVP scope**: web application only — no native mobile apps, no third-party booking integrations
  (see `docs/functional-requirements.md` Out of Scope section).

## Development Workflow

- Branching: feature work MUST be done on `feature/<descriptive-name>` branches. Direct commits to
  `main` are forbidden.
- Commits MUST follow Conventional Commits in English: `feat:`, `fix:`, `chore:`, `docs:`, `test:`,
  `refactor:`, etc.
- Releases follow Semantic Versioning 2.0.0 (SemVer): MAJOR.MINOR.PATCH.
- Pull requests MUST have all tests passing and all lint checks clean before merge.
- Quality gate order: lint → unit tests → integration tests → E2E tests.

## Team Structure

TrAIveler is developed by a team of four roles. Each role has a defined **primary ownership
boundary** — the area of the codebase and process for which that role is the first decision-maker
and required reviewer.

| Role | Primary Ownership | Secondary Involvement |
|------|------------------|----------------------|
| **Frontend Engineer** | `frontend/`, design tokens, Playwright E2E tests | API contracts, accessibility review |
| **Backend Engineer** | `backend/`, REST API contracts, business logic, unit/integration tests | CI pipeline for backend jobs |
| **Infrastructure Engineer** | `infra/`, `.github/workflows/`, deployment pipelines, AWS resources | Backend Dockerfile, environment config (`.env.example`) |
| **Product Manager (PM)** | `specs/`, `docs/functional-requirements.md`, user stories, acceptance criteria | Feature branch naming, spec quality reviews |

**Ownership rules**:
- A role's primary owner MUST be consulted before their ownership boundary is modified by another
  role. Unilateral changes to another role's area are blocked.
- Roles are not silos. Every engineer is encouraged to review and understand work outside their
  boundary, but accountability follows the table above.
- One person MAY hold multiple roles in a small team. In that case, a peer from any other role
  fulfils the required review obligation.

## Team Collaboration

Good engineering is a team sport. These rules govern how work flows from idea to merged code and
protect every contributor's time and focus.

### Pull Requests

- Every PR MUST receive at least one peer review and approval before it can be merged into `main`.
- A PR author MUST NOT merge their own PR, except for emergency hotfixes — which MUST receive a
  post-merge review within one working day.
- PR descriptions MUST explain **why** the change is needed, not just what changed. Link to the
  relevant spec (`specs/NNN-*/spec.md`) if one exists.
- PRs MUST be kept small and focused on a single concern. A PR that touches more than one
  unrelated concern MUST be split unless doing so would make the change incoherent.
- Draft PRs are encouraged for work in progress. Convert to "Ready for Review" only when all CI
  checks pass and the author considers the code complete.

**Review routing by role**: Reviewers MUST be assigned according to the ownership boundaries
defined in *Team Structure*. At minimum:

| PR touches… | Required reviewer role |
|-------------|------------------------|
| `frontend/` | Frontend Engineer |
| `backend/` | Backend Engineer |
| `infra/` or `.github/workflows/` | Infrastructure Engineer |
| `specs/` or `docs/functional-requirements.md` | Product Manager |
| Multiple domains | One reviewer per affected role |

If the author is the only person with a required role, they MUST request review from the
next-closest role and document the exception in the PR description.

### Code Reviews

- Reviewers MUST verify compliance with all five Core Principles. A principle violation is a
  blocking comment, not a suggestion.
- Non-blocking feedback MUST be prefixed with `nit:` or `suggestion:` so the author can clearly
  distinguish blocking from advisory comments.
- Reviews MUST be completed within one working day of the "Ready for Review" status being set.
  If a reviewer cannot meet this timeline, they MUST say so in the PR thread so the author can
  request a different reviewer.
- Approval means the reviewer accepts responsibility for the change landing in `main`. Rubber-stamp
  approvals without reading the diff are a constitution violation.

### Branch Hygiene

- Branches MUST be deleted after merge. Stale branches older than 30 days with no open PR are
  subject to deletion without notice.
- Branches MUST be rebased (not merged) onto `main` before opening a PR to keep history linear
  and readable.
- Force-pushes to `main` are forbidden under all circumstances.

## Governance

This constitution supersedes all other written or informal practices. Any conflict between this
document and any other guideline is resolved in favour of this constitution.

**Amendment procedure**:
1. Open a PR that modifies this file with a clear rationale.
2. Increment `CONSTITUTION_VERSION` following SemVer rules defined in the spec kit documentation.
3. Update `LAST_AMENDED_DATE` to the date the PR is merged.
4. Run the `speckit.constitution` command to propagate changes to all dependent templates.
5. Document the change in the PR description and in `.github/memory/session-notes.md`.

All PRs and code reviews MUST verify compliance with the five Core Principles above. Violations are
blocking. Complexity that cannot be justified against Principle II (Simplicity) MUST be removed.

For runtime development guidance refer to `docs/coding-guidelines.md`, `docs/testing-guidelines.md`,
and `docs/ui-guidelines.md`.

**Version**: 1.2.0 | **Ratified**: 2026-07-02 | **Last Amended**: 2026-07-02
