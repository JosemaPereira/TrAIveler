import { describe, expect, it } from 'vitest'

import { APIError } from './api-client'
import { mapApiError } from './error-handler'

describe('mapApiError', () => {
  describe('when handling a 401 APIError', () => {
    it('should return a non-retryable session-expired result', () => {
      const error = new APIError(
        401,
        'authentication_required',
        'Authentication required.',
        'req-401'
      )

      const result = mapApiError(error)

      expect(result).toEqual({
        title: 'Session Expired',
        message: 'Authentication required.',
        requestId: 'req-401',
        isRetryable: false,
      })
    })
  })

  describe('when handling a 403 APIError', () => {
    it('should return a non-retryable permission-denied result', () => {
      const error = new APIError(
        403,
        'forbidden',
        'You cannot do that.',
        'req-403'
      )

      const result = mapApiError(error)

      expect(result).toEqual({
        title: 'Permission Denied',
        message: 'You cannot do that.',
        requestId: 'req-403',
        isRetryable: false,
      })
    })
  })

  describe('when handling a 404 APIError', () => {
    it('should return a non-retryable not-found result', () => {
      const error = new APIError(404, 'not_found', 'Trip not found.', 'req-404')

      const result = mapApiError(error)

      expect(result).toEqual({
        title: 'Not Found',
        message: 'Trip not found.',
        requestId: 'req-404',
        isRetryable: false,
      })
    })

    it('should omit requestId when the APIError carries an empty request id', () => {
      const error = new APIError(404, 'not_found', 'Trip not found.', '')

      const result = mapApiError(error)

      expect(result.requestId).toBeUndefined()
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

      const result = mapApiError(error)

      expect(result).toEqual({
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

      const result = mapApiError(error)

      expect(result).toMatchObject({
        title: 'Something Went Wrong',
        isRetryable: true,
      })
    })
  })

  describe('when handling a 409 APIError', () => {
    it('should return a non-retryable conflict result', () => {
      const error = new APIError(
        409,
        'conflict',
        'Email already registered',
        'req-409'
      )

      const result = mapApiError(error)

      expect(result).toEqual({
        title: 'Conflict',
        message: 'Email already registered',
        requestId: 'req-409',
        isRetryable: false,
      })
    })
  })

  describe('when handling a 422 APIError', () => {
    it('should return a non-retryable invalid-input result', () => {
      const error = new APIError(
        422,
        'validation_failed',
        'One or more fields failed validation',
        'req-422',
        [{ field: 'email', error: 'Email must be a valid email address' }]
      )

      const result = mapApiError(error)

      expect(result).toEqual({
        title: 'Invalid Input',
        message: 'One or more fields failed validation',
        requestId: 'req-422',
        isRetryable: false,
      })
    })
  })

  describe('when handling a 429 APIError', () => {
    it('should return a retryable too-many-requests result', () => {
      const error = new APIError(
        429,
        'rate_limit_exceeded',
        'Too many attempts; please retry later',
        'req-429',
        undefined,
        { retry_after_seconds: 8 }
      )

      const result = mapApiError(error)

      expect(result).toEqual({
        title: 'Too Many Requests',
        message: 'Too many attempts; please retry later',
        requestId: 'req-429',
        isRetryable: true,
      })
    })
  })

  describe('when handling a non-APIError failure', () => {
    it('should fall back to a retryable generic result with a safe message', () => {
      const error = new TypeError('Failed to fetch')

      const result = mapApiError(error)

      expect(result).toEqual({
        title: 'Something Went Wrong',
        message: 'An unexpected error occurred. Please try again.',
        requestId: undefined,
        isRetryable: true,
      })
    })
  })
})
