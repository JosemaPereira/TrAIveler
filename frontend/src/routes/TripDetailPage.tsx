import { useState } from 'react'
import type { ChangeEvent, SyntheticEvent } from 'react'
import { useParams } from 'react-router'

import { Navigation } from '@/components/composites/Navigation'
import { Button } from '@/components/primitives/Button'
import { ErrorMessage } from '@/components/primitives/ErrorMessage'
import { Input } from '@/components/primitives/Input'
import { LoadingSpinner } from '@/components/primitives/LoadingSpinner'
import { DeleteTripModal } from '@/features/trips/components/DeleteTripModal'
import { ItineraryView } from '@/features/trips/components/ItineraryView'
import {
  mapItineraryDaysToSections,
  useItinerary,
} from '@/features/trips/hooks/useItinerary'
import { useTripDetail } from '@/features/trips/hooks/useTripDetail'
import { useUpdateTrip } from '@/features/trips/hooks/useUpdateTrip'
import { useErrorHandler } from '@/hooks/useErrorHandler'
import { mapApiError } from '@/lib/error-handler'
import { useUser } from '@/stores/auth-store'
import styles from './TripDetailPage.module.css'

const EDIT_LABEL = 'Edit Trip'
const DELETE_LABEL = 'Delete Trip'
const SAVE_LABEL = 'Save'
const CANCEL_EDIT_LABEL = 'Cancel Edit'
const TITLE_INPUT_LABEL = 'Title'
const DESCRIPTION_INPUT_LABEL = 'Description'

/**
 * Trip detail page (`/trips/:id`). Fetches the flat `Trip` `useTripDetail`
 * resolves for the header/action bar, and the nested itinerary (days,
 * activities) separately via `useItinerary` — two queries against two
 * endpoints, not one combined fetch, since `GET /trips/:id` deliberately
 * still returns only the flat Trip. The itinerary fetch's own
 * loading/error/retry state is handed straight to `ItineraryView`'s built-in
 * props rather than hand-rolled here, since it only affects that one section
 * of the page, not the page as a whole (contrast with `tripQuery`'s
 * `errorInfo`, which gates the entire page).
 */
export function TripDetailPage() {
  const { id } = useParams()
  const user = useUser()
  const tripQuery = useTripDetail(id ?? '')
  const itineraryQuery = useItinerary(id ?? '')
  const updateTrip = useUpdateTrip()
  const errorInfo = useErrorHandler(tripQuery.error)
  const itineraryErrorInfo = itineraryQuery.isError
    ? mapApiError(itineraryQuery.error)
    : null

  const [isEditing, setIsEditing] = useState(false)
  const [editTitle, setEditTitle] = useState('')
  const [editDescription, setEditDescription] = useState('')
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)

  if (tripQuery.isLoading) {
    return (
      <div>
        <Navigation />
        <LoadingSpinner label="Loading trip..." />
      </div>
    )
  }

  if (tripQuery.isError || !tripQuery.data) {
    return (
      <div>
        <Navigation />
        {errorInfo && (
          <ErrorMessage
            title={errorInfo.title}
            message={errorInfo.message}
            requestId={errorInfo.requestId}
            onRetry={
              errorInfo.isRetryable ? () => void tripQuery.refetch() : undefined
            }
          />
        )}
      </div>
    )
  }

  const trip = tripQuery.data
  const isCreator = user?.id === trip.creator_id
  // Action bar (008-T118): only the trip's creator, with an active
  // subscription, may edit or delete it. has_subscription is the only signal
  // the frontend currently has for that (see the grace-period comment below).
  const canManageTrip = isCreator && user.has_subscription
  const updateErrorInfo = updateTrip.isError
    ? mapApiError(updateTrip.error)
    : null

  function startEditing() {
    setEditTitle(trip.title)
    setEditDescription(trip.description ?? '')
    setIsEditing(true)
  }

  function cancelEditing() {
    setIsEditing(false)
  }

  function handleSaveEdit(event: SyntheticEvent<HTMLFormElement>) {
    event.preventDefault()
    updateTrip.mutate(
      {
        id: trip.id,
        version: trip.version,
        request: {
          title: editTitle,
          description: editDescription || undefined,
          // Status is intentionally left unchanged — this ticket is scoped
          // to title/description editing only, not a status-changing UI.
          status: trip.status,
        },
      },
      {
        onSuccess: () => {
          setIsEditing(false)
        },
      }
    )
  }

  return (
    <div>
      <Navigation />
      <h1>{trip.title}</h1>
      {trip.description && <p>{trip.description}</p>}

      {/*
       * Grace-period banner deliberately not implemented: there is currently
       * no wired way to know whether this trip's creator is inside a
       * subscription grace period. `User` (stores/auth-store.ts) only
       * exposes `has_subscription: boolean`; `Subscription.grace_period_ends_at`
       * (features/auth/types.ts) is delivered exactly once, on the
       * `POST /auth/register` response, and is never persisted into the auth
       * store afterward — and there is no `GET /subscription` endpoint to
       * fetch it later. Treating `has_subscription === false` as "in grace
       * period" would be misleading: a creator who never had a subscription
       * at all is a different state from one whose subscription just entered
       * its grace window. Skipping the banner until a real subscription
       * status source exists, rather than rendering it off the wrong signal.
       */}

      {canManageTrip && (
        <div className={styles.actionBar}>
          {isEditing ? (
            <form onSubmit={handleSaveEdit} className={styles.editForm}>
              <Input
                label={TITLE_INPUT_LABEL}
                value={editTitle}
                onChange={(event: ChangeEvent<HTMLInputElement>) => {
                  setEditTitle(event.target.value)
                }}
                required
              />
              <Input
                label={DESCRIPTION_INPUT_LABEL}
                value={editDescription}
                onChange={(event: ChangeEvent<HTMLInputElement>) => {
                  setEditDescription(event.target.value)
                }}
              />
              {updateErrorInfo && (
                <p role="alert" className={styles.editError}>
                  {updateErrorInfo.message}
                </p>
              )}
              <div className={styles.editActions}>
                <Button type="submit" disabled={updateTrip.isPending}>
                  {SAVE_LABEL}
                </Button>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={cancelEditing}
                  disabled={updateTrip.isPending}
                >
                  {CANCEL_EDIT_LABEL}
                </Button>
              </div>
            </form>
          ) : (
            <>
              <Button onClick={startEditing}>{EDIT_LABEL}</Button>
              <Button
                variant="danger"
                onClick={() => {
                  setIsDeleteModalOpen(true)
                }}
              >
                {DELETE_LABEL}
              </Button>
            </>
          )}
        </div>
      )}

      <ItineraryView
        days={mapItineraryDaysToSections(itineraryQuery.data?.days ?? [])}
        isLoading={itineraryQuery.isLoading}
        error={itineraryErrorInfo?.message}
        onRetry={() => void itineraryQuery.refetch()}
      />

      <DeleteTripModal
        tripId={trip.id}
        isOpen={isDeleteModalOpen}
        onClose={() => {
          setIsDeleteModalOpen(false)
        }}
      />
    </div>
  )
}
