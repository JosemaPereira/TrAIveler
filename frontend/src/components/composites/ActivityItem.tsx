import { MapPin, Package, Plane, Sparkles, Utensils } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'

import styles from './ActivityItem.module.css'

export type ActivityType = 'visit' | 'food' | 'logistics' | 'transfer'

export interface ActivityData {
  id: string
  title: string
  type: ActivityType
  description?: string
  isAIGenerated: boolean
}

export interface ActivityItemProps extends ActivityData {
  isLoading?: boolean
}

const TYPE_ICONS: Record<ActivityType, LucideIcon> = {
  visit: MapPin,
  food: Utensils,
  logistics: Package,
  transfer: Plane,
}

// Visible text conveying the activity type, so the icon above can stay
// aria-hidden without losing the information for assistive technology.
const TYPE_LABELS: Record<ActivityType, string> = {
  visit: 'Visit',
  food: 'Food',
  logistics: 'Logistics',
  transfer: 'Transfer',
}

/**
 * Decorative skeleton matching ActivityItem's normal footprint, shown while
 * activity data is still loading. Hidden from assistive technology since it
 * carries no information of its own.
 */
function ActivityItemSkeleton() {
  return (
    <div className={styles.activityItem} aria-hidden="true">
      <div className={styles.skeletonIcon} />
      <div className={styles.skeletonBody}>
        <div className={styles.skeletonLine} />
        <div className={styles.skeletonLineShort} />
      </div>
    </div>
  )
}

/**
 * Renders a single trip activity (a visit, meal, logistics step, or
 * transfer). Purely presentational — consumes ActivityData/props only, no
 * data fetching. The type icon is decorative (aria-hidden); the visible
 * TYPE_LABELS text is what actually conveys the type to assistive
 * technology. The AI-generated indicator pairs an icon with a text label so
 * the distinction is never color-only.
 */
export function ActivityItem({
  title,
  type,
  description,
  isAIGenerated,
  isLoading = false,
}: ActivityItemProps) {
  if (isLoading) {
    return <ActivityItemSkeleton />
  }

  const Icon = TYPE_ICONS[type]

  return (
    <div className={styles.activityItem}>
      <Icon aria-hidden="true" className={styles.icon} />
      <div className={styles.body}>
        <div className={styles.header}>
          <h4 className={styles.title}>{title}</h4>
          <span className={styles.typeLabel}>{TYPE_LABELS[type]}</span>
          {isAIGenerated && (
            <span className={styles.aiBadge}>
              <Sparkles aria-hidden="true" className={styles.aiIcon} />
              AI-generated
            </span>
          )}
        </div>
        {description && <p className={styles.description}>{description}</p>}
      </div>
    </div>
  )
}
