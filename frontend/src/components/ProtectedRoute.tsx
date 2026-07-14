import { Outlet } from 'react-router'

/**
 * Route guard for authenticated-only pages.
 *
 * Scaffolding placeholder for Spec 004 (Security & Authentication/
 * Authorization Model, see specs/004-security-auth-model/). It currently
 * renders its child routes unconditionally via `<Outlet />`; the real
 * auth-gating (redirect to a login route when unauthenticated, backed by
 * `lib/authContext.tsx`) is implemented in a later Spec 004 issue, once the
 * backend auth endpoints exist.
 */
export function ProtectedRoute() {
  return <Outlet />
}
