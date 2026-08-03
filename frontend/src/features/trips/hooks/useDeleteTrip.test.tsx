import { afterEach, describe, expect, it, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'

import { API_BASE_URL, testTrip } from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { createQueryWrapper } from '@/test/queryWrapper'
import { useDeleteTrip } from './useDeleteTrip'

const mockNavigate = vi.fn()

vi.mock('react-router', async () => {
  const actual =
    await vi.importActual<typeof import('react-router')>('react-router')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

afterEach(() => {
  mockNavigate.mockClear()
})

describe('useDeleteTrip', () => {
  describe('when the trip is deleted', () => {
    it('should resolve successfully', async () => {
      server.use(
        http.delete(
          `${API_BASE_URL}/trips/${testTrip.id}`,
          () => new HttpResponse(null, { status: 204 })
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useDeleteTrip(), { wrapper })

      result.current.mutate(testTrip.id)

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
    })

    it('should invalidate the trips list query', async () => {
      server.use(
        http.delete(
          `${API_BASE_URL}/trips/${testTrip.id}`,
          () => new HttpResponse(null, { status: 204 })
        )
      )
      const { wrapper, queryClient } = createQueryWrapper()
      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries')
      const { result } = renderHook(() => useDeleteTrip(), { wrapper })

      result.current.mutate(testTrip.id)

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['trips'] })
    })

    it('should navigate to the dashboard', async () => {
      server.use(
        http.delete(
          `${API_BASE_URL}/trips/${testTrip.id}`,
          () => new HttpResponse(null, { status: 204 })
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useDeleteTrip(), { wrapper })

      result.current.mutate(testTrip.id)

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard')
    })
  })

  describe('when the request fails', () => {
    it('should surface the error and not navigate', async () => {
      server.use(
        http.delete(`${API_BASE_URL}/trips/${testTrip.id}`, () =>
          HttpResponse.json(
            {
              error: 'internal_error',
              message: 'Something went wrong',
              request_id: 'req_delete_500',
            },
            { status: 500 }
          )
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useDeleteTrip(), { wrapper })

      result.current.mutate(testTrip.id)

      await waitFor(() => {
        expect(result.current.isError).toBe(true)
      })
      expect(mockNavigate).not.toHaveBeenCalled()
    })
  })
})
