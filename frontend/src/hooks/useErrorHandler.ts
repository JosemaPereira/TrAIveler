import { useEffect } from 'react'
import { useNavigate } from 'react-router'

import { isAPIError } from '../lib/query-client'

export interface ErrorHandlerResult {
  title: string
  message: string
  requestId?: string
  isRetryable: boolean
}

const LOGIN_PATH = '/login'
const FALLBACK_MESSAGE = 'An unexpected error occurred. Please try again.'

/**
 * Turns an `unknown` error (typically a TanStack Query `error` field, or a
 * caught exception) into render-ready display info for `ErrorMessage`.
 * Returns `null` when there is no error, so callers can short-circuit with
 * `const errorInfo = useErrorHandler(queryError); if (!errorInfo) return null`.
 */
export function useErrorHandler(error: unknown): ErrorHandlerResult | null {
  const navigate = useNavigate()
  const apiError = isAPIError(error) ? error : undefined
  const status = apiError?.status

  // useNavigate() must not be invoked during render (React Router throws),
  // so the 401 redirect is a side effect that runs after commit.
  useEffect(() => {
    if (status === 401) {
      void navigate(LOGIN_PATH)
    }
  }, [status, navigate])

  if (error === null || error === undefined) {
    return null
  }

  const message = apiError?.message ?? FALLBACK_MESSAGE
  // requestId can be '' (no parseable error envelope), not just undefined —
  // treat that as "no id" rather than rendering an empty string.
  const requestId = apiError?.requestId ? apiError.requestId : undefined

  switch (status) {
    case 401:
      return { title: 'Session Expired', message, requestId, isRetryable: false }
    case 403:
      return { title: 'Permission Denied', message, requestId, isRetryable: false }
    case 404:
      return { title: 'Not Found', message, requestId, isRetryable: false }
    default:
      // Covers 500/503, any other 5xx, and non-APIError network failures.
      return { title: 'Something Went Wrong', message, requestId, isRetryable: true }
  }
}
