import { useEffect } from 'react'
import { useNavigate } from 'react-router'

import { isAPIError } from '../lib/query-client'
import { mapApiError } from '../lib/error-handler'
import type { ErrorDisplay } from '../lib/error-handler'

// Re-exported for existing consumers that imported the result type from here.
export type ErrorHandlerResult = ErrorDisplay

const LOGIN_PATH = '/login'

/**
 * Turns an `unknown` error (typically a TanStack Query `error` field, or a
 * caught exception) into render-ready display info for `ErrorMessage`, and
 * performs the 401 → login redirect side effect.
 *
 * The pure title/message/retry mapping lives in `mapApiError`
 * (`lib/error-handler.ts`); this hook only adds the routing side effect and
 * the null short-circuit, so callers can do
 * `const errorInfo = useErrorHandler(queryError); if (!errorInfo) return null`.
 */
export function useErrorHandler(error: unknown): ErrorHandlerResult | null {
  const navigate = useNavigate()
  const status = isAPIError(error) ? error.status : undefined

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

  return mapApiError(error)
}
