# Copilot Instructions

> **Auxiliary file.** Claude Code is the primary AI tool for this project as of the migration to
> Claude; GitHub Copilot is kept as an auxiliary/secondary tool only. The canonical, source-of-truth
> instructions live in [`/CLAUDE.md`](../CLAUDE.md). This file is a mirror for GitHub Copilot's
> auto-loading mechanism — it must be kept in sync **with `CLAUDE.md`**, never the other way around.
> If the two files disagree, `CLAUDE.md` wins.
>
> **Editing agents in `.github/agents/` or `.github/prompts/`?** Several of them are hand-ported to
> Claude Code subagents under `.claude/agents/`. Before and after editing one, run
> `python3 scripts/check-agent-drift.py` from the repo root and follow its instructions — see
> `CLAUDE.md`'s "Local Agents & Workflows" section for the full explanation and adaptation rules.

## Project Context
- Create a web tool to plan trips to any destination in the world. The tool offers suggestions for places, excursions, gastronomy, tips, and recommendations for exploring the most attractive and hidden treasures of a city or country within a specific timeframe.
- Stack: Backend: Golang 1.24 or higher, Frontend: React 19 or higher
- Actual Phase: Initial setup of the project, including the creation of the repository, initial commit, and basic project structure.

## Documentation References

Read the following files before generating code, tests, or UI for this project:

<!-- PROMOTED:doc-references START -->
<!-- Last updated: 2026-07-06 — includes promoted foundational decisions from specs 001-007 -->

### Product & Vision

- **docs/product-vision.md** — product identity, core problem, value proposition, target personas, MVP scope, explicit out-of-scope, and roles/permissions model.

### Requirements & Constraints

- **docs/functional-requirements.md** — normative requirements ("The system shall…"); defines what the application must do and what is out of scope for the MVP.
- **docs/nfrs.md** — measurable non-functional requirements with validation methods covering performance, scalability, availability, accessibility (WCAG 2.1 AA), security, maintainability, privacy, and observability.

### Architecture & Infrastructure

- **docs/architecture.md** — system component boundaries, integration rules, scalability constraints, security boundaries, and observability strategy. Defines how backend (Go on ECS Fargate), frontend (React on S3+CloudFront), database (PostgreSQL RDS), and external dependencies (Anthropic AI) interact.
- **docs/cloud-and-environments.md** — cloud provider (AWS us-east-1), environment topology (staging active, production dormant), IaC approach (Terraform), compute platform (ECS Fargate, not Lambda), CI/CD (GitHub Actions with OIDC), secrets management (AWS Secrets Manager), and cost strategy ($200 staging, $300-400 production).
- **docs/data-model.md** — complete catalog of 16 core entities with full attribute specifications, three-layer validation rules ([DB], [Logic], [API]), performance-critical indexes, forward-only state transitions, and cascade behavior. Includes authentication entities (User, RefreshToken, JWTSigningKey, SecurityEvent), subscription model (Plan, Subscription), trip domain (Trip, Day, Activity, Destination), collaboration (Collaborator, Suggestion), travel styles (TravelStyle, TripTravelStyle), AI conversation tracking (ConversationSession, ConversationMessage), relationships, validation rules, invariants, business rules, and database migration strategy with optimistic locking and GDPR compliance rules.

### Security & Authorization

- **docs/security.md** — authentication model (JWT RS256 with multi-key rotation), authorization roles (admin/partner), subscription limits, password security (bcrypt cost 12), session management, optimistic locking for concurrency, secrets management (AWS Secrets Manager), PII handling (GDPR-aware), input validation (prompt injection prevention), output sanitization, dependency security, security logging & monitoring (CloudWatch 30-day retention, alarms), and testing requirements.

### Development Standards

- **docs/api-design-standards.md** — comprehensive API design conventions covering resource naming (plural nouns, lowercase, hyphens), URL structure (max 2-level nesting), versioning (/api/v1), HTTP methods (GET/POST/PUT/PATCH/DELETE), request/response formats (snake_case JSON, ISO 8601 timestamps), standardized error format (error code, message, request_id, fields array), HTTP status codes (semantic 2xx/4xx/5xx), pagination (page/per_page with metadata envelope), filtering (query operators: [gte], [lte], [like]), sorting (minus prefix for descending), rate limiting (100/min authenticated, headers in all responses), and 7 endpoint patterns (list, get, create, update full/partial, delete, action). All new endpoints must follow these standards verified through code review checklist and integration tests.
- **docs/coding-guidelines.md** — formatting rules, import organization, naming conventions, and KISS/DRY principles for Go (backend) and React/TypeScript (frontend).
- **docs/testing-guidelines.md** — three-layer testing strategy (unit, integration, E2E), folder structure, naming conventions, and coverage targets (80% business logic, 80% shared components).
- **docs/ui-guidelines.md** — design tokens (color, spacing, typography), component layers (Atomic Design), responsive breakpoints, accessibility rules (WCAG 2.1 AA), and loading/error/empty state requirements.

<!-- PROMOTED:doc-references END -->

## Language Policy (MANDATORY)
- Conversation with the developer may be in English or Spanish; respond in whichever
  language the developer uses.
- ALL generated artifacts MUST be in English, without exception. This includes:
  source code, identifiers, comments, docstrings, commit messages, branch names,
  documentation (docs/, README, specs), test names, and any file content committed
  to the repository.
- Never mix languages inside an artifact. If the developer writes a request in
  Spanish, still produce the code/docs in English.

## Development Principles
- Test-Driven Development: Red-Green-Refactor
- Incremental, small, and testable changes
- Validation before commit: tests pass, no lint errors

## API Documentation Enforcement (MANDATORY)

Whenever a backend endpoint is added, **or an existing endpoint's request/response shape or
behavior changes**:

- Add/update its `swag` doc-comment annotations per `specs/009-api-documentation/contracts/api.md`'s
  shape (mirrors `internal/example/handler.go`, the canonical reference implementation — see
  `backend/README.md`).
- Regenerate `backend/docs/` via `make swagger` and commit the regenerated artifact alongside the
  code change (it's a generated file — never hand-edit it).
- Run the Update Documentation prompt/agent (`.github/agents/technical-writer.agent.md`) to review
  the annotation change as part of the same piece of work — a new or changed endpoint is not done
  until this review has confirmed the annotations are complete/accurate and the regenerated docs
  are committed.

This is a manual-but-mandatory step until the deferred `swagger-drift` CI gate (Sprint 5, spec 009
US3) lands and enforces it automatically — do not treat "CI will catch it later" as a reason to
skip this now.

## Git Workflow
- **Commit messages**: Follow `.github/COMMIT_GUIDELINES.md` (Conventional Commits, imperative mood, English only)
- **Pull requests**: Use `.github/PULL_REQUEST_TEMPLATE.md` (auto-loaded by GitHub, includes task checklist, dependencies, verification steps)
- **Branch naming**: feature/<descriptive-name> (English, lowercase, hyphens)
- **Protection**: Never commit directly to main
- **Versioning**: Semantic Versioning (SemVer) 2.0.0

## Memory System (MANDATORY)

This repository has **one** memory system, shared byte-for-byte between Copilot and Claude Code
under `.github/memory/`. The full protocol — including how to avoid conflicts and stale entries
between tools — is canonical in **`.github/memory/README.md`**. Read it once per session and
follow it; do not rely on the summary below if it ever seems to diverge, the README wins.

Quick reference:

1. **Session start**: read `.github/memory/session-notes.md`, `.github/memory/patterns-discovered.md`,
   and `.github/memory/scratch/working-notes.md`, then confirm loading:
   ```
   ✅ Memory System Loaded
   - Session notes: X sessions reviewed
   - Patterns discovered: Y patterns available
   - Working notes: [Active/Empty]
   - Ready to proceed with context-aware assistance
   ```
2. **While working**: take notes in `.github/memory/scratch/working-notes.md` (not committed).
3. **When a reusable pattern emerges**: append it to `.github/memory/patterns-discovered.md` using
   its template, tagged with today's date and `**Tool**: GitHub Copilot`.
4. **At session end**: append a summary to `.github/memory/session-notes.md` using its template,
   tagged with today's date and `**Tool**: GitHub Copilot`.
5. **Append-only**: never edit or delete another session's entry, regardless of which tool wrote it.

## Task Consolidation Policy (MANDATORY)

**Core Principle**: Consolidate atomic tasks BEFORE creating issues to reduce ticket waste and improve backlog health.

### Consolidation Rules

**DO consolidate** tasks into ONE issue when they meet ALL criteria:
- ✅ Same spec (never cross spec boundaries)
- ✅ Same area/module (same directory or logical component)
- ✅ Same tech stack (all Go OR all React, never mixed)
- ✅ Shared context (config files, middleware package, primitives set)
- ✅ Similar size (2-4 small tasks = 1 reviewable PR)
- ✅ No blocking dependencies between them
- ✅ Independent tracking not required

**DON'T consolidate** when:
- ❌ Different tech stacks (Go + TypeScript = separate)
- ❌ Critical blocker task (needs visibility)
- ❌ Different dependency chains
- ❌ Already large task (>200 LOC)
- ❌ Cross-spec boundary

**Group Column**: Use `docs/roadmap.md` Group column to mark consolidated tasks (e.g., `G-BACKEND-MIDDLEWARE`). Tasks sharing a Group value become ONE issue with checklist.

**Tracking**: Sprint planning and issue creation details are tracked in `.github/memory/session-notes.md`. Consolidation checklist available in `.github/SPRINT-CONSOLIDATION-CHECKLIST.md`.