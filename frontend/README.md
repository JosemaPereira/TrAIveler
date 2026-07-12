# frontend

> React 19 + TypeScript SPA for TrAIveler — the user interface for AI-powered travel planning.

This directory contains the single-page application: pages, components, state management,
typed API client, accessibility helpers, and unit/integration tests.

[← Back to root README](../README.md) | [Spec 001](../specs/001-product-vision-scope/spec.md) | [UI Guidelines](../docs/ui-guidelines.md)

---

> **Implementation Status**: ✅ Sprint 2 complete (closed 2026-07-11) — frontend application
> structure (Spec 005 Phase 4) shipped in full, across issues #59–#66 (PRs #79–#82). Sprint 3
> (Weeks 5–6) is mainly Terraform infrastructure modules, but also pulled the accessibility CI gate
> forward (issue #93, `G-SPRINT3-A11Y-CI`, tasks 002-T002/T004/T022/T024 — see "Accessibility" and
> "Continuous Integration" below). The next feature-scoped frontend work (auth pages, trip
> dashboard, feature components) lands in later sprints per [`docs/roadmap.md`](../docs/roadmap.md).
> - ✅ Sprint 3 (issue #93): accessibility CI gate — `@axe-core/playwright` + `@lhci/cli` dev
>   dependencies, root `lighthouserc.yml`, `checkPageA11y(page)` helper (`frontend/tests/helpers/a11y.ts`),
>   dedicated `.github/workflows/accessibility.yml`. Not yet exercising real pages — no E2E spec is
>   tagged `@accessibility` (002-T023, Sprint 9); the workflow bridges the gap with
>   `--pass-with-no-tests` in the meantime.
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
> - ✅ Sprint 2 (005-T044/T045, issue #60): API client and query infrastructure — `src/lib/api-client.ts`
>   (`apiFetch`, `api.get/post/put/patch/delete`, `APIError`) and `src/lib/query-client.ts`
>   (TanStack `queryClient` + `isAPIError`/`getErrorMessage`/`getFieldErrors` helpers)
> - ✅ Sprint 2 (005-T056–T058, issue #62): `src/features/` placeholder directory for future feature
>   modules, `ErrorBoundary` top-level component, and React Router v7 config (`src/routes/`:
>   `index.tsx`, `RootLayout`, placeholder `HomePage`)
> - ✅ Sprint 2 (005-T047/T059/T060, issue #66): `App.tsx` now wires the real provider stack
>   (`ErrorBoundary` > `QueryClientProvider` > `RouterProvider`), replacing the static placeholder
> - ✅ Sprint 2 (005-T046, issue #61): `useAuthStore` Zustand store — `isAuthenticated`, `user`,
>   `login`, `logout`, `refreshSession`, `setLoading` under `src/stores/auth-store.ts`; no
>   persistence middleware, since the HTTP-only JWT cookie is the actual session store (state is
>   re-derived via `refreshSession()` on reload, not rehydrated from local storage)
> - ✅ Sprint 2 (005-T052–T054, issue #64): state display primitives — `LoadingSpinner`,
>   `ErrorMessage`, `EmptyState` under `src/components/primitives/`, each with a token-driven CSS
>   Module and a co-located Vitest/RTL test file
>
> Spec 005 Phase 4 (frontend application structure) is now complete. The "Project Structure" and
> "Tech Stack" sections below still contain a **Target** subsection for work planned in later
> specs (e.g. `hooks/`, feature-scoped `routes/`, per-resource query hooks) — only what's marked
> ✅ exists in the codebase today.

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
| Routing | React Router v7 (`react-router` package) | ✅ wired up: `createBrowserRouter` config in `src/routes/index.tsx`, mounted via `RouterProvider` in `App.tsx` (issue #62) |
| Server state | TanStack Query v5 | ✅ wired up: `queryClient` in `src/lib/query-client.ts`, mounted via `QueryClientProvider` in `App.tsx` (issue #60) |
| API client | Native `fetch` wrapper | ✅ `src/lib/api-client.ts` — `apiFetch`/`api.*` helpers, `APIError` matching the `docs/api-design-standards.md` §7 error envelope (issue #60) |
| Client state | Zustand | ✅ wired up: `useAuthStore` in `src/stores/auth-store.ts` (issue #61) |
| Build tool | Vite | ✅ installed |
| Styling | CSS Modules + CSS custom properties (design tokens) | ✅ tokens/global styles in place; CSS Modules in use by `primitives/`, `composites/`, and `ErrorBoundary` (issues #63, #64, #65, #62) |
| Icons | Lucide React | ✅ installed, not yet used |
| Linting/formatting | ESLint (strict) + Prettier | ✅ installed |
| Unit/integration tests | Vitest + React Testing Library + MSW | ✅ installed; MSW still not wired up (tests mock `fetch` directly — see `api-client.test.ts`) |
| E2E + accessibility tests | Playwright (`e2e/`) + `@axe-core/playwright` | ✅ `@axe-core/playwright` installed in both `e2e/` (Sprint 1, issue #24) and `frontend/` (issue #93); `checkPageA11y(page)` helper in `frontend/tests/helpers/a11y.ts` — not yet called from any `e2e/*.spec.ts` (that wiring is 002-T023, Sprint 9) |
| Performance auditing | `@lhci/cli` (Lighthouse CI) | ✅ installed as a `frontend/` dev dependency (issue #93); config at root [`lighthouserc.yml`](../lighthouserc.yml) |

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
│   │   │   ├── Card.tsx / Card.module.css / Card.test.tsx
│   │   │   ├── LoadingSpinner.tsx / .module.css / .test.tsx      # role="status"/aria-live loading indicator; visually-hidden text label
│   │   │   ├── ErrorMessage.tsx / .module.css / .test.tsx        # role="alert" error display with an optional retry button
│   │   │   └── EmptyState.tsx / .module.css / .test.tsx          # "nothing to show" state with optional icon and action button
│   │   ├── composites/                # Compositions of primitives
│   │   │   └── Form.tsx / Form.module.css / Form.test.tsx        # config-driven form built on Button + Input; owns loading/error state
│   │   └── ErrorBoundary.tsx / .module.css / .test.tsx           # Top-level class-based error boundary (see "Component Architecture" below); not a primitive or composite — sits outside the Atomic Design layers
│   ├── routes/                       # React Router v7 route configuration
│   │   ├── index.tsx                 # createBrowserRouter config: RootLayout > index route (HomePage)
│   │   ├── RootLayout.tsx            # Shared layout; renders <Outlet /> (future: header/footer chrome)
│   │   ├── HomePage.tsx              # Placeholder index-route page
│   │   └── index.test.tsx
│   ├── stores/                       # Zustand client-state stores
│   │   ├── auth-store.ts             # useAuthStore: isAuthenticated/user/isLoading + login/logout/refreshSession/setLoading; no persistence (session cookie is the source of truth)
│   │   └── auth-store.test.ts
│   ├── features/                     # Feature-scoped components and logic (empty placeholder; see .gitkeep)
│   │   └── .gitkeep
│   ├── lib/                          # Framework/infra wiring shared across the app
│   │   ├── api-client.ts             # apiFetch base + api.get/post/put/patch/delete; APIError matching docs/api-design-standards.md §7
│   │   ├── api-client.test.ts
│   │   ├── query-client.ts           # TanStack queryClient config + isAPIError/getErrorMessage/getFieldErrors helpers
│   │   └── query-client.test.ts
│   ├── styles/
│   │   ├── tokens.css                # Design tokens: color, spacing, typography, radius, shadow, z-index, transitions
│   │   └── global.css                # Imports tokens.css; CSS reset + base element styles
│   ├── test/
│   │   ├── setup.ts                  # Vitest setup: extends expect with jest-dom matchers, registers RTL's afterEach(cleanup)
│   │   └── cssModule.ts              # Test helper: resolves a CSS Module class into a definite `string` for toHaveClass() assertions
│   ├── App.tsx                       # Root component: ErrorBoundary > QueryClientProvider > RouterProvider
│   ├── App.test.tsx
│   ├── main.tsx                      # React entry point (StrictMode + createRoot)
│   └── vite-env.d.ts                 # ImportMetaEnv typing for VITE_* variables
├── tests/
│   └── helpers/
│       └── a11y.ts / a11y.test.ts    # checkPageA11y(page): wraps @axe-core/playwright, throws on any WCAG 2.1 AA violation (issue #93); not yet called from e2e/*.spec.ts (002-T023, Sprint 9)
├── index.html
├── vite.config.ts
├── vitest.config.ts                  # jsdom environment, coverage via v8 (no enforced threshold yet)
├── tsconfig.json / tsconfig.node.json  # TypeScript strict mode
├── tsconfig.tests.json                # Extends tsconfig.json; includes tests/ so ESLint's typed linting can parse it (src/ build is unaffected)
├── eslint.config.js                  # Flat ESLint config
├── .prettierrc.json
├── .env.test                          # Committed, deterministic VITE_API_BASE_URL for Vitest (see "Environment Variables")
└── package.json
```

### Target (planned, future specs)

Where the codebase is still headed. Everything in "Current" above (`components/primitives/`,
`components/composites/`, `components/ErrorBoundary.tsx`, `routes/`, `stores/auth-store.ts`,
`features/` placeholder, `lib/`, `tests/helpers/a11y.ts`, and the wired-up `App.tsx`) is already
real — everything below is still aspirational.

```
frontend/
├── src/
│   ├── features/                     # Feature-scoped components and logic (currently empty)
│   │   └── ...                       # e.g. auth/, trips/, collaboration/ — added as their specs land
│   ├── components/
│   │   ├── primitives/               # More primitives to add: Label, Badge, PrivacyPolicyLink
│   │   ├── composites/               # More composites to add: TripCard, ActivityItem, DaySection, SuggestionBubble
│   │   └── features/                 # Feature-level UI blocks (ConversationPanel, ItineraryView, SuggestionQueue)
│   ├── routes/                       # More routes to add: register/login/dashboard/generate/trip/privacy-policy pages
│   ├── lib/                          # More TanStack Query hooks to add on top of api-client.ts: useTrips, useTrip, useSendMessage, etc.
│   └── hooks/                        # Shared custom React hooks (prefix: use)
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
| `VITE_API_BASE_URL` | ✅ | Backend API base, e.g. `http://localhost:8080/api/v1`. Read by `src/lib/api-client.ts`; there is no hardcoded fallback, so a missing value throws at request time instead of silently pointing at the wrong backend. |

**`.env.test`** is committed (unlike `.env.local`) and mirrors `.env.example` with a fixed
`VITE_API_BASE_URL`. Vite/Vitest auto-load it in `test` mode, so unit tests get a deterministic
value without depending on a developer's local `.env.local` — tests mock `fetch` directly, so the
URL is never dereferenced over the network, it only needs to satisfy `api-client.ts`'s
required-env-var check. Personal overrides for test mode go in the git-ignored `.env.test.local`.

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
integration tests that intercept HTTP calls but isn't wired up yet — `src/lib/api-client.test.ts`
currently mocks the global `fetch` directly instead. Tests run in Vite's `test` mode, which
auto-loads `.env.test` for a deterministic `VITE_API_BASE_URL` (see "Environment Variables" above).

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
| Primitives ("atoms") | `src/components/primitives/` | Single-responsibility UI primitives: `Button`, `Input`, `Card` ✅ implemented (issue #63); `LoadingSpinner`, `ErrorMessage`, `EmptyState` ✅ implemented (issue #64); `Label`, `Badge`, `PrivacyPolicyLink` ⏳ planned |
| Composites | `src/components/composites/` | Reusable combinations of primitives: `Form` ✅ implemented (issue #65); `TripCard`, `ActivityItem`, `DaySection`, `TravelStyleSelector`, `SuggestionBubble`, `CollaboratorInvite` ⏳ planned |
| Features | `src/components/features/` | Domain-specific, data-connected blocks: `ConversationPanel`, `ItineraryView`, `SuggestionQueue` ⏳ planned |

`ErrorBoundary` (`src/components/ErrorBoundary.tsx`) ✅ implemented (issue #62) sits outside this
layering — it's a single top-level app-shell component, not a reusable primitive/composite/feature
block, so it lives directly under `src/components/` rather than in one of the three subdirectories.

Every data-dependent component must explicitly handle **Loading**, **Error**, and **Empty** states.

---

## State Management

| Concern | Tool | Location |
|---------|------|----------|
| Server state (fetching, caching, mutations) | TanStack Query v5 | `src/lib/query-client.ts` (config) ✅; per-resource query/mutation hooks ⏳ planned in `src/lib/` |
| Client-side auth session | Zustand | `src/stores/auth-store.ts` ✅ implemented (issue #61) |
| URL / navigation state | React Router v7 | `src/routes/` (router config) ✅, mounted via `RouterProvider` in `src/App.tsx` |

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
- Automated scanning via `@axe-core/playwright` (`checkPageA11y(page)`, `frontend/tests/helpers/a11y.ts`) is designed to run in every E2E test tagged `@accessibility` — no spec is tagged yet (002-T023, Sprint 9), so the CI gate below is not yet exercising real pages.
- Lighthouse CI asserts accessibility score ≥ 90 and Core Web Vitals thresholds on every PR via [`.github/workflows/accessibility.yml`](../.github/workflows/accessibility.yml) — see "Continuous Integration" below for its current bridge state.

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

Two workflows cover the frontend: `frontend-ci.yml` (lint/test/build, every PR) and the dedicated
`accessibility.yml` (WCAG 2.1 AA + Core Web Vitals gate, issue #93).

**Workflow**: [`.github/workflows/frontend-ci.yml`](../.github/workflows/frontend-ci.yml)

**Triggers**:
- Pull requests modifying `frontend/**`
- Push to `main` branch modifying `frontend/**`
- Manual workflow dispatch

**Jobs**:

1. **Lint** — Runs ESLint with strict TypeScript rules (no-explicit-any enforced as error) and Prettier formatting check. Must pass with zero errors and consistent formatting before merge.

2. **Test** — Executes Vitest unit and component tests with coverage reporting (`npm test -- --coverage --run`). Coverage artifact uploaded for review.

3. **Build** — Builds production bundle with Vite, reports bundle size, and uploads `dist/` artifact. Validates that the production build completes successfully without errors.

**Future Enhancements** (TODO comments in workflow):
- **Sprint 10**: S3 + CloudFront deployment job with cache invalidation

---

**Workflow**: [`.github/workflows/accessibility.yml`](../.github/workflows/accessibility.yml) (issue #93, `docs/roadmap.md` tasks 002-T002/T004/T022/T024)

**Triggers**: Pull requests (any); manual workflow dispatch. Job itself gates on `frontend/**`, `e2e/**`, or `lighthouserc.yml` changing, via the same always-running/`dorny/paths-filter` pattern as `frontend-ci.yml`.

**Job**: builds and serves the production bundle, then runs (1) `playwright test --grep @accessibility` from `e2e/` and (2) `lhci autorun --config=lighthouserc.yml` from `frontend/`, asserting the thresholds in root [`lighthouserc.yml`](../lighthouserc.yml): accessibility ≥ 0.9, LCP ≤ 2500 ms, CLS ≤ 0.1, and Total Blocking Time ≤ 200 ms as the lab-mode proxy for INP — Lighthouse's `interaction-to-next-paint` audit only supports `timespan` mode with a real recorded interaction, so it can't produce a value in this workflow's standard single-navigation run (see the comment in `lighthouserc.yml` for detail).

> ⚠️ **Known temporary gap**: no `e2e/*.spec.ts` file is tagged `@accessibility` yet — that wiring is
> task 002-T023 (Sprint 9). Until then, the Playwright step runs with `--pass-with-no-tests` so it
> passes trivially instead of failing on "no tests found." This workflow is **not** a required
> branch-protection check yet for that reason — see `docs/roadmap.md`'s Sprint 3 planning note for the
> full follow-up (drop/reassess the flag, consider making it required) once T023 lands.

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
