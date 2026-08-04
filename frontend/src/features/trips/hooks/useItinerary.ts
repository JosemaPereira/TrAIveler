import { useQuery } from '@tanstack/react-query'

import type { ActivityData } from '@/components/composites/ActivityItem'
import type { DaySectionProps } from '@/components/composites/DaySection'
import { tripKeys, tripsApi } from '@/features/trips/services/tripsApi'
import type {
  Activity,
  GetItineraryResponse,
  ItineraryDay,
} from '@/features/trips/types'

export interface UseItineraryOptions {
  /**
   * Whether the query should run at all. Defaults to `true`, matching
   * `useTripDetail`'s always-on behavior. `GeneratePage` passes `false`
   * until the conversation reports `itinerary_ready`, so this doesn't fire
   * (and fail, since the trip exists but has no itinerary yet) the moment a
   * trip id becomes available.
   */
  enabled?: boolean
}

/**
 * Query for `GET /trips/:id/itinerary`.
 *
 * Resolves with the full nested itinerary (days, each with its destination
 * and activities) the backend assembles. Complements `useTripDetail`, which
 * only ever returns the flat `Trip` — the two are separate queries because
 * they hit separate endpoints, not one combined "trip + itinerary" call.
 */
export function useItinerary(id: string, options: UseItineraryOptions = {}) {
  return useQuery<GetItineraryResponse>({
    queryKey: tripKeys.itinerary(id),
    queryFn: () => tripsApi.getItinerary(id),
    enabled: options.enabled ?? true,
  })
}

function mapActivityToActivityData(activity: Activity): ActivityData {
  return {
    id: activity.id,
    title: activity.title,
    type: activity.type,
    description: activity.description,
    isAIGenerated: activity.is_ai_generated,
  }
}

function mapItineraryDayToSection(
  day: ItineraryDay
): Omit<DaySectionProps, 'isLoading'> {
  return {
    dayNumber: day.day_number,
    label: day.label,
    activities: day.activities.map(mapActivityToActivityData),
  }
}

/**
 * Maps `GetItineraryResponse['days']` (wire shape, snake_case) into the
 * props `ItineraryView`/`DaySection`/`ActivityItem` actually render
 * (camelCase, and dropping fields the UI doesn't consume: `destination`,
 * `day_id`, `sequence_order`, `metadata`, `version`, timestamps). Exported so
 * both `ItineraryView` call sites (`TripDetailPage`, `GeneratePage`) share
 * one mapping instead of duplicating it.
 */
export function mapItineraryDaysToSections(
  days: ItineraryDay[]
): Omit<DaySectionProps, 'isLoading'>[] {
  return days.map(mapItineraryDayToSection)
}
