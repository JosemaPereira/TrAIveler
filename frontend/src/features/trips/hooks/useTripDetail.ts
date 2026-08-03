import { useQuery } from '@tanstack/react-query'

import { tripKeys, tripsApi } from '@/features/trips/services/tripsApi'
import type { Trip } from '@/features/trips/types'

/**
 * Query for `GET /trips/:id`.
 *
 * Resolves with the flat `Trip` the backend actually sends. There is no
 * nested itinerary (days/activities) on this response — no backend endpoint
 * exposes that yet, even though the trip-detail page will eventually need it
 * — mirroring the `durationDays` gap already documented on
 * `components/TripCard.tsx`. Do not fabricate itinerary fields here; that is
 * a separate, not-yet-scoped piece of work.
 */
export function useTripDetail(id: string) {
  return useQuery<Trip>({
    queryKey: tripKeys.detail(id),
    queryFn: async () => (await tripsApi.get(id)).trip,
  })
}
