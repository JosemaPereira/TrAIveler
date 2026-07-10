# backend

> Go 1.26 REST API for TrAIveler — AI-powered travel itinerary generation.

This directory contains the server-side application: HTTP handlers, domain services, database
repositories, AI integration, observability middleware, and security primitives.

[← Back to root README](../README.md) | [Spec 001](../specs/001-product-vision-scope/spec.md) | [Spec 002 — NFRs](../specs/002-nfr-system-constraints/spec.md)

---

> **Implementation Status**: Configuration system, linting, environment templates, database client, and
> observability/security middleware (`internal/middleware/`) complete and unit-tested (Sprint 2). The
> middleware package is not yet wired into an HTTP router/server — that integration is tracked separately
> (issue #57, blocked). HTTP server wiring remains in progress (Sprint 2).

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
| Language | Go 1.26 |
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

## Database Client

The backend uses `internal/database/client.go` for PostgreSQL connection pooling via `pgx/v5`.

### Architecture

The database client follows **Dependency Injection with interface-first design**:
- `database.Client` interface defines the contract (Ping, Close, Pool methods)
- `pgxClient` struct is the private concrete implementation
- `NewClient()` factory returns the interface, allowing easy mocking and implementation swapping

**Benefits:**
- Easy to mock in tests without requiring a real database
- Can swap implementations (e.g., in-memory for testing, different database providers)
- Clear contract documented by the interface
- Loose coupling between components

### Configuration

Connection pool settings (configured in `config.go`):
- **MinConns**: 5 — minimum idle connections maintained
- **MaxConns**: 25 — maximum concurrent connections
- **MaxConnLifetime**: 1 hour — connection reuse limit
- **MaxConnIdleTime**: 30 minutes — idle connection timeout

### Usage

```go
import (
    "context"
    "time"
    "github.com/JosemaPereira/TrAIveler/backend/config"
    "github.com/JosemaPereira/TrAIveler/backend/internal/database"
)

// Initialize client (returns interface, not concrete type)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

var dbClient database.Client
dbClient, err = database.NewClient(ctx, cfg.Database.URL)
if err != nil {
    log.Fatalf("failed to connect: %v", err)
}
defer dbClient.Close()

// Health check
if err := dbClient.Ping(ctx); err != nil {
    log.Fatalf("health check failed: %v", err)
}

// Execute queries using Pool()
var version string
err = dbClient.Pool().QueryRow(ctx, "SELECT version()").Scan(&version)
```

### Mocking for Tests

The project uses **[vektra/mockery](https://github.com/vektra/mockery)** for automated type-safe mock generation.

**Generate mocks:**
```bash
# From backend/ directory
make mocks

# Mocks are generated in /mocks subdirectory
# Example: internal/database/mocks/client_mock.go
```

**Use generated mocks in tests:**
```go
import (
    "testing"
    "github.com/stretchr/testify/mock"
    "github.com/JosemaPereira/TrAIveler/backend/internal/database"
    dbmocks "github.com/JosemaPereira/TrAIveler/backend/internal/database/mocks"
)

func TestMyService_HealthCheck_Success(t *testing.T) {
    // Create mock from mocks subdirectory
    mockDB := dbmocks.NewMockClient(t)
    
    // Setup expectation using fluent API
    mockDB.EXPECT().Ping(mock.Anything).Return(nil).Once()
    
    // Use mock in service
    service := NewMyService(mockDB)
    err := service.HealthCheck(context.Background())
    
    assert.NoError(t, err)
    // mockDB.AssertExpectations(t) called automatically
}

// Match specific arguments
func TestMyService_WithSpecificContext(t *testing.T) {
    mockDB := database.NewMockClient(t)
    
    ctx := context.WithValue(context.Background(), "request_id", "123")
    mockDB.On("Ping", ctx).Return(nil).Once()
    
    service := NewMyService(mockDB)
    err := service.HealthCheck(ctx)
    assert.NoError(t, err)
}
```

**See full examples:** `internal/database/client_mock_example_test.go`

**Mock Location Standard:**
- All mocks are in `/mocks` subdirectory: `internal/database/mocks/client_mock.go`
- Import with alias: `dbmocks "path/to/package/mocks"`

**Documentation:**
- [Mock Standards](../docs/mock-standards.md) — Comprehensive mock generation reference
- [Testing Guidelines](../docs/testing-guidelines.md) — Mockery usage guide
- [Coding Guidelines](../docs/coding-guidelines.md) — Mock generation standards

### Retry Logic

`NewClient` implements fail-fast connection with retry logic:
- **3 attempts** with 2-second delays between retries
- Logs warnings on each failed attempt
- Returns error after all retries exhausted
- Respects context cancellation during retries

### Testing

**Makefile targets:**
```bash
make test              # Run unit tests (DEFAULT - safe for local dev)
make test-all          # Run ALL tests including testcontainers (requires Colima)
make test-integration  # Run integration tests (requires DATABASE_URL)
make test-coverage     # Generate coverage report (same config as CI)
make mocks             # Generate mocks using mockery
```

**Quick Start (Local Development):**
```bash
# Default: unit tests only (no Docker required)
make test

# If you have Colima running and want full test coverage:
colima start --cpu 2 --memory 4
make test-all
```

**Unit tests** (no database required):
```bash
make test
# Or directly:
go test -tags=test -v ./... -short
```

**All tests including testcontainers** (requires Colima):
```bash
# ⚠️ Requires Colima running
colima status  # Check if running
colima start --cpu 2 --memory 4  # Start if needed

make test-all
```

**Integration tests only** (requires `DATABASE_URL`):
```bash
# Run integration tests
make test-integration
# Or directly:
go test -v ./internal/database/ -tags=integration -run="TestIntegration"
```

---

## Middleware

The backend uses `internal/middleware/` for cross-cutting HTTP concerns: request correlation,
structured access logging, panic recovery, CORS, and request body size limiting.

### Architecture

Every middleware follows the same **constructor injection** pattern as the database client above —
no globals, no hidden `config.Load()` or `slog.Default()` calls inside the package:
- `RequestID` and `BodySize` are plain `func(http.Handler) http.Handler` — stateless, no dependencies
  to inject.
- `Logger(logger *slog.Logger)` and `Recovery(logger *slog.Logger)` take the `*slog.Logger` as a
  constructor argument and return the `func(http.Handler) http.Handler` wrapper.
- `CORS(allowedOrigins string)` takes the comma-separated allow-list — sourced by the caller from
  `cfg.Server.AllowedCORS` (the existing `ALLOWED_CORS_ORIGINS` env var; see
  [Environment Variables](#environment-variables)) — as a constructor argument.

Dependencies are passed in explicitly by whatever wires the middleware chain, keeping each one
trivially mockable and testable in isolation.

### Components

| Middleware | File | Purpose |
|------------|------|---------|
| `RequestID` | `request_id.go` | Reuses an incoming `X-Request-ID` header verbatim, or generates a UUID v4; stores the ID in request context and echoes it on the response |
| `Logger` | `logger.go` | Structured JSON access log (`log/slog`) per request — `method`, `path`, `status`, `duration_ms`, `request_id` |
| `Recovery` | `recovery.go` | Recovers panics from downstream handlers, logs the panic value and stack trace, and responds `500` without leaking details to the client |
| `CORS` | `cors.go` | Exact-match origin allow-list; echoes the matched origin (never `*`) and handles `OPTIONS` preflight with `204` |
| `BodySize` | `body_size.go` | Caps request bodies at 10 MB — eager `413` via `Content-Length`, lazy `413` via `http.MaxBytesReader` for chunked bodies |

`errors.go` holds a shared `errorEnvelope` type and `writeErrorEnvelope` helper, used by
`recovery.go` and `body_size.go` to emit the standard JSON error format defined in
[`docs/api-design-standards.md`](../docs/api-design-standards.md) §7.

### Testing

Each middleware is unit-tested in isolation (`*_test.go` alongside its implementation) with no
external dependencies required — 22 tests, 98.6% coverage for the package.

### Status

Implemented and unit-tested, but **not yet wired into an HTTP router**. Composing the chain (adding
`github.com/go-chi/chi/v5` and registering it in `cmd/api/main.go`) is tracked separately and
currently blocked (issue #57).

---

## Development Tools

### Makefile Commands

The backend includes a Makefile with common development tasks:

```bash
make help              # Show all available targets
make test              # Run all tests
make test-unit         # Run unit tests only
make test-integration  # Run integration tests
make test-coverage     # Generate coverage report
make mocks             # Generate mocks using mockery
make lint              # Run golangci-lint
make fmt               # Format code with gofmt
make vet               # Run go vet
make clean             # Remove build artifacts
make install-tools     # Install development tools (mockery)
```

### Installing Development Tools

```bash
# Install all required tools
make install-tools

# Or manually:
go install github.com/vektra/mockery/v2@latest

# Install golangci-lint (see: https://golangci-lint.run/usage/install/)
# macOS:
brew install golangci-lint
# Linux:
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

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
│   │   ├── request_id.go             # Generates/propagates X-Request-ID correlation ID (NFR-OBS-003)
│   │   ├── logger.go                 # Structured HTTP log middleware using log/slog (NFR-OBS-001)
│   │   ├── recovery.go               # Recovers handler panics, logs stack trace, responds 500
│   │   ├── cors.go                   # Exact-match CORS origin allow-list, handles OPTIONS preflight
│   │   ├── body_size.go              # Caps request bodies at 10 MB, responds 413 when exceeded
│   │   └── errors.go                 # Shared JSON error-envelope helper used by recovery.go/body_size.go
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
- Base: `golang:1.26-alpine`
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

| Tool | Version | Check | Notes |
|------|---------|-------|-------|
| Go | ≥ 1.24 | `go version` | |
| **Container Runtime** | Latest | `docker --version` | **Use Colima or Podman** (free alternatives to Docker Desktop) |
| Docker Compose | ≥ 2 | `docker-compose --version` | |
| `golangci-lint` | ≥ 1.59 | `golangci-lint --version` | |
| `gosec` | ≥ 2.21 | `gosec --version` | |
| `govulncheck` | latest | `govulncheck -version` | |
| `gitleaks` | ≥ 8 | `gitleaks version` | |

### Container Runtime Setup

**⚠️ IMPORTANT:** Use a free container runtime alternative to Docker Desktop:

```bash
# Install Colima (recommended for macOS)
brew install colima

# Start Colima with Docker compatibility
colima start --cpu 2 --memory 4

# Verify Docker CLI works
docker ps
```

**Alternatives:**
- **Colima** (macOS/Linux) — Lightweight, Docker-compatible
- **Podman** (macOS/Linux) — Daemonless, Docker-compatible
- **Rancher Desktop** (macOS/Windows/Linux) — Full Kubernetes support

**Why not Docker Desktop?** Requires paid license for commercial use.

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
| `HTTP_READ_TIMEOUT` | `30s` | Max duration for reading the entire incoming request |
| `HTTP_WRITE_TIMEOUT` | `30s` | Max duration before timing out writes of the response |
| `HTTP_IDLE_TIMEOUT` | `120s` | Max time to wait for the next request on a keep-alive connection |
| `ALLOWED_CORS_ORIGINS` | `http://localhost:5173` | Comma-separated exact-match allow-list consumed by the CORS middleware (`internal/middleware/cors.go`) |
| `DB_MAX_CONNECTIONS` | `25` | Maximum concurrent connections in the PostgreSQL pool |
| `DB_MIN_CONNECTIONS` | `5` | Minimum idle connections maintained in the PostgreSQL pool |
| `AI_MODEL` | `claude-3-5-sonnet-20241022` | Anthropic model used for itinerary generation |
| `AI_TIMEOUT` | `60s` | Timeout applied to AI provider requests |
| `AI_MAX_RETRIES` | `3` | Max retry attempts for failed AI provider requests |
| `AI_STREAMING_CHUNK_SIZE` | `4096` | Buffer size (bytes) for streaming AI responses |
| `COOKIE_DOMAIN` | `localhost` | Domain attribute set on the auth cookie |
| `JWT_EXPIRATION` | `24h` | Access token lifetime |
| `REFRESH_TOKEN_EXPIRATION` | `7 days` | Refresh token lifetime |
| `BCRYPT_COST` | `12` | bcrypt hashing cost factor for password storage |
| `COOKIE_SECURE` | `false` (`true` in production) | Whether the auth cookie requires HTTPS |
| `LOG_LEVEL` | `info` | Minimum log level (`debug`, `info`, `warn`, `error`) |
| `LOG_FORMAT` | `json` | Structured log output format |

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
| Request-ID middleware | `internal/middleware/request_id.go` | NFR-OBS-003 |
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
