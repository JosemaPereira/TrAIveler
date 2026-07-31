import { api } from '../../../lib/api-client'
import type {
  CurrentUserResponse,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
} from '../types'

/**
 * HTTP surface for authentication, built on the fetch-based
 * `lib/api-client.ts` (there is no Axios in this repo — see the 008-T059
 * reconciliation note in docs/roadmap.md). The client already resolves paths
 * against `VITE_API_BASE_URL` (which ends in `/api/v1`), sends
 * `credentials: 'include'` so the HttpOnly session cookies ride along, and
 * rejects with `APIError` on any non-2xx response.
 *
 * Access and refresh tokens are never handled here: the backend sets and
 * rotates them as HttpOnly cookies, which JavaScript cannot read by design
 * (docs/security.md).
 */
export const authApi = {
  /**
   * Creates an account. Resolves with the new user plus, when
   * `payment_method_token` was supplied, the subscription created for it.
   * Rejects with a 409 `conflict` APIError if the email is already
   * registered, or a 422 `validation_failed` APIError carrying per-field
   * messages.
   */
  register: (request: RegisterRequest): Promise<RegisterResponse> =>
    api.post<RegisterResponse>('/auth/register', request),

  /**
   * Starts a session. Rejects with a 401 `authentication_required` APIError
   * for bad credentials (the backend deliberately does not distinguish an
   * unknown email from a wrong password), or a 429 `rate_limit_exceeded`
   * APIError whose `details.retry_after_seconds` says how long to wait.
   */
  login: (request: LoginRequest): Promise<LoginResponse> =>
    api.post<LoginResponse>('/auth/login', request),

  /** Ends the session server-side (204 No Content), clearing the auth cookies. */
  logout: (): Promise<void> => api.post<undefined>('/auth/logout'),

  /**
   * Resolves the account behind the session cookies — the only way to learn who
   * is signed in after a page reload, since the cookies are HTTP-only and no
   * user data survives client-side.
   *
   * Rejects with a 401 `authentication_required` APIError when there is no
   * session at all (no cookie, an invalid token, or a deleted account), which
   * for this endpoint is a normal negative answer rather than session death:
   * the api client exempts that case from its session-expiry teardown, so
   * probing from a public page cannot bounce an anonymous visitor to `/login`.
   */
  me: (): Promise<CurrentUserResponse> =>
    api.get<CurrentUserResponse>('/auth/me'),
}
