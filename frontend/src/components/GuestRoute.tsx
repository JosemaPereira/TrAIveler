import { Navigate, Outlet } from 'react-router'

import { useIsAuthenticated } from '@/stores/auth-store'

/**
 * Route guard for guest-only pages (`/login`, `/register`).
 *
 * Renders the matched child routes via `<Outlet />` when the auth store
 * reports no authenticated session; otherwise redirects to `/dashboard`,
 * replacing the current history entry so the guest-only URL is not left in
 * the back stack for an already-authenticated user.
 *
 * Inverse of `ProtectedRoute` — see `components/ProtectedRoute.tsx` for the
 * authenticated-only case. Authentication is derived from `useAuthStore`
 * (the HTTP-only JWT cookie is the real credential; the store mirrors
 * session state — see `stores/auth-store.ts`).
 */
export function GuestRoute() {
  const isAuthenticated = useIsAuthenticated()

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />
  }

  return <Outlet />
}
