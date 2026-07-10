# backend

> Go 1.26 REST API for TrAIveler — AI-powered travel itinerary generation.

This directory contains the server-side application: HTTP handlers, domain services, database
repositories, AI integration, observability middleware, and security primitives.

[← Back to root README](../README.md) | [Spec 001](../specs/001-product-vision-scope/spec.md) | [Spec 002 — NFRs](../specs/002-nfr-system-constraints/spec.md)

---

> **Implementation Status**: Configuration, linting, environment templates, the PostgreSQL client
> (`internal/database/`), observability/security middleware (`internal/middleware/`), and the HTTP
> server entry point (`cmd/api/`, issue #57) are complete and unit-tested (Sprint 2). The Chi router
> exposes `GET /healthz`. Domain routes (trips, auth, itinerary, ...) are not yet built.

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
| Test assertions | `github.com/stretchr/testify` (+ `vektra/mockery` for mocks) |
| Linting | `golangci-lint` (errcheck, govet, staticcheck, revive, gosec) |
| Security scanning | `gosec`, `govulncheck`, `gitleaks` |

---

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       ├── main.go                   # Process entry point: config → DB client → HTTP server → graceful shutdown
│       ├── server.go                 # HTTPServer: builds the Chi router, middleware chain, and /healthz handler
│       └── routes.go                 # registerRoutes(): single place new endpoints are wired up
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
│   ├── errors/
│   │   ├── types.go                  # DomainError type + constructors (NotFound, Validation, Unauthorized, Forbidden, Conflict)
│   │   └── handler.go                # HandleError(): maps *DomainError to the standard JSON error envelope (docs/api-design-standards.md §7)
│   ├── ai/
│   │   ├── validator/
│   │   │   └── prompt_validator.go   # Pattern-based deny-list; blocks prompt injection (NFR-SEC-007)
│   │   └── sanitizer/
│   │       └── output_sanitizer.go   # bluemonday HTML sanitiser for AI responses (NFR-SEC-008)
│   └── database/
│       └── client.go                 # pgxpool connection pooling, fail-fast retry, DI-friendly interface
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

## Prerequisites

| Tool | Version | Check | Notes |
|------|---------|-------|-------|
| Go | ≥ 1.26 | `go version` | |
| Container runtime | Latest | `docker --version` | Use [Colima](https://github.com/abiosoft/colima) or Podman — free alternatives to Docker Desktop, which requires a paid license for commercial use |
| Docker Compose | ≥ 2 | `docker-compose --version` | |
| `golangci-lint` | ≥ 1.59 | `golangci-lint --version` | |
| `gosec` | ≥ 2.21 | `gosec --version` | |
| `govulncheck` | latest | `govulncheck -version` | |
| `gitleaks` | ≥ 8 | `gitleaks version` | |

```bash
brew install colima
colima start --cpu 2 --memory 4
docker ps   # verify the Docker CLI works
```

---

## Environment Variables

Copy `backend/.env.example` to `backend/.env` and populate with actual values. **Never commit
`.env` to version control.** Configuration is loaded via `config/config.go`, which fails fast with
a descriptive error if a required variable is missing or a value is out of range (e.g. bad
`HTTP_PORT` or `LOG_LEVEL`).

### Required

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string (format: `postgresql://user:pass@host:port/db?sslmode=disable`) |
| `ANTHROPIC_API_KEY` | Anthropic API key from https://console.anthropic.com/settings/keys |
| `JWT_SIGNING_KEY` | JWT signing secret (required in production; generate with: `openssl rand -base64 32`) |

### Optional (with defaults)

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

---

## Setup

**Docker Compose (recommended)** — starts PostgreSQL and the backend together:

```bash
cp backend/.env.example backend/.env
# edit backend/.env and add your ANTHROPIC_API_KEY
docker-compose up -d
docker-compose ps     # both traveler-db and traveler-api should be "healthy"
curl http://localhost:8080/healthz
```

**Local Go (for active development/debugging)** — runs the API directly against a
Dockerized PostgreSQL:

```bash
docker-compose up -d postgres
cp backend/.env.example backend/.env    # edit with your DATABASE_URL and ANTHROPIC_API_KEY
cd backend
goose -dir migrations postgres "$DATABASE_URL" up
go run ./cmd/api
# HTTP server listening on :8080 (or $HTTP_PORT); Ctrl+C (SIGINT) drains in-flight requests
```

---

## Development Commands

```bash
go mod download                 # install dependencies
go fmt ./... && goimports -w .  # format (run before committing)
golangci-lint run ./...         # lint — must pass with zero errors before merge

go test ./... -race -count=1                          # unit + integration tests
go test -tags=test -v ./... -short                     # unit tests only (no Docker required)
go test ./internal/... -coverprofile=coverage.out       # coverage (target: ≥80% for internal/)
go tool cover -func=coverage.out

gosec -severity high -confidence medium ./...   # SAST scan — zero HIGH findings required
govulncheck ./...                               # dependency vulnerability audit
gitleaks detect --source . --verbose            # secret scanning

make mocks                      # regenerate mocks (vektra/mockery) into */mocks subdirectories
```

Testcontainers-based integration tests (`make test-all`, `make test-integration`) additionally
require a running container runtime and `DATABASE_URL`; see `Makefile` (`make help` for the full
target list).

---

## Architecture Notes

### Database client (`internal/database/`)

`database.NewClient()` returns the `database.Client` interface (`Ping`, `Close`, `Pool`) backed by
a `pgxpool`; the interface-first design keeps callers mockable without a real database. It retries
the initial connection up to 3 times (2s delay) and pings before returning, so a returned error is
fail-fast and terminal — callers don't need a separate health check after construction. Pool sizing
is driven by `DB_MAX_CONNECTIONS`/`DB_MIN_CONNECTIONS` (see [Environment
Variables](#environment-variables)); connections are recycled after 1h (`MaxConnLifetime`) or 30m
idle (`MaxConnIdleTime`).

Mocks are generated with `vektra/mockery` (`make mocks`) into `<package>/mocks/`. See
[Mock Standards](../docs/mock-standards.md) and `internal/database/client_mock_example_test.go`
for usage examples.

### Middleware (`internal/middleware/`)

Cross-cutting HTTP concerns, each built via constructor injection (dependencies passed in
explicitly — no globals, no hidden `config.Load()`/`slog.Default()` calls) so every middleware is
trivially testable in isolation:

| Middleware | File | Purpose |
|------------|------|---------|
| `RequestID` | `request_id.go` | Reuses an incoming `X-Request-ID` header verbatim, or generates a UUID v4; stores it in request context and echoes it on the response |
| `Logger` | `logger.go` | Structured JSON access log (`log/slog`) per request — `method`, `path`, `status`, `duration_ms`, `request_id` |
| `Recovery` | `recovery.go` | Recovers panics from downstream handlers, logs the panic value and stack trace, and responds `500` without leaking details |
| `CORS` | `cors.go` | Exact-match origin allow-list; echoes the matched origin (never `*`) and handles `OPTIONS` preflight with `204` |
| `BodySize` | `body_size.go` | Caps request bodies at 10 MB — eager `413` via `Content-Length`, lazy `413` via `http.MaxBytesReader` for chunked bodies |

`errors.go` holds the shared JSON error-envelope helper (used by `recovery.go` and `body_size.go`)
that implements the standard error format in
[`docs/api-design-standards.md`](../docs/api-design-standards.md) §7.

### Error handling (`internal/errors/`)

`DomainError` represents a business-rule failure (`Code`, `Message`, an optional `Fields` slice for
per-field validation errors, a `Details` map, and an optional wrapped `Err`) and implements
`error`/`Unwrap`, so it keeps composing with `errors.Is`/`errors.As` even after a lower layer wraps
it (e.g. `fmt.Errorf("...: %w", err)`). Five constructors — `NotFound`, `Validation`, `Unauthorized`,
`Forbidden`, `Conflict` — each build a fresh `*DomainError`; none is a shared package-level value,
since every occurrence needs its own dynamic message, fields, and cause.

`HandleError(w, r, err)` maps a `*DomainError` to the standard JSON error envelope from
[`docs/api-design-standards.md`](../docs/api-design-standards.md) §7:

| Domain error code | HTTP status |
|--------------------|-------------|
| `not_found` | 404 |
| `validation_failed` | 422 |
| `authentication_required` | 401 |
| `forbidden` | 403 |
| `conflict` | 409 |
| anything unrecognized (incl. non-`*DomainError` errors) | 500 / `internal_error` |

Any error that is not a `*DomainError` (and does not wrap one) is logged server-side via
`slog.Default()` — its real message never reaches the client, which sees only a generic
`internal_error`/500 so internals never leak.

This is a second, independent implementation from `internal/middleware`'s `errors.go`
(`writeErrorEnvelope`), not a reuse of it: that helper is unexported and used only internally by
`Recovery`/`BodySize` for transport-layer failures (panics, oversized bodies). `internal/errors` is
the exported, richer package — it adds `fields`/`details` and error wrapping — meant for every
future domain/business handler (trips, suggestions, auth, ...) to use. The two can't share one
implementation because `internal/errors` already imports `middleware` (for
`RequestIDFromContext`); a reverse import would create an import cycle.

Not yet wired into any handler — no existing endpoint calls `errors.HandleError()` (`/healthz` uses
its own local `writeJSON`, unrelated). Issue #58, next in the sprint and currently blocked on this
one, is what will first import and use this package from a real handler.

### HTTP server (`cmd/api/`)

`server.go`'s `NewHTTPServer` builds the Chi router and registers the middleware chain in this
exact order: `RequestID → Logger → Recovery → CORS → BodySize` — `RequestID` must run first so
later middleware can correlate by request ID, and `Recovery` must wrap everything downstream so a
panic anywhere still yields a clean `500`. `routes.go`'s `registerRoutes()` is the single place new
endpoints get added as the API grows. `main.go` wires `config.Load()` → `database.NewClient()` →
`NewHTTPServer()` → `*http.Server` (serving in a goroutine) → waits on `SIGINT`/`SIGTERM` →
`server.Shutdown(ctx)` with a 30s timeout → closes the database client.

**`GET /healthz`** pings the database with a 2-second bounded timeout:

| Condition | Status | Body |
|-----------|--------|------|
| Database reachable | `200` | `{"status":"healthy","database":"connected"}` |
| Database ping fails or times out | `503` | `{"status":"unhealthy","database":"disconnected","error":"<message>"}` |

No authentication required. This is distinct from the simpler `{"status":"ok"}` contract described
for `NFR-OBS-002` (tracked separately under roadmap group `G-OBS-HEALTHZ`) — this endpoint actively
checks the database, that one does not.

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

## Docker

Multi-stage `Dockerfile`, optimized for AWS ECS Fargate (Graviton2/`linux/arm64`): a `golang:1.26-alpine`
builder compiles a static (`CGO_ENABLED=0`, stripped) binary, copied into a non-root
`alpine:3.19` runtime image. A `HEALTHCHECK` polls `GET /healthz` (30s interval, 10s timeout, 30s
start period, 3 retries) — ECS uses the same signal to replace unhealthy containers.

```bash
docker build -t traveler-backend:local backend/
docker run -p 8080:8080 \
  -e DATABASE_URL="postgresql://..." \
  -e ANTHROPIC_API_KEY="sk-..." \
  traveler-backend:local
```

---

## Continuous Integration

Workflow: [`.github/workflows/backend-ci.yml`](../.github/workflows/backend-ci.yml) — runs on every
PR and push to `main` touching `backend/**`.

1. **Lint** — `golangci-lint` (errcheck, govet, staticcheck, revive, gosec); must pass with zero errors.
2. **Test** — `go test ./... -race -coverprofile=coverage.out` against a PostgreSQL 15.4 service container (target: ≥80% coverage for `internal/`).
3. **Build** — multi-stage Docker image for `linux/arm64`, tagged with the git SHA and uploaded as an artifact (not yet pushed to ECR — planned for Sprint 3).

---

## Related Documentation

| Doc | Covers |
|-----|--------|
| [Coding Guidelines](../docs/coding-guidelines.md) | Go formatting, naming, import order, error handling |
| [Testing Guidelines](../docs/testing-guidelines.md) | Test strategy, folder structure, coverage targets |
| [Mock Standards](../docs/mock-standards.md) | `vektra/mockery` generation and usage conventions |
| [API Design Standards](../docs/api-design-standards.md) | Resource naming, error envelope, pagination, versioning |
| [Architecture](../docs/architecture.md) | Component boundaries and integration rules |
| [NFRs](../docs/nfrs.md) / [Spec 002 quickstart](../specs/002-nfr-system-constraints/quickstart.md) | Non-functional requirements and validation scenarios (NFR-OBS-*, NFR-SEC-*, NFR-PRIV-001 referenced in the project structure above) |
