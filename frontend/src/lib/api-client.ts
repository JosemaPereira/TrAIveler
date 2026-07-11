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
}

// Field-level errors use `error` (not `message`) as the message key per
// docs/api-design-standards.md §7 — keep this guard aligned with that shape.
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
 * validation failures (422-style, per-field messages).
 */
export class APIError extends Error {
  readonly status: number
  readonly code: string
  readonly requestId: string
  readonly fields?: APIErrorField[]

  constructor(
    status: number,
    code: string,
    message: string,
    requestId: string,
    fields?: APIErrorField[]
  ) {
    super(message)
    this.name = 'APIError'
    this.status = status
    this.code = code
    this.requestId = requestId
    this.fields = fields
  }
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
      body.fields
    )
  }

  return new APIError(
    response.status,
    'unknown_error',
    response.statusText || 'Unknown error',
    ''
  )
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
 */
export async function apiFetch<T>(
  endpoint: string,
  options: RequestInit = {}
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
    throw await buildAPIError(response)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
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

  post: <T>(endpoint: string, body?: unknown, options?: RequestInit): Promise<T> =>
    withBody<T>(endpoint, 'POST', body, options),

  put: <T>(endpoint: string, body?: unknown, options?: RequestInit): Promise<T> =>
    withBody<T>(endpoint, 'PUT', body, options),

  patch: <T>(endpoint: string, body?: unknown, options?: RequestInit): Promise<T> =>
    withBody<T>(endpoint, 'PATCH', body, options),

  delete: <T>(endpoint: string, options?: RequestInit): Promise<T> =>
    apiFetch<T>(endpoint, { ...options, method: 'DELETE' }),
}
