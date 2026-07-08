# frontend

> React 19 + TypeScript SPA for TrAIveler — the user interface for AI-powered travel planning.

This directory contains the single-page application: pages, components, state management,
typed API client, accessibility helpers, and unit/integration tests.

[← Back to root README](../README.md) | [Spec 001](../specs/001-product-vision-scope/spec.md) | [UI Guidelines](../docs/ui-guidelines.md)

---

> **Implementation Status**: ✅ Project scaffolding complete (Sprint 1, 2026-07-07)
> - Directory structure with Atomic Design layers created
> - Vite + React 19 + TypeScript strict mode initialized  
> - Core dependencies installed (TanStack Query v5, Zustand, React Router v7, Lucide React)
> - ESLint + Prettier configured with no-any enforcement
> - Development server functional at http://localhost:5173
> 
> ⏳ **Next**: Component library, routing, state management, and API client implementation pending (Sprint 2+)

---

## Responsibility

The frontend is the only client of the backend REST API. Its primary jobs are:

1. **Authentication flows** — register, login, logout, and stub subscription checkout.
2. **Conversational itinerary generation** — multi-turn chat panel with Server-Sent Events
   (SSE) streaming; transitions to the itinerary view when generation is complete.
3. **Trip management** — dashboard, trip detail, day/activity views with loading, error, and
   empty states on every data-dependent component.
4. **Travel style personalisation** — style selector that feeds into the AI prompt.
5. **Collaboration** — real-time suggestion submission (partner) and approve/reject workflow (admin).
6. **Accessibility** — WCAG 2.1 AA compliance enforced by automated axe-core scanning in CI.
7. **Privacy** — privacy policy page linked from the registration flow (GDPR-aware design).

---

## Tech Stack

| Concern | Library / Tool |
|---------|----------------|
| Language | TypeScript (strict mode) |
| UI framework | React 19 |
| Routing | React Router v7 |
| Server state | TanStack Query v5 |
| Client state | Zustand |
| Build tool | Vite |
| Styling | CSS Modules + CSS custom properties (design tokens) |
| Icons | Lucide React |
| Unit/integration tests | Vitest + React Testing Library + MSW |
| E2E + accessibility tests | Playwright + `@axe-core/playwright` |
| Performance auditing | `@lhci/cli` (Lighthouse CI) |
| Linting/formatting | ESLint (strict) + Prettier |

---

## Project Structure

```
frontend/
├── src/
│   ├── features/                     # Feature-scoped components and logic
│   │   ├── auth/
│   │   │   └── store.ts              # Zustand auth store: user, setUser, clearUser
│   │   └── ...
│   ├── components/                   # Atomic Design component layers
│   │   ├── atoms/                    # Smallest reusable UI elements (Button, Input, Badge, PrivacyPolicyLink)
│   │   ├── composites/               # Composed UI units (TripCard, ActivityItem, DaySection, SuggestionBubble)
│   │   └── features/                 # Feature-level UI blocks (ConversationPanel, ItineraryView, SuggestionQueue)
│   ├── pages/                        # Route-level page components
│   │   ├── RegisterPage.tsx          # User registration with privacy policy link
│   │   ├── LoginPage.tsx
│   │   ├── SubscribePage.tsx         # Stub checkout flow
│   │   ├── DashboardPage.tsx         # Trip grid + "New Trip" CTA
│   │   ├── GeneratePage.tsx          # ConversationPanel → ItineraryView flow
│   │   ├── TripPage.tsx              # Trip detail: ItineraryView + collaboration panel
│   │   └── PrivacyPolicyPage.tsx     # Static GDPR privacy policy (NFR-PRIV-003)
│   ├── services/                     # Typed API client functions + TanStack Query hooks
│   │   ├── api.ts                    # apiFetch base: credentials: include, JSON parsing, typed errors
│   │   ├── trips.ts                  # useTrips, useTrip, useSendMessage, etc.
│   │   ├── conversation.ts
│   │   ├── suggestions.ts
│   │   └── collaborators.ts
│   ├── hooks/                        # Shared custom React hooks (prefix: use)
│   ├── styles/
│   │   └── tokens.css                # CSS custom property design tokens (color, spacing, typography)
│   └── App.tsx                       # React Router v7 route declarations and layout wrappers
├── tests/
│   └── helpers/
│       └── a11y.ts                   # checkPageA11y(page): wraps @axe-core/playwright for E2E specs
├── public/
├── index.html
├── vite.config.ts
├── vitest.config.ts                  # Coverage thresholds: ≥ 80% for src/components/ and src/hooks/
├── tsconfig.json                     # TypeScript strict mode
├── .eslintrc.json
└── .prettierrc
```

---

## Prerequisites

| Tool | Version | Check |
|------|---------|-------|
| Node.js | ≥ 20 LTS | `node --version` |
| npm | ≥ 10 | `npm --version` |

---

## Environment Variables

Create a `.env.local` file (gitignored) at `frontend/`. **Never commit real values.**

| Variable | Required | Description |
|----------|----------|-------------|
| `VITE_API_BASE_URL` | ✅ | Backend API base, e.g. `http://localhost:8080/api/v1` |

---

## Setup

```bash
cd frontend
npm install
```

---

## Development Server

```bash
npm run dev
```

Available at `http://localhost:5173`. The dev server proxies API requests to `VITE_API_BASE_URL`.
Ensure the backend is running first — see [`backend/README.md`](../backend/README.md).

---

## Running Tests

```bash
# Unit and integration tests
npm test

# Unit tests with coverage report (must report ≥ 80% for src/components/ and src/hooks/)
npm run test:coverage

# Watch mode during development
npm run test -- --watch
```

---

## Linting and Formatting

```bash
# Lint (must report zero errors before merge)
npm run lint

# Format all files in place
npx prettier --write src/

# Type check without emitting
npx tsc --noEmit
```

---

## Build

```bash
npm run build        # outputs to dist/
npm run preview      # serve the production build locally for Lighthouse auditing
```

---

## Component Architecture — Atomic Design

Components are organised into three layers. Importing from a lower layer into a higher one is
always permitted; the reverse is forbidden.

| Layer | Location | Description |
|-------|----------|-------------|
| Atoms | `src/components/atoms/` | Single-responsibility UI primitives: `Button`, `Input`, `Label`, `Badge`, `PrivacyPolicyLink` |
| Composites | `src/components/composites/` | Reusable combinations of atoms: `TripCard`, `ActivityItem`, `DaySection`, `TravelStyleSelector`, `SuggestionBubble`, `CollaboratorInvite` |
| Features | `src/components/features/` | Domain-specific, data-connected blocks: `ConversationPanel`, `ItineraryView`, `SuggestionQueue` |

Every data-dependent component must explicitly handle **Loading**, **Error**, and **Empty** states.

---

## State Management

| Concern | Tool | Location |
|---------|------|----------|
| Server state (fetching, caching, mutations) | TanStack Query v5 | `src/services/` |
| Client-side auth session | Zustand | `src/features/auth/store.ts` |
| URL / navigation state | React Router v7 | `src/App.tsx` |

---

## Design Tokens

All components must reference CSS custom properties defined in `src/styles/tokens.css`. Hard-coded
colour, spacing, or typography values are **forbidden** — they break theme consistency and will be
flagged by code review.

---

## Accessibility

- Target: **WCAG 2.1 AA** compliance on all pages.
- All interactive elements must be keyboard-operable with a visible focus indicator.
- All images must have meaningful `alt` attributes. All form fields must have associated labels.
- Touch targets must be ≥ 44 × 44 px on mobile.
- Automated scanning via `@axe-core/playwright` runs in every E2E test tagged `@accessibility`.
- Lighthouse CI asserts accessibility score ≥ 90 and Core Web Vitals thresholds on every PR.

See [`specs/002-nfr-system-constraints/spec.md`](../specs/002-nfr-system-constraints/spec.md)
(NFR-A11Y section) for full accessibility targets.

---

## Code Standards

- All code, identifiers, and comments must be in **English** — see [coding guidelines](../docs/coding-guidelines.md).
- TypeScript strict mode is required. `any` is forbidden; use explicit types or `unknown` with guards.
- One component per file. File name matches component name in PascalCase.
- Custom hooks live in `src/hooks/` and must start with `use`.
- Import order: React → third-party → internal → styles/assets.

---

## Continuous Integration

The frontend CI pipeline runs automatically on every pull request and push to main that modifies frontend code.

**Workflow**: [`.github/workflows/frontend-ci.yml`](../.github/workflows/frontend-ci.yml)

**Triggers**:
- Pull requests modifying `frontend/**`
- Push to `main` branch modifying `frontend/**`
- Manual workflow dispatch

**Jobs**:

1. **Lint** — Runs ESLint with strict TypeScript rules (no-explicit-any enforced as error) and Prettier formatting check. Must pass with zero errors and consistent formatting before merge.

2. **Test** — Executes Vitest unit and component tests with coverage reporting (`npm test -- --coverage --run`). Coverage artifact uploaded for review.

3. **Build** — Builds production bundle with Vite, reports bundle size, and uploads `dist/` artifact. Validates that the production build completes successfully without errors.

4. **Accessibility** — Placeholder job for Lighthouse CI accessibility audit (WCAG 2.1 AA). Full implementation scheduled for Sprint 2 with `@lhci/cli` and `@axe-core/playwright` integration.

**Future Enhancements** (TODO comments in workflow):
- **Sprint 2**: Full Lighthouse CI with WCAG 2.1 AA compliance checks, performance audits, and Core Web Vitals thresholds
- **Sprint 10**: S3 + CloudFront deployment job with cache invalidation

**Local Equivalent**:

Run the same checks locally before pushing:

```bash
# Lint and format
npm run lint
npx prettier --check "src/**/*.{ts,tsx,css}"

# Test with coverage
npm test -- --coverage --run

# Build production bundle
npm run build
```

---

## Related Specifications

| Document | Relevance |
|----------|-----------|
| [`specs/001-product-vision-scope/spec.md`](../specs/001-product-vision-scope/spec.md) | User stories and acceptance criteria |
| [`specs/001-product-vision-scope/contracts/api.md`](../specs/001-product-vision-scope/contracts/api.md) | Full REST API contract consumed by this SPA |
| [`specs/002-nfr-system-constraints/spec.md`](../specs/002-nfr-system-constraints/spec.md) | Accessibility, performance, and privacy NFRs |
| [`docs/ui-guidelines.md`](../docs/ui-guidelines.md) | Design tokens, responsive breakpoints, loading/error/empty state requirements |
| [`docs/testing-guidelines.md`](../docs/testing-guidelines.md) | Vitest and Playwright testing conventions |
