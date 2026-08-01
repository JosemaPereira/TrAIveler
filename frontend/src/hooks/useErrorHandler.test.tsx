import { describe, expect, it, vi, afterEach } from 'vitest'
import { renderHook } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import type { ReactNode } from 'react'

import { APIError } from '@/lib/api-client'
import { useErrorHandler } from './useErrorHandler'

const mockNavigate = vi.fn()

vi.mock('react-router', async () => {
  const actual =
    await vi.importActual<typeof import('react-router')>('react-router')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

function wrapper({ children }: { children: ReactNode }) {
  return <MemoryRouter>{children}</MemoryRouter>
}

afterEach(() => {
  mockNavigate.mockClear()
})

describe('useErrorHandler', () => {
  describe('when there is no error', () => {
    it('should return null for a null error', () => {
      const { result } = renderHook(() => useErrorHandler(null), { wrapper })

      expect(result.current).toBeNull()
    })

    it('should return null for an undefined error', () => {
      const { result } = renderHook(() => useErrorHandler(undefined), {
        wrapper,
      })

      expect(result.current).toBeNull()
    })
  })

  describe('when handling a 401 APIError', () => {
    it('should redirect to /login and return a session-expired result', () => {
      const error = new APIError(
        401,
        'authentication_required',
        'Authentication required.',
        'req-401'
      )

      const { result } = renderHook(() => useErrorHandler(error), { wrapper })

      expect(mockNavigate).toHaveBeenCalledWith('/login')
      expect(result.current).toEqual({
        title: 'Session Expired',
        message: 'Authentication required.',
        requestId: 'req-401',
        isRetryable: false,
      })
    })
  })

  describe('when handling a 403 APIError', () => {
    it('should return a permission-denied result without redirecting', () => {
      const error = new APIError(
        403,
        'forbidden',
        'You cannot do that.',
        'req-403'
      )

      const { result } = renderHook(() => useErrorHandler(error), { wrapper })

      expect(mockNavigate).not.toHaveBeenCalled()
      expect(result.current).toEqual({
        title: 'Permission Denied',
        message: 'You cannot do that.',
        requestId: 'req-403',
        isRetryable: false,
      })
    })
  })

  describe('when handling a 404 APIError', () => {
    it('should return a not-found result', () => {
      const error = new APIError(404, 'not_found', 'Trip not found.', 'req-404')

      const { result } = renderHook(() => useErrorHandler(error), { wrapper })

      expect(mockNavigate).not.toHaveBeenCalled()
      expect(result.current).toEqual({
        title: 'Not Found',
        message: 'Trip not found.',
        requestId: 'req-404',
        isRetryable: false,
      })
    })

    it('should omit requestId when the APIError carries an empty request id', () => {
      const error = new APIError(404, 'not_found', 'Trip not found.', '')

      const { result } = renderHook(() => useErrorHandler(error), { wrapper })

      expect(result.current?.requestId).toBeUndefined()
    })
  })

  describe('when handling a server-side APIError', () => {
    it('should return a retryable result for a 503', () => {
      const error = new APIError(
        503,
        'service_unavailable',
        'The service is temporarily unavailable.',
        'req-503'
      )

      const { result } = renderHook(() => useErrorHandler(error), { wrapper })

      expect(mockNavigate).not.toHaveBeenCalled()
      expect(result.current).toEqual({
        title: 'Something Went Wrong',
        message: 'The service is temporarily unavailable.',
        requestId: 'req-503',
        isRetryable: true,
      })
    })

    it('should return a retryable result for a 500', () => {
      const error = new APIError(
        500,
        'internal_error',
        'Internal error.',
        'req-500'
      )

      const { result } = renderHook(() => useErrorHandler(error), { wrapper })

      expect(result.current).toMatchObject({
        title: 'Something Went Wrong',
        isRetryable: true,
      })
    })
  })

  describe('when handling a non-APIError network failure', () => {
    it('should fall back to a retryable, generic result', () => {
      const error = new TypeError('Failed to fetch')

      const { result } = renderHook(() => useErrorHandler(error), { wrapper })

      expect(mockNavigate).not.toHaveBeenCalled()
      expect(result.current).toMatchObject({
        title: 'Something Went Wrong',
        isRetryable: true,
      })
      expect(result.current?.requestId).toBeUndefined()
    })
  })
})
