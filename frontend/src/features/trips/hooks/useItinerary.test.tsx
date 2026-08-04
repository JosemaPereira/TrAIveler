import { describe, expect, it } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'

import {
  API_BASE_URL,
  testActivity,
  testItinerary,
  testItineraryDay,
  testTrip,
} from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { createQueryWrapper } from '@/test/queryWrapper'
import type { ItineraryDay } from '@/features/trips/types'
import { mapItineraryDaysToSections, useItinerary } from './useItinerary'

describe('useItinerary', () => {
  describe('when the trip exists and is owned by the caller', () => {
    it('should resolve with the itinerary the backend returns', async () => {
      server.use(
        http.get(`${API_BASE_URL}/trips/${testTrip.id}/itinerary`, () =>
          HttpResponse.json(testItinerary, { status: 200 })
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useItinerary(testTrip.id), {
        wrapper,
      })

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })
      expect(result.current.data).toEqual(testItinerary)
    })
  })

  describe('when the trip does not exist or is not owned by the caller', () => {
    it('should surface the 404 APIError', async () => {
      server.use(
        http.get(`${API_BASE_URL}/trips/not-mine/itinerary`, () =>
          HttpResponse.json(
            {
              error: 'not_found',
              message: 'Trip not found',
              request_id: 'req_itinerary_404',
            },
            { status: 404 }
          )
        )
      )
      const { wrapper } = createQueryWrapper()
      const { result } = renderHook(() => useItinerary('not-mine'), {
        wrapper,
      })

      await waitFor(() => {
        expect(result.current.isError).toBe(true)
      })
      expect(result.current.error).toMatchObject({
        status: 404,
        code: 'not_found',
      })
    })
  })
})

describe('mapItineraryDaysToSections', () => {
  describe('when a day has a label and activities', () => {
    it('should map it into DaySection/ActivityItem props', () => {
      const result = mapItineraryDaysToSections([testItineraryDay])

      expect(result).toEqual([
        {
          dayNumber: testItineraryDay.day_number,
          label: testItineraryDay.label,
          activities: [
            {
              id: testActivity.id,
              title: testActivity.title,
              type: testActivity.type,
              description: testActivity.description,
              isAIGenerated: testActivity.is_ai_generated,
            },
          ],
        },
      ])
    })
  })

  describe('when a day has no label', () => {
    it('should omit the label key rather than mapping it to undefined explicitly', () => {
      const dayWithoutLabel: ItineraryDay = {
        ...testItineraryDay,
        label: undefined,
      }

      const mapped = mapItineraryDaysToSections([dayWithoutLabel])[0]

      expect(mapped?.label).toBeUndefined()
    })
  })

  describe('when a day has no activities', () => {
    it('should map to an empty activities array', () => {
      const dayWithoutActivities: ItineraryDay = {
        ...testItineraryDay,
        activities: [],
      }

      const mapped = mapItineraryDaysToSections([dayWithoutActivities])[0]

      expect(mapped?.activities).toEqual([])
    })
  })

  describe('when given an empty days array', () => {
    it('should return an empty array', () => {
      expect(mapItineraryDaysToSections([])).toEqual([])
    })
  })
})
