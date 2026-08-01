# frontend

> React 19 + TypeScript SPA for TrAIveler — the user interface for AI-powered travel planning.

This directory contains the single-page application: pages, components, state management,
typed API client, accessibility helpers, and unit/integration tests.

[← Back to root README](../README.md) | [Spec 001](../specs/001-product-vision-scope/spec.md) | [UI Guidelines](../docs/ui-guidelines.md)

---

> **Implementation Status**: ✅ Sprint 6 complete (closed 2026-08-01) — frontend application
> structure (Spec 005 Phase 4) shipped in full in Sprint 2, across issues #59–#66 (PRs #79–#82).
> Sprint 3 (Weeks 5–6) was mainly Terraform infrastructure modules, but also pulled the
> accessibility CI gate forward (issue #93, `G-SPRINT3-A11Y-CI`, tasks 002-T002/T004/T022/T024 —
> see "Accessibility" and "Continuous Integration" below). Sprint 6 (issues #171/#180/#182/#183/#184,
> PRs #194/#196/#198/#203, 2026-07-30–08-01) then shipped the feature-scoped frontend work this
> banner used to describe as future: real auth pages, navigation, a dashboard shell, and a
> redirect-aware route guard, wired against the backend auth vertical that landed the same sprint
> (see [`backend/README.md`](../backend/README.md)).
> - ✅ Sprint 6 (issues #171/#180, PR #194): auth API/hooks plumbing — `src/features/auth/`
>   (`services/authApi.ts`; `types.ts` for the wire shapes; `validation.ts`, mirroring the backend's
>   password rules rule-for-rule; `hooks/useRegister.ts`/`useLogin.ts`/`useLogout.ts`, each a
>   TanStack Query mutation around `authApi`) and the `Label` primitive. `src/lib/api-client.ts`
>   gained a single-flight 401→refresh→retry interceptor reached through a handler registry
>   (`session-expiry.ts`'s `installSessionExpiryHandler`/`handleSessionExpired`, wired from
>   `main.tsx` before first render), not a direct import, so the client stays store-agnostic;
>   `src/stores/auth-store.ts`'s `User` type was corrected to the real wire shape (`full_name`,
>   `has_subscription`; no `role`/`subscription_id` — the backend never serializes those).
> - ✅ Sprint 6 (issue #195, PR #196): `GET /auth/me` session bootstrap — `useAuthStore.refreshSession()`
>   calls it to re-derive session state on reload (the HTTP-only cookie is the real store; nothing is
>   rehydrated from local storage). A 401 here is a normal negative answer, not session death, so it is
>   exempted from the interceptor's teardown.
> - ✅ Sprint 6 (issues #182/#183, PR #198): real `/login` and `/register` pages —
>   `src/routes/LoginPage.tsx`/`RegisterPage.tsx` rendering `features/auth/components/LoginForm.tsx`/
>   `RegisterForm.tsx`. `src/routes/login-redirect.ts`'s `resolveLoginRedirect` allow-lists only a
>   same-origin `?redirect=` target (a single leading `/`, rejecting `//host`/`/\host`) before
>   `LoginPage` navigates there on success, closing an open-redirect gap.
> - ✅ Sprint 6 (issue #184, PR #203, closes G-008-AUTH-SHELL 008-T120/T121/T122/T069):
>   `src/components/ProtectedRoute.tsx` — redirects an unauthenticated visitor to
>   `/login?redirect=<attempted-path>` (replacing history) instead of rendering the guarded
>   `<Outlet />`; `src/components/composites/Navigation.tsx` — user name, subscription badge, log
>   out button, and a responsive hamburger menu; `src/routes/DashboardPage.tsx` — wires `Navigation`
>   plus a welcome message and an `EmptyState` placeholder (real trip list deferred to Sprint 8);
>   `auth-store.ts` gained a `setUser` action to refresh user data without touching
>   `isAuthenticated`. `src/routes/index.tsx`'s route table now nests `dashboard`, `trips/:id`, and
>   `settings` under `ProtectedRoute`, alongside the public `login`/`register`/`password-reset`
>   routes and the index `HomePage`. `PasswordResetPage.tsx`, `SettingsPage.tsx`, and
>   `TripDetailPage.tsx` are still placeholder pages scaffolding their routes — real content is later
>   Sprint 8+ work. `src/lib/auth.ts` and `src/lib/authContext.tsx` remain the original Spec-004
>   scaffolding placeholders (`export {}` / an untyped `createContext`) — session state actually
>   lives in `stores/auth-store.ts` and `features/auth/`, not in these two files.
> - ✅ Sprint 3 (issue #93): accessibility CI gate — `@axe-core/playwright` + `@lhci/cli` dev
>   dependencies, root `lighthouserc.yml`, `checkPageA11y(page)` helper (`frontend/tests/helpers/a11y.ts`),
>   dedicated `Accessibility Audit` job in `.github/workflows/frontend-ci.yml`. Not yet exercising real pages — no E2E spec is
>   tagged `@accessibility` (002-T023, Sprint 9); the workflow bridges the gap with
>   `--pass-with-no-tests` in the meantime.
> - ✅ Sprint 1 (2026-07-07): project scaffolding — Vite + React 19 + TypeScript strict mode, Atomic
>   Design directories, core dependencies (TanStack Query v5, Zustand, React Router v7, Lucide
>   React), ESLint + Prettier configured, dev server functional at http://localhost:5173
> - ✅ Sprint 2 (005-T042/T043, issue #59): design system tokens — `src/styles/tokens.css` +
>   `src/styles/global.css`, imported once in `src/main.tsx`
> - ✅ Test tooling: Vitest + React Testing Library + jest-dom + MSW installed, `npm test` runs a
>   passing smoke test (`src/App.test.tsx`). Coverage threshold enforcement (originally roadmap
>   002-T041) shipped later — see "Running Tests" below: `vitest.config.ts` now enforces ≥90%.
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
> - ✅ Sprint 4 (005-T115/T116, issue #118): frontend integration patterns — `useErrorHandler`
>   (`src/hooks/useErrorHandler.ts`), the **first hook** under `src/hooks/`, maps an `unknown` error
>   (typically a TanStack Query `error` field) to render-ready `{ title, message, requestId?,
>   isRetryable }`, redirecting to `/login` on a 401 as a `useEffect` side effect (React Router
>   forbids navigating during render). `ErrorMessage` gained an optional `requestId` prop that
>   renders a de-emphasized "Reference ID" line so a user can read a correlation ID off to support.
>   005-T114 (sending `X-Request-ID` and parsing `requestId` out of error envelopes in
>   `src/lib/api-client.ts`) shipped earlier, in Sprint 2 (issue #81). Still not wired into any real
>   page as of Sprint 6: the auth pages that now do consume live API data (`LoginForm`/`RegisterForm`)
>   render their errors via `getErrorMessage`/`getFieldErrors`/`isAPIError` from
>   `src/lib/query-client.ts` directly rather than through this hook, so nothing calls
>   `useErrorHandler` or renders `ErrorMessage` with a live `requestId` today; adopting the intended
>   pattern (see "Error Handling" below) is still open follow-up work.
>
> Spec 005 Phase 4 (frontend application structure) is now complete. The "Project Structure" and
> "Tech Stack" sections below still contain a **Target** subsection for work planned in later
> specs (e.g. feature-scoped `routes/`, per-resource query hooks, more `hooks/` entries) — only what's
> marked ✅ exists in the codebase today.

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
| Icons | Lucide React | ✅ installed; in use — e.g. `AlertCircle` in `src/components/primitives/ErrorMessage.tsx` |
| Linting/formatting | ESLint (strict) + Prettier | ✅ installed |
| Unit/integration tests | Vitest + React Testing Library + MSW | ✅ installed; **MSW wired up** (issue #180) — shared server in `src/test/msw/`, lifecycle in `src/test/setup.ts`, `onUnhandledRequest: 'error'`. `api-client.test.ts` keeps its direct `fetch` stubs (it is the client's own unit test; a stubbed `fetch` bypasses MSW, so the two styles cannot share a file) |
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
│   │   │   ├── Label.tsx / Label.module.css / Label.test.tsx     # standalone label for composite fields; required indicator rendered twice (aria-hidden `*` + visually-hidden "(required)")
│   │   │   ├── LoadingSpinner.tsx / .module.css / .test.tsx      # role="status"/aria-live loading indicator; visually-hidden text label
│   │   │   ├── ErrorMessage.tsx / .module.css / .test.tsx        # role="alert" error display with an optional retry button
│   │   │   └── EmptyState.tsx / .module.css / .test.tsx          # "nothing to show" state with optional icon and action button
│   │   ├── composites/                # Compositions of primitives
│   │   │   ├── Form.tsx / Form.module.css / Form.test.tsx        # config-driven form built on Button + Input; owns loading/error state
│   │   │   └── Navigation.tsx / .module.css / .test.tsx          # Persistent authenticated-shell nav: user name, subscription badge, log out button, responsive hamburger menu (issue #184, PR #203)
│   │   ├── ProtectedRoute.tsx / .test.tsx                        # Route guard: renders <Outlet /> when useIsAuthenticated() is true, else redirects (replace) to /login?redirect=<attempted-path> (issue #184, PR #203)
│   │   └── ErrorBoundary.tsx / .module.css / .test.tsx           # Top-level class-based error boundary (see "Component Architecture" below); not a primitive or composite — sits outside the Atomic Design layers
│   ├── routes/                       # React Router v7 route configuration
│   │   ├── index.tsx                 # createBrowserRouter config: RootLayout > { public: index/login/register/password-reset; ProtectedRoute > dashboard/trips/:id/settings } (issues #182/#183/#184)
│   │   ├── RootLayout.tsx            # Shared layout; renders <Outlet /> (future: header/footer chrome — Navigation is rendered per-page today, e.g. by DashboardPage, not here)
│   │   ├── HomePage.tsx              # Placeholder index-route page
│   │   ├── LoginPage.tsx             # Renders LoginForm; on success navigates to the sanitized ?redirect= target (resolveLoginRedirect), else /dashboard (issue #183, PR #198)
│   │   ├── RegisterPage.tsx          # Renders RegisterForm; on success navigates to /dashboard, replacing history (issue #182, PR #198)
│   │   ├── login-redirect.ts / .test.ts # resolveLoginRedirect(): allow-lists a same-origin ?redirect= path (single leading /), rejecting //host and /\host open-redirect vectors (issue #183, PR #198)
│   │   ├── PasswordResetPage.tsx     # Placeholder scaffolding the public /password-reset route; real reset flow is later-sprint work
│   │   ├── DashboardPage.tsx / .test.tsx # Wires Navigation + a welcome message + an EmptyState placeholder (real trip list lands with TripDashboard/useTrips in Sprint 8) (issue #184, PR #203)
│   │   ├── SettingsPage.tsx          # Placeholder scaffolding the protected /settings route; real account/subscription settings are later-sprint work
│   │   ├── TripDetailPage.tsx        # Placeholder scaffolding the protected /trips/:id route; real itinerary content is later-sprint work
│   │   └── index.test.tsx
│   ├── stores/                       # Zustand client-state stores
│   │   ├── auth-store.ts             # useAuthStore: isAuthenticated/user/isLoading + login/logout/refreshSession/setUser/setLoading; no persistence (session cookie is the source of truth). setUser added issue #184/PR #203 to refresh user data without touching isAuthenticated
│   │   └── auth-store.test.ts
│   ├── features/                     # Feature-scoped components and logic
│   │   └── auth/                     # Auth feature module (issues #171/#180/#182/#183, PRs #194/#198)
│   │       ├── types.ts              # Wire types for /api/v1/auth/*: RegisterRequest/Response, LoginRequest/Response, CurrentUserResponse, Subscription
│   │       ├── validation.ts         # isValidEmail/validatePasswordStrength/isValidFullName — mirrors backend/internal/auth/validator.go rule-for-rule
│   │       ├── session-expiry.ts     # handleSessionExpired/installSessionExpiryHandler: on an unrecoverable 401, clears the auth store and window.location.assigns to /login?redirect=..., wired from main.tsx via a handler registry (not an import) to keep api-client.ts store-agnostic
│   │       ├── services/
│   │       │   └── authApi.ts        # register/login/logout/me built on lib/api-client.ts (credentials: 'include'; tokens ride as HttpOnly cookies, never handled here)
│   │       ├── hooks/
│   │       │   ├── useRegister.ts    # TanStack mutation for POST /auth/register; on success pushes the returned user into the auth store
│   │       │   ├── useLogin.ts       # TanStack mutation for POST /auth/login; also exposes retryAfterSeconds seeded from a 429's Retry-After
│   │       │   └── useLogout.ts      # TanStack mutation for POST /auth/logout; on success clears the store, drops the whole query cache, and navigates to /login
│   │       └── components/
│   │           ├── LoginForm.tsx     # Email/password form; renders a generic "Invalid email or password" on 401 (anti-enumeration) and a live rate-limit countdown on 429
│   │           └── RegisterForm.tsx  # Email/password/full_name form with an optional [DEMO] payment-token checkbox (DEMO_PAYMENT_TOKEN — the real checkout page is a separate, still-Backlog ticket)
│   ├── hooks/                        # Shared custom React hooks (prefix: use)
│   │   ├── useErrorHandler.ts        # Maps an unknown error (e.g. TanStack Query's `error`) to { title, message, requestId?, isRetryable }; redirects to /login on 401
│   │   └── useErrorHandler.test.tsx
│   ├── lib/                          # Framework/infra wiring shared across the app
│   │   ├── api-client.ts             # apiFetch base + api.get/post/put/patch/delete; APIError matching docs/api-design-standards.md §7; silent 401 token_expired → refresh → retry
│   │   ├── api-client.test.ts        # client unit tests; stubs the global `fetch` (see "Testing")
│   │   ├── api-client.refresh.test.ts # 401 → refresh → retry flow; MSW-backed, hence a separate file
│   │   ├── error-handler.ts          # Pure mapApiError(error) → { title, message, requestId?, isRetryable }
│   │   ├── error-handler.test.ts
│   │   ├── query-client.ts           # TanStack queryClient config + isAPIError/getErrorMessage/getFieldErrors/getRetryAfterSeconds helpers
│   │   └── query-client.test.ts
│   ├── styles/
│   │   ├── tokens.css                # Design tokens: color, spacing, typography, radius, shadow, z-index, transitions
│   │   └── global.css                # Imports tokens.css; CSS reset + base element styles
│   ├── test/
│   │   ├── setup.ts                  # Vitest setup: jest-dom matchers, RTL's afterEach(cleanup), and the MSW server lifecycle (listen/resetHandlers/close)
│   │   ├── msw/                      # Shared Mock Service Worker instance (server.ts) + default auth handlers and fixtures (handlers.ts)
│   │   ├── queryWrapper.tsx          # createQueryWrapper(): per-test QueryClientProvider + MemoryRouter wrapper for renderHook/render
│   │   └── cssModule.ts              # Test helper: resolves a CSS Module class into a definite `string` for toHaveClass() assertions
│   ├── App.tsx                       # Root component: ErrorBoundary > QueryClientProvider > RouterProvider
│   ├── App.test.tsx
│   ├── main.tsx                      # React entry point (StrictMode + createRoot); composition root that installs the session-expiry handler before first render
│   └── vite-env.d.ts                 # ImportMetaEnv typing for VITE_* variables
├── tests/
│   └── helpers/
│       └── a11y.ts / a11y.test.ts    # checkPageA11y(page): wraps @axe-core/playwright, throws on any WCAG 2.1 AA violation (issue #93); not yet called from e2e/*.spec.ts (002-T023, Sprint 9)
├── index.html
├── vite.config.ts
├── vitest.config.ts                  # jsdom environment, coverage via v8, thresholds enforced at 90% (statements/branches/functions/lines)
├── tsconfig.json / tsconfig.node.json  # TypeScript strict mode
├── tsconfig.tests.json                # Extends tsconfig.json; includes tests/ so ESLint's typed linting can parse it (src/ build is unaffected)
├── eslint.config.js                  # Flat ESLint config
├── .prettierrc.json
├── .env.test                          # Committed, deterministic VITE_API_BASE_URL for Vitest (see "Environment Variables")
└── package.json
```

### Target (planned, future specs)

Where the codebase is still headed. Everything in "Current" above (`components/primitives/`,
`components/composites/` including `Navigation`, `components/ProtectedRoute.tsx`,
`components/ErrorBoundary.tsx`, `routes/` including the real `login`/`register`/`password-reset`/
`dashboard`/`trips/:id`/`settings` route files, `stores/auth-store.ts`, `features/auth/`,
`hooks/useErrorHandler.ts`, `lib/`, `test/msw/`, `tests/helpers/a11y.ts`, and the wired-up `App.tsx`)
is already real — everything below is still aspirational. Note that some routes above are real
*files* wired into the router but still render placeholder content (`PasswordResetPage`,
`SettingsPage`, `TripDetailPage`, `HomePage`) — see the "Current" tree for which.

```
frontend/
├── src/
│   ├── features/                     # Feature-scoped components and logic (auth/ is real; see Current)
│   │   └── ...                       # e.g. trips/, collaboration/ — added as their specs land
│   ├── components/
│   │   ├── primitives/               # More primitives to add: Badge, PrivacyPolicyLink
│   │   ├── composites/               # More composites to add: TripCard, ActivityItem, DaySection, TravelStyleSelector, SuggestionBubble, CollaboratorInvite
│   │   └── features/                 # Feature-level UI blocks (ConversationPanel, ItineraryView, SuggestionQueue)
│   ├── routes/                       # More routes to add: itinerary-generation, privacy-policy pages; real content still to land inside the existing password-reset/settings/trip/home placeholder pages
│   ├── lib/                          # More TanStack Query hooks to add on top of api-client.ts: useTrips, useTrip, useSendMessage, etc.
│   └── hooks/                        # More shared hooks beyond useErrorHandler, as reusable non-query logic emerges
└── public/
```

Coverage thresholds (≥ 90% statements/branches/functions/lines, repo-wide) **are** enforced via
`thresholds` in `vitest.config.ts` — `npm run test:coverage` and the CI coverage step fail below
that floor. This supersedes the enforcement gap formerly tracked as roadmap task 002-T041; see
[`docs/testing-guidelines.md`](../docs/testing-guidelines.md#coverage-targets) for the full history
of the 80%→90% change.

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
`VITE_API_BASE_URL`. Vite/Vitest auto-load it in `test` mode, so tests get a deterministic value
without depending on a developer's local `.env.local`. The URL is never dereferenced over the
network — requests are answered either by MSW (whose handler URLs are built from this same value,
so the two cannot drift) or by a stubbed global `fetch` — but it still has to be set at all, to
satisfy `api-client.ts`'s required-env-var check. Personal overrides for test mode go in the
git-ignored `.env.test.local`.

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

# Unit tests with a coverage report (≥90% enforced — see "Coverage Targets" in docs/testing-guidelines.md)
npm run test:coverage

# Watch mode during development
npm run test:watch
```

Test files live next to the source file they test (e.g. `App.test.tsx` alongside `App.tsx`), per
[`docs/testing-guidelines.md`](../docs/testing-guidelines.md). MSW intercepts HTTP for feature-level
tests (`src/features/auth/`, `src/lib/api-client.refresh.test.ts`); the shared server lives in
`src/test/msw/` and its lifecycle in `src/test/setup.ts`. `src/lib/api-client.test.ts` still mocks
the global `fetch` directly — deliberately, since it is the unit test for the client itself and a
stubbed `fetch` would bypass MSW entirely. Tests run in Vite's `test` mode, which
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
| Primitives ("atoms") | `src/components/primitives/` | Single-responsibility UI primitives: `Button`, `Input`, `Card` ✅ implemented (issue #63); `LoadingSpinner`, `ErrorMessage`, `EmptyState` ✅ implemented (issue #64); `Label` ✅ implemented (issue #171); `Badge`, `PrivacyPolicyLink` ⏳ planned |
| Composites | `src/components/composites/` | Reusable combinations of primitives: `Form` ✅ implemented (issue #65); `Navigation` ✅ implemented (issue #184); `TripCard`, `ActivityItem`, `DaySection`, `TravelStyleSelector`, `SuggestionBubble`, `CollaboratorInvite` ⏳ planned |
| Features | `src/components/features/` | Domain-specific, data-connected blocks: `ConversationPanel`, `ItineraryView`, `SuggestionQueue` ⏳ planned |

`ErrorBoundary` (`src/components/ErrorBoundary.tsx`) ✅ implemented (issue #62) and `ProtectedRoute`
(`src/components/ProtectedRoute.tsx`) ✅ implemented (issue #184) both sit outside this layering —
single top-level app-shell/routing components, not reusable primitives/composites/features, so they
live directly under `src/components/` rather than in one of the three subdirectories.

Every data-dependent component must explicitly handle **Loading**, **Error**, and **Empty** states.

---

## State Management

| Concern | Tool | Location |
|---------|------|----------|
| Server state (fetching, caching, mutations) | TanStack Query v5 | `src/lib/query-client.ts` (config) ✅; per-resource query/mutation hooks ⏳ planned in `src/lib/` |
| Client-side auth session | Zustand | `src/stores/auth-store.ts` ✅ implemented (issue #61) |
| URL / navigation state | React Router v7 | `src/routes/` (router config) ✅, mounted via `RouterProvider` in `src/App.tsx` |

---

## Error Handling

Three pieces compose into the intended error-handling pattern for any future data-fetching
component: `APIError` (thrown at the fetch layer), `useErrorHandler` (turns it into display-ready
info), and `ErrorMessage` (renders that info). **Note (updated Sprint 6)**: the auth pages now make
live API calls (`LoginForm`/`RegisterForm` via `useLogin`/`useRegister`), but they render errors via
`getErrorMessage`/`getFieldErrors`/`isAPIError` from `src/lib/query-client.ts` directly rather than
through this hook — so `useErrorHandler`/`ErrorMessage` still aren't exercised anywhere in the
running app today. This section documents the intended pattern for future feature pages to adopt.

### `APIError` (`src/lib/api-client.ts`)

`apiFetch` (and the `api.get`/`post`/`put`/`patch`/`delete` helpers built on it) throws `APIError`
for any non-2xx response — callers never have to branch on a bare `Response`/`Error`:

```ts
export class APIError extends Error {
  readonly status: number
  readonly code: string
  readonly requestId: string
  readonly fields?: APIErrorField[]
  // ...
}
```

`status`/`code`/`message`/`requestId` (and, for 422 validation failures, `fields`) are parsed from
the backend's standard error envelope (`{error, message, request_id, fields?}` —
[`docs/api-design-standards.md`](../docs/api-design-standards.md) §7, the same envelope
`backend/README.md`'s [error handling section](../backend/README.md#error-handling-internalerrors)
documents from the server side). If the response body isn't a parseable envelope, `APIError` still
gets built, with `code: 'unknown_error'` and `requestId: ''`.

Every request also gets a fresh, client-generated `X-Request-ID` header via `crypto.randomUUID()`:

```ts
headers.set('X-Request-ID', crypto.randomUUID())
```

This is **not** the same ID as `APIError.requestId` — the header above is generated per outgoing
request before the server ever sees it, while `requestId` on a thrown `APIError` is whatever the
server's own response body/`X-Request-ID` response header echoed back (the backend's `RequestID`
middleware reuses an incoming header when present, per the correlation-ID pattern described in the
backend section linked above). In the normal case they're the same value round-tripped; only a
proxy/gateway rewriting the header in transit would make them diverge.

### `useErrorHandler` (`src/hooks/useErrorHandler.ts`)

The first hook in `src/hooks/`. Turns an `unknown` value — typically a TanStack Query `error`
field — into render-ready display info, or `null` when there's nothing to show:

```ts
export interface ErrorHandlerResult {
  title: string
  message: string
  requestId?: string
  isRetryable: boolean
}

export function useErrorHandler(error: unknown): ErrorHandlerResult | null
```

It switches on `APIError.status` (via the existing `isAPIError` guard from
`src/lib/query-client.ts`, not a fresh `instanceof` check):

- **401** — returns non-retryable "Session Expired" info, and separately triggers a redirect to
  `/login` via `useNavigate()`. The redirect runs inside a `useEffect`, not during render, because
  React Router forbids calling `navigate()` while rendering.
- **403** / **404** — non-retryable "Permission Denied" / "Not Found" info.
- **anything else** (500/503, any other 5xx, or a non-`APIError` failure such as a network error) —
  a generic, **retryable** "Something Went Wrong" result.

`requestId` is only set when `APIError.requestId` is a non-empty string — an empty string (a real
possible value, not just `undefined`) is treated as "no id."

Minimal usage:

```tsx
const errorInfo = useErrorHandler(error)
if (!errorInfo) return null
```

### `ErrorMessage` (`src/components/primitives/ErrorMessage.tsx`)

Renders an `ErrorHandlerResult` (or any equivalent props) as an alert-role error state:

```tsx
<ErrorMessage
  title={errorInfo.title}
  message={errorInfo.message}
  requestId={errorInfo.requestId}
  onRetry={refetch}
/>
```

`requestId?: string` is optional and only renders a de-emphasized "Reference ID: ..." line
(`data-testid="error-request-id"`) when it's truthy — an empty string from `APIError.requestId`
renders nothing, not a blank line. `onRetry` is likewise optional, for error states with no
meaningful retry action (e.g. a permanent 403).

### End-to-end pattern

The intended full chain for any future data-fetching component — a TanStack Query hook's `error` →
`useErrorHandler` → `ErrorMessage`:

```tsx
function TripView({ tripId }: { tripId: string }) {
  const { data, error, refetch } = useQuery({
    queryKey: ['trip', tripId],
    queryFn: () => api.get<Trip>(`/api/v1/trips/${tripId}`),
  })

  const errorInfo = useErrorHandler(error)
  if (errorInfo) {
    return (
      <ErrorMessage
        title={errorInfo.title}
        message={errorInfo.message}
        requestId={errorInfo.requestId}
        onRetry={errorInfo.isRetryable ? refetch : undefined}
      />
    )
  }

  // ...render data
}
```

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
- Lighthouse CI asserts accessibility score ≥ 90 and Core Web Vitals thresholds on every PR via the `Accessibility Audit` job in [`.github/workflows/frontend-ci.yml`](../.github/workflows/frontend-ci.yml) — see "Continuous Integration" below for its current bridge state.

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
the `Accessibility Audit` job in `frontend-ci.yml` (WCAG 2.1 AA + Core Web Vitals gate, issue #93;
folded in from the former standalone `accessibility.yml` on 2026-07-16).

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

**Workflow**: the `Accessibility Audit` job in [`.github/workflows/frontend-ci.yml`](../.github/workflows/frontend-ci.yml) (issue #93, `docs/roadmap.md` tasks 002-T002/T004/T022/T024; standalone `accessibility.yml` until 2026-07-16)

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
