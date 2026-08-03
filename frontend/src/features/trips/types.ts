/**
 * Wire types for the trips + conversation surface
 * (`/api/v1/trips`, `/api/v1/trips/:id/conversation`).
 *
 * Field names are snake_case because this codebase has no case-conversion
 * layer — see `lib/api-client.ts`. Shapes mirror the Go structs that actually
 * serialize them (`backend/internal/trip/model.go`,
 * `backend/internal/conversation/model.go`), not spec prose, which can be
 * older than the implementation.
 */

/** Trip status values, mirroring the trips table CHECK constraint. */
export type TripStatus = 'draft' | 'published'

export interface Trip {
  id: string
  creator_id: string
  title: string
  description?: string
  status: TripStatus
  archived: boolean
  /** Optimistic-locking version (docs/data-model.md) — required as an `If-Match` header to update. */
  version: number
  created_at: string
  updated_at: string
}

export interface CreateTripRequest {
  title: string
  description?: string
}

export interface CreateTripResponse {
  trip: Trip
}

export interface ListTripsResponse {
  trips: Trip[]
}

export interface GetTripResponse {
  trip: Trip
}

export interface UpdateTripRequest {
  title: string
  description?: string
  status: TripStatus
}

export interface UpdateTripResponse {
  trip: Trip
}

/** Conversation message role, mirroring the conversation_messages table CHECK constraint. */
export type MessageRole = 'system' | 'user' | 'assistant'

export interface Message {
  id: string
  session_id: string
  role: MessageRole
  content: string
  token_count: number
  timestamp: string
}

/** Conversation session status, mirroring the conversation_sessions table CHECK constraint. */
export type ConversationSessionStatus = 'in_progress' | 'completed' | 'abandoned'

/** `GET /trips/:id/conversation` body — the session's status plus every message so far. */
export interface ConversationHistoryResponse {
  session_id: string
  status: ConversationSessionStatus
  messages: Message[]
}

export interface SendMessageRequest {
  message: string
}

/**
 * `POST /trips/:id/conversation` decoded payload. Delivered as a single SSE
 * frame (`Content-Type: text/event-stream`, body `"data: <json>\n\n"`), not
 * plain JSON — see `services/conversationApi.ts` for the parsing this shape
 * requires.
 */
export interface SendMessageResponse {
  session_id: string
  role: MessageRole
  message: string
  /** True once the AI-guided conversation has gathered enough to generate a full itinerary. */
  itinerary_ready: boolean
}
