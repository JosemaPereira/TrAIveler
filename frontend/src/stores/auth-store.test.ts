import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { api, APIError } from '../lib/api-client'
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
  role: 'admin',
  subscription_id: null,
  last_login_at: null,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
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

describe('unit: useAuthStore initial state', () => {
  it('starts unauthenticated with no user and not loading', () => {
    const state = useAuthStore.getState()

    expect(state.isAuthenticated).toBe(false)
    expect(state.user).toBeNull()
    expect(state.isLoading).toBe(false)
  })
})

describe('unit: useAuthStore login', () => {
  it('sets isAuthenticated and stores the user', () => {
    useAuthStore.getState().login(testUser)

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(true)
    expect(state.user).toEqual(testUser)
  })
})

describe('unit: useAuthStore logout', () => {
  it('clears isAuthenticated and the user', () => {
    useAuthStore.getState().login(testUser)

    useAuthStore.getState().logout()

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(false)
    expect(state.user).toBeNull()
  })
})

describe('unit: useAuthStore setLoading', () => {
  it('toggles isLoading to the given value', () => {
    useAuthStore.getState().setLoading(true)
    expect(useAuthStore.getState().isLoading).toBe(true)

    useAuthStore.getState().setLoading(false)
    expect(useAuthStore.getState().isLoading).toBe(false)
  })
})

describe('unit: useAuthStore refreshSession', () => {
  it('sets isAuthenticated and the user on a successful /auth/me call', async () => {
    vi.spyOn(api, 'get').mockResolvedValue(testUser)

    await useAuthStore.getState().refreshSession()

    const state = useAuthStore.getState()
    expect(api.get).toHaveBeenCalledWith('/auth/me')
    expect(state.isAuthenticated).toBe(true)
    expect(state.user).toEqual(testUser)
    expect(state.isLoading).toBe(false)
  })

  it('clears auth state and does not throw when the call fails', async () => {
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

  it('sets isLoading to true while the call is in flight', async () => {
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

describe('unit: auth store selector hooks', () => {
  it('useUser returns the current user slice', () => {
    useAuthStore.getState().login(testUser)

    const { result } = renderHook(() => useUser())

    expect(result.current).toEqual(testUser)
  })

  it('useIsAuthenticated returns the current isAuthenticated slice', () => {
    useAuthStore.getState().login(testUser)

    const { result } = renderHook(() => useIsAuthenticated())

    expect(result.current).toBe(true)
  })

  it('useAuthLoading returns the current isLoading slice', () => {
    useAuthStore.getState().setLoading(true)

    const { result } = renderHook(() => useAuthLoading())

    expect(result.current).toBe(true)
  })

  it('re-renders selector hooks when the underlying store updates', () => {
    const { result } = renderHook(() => useUser())
    expect(result.current).toBeNull()

    act(() => {
      useAuthStore.getState().login(testUser)
    })

    expect(result.current).toEqual(testUser)
  })
})
