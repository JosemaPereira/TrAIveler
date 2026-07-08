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

| Concern | Tool |
|---------|------|
| Browser automation | Playwright |
| Accessibility scanning | `@axe-core/playwright` (via `frontend/tests/helpers/a11y.ts`) |
| Test runner | Playwright Test (built-in) |

Playwright is installed as a dev dependency of the `frontend` package. Run all commands from
`frontend/` or use the scripts below which handle the path correctly.

---

## Project Structure

```
e2e/
├── generate-itinerary.spec.ts        # US1: register → subscribe → generate itinerary (P1 MVP)
├── travel-style-personalization.spec.ts  # US2: gastronomy style → verify food activities
├── enrich-existing-plan.spec.ts      # US3: anchor places → confirm → verify all appear
├── suggestion-approval-flow.spec.ts  # US4: admin invites partner → approve/reject suggestions
├── health-check.spec.ts              # NFR-OBS: /healthz header and schema validation
└── privacy-policy.spec.ts            # NFR-PRIV-003: privacy policy link on registration page
```

Accessibility assertions (`@accessibility` tag) are woven into each spec using the shared helper —
they do not live in a separate suite.

---

## Prerequisites

| Requirement | Notes |
|-------------|-------|
| Node.js ≥ 20 | `node --version` |
| Playwright browsers installed | Run `npx playwright install` once after `npm install` |
| Backend running | `http://localhost:8080` — see [`backend/README.md`](../backend/README.md) |
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
