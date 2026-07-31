import { useMutation } from '@tanstack/react-query'

import { getRetryAfterSeconds } from '../../../lib/query-client'
import { useAuthStore } from '../../../stores/auth-store'
import { authApi } from '../services/authApi'
import type { LoginRequest, LoginResponse } from '../types'

/**
 * Mutation for `POST /auth/login`, plus the rate-limit countdown seed.
 *
 * On success the returned user is pushed into the auth store; the backend has
 * already set the session cookies on that same response.
 *
 * `retryAfterSeconds` is derived from the current mutation error and is only
 * a number when the backend answered 429 `rate_limit_exceeded` with a
 * `details.retry_after_seconds` hint. It is the raw seed value — rendering an
 * actual ticking countdown from it is the login form's job, not this hook's.
 *
 * Retries are deliberately not configured here: the shared `queryClient`
 * already sets `mutations: { retry: false }`, and auto-retrying a login would
 * burn through the server-side attempt limiter.
 */
export function useLogin() {
  const login = useAuthStore((state) => state.login)

  const mutation = useMutation<LoginResponse, Error, LoginRequest>({
    mutationFn: (request) => authApi.login(request),
    onSuccess: (data) => {
      login(data.user)
    },
  })

  return {
    ...mutation,
    retryAfterSeconds: getRetryAfterSeconds(mutation.error),
  }
}
