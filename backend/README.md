# backend

> Go 1.26 REST API for TrAIveler — AI-powered travel itinerary generation.

This directory contains the server-side application: HTTP handlers, domain services, database
repositories, AI integration, observability middleware, and security primitives.

[← Back to root README](../README.md) | [Spec 001](../specs/001-product-vision-scope/spec.md) | [Spec 002 — NFRs](../specs/002-nfr-system-constraints/spec.md)

---

> **Implementation Status**: ✅ Sprint 6 complete (closed 2026-08-01) — Sprints 1–4 delivered
> configuration, linting, the PostgreSQL client (`internal/database/`, issue #53),
> observability/security middleware (`internal/middleware/`, issue #54), domain error handling
> (`internal/errors/`, issue #56), the HTTP server entry point (`cmd/api/`, issue #57), the AI
> client foundation (`internal/ai/`, issue #55 — `AIClient` interface with **two** first-class
> implementations, `OllamaClient` for local dev/MVP and `AnthropicClient` for staging/production,
> issue #117), the model→repository→service→handler reference pattern (`internal/example/`, issue
> #58), and the generated Swagger contract + UI (`backend/docs/`, `/swagger/*`, Spec 009). Sprint 5
> (PRs #153–#157) added the security foundation: goose migrations `001`–`004`
> (users/refresh_tokens/jwt_signing_keys/security_events), the **flat** `internal/auth/` package
> (bcrypt password hashing + validation), `internal/observability/` (correlation IDs,
> `LogSecurityEvent`), the shared `internal/database/migrations` config package, and config-driven
> pool sizing. **Sprint 6** (issues #167–#184, #195, PRs #186–#202, closed 2026-08-01) shipped the
> full authentication vertical end-to-end, wired at runtime rather than left inert: `internal/auth/`
> grew from the flat bcrypt/validation helpers into real HTTP handlers (`handler.go` — `POST
> /auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout`, `GET /auth/me`), a service
> (`service.go` — registration, rate-limited login with no email-enumeration, optimistic-locking
> updates) and a repository (`repository.go`); its `jwt/` subpackage gained a `Generator`/`Validator`
> pair with `kid`-based multi-key rotation, a `KeyProvider`/`StaticKeyProvider`/`LoadKeyFromPEM` key
> seam, an `Issuer` (initial token pair) and `Refresher` (one-time-use refresh rotation); its new
> `ratelimit/` subpackage adds a progressive-delay, account-keyed login throttle. `middleware.Authenticate`
> (`internal/middleware/auth.go`) is now actually mounted — `cmd/api/auth.go` composes the whole
> vertical (key provider, generator, issuer, refresher, subscription service, auth service, handler)
> and `cmd/api/token_validator.go` adapts it into the gate `cmd/api/routes.go` applies to `/api/v1`.
> `internal/subscription/` grew past its `doc.go` scaffold into a real `models.go`/`repository.go`/
> `service.go` (charge-then-persist via a `payment/` subpackage's `PaymentProvider` seam and
> always-succeeds `StubPaymentProvider`) — but it still has **no HTTP handler/routes of its own**;
> registration/login reach it only through `internal/auth`'s composition. `internal/{collaboration,
> security}` remain `doc.go` scaffolds only (still unbuilt). The Chi router exposes `GET /healthz`,
> the auth endpoints above, the demo `/api/v1/examples` CRUD resource, and `/swagger/*`.
> `internal/{trip, itinerary, conversation, suggestion}` are still unbuilt, and `pkg/{health,
> middleware, response}` shown under "Target" remain superseded by the "everything under
> `internal/`" convention. `internal/example/` itself is still throwaway reference code — it must be
> deleted once the first real domain package ships (tracked in
> `.github/memory/patterns-discovered.md`), not a permanent feature.

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
| AI provider | Local dev/MVP: **Ollama** (`internal/ai/ollama_client.go`, `net/http`, no SDK) running a Gemma model — see [docs/local-ai-setup.md](../docs/local-ai-setup.md). Staging/production: **Anthropic Claude** (`internal/ai/anthropic.go`, `github.com/anthropics/anthropic-sdk-go`) |
| Structured logging | `log/slog` (stdlib, JSON handler) |
| HTML sanitisation | `github.com/microcosm-cc/bluemonday` |
| Unique IDs | `github.com/google/uuid` |
| Test assertions | `github.com/stretchr/testify` (+ `vektra/mockery` for mocks) |
| Linting | `golangci-lint` (errcheck, govet, staticcheck, revive, gosec) |
| Security scanning | `gosec`, `govulncheck`, `gitleaks` |

---

## Project Structure

### Current

Everything below is real, built, and unit-tested as of Sprint 6 closure (2026-08-01):

```
backend/
├── cmd/
│   └── api/
│       ├── main.go                   # Process entry point: config → DB client → HTTP server → graceful shutdown
│       ├── server.go                 # HTTPServer: builds the Chi router, middleware chain, /healthz handler, and the internal/example wiring
│       ├── auth.go                   # buildAuthComponents(): composes the auth vertical (JWT keys/generator/issuer/refresher, subscription + auth services, handler)
│       ├── token_validator.go        # *jwt.Validator → middleware.TokenValidator adapter (008-T207; lives in main to break the jwt→errors→middleware cycle)
│       └── routes.go                 # registerRoutes(): single place new endpoints are wired up (/healthz public; /api/v1 split into a public and an Authenticate-gated group; /swagger/* gated in production only)
├── internal/                         # Domain packages — not importable outside this module
│   ├── middleware/
│   │   ├── request_id.go             # Generates/propagates X-Request-ID correlation ID (NFR-OBS-003)
│   │   ├── logger.go                 # Structured HTTP log middleware using log/slog (NFR-OBS-001)
│   │   ├── recovery.go               # Recovers handler panics, logs stack trace, responds 500
│   │   ├── cors.go                   # Exact-match CORS origin allow-list, handles OPTIONS preflight
│   │   ├── body_size.go              # Caps request bodies at 10 MB, responds 413 when exceeded
│   │   ├── auth.go                    # Authenticate: requires a valid access_token cookie, validated via a TokenValidator/AuthClaims port; attaches user id + has_subscription to context, else 401
│   │   ├── rate_limit.go             # RateLimit(limit, window): per-IP fixed-window throttle (X-RateLimit-* headers, 429 + Retry-After); wired in server.go, disabled by default via config
│   │   ├── user_context.go           # UserIDFromContext / HasSubscriptionFromContext accessors for the values Authenticate stores
│   │   └── errors.go                 # Shared JSON error-envelope helper used by recovery.go/body_size.go/auth.go/rate_limit.go
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
│   │   ├── client.go                 # pgxpool pooling (min/max conns from DB_MIN/MAX_CONNECTIONS via caller), fail-fast retry, DI-friendly interface
│   │   ├── migrations/               # Shared goose config: Dir (absolute path to backend/migrations/) + SetDialect() — used by all integration tests
│   │   └── mocks/                    # Generated database.Client mock (vektra/mockery)
│   ├── auth/                         # Full auth vertical (Spec 004/008, Sprint 6, issues #167–#184): flat package + two subpackages — deliberately no password/ subpackage
│   │   ├── password.go               # HashPassword/ComparePassword (bcrypt cost 12, per docs/security.md)
│   │   ├── validator.go              # ValidatePassword: 8-72 chars, upper/lower/digit
│   │   ├── models.go                 # User/RefreshToken domain models + Register/Login request/response DTOs (issues #167/#169)
│   │   ├── repository.go             # PostgresUserRepository/PostgresRefreshTokenRepository: CRUD + optimistic locking on User.Version (issue #175)
│   │   ├── service.go                # Service: Register (uniqueness, hashing, optional subscription), Login (rate-limited, no email enumeration) (issue #176)
│   │   ├── handler.go                # HTTP handlers: POST /auth/register /login /refresh /logout, GET /auth/me — the canonical swag-annotation reference alongside internal/example (issues #177/#178/#195)
│   │   ├── jwt/                       # JWT token primitives (issue #144, 008-T019/T020/T021, wired live in Sprint 6)
│   │   │   ├── generator.go          # Generator (NewGenerator): signs RS256 access tokens with the primary key, stamps the kid header
│   │   │   ├── validator.go          # Validator (NewValidator): verifies access tokens, selecting the key per-token by kid (multi-key rotation); locked to RS256
│   │   │   ├── issuer.go             # Issuer (NewIssuer): mints the initial access+refresh token pair for a freshly authenticated session
│   │   │   ├── refresher.go          # Refresher (NewRefresher): one-time-use refresh-token rotation (validate → revoke old → issue new pair) against the real RefreshTokenStore + subscription.Resolver
│   │   │   ├── keys.go               # KeyProvider interface + StaticKeyProvider (in-memory) + LoadKeyFromPEM (raw-PEM loader, ≥2048-bit)
│   │   │   ├── claims.go             # Claims: registered claims + has_subscription; issuer constant
│   │   │   └── doc.go                # Package overview
│   │   ├── ratelimit/                # Account-level login-attempt throttle, consumed by auth.Service.Login (issue #176)
│   │   │   ├── limiter.go            # Limiter (New): progressive delay 2^(attempts-5)s after 5 failures in a 15-min window; CheckRateLimit/RecordFailure/Reset
│   │   │   └── store.go              # store: concurrency-safe in-memory TTL map keyed by email (package doc lives here)
│   │   └── mocks/                    # Generated UserRepository/RefreshTokenRepository mocks (vektra/mockery)
│   ├── observability/                # Correlation IDs + security event logging (Spec 004, PR #155)
│   │   ├── correlation.go            # GenerateCorrelationID
│   │   └── logger.go                 # LogSecurityEvent(correlationID, eventType, userID, severity, ipAddress, userAgent, details) — structured JSON, matches the security_events table
│   ├── subscription/                 # Billing domain (Spec 008, issues #167/#169/#168/#175); no HTTP handler/routes of its own yet — reached only through internal/auth's composition
│   │   ├── models.go                 # Subscription entity + Validate (matches migrations/007_create_subscriptions_table.sql — no period/version columns yet)
│   │   ├── repository.go             # PostgresRepository (Create/GetByUserID/Update/Cancel) + Resolver adapter (NewResolver) feeding jwt.Refresher's has_subscription re-check
│   │   ├── payment/                  # Payment-processing seam (issue #168)
│   │   │   ├── provider.go           # PaymentProvider interface: ProcessPayment(ctx, token, planID) (string, error)
│   │   │   ├── stub.go               # StubPaymentProvider: always succeeds, logs "[DEMO] Payment processed" — no real gateway integrated yet
│   │   │   └── mocks/                # Generated PaymentProvider mock (vektra/mockery)
│   │   ├── service.go                # Service.CreateSubscription: charges the payment provider, persists only on success (issue #176)
│   │   └── mocks/                    # Generated Repository mock (vektra/mockery)
│   ├── collaboration/                # doc.go scaffold only (Spec 008 Phase 3+, unbuilt)
│   ├── security/                     # doc.go scaffold only (Spec 008 Phase 3+, unbuilt)
│   └── example/                      # Canonical model→repository→service→handler reference pattern (issue #58)
│       ├── model.go                  # Example entity + validation
│       ├── repository.go             # PostgresRepository: CRUD + optimistic-locking (version column)
│       ├── service.go                # Business logic layer, calls repository
│       ├── handler.go                # HTTP handlers mounted at /api/v1/examples (If-Match optimistic-locking header)
│       └── mocks/                    # Generated repository/service mocks (vektra/mockery)
├── pkg/                              # Deliberately empty (.gitkeep in database/ and config/) — dead by convention; shared code goes under internal/
├── config/
│   ├── config.go                     # Env var loading with fail-fast validation on missing required vars
│   └── config_test.go
├── tests/                            # integration/ (testcontainers), fixtures/, security/, unit/, contract/ (last two are .gitkeep scaffolds)
└── migrations/                       # Single FLAT goose directory shared across specs — do not create per-spec migration dirs
    ├── 001_create_users_table.sql    # Spec 004 security tables (PR #154): users, refresh_tokens,
    ├── 002_create_refresh_tokens_table.sql   # jwt_signing_keys, security_events
    ├── 003_create_jwt_signing_keys_table.sql
    ├── 004_create_security_events_table.sql
    ├── 005_alter_users_add_auth_fields.sql       # Sprint 6 (Spec 008): full_name/has_subscription/
    ├── 006_create_plans_table.sql                # failed_login_attempts/version on users, plus
    ├── 007_create_subscriptions_table.sql        # plans/subscriptions/password_reset_tokens and the
    ├── 008_create_password_reset_tokens_table.sql # security_events.email ALTER — see docs/data-model.md's
    ├── 009_alter_security_events_add_email.sql   # "As-Built Migration Numbering" note for why 001-015
    ├── 010_create_trips_table.sql                # don't match that doc's target 16-entity numbering
    ├── 011_create_collaborators_table.sql
    ├── 012_create_suggestions_table.sql
    ├── 013_create_destinations_table.sql
    ├── 014_create_days_table.sql
    ├── 015_create_activities_table.sql
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
> CookieAuth`, `@Router`) directly above its function definition, matching the shape defined in
> [`specs/009-api-documentation/contracts/api.md`](../specs/009-api-documentation/contracts/api.md).
> `CookieAuth` is the single security definition declared in `cmd/api/docs.go` (`type: apiKey`,
> `in: header`, `name: Cookie`); it was named `BearerAuth`/`Authorization` until issue #179
> (008-T208), when the live gate made that name describe a scheme the server never accepted —
> authentication is the HTTP-only `access_token` cookie only. Add `@Security` to a handler if and
> only if it is mounted in the authenticated route group, and pair it with
> `@Failure 401 {object} errors.ErrorResponse`.
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
> `pkg/` is **dead by convention**: it holds only `.gitkeep` placeholders
> (`pkg/{database,config}/`), and every shared backend package so far has landed under
> `internal/` instead — even where a spec/task literally names a `pkg/...` path (e.g. Spec 008's
> `pkg/database/connection.go`, `pkg/config/config.go`, `pkg/database/migrations/` all resolved to
> `internal/database/`, `config/`, and `internal/database/migrations/` respectively). Treat any
> remaining `pkg/...` path in task text as stale. `config/prompt-rules.yml` / `alerts.yml` /
> `backup-policy.yml` referenced in earlier planning docs do not exist yet either.
>
> **`/swagger/*` routes are mounted** (issue #115): `GET /swagger/index.html` (interactive Swagger
> UI) and `GET /swagger/doc.json` (the generated Swagger 2.0 contract) serve the `backend/docs/`
> artifact above. Since issue #179 (008-T208) they are **authenticated** — they sit behind the same
> `middleware.Authenticate` gate as the rest of `/api/v1`, so a browser needs a live `access_token`
> cookie (log in first) to open them. See [Try the
> API](#try-the-api-interactive-swagger-ui) below, [docs/architecture.md](../docs/architecture.md),
> and [docs/api-design-standards.md](../docs/api-design-standards.md) §16.

### Target (planned, future specs)

Not yet built. Listed here so contributors know where new domain code is expected to land, per the
architecture in [`docs/architecture.md`](../docs/architecture.md) and the API contract in
[`specs/001-product-vision-scope/contracts/api.md`](../specs/001-product-vision-scope/contracts/api.md).
Two caveats vs. the tree below: `internal/auth/` already exists as a **flat** package holding the
password utilities (see "Current" above) — its future handler/service/repository files land in that
same flat package, not a subpackage; and the `pkg/{health, middleware, response}` block is
superseded by the "everything under `internal/`" convention (health lives in `cmd/api/server.go`,
middleware in `internal/middleware/`):

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
| `goose` | v3.27.2 | `goose --version` | Only needed for the **Local Go** setup path below (`make run` or a manual migration). Install with `go install github.com/pressly/goose/v3/cmd/goose@v3.27.2` |

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
| `REFRESH_TOKEN_EXPIRATION` | `720h` (30 days) | Refresh token lifetime |
| `BCRYPT_COST` | `12` | Loaded/validated by `config.Load()`, but **not currently consumed**: `internal/auth/password.go` hardcodes `const bcryptCost = 12` per `docs/security.md` — a non-default value has no effect today |
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

Or, once `.env` exists and Postgres is reachable, run the last two steps (migrate + start) in one
command with `make run` (from `backend/`) — it only skips Dockerizing the Go process itself, it
still expects Postgres to already be running per the step above.

### Try the API (interactive Swagger UI)

With the backend running (either setup path above), open the interactive Swagger UI in a browser:

```bash
open http://localhost:8080/swagger/index.html
```

The page renders every `swag`-annotated endpoint (today: the `internal/auth` endpoints and the
`internal/example` reference resource — see [Project Structure](#project-structure)) and lets you
send real requests against your locally running server via "Try it out" — the page loads its
contract from the live `GET /swagger/doc.json` route, not a static or hand-edited copy, so it
always reflects whatever `make swagger` last generated from the annotated handlers.

Ignore the **Authorize** box. The contract's `CookieAuth` definition is an `apiKey` in the `Cookie`
header (Swagger 2.0 has no cookie scheme — that is OpenAPI 3's `in: cookie`), and browsers refuse
to let a page set `Cookie` on an XHR, so typing a value there does nothing. You do not need it:
the UI is same-origin with the API and behind the same gate, so the browser that could open this
page already holds the `access_token` cookie and attaches it to every "Try it out" request
automatically.

For a step-by-step walkthrough (expand a tag, execute a real `POST /api/v1/examples` request, and
confirm the response matches what `curl` would return), see [Scenario 2 of
`specs/009-api-documentation/quickstart.md`](../specs/009-api-documentation/quickstart.md#scenario-2--interactive-ui-real-request).

`/swagger/*` is gated (008-T208) with the same `middleware.Authenticate` chain as the authenticated
`/api/v1` group in `cmd/api/routes.go`, but **only when `Config.IsProduction()`** (post-closure fix,
PR #204 — gating it unconditionally caused a 401 on a fresh local checkout hitting
`/swagger/index.html` before any login). In development it is open with no cookie required; in
production, open it in a browser that already holds an `access_token` cookie — e.g. after
`POST /api/v1/auth/login` — otherwise it answers `401 authentication_required`.

---

## Development Commands

```bash
go mod download                 # install dependencies
go fmt ./... && goimports -w .  # format (run before committing)
golangci-lint run ./...         # lint — must pass with zero errors before merge

make test                       # unit tests only (no Docker required) — go test -tags=test -v ./... -short
make test-all                   # ALL tests, including testcontainers — requires Colima/Docker running
make test-coverage              # coverage report (CI configuration) — target: ≥90% for internal/
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
it (e.g. `fmt.Errorf("...: %w", err)`). Six constructors — `NotFound`, `Validation`, `Unauthorized`,
`Forbidden`, `Conflict`, `ServiceUnavailable` — each build a fresh `*DomainError`; none is a shared
package-level value, since every occurrence needs its own dynamic message, fields, and cause.

`HandleError(w, r, err)` maps a `*DomainError` to the standard JSON error envelope from
[`docs/api-design-standards.md`](../docs/api-design-standards.md) §7:

| Domain error code | HTTP status |
|--------------------|-------------|
| `not_found` | 404 |
| `validation_failed` | 422 |
| `authentication_required` | 401 |
| `forbidden` | 403 |
| `conflict` | 409 |
| `service_unavailable` | 503 (also sets a `Retry-After` header from `Details["retry_after_seconds"]` — see [AI client foundation](#ai-client-foundation-internalai) below) |
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

**Concrete examples**, matching the real shipped code in `internal/example/`:

1. **Creating a domain error** — `repository.go`'s `Create` reports a racing duplicate email as a
   `Conflict`, and its `scanOne` helper (shared by `FindByID`/`FindByEmail`) reports a missing row
   as a `NotFound`:

   ```go
   // repository.go — Create: translate a UNIQUE-violation into a domain error
   if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
       return domainerrors.Conflict(fmt.Sprintf("an example with email %q already exists", ex.Email))
   }

   // repository.go — scanOne: translate "no rows" into a domain error
   if errors.Is(err, pgx.ErrNoRows) {
       return nil, domainerrors.NotFound("example", fmt.Sprintf("%v", arg))
   }
   ```

   `Conflict(message string)` and `NotFound(resource, id string)` are two of the six constructors
   listed above; both build a fresh `*DomainError` with no wrapped `Err`, since there is no
   lower-level cause worth preserving for either case.

2. **Wrapping instead of translating** — the same `scanOne` method also shows the *other* path, for
   an error that stays unmapped:

   ```go
   return nil, fmt.Errorf("find example: %w", err)
   ```

   This wrapped error is **not** a `*DomainError` and does not wrap one, so it falls straight
   through `HandleError`'s `errors.As(err, &domainErr)` check into the generic branch: the client
   only ever sees `internal_error`/500, while the real cause is logged server-side via
   `slog.Default()` with the request ID attached. This is exactly the mechanism
   `tests/integration/error_test.go` (a sibling PR, #131, not yet merged as of this writing)
   exercises for a DB-timeout scenario — asserting a generic 500 body reaches the client while the
   specific timeout error still lands in the logs.

3. **Handler translation** — every handler in `handler.go` follows the same one-line pattern to
   turn a service error into the standard envelope; `handleGet` is the simplest example:

   ```go
   ex, err := h.service.GetExample(r.Context(), id)
   if err != nil {
       domainerrors.HandleError(w, r, err)
       return
   }
   ```

   Any future domain handler (trips, suggestions, auth, ...) should follow this exact shape — the
   handler decides nothing about status codes or envelope shape, `HandleError` owns that entirely.

4. **Retry/backoff on an outbound dependency** — see
   [AI client foundation](#ai-client-foundation-internalai) below for `AnthropicClient`'s approach:
   it delegates retry/backoff to the SDK's own `option.WithMaxRetries`/`option.WithRequestTimeout`
   rather than a hand-rolled loop, and an exhausted-retries 429/503 ultimately surfaces as a
   `"service_unavailable"` `DomainError` carrying a `Retry-After` header, via `TranslateError` — see
   that section for the full mechanism rather than repeating it here.

5. **Correlation ID propagation** — the same request ID appears in two places on every response:
   the `X-Request-ID` response header (set once, first in the middleware chain, by `RequestID` in
   `cmd/api/server.go`) and the JSON error body's `request_id` field, pulled from the same context
   value by `HandleError`/`writeErrorResponse` via `middleware.RequestIDFromContext(r.Context())`.
   For a failing request they carry the same value — a client or support engineer can quote either
   one and land on the same request:

   ```text
   X-Request-ID: 6f1a9e2e-52f0-4b2b-9d21-9d6a7d9e0c31            <- response header

   {"error":"not_found","message":"example with ID \"abc\" not found",
    "request_id":"6f1a9e2e-52f0-4b2b-9d21-9d6a7d9e0c31"}         <- response body
   ```

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

`AnthropicClient` (`anthropic.go`) is the Anthropic-backed `AIClient` implementation for
staging/production (`AI_PROVIDER=anthropic`, `docs/roadmap.md` task `005-T112`), built on the
official `github.com/anthropics/anthropic-sdk-go`. It delegates retry/backoff to the SDK's own
built-in support (`option.WithMaxRetries`, `option.WithRequestTimeout`) rather than a hand-rolled
loop; a 429/503 response that survives every retry attempt is returned as a
`*ProviderUnavailableError` (wrapping the `ErrProviderUnavailable` sentinel) instead of an opaque
error, so callers can distinguish "provider is transiently unavailable" from other failures. Like
`OllamaClient`, it forwards `ConversationHistory` as-is and parses the model's response text
directly as itinerary JSON — schema-guided generation is spec 008's job.

`NewAIClient` (`client.go`) is the single factory that picks between `OllamaClient` and
`AnthropicClient` based on `config.AIConfig.Provider`, so callers depend only on the `AIClient`
interface and never construct a provider directly. `TranslateError` (`client.go`) is the
factory-adjacent helper that turns a `*ProviderUnavailableError` into a `"service_unavailable"`
`*errors.DomainError` carrying a `Retry-After` hint (`errors.ServiceUnavailable`,
`internal/errors/types.go`) — `writeErrorResponse` (`internal/errors/handler.go`) sets the
`Retry-After` response header from it (005-T113).

### HTTP server (`cmd/api/`)

`server.go`'s `NewHTTPServer` builds the Chi router and registers the middleware chain in this
exact order: `RequestID → Logger → Recovery → CORS → BodySize` — `RequestID` must run first so
later middleware can correlate by request ID, and `Recovery` must wrap everything downstream so a
panic anywhere still yields a clean `500`. `routes.go`'s `registerRoutes()` is the single place new
endpoints get added as the API grows. `main.go` wires `config.Load()` → `database.NewClient()` →
`NewHTTPServer()` → `*http.Server` (serving in a goroutine) → waits on `SIGINT`/`SIGTERM` →
`server.Shutdown(ctx)` with a 30s timeout → closes the database client.

**`GET /healthz`** pings the database with a 2-second bounded timeout. Per spec 002's
`HealthCheckResponse` validation rules
([`specs/002-nfr-system-constraints/data-model.md`](../specs/002-nfr-system-constraints/data-model.md)),
the HTTP status is **always `200 OK`** — load balancers key their routing decisions off the HTTP
status code, so a transient DB blip must never trigger a failover. The ping result is signaled only
through the JSON body's `status` field (`"ok"` or `"degraded"`), plus a `logger.Warn` so the failure
is still visible to operators instead of disappearing silently:

| Condition | Status | Body |
|-----------|--------|------|
| Database reachable | `200` | `{"status":"ok","version":"0.1.0","uptime_seconds":123.45}` |
| Database ping fails or times out | `200` | `{"status":"degraded","version":"0.1.0","uptime_seconds":123.45}` |

No authentication required. `version` and `uptime_seconds` come from `cmd/api/server.go`'s package
`version` variable and `startTime` respectively, matching the `HealthCheckResponse` struct defined
there.

---

## API Overview

All versioned endpoints are prefixed `/api/v1`. Authentication uses an HTTP-only cookie (`access_token`).

### Built today

Since issue #179 (008-T208) `cmd/api/routes.go` splits `/api/v1` into a **public** group and an
**authenticated** group carrying `middleware.Authenticate`; `/healthz` sits outside both, and
`/swagger/*` is gated the same way, but only when `Config.IsProduction()` (PR #204) — open in
development.

| Group | Endpoints | Auth |
|-------|-----------|------|
| Health | `GET /healthz` | None (public, unversioned) |
| Auth (public) | `POST /api/v1/auth/register`, `/api/v1/auth/login`, `/api/v1/auth/refresh` | None — refresh authenticates with the `refresh_token` cookie |
| Auth (authenticated) | `POST /api/v1/auth/logout`, `GET /api/v1/auth/me` | Required (`access_token` cookie) |
| Examples (reference pattern, throwaway — see [Project Structure](#project-structure)) | `GET/POST /api/v1/examples`, `GET/PUT/DELETE /api/v1/examples/:id` | Required (`access_token` cookie) |
| API docs | `GET /swagger/index.html`, `GET /swagger/doc.json` | Required in production only (`access_token` cookie); open in development |

### Planned (not yet built)

The domain routes below are the target API contract and have no HTTP surface yet (`internal/auth/`
and `internal/subscription/` exist as packages, but only the auth endpoints listed under "Built
today" are mounted; `internal/{trip, itinerary, conversation, suggestion}` are unbuilt — see
[Project Structure](#project-structure)'s "Target" tree):

| Group | Endpoints | Auth |
|-------|-----------|------|
| Auth (remaining) | `DELETE /users/me`, the password-reset/change endpoints | Mixed |
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
2. **Test** — `make test-coverage` (`go test -tags=test -short -race -coverprofile=coverage.out`). No PostgreSQL service container: `-short` skips every testcontainer-gated, DB-backed test, so nothing in CI needs a live database (one was configured here until 2026-07-16, but nothing ever connected to it) — target: ≥90% coverage for `internal/`.
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
