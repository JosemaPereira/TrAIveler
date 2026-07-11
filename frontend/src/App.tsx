import { QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from 'react-router'

import { ErrorBoundary } from './components/ErrorBoundary'
import { queryClient } from './lib/query-client'
import { router } from './routes'

function App() {
  // ErrorBoundary wraps QueryClientProvider and RouterProvider (not just the
  // routed pages), so it also catches errors thrown by those providers
  // themselves, not only errors thrown while rendering a matched route.
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </ErrorBoundary>
  )
}

export default App
