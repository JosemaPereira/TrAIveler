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
