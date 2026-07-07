# backend

> Go 1.24 REST API for TrAIveler — AI-powered travel itinerary generation.

This directory contains the server-side application: HTTP handlers, domain services, database
repositories, AI integration, observability middleware, and security primitives.

[← Back to root README](../README.md) | [Spec 001](../specs/001-product-vision-scope/spec.md) | [Spec 002 — NFRs](../specs/002-nfr-system-constraints/spec.md)

---

> **Implementation Status**: Go module initialized with core dependencies (Chi, pgx, goose, uuid). Directory structure in place. Source code implementation in progress following Sprint 1 tasks.

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

Copy `.env.example` at the repo root and create a `.env` file. **Never commit real secret values.**

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | ✅ | PostgreSQL connection string, e.g. `postgres://traiveler:pass@localhost:5432/traiveler` |
| `ANTHROPIC_API_KEY` | ✅ | Anthropic API key for Claude — never hard-coded in source |
| `JWT_SECRET` | ✅ | Random secret for signing JWT tokens (min 32 bytes) |
| `JWT_TTL_SECONDS` | ✅ | Access token lifetime, e.g. `86400` (24 hours) |
| `PORT` | ✅ | HTTP listen port, e.g. `8080` |
| `FRONTEND_ORIGIN` | ✅ | Allowed CORS origin, e.g. `http://localhost:5173` |
| `AI_SYSTEM_PROMPT` | ✅ | Claude system prompt text — treated as a secret; never logged or returned to clients |
| `ENV` | ❌ | `development` \| `production`; controls log format and error verbosity |

---

## Setup

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
