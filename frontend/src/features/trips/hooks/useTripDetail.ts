import { useQuery } from '@tanstack/react-query'

import { tripKeys, tripsApi } from '@/features/trips/services/tripsApi'
import type { Trip } from '@/features/trips/types'

/**
 * Query for `GET /trips/:id`.
 *
 * Resolves with the flat `Trip` the backend actually sends. There is
 * deliberately no nested itinerary (days/activities) on this response —
 * that lives on the separate `GET /trips/:id/itinerary` endpoint, fetched
 * via `useItinerary` instead. Do not widen this hook to also carry
 * itinerary data; the two stay separate queries against separate endpoints.
 */
export function useTripDetail(id: string) {
  return useQuery<Trip>({
    queryKey: tripKeys.detail(id),
    queryFn: async () => (await tripsApi.get(id)).trip,
  })
}
