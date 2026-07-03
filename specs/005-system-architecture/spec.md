# Feature Specification: System Architecture and Technology Stack

**Feature Branch**: `005-system-architecture`

**Created**: 2026-07-03

**Status**: Draft

**Input**: Define the architecture, technology alignments, and foundations for frontend, backend, and IaC. As an engineering team, we want a coherent architecture that satisfies the product scope, the non-functional requirements, the cloud strategy, and the security foundations already defined. Specify the chosen technologies and their rationale, the high-level component structure for frontend and backend, integration boundaries, and how IaC realizes the architecture. This must align with and reference the prior foundation specs — do not define specific endpoints or UI screens yet.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Backend Service Architecture (Priority: P1)

As a backend engineer, I need a clear, modular backend architecture with defined domain boundaries, middleware patterns, and integration points so that I can implement features consistently, maintain code quality, and ensure all NFRs (performance, security, observability) are enforced systematically.

**Why this priority**: The backend service structure is foundational. All API endpoints, business logic, database access, AI integration, and security controls depend on this architecture. Without a coherent backend structure, implementation will be inconsistent and NFR compliance will be ad-hoc.

**Independent Test**: Can be validated by creating a skeleton backend project with the defined structure, implementing one sample domain (e.g., authentication), and verifying that middleware, routing, logging, error handling, and database access patterns work as specified. Success means a new engineer can implement a new domain following established patterns without architectural decisions.

**Acceptance Scenarios**:

1. **Given** a new feature requires a new domain (e.g., "trips"), **When** a backend engineer creates the domain structure following the architecture, **Then** the code is organized under `internal/trips/` with clear separation of handler, service, repository, and model layers
2. **Given** a new API endpoint is added to any domain, **When** the endpoint is invoked, **Then** request ID middleware, logging middleware, and panic recovery middleware are automatically applied
3. **Given** a backend service needs to access the database, **When** the service layer calls a repository method, **Then** the repository uses connection pooling from a shared database client and returns domain errors (not raw SQL errors)
4. **Given** a backend service integrates with an external API (AI provider), **When** the service makes a request, **Then** it uses a centralized HTTP client with configured timeouts, retry logic, and observability instrumentation

---

### User Story 2 - Frontend Application Structure (Priority: P1)

As a frontend engineer, I need a well-defined React application architecture with component layering, state management patterns, API client conventions, and styling guidelines so that I can build accessible, performant, and maintainable UI features that meet WCAG 2.1 AA standards and align with design system tokens.

**Why this priority**: Frontend structure determines developer productivity, bundle size, accessibility compliance, and user experience consistency. Without clear patterns for component composition, state management, and API interaction, the codebase will fragment and accessibility/performance will degrade.

**Independent Test**: Can be validated by implementing one complete user flow (e.g., login → dashboard) using the defined architecture, and verifying that components follow Atomic Design layering, state is managed consistently, API calls use TanStack Query, CSS uses tokens, and Lighthouse/axe-core audits pass. Success means a new UI feature can be built following established patterns without reinventing structure.

**Acceptance Scenarios**:

1. **Given** a new UI feature is being implemented, **When** the engineer structures components, **Then** primitives are in `src/components/primitives/`, composites in `src/components/composites/`, and feature-specific components in `src/features/<feature-name>/components/`
2. **Given** a component needs to fetch data from the backend, **When** the component mounts, **Then** it uses TanStack Query hooks (e.g., `useQuery`, `useMutation`) with defined query keys and error handling
3. **Given** a component needs global state (e.g., authentication status), **When** the component accesses state, **Then** it uses a Zustand store with typed selectors and does not prop-drill authentication state through component trees
4. **Given** a component needs styling, **When** the engineer writes CSS, **Then** all colors, spacing, and typography reference CSS custom property tokens from `src/styles/tokens.css` (no hard-coded values)
5. **Given** a page is rendered, **When** automated accessibility tests run, **Then** axe-core reports zero WCAG 2.1 AA violations and Lighthouse accessibility score is ≥ 90

---

### User Story 3 - Infrastructure as Code Foundations (Priority: P1)

As an infrastructure engineer, I need a structured Terraform codebase with modular design, environment separation, state management, and CI/CD integration so that I can provision AWS resources reliably, manage staging and production environments independently, and support zero-downtime deployments.

**Why this priority**: IaC is the bridge from architecture diagrams to running systems. Without a coherent Terraform structure, infrastructure will drift, environments will be inconsistent, and deployments will be brittle.

**Independent Test**: Can be validated by running `terraform plan` for staging environment and verifying that all required resources (VPC, ECS cluster, RDS, ALB, CloudFront, S3, Secrets Manager) are defined, output variables are exported for CI/CD, and remote state locking works. Success means infrastructure can be provisioned, updated, and destroyed idempotently without manual steps.

**Acceptance Scenarios**:

1. **Given** staging infrastructure needs to be provisioned, **When** `terraform apply` is run with staging workspace, **Then** all AWS resources are created in the `10.0.0.0/16` VPC with cost-optimized configuration (ECS: 0.25 vCPU/0.5 GB RAM, db.t4g.micro, single-AZ, NAT instance)
2. **Given** production infrastructure is defined but dormant, **When** `terraform plan` is run with production workspace, **Then** the plan shows resources with production configuration (ECS: 1 vCPU/2 GB RAM, db.t4g.small, Multi-AZ, NAT Gateway) but does not create them until explicitly applied
3. **Given** Terraform state is persisted remotely, **When** multiple engineers work on infrastructure, **Then** state locking via DynamoDB prevents concurrent modifications
4. **Given** backend container image is pushed to ECR, **When** GitHub Actions deploys to staging, **Then** the ECS service is updated with the new task definition and performs a rolling update with health checks

---

### User Story 4 - Integration and Error Handling Patterns (Priority: P2)

As any engineer, I need standardized patterns for API integration, error propagation, logging correlation, and graceful degradation so that debugging is efficient, error messages are actionable, and system failures are observable and recoverable.

**Why this priority**: Integration boundaries (frontend ↔ backend, backend ↔ database, backend ↔ AI) are where most runtime errors occur. Consistent error handling and observability patterns are required to meet NFR-OBS and NFR-AVAIL targets.

**Independent Test**: Can be tested by simulating failure scenarios (database timeout, AI API 500 error, malformed client request) and verifying that errors are logged with correlation IDs, client receives structured error responses (not stack traces), and retries/fallbacks work as designed.

**Acceptance Scenarios**:

1. **Given** a backend API request fails due to database timeout, **When** the error occurs, **Then** a structured log entry is emitted with correlation ID, error type, and stack trace, and the client receives `500 Internal Server Error` with correlation ID in response body
2. **Given** a frontend API call fails with 401 Unauthorized, **When** the error is handled, **Then** the user is redirected to login page and the failed request is NOT retried automatically
3. **Given** the AI provider returns a 503 Service Unavailable, **When** the backend receives the error, **Then** it returns `503 Service Unavailable` with `Retry-After: 60` header to the client and logs the AI provider outage
4. **Given** a user submits invalid input (e.g., malformed email), **When** validation fails, **Then** the client receives `400 Bad Request` with a structured error response containing field-level validation messages

---

### Edge Cases

- **What happens when the database connection pool is exhausted?** Backend rejects new requests with `503 Service Unavailable` until connections are released; logs connection pool metrics.
- **How does the system handle AI provider rate limits?** Backend respects `Retry-After` headers, implements exponential backoff, and fails fast with `503` if the AI provider is unavailable (does not queue requests indefinitely).
- **What happens when Terraform state locking fails?** The `terraform apply` command aborts with a locking error; engineer must manually resolve lock contention (force-unlock only in documented emergency scenarios).
- **How does frontend handle stale data during deployments?** CloudFront cache is invalidated on deployment; React app uses TanStack Query's `staleTime` to refetch data when staleness thresholds are exceeded.
- **What happens when ECS tasks fail health checks?** ECS marks the task as unhealthy and starts a replacement task; ALB stops routing traffic to the failing task; logs and CloudWatch alarms capture the incident.

## Requirements *(mandatory)*

### Functional Requirements

#### Backend Architecture

- **FR-001**: Backend MUST be organized using domain-driven directory structure with `cmd/`, `internal/<domain>/`, `pkg/`, and `config/` top-level directories
- **FR-002**: Each domain MUST have clear separation between HTTP handlers (presentation), service layer (business logic), repository layer (data access), and models (domain types)
- **FR-003**: Middleware MUST be applied globally for: request ID generation, structured logging, panic recovery, CORS, and request/response size limits
- **FR-004**: Database access MUST use connection pooling with configurable min/max connections and connection lifetime limits; MVP default: min 5, max 25 connections per ECS task
- **FR-005**: All errors from repository layer MUST be wrapped with context before propagation to service layer
- **FR-006**: Service layer MUST NOT depend on HTTP-specific types (e.g., `http.Request`, `http.ResponseWriter`); all domain logic MUST be testable without HTTP infrastructure
- **FR-007**: External API clients (AI provider, future integrations) MUST be abstracted behind Go interfaces to enable mocking in tests
- **FR-008**: Configuration MUST be loaded from environment variables with validation on startup; the application MUST NOT start if required configuration is missing

#### Frontend Architecture

- **FR-009**: React application MUST follow Atomic Design component layering: `primitives/` → `composites/` → `features/<name>/components/`
- **FR-010**: All API communication MUST use TanStack Query with native fetch API; a thin wrapper MUST provide base URL, default headers, and error normalization
- **FR-011**: Global authentication state MUST be managed in a Zustand store; local component state MUST use React's `useState`/`useReducer`
- **FR-012**: All styling MUST use CSS Modules with token-based theming; no inline styles or hard-coded color/spacing values
- **FR-013**: Icons MUST come from a single library (Lucide React) project-wide; no mixed icon sources
- **FR-014**: Routes MUST be defined in a centralized routing configuration using React Router v7
- **FR-015**: All form inputs MUST have associated labels for accessibility; all interactive elements MUST be keyboard-operable
- **FR-016**: Error boundaries MUST catch and display errors at feature level; critical errors MUST provide recovery actions (e.g., retry, navigate home)

#### Infrastructure as Code

- **FR-017**: Terraform codebase MUST be organized into reusable modules: `modules/vpc/`, `modules/ecs/`, `modules/rds/`, `modules/alb/`, `modules/cloudfront/`, `modules/secrets/`
- **FR-018**: Environment-specific configurations MUST use separate `.tfvars` files (staging.tfvars, production.tfvars) with single default workspace; environment differences MUST be explicit and reviewable in version control
- **FR-019**: Remote state MUST be stored in S3 with DynamoDB locking; state bucket MUST have versioning enabled
- **FR-020**: All sensitive outputs (database passwords, API keys) MUST be marked as `sensitive = true` in Terraform
- **FR-021**: ECS task definitions MUST support blue/green deployments with health check grace periods and automatic rollback on failure; auto-scaling MUST be configured with CPU-based target tracking (70% target) and environment-specific bounds (staging: min 1, max 2 tasks)
- **FR-022**: CloudFront distribution MUST enforce HTTPS-only access with TLS 1.2+ and serve from S3 origin with OAI (Origin Access Identity)
- **FR-023**: All AWS resources MUST be tagged with: `Environment` (staging/production), `ManagedBy` (terraform), `Project` (traveler)

#### Integration Boundaries

- **FR-024**: Frontend ↔ Backend communication MUST use RESTful HTTP with JSON payloads; all requests MUST include CORS-compliant headers
- **FR-025**: Backend ↔ Database communication MUST use pgx connection pooling with prepared statements to prevent SQL injection
- **FR-026**: Backend ↔ AI provider communication MUST use HTTPS with configurable request timeouts (default: 60s for non-streaming, infinite for streaming)
- **FR-027**: All cross-service error responses MUST include correlation ID (`X-Request-ID` header) for traceability
- **FR-028**: Frontend MUST handle all HTTP error codes gracefully: 400 (show validation errors), 401 (redirect to login), 403 (show permission error), 404 (show not found), 500/503 (show retry option)

### Key Entities *(architectural components)*

#### Backend Core Components

- **HTTPServer**: Chi router-based HTTP server with middleware chain, health check endpoint, graceful shutdown
- **Middleware**: Request ID generator, structured logger (slog), panic recovery, CORS, body size limiter
- **DatabaseClient**: pgx connection pool with health checks, prepared statement cache, connection lifecycle management
- **AIClient**: Anthropic SDK wrapper with retry logic, timeout configuration, prompt validation, and response sanitization
- **RepositoryPattern**: Generic interface for data access with methods like `Create`, `FindByID`, `Update`, `Delete`, `List` (domain-specific implementations)
- **ServicePattern**: Business logic layer with dependency injection of repositories and external clients
- **ErrorHandler**: Centralized error-to-HTTP status mapping with structured error responses

#### Frontend Core Components

- **AppRouter**: React Router v7 configuration with route definitions, protected routes, lazy loading
- **QueryProvider**: TanStack Query provider with global defaults (staleTime, cacheTime, retry policy)
- **AuthStore**: Zustand store for authentication state with typed actions (`login`, `logout`, `refreshSession`)
- **APIClient**: Thin wrapper around native fetch API with base URL, default headers, response/error normalization, and auth token injection
- **DesignTokens**: CSS custom properties for colors, spacing, typography, border-radius, shadows
- **ComponentLibrary**: Reusable primitives (Button, Input, Card, Modal) and composites (Form, DataTable, NavigationBar)
- **ErrorBoundary**: React error boundary with fallback UI and error reporting

#### Infrastructure Core Modules

- **VPCModule**: Creates VPC with public/private subnets across 2 AZs, route tables, internet gateway, NAT gateway/instance
- **ECSModule**: Creates ECS cluster, task definition with environment-specific resource allocation (staging: 0.25 vCPU/0.5 GB RAM ARM64, production: 1 vCPU/2 GB RAM ARM64), service with auto-scaling policies (CPU target tracking at 70%, staging: min 1/max 2 tasks), CloudWatch log group with retention (staging: 7 days, production: 30 days)
- **RDSModule**: Creates PostgreSQL RDS instance, subnet group, security group, parameter group, automated backups with retention (staging: 1 day, production: 30 days)
- **ALBModule**: Creates Application Load Balancer, target group, listener rules, health check configuration
- **CloudFrontModule**: Creates CloudFront distribution, S3 bucket for static assets, OAI, SSL certificate
- **SecretsModule**: Creates AWS Secrets Manager secrets for database credentials, AI API keys, JWT signing keys
- **CICDOutputs**: Terraform outputs for GitHub Actions (ECR repository URL, ECS cluster name, S3 bucket name, CloudFront distribution ID)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new backend domain can be scaffolded and integrated by a backend engineer in under 2 hours following documented patterns (includes handler, service, repository, unit tests, integration tests)
- **SC-002**: A new React feature can be built by a frontend engineer with zero hard-coded color/spacing values; 100% of styles reference design tokens
- **SC-003**: Terraform apply for staging environment completes successfully in under 15 minutes on first run; subsequent applies with no infrastructure changes complete in under 2 minutes (plan-only)
- **SC-004**: All NFR-A11Y targets are met (axe-core zero violations, Lighthouse accessibility ≥ 90) without manual remediation; accessibility compliance is enforced by automated tests in CI
- **SC-005**: Backend service starts and reaches healthy state (responds to `/healthz` with `200 OK`) within 60 seconds of container launch
- **SC-006**: Frontend production build size is under 500 KB gzipped (excluding code-split chunks); initial page load achieves LCP ≤ 2.5s on simulated 4G connection
- **SC-007**: 95% of API errors result in structured log entries with correlation IDs that can be traced from frontend through backend to database/AI provider
- **SC-008**: Zero secrets are detected in Git history or CI/CD logs; all credentials are loaded from environment variables or AWS Secrets Manager

## Assumptions

- **Architecture stability**: The three-tier architecture (React SPA + Go API + PostgreSQL) and AWS platform choice are fixed for MVP; no alternative platforms or architectures will be evaluated during this phase
- **Single-region deployment**: All infrastructure is deployed to `us-east-1` only; multi-region or global distribution is deferred to post-MVP
- **Monolithic backend initially**: The Go backend is a single monolithic service (not microservices); domain separation is logical (directory structure) not physical (separate deployments)
- **No BFF pattern**: The backend serves the React frontend directly; there is no Backend-For-Frontend layer or GraphQL gateway
- **CloudFront as CDN**: CloudFront is the CDN for static assets; no third-party CDN providers (Cloudflare, Fastly) are in scope
- **Manual production deployments**: Staging deployments are automated on merge to main; production deployments require manual approval and are not part of automated CD pipeline for MVP
- **No service mesh**: ECS tasks communicate with RDS and external APIs directly; no Istio, Linkerd, or other service mesh for MVP
- **JWT in HTTP-only cookies**: Access tokens are stored in HTTP-only cookies (not localStorage) to prevent XSS attacks; frontend does not handle token storage or refresh logic directly
- **Database migrations run on deployment**: Database schema migrations (via goose) are executed as part of the ECS task startup using advisory lock mechanism to ensure only one task runs migrations during concurrent rolling deployments; no separate migration job or pre-deployment migration step
- **Production configuration validation**: Production infrastructure sizing (ECS resources, RDS instance class, auto-scaling bounds) will be validated via load testing staging with production-like scenarios; production .tfvars will be adjusted based on observed CPU/memory/query metrics before provisioning

## Clarifications

### Session 2026-07-03

- Q: How should database migrations handle concurrent execution when multiple ECS tasks start during a rolling deployment? → A: Use goose's advisory lock mechanism (lock table); all tasks attempt migration but only one proceeds; others wait and skip
- Q: Which environment separation strategy should be used for staging and production? → A: Separate .tfvars files (staging.tfvars, production.tfvars) with single default workspace
- Q: Which HTTP client library should the frontend use for API requests? → A: Native fetch API with a thin wrapper for base URL, headers, and error handling
- Q: What should be the default database connection pool sizing for MVP (staging environment supporting < 50 concurrent users)? → A: Min 5, Max 25 connections per ECS task (balanced for MVP scale)
- Q: What should be the ECS auto-scaling configuration for staging environment? → A: Min 1, Max 2 tasks with 70% CPU target (minimal cost for MVP)
- Q: What should be the ECS task resource allocation for staging environment? → A: 0.25 vCPU / 0.5 GB RAM ARM64 (cheapest Fargate tier, ~$9/month per task)
- Q: What should be the CloudWatch log retention period for staging environment? → A: 7 days retention (reasonable debugging window, cost-balanced)
- Q: What should be the RDS automated backup retention period for staging environment? → A: 1 day retention (minimum viable, cheapest backup storage cost)
- Q: How should production infrastructure configuration be validated and tuned before provisioning? → A: Load test staging with production-like scenarios; adjust production .tfvars based on observed metrics before provisioning
