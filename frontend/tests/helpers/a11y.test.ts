/**
 * Unit tests for the checkPageA11y accessibility helper.
 *
 * `AxeBuilder` is mocked so these tests run fast and deterministically
 * without launching a real browser. This exercises the helper's own
 * pass/fail contract (throw on violations, resolve cleanly otherwise).
 * Full end-to-end verification against real rendered pages happens once
 * this helper is wired into `e2e/*.spec.ts` (Sprint 9, task 002-T023),
 * using the `e2e/` package's real Playwright + axe-core setup.
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { Page } from 'playwright-core'
import type { AxeResults, Result } from 'axe-core'

import { checkPageA11y } from './a11y'

const analyzeMock = vi.fn<() => Promise<AxeResults>>()
const withTagsMock = vi.fn()

vi.mock('@axe-core/playwright', () => {
  return {
    // A regular function expression is required here (not an arrow function)
    // so that `new AxeBuilder(...)` in the helper under test can invoke it as
    // a constructor.
    default: vi.fn().mockImplementation(function AxeBuilderMock() {
      return {
        withTags: withTagsMock,
        analyze: analyzeMock,
      }
    }),
  }
})

function buildViolation(overrides: Partial<Result> = {}): Result {
  return {
    id: 'image-alt',
    impact: 'critical',
    help: 'Images must have alternate text',
    description: 'Ensures <img> elements have alternate text',
    helpUrl: 'https://dequeuniversity.com/rules/axe/4.12/image-alt',
    tags: ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'],
    nodes: [{ html: '<img src="trip.jpg">', target: ['img'] }],
    ...overrides,
  } as Result
}

// The page object is never touched directly in these unit tests — AxeBuilder
// itself is mocked — so an empty stand-in is sufficient to satisfy the type.
const fakePage = {} as Page

describe('checkPageA11y', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    withTagsMock.mockReturnValue({ analyze: analyzeMock })
  })

  it('throws when the scan reports a WCAG 2.1 AA violation', async () => {
    analyzeMock.mockResolvedValue({
      violations: [buildViolation()],
    } as AxeResults)

    await expect(checkPageA11y(fakePage)).rejects.toThrow(/image-alt/)
  })

  it('resolves without throwing when the scan reports no violations', async () => {
    analyzeMock.mockResolvedValue({ violations: [] } as unknown as AxeResults)

    await expect(checkPageA11y(fakePage)).resolves.toBeUndefined()
  })

  it('scopes the scan to WCAG 2.1 AA tags', async () => {
    analyzeMock.mockResolvedValue({ violations: [] } as unknown as AxeResults)

    await checkPageA11y(fakePage)

    expect(withTagsMock).toHaveBeenCalledWith([
      'wcag2a',
      'wcag2aa',
      'wcag21a',
      'wcag21aa',
    ])
  })
})
