import { describe, expect, it, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'

import { API_BASE_URL, testTrip } from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { createQueryWrapper } from '@/test/queryWrapper'
import { useUpdateTrip } from './useUpdateTrip'

describe('useUpdateTrip', () => {
  describe('when the version is current', () => {
    it('should resolve with the updated trip', async () => {
      const updatedTrip = { ...testTrip, title: 'Renamed trip', version: 2 }
      server.use(
        http.put(`${API_BASE_URL}/trips/${testTrip.id}`, () =>
          HttpResponse.json({ trip: updatedTrip }, { status: 200 })
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useUpdateTrip(), { wrapper })

      result.current.mutate({
        id: testTrip.id,
        version: testTrip.version,
        request: { title: 'Renamed trip', status: 'draft' },
      })

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(result.current.data).toEqual(updatedTrip)
    })

    it('should send the version as an If-Match header', async () => {
      let ifMatch: string | null = null
      server.use(
        http.put(`${API_BASE_URL}/trips/${testTrip.id}`, ({ request }) => {
          ifMatch = request.headers.get('If-Match')
          return HttpResponse.json({ trip: testTrip }, { status: 200 })
        })
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useUpdateTrip(), { wrapper })

      result.current.mutate({
        id: testTrip.id,
        version: 5,
        request: { title: testTrip.title, status: 'draft' },
      })

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(ifMatch).toBe('5')
    })

    it('should invalidate both the trip detail and trips list queries', async () => {
      server.use(
        http.put(`${API_BASE_URL}/trips/${testTrip.id}`, () =>
          HttpResponse.json({ trip: testTrip }, { status: 200 })
        )
      )
      const { wrapper, queryClient } = createQueryWrapper()
      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries')
      const { result } = renderHook(() => useUpdateTrip(), { wrapper })

      result.current.mutate({
        id: testTrip.id,
        version: testTrip.version,
        request: { title: testTrip.title, status: 'draft' },
      })

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(invalidateSpy).toHaveBeenCalledWith({
        queryKey: ['trips', testTrip.id],
      })
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['trips'] })
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
              request_id: 'req_update_conflict',
            },
            { status: 409 }
          )
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useUpdateTrip(), { wrapper })

      result.current.mutate({
        id: testTrip.id,
        version: 1,
        request: { title: testTrip.title, status: 'draft' },
      })

      await waitFor(() => {
        expect(result.current.isError).toBe(true)
      })
      expect(result.current.error).toMatchObject({
        status: 409,
        code: 'conflict',
      })
    })
  })
})
