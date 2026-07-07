# Copilot Instructions

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

## Git Workflow
- Conventional commits (in English): feat:, fix:, chore:, docs:, etc.
- Feature branches: feature/<descriptive-name>  (branch names in English)
- Never commit directly to main
- Versioning: Semantic Versioning (SemVer) 2.0.0

## Memory System

### Overview
- **Persistent memory**: This file (.github/copilot-instructions.md) holds foundational principles and workflows.
- **Working memory**: The .github/memory/ directory holds discoveries, patterns, and session history.
- **Scratch notes**: .github/memory/scratch/working-notes.md for active session notes (not committed).
- **Patterns**: .github/memory/patterns-discovered.md for reusable implementation patterns (committed).
- **Session history**: .github/memory/session-notes.md for completed session summaries (committed).

### Session Start Protocol (MANDATORY)

**Every new session MUST begin by loading memory in this order:**

1. **Load Session Notes** (`.github/memory/session-notes.md`)
   - Review all completed session summaries
   - Understand what has been built and decided
   - Note any pending follow-ups or blockers

2. **Load Patterns Discovered** (`.github/memory/patterns-discovered.md`)
   - Review all accumulated implementation patterns
   - Apply proven solutions to similar problems
   - Avoid re-discovering known patterns

3. **Load Working Notes** (`.github/memory/scratch/working-notes.md`)
   - Check for in-progress work from previous session
   - Resume from documented stopping point if applicable

4. **Confirm Memory Load**
   - After loading all memory files, output this confirmation:
   ```
   ✅ Memory System Loaded
   - Session notes: X sessions reviewed
   - Patterns discovered: Y patterns available
   - Working notes: [Active/Empty]
   - Ready to proceed with context-aware assistance
   ```

**During active work**: Take notes in .github/memory/scratch/working-notes.md

**When a reusable pattern emerges**: Document it in .github/memory/patterns-discovered.md

**At session end**: Summarize key findings into .github/memory/session-notes.md

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

**Reference**: See `.github/PM-WORKFLOW-CONSOLIDATION.md` for detailed guidelines.

## Project Reports (Audit Trail)
- **PROMOTION-REPORT.md**: Tracks which foundational specs (001-006) have been promoted to docs/ and constitution. Update when new foundation specs are created or existing ones are revised.
- **ROADMAP-RECONCILIATION-REPORT.md**: Documents integration of spec tasks into docs/roadmap.md. Update when new specs are added, tasks change status/priority, or critical path changes.
- **ISSUE-CREATION-GUIDELINES.md**: Defines rules for epic vs issue classification, relationship management (blocks/blocked by/related to), label strategy, GitHub issue creation workflow, and **task consolidation policy**. **MANDATORY** reading before creating any GitHub issues.
- **PM-WORKFLOW-CONSOLIDATION.md**: Detailed consolidation-first PM workflow with rules, examples, and success metrics. **MANDATORY** for sprint planning and issue creation.
- **Update trigger**: When running `/promote-fundations` or `/build-roadmap` workflows, update the relevant report with new spec information, date, and statistics.
- **Location**: All reports live in .github/ directory for centralized audit trail.