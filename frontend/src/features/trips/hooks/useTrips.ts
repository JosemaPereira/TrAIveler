import { useQuery } from '@tanstack/react-query'

import { tripKeys, tripsApi } from '@/features/trips/services/tripsApi'
import type { Trip } from '@/features/trips/types'

/**
 * Query for `GET /trips` — every trip owned by the signed-in user.
 */
export function useTrips() {
  return useQuery<Trip[]>({
    queryKey: tripKeys.list(),
    queryFn: async () => (await tripsApi.list()).trips,
  })
}
