# frontend

> React 19 + TypeScript SPA for TrAIveler — the user interface for AI-powered travel planning.

This directory contains the single-page application: pages, components, state management,
typed API client, accessibility helpers, and unit/integration tests.

[← Back to root README](../README.md) | [Spec 001](../specs/001-product-vision-scope/spec.md) | [UI Guidelines](../docs/ui-guidelines.md)

---

> **Implementation Status**: 🔄 Sprint 2 in progress (as of 2026-07-10)
> - ✅ Sprint 1 (2026-07-07): project scaffolding — Vite + React 19 + TypeScript strict mode, Atomic
>   Design directories, core dependencies (TanStack Query v5, Zustand, React Router v7, Lucide
>   React), ESLint + Prettier configured, dev server functional at http://localhost:5173
> - ✅ Sprint 2 (005-T042/T043, issue #59): design system tokens — `src/styles/tokens.css` +
>   `src/styles/global.css`, imported once in `src/main.tsx`
> - ✅ Test tooling: Vitest + React Testing Library + jest-dom + MSW installed, `npm test` runs a
>   passing smoke test (`src/App.test.tsx`); no coverage threshold enforced yet (roadmap 002-T041)
> - ✅ Sprint 2 (005-T048–T051, issue #63): core UI primitives — `Button`, `Input`, `Card` under
>   `src/components/primitives/`, each with a token-driven CSS Module and a co-located Vitest/RTL
>   test file
> - ✅ Sprint 2 (005-T055, issue #65): `Form` composite under `src/components/composites/`,
>   composing `Button` + `Input` with config-driven fields and loading/error state
> - ⏳ Next: state display primitives — `LoadingSpinner`, `ErrorMessage`, `EmptyState` (issue #64),
>   app shell — QueryClient/Router/ErrorBoundary (issue #66)
>
> The "Project Structure" and "Tech Stack" sections below describe the **target architecture**
> once all Sprint 2 tasks land. Only what's marked ✅ above exists in the codebase today.

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

| Concern | Library / Tool | Status |
|---------|----------------|--------|
| Language | TypeScript (strict mode) | ✅ installed |
| UI framework | React 19 | ✅ installed |
| Routing | React Router v7 | ✅ installed, not yet wired up |
| Server state | TanStack Query v5 | ✅ installed, not yet wired up |
| Client state | Zustand | ✅ installed, not yet wired up |
| Build tool | Vite | ✅ installed |
| Styling | CSS Modules + CSS custom properties (design tokens) | ✅ tokens/global styles in place; CSS Modules in use by `primitives/` and `composites/` components (issues #63, #65) |
| Icons | Lucide React | ✅ installed, not yet used |
| Linting/formatting | ESLint (strict) + Prettier | ✅ installed |
| Unit/integration tests | Vitest + React Testing Library + MSW | ✅ installed; MSW not wired up yet (no API client to mock) |
| E2E + accessibility tests | Playwright + `@axe-core/playwright` | ⏳ planned, not yet installed |
| Performance auditing | `@lhci/cli` (Lighthouse CI) | ⏳ planned, not yet installed |

---

## Project Structure

### Current

```
frontend/
├── src/
│   ├── components/                   # Atomic Design component layers (see "Component Architecture" below)
│   │   ├── primitives/                # Token-driven, reusable UI primitives ("atoms")
│   │   │   ├── Button.tsx / Button.module.css / Button.test.tsx
│   │   │   ├── Input.tsx / Input.module.css / Input.test.tsx     # labeled field; wires aria-invalid/aria-describedby to a role="alert" error
│   │   │   └── Card.tsx / Card.module.css / Card.test.tsx
│   │   └── composites/                # Compositions of primitives
│   │       └── Form.tsx / Form.module.css / Form.test.tsx        # config-driven form built on Button + Input; owns loading/error state
│   ├── styles/
│   │   ├── tokens.css                # Design tokens: color, spacing, typography, radius, shadow, z-index, transitions
│   │   └── global.css                # Imports tokens.css; CSS reset + base element styles
│   ├── test/
│   │   ├── setup.ts                  # Vitest setup: extends expect with jest-dom matchers, registers RTL's afterEach(cleanup)
│   │   └── cssModule.ts              # Test helper: resolves a CSS Module class into a definite `string` for toHaveClass() assertions
│   ├── App.tsx                       # Root component (placeholder shell, not yet wired to routing/state)
│   ├── App.test.tsx                  # Smoke test for App
│   ├── main.tsx                      # React entry point (StrictMode + createRoot)
│   └── vite-env.d.ts
├── index.html
├── vite.config.ts
├── vitest.config.ts                  # jsdom environment, coverage via v8 (no enforced threshold yet)
├── tsconfig.json / tsconfig.node.json  # TypeScript strict mode
├── eslint.config.js                  # Flat ESLint config
├── .prettierrc.json
└── package.json
```

### Target (planned, Sprint 2+)

Where the codebase is headed as the remaining Sprint 2 tasks land. `components/primitives/` and
`components/composites/` are already real (see Current, above) — everything below is still
aspirational.

```
frontend/
├── src/
│   ├── features/                     # Feature-scoped components and logic
│   │   ├── auth/
│   │   │   └── store.ts              # Zustand auth store: user, setUser, clearUser
│   │   └── ...
│   ├── components/                   # Atomic Design component layers
│   │   ├── primitives/               # More state-display primitives to add: LoadingSpinner, ErrorMessage, EmptyState (issue #64)
│   │   ├── composites/               # More composites to add: TripCard, ActivityItem, DaySection, SuggestionBubble
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
│   └── App.tsx                       # React Router v7 route declarations and layout wrappers
├── tests/
│   └── helpers/
│       └── a11y.ts                   # checkPageA11y(page): wraps @axe-core/playwright for E2E specs
└── public/
```

Coverage thresholds (≥ 80% for `src/components/` and `src/hooks/`) will be enforced in `vitest.config.ts` once those directories exist (roadmap 002-T041).

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

# Unit tests with a coverage report (no enforced threshold yet — see roadmap 002-T041)
npm run test:coverage

# Watch mode during development
npm run test:watch
```

Test files live next to the source file they test (e.g. `App.test.tsx` alongside `App.tsx`), per
[`docs/testing-guidelines.md`](../docs/testing-guidelines.md). MSW is installed for future
integration tests that intercept HTTP calls but isn't wired up yet — there's no API client to mock
against until `src/services/` exists.

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
| Primitives ("atoms") | `src/components/primitives/` | Single-responsibility UI primitives: `Button`, `Input`, `Card` ✅ implemented (issue #63); `Label`, `Badge`, `PrivacyPolicyLink`, `LoadingSpinner`, `ErrorMessage`, `EmptyState` ⏳ planned (issue #64) |
| Composites | `src/components/composites/` | Reusable combinations of primitives: `Form` ✅ implemented (issue #65); `TripCard`, `ActivityItem`, `DaySection`, `TravelStyleSelector`, `SuggestionBubble`, `CollaboratorInvite` ⏳ planned |
| Features | `src/components/features/` | Domain-specific, data-connected blocks: `ConversationPanel`, `ItineraryView`, `SuggestionQueue` ⏳ planned |

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

All design tokens are defined as CSS custom properties in `src/styles/tokens.css` (color, spacing,
typography, radius, shadow, z-index, transitions — see [`docs/ui-guidelines.md`](../docs/ui-guidelines.md)
for the authoritative values). `src/styles/global.css` imports `tokens.css` and applies the CSS
reset and base element styles; it's imported once in `src/main.tsx`. Components must reference
these custom properties — hard-coded colour, spacing, or typography values are **forbidden** and
will be flagged by code review.

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

4. **Accessibility** — Placeholder job for Lighthouse CI accessibility audit (WCAG 2.1 AA). Full implementation with `@lhci/cli` and `@axe-core/playwright` is not yet scheduled — see `docs/roadmap.md` tasks 002-T002/T004/T024 (currently unscheduled/Sprint 9, not Sprint 2).

**Future Enhancements** (TODO comments in workflow):
- **Not yet scheduled**: Full Lighthouse CI with WCAG 2.1 AA compliance checks, performance audits, and Core Web Vitals thresholds — pending roadmap 002-T002/T004/T024 (re-check at Sprint 3 planning)
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
