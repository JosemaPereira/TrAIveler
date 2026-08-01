/**
 * Sanitizes the `?redirect=` query param `features/auth/session-expiry.ts`
 * attaches to `/login` (`/login?redirect=<original-path>`), so LoginPage can
 * bounce a signed-in user back to where they came from without ever letting
 * an attacker-supplied value send them off-site.
 *
 * Only a same-origin relative path starting with a single `/` is accepted —
 * `//host`, `\\host`, and absolute URLs (`https://host`) are all rejected in
 * favor of the dashboard, since a browser would otherwise treat any of those
 * as a navigation to a different host (open-redirect).
 */

const DEFAULT_REDIRECT = '/dashboard'

// A single leading slash not immediately followed by another slash or a
// backslash. `//host` is protocol-relative; `/\host` is a legacy
// backslash-as-slash quirk some browsers still honor — both are rejected.
const SAFE_RELATIVE_PATH = /^\/(?!\/|\\)/

export function resolveLoginRedirect(redirectParam: string | null): string {
  if (!redirectParam) {
    return DEFAULT_REDIRECT
  }
  return SAFE_RELATIVE_PATH.test(redirectParam) ? redirectParam : DEFAULT_REDIRECT
}
