import type { DaySectionProps } from '@/components/composites/DaySection'
import { DaySection } from '@/components/composites/DaySection'
import { EmptyState } from '@/components/primitives/EmptyState'
import { ErrorMessage } from '@/components/primitives/ErrorMessage'
import styles from './ItineraryView.module.css'

export interface ItineraryViewProps {
  days: Omit<DaySectionProps, 'isLoading'>[]
  isLoading?: boolean
  error?: string
  onRetry?: () => void
}

const SKELETON_DAY_COUNT = 3
const EMPTY_STATE_TITLE = 'Itinerary not yet generated'
const EMPTY_STATE_MESSAGE =
  'Start a conversation to build your day-by-day plan.'

/**
 * Day-by-day itinerary view. Purely presentational — composes DaySection
 * from props only, no data fetching of its own. State priority when more
 * than one could apply: loading skeleton, then error, then the "not yet
 * generated" empty state, then the real day list (sorted by day number).
 */
export function ItineraryView({
  days,
  isLoading = false,
  error,
  onRetry,
}: ItineraryViewProps) {
  if (isLoading) {
    return (
      <div className={styles.itineraryView}>
        {Array.from({ length: SKELETON_DAY_COUNT }, (_, index) => (
          <DaySection
            key={index}
            dayNumber={index + 1}
            activities={[]}
            isLoading
          />
        ))}
      </div>
    )
  }

  if (error) {
    return <ErrorMessage message={error} onRetry={onRetry} />
  }

  if (days.length === 0) {
    return (
      <EmptyState title={EMPTY_STATE_TITLE} message={EMPTY_STATE_MESSAGE} />
    )
  }

  const sortedDays = [...days].sort((a, b) => a.dayNumber - b.dayNumber)

  return (
    <div className={styles.itineraryView}>
      {sortedDays.map((day) => (
        <DaySection key={day.dayNumber} {...day} />
      ))}
    </div>
  )
}
