import { afterEach, describe, expect, it, vi } from 'vitest'
import { http, HttpResponse } from 'msw'

import { API_BASE_URL } from '../test/msw/handlers'
import { server } from '../test/msw/server'
import { api, apiFetch, setSessionExpiredHandler } from './api-client'

/**
 * Tests for the 401 -> refresh -> retry flow (roadmap 008-T160). These are
 * MSW-backed rather than `fetch`-stubbed like the rest of `api-client.test.ts`
 * because the behavior under test is a multi-request sequence, and the two
 * styles cannot share a file: stubbing the global `fetch` would bypass MSW.
 */

const tokenExpiredEnvelope = {
  error: 'token_expired',
  message: 'Access token has expired',
  request_id: 'req_expired',
}

const sessionGoneEnvelope = {
  error: 'authentication_required',
  message: 'Session expired. Please log in again.',
  request_id: 'req_gone',
}

/** Answers `token_expired` for the first `failures` calls, then 200. */
function expiringTripsHandler(failures: number) {
  let calls = 0
  return http.get(`${API_BASE_URL}/trips`, () => {
    calls += 1
    if (calls <= failures) {
      return HttpResponse.json(tokenExpiredEnvelope, { status: 401 })
    }
    return HttpResponse.json({ trips: ['Rome'] }, { status: 200 })
  })
}

function countingRefreshHandler(response: () => Response) {
  const calls = { count: 0 }
  const handler = http.post(`${API_BASE_URL}/auth/refresh`, () => {
    calls.count += 1
    return response()
  })
  return { handler, calls }
}

afterEach(() => {
  setSessionExpiredHandler(null)
})

describe('apiFetch session refresh', () => {
  describe('when a request fails with token_expired and the refresh succeeds', () => {
    it('should retry the original request and resolve with its data', async () => {
      const { handler: refreshHandler } = countingRefreshHandler(() =>
        HttpResponse.json({ message: 'Token refreshed successfully' })
      )
      server.use(expiringTripsHandler(1), refreshHandler)

      await expect(apiFetch('/trips')).resolves.toEqual({ trips: ['Rome'] })
    })

    it('should refresh exactly once', async () => {
      const { handler: refreshHandler, calls } = countingRefreshHandler(() =>
        HttpResponse.json({ message: 'Token refreshed successfully' })
      )
      server.use(expiringTripsHandler(1), refreshHandler)

      await apiFetch('/trips')

      expect(calls.count).toBe(1)
    })

    it('should not notify the session-expired handler', async () => {
      const onSessionExpired = vi.fn()
      setSessionExpiredHandler(onSessionExpired)
      const { handler: refreshHandler } = countingRefreshHandler(() =>
        HttpResponse.json({ message: 'Token refreshed successfully' })
      )
      server.use(expiringTripsHandler(1), refreshHandler)

      await apiFetch('/trips')

      expect(onSessionExpired).not.toHaveBeenCalled()
    })
  })

  describe('when several requests expire concurrently', () => {
    it('should issue a single shared refresh for all of them', async () => {
      const { handler: refreshHandler, calls } = countingRefreshHandler(() =>
        HttpResponse.json({ message: 'Token refreshed successfully' })
      )
      server.use(expiringTripsHandler(3), refreshHandler)

      const results = await Promise.all([
        apiFetch('/trips'),
        apiFetch('/trips'),
        apiFetch('/trips'),
      ])

      expect(results).toEqual([
        { trips: ['Rome'] },
        { trips: ['Rome'] },
        { trips: ['Rome'] },
      ])
      expect(calls.count).toBe(1)
    })
  })

  describe('when the refresh itself fails', () => {
    it('should notify the session-expired handler', async () => {
      const onSessionExpired = vi.fn()
      setSessionExpiredHandler(onSessionExpired)
      const { handler: refreshHandler } = countingRefreshHandler(() =>
        HttpResponse.json(sessionGoneEnvelope, { status: 401 })
      )
      server.use(expiringTripsHandler(1), refreshHandler)

      await expect(apiFetch('/trips')).rejects.toMatchObject({ status: 401 })

      expect(onSessionExpired).toHaveBeenCalledOnce()
    })

    it('should treat a network failure during the refresh as a dead session', async () => {
      const onSessionExpired = vi.fn()
      setSessionExpiredHandler(onSessionExpired)
      const { handler: refreshHandler } = countingRefreshHandler(() =>
        HttpResponse.error()
      )
      server.use(expiringTripsHandler(1), refreshHandler)

      await expect(apiFetch('/trips')).rejects.toMatchObject({
        code: 'token_expired',
      })

      expect(onSessionExpired).toHaveBeenCalledOnce()
    })

    it('should reject with the original token_expired error', async () => {
      const { handler: refreshHandler } = countingRefreshHandler(() =>
        HttpResponse.json(sessionGoneEnvelope, { status: 401 })
      )
      server.use(expiringTripsHandler(1), refreshHandler)

      await expect(apiFetch('/trips')).rejects.toMatchObject({
        status: 401,
        code: 'token_expired',
      })
    })
  })

  describe('when the expired request carried a body', () => {
    it('should replay it with the same method and payload', async () => {
      const received: { method: string; body: unknown }[] = []
      const { handler: refreshHandler } = countingRefreshHandler(() =>
        HttpResponse.json({ message: 'Token refreshed successfully' })
      )
      server.use(
        http.post(`${API_BASE_URL}/trips`, async ({ request }) => {
          received.push({
            method: request.method,
            body: await request.json(),
          })
          if (received.length === 1) {
            return HttpResponse.json(tokenExpiredEnvelope, { status: 401 })
          }
          return HttpResponse.json({ id: 'trip_1' }, { status: 201 })
        }),
        refreshHandler
      )

      await expect(
        api.post('/trips', { destination: 'Rome' })
      ).resolves.toEqual({ id: 'trip_1' })

      // The replay reuses the original RequestInit; this pins that the body
      // survives it (it is an already-serialized string, not a consumed
      // stream) rather than the retry silently going out empty.
      expect(received).toEqual([
        { method: 'POST', body: { destination: 'Rome' } },
        { method: 'POST', body: { destination: 'Rome' } },
      ])
    })
  })

  describe('when the retried request is still rejected', () => {
    it('should not refresh a second time and should notify the handler once', async () => {
      const onSessionExpired = vi.fn()
      setSessionExpiredHandler(onSessionExpired)
      const { handler: refreshHandler, calls } = countingRefreshHandler(() =>
        HttpResponse.json({ message: 'Token refreshed successfully' })
      )
      server.use(expiringTripsHandler(Number.POSITIVE_INFINITY), refreshHandler)

      await expect(apiFetch('/trips')).rejects.toMatchObject({
        code: 'token_expired',
      })

      expect(calls.count).toBe(1)
      expect(onSessionExpired).toHaveBeenCalledOnce()
    })
  })

  describe('when a 401 is not token_expired', () => {
    it('should notify the session-expired handler without attempting a refresh', async () => {
      const onSessionExpired = vi.fn()
      setSessionExpiredHandler(onSessionExpired)
      const { handler: refreshHandler, calls } = countingRefreshHandler(() =>
        HttpResponse.json({ message: 'Token refreshed successfully' })
      )
      server.use(
        http.get(`${API_BASE_URL}/trips`, () =>
          HttpResponse.json(sessionGoneEnvelope, { status: 401 })
        ),
        refreshHandler
      )

      await expect(apiFetch('/trips')).rejects.toMatchObject({
        code: 'authentication_required',
      })

      expect(calls.count).toBe(0)
      expect(onSessionExpired).toHaveBeenCalledOnce()
    })
  })

  describe('when the refresh endpoint itself answers token_expired', () => {
    it('should not recurse into another refresh', async () => {
      const onSessionExpired = vi.fn()
      setSessionExpiredHandler(onSessionExpired)
      const { handler: refreshHandler, calls } = countingRefreshHandler(() =>
        HttpResponse.json(tokenExpiredEnvelope, { status: 401 })
      )
      server.use(refreshHandler)

      await expect(
        apiFetch('/auth/refresh', { method: 'POST' })
      ).rejects.toMatchObject({ code: 'token_expired' })

      expect(calls.count).toBe(1)
      expect(onSessionExpired).toHaveBeenCalledOnce()
    })
  })

  describe('when a credential endpoint answers 401', () => {
    it.each(['/auth/login', '/auth/register'])(
      'should treat a 401 from %s as bad credentials, not a dead session',
      async (endpoint) => {
        const onSessionExpired = vi.fn()
        setSessionExpiredHandler(onSessionExpired)
        server.use(
          http.post(`${API_BASE_URL}${endpoint}`, () =>
            HttpResponse.json(
              {
                error: 'authentication_required',
                message: 'Invalid email or password',
                request_id: 'req_bad_creds',
              },
              { status: 401 }
            )
          )
        )

        await expect(
          apiFetch(endpoint, { method: 'POST' })
        ).rejects.toMatchObject({ status: 401 })

        expect(onSessionExpired).not.toHaveBeenCalled()
      }
    )
  })

  describe('when no session-expired handler is registered', () => {
    it('should still reject without throwing from the notification', async () => {
      server.use(
        http.get(`${API_BASE_URL}/trips`, () =>
          HttpResponse.json(sessionGoneEnvelope, { status: 401 })
        )
      )

      await expect(apiFetch('/trips')).rejects.toMatchObject({ status: 401 })
    })
  })

  describe('when a non-401 error occurs', () => {
    it('should not attempt a refresh or notify the handler', async () => {
      const onSessionExpired = vi.fn()
      setSessionExpiredHandler(onSessionExpired)
      const { handler: refreshHandler, calls } = countingRefreshHandler(() =>
        HttpResponse.json({ message: 'Token refreshed successfully' })
      )
      server.use(
        http.get(`${API_BASE_URL}/trips`, () =>
          HttpResponse.json(
            {
              error: 'internal_error',
              message: 'Something went wrong',
              request_id: 'req_500',
            },
            { status: 500 }
          )
        ),
        refreshHandler
      )

      await expect(apiFetch('/trips')).rejects.toMatchObject({ status: 500 })

      expect(calls.count).toBe(0)
      expect(onSessionExpired).not.toHaveBeenCalled()
    })
  })
})

describe('apiFetch expected-401 endpoints', () => {
  // /auth/me is the session probe: "no, you are not signed in" is its normal
  // negative answer. Treating that as session death would make the app's own
  // boot sequence bounce every anonymous visitor to /login.
  it('should not notify the session-expired handler for a 401 from /auth/me', async () => {
    const onSessionExpired = vi.fn()
    setSessionExpiredHandler(onSessionExpired)
    const { handler: refreshHandler, calls } = countingRefreshHandler(() =>
      HttpResponse.json({ message: 'Token refreshed successfully' })
    )
    server.use(
      http.get(`${API_BASE_URL}/auth/me`, () =>
        HttpResponse.json(sessionGoneEnvelope, { status: 401 })
      ),
      refreshHandler
    )

    await expect(apiFetch('/auth/me')).rejects.toMatchObject({
      code: 'authentication_required',
    })

    expect(onSessionExpired).not.toHaveBeenCalled()
    expect(calls.count).toBe(0)
  })

  // The carve-out is about *session death*, not about skipping recovery: an
  // access token that merely expired mid-probe should still be refreshed.
  it('should still refresh and retry a token_expired from /auth/me', async () => {
    const onSessionExpired = vi.fn()
    setSessionExpiredHandler(onSessionExpired)
    const { handler: refreshHandler, calls } = countingRefreshHandler(() =>
      HttpResponse.json({ message: 'Token refreshed successfully' })
    )
    let attempts = 0
    server.use(
      http.get(`${API_BASE_URL}/auth/me`, () => {
        attempts += 1
        if (attempts === 1) {
          return HttpResponse.json(tokenExpiredEnvelope, { status: 401 })
        }
        return HttpResponse.json({ user: { id: 'u1' } }, { status: 200 })
      }),
      refreshHandler
    )

    await expect(apiFetch('/auth/me')).resolves.toEqual({ user: { id: 'u1' } })

    expect(calls.count).toBe(1)
    expect(onSessionExpired).not.toHaveBeenCalled()
  })
})
