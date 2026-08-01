import { setSessionExpiredHandler } from '@/lib/api-client'
import { useAuthStore } from '@/stores/auth-store'

const LOGIN_PATH = '/login'

/**
 * Reacts to an unrecoverable 401 from the API client: drop the in-memory
 * session and send the user to the login page, remembering where they were so
 * the login form can bounce them back afterwards (`?redirect=<original-path>`,
 * roadmap 008-T160).
 *
 * Navigation goes through `window.location.assign` rather than the router on
 * purpose. `lib/api-client.ts` has no router instance to call — it is plain
 * module code, not a component — and a full document load is desirable here
 * anyway: it guarantees no stale authenticated state survives in any React
 * tree, store, or in-flight query.
 */
export function handleSessionExpired(): void {
  useAuthStore.getState().logout()

  const { pathname, search } = window.location
  if (pathname === LOGIN_PATH) {
    // Already where we would send them; redirecting would discard whatever
    // they have typed into the login form and could loop.
    return
  }

  const redirectTo = encodeURIComponent(`${pathname}${search}`)
  window.location.assign(`${LOGIN_PATH}?redirect=${redirectTo}`)
}

/**
 * Wires `handleSessionExpired` into the API client. Called once from the
 * composition root (`main.tsx`) — this indirection is what keeps
 * `api-client.ts` free of any import of the auth store, which imports it.
 */
export function installSessionExpiryHandler(): void {
  setSessionExpiredHandler(handleSessionExpired)
}
