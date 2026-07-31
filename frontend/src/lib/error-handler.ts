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
 *   redirect — this function stays side-effect free). A 401 that reaches a
 *   component has already survived the api client's silent refresh attempt
 *   (008-T160), so the session really is over by the time this runs.
 * - 403 → "Permission Denied" (not retryable).
 * - 404 → "Not Found" (not retryable).
 * - 409 → "Conflict" (not retryable): the request clashed with existing state —
 *   a duplicate email on register, or a stale `version` on an optimistic-locked
 *   update. Retrying the identical request cannot help; the user has to change
 *   something or reload.
 * - 422 → "Invalid Input" (not retryable): per-field messages are in
 *   `APIError.fields`; render them against the form via `getFieldErrors` rather
 *   than only showing this summary.
 * - 429 → "Too Many Requests" (retryable, but not immediately — read the wait
 *   from `getRetryAfterSeconds`).
 * - Everything else (500/503, other 5xx, and non-`APIError` network failures)
 *   → generic "Something Went Wrong", treated as retryable.
 *
 * 409/422/429 used to fall through to the retryable "Something Went Wrong"
 * default, which was wrong in both directions: it invited a pointless retry of
 * a request that can only fail again, and it hid a fixable validation error
 * behind a generic failure. They only became reachable once the auth endpoints
 * went live (#177/#178/#194).
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
  const base = { message, requestId }

  switch (apiError?.status) {
    case 401:
      return { ...base, title: 'Session Expired', isRetryable: false }
    case 403:
      return { ...base, title: 'Permission Denied', isRetryable: false }
    case 404:
      return { ...base, title: 'Not Found', isRetryable: false }
    case 409:
      return { ...base, title: 'Conflict', isRetryable: false }
    case 422:
      return { ...base, title: 'Invalid Input', isRetryable: false }
    case 429:
      return { ...base, title: 'Too Many Requests', isRetryable: true }
    default:
      return { ...base, title: 'Something Went Wrong', isRetryable: true }
  }
}
