import { api, APIError, type APIErrorField } from '@/lib/api-client'
import type {
  ConversationHistoryResponse,
  SendMessageRequest,
  SendMessageResponse,
} from '@/features/trips/types'

// No hardcoded fallback, same rationale as `lib/api-client.ts`'s own
// `getBaseUrl`: `VITE_API_BASE_URL` is environment-specific and baked in at
// Vite build time, so a silent default here could point a deployed build at
// the wrong backend. Duplicated rather than imported because `api-client.ts`
// does not export it, and this is the one call site in the codebase that
// cannot go through `apiFetch` (see `postSseFrame` below).
function getBaseUrl(): string {
  const baseUrl = import.meta.env.VITE_API_BASE_URL
  if (!baseUrl) {
    throw new Error(
      'VITE_API_BASE_URL is not set. Copy frontend/.env.example to frontend/.env.local and set it.'
    )
  }
  return baseUrl
}

interface ErrorEnvelope {
  error: string
  message: string
  request_id: string
  fields?: APIErrorField[]
  details?: Record<string, unknown>
}

// Mirrors `lib/api-client.ts`'s own `isErrorEnvelope` guard.
function isErrorEnvelope(value: unknown): value is ErrorEnvelope {
  return (
    typeof value === 'object' &&
    value !== null &&
    typeof (value as Record<string, unknown>).error === 'string' &&
    typeof (value as Record<string, unknown>).message === 'string' &&
    typeof (value as Record<string, unknown>).request_id === 'string'
  )
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
      body.details
    )
  }

  return new APIError(
    response.status,
    'unknown_error',
    response.statusText || 'Unknown error',
    ''
  )
}

const SSE_DATA_PREFIX = 'data: '

/**
 * Strips the `data: ` prefix and trailing newlines off a single SSE frame
 * before JSON-decoding it. Returns `unknown` rather than a generic `<T>` —
 * the raw frame carries no evidence of its payload shape, so a generic here
 * would just be an uninferrable, uncheckable cast; callers assert the
 * concrete type themselves (see `postSseMessage`).
 */
function parseSseFrame(raw: string): unknown {
  const trimmed = raw.trim()
  const json = trimmed.startsWith(SSE_DATA_PREFIX)
    ? trimmed.slice(SSE_DATA_PREFIX.length)
    : trimmed
  return JSON.parse(json) as unknown
}

/**
 * POSTs a JSON body and parses a single-frame SSE response
 * (`Content-Type: text/event-stream`, body literally `"data: <json>\n\n"`).
 *
 * `POST /trips/:id/conversation` answers this way because
 * `internal/conversation/handler.go` was built against the streaming
 * `Response` contract meant for a future token-by-token implementation, even
 * though today's `conversation.Service.SendMessage` is synchronous and
 * returns one complete reply per turn. `api.post()` in `lib/api-client.ts`
 * calls `response.json()` unconditionally and cannot parse this body, so this
 * is a raw `fetch` instead — mirroring `api-client.ts`'s base URL
 * resolution, `credentials: 'include'`, and `APIError`-on-non-2xx, but
 * without its 401-refresh-and-retry machinery: a session dying mid-
 * conversation is rare enough that surfacing the 401 as-is is an acceptable
 * simplification for this issue's scope, and refresh-and-replay would need
 * to redo this same raw fetch + SSE parse rather than reuse `apiFetch`'s.
 */
async function postSseMessage<T>(
  endpoint: string,
  body: unknown
): Promise<T> {
  const response = await fetch(`${getBaseUrl()}${endpoint}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Request-ID': crypto.randomUUID(),
    },
    credentials: 'include',
    body: JSON.stringify(body),
  })

  if (!response.ok) {
    throw await buildAPIError(response)
  }

  return parseSseFrame(await response.text()) as T
}

/**
 * HTTP surface for a trip's AI-guided planning conversation
 * (`/trips/:id/conversation`).
 */
export const conversationApi = {
  /** Fetches the full conversation history for a trip: session status plus every message so far. */
  getHistory: (tripId: string): Promise<ConversationHistoryResponse> =>
    api.get<ConversationHistoryResponse>(`/trips/${tripId}/conversation`),

  /**
   * Sends a user message and returns the assistant's single reply for this
   * turn. `itinerary_ready` signals the AI-guided conversation has gathered
   * enough to generate a full itinerary. See `postSseMessage`'s doc comment
   * for why this cannot use `api.post`.
   */
  sendMessage: (
    tripId: string,
    request: SendMessageRequest
  ): Promise<SendMessageResponse> =>
    postSseMessage<SendMessageResponse>(
      `/trips/${tripId}/conversation`,
      request
    ),
}
