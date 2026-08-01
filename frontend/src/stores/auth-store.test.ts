import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { api, APIError } from '@/lib/api-client'
import type { User } from './auth-store'
import {
  useAuthLoading,
  useAuthStore,
  useIsAuthenticated,
  useUser,
} from './auth-store'

const testUser: User = {
  id: 'user_1',
  email: 'traveler@example.com',
  full_name: 'Ada Traveler',
  has_subscription: false,
  created_at: '2026-01-01T00:00:00Z',
}

// Zustand stores are module-singletons: reset state before every test so
// assertions never leak from one test into the next.
beforeEach(() => {
  useAuthStore.setState({
    isAuthenticated: false,
    user: null,
    isLoading: false,
  })
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe('useAuthStore', () => {
  describe('when initialized', () => {
    it('should start unauthenticated with no user and not loading', () => {
      const state = useAuthStore.getState()

      expect(state.isAuthenticated).toBe(false)
      expect(state.user).toBeNull()
      expect(state.isLoading).toBe(false)
    })
  })

  describe('when login is called', () => {
    it('should set isAuthenticated and store the user', () => {
      useAuthStore.getState().login(testUser)

      const state = useAuthStore.getState()

      expect(state.isAuthenticated).toBe(true)
      expect(state.user).toEqual(testUser)
    })
  })

  describe('when logout is called', () => {
    it('should clear isAuthenticated and the user', () => {
      useAuthStore.getState().login(testUser)

      useAuthStore.getState().logout()

      const state = useAuthStore.getState()
      expect(state.isAuthenticated).toBe(false)
      expect(state.user).toBeNull()
    })
  })

  describe('when setUser is called', () => {
    it('should update the stored user', () => {
      useAuthStore.getState().login(testUser)
      const updatedUser: User = { ...testUser, full_name: 'Ada Explorer' }

      useAuthStore.getState().setUser(updatedUser)

      expect(useAuthStore.getState().user).toEqual(updatedUser)
    })

    it('should not change isAuthenticated', () => {
      useAuthStore.setState({ isAuthenticated: false, user: null })

      useAuthStore.getState().setUser(testUser)

      const state = useAuthStore.getState()
      expect(state.user).toEqual(testUser)
      expect(state.isAuthenticated).toBe(false)
    })
  })

  describe('when setLoading is called', () => {
    it('should toggle isLoading to the given value', () => {
      useAuthStore.getState().setLoading(true)
      expect(useAuthStore.getState().isLoading).toBe(true)

      useAuthStore.getState().setLoading(false)

      expect(useAuthStore.getState().isLoading).toBe(false)
    })
  })

  describe('when refreshSession succeeds', () => {
    // The endpoint answers with a { user } envelope, matching register and
    // login. This test previously mocked a bare user, encoding a contract the
    // backend never had — it passed only because no such endpoint existed yet.
    it('should set isAuthenticated and the user from /auth/me', async () => {
      vi.spyOn(api, 'get').mockResolvedValue({ user: testUser })

      await useAuthStore.getState().refreshSession()

      const state = useAuthStore.getState()
      expect(api.get).toHaveBeenCalledWith('/auth/me')
      expect(state.isAuthenticated).toBe(true)
      expect(state.user).toEqual(testUser)
      expect(state.isLoading).toBe(false)
    })

    it('should unwrap the user envelope rather than storing the envelope', async () => {
      vi.spyOn(api, 'get').mockResolvedValue({ user: testUser })

      await useAuthStore.getState().refreshSession()

      expect(useAuthStore.getState().user).not.toHaveProperty('user')
    })
  })

  describe('when refreshSession fails', () => {
    it('should clear auth state and not throw', async () => {
      vi.spyOn(api, 'get').mockRejectedValue(
        new APIError(401, 'authentication_required', 'Missing session', 'req_1')
      )
      useAuthStore.setState({ isAuthenticated: true, user: testUser })

      await expect(
        useAuthStore.getState().refreshSession()
      ).resolves.toBeUndefined()

      const state = useAuthStore.getState()
      expect(state.isAuthenticated).toBe(false)
      expect(state.user).toBeNull()
      expect(state.isLoading).toBe(false)
    })
  })

  describe('when refreshSession is in flight', () => {
    it('should set isLoading to true until the call settles', async () => {
      let resolveGet!: (value: User) => void
      vi.spyOn(api, 'get').mockReturnValue(
        new Promise((resolve) => {
          resolveGet = resolve
        })
      )

      const pending = useAuthStore.getState().refreshSession()
      expect(useAuthStore.getState().isLoading).toBe(true)
      resolveGet(testUser)
      await pending

      expect(useAuthStore.getState().isLoading).toBe(false)
    })
  })
})

describe('auth store selector hooks', () => {
  describe('when the store holds state', () => {
    it('should return the current user slice from useUser', () => {
      useAuthStore.getState().login(testUser)

      const { result } = renderHook(() => useUser())

      expect(result.current).toEqual(testUser)
    })

    it('should return the current isAuthenticated slice from useIsAuthenticated', () => {
      useAuthStore.getState().login(testUser)

      const { result } = renderHook(() => useIsAuthenticated())

      expect(result.current).toBe(true)
    })

    it('should return the current isLoading slice from useAuthLoading', () => {
      useAuthStore.getState().setLoading(true)

      const { result } = renderHook(() => useAuthLoading())

      expect(result.current).toBe(true)
    })
  })

  describe('when the underlying store updates', () => {
    it('should re-render subscribed selector hooks', () => {
      const { result } = renderHook(() => useUser())
      expect(result.current).toBeNull()

      act(() => {
        useAuthStore.getState().login(testUser)
      })

      expect(result.current).toEqual(testUser)
    })
  })
})
