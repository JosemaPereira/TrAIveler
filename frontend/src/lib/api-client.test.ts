import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { api, apiFetch, APIError } from './api-client'

// `headers` is a real Headers instance rather than being omitted: buildAPIError
// reads Retry-After off it, and a fake that lacks it would throw here while a
// real Response never can.
function jsonResponse(
  body: unknown,
  init: { status: number; ok: boolean; headers?: HeadersInit }
): Response {
  return {
    ok: init.ok,
    status: init.status,
    headers: new Headers(init.headers),
    json: () => Promise.resolve(body),
  } as Response
}

describe('apiFetch', () => {
  const fixedUuid = '11111111-1111-1111-1111-111111111111'

  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
    vi.spyOn(crypto, 'randomUUID').mockReturnValue(fixedUuid)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.unstubAllEnvs()
    vi.restoreAllMocks()
  })

  describe('when the response is successful', () => {
    it('should parse and return JSON on a GET response', async () => {
      const payload = { id: 'trip_1', name: 'Rome' }
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse(payload, { status: 200, ok: true })
      )

      const result = await apiFetch<typeof payload>('/trips/trip_1')

      expect(result).toEqual(payload)
    })

    it('should parse and return JSON on a POST response', async () => {
      const payload = { id: 'trip_2', name: 'Paris' }
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse(payload, { status: 201, ok: true })
      )

      const result = await apiFetch<typeof payload>('/trips', {
        method: 'POST',
        body: JSON.stringify({ name: 'Paris' }),
      })

      expect(result).toEqual(payload)
    })
  })

  describe('when building a request', () => {
    it('should send Content-Type: application/json on every request', async () => {
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse({}, { status: 200, ok: true })
      )

      await apiFetch('/trips')

      const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
      const headers = new Headers(requestInit?.headers)
      expect(headers.get('Content-Type')).toBe('application/json')
    })

    it('should generate and send a client X-Request-ID header on every request', async () => {
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse({}, { status: 200, ok: true })
      )

      await apiFetch('/trips')

      const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
      const headers = new Headers(requestInit?.headers)
      expect(headers.get('X-Request-ID')).toBe(fixedUuid)
    })

    it('should always set credentials to include', async () => {
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse({}, { status: 200, ok: true })
      )

      await apiFetch('/trips')

      const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
      expect(requestInit?.credentials).toBe('include')
    })
  })

  describe('when resolving the API base URL', () => {
    it('should throw a clear error instead of silently defaulting when VITE_API_BASE_URL is unset', async () => {
      vi.stubEnv('VITE_API_BASE_URL', undefined)

      await expect(apiFetch('/trips')).rejects.toThrow(
        'VITE_API_BASE_URL is not set'
      )
      expect(fetch).not.toHaveBeenCalled()
    })

    it('should use VITE_API_BASE_URL when it is set', async () => {
      vi.stubEnv(
        'VITE_API_BASE_URL',
        'https://api-staging.traveler.example.com/api/v1'
      )
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse({}, { status: 200, ok: true })
      )

      await apiFetch('/trips')

      const [url] = vi.mocked(fetch).mock.calls[0] ?? []
      expect(url).toBe('https://api-staging.traveler.example.com/api/v1/trips')
    })
  })

  describe('when the response is an error envelope', () => {
    it.each([400, 401, 404, 422, 500])(
      'should throw an APIError with status, code, message, requestId, and fields on a %s response',
      async (status) => {
        const envelope = {
          error: 'validation_failed',
          message: 'One or more fields failed validation',
          request_id: 'req_abc123xyz',
          fields: [
            { field: 'email', error: 'Email address is already registered' },
          ],
        }
        vi.mocked(fetch).mockResolvedValue(
          jsonResponse(envelope, { status, ok: false })
        )

        await expect(apiFetch('/trips')).rejects.toMatchObject({
          status,
          code: 'validation_failed',
          message: 'One or more fields failed validation',
          requestId: 'req_abc123xyz',
          fields: [
            { field: 'email', error: 'Email address is already registered' },
          ],
        })
      }
    )

    it('should throw an APIError instance', async () => {
      const envelope = {
        error: 'not_found',
        message: 'Trip not found',
        request_id: 'req_xyz789',
      }
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse(envelope, { status: 404, ok: false })
      )

      await expect(apiFetch('/trips/missing')).rejects.toBeInstanceOf(APIError)
    })

    it('should carry the envelope details through to the APIError', async () => {
      const envelope = {
        error: 'rate_limit_exceeded',
        message: 'Too many login attempts. Please try again later.',
        request_id: 'req_rate001',
        details: { retry_after_seconds: 42 },
      }
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse(envelope, { status: 429, ok: false })
      )

      await expect(apiFetch('/auth/login')).rejects.toMatchObject({
        status: 429,
        code: 'rate_limit_exceeded',
        details: { retry_after_seconds: 42 },
      })
    })

    it('should leave details undefined when the envelope has no details object', async () => {
      const envelope = {
        error: 'not_found',
        message: 'Trip not found',
        request_id: 'req_nf001',
      }
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse(envelope, { status: 404, ok: false })
      )

      await expect(apiFetch('/trips/missing')).rejects.toMatchObject({
        details: undefined,
      })
    })

    it('should fall back to the Retry-After header when the envelope omits the wait', async () => {
      const envelope = {
        error: 'rate_limit_exceeded',
        message: 'Too many requests',
        request_id: 'req_rate002',
      }
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse(envelope, {
          status: 429,
          ok: false,
          headers: { 'Retry-After': '30' },
        })
      )

      await expect(apiFetch('/auth/login')).rejects.toMatchObject({
        details: { retry_after_seconds: 30 },
      })
    })

    it('should prefer the envelope wait over the Retry-After header', async () => {
      const envelope = {
        error: 'rate_limit_exceeded',
        message: 'Too many requests',
        request_id: 'req_rate003',
        details: { retry_after_seconds: 7 },
      }
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse(envelope, {
          status: 429,
          ok: false,
          headers: { 'Retry-After': '30' },
        })
      )

      await expect(apiFetch('/auth/login')).rejects.toMatchObject({
        details: { retry_after_seconds: 7 },
      })
    })

    it('should ignore a non-numeric Retry-After header', async () => {
      // The HTTP-date form is legal but this API never emits it, and parsing it
      // would mean trusting client/server clock agreement.
      const envelope = {
        error: 'rate_limit_exceeded',
        message: 'Too many requests',
        request_id: 'req_rate004',
      }
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse(envelope, {
          status: 429,
          ok: false,
          headers: { 'Retry-After': 'Wed, 21 Oct 2026 07:28:00 GMT' },
        })
      )

      await expect(apiFetch('/auth/login')).rejects.toMatchObject({
        details: undefined,
      })
    })

    it('should keep the Retry-After header when no error envelope is returned', async () => {
      // An intermediary (load balancer, CDN) shedding load answers 429 with the
      // header and no envelope of ours; the countdown must survive that.
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse('<html>Too Many Requests</html>', {
          status: 429,
          ok: false,
          headers: { 'Retry-After': '120' },
        })
      )

      await expect(apiFetch('/auth/login')).rejects.toMatchObject({
        code: 'unknown_error',
        details: { retry_after_seconds: 120 },
      })
    })

    it('should leave fields undefined when the envelope has no fields array', async () => {
      const envelope = {
        error: 'authentication_required',
        message: 'Missing or invalid auth token',
        request_id: 'req_auth001',
      }
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse(envelope, { status: 401, ok: false })
      )

      await expect(apiFetch('/trips')).rejects.toMatchObject({
        fields: undefined,
      })
    })
  })
})

describe('api method helpers', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
    vi.spyOn(crypto, 'randomUUID').mockReturnValue(
      '22222222-2222-2222-2222-222222222222'
    )
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  describe('when calling a verb helper', () => {
    it('should send a GET request', async () => {
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse({ ok: true }, { status: 200, ok: true })
      )

      await api.get('/trips')

      const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
      expect(requestInit?.method).toBe('GET')
    })

    it('should send a POST request with a JSON-serialized body', async () => {
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse({ ok: true }, { status: 201, ok: true })
      )

      await api.post('/trips', { name: 'Tokyo' })

      const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
      expect(requestInit?.method).toBe('POST')
      expect(requestInit?.body).toBe(JSON.stringify({ name: 'Tokyo' }))
    })

    it('should send a PUT request with a JSON-serialized body', async () => {
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse({ ok: true }, { status: 200, ok: true })
      )

      await api.put('/trips/trip_1', { name: 'Tokyo Updated' })

      const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
      expect(requestInit?.method).toBe('PUT')
      expect(requestInit?.body).toBe(JSON.stringify({ name: 'Tokyo Updated' }))
    })

    it('should send a PATCH request with a JSON-serialized body', async () => {
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse({ ok: true }, { status: 200, ok: true })
      )

      await api.patch('/trips/trip_1', { name: 'Tokyo Patched' })

      const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
      expect(requestInit?.method).toBe('PATCH')
      expect(requestInit?.body).toBe(JSON.stringify({ name: 'Tokyo Patched' }))
    })

    it('should send a DELETE request', async () => {
      vi.mocked(fetch).mockResolvedValue(
        jsonResponse({}, { status: 204, ok: true })
      )

      await api.delete('/trips/trip_1')

      const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
      expect(requestInit?.method).toBe('DELETE')
    })
  })

  describe('when the response is 204 No Content', () => {
    it('should resolve to undefined without parsing a body', async () => {
      vi.mocked(fetch).mockResolvedValue({
        ok: true,
        status: 204,
        json: () => Promise.reject(new Error('no body to parse')),
      } as Response)

      await expect(api.delete('/trips/trip_1')).resolves.toBeUndefined()
    })
  })
})
