# Implementation Plan: Product Vision and Scope

**Branch**: `001-product-vision-scope` | **Date**: 2026-07-02 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-product-vision-scope/spec.md`

## Summary

TrAIveler is a POC/MVP web application for AI-powered, conversational travel itinerary generation,
customization, and asynchronous collaboration. The backend is a Go 1.24 REST API (Chi router, pgx/v5
on PostgreSQL, Anthropic Claude for multi-turn AI). The frontend is a React 19 + TypeScript SPA.
Subscription plan enforcement (basic: 1 admin + 1 partner) is active application logic; payment
collection is a visible mock stub satisfying the same Go interface as a future real provider.
Collaboration uses a suggest-then-approve workflow; only the admin can apply changes.
See [research.md](research.md) for all dependency decisions.

## Technical Context

**Language/Version**: Go 1.24 (backend), TypeScript 5.x + React 19 (frontend)

**Primary Dependencies**:
- Backend: `github.com/go-chi/chi/v5` (HTTP router), `github.com/jackc/pgx/v5` (PostgreSQL),
  `github.com/golang-jwt/jwt/v5` (auth, HTTP-only cookies),
  `github.com/anthropics/anthropic-sdk-go` (AI provider),
  `github.com/pressly/goose/v3` (DB migrations), `github.com/stretchr/testify` (tests)
- Frontend: React 19, TypeScript strict, TanStack Query v5, React Router v7, Zustand,
  Vitest, React Testing Library, MSW, Playwright, Lucide React (icons)

**Storage**: PostgreSQL 16 — relational schema; JSONB for AI-generated content fields

**Testing**: Go `testing` + `testify` + `net/http/httptest`; Vitest + RTL + MSW; Playwright E2E

**Target Platform**: Linux server, single Docker instance, web application

**Project Type**: Web application — RESTful JSON API backend + SPA frontend

**Performance Goals**: Sustain < 50 concurrent users on a single instance; AI generation responses
streamed within 5 s of first token, completed within 90 s; API p95 < 200 ms for non-AI endpoints

**Constraints**: All secrets via environment variables (never hard-coded); AI provider API key
required at runtime; JWT in HTTP-only cookies; mock payment stub exposes identical Go interface to
real provider; no `log.Fatal` / `os.Exit` outside `main`

**Scale/Scope**: POC/MVP — < 50 concurrent users, best-effort uptime, single instance

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

[Gates determined based on constitution file]

## Project Structure

### Documentation (this feature)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | ✅ PASS | TDD enforced; all tests written before implementation; three layers (unit, integration, E2E) required |
| II. Simplicity — KISS & DRY | ✅ PASS | Chi (no magic), pgx (no ORM), no speculative abstractions; extraction only when pattern appears 3+ times |
| III. Code Quality & Consistency | ✅ PASS | `gofmt` + `golangci-lint` (Go); Prettier + ESLint strict (TypeScript); English identifiers throughout |
| IV. Accessible & Token-Driven UI | ✅ PASS | CSS custom property tokens in `src/styles/tokens.css`; WCAG 2.1 AA; Atomic Design layering enforced |
| V. Secure Configuration | ✅ PASS | All secrets via env vars; JWT in HTTP-only cookies; payment stub never processes real credentials; no hard-coded keys |

**No violations — Complexity Tracking table not required.**

## Project Structure

### Documentation (this feature)

```text
specs/001-product-vision-scope/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── api.md
└── tasks.md             # Phase 2 output (/speckit.tasks — not created by /speckit.plan)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/
│   │   ├── handler.go
│   │   ├── handler_test.go
│   │   ├── service.go
│   │   ├── service_test.go
│   │   └── repository.go
│   ├── trip/
│   │   ├── handler.go
│   │   ├── handler_test.go
│   │   ├── service.go
│   │   ├── service_test.go
│   │   └── repository.go
│   ├── itinerary/
│   │   ├── handler.go
│   │   ├── handler_test.go
│   │   ├── service.go
│   │   └── service_test.go
│   ├── suggestion/
│   │   ├── handler.go
│   │   ├── handler_test.go
│   │   ├── service.go
│   │   ├── service_test.go
│   │   └── repository.go
│   ├── subscription/
│   │   ├── handler.go
│   │   ├── handler_test.go
│   │   ├── service.go
│   │   ├── service_test.go
│   │   ├── repository.go
│   │   └── payment/
│   │       ├── provider.go       # PaymentProvider interface
│   │       ├── provider_test.go
│   │       └── stub.go           # StubProvider (mock checkout, always succeeds)
│   └── conversation/
│       ├── handler.go
│       ├── handler_test.go
│       ├── service.go
│       └── service_test.go
├── pkg/
│   ├── middleware/
│   │   ├── auth.go
│   │   └── auth_test.go
│   └── response/
│       └── response.go
├── config/
│   └── config.go
├── migrations/
│   ├── 001_create_users.sql
│   ├── 002_create_plans_subscriptions.sql
│   ├── 003_create_trips.sql
│   ├── 004_create_days_activities.sql
│   ├── 005_create_collaborators.sql
│   ├── 006_create_suggestions.sql
│   └── 007_create_conversation_sessions.sql
└── tests/
    └── integration/

frontend/
├── src/
│   ├── features/
│   │   ├── auth/
│   │   ├── trips/
│   │   ├── itinerary/
│   │   ├── suggestions/
│   │   └── subscription/
│   ├── components/
│   │   ├── primitives/       # Atomic: Button, Input, Badge, etc.
│   │   ├── composites/       # Molecules: TripCard, ActivityItem, SuggestionBubble
│   │   └── features/         # Organisms: ItineraryView, ConversationPanel, SuggestionQueue
│   ├── pages/
│   │   ├── HomePage.tsx
│   │   ├── TripPage.tsx
│   │   ├── GeneratePage.tsx
│   │   └── SubscribePage.tsx  # stub checkout flow
│   ├── hooks/
│   ├── services/              # API client (typed fetch wrappers)
│   └── styles/
│       └── tokens.css         # CSS custom property tokens
└── tests/
    └── integration/

e2e/
├── auth.spec.ts
├── subscribe-mock-checkout.spec.ts
├── generate-itinerary.spec.ts
├── share-trip.spec.ts
└── suggestion-approval-flow.spec.ts
```

**Structure Decision**: Option 2 — Web application. Backend uses the domain-based layout defined in
`docs/coding-guidelines.md` (`cmd/`, `internal/<domain>/`, `pkg/`, `config/`). Frontend follows
Atomic Design layering per `docs/ui-guidelines.md`. The payment stub is isolated under
`internal/subscription/payment/` to enforce the swappable-interface contract.

## Complexity Tracking

> No constitution violations — table omitted.
