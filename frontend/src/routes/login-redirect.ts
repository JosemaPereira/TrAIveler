/**
 * Sanitizes the `?redirect=` query param `features/auth/session-expiry.ts`
 * attaches to `/login`, so LoginPage can bounce a signed-in user back to
 * where they came from without ever letting an attacker-supplied value send
 * them off-site (open-redirect).
 */

const DEFAULT_REDIRECT = '/dashboard'

// Only a same-origin relative path is safe: a single leading slash not
// followed by another slash or backslash. `//host` is protocol-relative and
// `/\host` is a legacy backslash-as-slash quirk some browsers still honor —
// a browser would treat either as navigation to a different host, so both
// are rejected in favor of the dashboard, same as any absolute URL.
const SAFE_RELATIVE_PATH = /^\/(?!\/|\\)/

export function resolveLoginRedirect(redirectParam: string | null): string {
  if (!redirectParam) {
    return DEFAULT_REDIRECT
  }
  return SAFE_RELATIVE_PATH.test(redirectParam) ? redirectParam : DEFAULT_REDIRECT
}
