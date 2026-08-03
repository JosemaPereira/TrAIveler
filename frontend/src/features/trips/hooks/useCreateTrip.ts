import { useMutation, useQueryClient } from '@tanstack/react-query'

import { tripKeys, tripsApi } from '@/features/trips/services/tripsApi'
import type { CreateTripRequest, Trip } from '@/features/trips/types'

/**
 * Mutation for `POST /trips`.
 *
 * On success invalidates the trips list query so a newly created trip shows
 * up without a manual refetch.
 */
export function useCreateTrip() {
  const queryClient = useQueryClient()

  return useMutation<Trip, Error, CreateTripRequest>({
    mutationFn: async (request) => (await tripsApi.create(request)).trip,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: tripKeys.list() })
    },
  })
}
