import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { http, HttpResponse } from 'msw'

import { apiFetch, setSessionExpiredHandler } from '../../lib/api-client'
import { useAuthStore } from '../../stores/auth-store'
import { API_BASE_URL, testUser } from '../../test/msw/handlers'
import { server } from '../../test/msw/server'
import {
  handleSessionExpired,
  installSessionExpiryHandler,
} from './session-expiry'

const assign = vi.fn()

function stubLocation(pathname: string, search = '') {
  vi.stubGlobal('location', { pathname, search, assign })
}

beforeEach(() => {
  useAuthStore.setState({
    isAuthenticated: true,
    user: testUser,
    isLoading: false,
  })
})

afterEach(() => {
  vi.unstubAllGlobals()
  assign.mockClear()
  setSessionExpiredHandler(null)
})

describe('handleSessionExpired', () => {
  describe('when the session dies on a protected page', () => {
    it('should clear the auth store', () => {
      stubLocation('/dashboard')

      handleSessionExpired()

      expect(useAuthStore.getState().isAuthenticated).toBe(false)
      expect(useAuthStore.getState().user).toBeNull()
    })

    it('should send the user to /login with the current path as the redirect target', () => {
      stubLocation('/trips/42')

      handleSessionExpired()

      expect(assign).toHaveBeenCalledWith('/login?redirect=%2Ftrips%2F42')
    })

    it('should preserve the query string in the redirect target', () => {
      stubLocation('/trips', '?sort=start_date')

      handleSessionExpired()

      expect(assign).toHaveBeenCalledWith(
        '/login?redirect=%2Ftrips%3Fsort%3Dstart_date'
      )
    })
  })

  describe('when the user is already on the login page', () => {
    it('should clear the auth store without navigating', () => {
      stubLocation('/login')

      handleSessionExpired()

      expect(useAuthStore.getState().isAuthenticated).toBe(false)
      expect(assign).not.toHaveBeenCalled()
    })
  })
})

describe('installSessionExpiryHandler', () => {
  describe('when the api client rejects a request as unauthenticated', () => {
    it('should tear the session down and redirect', async () => {
      stubLocation('/dashboard')
      server.use(
        http.get(`${API_BASE_URL}/trips`, () =>
          HttpResponse.json(
            {
              error: 'authentication_required',
              message: 'Session expired. Please log in again.',
              request_id: 'req_gone',
            },
            { status: 401 }
          )
        )
      )
      installSessionExpiryHandler()

      await expect(apiFetch('/trips')).rejects.toMatchObject({ status: 401 })

      expect(useAuthStore.getState().isAuthenticated).toBe(false)
      expect(assign).toHaveBeenCalledWith('/login?redirect=%2Fdashboard')
    })
  })
})
