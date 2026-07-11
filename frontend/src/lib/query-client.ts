import { QueryClient } from '@tanstack/react-query'

import { APIError } from './api-client'

const FIVE_MINUTES_MS = 5 * 60 * 1000

// Business rule: 4xx responses are client errors (bad input, auth, not
// found) that a retry cannot fix, so fail fast on those. Everything else
// (5xx, network errors) is treated as potentially transient and gets one
// retry before the query is marked as failed.
function shouldRetryQuery(failureCount: number, error: Error): boolean {
  if (error instanceof APIError && error.status >= 400 && error.status < 500) {
    return false
  }
  return failureCount < 1
}

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: FIVE_MINUTES_MS,
      refetchOnWindowFocus: false,
      refetchOnReconnect: true,
      retry: shouldRetryQuery,
    },
    mutations: {
      retry: false,
    },
  },
})

export function isAPIError(error: unknown): error is APIError {
  return error instanceof APIError
}

export function getErrorMessage(error: unknown): string {
  if (isAPIError(error)) {
    return error.message
  }
  if (error instanceof Error) {
    return error.message
  }
  return 'An unexpected error occurred.'
}

export function getFieldErrors(error: unknown): Record<string, string> {
  if (!isAPIError(error) || !error.fields) {
    return {}
  }
  return error.fields.reduce<Record<string, string>>((acc, field) => {
    acc[field.field] = field.error
    return acc
  }, {})
}
