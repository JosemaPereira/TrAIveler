import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router'

import { useAuthStore } from '../../../stores/auth-store'
import { authApi } from '../services/authApi'

/**
 * Mutation for `POST /auth/logout`.
 *
 * On success it tears down every trace of the session in this tab: the auth
 * store is cleared, the whole TanStack Query cache is dropped (so the next
 * user of the browser can never see the previous user's cached trips), and
 * the caller is sent to `/login`.
 *
 * The teardown deliberately runs only on success — if the request fails the
 * server-side session may still be alive, so clearing local state would leave
 * the UI logged out while the cookies remain valid.
 */
export function useLogout() {
  const clearSession = useAuthStore((state) => state.logout)
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  return useMutation({
    mutationFn: () => authApi.logout(),
    onSuccess: () => {
      clearSession()
      queryClient.clear()
      // react-router's navigate is async; nothing here depends on the
      // transition completing, so the promise is explicitly discarded (same
      // idiom as hooks/useErrorHandler.ts).
      void navigate('/login')
    },
  })
}
