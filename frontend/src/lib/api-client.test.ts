import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { api, apiFetch, APIError } from './api-client'

function jsonResponse(
  body: unknown,
  init: { status: number; ok: boolean }
): Response {
  return {
    ok: init.ok,
    status: init.status,
    json: () => Promise.resolve(body),
  } as Response
}

describe('unit: apiFetch', () => {
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

  it('parses and returns JSON on a successful GET response', async () => {
    const payload = { id: 'trip_1', name: 'Rome' }
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse(payload, { status: 200, ok: true })
    )

    const result = await apiFetch<typeof payload>('/trips/trip_1')

    expect(result).toEqual(payload)
  })

  it('parses and returns JSON on a successful POST response', async () => {
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

  it('sends Content-Type: application/json on every request', async () => {
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse({}, { status: 200, ok: true })
    )

    await apiFetch('/trips')

    const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
    const headers = new Headers(requestInit?.headers)
    expect(headers.get('Content-Type')).toBe('application/json')
  })

  it('generates and sends a client X-Request-ID header on every request', async () => {
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse({}, { status: 200, ok: true })
    )

    await apiFetch('/trips')

    const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
    const headers = new Headers(requestInit?.headers)
    expect(headers.get('X-Request-ID')).toBe(fixedUuid)
  })

  it('always sets credentials to include', async () => {
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse({}, { status: 200, ok: true })
    )

    await apiFetch('/trips')

    const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
    expect(requestInit?.credentials).toBe('include')
  })

  it('throws a clear error instead of silently defaulting when VITE_API_BASE_URL is unset', async () => {
    vi.stubEnv('VITE_API_BASE_URL', undefined)

    await expect(apiFetch('/trips')).rejects.toThrow(
      'VITE_API_BASE_URL is not set'
    )
    expect(fetch).not.toHaveBeenCalled()
  })

  it('uses VITE_API_BASE_URL when it is set', async () => {
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

  it.each([400, 401, 404, 422, 500])(
    'throws an APIError with status, code, message, requestId, and fields on a %s error response',
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

  it('throws an APIError instance on error responses', async () => {
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

  it('leaves fields undefined when the error envelope has no fields array', async () => {
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

describe('unit: api method helpers', () => {
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

  it('sends a GET request', async () => {
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse({ ok: true }, { status: 200, ok: true })
    )

    await api.get('/trips')

    const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
    expect(requestInit?.method).toBe('GET')
  })

  it('sends a POST request with a JSON-serialized body', async () => {
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse({ ok: true }, { status: 201, ok: true })
    )

    await api.post('/trips', { name: 'Tokyo' })

    const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
    expect(requestInit?.method).toBe('POST')
    expect(requestInit?.body).toBe(JSON.stringify({ name: 'Tokyo' }))
  })

  it('sends a PUT request with a JSON-serialized body', async () => {
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse({ ok: true }, { status: 200, ok: true })
    )

    await api.put('/trips/trip_1', { name: 'Tokyo Updated' })

    const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
    expect(requestInit?.method).toBe('PUT')
    expect(requestInit?.body).toBe(JSON.stringify({ name: 'Tokyo Updated' }))
  })

  it('sends a PATCH request with a JSON-serialized body', async () => {
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse({ ok: true }, { status: 200, ok: true })
    )

    await api.patch('/trips/trip_1', { name: 'Tokyo Patched' })

    const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
    expect(requestInit?.method).toBe('PATCH')
    expect(requestInit?.body).toBe(JSON.stringify({ name: 'Tokyo Patched' }))
  })

  it('sends a DELETE request', async () => {
    vi.mocked(fetch).mockResolvedValue(
      jsonResponse({}, { status: 204, ok: true })
    )

    await api.delete('/trips/trip_1')

    const [, requestInit] = vi.mocked(fetch).mock.calls[0] ?? []
    expect(requestInit?.method).toBe('DELETE')
  })

  it('resolves to undefined for a 204 No Content response without parsing a body', async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      status: 204,
      json: () => Promise.reject(new Error('no body to parse')),
    } as Response)

    await expect(api.delete('/trips/trip_1')).resolves.toBeUndefined()
  })
})
