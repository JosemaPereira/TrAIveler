import { create } from 'zustand'

import { api } from '../lib/api-client'

// Wire-format User, typed with the same snake_case field names the backend
// returns — this codebase has no case-conversion layer (see api-client.ts),
// so the frontend type mirrors the API response as-is.
//
// This is exactly the shape `POST /auth/register` and `POST /auth/login`
// serialize today (backend/internal/auth/models.go): every other field on the
// backend's User entity — password hash, role, version, failed login counters,
// timestamps other than created_at — carries `json:"-"` and never reaches the
// client. In particular `role` is deliberately NOT exposed by the API today,
// so authorization decisions cannot be made from this type; add it here only
// once the backend actually serializes it.
export interface User {
  id: string
  email: string
  full_name: string
  has_subscription: boolean
  created_at: string
}

export interface AuthState {
  isAuthenticated: boolean
  user: User | null
  isLoading: boolean
  login: (user: User) => void
  logout: () => void
  refreshSession: () => Promise<void>
  setLoading: (loading: boolean) => void
}

// No persistence middleware: this is deliberately ephemeral state. The
// backend's HTTP-only JWT cookie is the actual persistence mechanism
// (docs/security.md), so a page reload is expected to re-derive this state
// via refreshSession() rather than rehydrate it from local storage.
export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  user: null,
  isLoading: false,

  login: (user) => {
    set({ isAuthenticated: true, user })
  },

  logout: () => {
    set({ isAuthenticated: false, user: null })
  },

  refreshSession: async () => {
    set({ isLoading: true })
    try {
      const user = await api.get<User>('/auth/me')
      set({ isAuthenticated: true, user })
    } catch {
      // Any failure (expired/invalid JWT, network error) means the caller
      // is not authenticated — never let this reject out of the action.
      set({ isAuthenticated: false, user: null })
    } finally {
      set({ isLoading: false })
    }
  },

  setLoading: (loading) => {
    set({ isLoading: loading })
  },
}))

export const useUser = (): User | null => useAuthStore((state) => state.user)

export const useIsAuthenticated = (): boolean =>
  useAuthStore((state) => state.isAuthenticated)

export const useAuthLoading = (): boolean =>
  useAuthStore((state) => state.isLoading)
