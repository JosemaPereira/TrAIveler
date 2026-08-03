import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'

import { API_BASE_URL, testTrip } from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { tripsApi } from './tripsApi'

describe('tripsApi.list', () => {
  describe('when the request succeeds', () => {
    it('should GET /trips and resolve with the trip list', async () => {
      let method: string | undefined
      server.use(
        http.get(`${API_BASE_URL}/trips`, ({ request }) => {
          method = request.method
          return HttpResponse.json({ trips: [testTrip] }, { status: 200 })
        })
      )

      const result = await tripsApi.list()

      expect(method).toBe('GET')
      expect(result.trips).toEqual([testTrip])
    })

    it('should send the session cookie along with the request', async () => {
      let credentials: RequestCredentials | undefined
      server.use(
        http.get(`${API_BASE_URL}/trips`, ({ request }) => {
          credentials = request.credentials
          return HttpResponse.json({ trips: [] }, { status: 200 })
        })
      )

      await tripsApi.list()

      expect(credentials).toBe('include')
    })
  })

  describe('when the caller is not authenticated', () => {
    it('should reject with a 401 APIError', async () => {
      server.use(
        http.get(`${API_BASE_URL}/trips`, () =>
          HttpResponse.json(
            {
              error: 'authentication_required',
              message: 'Session expired. Please log in again.',
              request_id: 'req_trips_401',
            },
            { status: 401 }
          )
        )
      )

      await expect(tripsApi.list()).rejects.toMatchObject({
        status: 401,
        code: 'authentication_required',
      })
    })
  })
})

describe('tripsApi.create', () => {
  describe('when the payload is valid', () => {
    it('should POST the title and description to /trips', async () => {
      let receivedBody: unknown
      server.use(
        http.post(`${API_BASE_URL}/trips`, async ({ request }) => {
          receivedBody = await request.json()
          return HttpResponse.json({ trip: testTrip }, { status: 201 })
        })
      )

      await tripsApi.create({
        title: 'Two weeks in Japan',
        description: 'Tokyo, Kyoto, Osaka',
      })

      expect(receivedBody).toEqual({
        title: 'Two weeks in Japan',
        description: 'Tokyo, Kyoto, Osaka',
      })
    })

    it('should resolve with the created trip', async () => {
      server.use(
        http.post(`${API_BASE_URL}/trips`, () =>
          HttpResponse.json({ trip: testTrip }, { status: 201 })
        )
      )

      const result = await tripsApi.create({ title: testTrip.title })

      expect(result.trip).toEqual(testTrip)
    })
  })

  describe('when the payload fails validation', () => {
    it('should reject with a 422 APIError carrying per-field errors', async () => {
      server.use(
        http.post(`${API_BASE_URL}/trips`, () =>
          HttpResponse.json(
            {
              error: 'validation_failed',
              message: 'One or more fields failed validation',
              request_id: 'req_trip_validation',
              fields: [{ field: 'title', error: 'Title is required' }],
            },
            { status: 422 }
          )
        )
      )

      await expect(tripsApi.create({ title: '' })).rejects.toMatchObject({
        status: 422,
        code: 'validation_failed',
        fields: [{ field: 'title', error: 'Title is required' }],
      })
    })
  })
})

describe('tripsApi.get', () => {
  describe('when the trip exists and is owned by the caller', () => {
    it('should GET /trips/:id and resolve with the trip', async () => {
      let method: string | undefined
      server.use(
        http.get(`${API_BASE_URL}/trips/${testTrip.id}`, ({ request }) => {
          method = request.method
          return HttpResponse.json({ trip: testTrip }, { status: 200 })
        })
      )

      const result = await tripsApi.get(testTrip.id)

      expect(method).toBe('GET')
      expect(result.trip).toEqual(testTrip)
    })
  })

  describe('when the trip does not exist or is not owned by the caller', () => {
    it('should reject with a 404 APIError', async () => {
      server.use(
        http.get(`${API_BASE_URL}/trips/not-mine`, () =>
          HttpResponse.json(
            {
              error: 'not_found',
              message: 'Trip not found',
              request_id: 'req_trip_404',
            },
            { status: 404 }
          )
        )
      )

      await expect(tripsApi.get('not-mine')).rejects.toMatchObject({
        status: 404,
        code: 'not_found',
      })
    })
  })
})

describe('tripsApi.update', () => {
  describe('when the version is current', () => {
    it('should PUT the body and send the version as an If-Match header', async () => {
      let receivedBody: unknown
      let ifMatch: string | null = null
      server.use(
        http.put(
          `${API_BASE_URL}/trips/${testTrip.id}`,
          async ({ request }) => {
            receivedBody = await request.json()
            ifMatch = request.headers.get('If-Match')
            return HttpResponse.json(
              { trip: { ...testTrip, version: 4 } },
              { status: 200 }
            )
          }
        )
      )

      const result = await tripsApi.update(testTrip.id, 3, {
        title: 'Updated title',
        status: 'published',
      })

      expect(receivedBody).toEqual({
        title: 'Updated title',
        status: 'published',
      })
      expect(ifMatch).toBe('3')
      expect(result.trip.version).toBe(4)
    })
  })

  describe('when the version is stale', () => {
    it('should reject with a 409 conflict APIError', async () => {
      server.use(
        http.put(`${API_BASE_URL}/trips/${testTrip.id}`, () =>
          HttpResponse.json(
            {
              error: 'conflict',
              message: 'Trip was modified by another request',
              request_id: 'req_trip_conflict',
            },
            { status: 409 }
          )
        )
      )

      await expect(
        tripsApi.update(testTrip.id, 1, {
          title: 'Stale update',
          status: 'draft',
        })
      ).rejects.toMatchObject({ status: 409, code: 'conflict' })
    })
  })
})

describe('tripsApi.delete', () => {
  describe('when the trip is deleted', () => {
    it('should DELETE /trips/:id and resolve on 204 No Content', async () => {
      let method: string | undefined
      server.use(
        http.delete(`${API_BASE_URL}/trips/${testTrip.id}`, ({ request }) => {
          method = request.method
          return new HttpResponse(null, { status: 204 })
        })
      )

      await expect(tripsApi.delete(testTrip.id)).resolves.toBeUndefined()
      expect(method).toBe('DELETE')
    })
  })

  describe('when the trip does not exist', () => {
    it('should reject with a 404 APIError', async () => {
      server.use(
        http.delete(`${API_BASE_URL}/trips/not-mine`, () =>
          HttpResponse.json(
            {
              error: 'not_found',
              message: 'Trip not found',
              request_id: 'req_delete_404',
            },
            { status: 404 }
          )
        )
      )

      await expect(tripsApi.delete('not-mine')).rejects.toMatchObject({
        status: 404,
        code: 'not_found',
      })
    })
  })
})
