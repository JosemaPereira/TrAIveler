/**
 * Accessibility E2E test helper.
 *
 * Wraps `@axe-core/playwright` to provide a single reusable assertion that
 * fails the calling test with a descriptive error whenever the scanned page
 * has any WCAG 2.1 Level AA violation (NFR-A11Y-001, docs/nfrs.md).
 *
 * This helper is intentionally framework-agnostic beyond Playwright itself —
 * it does not assume any particular test runner tag. Sprint 9 (task
 * 002-T023) wires `checkPageA11y(page)` into every `e2e/*.spec.ts` file and
 * tags those assertions with `@accessibility` for the CI gate
 * (`.github/workflows/accessibility.yml`).
 */
import AxeBuilder from '@axe-core/playwright'
import type { Page } from 'playwright-core'

/**
 * WCAG 2.1 Level AA tag set used to scope the axe-core scan. Level A tags are
 * included because AA conformance requires all A-level success criteria too.
 */
const WCAG_2_1_AA_TAGS = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa']

/**
 * Runs an axe-core accessibility scan against the given Playwright page and
 * throws when any WCAG 2.1 Level AA violation is found.
 *
 * @param page - A Playwright page already navigated to the target URL.
 * @throws {Error} Listing every violation's rule ID, impact, and affected
 * node count, when one or more WCAG 2.1 AA violations are detected.
 */
export async function checkPageA11y(page: Page): Promise<void> {
  const results = await new AxeBuilder({ page })
    .withTags(WCAG_2_1_AA_TAGS)
    .analyze()

  if (results.violations.length === 0) {
    return
  }

  const summary = results.violations
    .map(
      (violation) =>
        `- [${violation.impact ?? 'unknown'}] ${violation.id}: ${violation.help} (${violation.nodes.length.toString()} node(s))`
    )
    .join('\n')

  throw new Error(`Accessibility violations found (WCAG 2.1 AA):\n${summary}`)
}
