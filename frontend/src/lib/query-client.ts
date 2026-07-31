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

/**
 * Reads the `retry_after_seconds` wait hint off an error's `details`, for UIs
 * that render a "try again in N seconds" countdown. It is set on 429
 * `rate_limit_exceeded` and 503 `service_unavailable` errors, either by the
 * backend envelope (`internal/errors/types.go`) or by `api-client.ts` falling
 * back to the response's `Retry-After` header.
 *
 * Returns `undefined` unless the value is genuinely a finite number, so a
 * malformed or absent hint degrades to "no countdown" instead of rendering
 * `NaN`.
 */
export function getRetryAfterSeconds(error: unknown): number | undefined {
  if (!isAPIError(error)) {
    return undefined
  }
  const retryAfter = error.details?.retry_after_seconds
  return typeof retryAfter === 'number' && Number.isFinite(retryAfter)
    ? retryAfter
    : undefined
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
