import '@testing-library/jest-dom/vitest'
import { afterEach } from 'vitest'
import { cleanup } from '@testing-library/react'

// Vitest globals are disabled (test.globals is not set in vitest.config.ts),
// so React Testing Library cannot auto-detect the test framework's afterEach
// hook to run its automatic DOM cleanup. Register it explicitly so each test
// starts from an empty document instead of accumulating previous renders.
afterEach(() => {
  cleanup()
})
