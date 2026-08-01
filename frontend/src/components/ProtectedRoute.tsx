import { Navigate, Outlet, useLocation } from 'react-router'

import { useIsAuthenticated } from '@/stores/auth-store'

/**
 * Route guard for authenticated-only pages.
 *
 * Renders the matched child routes via `<Outlet />` when the auth store
 * reports an authenticated session; otherwise redirects to
 * `/login?redirect=<attempted-path>`, replacing the current history entry so
 * the guarded URL is not left in the back stack. The path is URL-encoded so
 * `resolveLoginRedirect` can send the user back there after login.
 *
 * Authentication is derived from `useAuthStore` (the HTTP-only JWT cookie is
 * the real credential; the store mirrors session state — see
 * `stores/auth-store.ts`).
 */
export function ProtectedRoute() {
  const isAuthenticated = useIsAuthenticated()
  const location = useLocation()

  if (!isAuthenticated) {
    const attemptedPath = `${location.pathname}${location.search}`
    return (
      <Navigate
        to={`/login?redirect=${encodeURIComponent(attemptedPath)}`}
        replace
      />
    )
  }

  return <Outlet />
}
