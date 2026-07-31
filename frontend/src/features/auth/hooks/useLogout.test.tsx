import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'

import { useAuthStore } from '../../../stores/auth-store'
import { API_BASE_URL, testUser } from '../../../test/msw/handlers'
import { server } from '../../../test/msw/server'
import { createQueryWrapper } from '../../../test/queryWrapper'
import { useLogout } from './useLogout'

const mockNavigate = vi.fn()

vi.mock('react-router', async () => {
  const actual =
    await vi.importActual<typeof import('react-router')>('react-router')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

beforeEach(() => {
  useAuthStore.setState({
    isAuthenticated: true,
    user: testUser,
    isLoading: false,
  })
})

afterEach(() => {
  mockNavigate.mockClear()
})

describe('useLogout', () => {
  describe('when the logout succeeds', () => {
    it('should clear the auth store', async () => {
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useLogout(), { wrapper })

      result.current.mutate()

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(useAuthStore.getState().isAuthenticated).toBe(false)
      expect(useAuthStore.getState().user).toBeNull()
    })

    it('should clear the query cache so no other user sees the previous session data', async () => {
      const { wrapper, queryClient } = createQueryWrapper()
      const clearSpy = vi.spyOn(queryClient, 'clear')
      const { result } = renderHook(() => useLogout(), { wrapper })

      result.current.mutate()

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(clearSpy).toHaveBeenCalledOnce()
    })

    it('should navigate to the login page', async () => {
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useLogout(), { wrapper })

      result.current.mutate()

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(mockNavigate).toHaveBeenCalledWith('/login')
    })
  })

  describe('when the logout request fails', () => {
    it('should surface the error and keep the local session untouched', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/logout`, () =>
          HttpResponse.json(
            {
              error: 'internal_error',
              message: 'Something went wrong',
              request_id: 'req_logout_500',
            },
            { status: 500 }
          )
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useLogout(), { wrapper })

      result.current.mutate()

      await waitFor(() => {
        expect(result.current.isError).toBe(true)
      })
      expect(useAuthStore.getState().isAuthenticated).toBe(true)
      expect(mockNavigate).not.toHaveBeenCalled()
    })
  })
})
