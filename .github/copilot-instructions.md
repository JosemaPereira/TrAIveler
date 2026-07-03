# Copilot Instructions

## Project Context
- Create a web tool to plan trips to any destination in the world. The tool offers suggestions for places, excursions, gastronomy, tips, and recommendations for exploring the most attractive and hidden treasures of a city or country within a specific timeframe.
- Stack: Backend: Golang 1.24 or higher, Frontend: React 19 or higher
- Actual Phase: Initial setup of the project, including the creation of the repository, initial commit, and basic project structure.

## Documentation References

Read the following files before generating code, tests, or UI for this project:

<!-- PROMOTED:doc-references START -->
<!-- Last updated: 2026-07-03 — includes promoted foundational decisions -->

### Product & Vision

- **docs/product-vision.md** — product identity, core problem, value proposition, target personas, MVP scope, explicit out-of-scope, and roles/permissions model.

### Requirements & Constraints

- **docs/functional-requirements.md** — normative requirements ("The system shall…"); defines what the application must do and what is out of scope for the MVP.
- **docs/nfrs.md** — measurable non-functional requirements with validation methods covering performance, scalability, availability, accessibility (WCAG 2.1 AA), security, maintainability, privacy, and observability.

### Architecture & Infrastructure

- **docs/architecture.md** — system component boundaries, integration rules, scalability constraints, security boundaries, and observability strategy. Defines how backend (Go on ECS Fargate), frontend (React on S3+CloudFront), database (PostgreSQL RDS), and external dependencies (Anthropic AI) interact.
- **docs/cloud-and-environments.md** — cloud provider (AWS us-east-1), environment topology (staging active, production dormant), IaC approach (Terraform), compute platform (ECS Fargate, not Lambda), CI/CD (GitHub Actions with OIDC), secrets management (AWS Secrets Manager), and cost strategy ($200 staging, $300-400 production).
- **docs/data-model.md** — core entities (User, Trip, Day, Activity, RefreshToken, JWTSigningKey, SecurityEvent), relationships, validation rules, invariants, business rules, and database migration strategy. Includes authentication entities, optimistic locking, and GDPR compliance rules.

### Security & Authorization

- **docs/security.md** — authentication model (JWT RS256 with multi-key rotation), authorization roles (admin/partner), subscription limits, password security (bcrypt cost 12), session management, optimistic locking for concurrency, secrets management (AWS Secrets Manager), PII handling (GDPR-aware), input validation (prompt injection prevention), output sanitization, dependency security, security logging & monitoring (CloudWatch 30-day retention, alarms), and testing requirements.

### Development Standards

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
- Persistent memory: this file (.github/copilot-instructions.md) holds foundational
  principles and workflows.
- Working memory: the .github/memory/ directory holds discoveries and patterns.
- During active work, take notes in .github/memory/scratch/working-notes.md (not committed).
- When a reusable pattern emerges, document it in .github/memory/patterns-discovered.md (committed).
- At the end of a session, summarize key findings into .github/memory/session-notes.md (committed).
- Reference these files when giving context-aware suggestions.
- At the start of every session, read session-notes.md, patterns-discovered.md, and working-notes.md before doing any work.