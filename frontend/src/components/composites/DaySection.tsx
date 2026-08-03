import { useId } from 'react'

import type { ActivityData } from './ActivityItem'
import { ActivityItem } from './ActivityItem'
import styles from './DaySection.module.css'

export interface DaySectionProps {
  dayNumber: number
  label?: string
  activities: ActivityData[]
  isLoading?: boolean
}

const EMPTY_DAY_MESSAGE = 'No activities planned for this day.'
const SKELETON_ACTIVITY_COUNT = 2

/**
 * Section-shaped skeleton shown while a day's activities are still loading.
 * Hidden from assistive technology since it carries no information of its
 * own — the real heading/list replace it once data arrives.
 */
function DaySectionSkeleton() {
  return (
    <section className={styles.daySection} aria-hidden="true">
      <div className={styles.skeletonHeading} />
      <div className={styles.skeletonList}>
        {Array.from({ length: SKELETON_ACTIVITY_COUNT }, (_, index) => (
          <div key={index} className={styles.skeletonActivity} />
        ))}
      </div>
    </section>
  )
}

/**
 * Renders one day of an itinerary: a "Day N" heading (plus an optional
 * label) and its activities, in sequence order, as an ordered list of
 * ActivityItem. Purely presentational — composes ActivityItem from
 * props only, no data fetching of its own.
 */
export function DaySection({
  dayNumber,
  label,
  activities,
  isLoading = false,
}: DaySectionProps) {
  const headingId = useId()

  if (isLoading) {
    return <DaySectionSkeleton />
  }

  const dayLabel = String(dayNumber)
  const heading = label ? `Day ${dayLabel} — ${label}` : `Day ${dayLabel}`

  return (
    <section className={styles.daySection} aria-labelledby={headingId}>
      <h2 id={headingId} className={styles.heading}>
        {heading}
      </h2>
      {activities.length === 0 ? (
        <p className={styles.emptyMessage}>{EMPTY_DAY_MESSAGE}</p>
      ) : (
        <ol className={styles.activityList}>
          {activities.map((activity) => (
            <li key={activity.id} className={styles.activityListItem}>
              <ActivityItem {...activity} />
            </li>
          ))}
        </ol>
      )}
    </section>
  )
}
