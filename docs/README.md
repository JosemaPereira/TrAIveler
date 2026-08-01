# docs/

> Index of TrAIveler's documentation — start here to find the right document instead of browsing
> the directory listing.

[← Back to root README](../README.md)

Most files below carry a `<!-- PROMOTED:... -->` marker: they are generated from `specs/*/spec.md`
via the `/promote-foundations` workflow and represent durable, agreed decisions. Don't hand-edit
content inside a `PROMOTED:...START`/`END` block directly — re-run `/promote-foundations`, or (for a
local-only clarification that isn't a real decision change) append a new section after the block's
`END` marker instead, as several files already do (e.g. the "Local Development Note" addenda below).

---

## Product & Vision

| Document | Purpose |
|----------|---------|
| [product-vision.md](product-vision.md) | Product identity, core problem, value proposition, target personas, MVP scope, explicit out-of-scope, roles/permissions model |
| [capstone-idea.md](capstone-idea.md) | The original user story and pitch this project was bootstrapped from (historical — superseded in detail by `product-vision.md`, kept for context) |

## Requirements & Constraints

| Document | Purpose |
|----------|---------|
| [functional-requirements.md](functional-requirements.md) | Normative requirements ("The system shall…"); what the application must do and what's out of scope for the MVP |
| [nfrs.md](nfrs.md) | Measurable non-functional requirements with validation methods: performance, scalability, availability, accessibility (WCAG 2.1 AA), security, maintainability, privacy, observability |

## Architecture & Infrastructure

| Document | Purpose |
|----------|---------|
| [architecture.md](architecture.md) | System component boundaries, integration rules, scalability constraints, security boundaries, observability strategy |
| [cloud-and-environments.md](cloud-and-environments.md) | Cloud provider (AWS us-east-1), environment topology (staging active, production dormant), IaC approach (Terraform), compute platform (ECS Fargate), CI/CD (GitHub Actions + OIDC), secrets management, cost strategy |
| [data-model.md](data-model.md) | Complete catalog of 16 core entities with full attribute specs, three-layer validation rules (`[DB]`/`[Logic]`/`[API]`), performance indexes, forward-only state transitions, cascade behavior |
| [local-ai-setup.md](local-ai-setup.md) | Local-development-only addendum: running the backend against a free local Ollama+Gemma server instead of Anthropic Claude |

Note: `architecture.md`/`cloud-and-environments.md` describe the target design agreed at the
foundation phase and generally don't need a per-module edit every time a Terraform module ships —
see [infra/README.md](../infra/README.md) for what's actually built so far, and `docs/roadmap.md`
for per-task status. Only touch the `PROMOTED:...` block itself if the target design genuinely
changed, not just because an implementation landed.

## Security & Authorization

| Document | Purpose |
|----------|---------|
| [security.md](security.md) | Authentication (JWT RS256, multi-key rotation), authorization roles (admin/partner), subscription limits, password security (bcrypt cost 12), session management, optimistic locking, secrets management, PII/GDPR handling, input validation (prompt injection prevention), output sanitization, dependency security, security logging & monitoring |

## Development Standards

| Document | Purpose |
|----------|---------|
| [api-design-standards.md](api-design-standards.md) | API conventions: resource naming, URL structure, versioning (`/api/v1`), HTTP methods, request/response formats, standardized error format, status codes, pagination, filtering, sorting, rate limiting |
| [coding-guidelines.md](coding-guidelines.md) | Formatting rules, import organization, naming conventions, KISS/DRY principles for Go (backend) and React/TypeScript (frontend) |
| [testing-guidelines.md](testing-guidelines.md) | Three-layer testing strategy (unit, integration, E2E), folder structure, naming conventions, coverage targets (90% business logic, 90% shared components) |
| [ui-guidelines.md](ui-guidelines.md) | Design tokens (color, spacing, typography), component layers (Atomic Design), responsive breakpoints, accessibility rules, loading/error/empty state requirements |
| [mock-standards.md](mock-standards.md) | Standardized mock generation pattern and location convention used across the backend |

## Project Planning & Process

| Document | Purpose |
|----------|---------|
| [roadmap.md](roadmap.md) | Source of truth for task existence, sprint assignment, and status — generated/reconciled by `/build-roadmap`, human-owned Priority/Status/Phase/Issue/Notes columns |
| [project-workflow.md](project-workflow.md) | End-to-end operating manual: how the project was bootstrapped, the SpecKit + agentic-kit tooling, and the repeatable flow from a spec to a shipped, tracked piece of work (documents the Copilot invocation surface; see `CLAUDE.md` for Claude Code's equivalent, canonical mechanism) |

---

## Related, not in this directory

- [CLAUDE.md](../CLAUDE.md) — canonical AI-agent instructions for this repo (source of truth; `.github/copilot-instructions.md` mirrors it)
- [.github/memory/](../.github/memory/) — shared, tool-agnostic session notes and discovered patterns
- [specs/](../specs/) — the SpecKit spec/plan/tasks artifacts these docs are generated from
- [infra/README.md](../infra/README.md), [backend/README.md](../backend/README.md), [frontend/README.md](../frontend/README.md), [e2e/README.md](../e2e/README.md) — per-area setup and current implementation status
