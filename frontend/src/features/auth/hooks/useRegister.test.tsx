import { beforeEach, describe, expect, it } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'

import { useAuthStore } from '@/stores/auth-store'
import {
  API_BASE_URL,
  testSubscription,
  testUser,
} from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { createQueryWrapper } from '@/test/queryWrapper'
import { useRegister } from './useRegister'

const validRegistration = {
  email: 'traveler@example.com',
  password: 'CorrectHorse1!',
  full_name: 'Ada Traveler',
}

beforeEach(() => {
  useAuthStore.setState({
    isAuthenticated: false,
    user: null,
    isLoading: false,
  })
})

describe('useRegister', () => {
  describe('when the registration succeeds', () => {
    it('should resolve with the created user', async () => {
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useRegister(), { wrapper })

      result.current.mutate(validRegistration)

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(result.current.data?.user).toEqual(testUser)
    })

    it('should authenticate the user in the auth store', async () => {
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useRegister(), { wrapper })

      result.current.mutate(validRegistration)

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(useAuthStore.getState().isAuthenticated).toBe(true)
      expect(useAuthStore.getState().user).toEqual(testUser)
    })

    it('should expose the subscription created for a paid registration', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/register`, () =>
          HttpResponse.json(
            {
              user: { ...testUser, has_subscription: true },
              subscription: testSubscription,
            },
            { status: 201 }
          )
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useRegister(), { wrapper })

      result.current.mutate({
        ...validRegistration,
        payment_method_token: 'tok_visa_demo',
      })

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(result.current.data?.subscription).toEqual(testSubscription)
      expect(useAuthStore.getState().user?.has_subscription).toBe(true)
    })
  })

  describe('when the email is already registered', () => {
    it('should surface the 409 APIError and leave the store unauthenticated', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/register`, () =>
          HttpResponse.json(
            {
              error: 'conflict',
              message: 'Email already registered',
              request_id: 'req_conflict',
            },
            { status: 409 }
          )
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useRegister(), { wrapper })

      result.current.mutate(validRegistration)

      await waitFor(() => {
        expect(result.current.isError).toBe(true)
      })
      expect(result.current.error).toMatchObject({
        status: 409,
        code: 'conflict',
      })
      expect(useAuthStore.getState().isAuthenticated).toBe(false)
    })
  })
})
