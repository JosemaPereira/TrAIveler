import { api } from '@/lib/api-client'
import type {
  CreateTripRequest,
  CreateTripResponse,
  GetTripResponse,
  ListTripsResponse,
  UpdateTripRequest,
  UpdateTripResponse,
} from '@/features/trips/types'

/**
 * Query-key convention for the trips feature, following TanStack Query's
 * hierarchical array convention (`['trips']` for the list, `['trips', id]`
 * for a single trip, so invalidating the list also matches any per-trip
 * detail lookup that shares the prefix). There is no prior precedent for
 * query keys in this codebase — this is the first `useQuery` hook — so this
 * shape is defined once here, next to the service that fetches the data, and
 * every hook in `hooks/` imports it rather than hand-rolling its own key.
 */
export const tripKeys = {
  list: () => ['trips'] as const,
  detail: (id: string) => ['trips', id] as const,
}

/**
 * HTTP surface for trips (`/trips`), built on the same fetch-based
 * `lib/api-client.ts` as `authApi.ts` — see that file's doc comment for the
 * shared conventions (base URL resolution, credentials, APIError on non-2xx).
 */
export const tripsApi = {
  /** Lists every trip owned by the signed-in user. No pagination envelope — the backend doesn't implement one yet. */
  list: (): Promise<ListTripsResponse> => api.get<ListTripsResponse>('/trips'),

  /** Creates a trip, defaulting to 'draft' status server-side. */
  create: (request: CreateTripRequest): Promise<CreateTripResponse> =>
    api.post<CreateTripResponse>('/trips', request),

  /**
   * Resolves a single trip by id. Rejects with a 404 APIError both when the
   * trip does not exist and when it belongs to another user — anti-
   * enumeration, same shape either way, nothing to special-case here.
   */
  get: (id: string): Promise<GetTripResponse> =>
    api.get<GetTripResponse>(`/trips/${id}`),

  /**
   * Updates a trip. `version` is the trip's current optimistic-locking
   * version (docs/data-model.md) and travels as an `If-Match` header rather
   * than a body field — a stale version rejects with a 409 conflict
   * APIError.
   */
  update: (
    id: string,
    version: number,
    request: UpdateTripRequest
  ): Promise<UpdateTripResponse> =>
    api.put<UpdateTripResponse>(`/trips/${id}`, request, {
      headers: { 'If-Match': String(version) },
    }),

  /**
   * Deletes a trip. No `If-Match` header — `trip.Service.Delete` takes no
   * version parameter, a deliberate backend deviation from the blanket
   * "all PUT/DELETE require If-Match" statement in docs/data-model.md.
   */
  delete: (id: string): Promise<void> => api.delete(`/trips/${id}`),
}
