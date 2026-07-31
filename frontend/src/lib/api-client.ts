// No hardcoded fallback: VITE_API_BASE_URL is environment-specific (docs/coding-guidelines.md
// forbids hardcoding those) and is baked in at Vite build time, so a silent default here could
// point a deployed build at the wrong backend. Fail loudly instead.
function getBaseUrl(): string {
  const baseUrl = import.meta.env.VITE_API_BASE_URL
  if (!baseUrl) {
    throw new Error(
      'VITE_API_BASE_URL is not set. Copy frontend/.env.example to frontend/.env.local and set it.'
    )
  }
  return baseUrl
}

export interface APIErrorField {
  field: string
  error: string
}

interface ErrorEnvelope {
  error: string
  message: string
  request_id: string
  fields?: APIErrorField[]
  details?: Record<string, unknown>
}

// Field-level errors use `error` (not `message`) as the message key per
// docs/api-design-standards.md §7 — keep this guard aligned with that shape.
// `fields` and `details` are both optional in that envelope, so neither is
// part of the discriminating check.
function isErrorEnvelope(value: unknown): value is ErrorEnvelope {
  return (
    typeof value === 'object' &&
    value !== null &&
    typeof (value as Record<string, unknown>).error === 'string' &&
    typeof (value as Record<string, unknown>).message === 'string' &&
    typeof (value as Record<string, unknown>).request_id === 'string'
  )
}

/**
 * Thrown by `apiFetch` (and the `api.*` helpers) for any non-2xx response.
 * `code` and `message` come from the response's error envelope (see
 * `docs/api-design-standards.md` §7); `fields` is only present for
 * validation failures (422-style, per-field messages), and `details` carries
 * code-specific extras such as `retry_after_seconds` on a 429 (read it via
 * `getRetryAfterSeconds` in `query-client.ts` rather than indexing it here).
 */
export class APIError extends Error {
  readonly status: number
  readonly code: string
  readonly requestId: string
  readonly fields?: APIErrorField[]
  readonly details?: Record<string, unknown>

  constructor(
    status: number,
    code: string,
    message: string,
    requestId: string,
    fields?: APIErrorField[],
    details?: Record<string, unknown>
  ) {
    super(message)
    this.name = 'APIError'
    this.status = status
    this.code = code
    this.requestId = requestId
    this.fields = fields
    this.details = details
  }
}

/**
 * Reads a `Retry-After` header as delta-seconds.
 *
 * Only the numeric form is honored. The HTTP-date form is legal but this API
 * never emits it (`internal/errors/handler.go` echoes an integer second count),
 * and parsing a date here would mean trusting client/server clock agreement.
 */
function retryAfterHeaderSeconds(response: Response): number | undefined {
  const header = response.headers.get('Retry-After')
  if (header === null) {
    return undefined
  }
  const seconds = Number(header)
  return Number.isInteger(seconds) && seconds >= 0 ? seconds : undefined
}

/**
 * Ensures a rate-limit/unavailable error carries `retry_after_seconds` even
 * when the wait only arrived as a header.
 *
 * The backend sends both (body `details` plus the `Retry-After` header), but an
 * intermediary — a load balancer or CDN shedding load — can answer 429/503 with
 * the header and no envelope of ours at all. Without this the UI would silently
 * lose the countdown in exactly the case where the wait matters most.
 */
function withRetryAfter(
  details: Record<string, unknown> | undefined,
  response: Response
): Record<string, unknown> | undefined {
  if (details?.retry_after_seconds !== undefined) {
    return details
  }
  const seconds = retryAfterHeaderSeconds(response)
  if (seconds === undefined) {
    return details
  }
  return { ...details, retry_after_seconds: seconds }
}

async function buildAPIError(response: Response): Promise<APIError> {
  let body: unknown
  try {
    body = await response.json()
  } catch {
    body = undefined
  }

  if (isErrorEnvelope(body)) {
    return new APIError(
      response.status,
      body.error,
      body.message,
      body.request_id,
      body.fields,
      withRetryAfter(body.details, response)
    )
  }

  // No envelope of ours — typically an intermediary (load balancer, CDN)
  // answering before the request reached the app. A Retry-After header is the
  // only structured thing such a response carries, so keep it.
  return new APIError(
    response.status,
    'unknown_error',
    response.statusText || 'Unknown error',
    '',
    undefined,
    withRetryAfter(undefined, response)
  )
}

const REFRESH_ENDPOINT = '/auth/refresh'

// A 401 from these endpoints means "those credentials are wrong", not "your
// session died", so it must not tear down the session or bounce the user to
// the login page they are already looking at.
const CREDENTIAL_ENDPOINTS = ['/auth/login', '/auth/register']

let sessionExpiredHandler: (() => void) | null = null

/**
 * Registers the callback invoked when the API says the session is gone for
 * good (a 401 that a token refresh could not rescue).
 *
 * The client cannot call the auth store or the router itself: `auth-store.ts`
 * imports this module, so importing it back would create a cycle, and the
 * client has no access to a router instance. The composition root wires the
 * real behavior instead — see `features/auth/session-expiry.ts`, installed
 * from `main.tsx`. Pass `null` to unregister (tests do this on teardown).
 */
export function setSessionExpiredHandler(handler: (() => void) | null): void {
  sessionExpiredHandler = handler
}

function notifySessionExpired(): void {
  sessionExpiredHandler?.()
}

function isCredentialEndpoint(endpoint: string): boolean {
  return CREDENTIAL_ENDPOINTS.some((path) => endpoint.startsWith(path))
}

// Only an *expired* access token is silently recoverable. Every other 401
// (missing cookie, revoked refresh token, tampered signature) comes back as
// `authentication_required` and means the session is genuinely over.
function isRecoverableExpiry(error: APIError, endpoint: string): boolean {
  return (
    error.status === 401 &&
    error.code === 'token_expired' &&
    !endpoint.startsWith(REFRESH_ENDPOINT)
  )
}

let inFlightRefresh: Promise<boolean> | null = null

// Deliberately a bare `fetch` rather than `apiFetch`: routing the refresh
// through the wrapper would let a 401 from the refresh endpoint trigger
// another refresh. Rotated tokens come back as HttpOnly cookies, so there is
// no response body worth reading — only whether it succeeded.
async function performRefresh(): Promise<boolean> {
  try {
    const response = await fetch(`${getBaseUrl()}${REFRESH_ENDPOINT}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Request-ID': crypto.randomUUID(),
      },
      credentials: 'include',
    })
    return response.ok
  } catch {
    // A network failure during refresh is indistinguishable from a rejected
    // one as far as the caller is concerned: the request cannot be retried.
    return false
  }
}

// Single-flight: several requests can fail with `token_expired` at the same
// moment (a dashboard firing three queries in parallel), and each of them
// refreshing would rotate the refresh token repeatedly and invalidate the
// others. They all await the same promise instead.
//
// Named for the token it renews, not the session: `auth-store.ts` exports an
// unrelated `refreshSession()` action that re-derives the *user* from the API,
// and two same-named functions in adjacent modules would read as one.
function refreshAccessToken(): Promise<boolean> {
  inFlightRefresh ??= performRefresh().finally(() => {
    inFlightRefresh = null
  })
  return inFlightRefresh
}

async function request<T>(
  endpoint: string,
  options: RequestInit,
  allowRefresh: boolean
): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set('Content-Type', 'application/json')
  headers.set('X-Request-ID', crypto.randomUUID())

  const response = await fetch(`${getBaseUrl()}${endpoint}`, {
    ...options,
    headers,
    credentials: 'include',
  })

  if (!response.ok) {
    const error = await buildAPIError(response)

    if (allowRefresh && isRecoverableExpiry(error, endpoint)) {
      if (await refreshAccessToken()) {
        // Replayed with `allowRefresh: false` so a still-401 retry falls
        // through to the session-expired path instead of looping. `options`
        // is safe to reuse because every body this client sends is an
        // already-serialized string (see `withBody`), not a consumed stream.
        return request<T>(endpoint, options, false)
      }
      notifySessionExpired()
      throw error
    }

    if (error.status === 401 && !isCredentialEndpoint(endpoint)) {
      notifySessionExpired()
    }

    throw error
  }

  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}

/**
 * Base fetch wrapper for the backend REST API. Resolves `endpoint` against
 * `VITE_API_BASE_URL`, attaches JSON headers, a per-request `X-Request-ID`,
 * and credentials for the session cookie.
 *
 * Rejects with `APIError` (never a bare `Response`/`Error`) whenever the
 * response status is not 2xx — callers that need the machine-readable code,
 * request ID, or per-field validation messages should catch and narrow on
 * `APIError` (see `isAPIError`/`getErrorMessage`/`getFieldErrors` in
 * `query-client.ts`).
 *
 * Transparently recovers from an expired access token: a 401 `token_expired`
 * triggers one shared `POST /auth/refresh` and one replay of the original
 * request. If the refresh fails — or any other 401 arrives — the registered
 * session-expired handler runs and the error is rethrown unchanged.
 */
export async function apiFetch<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  return request<T>(endpoint, options, true)
}

function withBody<T>(
  endpoint: string,
  method: 'POST' | 'PUT' | 'PATCH',
  body: unknown,
  options?: RequestInit
): Promise<T> {
  return apiFetch<T>(endpoint, {
    ...options,
    method,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
}

export const api = {
  get: <T>(endpoint: string, options?: RequestInit): Promise<T> =>
    apiFetch<T>(endpoint, { ...options, method: 'GET' }),

  post: <T>(
    endpoint: string,
    body?: unknown,
    options?: RequestInit
  ): Promise<T> => withBody<T>(endpoint, 'POST', body, options),

  put: <T>(
    endpoint: string,
    body?: unknown,
    options?: RequestInit
  ): Promise<T> => withBody<T>(endpoint, 'PUT', body, options),

  patch: <T>(
    endpoint: string,
    body?: unknown,
    options?: RequestInit
  ): Promise<T> => withBody<T>(endpoint, 'PATCH', body, options),

  delete: <T>(endpoint: string, options?: RequestInit): Promise<T> =>
    apiFetch<T>(endpoint, { ...options, method: 'DELETE' }),
}
