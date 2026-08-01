import { beforeEach, describe, expect, it } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'

import { useAuthStore } from '@/stores/auth-store'
import { API_BASE_URL, testUser } from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { createQueryWrapper } from '@/test/queryWrapper'
import { useLogin } from './useLogin'

const validCredentials = {
  email: 'traveler@example.com',
  password: 'CorrectHorse1!',
}

beforeEach(() => {
  useAuthStore.setState({
    isAuthenticated: false,
    user: null,
    isLoading: false,
  })
})

describe('useLogin', () => {
  describe('when the credentials are valid', () => {
    it('should resolve with the authenticated user', async () => {
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useLogin(), { wrapper })

      result.current.mutate(validCredentials)

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(result.current.data?.user).toEqual(testUser)
    })

    it('should authenticate the user in the auth store', async () => {
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useLogin(), { wrapper })

      result.current.mutate(validCredentials)

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(useAuthStore.getState().isAuthenticated).toBe(true)
      expect(useAuthStore.getState().user).toEqual(testUser)
    })
  })

  describe('when the login has not been attempted yet', () => {
    it('should report no retry countdown', () => {
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useLogin(), { wrapper })

      expect(result.current.retryAfterSeconds).toBeUndefined()
    })
  })

  describe('when the credentials are rejected', () => {
    it('should surface the 401 APIError without a retry countdown', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/login`, () =>
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
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useLogin(), { wrapper })

      result.current.mutate({ ...validCredentials, password: 'wrong' })

      await waitFor(() => {
        expect(result.current.isError).toBe(true)
      })
      expect(result.current.error).toMatchObject({
        status: 401,
        code: 'authentication_required',
      })
      expect(result.current.retryAfterSeconds).toBeUndefined()
      expect(useAuthStore.getState().isAuthenticated).toBe(false)
    })
  })

  describe('when the caller is rate limited', () => {
    it('should expose the retry countdown seeded from the 429 details', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/login`, () =>
          HttpResponse.json(
            {
              error: 'rate_limit_exceeded',
              message: 'Too many login attempts. Please try again later.',
              request_id: 'req_rate',
              details: { retry_after_seconds: 45 },
            },
            { status: 429 }
          )
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useLogin(), { wrapper })

      result.current.mutate({ ...validCredentials, password: 'wrong' })

      await waitFor(() => {
        expect(result.current.isError).toBe(true)
      })
      expect(result.current.retryAfterSeconds).toBe(45)
    })
  })
})
