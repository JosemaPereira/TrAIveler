import { useMutation } from '@tanstack/react-query'

import { useAuthStore } from '@/stores/auth-store'
import { authApi } from '@/features/auth/services/authApi'
import type { RegisterRequest, RegisterResponse } from '@/features/auth/types'

/**
 * Mutation for `POST /auth/register`.
 *
 * On success the returned user is pushed into the auth store, so a successful
 * registration also logs the caller in — the backend has already set the
 * session cookies on that same response.
 *
 * Retries are deliberately not configured here: the shared `queryClient`
 * already sets `mutations: { retry: false }`, and replaying a registration
 * could double-charge the stub payment provider.
 *
 * Errors reach the caller as `APIError` on `mutation.error` (409 `conflict`
 * for a duplicate email, 422 `validation_failed` with per-field messages);
 * render them with `getErrorMessage`/`getFieldErrors` from `lib/query-client`.
 */
export function useRegister() {
  const login = useAuthStore((state) => state.login)

  return useMutation<RegisterResponse, Error, RegisterRequest>({
    mutationFn: (request) => authApi.register(request),
    onSuccess: (data) => {
      login(data.user)
    },
  })
}
