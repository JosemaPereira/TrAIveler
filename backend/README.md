# backend

> Go 1.26 REST API for TrAIveler — AI-powered travel itinerary generation.

This directory contains the server-side application: HTTP handlers, domain services, database
repositories, AI integration, observability middleware, and security primitives.

[← Back to root README](../README.md) | [Spec 001](../specs/001-product-vision-scope/spec.md) | [Spec 002 — NFRs](../specs/002-nfr-system-constraints/spec.md)

---

> **Implementation Status**: ✅ Sprint 2 complete (closed 2026-07-11) — Configuration, linting,
> environment templates, the PostgreSQL client (`internal/database/`, issue #53),
> observability/security middleware (`internal/middleware/`, issue #54), domain error handling
> (`internal/errors/`, issue #56), the HTTP server entry point (`cmd/api/`, issue #57), the AI
> client foundation (`internal/ai/`, issue #55 — `AIClient` interface, prompt-validator/output-
> sanitizer stubs, and a working `OllamaClient` for local dev/MVP testing), and the
> model→repository→service→handler reference pattern (`internal/example/`, issue #58) are complete
> and unit-tested. The Chi router exposes `GET /healthz` and, via the reference pattern, a demo
> `/api/v1/examples` CRUD resource. **No real domain package exists yet**: `internal/{auth, trip,
> itinerary, conversation, suggestion, subscription}` and `pkg/{health, middleware, response}`
> shown below under "Target" are still unbuilt, and the Anthropic-backed `AIClient` implementation
> is not yet written. `internal/example/` itself is throwaway reference code — it must be deleted
> once the first real domain package ships (tracked in `.github/memory/patterns-discovered.md`),
> not a permanent feature.

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
| AI provider | Local dev/MVP: **Ollama** (`internal/ai/ollama_client.go`, `net/http`, no SDK) running a Gemma model — see [docs/local-ai-setup.md](../docs/local-ai-setup.md). Staging/production: `github.com/anthropics/anthropic-sdk-go` (not yet implemented, tracked as `005-T112`) |
| Structured logging | `log/slog` (stdlib, JSON handler) |
| HTML sanitisation | `github.com/microcosm-cc/bluemonday` |
| Unique IDs | `github.com/google/uuid` |
| Test assertions | `github.com/stretchr/testify` (+ `vektra/mockery` for mocks) |
| Linting | `golangci-lint` (errcheck, govet, staticcheck, revive, gosec) |
| Security scanning | `gosec`, `govulncheck`, `gitleaks` |

---

## Project Structure

### Current

Everything below is real, built, and unit-tested as of Sprint 2's close (2026-07-11):

```
backend/
├── cmd/
│   └── api/
│       ├── main.go                   # Process entry point: config → DB client → HTTP server → graceful shutdown
│       ├── server.go                 # HTTPServer: builds the Chi router, middleware chain, /healthz handler, and the internal/example wiring
│       └── routes.go                 # registerRoutes(): single place new endpoints are wired up (currently /healthz + /api/v1/examples)
├── internal/                         # Domain packages — not importable outside this module
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
│   │   ├── client.go                 # AIClient interface (GenerateItinerary, StreamItinerary)
│   │   ├── types.go                  # Shared request/response types (Message, ItineraryRequest/Response, ...)
│   │   ├── validator.go              # PromptValidator stub — always valid; deny-list rules land in NFR-SEC-007
│   │   ├── sanitizer.go              # OutputSanitizer stub — pass-through; bluemonday policy lands in NFR-SEC-008
│   │   ├── ollama_client.go          # OllamaClient: real AIClient impl for local dev/MVP (see docs/local-ai-setup.md)
│   │   └── mocks/
│   │       └── ai_client_mock.go     # Generated AIClient mock (vektra/mockery)
│   ├── database/
│   │   └── client.go                 # pgxpool connection pooling, fail-fast retry, DI-friendly interface
│   └── example/                      # Canonical model→repository→service→handler reference pattern (issue #58)
│       ├── model.go                  # Example entity + validation
│       ├── repository.go             # PostgresRepository: CRUD + optimistic-locking (version column)
│       ├── service.go                # Business logic layer, calls repository
│       ├── handler.go                # HTTP handlers mounted at /api/v1/examples (If-Match optimistic-locking header)
│       └── mocks/                    # Generated repository/service mocks (vektra/mockery)
├── config/
│   ├── config.go                     # Env var loading with fail-fast validation on missing required vars
│   └── config_test.go
└── migrations/
    └── 20260710120000_create_examples_table.sql  # Goose timestamp-versioned migration for the examples table
```

> **`internal/example/` is throwaway reference code, not a permanent feature.** It exists solely to
> demonstrate the model→repository→service→handler layering, optimistic locking, and pagination
> conventions new domain packages should follow — it must be deleted once the first real domain
> package (e.g. Trip) ships. Its migration is deliberately timestamp-versioned
> (`20260710120000_...`) rather than sequentially numbered, because the sequential `001`–`016`
> range is reserved for the real foundational domain tables cataloged in
> [`docs/data-model.md`](../docs/data-model.md).
>
> **`internal/example/handler.go` is also the canonical example for `swag` doc-comment
> annotations.** Every handler function in that file carries a `swag` doc block (`@Summary`,
> `@Description`, `@Tags`, `@Accept`/`@Produce`, `@Param`, `@Success`, `@Failure`, `@Security
> BearerAuth`, `@Router`) directly above its function definition, matching the shape defined in
> [`specs/009-api-documentation/contracts/api.md`](../specs/009-api-documentation/contracts/api.md).
> When adding a real domain handler (Trip, Auth, ...), copy this file's annotation pattern rather
> than inventing a new one: reference request/response types with `{object} <TypeName>` (unexported
> types in the same package resolve fine — see `example.createRequest`/`example.listResponse` in
> the generated `backend/docs/swagger.json`), reference the shared error envelope as
> `errors.ErrorResponse` (its exported name in `internal/errors/handler.go`, added specifically so
> `swag` annotations elsewhere in the codebase can resolve it — always use the target package's
> real name, e.g. `errors`, not a local import alias like `domainerrors`), and run `make swagger`
> to regenerate `backend/docs/` after any annotation change. A handler with no annotations is
> silently excluded from the generated contract (e.g. `/healthz`) — this is intentional, not a bug.
>
> `pkg/` and `config/prompt-rules.yml` / `alerts.yml` / `backup-policy.yml` referenced in earlier
> planning docs do not exist yet — `pkg/` currently holds only a `.gitkeep` placeholder.
>
> **Planned**: `/swagger/*` routes (Swagger UI + `GET /swagger/doc.json`) serving the
> `backend/docs/` artifact below are defined in `specs/009-api-documentation/` but not yet mounted
> — see [docs/architecture.md](../docs/architecture.md) and
> [docs/api-design-standards.md](../docs/api-design-standards.md) §16.

### Target (planned, future specs)

Not yet built. Listed here so contributors know where new domain code is expected to land, per the
architecture in [`docs/architecture.md`](../docs/architecture.md) and the API contract in
[`specs/001-product-vision-scope/contracts/api.md`](../specs/001-product-vision-scope/contracts/api.md):

```
backend/
├── internal/
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
│   └── subscription/
│       ├── handler.go                # GET /plans, POST /checkout /confirm, GET /current
│       ├── service.go                # Plan management, checkout via payment provider
│       ├── repository.go             # Plan and Subscription DB operations
│       └── payment/
│           ├── provider.go           # PaymentProvider interface (swappable for real provider post-MVP)
│           └── stub.go               # StubProvider: always succeeds, logs [STUB] to stdout
├── pkg/                              # Packages safe to import from outside internal/ (currently empty)
│   ├── health/
│   │   └── handler.go                # GET /healthz — HealthCheckResponse (NFR-OBS-002); today this lives inline in cmd/api/server.go instead
│   ├── middleware/
│   │   ├── auth.go                   # JWT cookie validation; attaches User to request context
│   │   └── role.go                   # RequireRole("admin"|"partner") guard factory
│   └── response/
│       └── response.go               # JSON Success() and Error() response helpers
├── config/
│   ├── prompt-rules.yml              # Versioned deny-list patterns for the PromptValidator
│   ├── alerts.yml                    # Error-rate alerting rule definitions (NFR-OBS-004)
│   └── backup-policy.yml             # RPO ≤ 24 h backup schedule and restore instructions
└── migrations/
    └── 001_*.sql … 016_*.sql         # Sequential Goose migrations for the real domain tables (docs/data-model.md)
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
| Ollama | Latest | `ollama --version` | Only needed for the **Local Go** setup path below — `docker-compose` starts it automatically. See [docs/local-ai-setup.md](../docs/local-ai-setup.md) |
| `swag` | ≥ 1.16 | `swag --version` | Only needed to regenerate the OpenAPI/Swagger contract (`make swagger`). Install with `go install github.com/swaggo/swag/cmd/swag@latest`. See [specs/009-api-documentation/](../specs/009-api-documentation/) |

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
| `JWT_SIGNING_KEY` | JWT signing secret (required in production; generate with: `openssl rand -base64 32`) |
| `ANTHROPIC_API_KEY` | Anthropic API key from <https://console.anthropic.com/settings/keys> — **only required when `AI_PROVIDER=anthropic`** (default is `ollama`, no key needed; see [docs/local-ai-setup.md](../docs/local-ai-setup.md)) |

### Optional (with defaults)

| Variable | Default | Description |
|----------|---------|-------------|
| `GO_ENV` | `development` | Deployment environment (`development`, `staging`, `production`); gates the `JWT_SIGNING_KEY` requirement and selects `AI_PROVIDER`'s default (`ollama` unless `production`, see below) |
| `HTTP_PORT` | `8080` | HTTP server listen port |
| `HTTP_READ_TIMEOUT` | `30s` | Max duration for reading the entire incoming request |
| `HTTP_WRITE_TIMEOUT` | `30s` | Max duration before timing out writes of the response |
| `HTTP_IDLE_TIMEOUT` | `120s` | Max time to wait for the next request on a keep-alive connection |
| `ALLOWED_CORS_ORIGINS` | `http://localhost:5173` | Comma-separated exact-match allow-list consumed by the CORS middleware (`internal/middleware/cors.go`) |
| `DB_MAX_CONNECTIONS` | `25` | Maximum concurrent connections in the PostgreSQL pool |
| `DB_MIN_CONNECTIONS` | `5` | Minimum idle connections maintained in the PostgreSQL pool |
| `AI_PROVIDER` | `ollama` (dev) / `anthropic` (prod) | Selects the `ai.AIClient` backend; `config.Load()` fails fast on any other value — see [docs/local-ai-setup.md](../docs/local-ai-setup.md) |
| `OLLAMA_HOST` | `http://localhost:11434` | Local Ollama server base URL (`http://ollama:11434` under `docker-compose`) |
| `OLLAMA_MODEL` | `gemma3:4b` | Ollama model tag; must be pulled first (`ollama pull gemma3:4b`) |
| `ANTHROPIC_MODEL` | `claude-3-5-sonnet-20241022` | Anthropic model used when `AI_PROVIDER=anthropic` |
| `AI_TIMEOUT` | `60s` | Timeout applied to AI provider requests (shared across providers) |
| `AI_MAX_RETRIES` | `3` | Max retry attempts for failed AI provider requests (shared across providers) |
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

**Docker Compose (recommended)** — starts PostgreSQL, a local Ollama (AI) server, and the backend
together; no API key needed with the default `AI_PROVIDER=ollama` (see
[docs/local-ai-setup.md](../docs/local-ai-setup.md)):

```bash
cp backend/.env.example backend/.env
docker-compose up -d
docker-compose exec ollama ollama pull gemma3:4b   # one-time: download the local AI model
docker-compose ps     # traveler-db, traveler-ollama, and traveler-api should all be "healthy"
curl http://localhost:8080/healthz
```

**Local Go (for active development/debugging)** — runs the API directly against a
Dockerized PostgreSQL and a natively-installed Ollama:

```bash
docker-compose up -d postgres
cp backend/.env.example backend/.env    # defaults to AI_PROVIDER=ollama; edit DATABASE_URL as needed
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

make test                       # unit tests only (no Docker required) — go test -tags=test -v ./... -short
make test-all                   # ALL tests, including testcontainers — requires Colima/Docker running
make test-coverage              # coverage report (CI configuration) — target: ≥80% for internal/
go tool cover -func=coverage.out

gosec -severity high -confidence medium ./...   # SAST scan — zero HIGH findings required
govulncheck ./...                               # dependency vulnerability audit
gitleaks detect --source . --verbose            # secret scanning

make mocks                      # regenerate mocks (vektra/mockery) into */mocks subdirectories
make swagger                    # regenerate OpenAPI/Swagger docs (swaggo/swag) into backend/docs/
```

Most test files carry a `//go:build test` constraint, so plain `go test ./...` (no `-tags=test`)
silently compiles and runs only the untagged subset — always pass `-tags=test`, or use the `make`
targets above, which already do. `make test-integration` is a deprecated alias that now just prints
a pointer to `make test-all` and exits nonzero — testcontainers-based integration tests were folded
into `test-all` to avoid duplicate coverage. See `Makefile` (`make help` for the full target list).

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
[Mock Standards](../docs/mock-standards.md) and `internal/example/service_test.go` for a complete
usage example (`examplemocks.NewMockRepository(t)` with the fluent `EXPECT()` API).

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

Wired into a real handler as of issue #58 (`internal/example/`, the reference pattern described in
[Project Structure](#project-structure)): its `handler.go` calls `errors.HandleError()` on every
error path, and `service.go`/`repository.go` construct `Conflict`/`NotFound` `DomainError`s for
duplicate emails, missing rows, and optimistic-locking version mismatches. `/healthz` still uses its
own local `writeJSON` in `cmd/api/server.go`, unrelated to this package — it predates `internal/errors`
and has no domain-error case to report.

### AI client foundation (`internal/ai/`)

`AIClient` is the stable contract for generating trip itineraries from a conversation history
(`GenerateItinerary`, `StreamItinerary`); `types.go` holds the shared request/response shapes
(`Message`, `ItineraryRequest`, `ItineraryResponse`, `StreamChunk`, ...), matching the entity shapes
in [`docs/data-model.md`](../docs/data-model.md).

`PromptValidator.Validate` (`validator.go`) and `OutputSanitizer.Sanitize` (`sanitizer.go`) are
intentional pass-through stubs for this ticket — real prompt-injection deny-list rules
(NFR-SEC-007) and bluemonday HTML stripping (NFR-SEC-008) land in a future spec 002 ticket, per
[`docs/security.md`](../docs/security.md).

`OllamaClient` (`ollama_client.go`) is a real, non-stub implementation of `AIClient` against a
local [Ollama](https://ollama.com) server's native `/api/chat` endpoint, used for local development
and MVP testing (product decision, not upstream-mandated by this ticket) — free, no API key. It
requests `format: "json"` from Ollama to force valid-JSON model output, retries transient failures
(connection errors, 5xx) with linear backoff, and streams newline-delimited response chunks over a
channel for `StreamItinerary`. It does not build itinerary-specific system prompts or do
schema-guided generation — it forwards `ConversationHistory` as-is; that's spec 008's job. See
[docs/local-ai-setup.md](../docs/local-ai-setup.md) for setup.

The Anthropic-backed `AIClient` implementation for staging/production (`AI_PROVIDER=anthropic`) is
tracked separately (`docs/roadmap.md` task `005-T112`) and not yet built; `AIClient`'s interface is
designed so adding it won't require a breaking change.

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

All versioned endpoints are prefixed `/api/v1`. Authentication uses an HTTP-only cookie (`auth_token`).

### Built today

| Group | Endpoints | Auth |
|-------|-----------|------|
| Health | `GET /healthz` | None |
| Examples (reference pattern, throwaway — see [Project Structure](#project-structure)) | `GET/POST /api/v1/examples`, `GET/PUT/DELETE /api/v1/examples/:id` | None (no auth wired into this demo resource) |

### Planned (not yet built)

The domain routes below are the target API contract — none of them exist in the codebase yet (no
`auth`/`trip`/`itinerary`/`conversation`/`suggestion`/`subscription` package has been created, see
[Project Structure](#project-structure)'s "Target" tree):

| Group | Endpoints | Auth |
|-------|-----------|------|
| Auth | `POST /auth/register`, `/auth/login`, `/auth/logout`, `GET /auth/me`, `DELETE /users/me` | Mixed |
| Subscription | `GET /plans`, `POST /subscription/checkout`, `/confirm`, `GET /subscription/current` | Required |
| Trips | `GET/POST /trips`, `GET/PUT/DELETE /trips/:id` | Required (admin write) |
| Itinerary | `POST /trips/:id/generate` | Required (admin) |
| Conversation | `POST/GET /trips/:id/conversation` | Required |
| Collaborators | `POST/DELETE /trips/:id/collaborators/:user_id` | Required (admin) |
| Suggestions | `GET/POST /trips/:id/suggestions`, `PATCH .../approve`, `PATCH .../reject` | Required |

Full request/response schemas: [`specs/001-product-vision-scope/contracts/api.md`](../specs/001-product-vision-scope/contracts/api.md) and [`specs/002-nfr-system-constraints/contracts/api.md`](../specs/002-nfr-system-constraints/contracts/api.md).

---

## Docker

Multi-stage `Dockerfile`, optimized for AWS ECS Fargate (Graviton2/`linux/arm64`): a `golang:1.26-alpine`
builder stage — pinned to `--platform=$BUILDPLATFORM` so it always compiles natively rather than
under QEMU emulation, then cross-compiles the target arch via `GOOS`/`GOARCH` — produces a static
(`CGO_ENABLED=0`, stripped) binary, copied into a non-root `alpine:3.23` runtime image (bumped from
3.19, which had fallen out of its security-support window). A `HEALTHCHECK` polls `GET /healthz`
(30s interval, 10s timeout, 30s start period, 3 retries) — ECS uses the same signal to replace
unhealthy containers.

This image mirrors the ECS deployment target, where `AI_PROVIDER=anthropic` is expected — pass it
explicitly, since `AI_PROVIDER` otherwise defaults to `ollama` (see [Environment
Variables](#environment-variables)):

```bash
docker build -t traveler-backend:local backend/
docker run -p 8080:8080 \
  -e DATABASE_URL="postgresql://..." \
  -e AI_PROVIDER="anthropic" \
  -e ANTHROPIC_API_KEY="sk-..." \
  traveler-backend:local
```

---

## Continuous Integration

Workflow: [`.github/workflows/backend-ci.yml`](../.github/workflows/backend-ci.yml) — the `pull_request`
trigger runs unconditionally on every PR (deliberately not `paths:`-filtered, so the required status
check never gets stuck at "Expected"), but its `lint`/`test`/`build` jobs are skipped via an `if:`
gate unless the PR actually touches `backend/**`; the `push` trigger to `main` is `paths:`-filtered
to `backend/**` directly.

1. **Lint** — `golangci-lint` (errcheck, govet, staticcheck, revive, gosec); must pass with zero errors.
2. **Test** — `make test-coverage` (`go test -tags=test -short -race -coverprofile=coverage.out`) against a `postgres:15.4-alpine` service container — pinned to match the RDS `engine_version` planned for staging/production, not a stale version (target: ≥80% coverage for `internal/`).
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
| [Local AI Setup](../docs/local-ai-setup.md) | Installing Ollama, pulling a Gemma model, and configuring `AI_PROVIDER` for local development |
| [NFRs](../docs/nfrs.md) / [Spec 002 quickstart](../specs/002-nfr-system-constraints/quickstart.md) | Non-functional requirements and validation scenarios (NFR-OBS-*, NFR-SEC-*, NFR-PRIV-001 referenced in the project structure above) |
