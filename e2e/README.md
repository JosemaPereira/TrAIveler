# e2e

> Playwright end-to-end test suite for TrAIveler — validates complete user flows through a real browser.

This directory contains E2E test specs and shared helpers. Tests run against a fully started
local environment (backend + frontend + database) and exercise the application the same way a
real user would.

[← Back to root README](../README.md) | [Testing Guidelines](../docs/testing-guidelines.md) | [Quickstart](../specs/001-product-vision-scope/quickstart.md)

---

> **Implementation Status**:
>
> - ✅ Sprint 1 (2026-07-08) — Playwright installed, browsers configured, sample smoke test passing.
> - ✅ Sprint 3 (2026-07-12, issue #93) — `@axe-core/playwright` added as a dependency and the
>   dedicated accessibility CI gate (the `Accessibility Audit` job in
>   `.github/workflows/frontend-ci.yml`) now runs
>   `npx playwright test --grep @accessibility` from this directory on every pull request that
>   touches `frontend/**` or `e2e/**`.
> - 🔲 No spec in this directory is tagged `@accessibility` yet, so that CI step currently passes
>   via `--pass-with-no-tests` rather than actually asserting anything — wiring the shared
>   `checkPageA11y(page)` helper (already implemented at
>   [`frontend/tests/helpers/a11y.ts`](../frontend/tests/helpers/a11y.ts)) into every spec below is
>   tracked as task `002-T023` (Sprint 9). Feature spec files (`generate-itinerary.spec.ts` and the
>   rest of the table in [Validation Scenarios](#validation-scenarios)) are authored alongside their
>   corresponding feature implementation and do not exist yet — only `tests/sample.spec.ts`
>   (setup-verification placeholder) is present today.

---

## Responsibility

The E2E suite validates that the integrated system delivers the behaviour described in
[`specs/001-product-vision-scope/spec.md`](../specs/001-product-vision-scope/spec.md). It is the
final quality gate before a feature is considered shippable. E2E tests are not a substitute for
unit or integration tests — they verify user-visible outcomes, not implementation details.

---

## Tech Stack

| Concern | Tool | Version |
|---------|------|---------|
| Browser automation | Playwright | ^1.61.1 |
| Accessibility scanning | @axe-core/playwright | ^4.12.1 |
| Test runner | Playwright Test (built-in) | — |
| JavaScript runtime | Node.js | ≥ 20 |

Playwright is installed as a standalone package in this directory with its own `node_modules` and
`package-lock.json` (the lockfile is gitignored — CI uses `npm install`, not `npm ci`, for this
directory). All test commands below run from the `e2e/` directory.

---

## Project Structure

Current layout:

```
e2e/
├── tests/
│   └── sample.spec.ts                # Setup verification test (placeholder)
├── fixtures/                         # Test data and fixtures (empty - for future use)
├── playwright.config.ts              # Playwright configuration
├── package.json                      # Dependencies and npm scripts
├── .gitignore                        # Excludes node_modules, test artifacts, package-lock.json
└── README.md                         # This file
```

Planned structure (spec files are authored alongside their corresponding feature, not yet present):

```
e2e/tests/
├── generate-itinerary.spec.ts            # US1: register → subscribe → generate
├── travel-style-personalization.spec.ts  # US2: gastronomy style preferences
├── enrich-existing-plan.spec.ts          # US3: anchor places integration
├── suggestion-approval-flow.spec.ts      # US4: admin/partner collaboration
├── health-check.spec.ts                  # NFR-OBS: /healthz validation
└── privacy-policy.spec.ts                # NFR-PRIV-003: privacy policy link
```

---

## Prerequisites

| Tool | Version | Check Command |
|------|---------|----------------|
| Node.js | ≥ 20 | `node --version` |
| npm | ≥ 9 | `npm --version` |

For running tests against the application (future, once feature specs exist):

| Requirement | Notes |
|-------------|-------|
| Backend running | `http://localhost:8080` — see [backend/README.md](../backend/README.md) |
| Frontend dev server | `http://localhost:5173` — `cd frontend && npm run dev` |
| Database ready | Migrations applied — see [backend/README.md](../backend/README.md) for the exact command |
| Environment variables | Backend `.env` populated — see [backend/README.md](../backend/README.md) |

## Setup

```bash
# Navigate to e2e directory
cd e2e

# Install dependencies
npm install

# Install Playwright browsers (Chromium, Firefox, WebKit)
# This downloads ~350MB of browser binaries to ~/.cache/ms-playwright/
npx playwright install --with-deps chromium
```

## Running Tests

All commands below run from the `e2e/` directory:

```bash
cd e2e

# Run all tests (headless, default)
npm test

# Run with browser visible (headed mode — useful for debugging)
npm run test:headed

# Run in debug mode (step through tests, pause on errors)
npm run test:debug

# Open Playwright interactive UI (visual test explorer)
npm run test:ui

# Run specific test file
npx playwright test tests/sample.spec.ts

# Run tests matching a pattern
npx playwright test --grep "should load"

# Run only accessibility-tagged specs (once any exist — see Implementation Status above)
npx playwright test --grep @accessibility --pass-with-no-tests

# View HTML report from last run
npm run report
```

### Advanced Usage

```bash
# Run with a different browser
npx playwright test --project=firefox

# Run with multiple workers (only once tests are isolated from shared DB state)
npx playwright test --workers=4

# Update snapshots (for visual regression tests)
npx playwright test --update-snapshots

# Run only failed tests from the previous run
npx playwright test --last-failed

# Generate code from browser interactions (test recorder)
npx playwright codegen http://localhost:5173
```

## Configuration

Configuration is centralized in [playwright.config.ts](./playwright.config.ts). Key settings:

| Setting | Value | Rationale |
|---------|-------|-----------|
| `workers` | 1 | Shared database state — tests must run sequentially |
| `retries` | 1 (local), 2 (CI) | Handles transient network/timing issues |
| `fullyParallel` | false | Enforces sequential execution |
| `baseURL` | `http://localhost:5173` | Frontend dev server |
| `screenshot` | `only-on-failure` | Captures visual state on errors |
| `trace` | `retain-on-failure` | Video-like debugging for failed tests |
| `actionTimeout` | 10000ms | Default timeout for clicks, fills, etc. |

**Browser projects:**
- **Local development:** Chromium only (faster feedback)
- **CI:** All browsers (Chromium, Firefox, WebKit) for cross-browser validation

## Environment Variables

E2E tests do not require their own `.env` file. They connect to the running backend/frontend which
already have their environment configured.

For test-specific configuration (future):

| Variable | Purpose | Default |
|----------|---------|---------|
| `E2E_BASE_URL` | Override frontend URL | `http://localhost:5173` |
| `E2E_API_URL` | Override backend URL | `http://localhost:8080` |
| `E2E_HEADED` | Run tests with visible browser | `false` |

## Debugging Failed Tests

When a test fails:

1. **Check the terminal output** — shows the exact assertion that failed
2. **Open the HTML report** — `npm run report` opens detailed results in browser
3. **View screenshots** — saved in `test-results/` for failed tests
4. **Replay traces** — click "View trace" in HTML report for step-by-step replay
5. **Run in headed mode** — `npm run test:headed` to watch the browser
6. **Use debug mode** — `npm run test:debug` to pause execution and step through

## Writing New Tests

When implementing a new feature:

1. Create test file in `tests/` — follow naming: `<feature-name>.spec.ts`
2. Import test helpers:
   ```typescript
   import { test, expect } from '@playwright/test';
   ```
3. Structure tests with `describe` blocks:
   ```typescript
   test.describe('Feature Name', () => {
     test('should do something', async ({ page }) => {
       // Test implementation
     });
   });
   ```
4. Use Playwright locators (prefer role-based):
   ```typescript
   await page.getByRole('button', { name: 'Submit' }).click();
   await page.getByLabel('Email').fill('user@example.com');
   ```
5. Tag accessibility assertions with `@accessibility` in the test name so CI can filter them
   (`--grep @accessibility`), and call the shared
   [`checkPageA11y(page)`](../frontend/tests/helpers/a11y.ts) helper after navigating to the page
   under test (see [Accessibility Testing](#accessibility-testing) below).
6. Follow testing guidelines in [docs/testing-guidelines.md](../docs/testing-guidelines.md)

---

## Accessibility Testing

The `@axe-core/playwright`-based scan helper, `checkPageA11y(page)`, already lives at
[`frontend/tests/helpers/a11y.ts`](../frontend/tests/helpers/a11y.ts) — it wraps an axe-core scan
scoped to the WCAG 2.1 AA tag set and throws (failing the test) if any violation is found. Wiring
it into every spec in this directory, tagged `@accessibility`, is task `002-T023` (Sprint 9) and
has not happened yet, so today's `sample.spec.ts` does not call it.

The CI accessibility gate (the `Accessibility Audit` job in `.github/workflows/frontend-ci.yml`,
originally added as a standalone workflow in issue #93) already runs against
this directory on every pull request touching `frontend/**` or `e2e/**`:

1. Builds and serves the frontend production bundle.
2. Runs `npx playwright test --grep @accessibility --pass-with-no-tests` from `e2e/` — a no-op pass
   until `002-T023` adds tagged specs.
3. Runs Lighthouse CI (`lighthouserc.yml` at the repo root) against the served build, asserting a
   score ≥ 90 and the Core Web Vitals thresholds from `docs/nfrs.md` (NFR-PERF-003, NFR-A11Y-004).

Example of what a tagged spec will look like once `002-T023` lands:

```ts
import { checkPageA11y } from '../../frontend/tests/helpers/a11y';

test('@accessibility privacy policy page has no violations', async ({ page }) => {
  await page.goto('/privacy-policy');
  await checkPageA11y(page);   // throws if any WCAG 2.1 AA violation is found
});
```

See [`specs/002-nfr-system-constraints/spec.md`](../specs/002-nfr-system-constraints/spec.md)
(NFR-A11Y) for the full accessibility targets.

---

## Validation Scenarios

Once authored, the specs in this directory will correspond to the quickstart scenarios documented
in [`specs/001-product-vision-scope/quickstart.md`](../specs/001-product-vision-scope/quickstart.md).
Refer to that document for the exact step-by-step user flows and expected outcomes.

| Spec file (planned) | Quickstart scenario | Story | Priority |
|-----------|---------------------|-------|----------|
| `generate-itinerary.spec.ts` | Scenario 1 + 2 | US1 | P1 🎯 MVP |
| `enrich-existing-plan.spec.ts` | Scenario 3 | US3 | P2 |
| `travel-style-personalization.spec.ts` | Scenario 4 | US2 | P2 |
| `suggestion-approval-flow.spec.ts` | Scenario 5 + 6 | US4 | P3 |
| `health-check.spec.ts` | NFR Scenario 2 (observability) | — | P1 |
| `privacy-policy.spec.ts` | NFR Scenario 12 (PRIV-003) | — | P1 |

---

## Related Documentation

| Document | Relevance |
|----------|-----------|
| [`specs/001-product-vision-scope/quickstart.md`](../specs/001-product-vision-scope/quickstart.md) | Step-by-step flows and expected outcomes for all scenarios |
| [`specs/001-product-vision-scope/spec.md`](../specs/001-product-vision-scope/spec.md) | Acceptance scenarios this suite validates |
| [`specs/002-nfr-system-constraints/quickstart.md`](../specs/002-nfr-system-constraints/quickstart.md) | NFR validation scenarios (health check, accessibility, privacy policy) |
| [`docs/testing-guidelines.md`](../docs/testing-guidelines.md) | Three-layer strategy, naming conventions, coverage targets |
| [Playwright Documentation](https://playwright.dev/docs/intro) | Official Playwright docs |
