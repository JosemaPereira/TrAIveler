import { setupServer } from 'msw/node'

import { handlers } from './handlers'

/**
 * Shared Mock Service Worker instance for Node-side (Vitest) tests.
 *
 * Lifecycle (`listen`/`resetHandlers`/`close`) is owned by `src/test/setup.ts`
 * so every test file gets it for free; import `server` directly only to add
 * per-test overrides via `server.use(...)`.
 *
 * Per docs/testing-guidelines.md (Layer 2, Frontend), feature-level tests
 * intercept HTTP here instead of mocking the global `fetch`.
 */
export const server = setupServer(...handlers)
