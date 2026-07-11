import { describe, expect, it } from 'vitest'

import { APIError } from './api-client'
import { getErrorMessage, getFieldErrors, isAPIError, queryClient } from './query-client'

describe('unit: queryClient configuration', () => {
  it('sets staleTime to 5 minutes', () => {
    const { queries } = queryClient.getDefaultOptions()

    expect(queries?.staleTime).toBe(5 * 60 * 1000)
  })

  it('disables refetchOnWindowFocus', () => {
    const { queries } = queryClient.getDefaultOptions()

    expect(queries?.refetchOnWindowFocus).toBe(false)
  })

  it('enables refetchOnReconnect', () => {
    const { queries } = queryClient.getDefaultOptions()

    expect(queries?.refetchOnReconnect).toBe(true)
  })

  it('never retries mutations', () => {
    const { mutations } = queryClient.getDefaultOptions()

    expect(mutations?.retry).toBe(false)
  })
})

describe('unit: queryClient query retry logic', () => {
  function retry(failureCount: number, error: unknown): boolean {
    const { queries } = queryClient.getDefaultOptions()
    const retryFn = queries?.retry
    if (typeof retryFn !== 'function') {
      throw new Error('expected queries.retry to be a function')
    }
    return retryFn(failureCount, error as Error)
  }

  it.each([400, 401, 403, 404, 422])(
    'does not retry on a %s APIError',
    (status) => {
      const error = new APIError(status, 'code', 'message', 'req_1')

      expect(retry(0, error)).toBe(false)
    }
  )

  it('retries once on a 5xx APIError', () => {
    const error = new APIError(500, 'internal_error', 'message', 'req_1')

    expect(retry(0, error)).toBe(true)
  })

  it('does not retry a second time on a 5xx APIError', () => {
    const error = new APIError(500, 'internal_error', 'message', 'req_1')

    expect(retry(1, error)).toBe(false)
  })
})

describe('unit: isAPIError', () => {
  it('returns true for an APIError instance', () => {
    expect(isAPIError(new APIError(400, 'code', 'message', 'req_1'))).toBe(true)
  })

  it('returns false for a generic Error', () => {
    expect(isAPIError(new Error('boom'))).toBe(false)
  })

  it('returns false for a non-error value', () => {
    expect(isAPIError('boom')).toBe(false)
    expect(isAPIError(undefined)).toBe(false)
  })
})

describe('unit: getErrorMessage', () => {
  it('returns the message of an APIError', () => {
    const error = new APIError(422, 'validation_failed', 'Field is invalid', 'req_1')

    expect(getErrorMessage(error)).toBe('Field is invalid')
  })

  it('returns the message of a generic Error', () => {
    expect(getErrorMessage(new Error('network down'))).toBe('network down')
  })

  it('returns a generic fallback message for unknown error shapes', () => {
    expect(getErrorMessage('boom')).toBe('An unexpected error occurred.')
    expect(getErrorMessage(undefined)).toBe('An unexpected error occurred.')
  })
})

describe('unit: getFieldErrors', () => {
  it('maps an APIError fields array into a field-to-error record', () => {
    const error = new APIError(422, 'validation_failed', 'Invalid fields', 'req_1', [
      { field: 'email', error: 'Email address is already registered' },
      { field: 'start_date', error: 'Start date must be in the future' },
    ])

    expect(getFieldErrors(error)).toEqual({
      email: 'Email address is already registered',
      start_date: 'Start date must be in the future',
    })
  })

  it('returns an empty record when the APIError has no fields', () => {
    const error = new APIError(500, 'internal_error', 'Server error', 'req_1')

    expect(getFieldErrors(error)).toEqual({})
  })

  it('returns an empty record for a non-APIError value', () => {
    expect(getFieldErrors(new Error('boom'))).toEqual({})
    expect(getFieldErrors(undefined)).toEqual({})
  })
})
