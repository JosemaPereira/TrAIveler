import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router'

import { tripKeys, tripsApi } from '@/features/trips/services/tripsApi'

/**
 * Mutation for `DELETE /trips/:id`.
 *
 * On success invalidates the trips list query and redirects to `/dashboard`
 * — same idiom as `features/auth/hooks/useLogout.ts`'s `void navigate(...)`,
 * since react-router's navigate is async and nothing here depends on the
 * transition completing.
 */
export function useDeleteTrip() {
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  return useMutation({
    mutationFn: (id: string) => tripsApi.delete(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: tripKeys.list() })
      void navigate('/dashboard')
    },
  })
}
