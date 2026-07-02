<!--
SYNC IMPACT REPORT
==================
Version change: (none — initial ratification) → 1.0.0
Modified principles: N/A (first version)
Added sections: Core Principles, Technology Stack, Development Workflow, Governance
Removed sections: N/A
Templates requiring updates:
  ✅ .specify/templates/plan-template.md — Constitution Check section is generic; no update required
  ✅ .specify/templates/spec-template.md — no principle-driven mandatory sections need updating
  ✅ .specify/templates/tasks-template.md — task categories (testing, linting, setup) align with principles
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

**Version**: 1.0.0 | **Ratified**: 2026-07-02 | **Last Amended**: 2026-07-02
