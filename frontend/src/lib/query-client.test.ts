import { describe, expect, it } from 'vitest'

import { APIError } from './api-client'
import {
  getErrorMessage,
  getFieldErrors,
  getRetryAfterSeconds,
  isAPIError,
  queryClient,
} from './query-client'

describe('queryClient', () => {
  describe('having the default options', () => {
    it('should set staleTime to 5 minutes', () => {
      const { queries } = queryClient.getDefaultOptions()

      expect(queries?.staleTime).toBe(5 * 60 * 1000)
    })

    it('should disable refetchOnWindowFocus', () => {
      const { queries } = queryClient.getDefaultOptions()

      expect(queries?.refetchOnWindowFocus).toBe(false)
    })

    it('should enable refetchOnReconnect', () => {
      const { queries } = queryClient.getDefaultOptions()

      expect(queries?.refetchOnReconnect).toBe(true)
    })

    it('should never retry mutations', () => {
      const { mutations } = queryClient.getDefaultOptions()

      expect(mutations?.retry).toBe(false)
    })
  })

  describe('when a query fails', () => {
    function retry(failureCount: number, error: unknown): boolean {
      const { queries } = queryClient.getDefaultOptions()
      const retryFn = queries?.retry
      if (typeof retryFn !== 'function') {
        throw new Error('expected queries.retry to be a function')
      }
      return retryFn(failureCount, error as Error)
    }

    it.each([400, 401, 403, 404, 422])(
      'should not retry on a %s APIError',
      (status) => {
        const error = new APIError(status, 'code', 'message', 'req_1')

        expect(retry(0, error)).toBe(false)
      }
    )

    it('should retry once on a 5xx APIError', () => {
      const error = new APIError(500, 'internal_error', 'message', 'req_1')

      expect(retry(0, error)).toBe(true)
    })

    it('should not retry a second time on a 5xx APIError', () => {
      const error = new APIError(500, 'internal_error', 'message', 'req_1')

      expect(retry(1, error)).toBe(false)
    })
  })
})

describe('isAPIError', () => {
  describe('when given an APIError instance', () => {
    it('should return true', () => {
      const error = new APIError(400, 'code', 'message', 'req_1')

      expect(isAPIError(error)).toBe(true)
    })
  })

  describe('when given anything else', () => {
    it('should return false for a generic Error', () => {
      expect(isAPIError(new Error('boom'))).toBe(false)
    })

    it('should return false for a non-error value', () => {
      expect(isAPIError('boom')).toBe(false)
      expect(isAPIError(undefined)).toBe(false)
    })
  })
})

describe('getErrorMessage', () => {
  describe('when given an error with a message', () => {
    it('should return the message of an APIError', () => {
      const error = new APIError(
        422,
        'validation_failed',
        'Field is invalid',
        'req_1'
      )

      expect(getErrorMessage(error)).toBe('Field is invalid')
    })

    it('should return the message of a generic Error', () => {
      expect(getErrorMessage(new Error('network down'))).toBe('network down')
    })
  })

  describe('when given an unknown error shape', () => {
    it('should return a generic fallback message', () => {
      expect(getErrorMessage('boom')).toBe('An unexpected error occurred.')
      expect(getErrorMessage(undefined)).toBe('An unexpected error occurred.')
    })
  })
})

describe('getFieldErrors', () => {
  describe('when given an APIError with field details', () => {
    it('should map the fields array into a field-to-error record', () => {
      const error = new APIError(
        422,
        'validation_failed',
        'Invalid fields',
        'req_1',
        [
          { field: 'email', error: 'Email address is already registered' },
          { field: 'start_date', error: 'Start date must be in the future' },
        ]
      )

      expect(getFieldErrors(error)).toEqual({
        email: 'Email address is already registered',
        start_date: 'Start date must be in the future',
      })
    })
  })

  describe('when given an error without field details', () => {
    it('should return an empty record for an APIError with no fields', () => {
      const error = new APIError(500, 'internal_error', 'Server error', 'req_1')

      expect(getFieldErrors(error)).toEqual({})
    })

    it('should return an empty record for a non-APIError value', () => {
      expect(getFieldErrors(new Error('boom'))).toEqual({})
      expect(getFieldErrors(undefined)).toEqual({})
    })
  })
})

describe('getRetryAfterSeconds', () => {
  function rateLimitError(details?: Record<string, unknown>): APIError {
    return new APIError(
      429,
      'rate_limit_exceeded',
      'Too many login attempts. Please try again later.',
      'req_1',
      undefined,
      details
    )
  }

  describe('when given a rate-limited APIError', () => {
    it('should return the retry_after_seconds value from details', () => {
      const error = rateLimitError({ retry_after_seconds: 30 })

      expect(getRetryAfterSeconds(error)).toBe(30)
    })

    it('should return 0 when the server reports no remaining wait', () => {
      const error = rateLimitError({ retry_after_seconds: 0 })

      expect(getRetryAfterSeconds(error)).toBe(0)
    })
  })

  describe('when the retry hint is missing or unusable', () => {
    it('should return undefined when details has no retry_after_seconds', () => {
      expect(getRetryAfterSeconds(rateLimitError({}))).toBeUndefined()
    })

    it('should return undefined when the APIError has no details at all', () => {
      expect(getRetryAfterSeconds(rateLimitError())).toBeUndefined()
    })

    it.each([
      ['a string', '30'],
      ['null', null],
      ['NaN', Number.NaN],
      ['Infinity', Number.POSITIVE_INFINITY],
    ])('should return undefined when the value is %s', (_label, value) => {
      const error = rateLimitError({ retry_after_seconds: value })

      expect(getRetryAfterSeconds(error)).toBeUndefined()
    })
  })

  describe('when given a non-APIError value', () => {
    it('should return undefined', () => {
      expect(getRetryAfterSeconds(new Error('boom'))).toBeUndefined()
      expect(getRetryAfterSeconds(undefined)).toBeUndefined()
    })
  })
})
