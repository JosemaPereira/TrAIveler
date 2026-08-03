import { describe, expect, it } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'

import { API_BASE_URL, testTrip } from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { createQueryWrapper } from '@/test/queryWrapper'
import { useTripDetail } from './useTripDetail'

describe('useTripDetail', () => {
  describe('when the trip exists and is owned by the caller', () => {
    it('should resolve with the flat trip the backend returns', async () => {
      server.use(
        http.get(`${API_BASE_URL}/trips/${testTrip.id}`, () =>
          HttpResponse.json({ trip: testTrip }, { status: 200 })
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useTripDetail(testTrip.id), {
        wrapper,
      })

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(result.current.data).toEqual(testTrip)
    })
  })

  describe('when the trip does not exist or is not owned by the caller', () => {
    it('should surface the 404 APIError', async () => {
      server.use(
        http.get(`${API_BASE_URL}/trips/not-mine`, () =>
          HttpResponse.json(
            {
              error: 'not_found',
              message: 'Trip not found',
              request_id: 'req_trip_detail_404',
            },
            { status: 404 }
          )
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useTripDetail('not-mine'), {
        wrapper,
      })

      await waitFor(() => {
        expect(result.current.isError).toBe(true)
      })
      expect(result.current.error).toMatchObject({
        status: 404,
        code: 'not_found',
      })
    })
  })
})
