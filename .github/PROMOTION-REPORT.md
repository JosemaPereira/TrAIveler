# Foundation Specs Promotion Report

**Generated**: 2026-07-03 (Updated: 2026-07-06)  
**Status**: ✅ COMPLETE

## Executive Summary

All 7 foundational specifications have been successfully promoted to the project's persistent context (constitution + docs/). Subsequent `/speckit.plan`, `/speckit.tasks`, and implementation work will automatically inherit these decisions without re-reading individual specs.

---

## Foundational Specs Analyzed

| Spec ID | Title | Status | Primary Destination |
|---------|-------|--------|---------------------|
| **001** | Product Vision and Scope | ✅ Promoted | docs/product-vision.md |
| **002** | Non-Functional Requirements | ✅ Promoted | docs/nfrs.md |
| **003** | Cloud & Environments Strategy | ✅ Promoted | docs/cloud-and-environments.md |
| **004** | Security & Authentication Model | ✅ Promoted | docs/security.md |
| **005** | System Architecture & Technology Stack | ✅ Promoted | docs/architecture.md, docs/data-model.md, constitution |
| **006** | Core Domain and Data Model Foundations | ✅ Promoted | docs/data-model.md |
| **007** | API Design Standards and Conventions | ✅ Promoted | docs/api-design-standards.md |

---

## Promotion Details by Spec

### Spec 001: Product Vision and Scope

**Durable Decisions Promoted**:
- Product identity (TrAIveler tagline, core problem, value proposition)
- Target personas (4 personas: First-Timer, Repeat Explorer, Detail Planner, Group Organizer)
- MVP scope (6 capabilities: auth/subscription, AI generation, travel styles, customization, enrichment, collaboration)
- Explicit out-of-scope (third-party integrations, mobile apps, booking, offline, payment processing, social features, i18n, MFA, compliance certs)
- Roles & permissions (admin: full CRUD + approve/reject, partner: view + suggest only)
- Subscription limits (basic plan: 1 admin + 1 partner max)

**Destination**: `docs/product-vision.md`

**Markers**: `<!-- PROMOTED:product-vision START/END -->`

**Updates Made**:
- ✅ Expanded MVP scope #2 (AI generation) to explicitly include "well-known landmarks AND lesser-known local points of interest"
- ✅ Clarified enrichment requirement (#5) that all user-provided places MUST be included without omission
- ✅ Added collaboration workflow details (#6) with suggestion status transitions and permission enforcement
- ✅ Expanded out-of-scope section with explicit items: i18n, MFA, OAuth, compliance certs, session management features

---

### Spec 002: Non-Functional Requirements and System Constraints

**Durable Decisions Promoted**:
- Performance targets (NFR-PERF-001 to NFR-PERF-004): p95 ≤ 500ms @ 100 RPS, AI ack ≤ 3s, Core Web Vitals LCP ≤ 2.5s
- Scalability targets (NFR-SCALE-001 to NFR-SCALE-003): 500 concurrent VUs sustained 10 min, stateless backend, 503 w/ Retry-After on overload
- Availability targets (NFR-AVAIL-001 to NFR-AVAIL-004): 99.5% uptime, RPO ≤ 24h, cold-start ≤ 60s, MTTR ≤ 30 min
- Accessibility requirements (NFR-A11Y-001 to NFR-A11Y-004): WCAG 2.1 AA zero violations, keyboard-operable, 4.5:1 contrast, Lighthouse ≥ 90
- Security requirements (NFR-SEC-001 to NFR-SEC-008): zero secrets in Git, TLS 1.2+, zero gosec high findings, zero critical CVEs, prompt injection validation, output sanitization
- Maintainability (NFR-MAINT-001 to NFR-MAINT-005): 80% coverage backend+frontend, zero lint errors, API docs mandatory
- Privacy (NFR-PRIV-001 to NFR-PRIV-003): right-to-deletion, minimal PII, privacy policy
- Observability (NFR-OBS-001 to NFR-OBS-004): 100% structured logs, /healthz < 100ms, correlation IDs, alerting rules

**Destination**: `docs/nfrs.md`

**Markers**: `<!-- PROMOTED:nfrs START/END -->`

**Status**: ✅ Already complete (no changes needed - all NFRs already promoted)

---

### Spec 003: Cloud & Environments Strategy

**Durable Decisions Promoted**:
- Cloud provider: AWS us-east-1, single account, VPC-based isolation
- Environment topology: 2 environments (staging active MVP @ $200/mo, production dormant @ $300-400/mo when provisioned)
- IaC tool: Terraform 1.5+ with HCL, workspaces, remote state (S3 + DynamoDB locking)
- Compute platform: ECS Fargate ARM64 Graviton2 (NOT Lambda - AI workloads need unlimited execution time)
- Frontend delivery: S3 + CloudFront CDN
- CI/CD: GitHub Actions with OIDC federation (no long-lived credentials)
- Secrets: AWS Secrets Manager (runtime retrieval via IAM roles)
- Staging config: db.t4g.micro single-AZ, 0.5 vCPU/1GB tasks, NAT instance, 7-day logs, 7-day backups
- Production config: db.t4g.small Multi-AZ, 1 vCPU/2GB tasks, NAT Gateway, 30-day logs, 30-day backups
- Deployment strategy: Auto to staging on main merge, manual-only to production

**Destination**: `docs/cloud-and-environments.md`

**Markers**: `<!-- PROMOTED:cloud-environments START/END -->`

**Status**: ✅ Already complete (includes comprehensive diagrams and configuration tables)

---

### Spec 004: Security & Authentication/Authorization Model

**Durable Decisions Promoted**:
- Authentication: JWT RS256 with multi-key rotation, 24h access tokens, 30-day refresh tokens, HTTP-only cookies
- Password security: bcrypt cost 12, 8-72 chars with uppercase/lowercase/digit requirements
- JWT key rotation: Zero-downtime via multi-key strategy (new tokens use primary key, validation accepts any active key)
- Password change: Optional "log out all devices" checkbox during password change
- Authorization: RBAC with admin (full CRUD, approve suggestions, invite 1 partner) and partner (view, suggest only) roles
- Concurrency control: Optimistic locking on all admin trip modifications (version/timestamp check, 409 on conflict)
- Subscription limits: 1 admin + max 1 partner per trip (basic plan)
- Secrets management: AWS Secrets Manager, IAM role-based retrieval, zero-downtime rotation support
- PII handling: GDPR-aware (minimal collection, right-to-deletion within 30 days, privacy policy)
- Input validation: Prompt injection detection (instruction override, system prompt extraction, off-topic), SQL injection, XSS, path traversal all rejected with 400
- Output sanitization: HTML/script stripping, executable content filtering before persist/render
- Dependency security: Zero critical/high CVEs, gosec zero high findings, npm audit --audit-level=high pass required
- Security logging: CloudWatch 30-day retention for auth/authz/validation failures, alarms @ 100 auth failures/min or 10 prompt injections/min (manual review, no auto-blocking)

**Destination**: `docs/security.md`

**Markers**: `<!-- PROMOTED:security START/END -->`

**Status**: ✅ Already complete (includes comprehensive validation flow diagrams)

---

### Spec 005: System Architecture & Technology Stack

**Durable Decisions Promoted**:

**To docs/architecture.md**:
- Three-tier architecture: React 19 SPA (S3+CloudFront) → Go 1.24+ REST API (ECS Fargate) → PostgreSQL 15.4 (RDS)
- Backend tech: Chi router, pgx/v5, goose/v3 migrations, Anthropic SDK, slog logging
- Frontend tech: Vite, TanStack Query v5, Zustand, CSS Modules, Lucide React, React Router v7
- Component boundaries: stateless backend, database connection pooling (min 5, max 25 per task), middleware chain (RequestID → Logger → Recovery → CORS → BodySize)
- Integration rules: Backend ↔ DB (goose migrations, transactions for multi-table ops), Backend ↔ AI (prompt validation, output sanitization, streaming, 429 fast-fail), Frontend ↔ Backend (HTTPS, JWT cookies, JSON, X-Request-ID correlation), CI/CD ↔ AWS (OIDC, Terraform state, ECR, rolling updates, circuit breaker rollback)
- Scalability constraints: stateless ECS tasks, auto-scaling 1-5 (staging) / 2-20 (production) @ 70% CPU, load shedding with 503 + Retry-After
- Security boundaries: network isolation (public subnets ALB only, private subnets ECS+RDS), authentication (JWT 24h expiration), authorization (role-based API enforcement), input validation (API + prompt layers), secrets (AWS Secrets Manager, IAM roles)
- Observability: structured JSON logs (100% of requests), /healthz endpoint, correlation IDs, alerting rules

**To docs/data-model.md**:
- Architectural components (not database entities - those are already in data-model.md)
- Backend: HTTPServer, Middleware Chain, DatabaseClient, AIClient, RepositoryPattern, ServicePattern, ErrorHandler
- Frontend: AppRouter, QueryProvider, AuthStore, APIClient, DesignTokens, ComponentLibrary, ErrorBoundary
- Infrastructure: VPCModule, ECSModule, RDSModule, ALBModule, CloudFrontModule, SecretsModule, CICDOutputs

**To constitution** (.specify/memory/constitution.md):
- Mandated technology stack (already present): Go 1.24+, React 19+, PostgreSQL 15.4, AWS us-east-1, ECS Fargate ARM64, Terraform 1.5+
- Infrastructure layer (already present): ECS Fargate for backend (NOT Lambda), S3+CloudFront for frontend, remote state S3+DynamoDB
- Security rules (already present): prompt injection prevention, output sanitization

**Markers**: 
- `<!-- PROMOTED:architecture START/END -->` in docs/architecture.md
- `<!-- PROMOTED:data-model START/END -->` in docs/data-model.md
- `<!-- PROMOTED:mandated-stack START/END -->` in constitution
- `<!-- PROMOTED:security-rules START/END -->` in constitution

**Status**: ✅ Already complete (all architectural decisions already promoted to appropriate destinations)

---

### Spec 006: Core Domain and Data Model Foundations

**Durable Decisions Promoted**:
- Complete 16-entity catalog (User, RefreshToken, JWTSigningKey, SecurityEvent, Plan, Subscription, Trip, Destination, Day, Activity, Collaborator, Suggestion, TravelStyle, TripTravelStyle, ConversationSession, ConversationMessage)
- Three-layer validation taxonomy ([DB], [Logic], [API]) for all entities
- Performance-critical indexes (26 indexes across entities for optimal query patterns)
- Forward-only state transitions (User: Active → Deleted, Subscription: stub_pending → active → cancelled, Trip: draft → published, Suggestion: pending → approved/rejected, ConversationSession: in_progress → completed/abandoned)
- Cascade behavior for all foreign keys (Trip cascades Days/Activities/Collaborators/Suggestions/ConversationSessions, User cascades RefreshToken, etc.)
- Concurrency control strategy (optimistic locking on Trip and Activity with version field, If-Match header, 409 Conflict response)
- 26 invariants and business rules (authentication, authorization, concurrency, data protection, plan limits, GDPR, state transition immutability)
- Database migration sequence (16 migrations in dependency order)
- Destination geographic attributes (latitude/longitude coordinates with spatial index support)

**Destination**: `docs/data-model.md`

**Markers**: `<!-- PROMOTED:data-model START/END -->`

**Updates Made**:
- ✅ Added ConversationSession entity (trip AI conversation tracking: session_id, trip_id, status, token count)
- ✅ Added ConversationMessage entity (individual messages: session_id, role, content, token_count, timestamp)
- ✅ Completed Destination entity with latitude/longitude coordinates and spatial index notes
- ✅ Added three-layer validation rules ([DB], [Logic], [API]) for all 16 entities with 101+ validation tags
- ✅ Documented 26 performance-critical indexes (User: email case-insensitive, Trip: creator_id + status, Day: trip_id + destination_id, Activity: day_id, Collaborator: trip_id + user_id, Suggestion: trip_id + author_id + status, Subscription: user_id UNIQUE + status, ConversationSession: trip_id + status, ConversationMessage: session_id, Destination: country + optional spatial)
- ✅ Added State Transitions sections documenting forward-only flows for User, Subscription, Trip, Suggestion, ConversationSession
- ✅ Added Cascade Behavior sections documenting CASCADE/SET NULL/RESTRICT for all foreign keys
- ✅ Added Concurrency Control sections for Trip and Activity (optimistic locking with version increment)
- ✅ Documented Invariants and Business Rules section with 26 global rules (authentication bcrypt cost 12, JWT RS256 rotation, role enforcement, plan limits, GDPR right-to-deletion, state transition immutability, optimistic locking enforcement, password/token security, PII handling)
- ✅ Added Database Migrations section listing all 16 migrations in dependency order
- ✅ Enhanced .github/copilot-instructions.md with expanded data-model.md description (16 entities, validation layers, indexes, state transitions)

**Status**: ✅ Completed 2026-07-06 (gap-filled all missing content)

---

### Spec 007: API Design Standards and Conventions

**Durable Decisions Promoted**:
- Resource naming: Plural nouns (`/trips`, `/users`), lowercase with hyphens (`/travel-styles`), verbs only for non-CRUD actions (`/trips/:id/generate`)
- URL structure: Maximum 2-level nesting (`/trips/:id/days/:day_id`), top-level endpoints for deep resources (`/activities/:id`), query params for cross-hierarchy filtering
- API versioning: URL path `/api/v1` (major version only), breaking changes require new version, 6-month deprecation period
- HTTP methods: Semantic usage (GET read-only, POST create, PUT full replace, PATCH partial, DELETE remove), idempotent PUT/DELETE
- Request format: `application/json`, `snake_case` fields, ISO 8601 timestamps/dates, validation on all input (400 malformed, 422 validation)
- Response format: Flat JSON for singles, envelope for lists with pagination metadata, empty arrays (not 404)
- Error format: Standardized structure (`{error, message, request_id, fields?}`), 11 machine-readable error codes, field-level validation errors
- HTTP status codes: Semantic usage (200/201/204 success, 400/401/403/404/409/422/429 client errors, 500/503 server errors)
- Pagination: Offset-based (`page`, `per_page` with default 20, max 100), metadata envelope with `has_next`/`has_prev`, page beyond range returns empty array
- Filtering: Query param operators (`[gte]`, `[lte]`, `[like]`, `[in]`), comma-separated OR, text search with `?q=`, validate field names
- Sorting: Minus prefix for descending (`?sort=-created_at`), comma-separated multi-field, left-to-right priority
- Rate limiting: 100 req/min authenticated (per user ID), 10 req/min unauthenticated (per IP), headers in all responses (`X-RateLimit-*`), 429 with `Retry-After`
- Timestamp format: ISO 8601 with UTC (`2026-07-06T10:00:00Z`), consistent across all endpoints
- Endpoint patterns: 7 reusable templates (List, Get Single, Create, Update Full, Update Partial, Delete, Action)

**Destination**: `docs/api-design-standards.md`

**Markers**: `<!-- PROMOTED:api-standards START/END -->`

**Updates Made**:
- ✅ Created docs/api-design-standards.md with complete 15-section standards document
- ✅ Added comprehensive examples (✅ DO / ❌ DON'T) for every convention
- ✅ Documented 11 machine-readable error codes with HTTP status mapping
- ✅ Defined 7 endpoint patterns with request/response examples
- ✅ Added compliance section (code review checklist, integration tests, linting, exception process)
- ✅ Updated .github/copilot-instructions.md to reference api-design-standards.md with full description (resource naming, URL structure, versioning, formats, error handling, pagination, filtering, sorting, rate limiting, endpoint patterns)

**Status**: ✅ Completed 2026-07-06

---

## Constitution Updates

**File**: `.specify/memory/constitution.md`

**Sections Verified**:
1. ✅ **Technology Stack → Application Layer**: Mandated Go 1.24+, React 19+, PostgreSQL 15.4, Anthropic Claude API
2. ✅ **Technology Stack → Infrastructure Layer**: Mandated AWS us-east-1, Terraform 1.5+, ECS Fargate ARM64 (NOT Lambda), ECR, S3+CloudFront, ALB, VPC isolation, Secrets Manager, CloudWatch, GitHub Actions OIDC
3. ✅ **Environment Strategy**: Staging active MVP ($200), Production dormant ($300-400), separate .tfvars, auto-deploy staging, manual-only production
4. ✅ **Security Rules (V. Secure Configuration)**: Prompt injection prevention (non-negotiable), output sanitization (non-negotiable), authorization enforcement (admin/partner roles)

**Version**: Constitution version remains 1.3.0 (no new amendments needed - all architectural decisions align with existing principles)

---

## Documentation References Update

**File**: `.github/copilot-instructions.md`

**Section**: `<!-- PROMOTED:doc-references START/END -->`

**Status**: ✅ Updated 2026-07-06 (added api-design-standards.md)

**Referenced Docs**:
1. docs/product-vision.md — product identity, personas, MVP scope, out-of-scope, roles/permissions
2. docs/functional-requirements.md — normative requirements, in-scope / out-of-scope for MVP
3. docs/nfrs.md — measurable NFRs with validation methods (performance, scalability, availability, accessibility, security, maintainability, privacy, observability)
4. docs/architecture.md — component boundaries, integration rules, scalability constraints, security boundaries, observability strategy
5. docs/cloud-and-environments.md — AWS us-east-1, staging/production topology, Terraform IaC, ECS Fargate, CI/CD OIDC, secrets management, cost strategy
6. docs/data-model.md — 16 core entities with attributes, three-layer validation, indexes, state transitions, cascade behavior, invariants, business rules, migrations
7. docs/security.md — JWT RS256 multi-key rotation, admin/partner roles, subscription limits, bcrypt password security, optimistic locking, secrets management, PII handling, prompt injection prevention, output sanitization, dependency security, logging & monitoring
8. docs/api-design-standards.md — API conventions (resource naming, URL structure, versioning, HTTP methods, request/response formats, error handling, status codes, pagination, filtering, sorting, rate limiting, endpoint patterns)
9. docs/coding-guidelines.md — formatting, import organization, naming conventions, KISS/DRY for Go and React/TypeScript
10. docs/testing-guidelines.md — three-layer testing (unit, integration, E2E), folder structure, naming, coverage targets
11. docs/ui-guidelines.md — design tokens, Atomic Design, responsive breakpoints, WCAG 2.1 AA, loading/error/empty states

---

## Validation Results

### ✅ Completeness Check

All 7 foundational specs have been analyzed and promoted:
- ✅ Spec 001: Product vision, personas, MVP scope, out-of-scope, roles
- ✅ Spec 002: All 31 NFRs with measurable targets and validation methods
- ✅ Spec 003: Cloud provider, environments, IaC, compute, CI/CD, secrets, costs
- ✅ Spec 004: Authentication, authorization, concurrency, security, PII, validation, sanitization, logging
- ✅ Spec 005: Architecture, tech stack, components, integration rules, scalability, observability
- ✅ Spec 006: Complete entity catalog, validation layers, indexes, state transitions, cascade behavior, concurrency control, business rules
- ✅ Spec 007: API design standards covering resource naming, URL structure, versioning, HTTP methods, request/response formats, error handling, status codes, pagination, filtering, sorting, rate limiting

### ✅ Constitution Compliance

All promoted decisions align with existing constitution principles:
- ✅ Test-First Development (comprehensive testing strategy defined in NFRs)
- ✅ Simplicity (industry-standard patterns: domain-driven backend, Atomic Design frontend, modular Terraform)
- ✅ Code Quality (automated lint/test gates in NFRs, 80% coverage targets)
- ✅ Accessible UI (WCAG 2.1 AA non-negotiable, design tokens enforced)
- ✅ Secure Configuration (secrets in AWS Secrets Manager, prompt injection prevention, output sanitization)

### ✅ Marker Consistency

All promoted sections use consistent markers:
- Format: `<!-- PROMOTED:<topic> START -->` and `<!-- PROMOTED:<topic> END -->`
- Last updated date included: `<!-- Last promoted: 2026-07-03 -->`
- Source attribution included where applicable

### ✅ No Gaps Detected

All durable decisions from foundational specs have been promoted. No open ambiguities or unresolved decisions flagged for user resolution.

---

## Impact on Future Work

### Automatic Inheritance

All subsequent work (new specs, `/speckit.plan`, `/speckit.tasks`, implementation) will now automatically inherit:

1. **Product boundaries**: MVP scope and explicit out-of-scope constraints prevent scope creep
2. **Quality gates**: NFR targets automatically enforced in CI/CD (coverage, lint, security scans)
3. **Technology stack**: No re-decisions on languages, frameworks, or infrastructure platforms
4. **Security model**: Consistent authentication, authorization, validation, and sanitization across all features
5. **Cost strategy**: Environment-specific resource configurations prevent budget overruns
6. **Architecture patterns**: Proven patterns for backend (handler/service/repository), frontend (Atomic Design), infrastructure (modular Terraform)
7. **API design conventions**: Consistent endpoint patterns (naming, structure, versioning, error handling, pagination) across all backend APIs

### Re-Promotion Trigger

Re-run this promotion when:
- ✅ Any foundational spec (001-007) is revised or refined
- ✅ Constitution is amended with new non-negotiable principles
- ✅ New foundation-level specs are created (e.g., 008-observability-detailed)

### Maintenance

- Promoted sections are **idempotent**: Re-running promotion updates marked sections in place without duplication
- Human edits outside marked sections are preserved
- Marker format enables automated detection and differential updates

---

## Recommendation
7 are now in persistent context.

**Next Steps**:
1. Review this report and verify all promoted decisions align with team understanding
2. If approved, commit all changes (docs/, .github/copilot-instructions.md)
3. Use `/commit-and-push` with conventional commit message: `docs: promote API design standards (spec 007) to persistent context`
4. Begin feature development knowing all architectural, security, quality, domain model, and API conventions are inherited automatically
5. Review [.github/ROADMAP-RECONCILIATION-REPORT.md](.github/ROADMAP-RECONCILIATION-REPORT.md) for spec 006 task integration (77 tasks added to docs/roadmap.md)

---

## Files Modified

| File | Status | Changes |
|------|--------|---------|
| `.specify/memory/constitution.md` | ✅ Verified | No changes needed (already complete) |
| `docs/product-vision.md` | ✅ Updated | Expanded MVP scope details, enhanced out-of-scope section |
| `docs/nfrs.md` | ✅ Verified | No changes needed (already complete) |
| `docs/cloud-and-environments.md` | ✅ Verified | No changes needed (already complete) |
| `docs/security.md` | ✅ Verified | No changes needed (already complete) |
| `docs/architecture.md` | ✅ Verified | No changes needed (already complete) |
| `docs/data-model.md` | ✅ Updated | Gap-filled with spec 006 (ConversationSession, ConversationMessage, validation layers, indexes, state transitions, cascade behavior) |
| `.github/copilot-instructions.md` | ✅ Updated | Enhanced data-model.md description with 16 entities and validation layers |
| `docs/api-design-standards.md` | ✅ Created | Complete API standards (resource naming, URL structure, versioning, HTTP methods, request/response formats, error handling, pagination, filtering, sorting, rate limiting) |
| `.github/copilot-instructions.md` | ✅ Updated | Enhanced data-model.md description with 16 entities and validation layers; added api-design-standards.md reference with full description |
| `.github/PROMOTION-REPORT.md` | ✅ Updated | Added spec 007 section (this report document) |

**Total Files Modified**: 4 (data-model.md updated 2026-07-06, api-design-standards.md created 2026-07-06, copilot-instructions.md updated 2026-07-06, PROMOTION-REPORT.md updated 2026-07-06)  
**Total Files Verified**: 6 (constitution, nfrs.md, cloud-and-environments.md, security.md, architecture.md, product-vision.md already complete)

---

**Generated by**: `/promote-fundations` workflow  
**Execution Date**: 2026-07-03 (Updated: 2026-07-06 with specs 006 and 007)  
**Next Promotion**: When foundational specs 001-007
---

## Related Reports

- [ROADMAP-RECONCILIATION-REPORT.md](.github/ROADMAP-RECONCILIATION-REPORT.md) - Documents integration of spec 006 tasks into docs/roadmap.md (77 tasks added, 539 total)
