# backend

> Go 1.24 REST API for TrAIveler — AI-powered travel itinerary generation.

This directory contains the server-side application: HTTP handlers, domain services, database
repositories, AI integration, observability middleware, and security primitives.

[← Back to root README](../README.md) | [Spec 001](../specs/001-product-vision-scope/spec.md) | [Spec 002 — NFRs](../specs/002-nfr-system-constraints/spec.md)

---

> **Implementation Status**: Configuration system, linting, and environment templates complete (Sprint 1). HTTP server, middleware chain, and database client implementation in progress (Sprint 2).

---

## Responsibility

The backend exposes a RESTful JSON API consumed by the frontend SPA. Its primary jobs are:

1. **Authentication and subscription** — issue and validate JWTs in HTTP-only cookies; enforce
   plan limits (basic: 1 admin + 1 partner per subscription).
2. **AI itinerary generation** — conduct a multi-turn Claude conversation, parse structured
   tool-use output into a typed itinerary, and persist it to PostgreSQL.
3. **Trip management** — CRUD operations on trips, days, activities, and destinations.
4. **Collaboration** — suggestion-based workflow where partners submit and admins approve/reject.
5. **Observability** — structured JSON logging, correlation IDs, and a `/healthz` endpoint.
6. **Security** — server-side prompt injection defence, AI output sanitisation, and a GDPR
   right-to-deletion endpoint.

---

## Tech Stack

| Concern | Library / Tool |
|---------|----------------|
| Language | Go 1.24 |
| HTTP router | `github.com/go-chi/chi/v5` |
| Database driver | `github.com/jackc/pgx/v5` (PostgreSQL 15.4, no ORM) |
| Migrations | `github.com/pressly/goose/v3` |
| Auth tokens | `github.com/golang-jwt/jwt/v5` (HTTP-only cookies) |
| AI provider | `github.com/anthropics/anthropic-sdk-go` (streaming, multi-turn, tool-use) |
| Structured logging | `log/slog` (stdlib, JSON handler) |
| HTML sanitisation | `github.com/microcosm-cc/bluemonday` |
| Unique IDs | `github.com/google/uuid` |
| Test assertions | `github.com/stretchr/testify` |
| Linting | `golangci-lint` (errcheck, govet, staticcheck, revive, gosec) |
| Security scanning | `gosec`, `govulncheck` |

---

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go                   # Process entry point: wires config, DB, router, and starts server
├── internal/                         # Domain packages — not importable outside this module
│   ├── auth/
│   │   ├── handler.go                # HTTP handlers: POST /register /login /logout, GET /me, DELETE /users/me
│   │   ├── service.go                # Auth business logic: Register (bcrypt), Login (JWT), Me
│   │   ├── repository.go             # User DB operations: Create, FindByEmail, FindByID, UpdateSubscription
│   │   └── user_deletion_test.go     # Integration test: right-to-deletion (NFR-PRIV-001)
│   ├── trip/
│   │   ├── handler.go                # Trip + collaborator HTTP handlers (GET/POST/PUT/DELETE /trips)
│   │   ├── service.go                # Trip service: CRUD, plan-limit enforcement for collaborators
│   │   └── repository.go             # Trip, Destination, Day, Activity, Collaborator DB operations
│   ├── itinerary/
│   │   ├── handler.go                # POST /trips/:id/generate — prompt validation + AI dispatch
│   │   └── service.go                # Build Claude prompt, stream response, parse tool-use, persist
│   ├── conversation/
│   │   ├── handler.go                # POST/GET /trips/:id/conversation — SSE streaming
│   │   └── service.go                # Multi-turn session management, anchor-place detection
│   ├── suggestion/
│   │   ├── handler.go                # GET/POST /trips/:id/suggestions, PATCH .../approve|reject
│   │   └── service.go                # Submit, Approve (applies to itinerary), Reject (preserves history)
│   ├── subscription/
│   │   ├── handler.go                # GET /plans, POST /checkout /confirm, GET /current
│   │   ├── service.go                # Plan management, checkout via payment provider
│   │   ├── repository.go             # Plan and Subscription DB operations
│   │   └── payment/
│   │       ├── provider.go           # PaymentProvider interface (swappable for real provider post-MVP)
│   │       └── stub.go               # StubProvider: always succeeds, logs [STUB] to stdout
│   ├── middleware/
│   │   ├── requestid.go              # Generates/propagates X-Request-ID correlation ID (NFR-OBS-003)
│   │   └── logger.go                 # Structured HTTP log middleware using log/slog (NFR-OBS-001)
│   ├── ai/
│   │   ├── validator/
│   │   │   └── prompt_validator.go   # Pattern-based deny-list; blocks prompt injection (NFR-SEC-007)
│   │   └── sanitizer/
│   │       └── output_sanitizer.go   # bluemonday HTML sanitiser for AI responses (NFR-SEC-008)
│   └── database/
│       └── db.go                     # pgxpool initialisation and teardown
├── pkg/                              # Packages safe to import from outside internal/
│   ├── health/
│   │   └── handler.go                # GET /healthz — HealthCheckResponse (NFR-OBS-002)
│   ├── middleware/
│   │   ├── auth.go                   # JWT cookie validation; attaches User to request context
│   │   └── role.go                   # RequireRole("admin"|"partner") guard factory
│   └── response/
│       └── response.go               # JSON Success() and Error() response helpers
├── config/
│   ├── config.go                     # Env var loading with fail-fast validation on missing required vars
│   ├── prompt-rules.yml              # Versioned deny-list patterns for the PromptValidator
│   ├── alerts.yml                    # Error-rate alerting rule definitions (NFR-OBS-004)
│   └── backup-policy.yml             # RPO ≤ 24 h backup schedule and restore instructions
└── migrations/
    └── 001_*.sql … 008_*.sql         # Goose SQL migration files (sequential, never edited after merge)
```

---

## Docker

The backend uses a multi-stage Dockerfile optimized for production deployment on AWS ECS Fargate.

### Build Strategy

**Stage 1 (builder):**
- Base: `golang:1.25-alpine`
- Installs build dependencies (git, ca-certificates, tzdata)
- Downloads Go modules (cached layer when go.mod/go.sum unchanged)
- Compiles static binary with `CGO_ENABLED=0` and stripped debug symbols (`-ldflags="-w -s"`)

**Stage 2 (runtime):**
- Base: `alpine:3.19` (minimal ~5MB base image)
- Copies only the compiled binary and migrations
- Runs as non-root user (`appuser:1000`) for security
- Includes health check polling `/healthz` endpoint every 30s

### Building the Image

```bash
# Build locally
cd backend
docker build -t traveler-backend:local .

# Build with custom tags
docker build -t traveler-backend:v1.0.0 .

# Test the image (requires DATABASE_URL and ANTHROPIC_API_KEY)
docker run -p 8080:8080 \
  -e DATABASE_URL="postgresql://..." \
  -e ANTHROPIC_API_KEY="sk-..." \
  traveler-backend:local
```

### Image Size

The multi-stage build produces a minimal runtime image:
- Builder stage: ~1.2GB (includes full Go toolchain)
- Runtime image: ~50MB (Alpine + binary + migrations)

Only the 50MB runtime image is pushed to ECR and deployed to ECS.

### Health Checks

The container includes a `HEALTHCHECK` instruction that Docker and ECS use to determine container health:
- **Endpoint**: `GET http://localhost:8080/healthz`
- **Interval**: 30s (check every 30 seconds)
- **Timeout**: 10s (wait up to 10 seconds for response)
- **Start Period**: 30s (wait 30s before first check to allow startup)
- **Retries**: 3 (3 consecutive failures mark container unhealthy)

When deployed to ECS, failed health checks trigger automatic container replacement.

---

## Prerequisites

| Tool | Version | Check |
|------|---------|-------|
| Go | ≥ 1.24 | `go version` |
| Docker + Docker Compose | ≥ 24 / ≥ 2 | `docker --version` |
| `golangci-lint` | ≥ 1.59 | `golangci-lint --version` |
| `gosec` | ≥ 2.21 | `gosec --version` |
| `govulncheck` | latest | `govulncheck -version` |
| `gitleaks` | ≥ 8 | `gitleaks version` |

---

## Environment Variables

Copy `backend/.env.example` to `backend/.env` and populate with actual values. **Never commit .env to version control.**

Configuration is loaded via `config/config.go` with fail-fast validation on startup.

### Required Variables

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string (format: `postgresql://user:pass@host:port/db?sslmode=disable`) |
| `ANTHROPIC_API_KEY` | Anthropic API key from https://console.anthropic.com/settings/keys |
| `JWT_SIGNING_KEY` | JWT signing secret (required in production; generate with: `openssl rand -base64 32`) |

### Optional Variables (with defaults)

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PORT` | `8080` | HTTP server listen port |
### Quick Start

**Option A: Using Docker Compose (Recommended for Local Development)**

The project includes a `docker-compose.yml` at the repository root that orchestrates PostgreSQL and the backend service.

```bash
# 1. Set up environment variables (from repository root)
cp backend/.env.example backend/.env
# Edit backend/.env and add your ANTHROPIC_API_KEY

# 2. Start all services (PostgreSQL + backend)
docker-compose up -d

# 3. Verify services are healthy
docker-compose ps
# Both traveler-db and traveler-api should show "healthy" status

# 4. View logs
docker-compose logs -f backend

# 5. Stop services
docker-compose down
```

The API will be available at `http://localhost:8080`. Test with:

```bash
curl http://localhost:8080/healthz
# Expected: {"status":"ok","database":"connected"}
```

**Option B: Local Go Development (without Docker)**

Useful when actively developing and debugging backend code.

```bash
# 1. Start only PostgreSQL via Docker Compose
docker-compose up -d postgres

# 2. Copy environment template and configure
cp backend/.env.example backend/.env
# Edit backend/.env with your DATABASE_URL and ANTHROPIC_API_KEY

# 3. Run database migrations (when migrations/ is implemented)
cd backend
goose -dir migrations postgres "$DATABASE_URL" up

# 4. Start the API server (when cmd/api/main.go is implemented)
go run ./cmd/api
```

**Option C: Running Tests**

Tests can run without Docker using in-memory or test fixtures.

```bash
cd backend
go test ./... -race -count=1
```

### Development Workflow

```bash
# Install dependencies
go mod download

# Format code (run before committing)
go fmt ./...
goimports -w .

# Run linter (must pass before merge)
golangci-lint run ./...

# Run tests with race detector
go test ./... -race -count=1

# Run tests with coverage (target: 80% for business logic)
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Security scanning
gosec -severity high -confidence medium ./...
govulncheck ./...
```

### Configuration Validation

The `config` package validates all configuration on startup with descriptive error messages:

```bash
# Missing required variable
$ unset DATABASE_URL
$ go run ./cmd/api
# Error: config validation failed: DATABASE_URL is required

# Invalid port
$ export HTTP_PORT=99999
$ go run ./cmd/api
# Error: HTTP_PORT must be between 1 and 65535, got 99999

# Invalid log level
$ export LOG_LEVEL=verbose
$ go run ./cmd/api
# Error: LOG_LEVEL must be one of: debug, info, warn, error; got verbose
```bash
# 1. Start PostgreSQL
docker compose up -d postgres

# 2. Run database migrations
cd backend
goose -dir migrations postgres "$DATABASE_URL" up

# 3. Start the API server
go run ./cmd/api
```

API is available at `http://localhost:8080`. The `/healthz` endpoint confirms it is ready:

```bash
curl http://localhost:8080/healthz
# {"status":"ok","version":"0.1.0","uptime_seconds":2.3}
```

---

## Running Tests

```bash
# Unit and integration tests (with race detector)
go test ./... -race -count=1

# Unit tests only, with coverage report
go test ./internal/... -coverprofile=coverage.out
go tool cover -func=coverage.out   # must report ≥ 80% for internal/

# Specific package
go test ./internal/auth/... -v -run TestRegister
```

---

## Linting and Security Scanning

```bash
# Lint (must report zero errors before merge)
golangci-lint run ./...

# SAST security scan (must report zero HIGH findings)
gosec -severity high -confidence medium ./...

# Dependency vulnerability audit
govulncheck ./...

# Secret scanning
gitleaks detect --source . --verbose
```

---

## API Overview

All endpoints are prefixed `/api/v1`. Authentication uses an HTTP-only cookie (`auth_token`).

| Group | Endpoints | Auth |
|-------|-----------|------|
| Auth | `POST /auth/register`, `/auth/login`, `/auth/logout`, `GET /auth/me`, `DELETE /users/me` | Mixed |
| Subscription | `GET /plans`, `POST /subscription/checkout`, `/confirm`, `GET /subscription/current` | Required |
| Trips | `GET/POST /trips`, `GET/PUT/DELETE /trips/:id` | Required (admin write) |
| Itinerary | `POST /trips/:id/generate` | Required (admin) |
| Conversation | `POST/GET /trips/:id/conversation` | Required |
| Collaborators | `POST/DELETE /trips/:id/collaborators/:user_id` | Required (admin) |
| Suggestions | `GET/POST /trips/:id/suggestions`, `PATCH .../approve`, `PATCH .../reject` | Required |
| Health | `GET /healthz` | None |

Full request/response schemas: [`specs/001-product-vision-scope/contracts/api.md`](../specs/001-product-vision-scope/contracts/api.md) and [`specs/002-nfr-system-constraints/contracts/api.md`](../specs/002-nfr-system-constraints/contracts/api.md).

---

## Code Standards

- All code, identifiers, and comments must be in **English** — see [coding guidelines](../docs/coding-guidelines.md).
- Format all Go files with `gofmt` before committing.
- Import order: stdlib → third-party → internal (see guidelines for exact format).
- Errors must never be silently ignored. Wrap with `fmt.Errorf("context: %w", err)`.
- Do not use `log.Fatal` or `os.Exit` outside of `main`.

---

## NFR Infrastructure

This package implements the cross-cutting NFR primitives defined in
[`specs/002-nfr-system-constraints/`](../specs/002-nfr-system-constraints/):

| Component | File | NFR |
|-----------|------|-----|
| Request-ID middleware | `internal/middleware/requestid.go` | NFR-OBS-003 |
| Structured-log middleware | `internal/middleware/logger.go` | NFR-OBS-001 |
| Health endpoint | `pkg/health/handler.go` | NFR-OBS-002 |
| Prompt injection defence | `internal/ai/validator/prompt_validator.go` | NFR-SEC-007 |
| AI output sanitisation | `internal/ai/sanitizer/output_sanitizer.go` | NFR-SEC-008 |
| Right-to-deletion endpoint | `internal/auth/handler.go` — `DELETE /users/me` | NFR-PRIV-001 |

Validation scenarios for each: [`specs/002-nfr-system-constraints/quickstart.md`](../specs/002-nfr-system-constraints/quickstart.md).

---

## Continuous Integration

The backend CI pipeline runs automatically on every pull request and push to main that modifies backend code.

**Workflow**: [`.github/workflows/backend-ci.yml`](../.github/workflows/backend-ci.yml)

**Triggers**:
- Pull requests modifying `backend/**`
- Push to `main` branch modifying `backend/**`
- Manual workflow dispatch

**Jobs**:

1. **Lint** — Runs `golangci-lint` with all configured linters (errcheck, govet, staticcheck, revive, gosec). Must pass with zero errors before merge.

2. **Test** — Executes `go test ./... -race -coverprofile=coverage.out` against a PostgreSQL 15.4 service container. Runs with race detector enabled and generates coverage report (target: ≥80% for `internal/` packages). Coverage artifact uploaded for review.

3. **Build** — Builds Docker image using multi-stage Dockerfile for `linux/arm64` (ECS Fargate Graviton2). Image is tagged with git SHA and uploaded as artifact. Does not push to ECR yet (Sprint 3).

**Future Enhancements** (TODO comments in workflow):
- **Sprint 3**: ECR push job using AWS OIDC authentication
- **Sprint 10**: ECS deployment job with rolling updates and health checks

**Local Equivalent**:

Run the same checks locally before pushing:

```bash
# Lint
golangci-lint run ./...

# Test with race detection
go test ./... -race -count=1

# Build Docker image
docker buildx build --platform linux/arm64 --tag traveler-backend:local backend/
```
