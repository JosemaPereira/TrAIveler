# Implementation Plan: System Architecture and Technology Stack

**Branch**: `005-system-architecture` | **Date**: 2026-07-03 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/005-system-architecture/spec.md`

**Note**: This document defines the architectural foundation for backend, frontend, and infrastructure. All subsequent feature implementations will build upon these patterns.

## Summary

This specification establishes the complete architectural foundation for the TrAIveler application across three primary domains: backend (Go REST API), frontend (React SPA), and infrastructure (AWS via Terraform). The architecture defines:

- **Backend**: Domain-driven directory structure with layered separation (handlers, services, repositories, models), global middleware chain (request ID, logging, panic recovery, CORS), connection pooling (pgx for PostgreSQL, min 5/max 25 connections per task), and interface abstractions for external dependencies (AI provider)
- **Frontend**: Atomic Design component layering (primitives → composites → features), TanStack Query for API communication with native fetch wrapper, Zustand for global state, CSS Modules with design token system, and React Router v7 for routing
- **Infrastructure**: Terraform modules for AWS resources (VPC, ECS Fargate, RDS, ALB, CloudFront, Secrets Manager), environment separation via .tfvars files (staging: cost-optimized at $150-180/month, production: dormant until validated via load testing), and CI/CD integration with GitHub Actions OIDC

Technical approach prioritizes cost optimization for staging MVP (0.25 vCPU/0.5 GB RAM ECS tasks, 1-day RDS backups, 7-day log retention) while maintaining production-ready patterns that can scale via configuration changes validated through load testing.

## Technical Context

**Language/Version**: 
- Backend: Go 1.24 or higher
- Frontend: React 19 with TypeScript (strict mode)
- Infrastructure: Terraform 1.5+ (HCL syntax)

**Primary Dependencies**:
- Backend: Chi (HTTP router), pgx/v5 (PostgreSQL driver), goose/v3 (migrations), Anthropic SDK (AI), log/slog (structured logging), golang-jwt/jwt v5 (authentication)
- Frontend: Vite (build tool), TanStack Query v5 (data fetching), Zustand (state management), React Router v7 (routing), Lucide React (icons), Vitest + React Testing Library + MSW (testing)
- Infrastructure: AWS SDK, Terraform AWS provider

**Storage**: PostgreSQL 15.4 on Amazon RDS (db.t4g.micro for staging, Multi-AZ for production)

**Testing**: 
- Backend: Go testing framework (unit/integration), gosec (SAST), gole aks (secrets)
- Frontend: Vitest (unit), React Testing Library (component), MSW (API mocking), Playwright (E2E), axe-core (accessibility), Lighthouse CI (performance)
- Integration: k6 (load testing)

**Target Platform**: 
- Backend: AWS ECS Fargate ARM64 Graviton2 (0.25 vCPU/0.5 GB RAM staging, 1 vCPU/2 GB RAM production)
- Frontend: AWS S3 + CloudFront CDN (global edge caching, HTTPS-only)
- Database: Amazon RDS PostgreSQL in private VPC subnets
- Networking: VPC with public/private subnets across 2 AZs, ALB, NAT instance (staging)/NAT Gateway (production)

**Project Type**: Web application (three-tier architecture: React SPA + Go REST API + PostgreSQL)

**Performance Goals**: 
- Non-AI API endpoints: p95 latency ≤ 500 ms at 100 RPS (NFR-PERF-001)
- AI itinerary generation: acknowledgement ≤ 3s, full result ≤ 30s (NFR-PERF-002)
- Frontend: LCP ≤ 2.5s, CLS ≤ 0.1, INP ≤ 200ms on 4G (NFR-PERF-003)
- Frontend bundle: < 500 KB gzipped (SC-006)
- Backend cold start: ready-to-serve (healthz 200 OK) ≤ 60s (NFR-AVAIL-003, SC-005)

**Constraints**: 
- Cost: $200/month staging budget (must use cheapest viable AWS tiers)
- Accessibility: WCAG 2.1 AA compliance mandatory (4.5:1 contrast, keyboard operability, zero axe-core violations)
- Test Coverage: 80% backend business logic, 80% frontend shared components, 100% API endpoints (integration), all primary flows (E2E)
- Security: No secrets in Git, prompt injection validation, output sanitization, JWT RS256 with multi-key rotation
- Observability: 95% of errors traceable via correlation IDs (SC-007)

**Scale/Scope**: 
- MVP Staging: < 50 concurrent users, auto-scaling 1-2 ECS tasks, single RDS instance (single-AZ)
- Production Target: 500 concurrent VUs sustained 10 min (NFR-SCALE-001), auto-scaling 2-20 tasks, Multi-AZ RDS
- Database connections: min 5, max 25 per ECS task (clarification #4)
- CloudWatch retention: 7 days staging, 30 days production (clarification #7)
- RDS backups: 1 day staging, 30 days production (clarification #8)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence / Notes |
|-----------|--------|------------------|
| **I. Test-First Development** | ✅ PASS | Spec defines comprehensive testing strategy: Go testing (unit/integration), Vitest+RTL+MSW (frontend unit/component), Playwright (E2E), 80% coverage targets for backend business logic and frontend shared components (FR-MAINT-001, FR-MAINT-002), all API endpoints (integration), all primary flows (E2E). TDD workflow enforced via constitution. |
| **II. Simplicity (KISS & DRY)** | ✅ PASS | Architecture follows industry-standard patterns without speculative abstractions: domain-driven backend (handler/service/repository/model layers per FR-002), Atomic Design frontend (primitives → composites → features per FR-009), modular Terraform (reusable modules per FR-017). No custom frameworks or unnecessary complexity. Single monolithic backend (not microservices) keeps deployment simple for MVP (Assumptions). |
| **III. Code Quality & Consistency** | ✅ PASS | Spec mandates automated quality gates: golangci-lint for Go (errcheck, govet, staticcheck, revive, gosec), ESLint + Prettier for TypeScript/React, gofmt, zero errors required before merge. TypeScript strict mode enforced. Import order conventions in docs/coding-guidelines.md. All code in English (constitution III). |
| **IV. Accessible & Token-Driven UI** | ✅ PASS | Frontend requires WCAG 2.1 AA compliance (NFR-A11Y-001 to NFR-A11Y-004): 4.5:1 contrast, keyboard operability, zero axe-core violations, Lighthouse ≥ 90. All styling via CSS Modules + design tokens (FR-012), no hard-coded values (SC-002). Atomic Design layering enforced (FR-009). Loading/Error/Empty states mandatory (constitution IV). |
| **V. Secure Configuration** | ✅ PASS | Secrets stored in AWS Secrets Manager (FR-020, SecretsModule), loaded via environment variables (FR-008). No secrets in Git enforced via gitleaks CI gate (NFR-SEC-001). Prompt injection validation (FR-007, constitution prompt injection rules), output sanitization (constitution output sanitization rules). No raw errors to users (NFR-SEC-005). |

**Overall**: ✅ ALL GATES PASS — No constitution violations. Architecture follows established patterns with appropriate justification for all complexity.

## Project Structure

### Documentation (this feature)

```text
specs/005-system-architecture/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (technology decisions - generated below)
├── data-model.md        # Phase 1 output (architectural components - generated below)
├── quickstart.md        # Phase 1 output (validation scenarios - generated below)
├── contracts/           # Phase 1 output (integration contracts - generated below)
│   ├── backend-layers.md      # Backend layer contracts
│   ├── frontend-patterns.md   # Frontend architectural patterns
│   └── infrastructure.md      # Terraform module contracts
├── checklists/
│   └── requirements.md  # Specification quality checklist (completed)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   └── api/                    # Main HTTP server entry point
│       └── main.go
├── internal/                   # Private application code (domain-driven structure)
│   ├── <domain>/              # Per-domain packages (e.g., auth, trips, suggestions)
│   │   ├── handler.go         # HTTP handlers (presentation layer)
│   │   ├── service.go         # Business logic (service layer)
│   │   ├── repository.go      # Data access (repository layer)
│   │   └── model.go           # Domain types (models)
│   ├── middleware/            # Global HTTP middleware
│   │   ├── request_id.go
│   │   ├── logger.go
│   │   ├── recovery.go
│   │   └── cors.go
│   ├── database/              # Database client and connection pooling
│   │   └── client.go
│   ├── ai/                    # AI provider client and abstractions
│   │   ├── client.go
│   │   ├── validator.go       # Prompt injection validation
│   │   └── sanitizer.go       # Output sanitization
│   └── errors/                # Centralized error handling
│       └── handler.go
├── pkg/                       # Public packages (if needed for sharing)
├── config/                    # Configuration loading and validation
│   ├── config.go
│   └── env.go
├── migrations/                # Database schema migrations (goose)
│   ├── 00001_initial_schema.sql
│   └── ...
├── tests/
│   ├── integration/           # Integration tests (API + database)
│   └── fixtures/              # Test data and mocks
├── .env.example               # Environment variable template
├── go.mod
├── go.sum
├── Dockerfile                 # Multi-stage build for ECS deployment
└── README.md

frontend/
├── src/
│   ├── components/
│   │   ├── primitives/        # Atomic Design: atoms (Button, Input, Card, Modal)
│   │   └── composites/        # Atomic Design: molecules (Form, DataTable, NavigationBar)
│   ├── features/              # Feature-specific components
│   │   ├── <feature-name>/
│   │   │   ├── components/    # Atomic Design: organisms/templates
│   │   │   ├── hooks/         # Feature-specific hooks
│   │   │   └── types/         # Feature-specific types
│   ├── hooks/                 # Shared React hooks
│   ├── lib/                   # Utility libraries
│   │   ├── api-client.ts      # Fetch API wrapper with base URL, headers, error handling
│   │   └── query-client.ts    # TanStack Query configuration
│   ├── stores/                # Zustand global state stores
│   │   └── auth-store.ts      # Authentication state
│   ├── styles/
│   │   ├── tokens.css         # CSS custom properties (colors, spacing, typography)
│   │   └── global.css         # Global styles
│   ├── routes/                # React Router v7 configuration
│   │   └── index.tsx
│   ├── App.tsx                # Root component with providers
│   ├── main.tsx               # Vite entry point
│   └── vite-env.d.ts
├── tests/
│   ├── unit/                  # Vitest unit tests
│   ├── component/             # React Testing Library component tests
│   └── __mocks__/             # MSW API mocks
├── public/                    # Static assets
├── .env.example
├── package.json
├── tsconfig.json              # TypeScript strict mode
├── vite.config.ts
├── vitest.config.ts
└── README.md

e2e/
├── tests/
│   ├── auth.spec.ts           # Playwright E2E tests for authentication
│   ├── trips.spec.ts          # Playwright E2E tests for trip management
│   └── accessibility.spec.ts  # axe-core accessibility tests
├── fixtures/
├── playwright.config.ts
├── package.json
└── README.md

infra/                         # Terraform infrastructure as code
├── modules/                   # Reusable Terraform modules
│   ├── vpc/                   # VPC with public/private subnets, NAT, IGW
│   ├── ecs/                   # ECS cluster, task definition, service, auto-scaling
│   ├── rds/                   # RDS PostgreSQL instance, subnet group, security group
│   ├── alb/                   # Application Load Balancer, target group, listeners
│   ├── cloudfront/            # CloudFront distribution, S3 origin, OAI
│   └── secrets/               # AWS Secrets Manager secrets
├── environments/
│   ├── staging.tfvars         # Staging-specific configuration (cost-optimized)
│   └── production.tfvars      # Production configuration (dormant until validated)
├── main.tf                    # Root module that calls reusable modules
├── variables.tf               # Input variables
├── outputs.tf                 # Output values for CI/CD (ECR URL, ECS cluster, S3 bucket, CloudFront ID)
├── backend.tf                 # S3 + DynamoDB remote state configuration
├── versions.tf                # Terraform and provider versions
└── README.md

.github/
├── workflows/
│   ├── backend-ci.yml         # Backend lint, test, build, ECR push
│   ├── frontend-ci.yml        # Frontend lint, test, build, S3 sync
│   ├── e2e-ci.yml             # E2E tests on staging
│   ├── infra-plan.yml         # Terraform plan on PR
│   └── infra-apply.yml        # Terraform apply to staging on main merge
├── copilot-instructions.md    # Copilot workspace instructions
└── memory/                    # Memory system for patterns and session notes

docs/
├── architecture.md            # System architecture (promoted from specs)
├── coding-guidelines.md       # Go and TypeScript/React coding standards
├── testing-guidelines.md      # Testing strategy and conventions
├── ui-guidelines.md           # Design tokens, Atomic Design, accessibility
├── product-vision.md          # Product identity, personas, MVP scope
├── nfrs.md                    # Non-functional requirements
├── security.md                # Authentication, authorization, secrets management
├── cloud-and-environments.md  # AWS strategy, environment topology
├── data-model.md              # Core entities and relationships
└── roadmap.md                 # Project roadmap with all tasks
```

**Structure Decision**: The project uses **Option 2: Web application** structure with separate `backend/`, `frontend/`, and `e2e/` top-level directories, plus `infra/` for Terraform. This separation aligns with team ownership boundaries (Backend Engineer owns backend/, Frontend Engineer owns frontend/, Infrastructure Engineer owns infra/). The backend follows domain-driven structure (internal/<domain>/ with handler/service/repository/model layers per FR-002). The frontend follows Atomic Design layering (primitives → composites → features per FR-009). Infrastructure uses modular Terraform design (reusable modules/ called by root module per FR-017).

## Phase 1 Design Artifacts

**Completion Date**: 2026-07-03

### Generated Artifacts

1. **[research.md](research.md)** (Phase 0 output): 9 technology decisions resolving all NEEDS CLARIFICATION markers from Technical Context:
   - Frontend HTTP client selection (native fetch vs Axios)
   - Terraform environment separation strategy (.tfvars vs workspaces)
   - Database migration concurrency handling (goose advisory locks)
   - Database connection pool sizing (min 5, max 25 per task)
   - ECS auto-scaling configuration (min 1, max 2 tasks @ 70% CPU)
   - ECS task resource allocation (0.25 vCPU / 0.5 GB RAM ARM64 for staging)
   - CloudWatch log retention (7 days staging, 30 days production)
   - RDS backup retention (1 day staging, 30 days production)
   - Production configuration validation strategy (load test → adjust .tfvars → provision)

2. **[data-model.md](data-model.md)** (Phase 1 output): 26 architectural components across all layers:
   - Backend (7): HTTPServer, Middleware Chain, DatabaseClient, AIClient, RepositoryPattern, ServicePattern, ErrorHandler
   - Frontend (7): AppRouter, QueryProvider, AuthStore, APIClient, DesignTokens, ComponentLibrary, ErrorBoundary
   - Infrastructure (6 modules + 1 outputs): VPCModule, ECSModule, RDSModule, ALBModule, CloudFrontModule, SecretsModule, CICDOutputs
   - Includes 12 invariants across all layers ensuring consistency and correctness

3. **[contracts/backend-layers.md](contracts/backend-layers.md)** (Phase 1 output): Integration contracts between backend architectural layers:
   - Layer responsibilities (Handler/Service/Repository)
   - Handler ↔ Service contract (HTTP request → service method → domain response → HTTP response)
   - Service ↔ Repository contract (business logic → SQL queries → domain models)
   - Service ↔ AI Client contract (prompt validation → AI call → output sanitization)
   - Middleware ↔ Handler contract (context propagation for requestID, userID)
   - Testing contracts for service and repository layers
   - Error handling rules (domain errors → HTTP status codes via ErrorHandler)

4. **[contracts/frontend-patterns.md](contracts/frontend-patterns.md)** (Phase 1 output): Frontend architectural patterns:
   - Pattern 1: Atomic Design component layering (primitives → composites → features)
   - Pattern 2: TanStack Query for API communication (query key convention, hook patterns, error handling)
   - Pattern 3: Zustand for global state (store definition, typed selectors, persistence)
   - Pattern 4: API client integration (fetch wrapper, error handling, correlation IDs)
   - Pattern 5: Design tokens and CSS Modules (no hard-coded values, token usage)
   - Pattern 6: Accessibility requirements (keyboard operability, focus indicators, ARIA labels)
   - Pattern 7: Loading, Error, Empty states (mandatory for all data components)
   - Pattern 8: Testing patterns (React Testing Library, MSW, accessibility tests)

5. **[contracts/infrastructure.md](contracts/infrastructure.md)** (Phase 1 output): Terraform module and CI/CD integration contracts:
   - Module input/output conventions (standardized tagging, resource identifiers)
   - VPC Module contract (isolated VPC, public/private subnets, NAT instance vs NAT Gateway)
   - ECS Module contract (Fargate cluster, task definition, auto-scaling, health checks)
   - RDS Module contract (PostgreSQL instance, backups, security groups, secrets)
   - ALB Module contract (HTTPS termination, target groups, SSL certificates)
   - CloudFront Module contract (S3 origin, OAI, cache behavior, invalidation)
   - Secrets Module contract (AWS Secrets Manager integration, IAM permissions)
   - Environment configuration contract (.tfvars structure for staging/production)
   - Remote state management contract (S3 backend with DynamoDB locking)
   - CI/CD integration contracts (backend/frontend/infra workflows, OIDC authentication)

6. **[quickstart.md](quickstart.md)** (Phase 1 output): Runnable validation scenarios proving architecture works end-to-end:
   - Scenario 1: Backend skeleton validation (middleware, database client, health check)
   - Scenario 2: Frontend scaffold validation (TanStack Query, Zustand, design tokens)
   - Scenario 3: Infrastructure plan validation (Terraform modules, remote state, staging plan)
   - Scenario 4: End-to-end integration validation (frontend → backend → database)
   - Scenario 5: Accessibility validation (WCAG 2.1 AA compliance via axe-core)

### Phase 1 Constitution Re-check

*Re-evaluate constitution principles after Phase 1 design artifacts complete.*

| Principle | Status | Evidence / Notes |
|-----------|--------|------------------|
| **I. Test-First Development** | ✅ PASS | Design artifacts include comprehensive testing contracts: [contracts/backend-layers.md](contracts/backend-layers.md) defines service/repository testing contracts (unit tests with mocks, integration tests with real database). [contracts/frontend-patterns.md](contracts/frontend-patterns.md) Pattern 8 defines React Testing Library + MSW patterns. [quickstart.md](quickstart.md) includes 5 validation scenarios proving architecture works. No violations introduced. |
| **II. Simplicity (KISS & DRY)** | ✅ PASS | Design artifacts enforce clear separation of concerns: Backend uses Handler/Service/Repository pattern (no direct DB access in handlers, no HTTP dependencies in services per [contracts/backend-layers.md](contracts/backend-layers.md)). Frontend uses Atomic Design layering (primitives → composites → features per [contracts/frontend-patterns.md](contracts/frontend-patterns.md)). Infrastructure uses modular Terraform with reusable modules per [contracts/infrastructure.md](contracts/infrastructure.md). No speculative abstractions introduced. |
| **III. Code Quality & Consistency** | ✅ PASS | Design artifacts enforce quality standards: [contracts/backend-layers.md](contracts/backend-layers.md) mandates error wrapping (`fmt.Errorf("context: %w", err)`), no logging in service layer (middleware handles). [contracts/frontend-patterns.md](contracts/frontend-patterns.md) requires TypeScript strict mode, no `any`, typed selectors for Zustand. [contracts/infrastructure.md](contracts/infrastructure.md) enforces resource tagging convention (Environment/ManagedBy/Project). All code examples use English identifiers. No violations introduced. |
| **IV. Accessible & Token-Driven UI** | ✅ PASS | Design artifacts enforce accessible UI: [contracts/frontend-patterns.md](contracts/frontend-patterns.md) Pattern 5 mandates design tokens (no hard-coded values in CSS, all styling references `var(--token-name)`). Pattern 6 requires WCAG 2.1 AA compliance (keyboard operability, focus indicators, ARIA labels). Pattern 7 mandates Loading/Error/Empty states for all data components. [quickstart.md](quickstart.md) Scenario 5 validates zero axe-core violations. No violations introduced. |
| **V. Secure Configuration** | ✅ PASS | Design artifacts enforce secure configuration: [contracts/backend-layers.md](contracts/backend-layers.md) includes Service ↔ AI Client contract with prompt validation (reject instruction-override patterns) and output sanitization (strip HTML/script tags). [contracts/infrastructure.md](contracts/infrastructure.md) defines SecretsModule (AWS Secrets Manager) and OIDC authentication (no long-lived credentials). [quickstart.md](quickstart.md) uses environment variables (DATABASE_URL, VITE_API_BASE_URL). No violations introduced. |

**Overall**: ✅ ALL GATES PASS — Design artifacts align with all 5 constitution principles. No violations introduced during Phase 1.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

N/A — All constitution checks passed in both initial gate and Phase 1 re-check. No violations to justify.
