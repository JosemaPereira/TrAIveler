import { isAPIError } from './query-client'

/**
 * Render-ready display info derived from an arbitrary error. `mapApiError`
 * turns an `unknown` error (an `APIError`, a bare `Error`, a rejected fetch)
 * into a user-friendly title/message and a retry hint, without any React or
 * routing side effects — those belong to the `useErrorHandler` hook, which
 * builds on this pure mapping.
 */
export interface ErrorDisplay {
  title: string
  message: string
  requestId?: string
  isRetryable: boolean
}

const FALLBACK_MESSAGE = 'An unexpected error occurred. Please try again.'

/**
 * Maps an API (or generic) error to user-friendly display info.
 *
 * - 401 → "Session Expired" (not retryable; the caller is responsible for any
 *   redirect — this function stays side-effect free).
 * - 403 → "Permission Denied" (not retryable).
 * - 404 → "Not Found" (not retryable).
 * - Everything else (500/503, other 5xx, and non-`APIError` network failures)
 *   → generic "Something Went Wrong", treated as retryable.
 *
 * The user-facing `message` comes from the API error envelope when available
 * (see `api-client.ts`), otherwise a safe generic fallback — the raw error is
 * never surfaced. `requestId` is included only when the envelope carried a
 * non-empty one, so operators can correlate a report to a server log.
 */
export function mapApiError(error: unknown): ErrorDisplay {
  const apiError = isAPIError(error) ? error : undefined
  const message = apiError?.message ?? FALLBACK_MESSAGE
  // requestId can be '' (no parseable error envelope), not just undefined —
  // treat that as "no id" rather than rendering an empty string.
  const requestId = apiError?.requestId ? apiError.requestId : undefined

  switch (apiError?.status) {
    case 401:
      return {
        title: 'Session Expired',
        message,
        requestId,
        isRetryable: false,
      }
    case 403:
      return {
        title: 'Permission Denied',
        message,
        requestId,
        isRetryable: false,
      }
    case 404:
      return { title: 'Not Found', message, requestId, isRetryable: false }
    default:
      return {
        title: 'Something Went Wrong',
        message,
        requestId,
        isRetryable: true,
      }
  }
}
