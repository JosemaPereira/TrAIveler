import { http, HttpResponse } from 'msw'

import type { Subscription } from '../../features/auth/types'
import type { User } from '../../stores/auth-store'

/**
 * Absolute base URL every handler is registered against. Read from the same
 * env var `lib/api-client.ts` resolves requests with (committed for tests in
 * `frontend/.env.test`), so handler URLs can never drift from the client's.
 */
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL

/** Wire-shape user fixture, matching what `backend/internal/auth/models.go` serializes. */
export const testUser: User = {
  id: '11111111-1111-1111-1111-111111111111',
  email: 'traveler@example.com',
  full_name: 'Ada Traveler',
  has_subscription: false,
  created_at: '2026-01-01T00:00:00Z',
}

/** Wire-shape subscription fixture, matching `backend/internal/subscription/models.go`. */
export const testSubscription: Subscription = {
  id: '22222222-2222-2222-2222-222222222222',
  user_id: testUser.id,
  plan_id: '00000000-0000-0000-0000-000000000001',
  status: 'active',
  created_at: '2026-01-01T00:00:00Z',
}

/**
 * Default happy-path handlers for the auth surface. The server runs with
 * `onUnhandledRequest: 'error'`, so any request a test does not explicitly
 * account for fails loudly instead of hitting the network. Tests override an
 * individual endpoint (error responses, request capture, multi-step
 * sequences) with `server.use(...)`, which `resetHandlers()` undoes after
 * each test.
 */
export const authHandlers = [
  http.post(`${API_BASE_URL}/auth/register`, () =>
    HttpResponse.json({ user: testUser }, { status: 201 })
  ),

  http.post(`${API_BASE_URL}/auth/login`, () =>
    HttpResponse.json({ user: testUser }, { status: 200 })
  ),

  http.post(
    `${API_BASE_URL}/auth/logout`,
    () => new HttpResponse(null, { status: 204 })
  ),
]

// `POST /auth/refresh` and `GET /auth/me` have deliberately no default handler: it is only ever
// reached through the api client's silent 401 recovery, and a test that hits
// it without saying so is a bug worth failing on rather than a request worth
// answering. The 008-T160 tests register their own refresh handler so they
// can count how many times it fires, and every /auth/me test states the exact
// response it is exercising (signed in, signed out, expired mid-probe).

export const handlers = [...authHandlers]
