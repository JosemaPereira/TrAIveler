# TrAIveler

> **Plan smarter trips. Discover hidden gems. Travel together.**

TrAIveler is an AI-powered web application that helps travelers generate, customize, and
collaboratively refine travel itineraries. It turns an overwhelming planning process — researching
destinations, sequencing visits, finding local food, and coordinating with companions — into a
guided, personalized, and shareable experience.

This project is the capstone of the AI Bootcamp at Slalom.

---

## The Problem

Planning a trip from scratch is time-consuming and fragmented. Travelers either over-plan (hours in
spreadsheets) or under-plan (missing local gems). TrAIveler removes that friction by generating a
coherent, day-by-day itinerary tailored to the traveler's style, experience level, and timeframe —
with no prior knowledge of the destination required.

## Core Value Proposition

A traveler describes where they want to go and for how long. Through a **conversational,
multi-turn AI flow**, the system asks targeted questions to understand their preferences, then
produces a complete itinerary covering must-see attractions, off-the-beaten-path discoveries, local
gastronomy, and inter-city logistics — which they can refine and share with companions.

---

## Project Status

**POC / MVP** — Currently in the planning and specification phase.

| Artifact | Status |
|----------|--------|
| Product vision & scope | ✅ Complete — [`specs/001-product-vision-scope/spec.md`](specs/001-product-vision-scope/spec.md) |
| Implementation plan | ✅ Complete — [`specs/001-product-vision-scope/plan.md`](specs/001-product-vision-scope/plan.md) |
| Technology research | ✅ Complete — [`specs/001-product-vision-scope/research.md`](specs/001-product-vision-scope/research.md) |
| Data model | ✅ Complete — [`specs/001-product-vision-scope/data-model.md`](specs/001-product-vision-scope/data-model.md) |
| REST API contract | ✅ Complete — [`specs/001-product-vision-scope/contracts/api.md`](specs/001-product-vision-scope/contracts/api.md) |
| Task list (81 tasks) | ✅ Complete — [`specs/001-product-vision-scope/tasks.md`](specs/001-product-vision-scope/tasks.md) |
| Backend implementation | 🔲 Not started |
| Frontend implementation | 🔲 Not started |

---

## MVP Scope

1. **User authentication + subscription** — account registration with a visible stub checkout flow;
   basic plan (1 admin + 1 partner collaborator per subscription).
2. **AI itinerary generation** — conversational multi-turn flow; free-form natural language input;
   day-by-day plan with visits, food, logistics, and inter-city transfers.
3. **Travel style personalization** — gastronomy, sports, technology, museums & art, film &
   audiovisual media.
4. **Itinerary customization** — add, edit, remove, and reorder destinations and activities.
5. **Existing plan enrichment** — provide a list of places; AI anchors around them and fills gaps.
6. **Collaboration** — admin invites one partner; partner submits suggestions; admin approves or
   rejects; full suggestion history preserved.

**Out of scope (MVP)**: real payment processing, native mobile apps, real-time booking, third-party
map integrations (Google Maps, etc.), offline access, social feeds.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.24, Chi router, pgx/v5 (PostgreSQL), Goose migrations, JWT (HTTP-only cookies) |
| AI provider | Anthropic Claude (`anthropic-sdk-go`) — streaming, multi-turn, tool-use |
| Frontend | React 19, TypeScript strict, React Router v7, TanStack Query v5, Zustand |
| Testing | Go `testing` + testify + `net/http/httptest`; Vitest + RTL + MSW; Playwright E2E |
| Database | PostgreSQL 16 |
| Infrastructure | Docker + Docker Compose (local), single Linux instance |

---

## Planned Repository Structure

```
backend/          # Go 1.24 REST API
  cmd/server/     # main entrypoint
  internal/       # domain packages: auth, trip, itinerary, conversation, suggestion, subscription
  pkg/            # shared middleware and response helpers
  config/         # environment-based configuration
  migrations/     # Goose SQL migration files

frontend/         # React 19 + TypeScript SPA
  src/
    features/     # domain slices: auth, trips, itinerary, suggestions, subscription
    components/   # Atomic Design layers: primitives, composites, features
    pages/        # route-level page components
    services/     # typed API client functions
    styles/       # CSS custom property design tokens

e2e/              # Playwright end-to-end test specs

specs/            # Feature specifications, plans, research, data models, contracts
  001-product-vision-scope/

.github/
  copilot-instructions.md   # project-wide AI coding standards
  prompts/                  # reusable agent prompts
  agents/                   # custom agent definitions
```

---

## Development Workflow

- **Branching**: all work on `feature/<descriptive-name>` branches — never commit directly to `main`.
- **Commits**: [Conventional Commits](https://www.conventionalcommits.org/) in English (`feat:`, `fix:`, `chore:`, `docs:`, etc.).
- **TDD**: Red → Green → Refactor, mandatory. Tests are written before implementation.
- **Quality gates** (must pass before merge): `golangci-lint` (backend) · Prettier + ESLint (frontend) · unit tests · integration tests · E2E tests.
- **Versioning**: [Semantic Versioning 2.0.0](https://semver.org/).

---

## Key Design Decisions

| Topic | Decision |
|-------|---------|
| AI provider | Anthropic Claude — Go 1.24+ native SDK, first-class multi-turn support |
| HTTP router | `go-chi/chi/v5` — idiomatic, net/http compatible, fully testable |
| DB driver | `jackc/pgx/v5` — context-first, built-in connection pool, no ORM |
| Auth | `golang-jwt/jwt/v5` in HTTP-only cookies (XSS-safe, CSRF-mitigated) |
| Payment | Visible mock stub (always succeeds); same Go interface as real provider for post-MVP swap |
| Collaboration | Suggest-then-approve: partner submits; admin approves/rejects; history never deleted |

See [`specs/001-product-vision-scope/research.md`](specs/001-product-vision-scope/research.md) for
full rationale on every dependency choice.

---

## Documentation

| Document | Description |
|----------|-------------|
| [`specs/001-product-vision-scope/spec.md`](specs/001-product-vision-scope/spec.md) | Product vision, personas, MVP scope, success criteria |
| [`specs/001-product-vision-scope/plan.md`](specs/001-product-vision-scope/plan.md) | Implementation plan, tech context, constitution check, project structure |
| [`specs/001-product-vision-scope/research.md`](specs/001-product-vision-scope/research.md) | Technology decisions with rationale |
| [`specs/001-product-vision-scope/data-model.md`](specs/001-product-vision-scope/data-model.md) | PostgreSQL schema for all 11 entities |
| [`specs/001-product-vision-scope/contracts/api.md`](specs/001-product-vision-scope/contracts/api.md) | Full REST API contract (endpoints, request/response shapes) |
| [`specs/001-product-vision-scope/quickstart.md`](specs/001-product-vision-scope/quickstart.md) | Local setup and end-to-end validation scenarios |
| [`specs/001-product-vision-scope/tasks.md`](specs/001-product-vision-scope/tasks.md) | 81 dependency-ordered implementation tasks across 7 phases |
| [`docs/functional-requirements.md`](docs/functional-requirements.md) | Normative functional requirements |
| [`docs/coding-guidelines.md`](docs/coding-guidelines.md) | Go and TypeScript formatting and style rules |
| [`docs/testing-guidelines.md`](docs/testing-guidelines.md) | Three-layer testing strategy and coverage targets |
| [`docs/ui-guidelines.md`](docs/ui-guidelines.md) | Design tokens, Atomic Design layers, accessibility rules |