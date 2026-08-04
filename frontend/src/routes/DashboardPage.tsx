import { useNavigate } from 'react-router'

import { Navigation } from '@/components/composites/Navigation'
import { Button } from '@/components/primitives/Button'
import { EmptyState } from '@/components/primitives/EmptyState'
import { ErrorMessage } from '@/components/primitives/ErrorMessage'
import { TripCard } from '@/features/trips/components/TripCard'
import { useTrips } from '@/features/trips/hooks/useTrips'
import { useErrorHandler } from '@/hooks/useErrorHandler'
import { useUser } from '@/stores/auth-store'
import styles from './DashboardPage.module.css'

const NEW_TRIP_LABEL = 'New Trip'
const EMPTY_STATE_TITLE = 'No trips yet'
const LOADING_SKELETON_COUNT = 3
const GENERATE_PATH = '/generate'

// Neither the trip-list nor trip-detail API contracts expose destination
// names (see TripCard.tsx's own doc comment); TripCard already renders a
// "Destinations coming soon" fallback for an empty array, so that is passed
// as-is rather than inventing data.
const NO_DESTINATIONS: string[] = []

// durationDays is a required TripCard prop, but there is no source data for
// it anywhere in the Trip payload (no start/end date, no day count) — the
// same gap TripCard.tsx's own doc comment names for the missing "duration"
// field. A fixed placeholder is used here, deliberately, until the backend
// exposes real trip-duration data, rather than silently fabricating a
// per-trip number that would look real but isn't.
const PLACEHOLDER_DURATION_DAYS = 1

/**
 * Authenticated dashboard (`/dashboard`): lists every trip owned by the
 * signed-in user and offers a "New Trip" CTA into the trip-generation flow.
 */
export function DashboardPage() {
  const user = useUser()
  const navigate = useNavigate()
  const tripsQuery = useTrips()
  const errorInfo = useErrorHandler(tripsQuery.error)

  // Same has_subscription gate as the rest of the trip-management surface
  // (see TripDetailPage.tsx's action-bar comment) — no prior "disable a CTA
  // off has_subscription" precedent exists elsewhere in the codebase to
  // mirror (SettingsPage.tsx is still a placeholder), so this is the first
  // instance of the idiom.
  const canCreateTrip = user?.has_subscription === true

  return (
    <div>
      <Navigation />
      <h1>Welcome back{user ? `, ${user.full_name}` : ''}</h1>

      <Button
        onClick={() => {
          void navigate(GENERATE_PATH)
        }}
        disabled={!canCreateTrip}
      >
        {NEW_TRIP_LABEL}
      </Button>

      {tripsQuery.isLoading && (
        <div className={styles.grid}>
          {Array.from({ length: LOADING_SKELETON_COUNT }, (_, index) => (
            <TripCard
              key={index}
              id=""
              title=""
              status="draft"
              destinationNames={NO_DESTINATIONS}
              durationDays={PLACEHOLDER_DURATION_DAYS}
              createdAt=""
              isLoading
            />
          ))}
        </div>
      )}

      {tripsQuery.isError && errorInfo && (
        <ErrorMessage
          title={errorInfo.title}
          message={errorInfo.message}
          requestId={errorInfo.requestId}
          onRetry={
            errorInfo.isRetryable ? () => void tripsQuery.refetch() : undefined
          }
        />
      )}

      {tripsQuery.isSuccess && tripsQuery.data.length === 0 && (
        <EmptyState title={EMPTY_STATE_TITLE} />
      )}

      {tripsQuery.isSuccess && tripsQuery.data.length > 0 && (
        <div className={styles.grid}>
          {tripsQuery.data.map((trip) => (
            <TripCard
              key={trip.id}
              id={trip.id}
              title={trip.title}
              status={trip.status}
              destinationNames={NO_DESTINATIONS}
              durationDays={PLACEHOLDER_DURATION_DAYS}
              createdAt={trip.created_at}
            />
          ))}
        </div>
      )}
    </div>
  )
}
