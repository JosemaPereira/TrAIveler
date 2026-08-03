import { useMutation, useQueryClient } from '@tanstack/react-query'

import { tripKeys, tripsApi } from '@/features/trips/services/tripsApi'
import type { Trip, UpdateTripRequest } from '@/features/trips/types'

export interface UpdateTripVariables {
  id: string
  /** The trip's current optimistic-locking version (docs/data-model.md), sent as an `If-Match` header. */
  version: number
  request: UpdateTripRequest
}

/**
 * Mutation for `PUT /trips/:id`.
 *
 * `version` travels separately from `request` because the backend expects it
 * as an `If-Match` header, not a body field — see `tripsApi.update`. A stale
 * version rejects with a 409 conflict APIError; surfacing that to the user
 * (e.g. "someone else edited this trip, refresh") is the caller's job.
 *
 * On success invalidates both this trip's detail query and the trips list,
 * since a title/status change can affect either view.
 */
export function useUpdateTrip() {
  const queryClient = useQueryClient()

  return useMutation<Trip, Error, UpdateTripVariables>({
    mutationFn: async ({ id, version, request }) =>
      (await tripsApi.update(id, version, request)).trip,
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({
        queryKey: tripKeys.detail(variables.id),
      })
      void queryClient.invalidateQueries({ queryKey: tripKeys.list() })
    },
  })
}
