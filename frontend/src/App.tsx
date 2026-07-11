import { QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from 'react-router'

import { ErrorBoundary } from './components/ErrorBoundary'
import { queryClient } from './lib/query-client'
import { router } from './routes'

function App() {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </ErrorBoundary>
  )
}

export default App
