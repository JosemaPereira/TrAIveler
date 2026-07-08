# e2e

> Playwright end-to-end test suite for TrAIveler — validates complete user flows through a real browser.

This directory contains E2E test specs and shared helpers. Tests run against a fully started
local environment (backend + frontend + database) and exercise the application the same way a
real user would.

[← Back to root README](../README.md) | [Testing Guidelines](../docs/testing-guidelines.md) | [Quickstart](../specs/001-product-vision-scope/quickstart.md)

---

> **Implementation Status**: ✅ **Sprint 1 Complete** (2026-07-08)  
> Infrastructure setup complete: Playwright installed, browsers configured, sample tests passing.  
> Test specs will be authored alongside feature implementation starting in Sprint 2.

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

Playwright is installed as a standalone package in this directory with its own `node_modules`.
All test commands run from the `e2e/` directory.

---

## Project Structure

```
e2e/tests/                            # Test specification files
│   └── sample.spec.ts                # Setup verification test (placeholder)
├── fixtures/                         # Test data and fixtures (empty - for future use)
├── playwright.config.ts              # Playwright configuration
├── package.json                      # Dependencies and npm scripts
├── .gitignore                        # Excludes node_modules, test artifacts
└── README.md                         # This file

Future structure (as features are implemented):
├── tests/
│   ├── generate-itinerary.spec.ts        # US1: register → subscribe → generate
│   ├── travel-style-personalization.spec.ts  # US2: gastronomy style preferences
│   ├── enrich-existing-plan.spec.ts      # US3: anchor places integration
│   ├── suggestion-approval-flow.spec.ts  # US4: admin/partner collaboration
│   ├── health-check.spec.ts              # NFR-OBS: /healthz validation
│   └── privacy-policy.spec.ts            # NFR-PRIV-003: privacy policy link
└── fixtures/Version | Check Command |
|-------------|---------|---------------|
| Node.js | ≥ 20 | `node --version` |
| npm | ≥ 9 | `npm --version` |

For running tests against the application (future):

| Requirement | Notes |
|-------------|-------|
| Backend running | `http://localhost:8080` — see [backend/README.md](../backend/README.md) |
| Frontend dev server | `http://localhost:5173` — `cd frontend && npm run dev` |
| Database ready | Migrations applied: `cd backend && goose -dir migrations postgres "$DATABASE_URL" up` |
| Environment variables | Backend `.env` with `DATABASE_URL`, `ANTHROPIC_API_KEY`, etc. |

## Setup

```bash
# Navigate to e2e directory
cd e2e

# Install dependencies
npm install

# Install Playwright browsers (Chromium, Firefox, WebKit)
# This downloads ~350MB of browser binaries to ~/.cache/ms-playwright/
npx playwrighrun from the `e2e/` directory:

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

# View HTML report from last run
npm run report
```
Configuration

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

E2E tests do not require their own `.env` file. They connect to the running backend/frontend which already have their environment configured.

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
5. Follow testing guidelines in [docs/testing-guidelines.md](../docs/testing-guidelines.md)

## Related Documentation

- [Testing Guidelines](../docs/testing-guidelines.md) — Project-wide testing strategy
- [Product Vision](../docs/product-vision.md) — User stories and acceptance criteria
- [Quickstart](../specs/001-product-vision-scope/quickstart.md) — End-to-end validation scenarios
- [Playwright Documentation](https://playwright.dev/docs/intro) — Official Playwright docs

---

## 
### Advanced Usage

```bash
# Run with different browser
npx playwright test --project=firefox

# Run with multiple workers (when tests are isolated)
npx playwright test --workers=4

# Update snapshots (for visual regression tests)
npx playwright test --update-snapshots

# Run only failed tests from previous run
npx playwright test --last-failed

# Generate code from browser interactions (test recorder)
npx playwright codegen http://localhost:5173//localhost:8080` — see [`backend/README.md`](../backend/README.md) |
| Frontend dev server running | `http://localhost:5173` — `cd frontend && npm run dev` |
| Database seeded | Migrations must have run: `cd backend && go run ./cmd/migrate up` |
| `.env` populated | `ANTHROPIC_API_KEY` and all required backend vars set |

---

## Running Tests

All commands are run from the `frontend/` directory (where Playwright is installed):

```bash
cd frontend

# Run the full E2E suite
npx playwright test

# Run a specific spec file
npx playwright test ../e2e/generate-itinerary.spec.ts

# Run only accessibility-tagged assertions (fast feedback)
npx playwright test --grep @accessibility

# Run with browser visible (headed mode — useful for debugging)
npx playwright test --headed

# Open the Playwright interactive UI
npx playwright test --ui

# View the HTML test report after a run
npx playwright show-report
```

---

## Validation Scenarios

The specs in this directory correspond to the quickstart scenarios documented in
[`specs/001-product-vision-scope/quickstart.md`](../specs/001-product-vision-scope/quickstart.md).
Refer to that document for the exact step-by-step user flows and expected outcomes.

| Spec file | Quickstart scenario | Story | Priority |
|-----------|---------------------|-------|----------|
| `generate-itinerary.spec.ts` | Scenario 1 + 2 | US1 | P1 🎯 MVP |
| `enrich-existing-plan.spec.ts` | Scenario 3 | US3 | P2 |
| `travel-style-personalization.spec.ts` | Scenario 4 | US2 | P2 |
| `suggestion-approval-flow.spec.ts` | Scenario 5 + 6 | US4 | P3 |
| `health-check.spec.ts` | NFR Scenario 2 (observability) | — | P1 |
| `privacy-policy.spec.ts` | NFR Scenario 12 (PRIV-003) | — | P1 |

---

## Accessibility Testing

Every spec imports the shared accessibility helper from
`frontend/tests/helpers/a11y.ts` and calls `checkPageA11y(page)` after navigating to each
page. Tests tagged `@accessibility` run as a subset of the full suite in the CI
`accessibility.yml` workflow.

```ts
// Example — annotated with @accessibility for CI filtering
test('@accessibility privacy policy page has no violations', async ({ page }) => {
  await page.goto('/privacy-policy');
  await checkPageA11y(page);   // throws if any WCAG 2.1 AA violation is found
});
```

The accessibility workflow also runs Lighthouse CI (`lighthouserc.yml` at the repo root) and
asserts score ≥ 90 and Core Web Vitals thresholds. See
[`specs/002-nfr-system-constraints/spec.md`](../specs/002-nfr-system-constraints/spec.md)
(NFR-A11Y) for the full targets.

---

## Writing New Tests

- One spec file per user story or NFR scenario.
- Tag accessibility assertions with `@accessibility` in the test name so CI can filter them.
- Use accessible locators (`getByRole`, `getByLabel`, `getByText`) rather than CSS selectors.
- Do not assert on internal state or DOM structure — assert on what the user sees and can do.
- Clean up any created test data at the end of each test (use `afterEach` / `afterAll`).
- Never make real API calls to external services in tests; mock the AI provider if needed.

See [`docs/testing-guidelines.md`](../docs/testing-guidelines.md) for naming conventions and
additional E2E guidance.

---

## Related Specifications

| Document | Relevance |
|----------|-----------|
| [`specs/001-product-vision-scope/quickstart.md`](../specs/001-product-vision-scope/quickstart.md) | Step-by-step flows and expected outcomes for all scenarios |
| [`specs/001-product-vision-scope/spec.md`](../specs/001-product-vision-scope/spec.md) | Acceptance scenarios this suite validates |
| [`specs/002-nfr-system-constraints/quickstart.md`](../specs/002-nfr-system-constraints/quickstart.md) | NFR validation scenarios (health check, accessibility, privacy policy) |
| [`docs/testing-guidelines.md`](../docs/testing-guidelines.md) | Three-layer strategy, naming conventions, coverage targets |
