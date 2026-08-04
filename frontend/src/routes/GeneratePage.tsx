import { useState } from 'react'
import type { ChangeEvent, SyntheticEvent } from 'react'

import { Navigation } from '@/components/composites/Navigation'
import { Button } from '@/components/primitives/Button'
import { Input } from '@/components/primitives/Input'
import { ConversationPanel } from '@/features/trips/components/ConversationPanel'
import { ItineraryView } from '@/features/trips/components/ItineraryView'
import { useConversation } from '@/features/trips/hooks/useConversation'
import { useCreateTrip } from '@/features/trips/hooks/useCreateTrip'
import { mapApiError } from '@/lib/error-handler'
import styles from './GeneratePage.module.css'

const TITLE_LABEL = 'Trip title'
const DESCRIPTION_LABEL = 'Description'
const CREATE_LABEL = 'Start Planning'

/**
 * Trip-generation page (`/generate`): a small "new trip" form hands off to
 * the AI-guided planning conversation once a trip exists, and finally to the
 * itinerary view once the conversation reports `itinerary_ready`.
 *
 * State is local page state (`useCreateTrip` for the form, the
 * `useConversation` hook for the chat), not a single combined hook — the two
 * halves of this flow have independent lifecycles (a trip is created once;
 * the conversation runs many turns after), so splitting them keeps each
 * piece simple rather than forcing one hook to model both.
 */
export function GeneratePage() {
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [tripId, setTripId] = useState<string>()
  const createTrip = useCreateTrip()
  const conversation = useConversation(tripId ?? '')

  const createErrorInfo = createTrip.isError
    ? mapApiError(createTrip.error)
    : null

  function handleCreateTrip(event: SyntheticEvent<HTMLFormElement>) {
    event.preventDefault()
    createTrip.mutate(
      { title, description: description || undefined },
      {
        onSuccess: (trip) => {
          setTripId(trip.id)
        },
      }
    )
  }

  return (
    <div>
      <Navigation />
      <h1>Plan a New Trip</h1>

      {!tripId && (
        <form onSubmit={handleCreateTrip} className={styles.form}>
          <Input
            label={TITLE_LABEL}
            value={title}
            onChange={(event: ChangeEvent<HTMLInputElement>) => {
              setTitle(event.target.value)
            }}
            required
          />
          <Input
            label={DESCRIPTION_LABEL}
            value={description}
            onChange={(event: ChangeEvent<HTMLInputElement>) => {
              setDescription(event.target.value)
            }}
          />
          {createErrorInfo && (
            <p role="alert" className={styles.error}>
              {createErrorInfo.message}
            </p>
          )}
          <Button type="submit" disabled={createTrip.isPending}>
            {CREATE_LABEL}
          </Button>
        </form>
      )}

      {tripId && !conversation.itineraryReady && (
        <ConversationPanel
          messages={conversation.messages}
          onSendMessage={conversation.sendMessage}
          isSending={conversation.isSending}
          error={conversation.error}
        />
      )}

      {tripId && conversation.itineraryReady && (
        // No itinerary-content endpoint exists yet — same documented gap as
        // TripDetailPage.tsx / useTripDetail.ts. ItineraryView's own "not yet
        // generated" empty state is the correct representation here, not
        // fabricated day/activity data.
        <ItineraryView days={[]} />
      )}
    </div>
  )
}
