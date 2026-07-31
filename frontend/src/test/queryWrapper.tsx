import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router'
import type { ReactNode } from 'react'

/**
 * Builds the provider wrapper `renderHook`/`render` need for hooks that use
 * TanStack Query and/or react-router.
 *
 * Each call creates its own `QueryClient` so cached data and mutation state
 * never leak between tests, and retries are disabled so a failing request
 * settles immediately instead of making the test wait out a backoff.
 *
 * The client is returned alongside the wrapper so a test can assert against
 * it (e.g. spying on `clear()`).
 */
export function createQueryWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  function wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>{children}</MemoryRouter>
      </QueryClientProvider>
    )
  }

  return { wrapper, queryClient }
}
