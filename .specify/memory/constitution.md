<!--
SYNC IMPACT REPORT
==================
Version change: 1.4.0 → 1.4.1
Modified principles: none
Added sections: none
Modified sections: Technology Stack (Database line corrected 15.4 → 15.18 — routine CVE-review
version bump, issue #211; not a mandated-stack decision change, PostgreSQL 15.x remains the
mandate)
Removed sections: none
Templates requiring updates:
  ✅ .specify/templates/plan-template.md — no impact
  ✅ .specify/templates/spec-template.md — no impact
  ✅ .specify/templates/tasks-template.md — no impact
Follow-up TODOs: none — this is a patch-level factual correction, not a new decision.
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

<!-- PROMOTED:security-rules START -->
<!-- Last updated: 2026-07-03 from specs 002 (NFR-SEC-007, NFR-SEC-008) and 004 (security-auth-model) -->

#### Prompt Injection Prevention (NON-NEGOTIABLE)

All user input sent to the AI provider MUST pass through a prompt-validation layer that:
- Detects and rejects instruction-override patterns (e.g., "ignore previous instructions")
- Detects and rejects attempts to extract system prompts or internal configuration
- Detects and rejects off-topic prompts unrelated to travel planning
- Returns `400 Bad Request` with correlation ID logged for security review

The prompt-validation layer operates server-side only. Its classification rules MUST NOT be disclosed
in client-side code.

#### Output Sanitization (NON-NEGOTIABLE)

All AI-generated content MUST be sanitized before:
- Rendering in the browser
- Persisting to the database

Sanitization MUST strip or escape:
- HTML/script tags and event handlers
- Executable content (JavaScript, data URIs)
- Raw SQL or template injection attempts

No AI-generated content may be executed as code or used as a raw SQL/template value.

#### Authorization Enforcement

- Admin role: Full CRUD on own trips; exclusive authority to approve/reject partner suggestions
- Partner role: View-only on shared trips; suggestion-only modifications (no direct edits)
- Basic plan limit: One admin user + maximum one partner collaborator per subscription

These role boundaries MUST be enforced at the API layer on every request. Client-side enforcement is
advisory only.

<!-- PROMOTED:security-rules END -->

## Technology Stack

<!-- PROMOTED:mandated-stack START -->
<!-- Last updated: 2026-07-10 from specs 001, 002, 003, 009 -->

### Application Layer

- **Backend**: Go 1.24 or higher, structured following the domain-based layout in
  `docs/coding-guidelines.md` (`cmd/`, `internal/<domain>/`, `pkg/`, `config/`).
- **Frontend**: React 19 or higher with TypeScript (strict mode), Vitest, React Testing Library, MSW,
  Playwright for E2E.
- **API**: RESTful HTTP; JSON request/response bodies.
- **API Documentation**: OpenAPI v3 contract generated code-first from Go doc-comment annotations via
  `github.com/swaggo/swag`, served through Swagger UI via `github.com/swaggo/http-swagger/v2` (with
  `github.com/swaggo/files` for embedded UI assets); the generated `backend/docs/` artifact is
  committed and its freshness enforced by a CI drift-check (see `docs/testing-guidelines.md`). This
  complements, and does not replace, `docs/api-design-standards.md` as the human-readable source of
  truth for conventions.
- **Styling**: CSS Modules + CSS custom properties (no CSS-in-JS runtime or utility-class framework
  unless explicitly adopted by constitution amendment).
- **Icons**: single icon library project-wide (Lucide React is the default).
- **Database**: PostgreSQL 15.18 on Amazon RDS.
- **AI Provider**: Anthropic Claude API for itinerary generation.

### Infrastructure Layer (NON-NEGOTIABLE)

- **Cloud Provider**: AWS (Amazon Web Services), us-east-1 region
- **Infrastructure as Code**: Terraform 1.5+ with HCL syntax; remote state in S3 with DynamoDB locking
- **Compute Platform**: ECS Fargate with ARM64 Graviton2 containers (NOT Lambda — AI workloads require
  unlimited execution time, no payload limits, persistent HTTP connections)
- **Container Registry**: Amazon ECR for private Docker images
- **Frontend Delivery**: S3 for static assets + CloudFront CDN for global delivery
- **Load Balancing**: Application Load Balancer (ALB) with HTTPS termination
- **Networking**: VPC per environment with isolated public/private subnets
- **Secrets Management**: AWS Secrets Manager (no secrets in code or Git)
- **Logging**: Structured JSON logs to AWS CloudWatch Logs
- **Monitoring**: AWS CloudWatch metrics and dashboards
- **CI/CD Platform**: GitHub Actions with OIDC federation to AWS (no long-lived credentials)

### Environment Strategy

- **Staging**: Active MVP environment with cost-optimized configuration ($200/month budget)
- **Production**: IaC-defined but dormant until alpha release ($300-400/month when provisioned)
- **Isolation**: Separate VPCs within single AWS account (staging: 10.0.0.0/16, production: 10.1.0.0/16)
- **Deployment**: Auto-deploy to staging on main merge; manual-only production deployment

### MVP Scope

- Web application only — no native mobile apps, no third-party booking integrations
  (see `docs/functional-requirements.md` Out of Scope section)
- Subscription payment collection is mocked with a stub; plan enforcement is active application logic

<!-- PROMOTED:mandated-stack END -->

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

**Version**: 1.4.1 | **Ratified**: 2026-07-02 | **Last Amended**: 2026-08-01
