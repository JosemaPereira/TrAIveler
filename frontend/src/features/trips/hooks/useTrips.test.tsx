import { describe, expect, it } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'

import { API_BASE_URL, testTrip } from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { createQueryWrapper } from '@/test/queryWrapper'
import { useTrips } from './useTrips'

describe('useTrips', () => {
  describe('when the request succeeds', () => {
    it('should resolve with the trip list', async () => {
      server.use(
        http.get(`${API_BASE_URL}/trips`, () =>
          HttpResponse.json({ trips: [testTrip] }, { status: 200 })
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useTrips(), { wrapper })

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(result.current.data).toEqual([testTrip])
    })
  })

  describe('when the request fails', () => {
    it('should surface the APIError', async () => {
      server.use(
        http.get(`${API_BASE_URL}/trips`, () =>
          HttpResponse.json(
            {
              error: 'authentication_required',
              message: 'Session expired. Please log in again.',
              request_id: 'req_trips_hook_401',
            },
            { status: 401 }
          )
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useTrips(), { wrapper })

      await waitFor(() => {
        expect(result.current.isError).toBe(true)
      })
      expect(result.current.error).toMatchObject({
        status: 401,
        code: 'authentication_required',
      })
    })
  })
})
