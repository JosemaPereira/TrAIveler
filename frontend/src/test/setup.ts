import '@testing-library/jest-dom/vitest'
import { afterAll, afterEach, beforeAll } from 'vitest'
import { cleanup } from '@testing-library/react'

import { server } from './msw/server'

// Mock Service Worker intercepts HTTP for feature-level tests
// (docs/testing-guidelines.md, Layer 2). `onUnhandledRequest: 'error'` makes
// an un-stubbed request a hard failure rather than a silent real network
// call, so a forgotten handler can never turn into a flaky test.
//
// Tests that stub the global `fetch` directly (src/lib/api-client.test.ts —
// the unit test for the client itself) replace `fetch` wholesale and never
// reach the interceptor, so the two styles coexist.
beforeAll(() => {
  server.listen({ onUnhandledRequest: 'error' })
})

// Vitest globals are disabled (test.globals is not set in vitest.config.ts),
// so React Testing Library cannot auto-detect the test framework's afterEach
// hook to run its automatic DOM cleanup. Register it explicitly so each test
// starts from an empty document instead of accumulating previous renders.
afterEach(() => {
  cleanup()
  server.resetHandlers()
})

afterAll(() => {
  server.close()
})
