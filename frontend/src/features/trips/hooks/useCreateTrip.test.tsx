import { describe, expect, it, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'

import { API_BASE_URL, testTrip } from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { createQueryWrapper } from '@/test/queryWrapper'
import { useCreateTrip } from './useCreateTrip'

describe('useCreateTrip', () => {
  describe('when the trip is created', () => {
    it('should resolve with the created trip', async () => {
      server.use(
        http.post(`${API_BASE_URL}/trips`, () =>
          HttpResponse.json({ trip: testTrip }, { status: 201 })
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useCreateTrip(), { wrapper })

      result.current.mutate({ title: testTrip.title })

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(result.current.data).toEqual(testTrip)
    })

    it('should invalidate the trips list query', async () => {
      server.use(
        http.post(`${API_BASE_URL}/trips`, () =>
          HttpResponse.json({ trip: testTrip }, { status: 201 })
        )
      )
      const { wrapper, queryClient } = createQueryWrapper()
      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries')
      const { result } = renderHook(() => useCreateTrip(), { wrapper })

      result.current.mutate({ title: testTrip.title })

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['trips'] })
    })
  })

  describe('when the request fails', () => {
    it('should surface the APIError', async () => {
      server.use(
        http.post(`${API_BASE_URL}/trips`, () =>
          HttpResponse.json(
            {
              error: 'validation_failed',
              message: 'Title is required',
              request_id: 'req_create_trip_422',
              fields: [{ field: 'title', error: 'Title is required' }],
            },
            { status: 422 }
          )
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useCreateTrip(), { wrapper })

      result.current.mutate({ title: '' })

      await waitFor(() => {
        expect(result.current.isError).toBe(true)
      })
      expect(result.current.error).toMatchObject({
        status: 422,
        code: 'validation_failed',
      })
    })
  })
})
