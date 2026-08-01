import { Navigate, Outlet } from 'react-router'

import { useIsAuthenticated } from '@/stores/auth-store'

/**
 * Route guard for authenticated-only pages (Spec 004 / Spec 008 008-T036).
 *
 * Renders the matched child routes via `<Outlet />` when the auth store
 * reports an authenticated session; otherwise redirects to `/login`,
 * replacing the current history entry so the guarded URL is not left in the
 * back stack.
 *
 * Authentication is derived from `useAuthStore` (the HTTP-only JWT cookie is
 * the real credential; the store mirrors session state — see
 * `stores/auth-store.ts`). Until the backend auth endpoints exist and a login
 * flow can populate the store, this correctly keeps every protected route
 * behind the login redirect.
 */
export function ProtectedRoute() {
  const isAuthenticated = useIsAuthenticated()

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <Outlet />
}
