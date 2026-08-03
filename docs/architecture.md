# System Architecture

<!-- PROMOTED:architecture START -->
<!-- Generated from specs/001-product-vision-scope/spec.md, specs/002-nfr-system-constraints/spec.md, and specs/003-cloud-env-strategy/spec.md -->
<!-- Last promoted: 2026-07-03 -->

## Overview

TrAIveler is a three-tier web application with a stateless RESTful API backend, a React single-page application frontend, and AWS-managed infrastructure for compute, storage, and content delivery.

## Component Architecture

```mermaid
%%{init: {'flowchart': {'curve': 'linear'}}}%%
graph TD
    User[End User Browser]
    CF[CloudFront CDN<br/>React build artifacts from S3<br/>Global edge caching, HTTPS]
    ALB[Application Load Balancer<br/>HTTPS termination<br/>Health checks → /healthz<br/>Connection draining]
    ECS[ECS Fargate Cluster<br/>Go 1.26+ RESTful API<br/>ARM64 Graviton2 containers<br/>Auto-scaling 1-5 tasks staging<br/>Stateless horizontal scaling]
    RDS[Amazon RDS<br/>PostgreSQL 15.18<br/>Multi-AZ prod<br/>Single-AZ stage]
    AI[Anthropic AI API<br/>Claude external<br/>Itinerary generation<br/>Multi-turn conversation]
    Secrets[AWS Secrets Manager<br/>DB passwords<br/>API keys<br/>JWT secrets]

    User -->|HTTPS| CF
    User -->|API Requests HTTPS| ALB
    ALB -->|HTTP| ECS
    ECS -->|PostgreSQL| RDS
    ECS -->|HTTPS| AI
    ECS -.->|Retrieves secrets at runtime| Secrets

    style User fill:#e1f5ff
    style CF fill:#fff4e6
    style ALB fill:#e8f5e9
    style ECS fill:#f3e5f5
    style RDS fill:#fce4ec
    style AI fill:#fff9c4
    style Secrets fill:#ffebee
```

## Component Boundaries

### Frontend (React 19+ SPA)

**Technology**:
- React 19 with TypeScript (strict mode)
- Vite build tool
- TanStack Query for data fetching
- Zustand for state management
- CSS Modules + CSS custom properties for styling
- Lucide React for icons

**Responsibilities**:
- Render all user-facing UI
- Handle user authentication state (JWT stored in HTTP-only cookies)
- Make HTTP requests to backend REST API
- Client-side routing (React Router v7)
- Form validation and error handling
- Accessibility compliance (WCAG 2.1 AA)

**Deployment**:
- Static build artifacts (`dist/`) synced to S3
- Served globally via CloudFront CDN
- Invalidation on deployment for cache busting

**Entry Points**:
- `/` - Landing page
- `/login`, `/register` - Authentication
- `/dashboard` - Trip list
- `/trips/:id` - Trip detail view
- `/generate` - AI itinerary generation

### Backend (Go 1.24+ REST API)

**Technology**:
- Go 1.24 or higher
- Chi router for HTTP routing
- pgx/v5 for PostgreSQL database access
- goose/v3 for database migrations
- Anthropic SDK for AI integration
- slog for structured JSON logging

**Responsibilities**:
- Authenticate users (JWT tokens)
- Enforce authorization rules (admin vs partner roles)
- CRUD operations for trips, destinations, days, activities
- AI conversation management (multi-turn input flow)
- Itinerary generation via Claude API
- Collaboration features (suggestions, approvals)
- Prompt validation (injection prevention)
- Output sanitization (XSS prevention)
- Structured logging with correlation IDs
- Health check endpoint (`/healthz`)

**Deployment**:
- Docker containers on ECS Fargate
- ARM64 Graviton2 for cost efficiency
- Auto-scaling based on CPU utilization (70% target)
- Rolling updates with health checks
- Automatic rollback on failures

**API Endpoints** (see `specs/*/contracts/api.md` for full schemas):
- `POST /auth/register`, `/auth/login`, `/auth/logout`
- `GET /auth/me`
- `GET /subscription/plans`, `POST /subscription/checkout`, `GET /subscription/current`
- `GET /trips`, `POST /trips`, `GET /trips/:id`, `PUT /trips/:id`, `DELETE /trips/:id`
- `POST /trips/:id/conversation`, `GET /trips/:id/conversation`
- `POST /trips/:id/collaborators`, `DELETE /trips/:id/collaborators/:userId`
- `GET /suggestions`, `POST /suggestions`, `PATCH /suggestions/:id`
- `GET /healthz`

### Database (PostgreSQL 15.18 on Amazon RDS)

**Schema** (8 core tables):
1. `users` - User accounts (email, password hash, subscription_id)
2. `subscription_plans` - Plan definitions (name, collaborator limit, feature flags)
3. `subscriptions` - Active subscriptions (user_id, plan_id, status)
4. `trips` - Trip metadata (title, start/end dates, status, owner_id)
5. `destinations` - Destinations within trips (city, country, order)
6. `days` - Daily itinerary structure (date, destination_id)
7. `activities` - Activities within days (time, title, description, type, order)
8. `trip_collaborators` - Many-to-many for shared trips (trip_id, user_id, role)
9. `conversation_sessions` - AI conversation context (trip_id, status)
10. `conversation_messages` - Message history (session_id, role, content, timestamp)
11. `suggestions` - Partner modification proposals (trip_id, target_type, target_id, content, status)

**Relationships**:
- User → Subscription (1:1)
- User → Trips (1:N as owner)
- Trip → Destinations (1:N)
- Destination → Days (1:N)
- Day → Activities (1:N)
- Trip ↔ Users (M:N via trip_collaborators)
- Trip → ConversationSession (1:N)
- ConversationSession → ConversationMessages (1:N)
- Trip → Suggestions (1:N)

**Configuration**:
- Staging: db.t4g.micro, single-AZ, 7-day backups
- Production: db.t4g.small, Multi-AZ, 30-day backups
- Connection pooling configured in backend
- Secrets retrieved from AWS Secrets Manager at runtime

### External Dependencies

**Anthropic Claude API**:
- Multi-turn conversational AI for itinerary generation
- Streaming responses for real-time feedback
- Tool use for structured itinerary extraction
- Connection pooling from backend for efficiency

**AWS Services**:
- **ECS Fargate**: Container orchestration
- **RDS**: Managed PostgreSQL database
- **S3**: Static asset storage
- **CloudFront**: CDN for global delivery
- **ALB**: Load balancing and HTTPS termination
- **Secrets Manager**: Credential storage
- **CloudWatch**: Logging and monitoring
- **IAM**: Access control and service roles

## Integration Rules

### Backend ↔ Database

- **Connection Pooling**: pgxpool with configurable min/max connections
- **Migrations**: goose for versioned schema changes
- **Transactions**: Used for all multi-table operations to ensure consistency
- **Error Handling**: All SQL errors wrapped with context before returning

### Backend ↔ Anthropic AI

- **Authentication**: API key retrieved from AWS Secrets Manager
- **Request Format**: Multi-turn conversation with system prompt
- **Response Handling**: Streaming for real-time feedback; tool-use parsing for structured data
- **Error Handling**: Rate limits (429) fast-fail with user-facing message; other errors logged with correlation ID
- **Validation**: Prompt validation layer runs before every AI request (injection prevention)
- **Sanitization**: Output sanitization runs on all AI responses before storage or rendering

### Frontend ↔ Backend

- **Protocol**: HTTPS only (HTTP redirects to HTTPS)
- **Authentication**: JWT tokens in HTTP-only cookies
- **Request Format**: JSON request bodies, standard HTTP methods (GET, POST, PUT, PATCH, DELETE)
- **Response Format**: JSON with consistent error structure
- **Correlation**: X-Request-ID header on all requests/responses for tracing
- **Error Handling**: Typed error responses with actionable messages

### CI/CD ↔ AWS

- **Authentication**: OIDC federation (GitHub Actions assumes IAM role)
- **Infrastructure**: Terraform state in S3 with DynamoDB locking
- **Container Registry**: ECR for Docker images
- **Deployment**: Rolling updates via ECS service updates
- **Rollback**: Automatic via deployment circuit breaker on health check failures

## Scalability Constraints

### Stateless Backend

- All ECS tasks are stateless - no in-process shared state
- Sessions stored in database or client-side (JWT)
- Horizontal scaling supported via ECS auto-scaling

### Database Connection Pooling

- Configured pool size prevents connection exhaustion under load
- Health checks ensure connections released properly

### Auto-Scaling

- ECS tasks scale based on CPU utilization (70% target)
- Limits: 1-5 tasks (staging), 2-20 tasks (production)
- Prevents runaway costs from unexpected spikes

### Load Shedding

- Returns `503 Service Unavailable` with `Retry-After` header when capacity exceeded
- Protects backend from overload and database connection exhaustion

## Security Boundaries

### Network Isolation

- **Public Subnets**: ALB only (internet-facing)
- **Private Subnets**: ECS tasks, RDS (no direct internet access)
- **Security Groups**: Layered (alb-sg → ecs-sg → rds-sg)
- **NACLs**: Environment VPCs isolated (10.0.0.0/16 staging, 10.1.0.0/16 production)

### Authentication & Authorization

- **Authentication**: JWT tokens with 24-hour expiration
- **Authorization**: Role-based (admin vs partner) enforced at API layer
- **Session Management**: HTTP-only cookies prevent XSS token theft

### Input Validation

- **API Layer**: All inputs validated before processing (SQL injection, XSS, path traversal blocked)
- **Prompt Validation**: AI inputs validated for injection attempts before sending to provider
- **Output Sanitization**: AI responses sanitized before rendering or storage

### Secrets Management

- No secrets in code or Git
- All secrets in AWS Secrets Manager
- IAM role-based access from ECS tasks
- Secret rotation supported without redeployment

## Observability

### Structured Logging

- **Format**: JSON with required fields (timestamp, correlation ID, method, path, status, duration, service)
- **Destination**: CloudWatch Logs
- **Retention**: 7 days (staging), 30 days (production)

### Health Checks

- **Endpoint**: `GET /healthz`
- **Response**: `200 OK` with `{"status":"ok"}` in < 100ms
- **ALB Configuration**: 30-second intervals, 3 healthy / 2 unhealthy thresholds

### Metrics & Alerts

- **ECS**: Task count, CPU/memory utilization
- **ALB**: Request count, target response time
- **RDS**: CPU utilization, database connections
- **Alerts**: CPU > 80%, connections > 80, unhealthy targets = 0

### Request Tracing

- **X-Request-ID**: Generated or propagated on all requests
- **Log Correlation**: Same ID in logs and response headers
- **Debugging**: Search CloudWatch by request ID

<!-- PROMOTED:architecture END -->

## Implementation Status (Spec 005 — System Architecture, Sprints 1–4; Sprint 5 partial update 2026-07-15)

This section is an addendum outside the promoted architecture above — it does not change any
component boundary agreed in specs 001-003, it documents what is actually built today (Sprint 4
close plus the Sprint 5 security-foundation work merged so far — PRs #151, #153–#157 — spec
`specs/005-system-architecture/` and specs 004/008) versus what remains target/planned. Component-level
detail lives in each area's own README (`backend/README.md`, `frontend/README.md`,
`infra/README.md`); this section is a cross-area summary kept in sync with them, not a duplicate
source of truth — see those files for exhaustive project-structure trees, environment variables,
and per-module notes.

### As-built component diagram

```mermaid
%%{init: {'flowchart': {'curve': 'linear'}}}%%
graph TD
    User[End User Browser]
    FE[React 19 SPA — Vite dev server :5173<br/>ErrorBoundary → QueryClientProvider → RouterProvider<br/>Single real route: placeholder HomePage]
    API[Go 1.26 API — cmd/api<br/>Chi router<br/>Middleware: RequestID → Logger → Recovery → CORS → BodySize<br/>GET /healthz, /api/v1/auth/* (Sprint 6), /swagger/*]
    DB[(PostgreSQL 15.18<br/>docker-compose locally; RDS module authored, not applied)]
    AI[Ollama + Gemma — local dev default<br/>AnthropicClient built for staging/production]
    TF[6 Terraform modules<br/>vpc · ecs · rds · alb · cloudfront · secrets<br/>validate/fmt clean, never applied]

    User -->|http://localhost:5173| FE
    FE -.->|VITE_API_BASE_URL configured; no route calls it yet| API
    API -->|pgx pool, DB_MIN/MAX_CONNECTIONS defaults 5/25| DB
    API -->|AI_PROVIDER=ollama default in development| AI
    TF -.->|would provision; AWS-cost-avoidance policy in force| DB

    style User fill:#e1f5ff
    style FE fill:#f3e5f5
    style API fill:#e8f5e9
    style DB fill:#fce4ec
    style AI fill:#fff9c4
    style TF fill:#fff4e6
```

### Backend (`backend/`)

The Chi router is built in `cmd/api/server.go`; `cmd/api/routes.go`'s `registerRoutes()` is the
single place new endpoints get mounted. The middleware chain runs in this exact order —
`RequestID → Logger → Recovery → CORS → BodySize` (`internal/middleware/`) — `RequestID` first so
downstream middleware can correlate by request ID, `Recovery` wrapping everything so a panic
anywhere still yields a clean `500`. `internal/database/client.go` wraps a `pgxpool` connection
pool behind a `database.Client` interface; since PR #157 its min/max size is genuinely driven by
`DB_MIN_CONNECTIONS`/`DB_MAX_CONNECTIONS` (defaults 5/25) rather than hardcoded. `internal/errors/`
defines `DomainError` (six constructors: `NotFound`, `Validation`, `Unauthorized`, `Forbidden`,
`Conflict`, `ServiceUnavailable`) and `HandleError`, which maps a domain error to the standard JSON
error envelope (`docs/api-design-standards.md` §7), including the correlation ID from context.

`internal/ai/` defines the `AIClient` interface (`GenerateItinerary`, `StreamItinerary`) with **two
real implementations** selected by the `AI_PROVIDER` environment variable: `OllamaClient` (default
in `development` — a free, local, no-API-key backend against a local Ollama server running a Gemma
model) and `AnthropicClient` (default in `production` — the official `anthropic-sdk-go`, delegating
retry/backoff to the SDK's own `option.WithMaxRetries`/`option.WithRequestTimeout`, surfacing an
exhausted-retries 429/503 as a `service_unavailable` `DomainError` with a `Retry-After` hint). See
[Local AI Setup](local-ai-setup.md).

`internal/example/` was a throwaway model→repository→service→handler reference implementation
(CRUD + optimistic locking via a `version` column, pagination, `swag` annotations for every
handler) that demonstrated the layering new domain packages should copy. Per that plan, it was
deleted in Sprint 8 (issue #234) once the first real domain packages —
`internal/trip/`/`internal/conversation/`, both repository-only so far — shipped; see the
"Divergences" note below and `backend/README.md` for the current state.

Sprint 5 (specs 004/008, PRs #153–#157) added the security foundation:

- **`internal/auth/`** — a **flat package** (`password.go`: `HashPassword`/`ComparePassword`,
  bcrypt cost 12; `validator.go`: `ValidatePassword`), deliberately *not* the
  `internal/auth/password/` subpackage some earlier spec/task text describes. HTTP
  handlers/services/repositories for auth are still unbuilt (Spec 008 Phase 3).
- **`internal/observability/`** — `correlation.go` (`GenerateCorrelationID`) and `logger.go`
  (`LogSecurityEvent(correlationID, eventType, userID, severity, ipAddress, userAgent, details)`,
  structured JSON via `log/slog`; matches the migrated `security_events` table).
- **`internal/{subscription, collaboration, security}/`** — `doc.go` scaffolding only, no
  implementation yet.
- **`internal/database/migrations/`** — shared goose migration config (`Dir`, `SetDialect()`)
  used by all integration tests; the `.sql` files themselves live flat in `backend/migrations/`
  (shared across specs — `001`–`004` create `users`/`refresh_tokens`/`jwt_signing_keys`/
  `security_events`; later migrations add the trip/subscription domain tables and, as of Sprint 8,
  `internal/conversation`'s tables (`017`–`018`, issue #234)). `pkg/` remains
  deliberately empty (`.gitkeep` only) — shared backend code goes under `internal/`, per the
  established convention.
- **Config-driven pool sizing** — `database.NewClient` now takes `minConns`/`maxConns` from
  `DB_MIN_CONNECTIONS`/`DB_MAX_CONNECTIONS` (defaults 5/25) instead of hardcoding them.

Beyond auth, domain packages remained scaffolds/unbuilt as of this sub-section's Sprint 5 writing:
`internal/{trip, itinerary, conversation, suggestion}` from the original spec text were still
unbuilt, and `internal/{subscription, collaboration, security}` were empty scaffolds. **Updated for
Sprint 8 (issue #234):** `internal/trip/` and `internal/conversation/` now exist as real,
repository-only packages (`model.go` + `repository.go`, no service/handler yet) — see the
"Divergences" note below. **Updated 2026-08-02 (Sprint 8, issues #235/#236):** both packages have
since gained a `service.go` (business logic, no knowledge of HTTP or SQL), and a third package,
`internal/itinerary/`, now exists alongside them: `Service` implements
`conversation.ItineraryGenerator`, driving `AIClient.StreamItinerary` through a schema-guided system
prompt and persisting the resulting Destination/Day/Activity rows through `trip.Repository` once the
AI reports the conversation ready. All three remain **HTTP-handler-free** — no domain route beyond
auth is mounted yet — see `backend/README.md`'s "Itinerary service" section and the "Divergences"
note below for the fuller picture. `backend/docs/` (a generated Swagger 2.0/OpenAPI contract, `swaggo/swag`)
is produced from Go doc-comment annotations on real handlers — as of Sprint 8, that means
`internal/auth/handler.go`'s (the reference implementation this sub-section originally cited,
`internal/example/handler.go`, was deleted once real domain packages shipped, issue #234) — and
served at `/swagger/index.html` / `/swagger/doc.json`, unauthenticated for now (see the "Planned
Addendum" section below, which predates this status update and remains accurate).

### Frontend (`frontend/`)

Vite + React 19 + TypeScript strict mode. `src/App.tsx` composes `ErrorBoundary` >
`QueryClientProvider` (`src/lib/query-client.ts`) > `RouterProvider` (`src/routes/index.tsx`).
Atomic Design layering under `src/components/`: `primitives/` (`Button`, `Input`, `Card`,
`Label`, `LoadingSpinner`, `ErrorMessage`, `EmptyState`, each token-driven via CSS Modules),
`composites/` (`Form`, composing `Button` + `Input`), and `features/` for feature-scoped modules —
`features/auth/` is the first real one (issue #180: `authApi`, `useRegister`/`useLogin`/`useLogout`,
and the session-expiry handler). `ErrorBoundary` itself sits outside this layering,
directly under `src/components/`. `src/stores/auth-store.ts` is a Zustand store
(`isAuthenticated`/`user`/`isLoading` + `login`/`logout`/`refreshSession`) with **no persistence
middleware** — the HTTP-only JWT cookie is the real session store, state is re-derived via
`refreshSession()` on reload. `src/lib/api-client.ts` (typed `fetch` wrapper, `APIError` matching
the backend's error envelope, a fresh `X-Request-ID` per request) and
`src/hooks/useErrorHandler.ts` (the first hook under `src/hooks/`, mapping an API error to
render-ready `{title, message, requestId?, isRetryable}`) form a complete, unit-tested
error-handling chain — but **no route consumes live API data yet**: `src/routes/` holds a single
real route, `HomePage`, a documented placeholder (`<h1>TrAIveler</h1>`). Auth/trip/generate pages
are future-sprint work.

### Infrastructure (`infra/`)

Six Terraform modules exist and pass `terraform validate` / `terraform fmt -check -recursive`:
`vpc`, `ecs` (Fargate, ARM64/Graviton2 task definitions), `rds` (PostgreSQL 15.18), `alb`
(`/healthz`-checked target group), `cloudfront` (S3 origin via OAI, SPA 404→`/index.html`
rewrite), and `secrets` (AI API key, JWT signing key — DB credentials are created directly by the
`rds` module). The root `infra/main.tf` wires all six via module outputs; `infra/environments/
{staging,production}.tfvars` hold environment-specific sizing matching the "Environment
Configurations" table above. `.github/workflows/infra-plan.yml` and `infra-apply.yml` exist; their
non-AWS steps (`terraform init -backend=false`, `validate`, `fmt -check`) run for real on every
pull request, but **every AWS-touching step is gated behind the `AWS_ROLE_ARN` repository secret**,
which has never been set — no `terraform apply` has run against real AWS infrastructure, per this
repo's standing AWS-cost-avoidance policy (see `.github/memory/session-notes.md`). OIDC IAM role
creation is fully documented and reproducible (`infra/README.md`) but deliberately not provisioned.

### Divergences from the original Spec 005 plan worth flagging

- The promoted architecture above (from specs 001-003) describes 8+ domain database tables and
  roughly 15 REST endpoint groups; most of that domain layer (trips, days, activities,
  destinations, collaborators, suggestions, conversation sessions) is still unbuilt. **Updated
  2026-08-01 (Sprint 6):** the authentication vertical is no longer aspirational — `POST
  /auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout`, and `GET /auth/me` are real,
  tested endpoints (`backend/internal/auth/handler.go`), gated by `middleware.Authenticate` on
  `/api/v1` (see `backend/README.md`'s Project Structure for the full `internal/auth/`,
  `internal/auth/jwt/`, `internal/auth/ratelimit/`, and `internal/subscription/` breakdown). At that
  point the only remaining purely-scaffolded surface was `/healthz` and the throwaway
  `internal/example` reference CRUD resource, alongside the still-unbuilt trip/collaboration domain
  above. **Updated 2026-08-02 (Sprint 8, issue #234):** `internal/example/` has since been deleted
  entirely — its route mount, migration, and mocks are gone — replaced by the first real domain
  repositories, `internal/trip/` (Trip/Destination/Day/Activity) and `internal/conversation/`
  (ConversationSession/ConversationMessage). Both are **repository-only**, with no HTTP surface yet,
  so they don't change the "no domain routes mounted beyond auth" picture above; the remaining
  scaffolded surface is now `/healthz` plus the still-unbuilt trip/collaboration HTTP layer.
  **Updated 2026-08-02 (Sprint 8, issues #235/#236):** both packages gained a `service.go`, and a
  third package, `internal/itinerary/` (no HTTP surface either), now implements
  `conversation.ItineraryGenerator` end-to-end against a real `AIClient`. This still doesn't change
  the "no domain routes mounted beyond auth" picture — all three remain HTTP-handler-free until
  001-T038/T039/T040 (issue #237). The
  `users` table's real columns (`has_subscription` instead of the `subscription_id` shown in the
  schema list above, plus other as-built differences) are tracked in `docs/data-model.md`'s "As-
  Built Schema Diagram" — that document is the authority for real column names, not this section.
- Local development defaults to **Ollama + Gemma**, not Anthropic Claude — a documented product
  decision made during implementation (not present in the original spec text), to avoid requiring
  an API key for MVP testing. Staging/production still target Anthropic Claude as originally
  planned, and the `AnthropicClient` implementation is complete.
- The frontend's error-handling chain (`APIError` → `useErrorHandler` → `ErrorMessage`) is fully
  implemented and tested, but not yet visible anywhere in the running app — it documents the
  pattern future feature pages should adopt, rather than something exercised today.
- Infrastructure modules are code-complete and validated but never applied — the Component
  Architecture diagram above depicts the target topology, not a currently running deployment.

## Local Development Note: AI Provider Override (Ollama)

This section is a local-development addendum, outside the promoted architecture above — it does
not change the staging/production architecture, which still targets **Anthropic Claude** as shown
in the Component Architecture diagram.

For local development and MVP testing, the backend defaults to a local **Ollama** server running a
**Gemma** model instead of Anthropic — free, no API key, no external network dependency once the
model is downloaded. `backend/internal/ai.AIClient` is the stable interface both backends
implement; `OllamaClient` (`backend/internal/ai/ollama_client.go`) is the concrete local
implementation, selected via the `AI_PROVIDER` environment variable (default `ollama` when
`GO_ENV=development`, `anthropic` when `GO_ENV=production` — see `backend/config/config.go`).

See **[docs/local-ai-setup.md](local-ai-setup.md)** for installation, model selection, and
`docker-compose` wiring.

## Planned Addendum: Generated API Documentation (Swagger/OpenAPI)

This section is an addendum outside the promoted architecture above — it does not change any
component boundary, it only documents a new generated-artifact surface. Defined in
`specs/009-api-documentation/` (spec/plan/research complete; implementation not yet built as of this
writing).

The backend will gain a machine-readable OpenAPI v3 contract, generated code-first from Go
doc-comment annotations via `swaggo/swag`, and served through Swagger UI via
`swaggo/http-swagger/v2`:

- **`backend/docs/`** — new generated-artifact directory (`docs.go`, `swagger.json`,
  `swagger.yaml`), committed to the repository like other generated artifacts in this codebase
  (e.g. `*mocks` packages per `docs/mock-standards.md`), and drift-checked in CI (see
  `docs/testing-guidelines.md`).
- **`/swagger/*` routes** (`GET /swagger/doc.json`, `GET /swagger/index.html`) — mounted inside the
  same Chi route group as `/api/v1`, so Swagger UI automatically inherits whatever authentication
  middleware Sprint 5 adds to that group. No bespoke auth is invented ahead of that; until Sprint 5
  lands, these routes have no auth, matching every other route today.

This complements, and does not replace, `docs/api-design-standards.md` §16 (the human-readable
conventions source of truth).
