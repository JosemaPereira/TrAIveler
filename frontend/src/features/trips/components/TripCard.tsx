import { Link } from 'react-router'

import { Card } from '@/components/primitives/Card'
import styles from './TripCard.module.css'

export interface TripCardProps {
  id: string
  title: string
  status: 'draft' | 'published'
  destinationNames: string[]
  durationDays: number
  createdAt: string
  isLoading?: boolean
}

const NO_DESTINATIONS_FALLBACK = 'Destinations coming soon'

// Trip status is a distinct domain from the Badge primitive's subscription-tier
// variant ('free' | 'paid' | 'pending') — see Badge.tsx's doc comment — so status
// is rendered as local, token-driven markup instead of forcing it through Badge.
const STATUS_LABELS: Record<TripCardProps['status'], string> = {
  draft: 'Draft',
  published: 'Published',
}

function formatDuration(days: number): string {
  return `${String(days)} ${days === 1 ? 'day' : 'days'}`
}

function formatCreatedAt(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

/**
 * Decorative skeleton matching TripCard's normal footprint, shown while trip
 * data is still loading. Not wrapped in a Link — there is nothing to
 * navigate to yet — and hidden from assistive technology.
 */
function TripCardSkeleton() {
  return (
    <Card className={styles.card}>
      <div aria-hidden="true" className={styles.skeleton}>
        <div className={styles.skeletonTitle} />
        <div className={styles.skeletonLine} />
        <div className={styles.skeletonLine} />
      </div>
    </Card>
  )
}

/**
 * Summary card for a trip, linking to its detail page. Purely
 * presentational — durationDays is passed in directly rather than derived,
 * since neither the trip-list nor trip-detail API contracts expose a single
 * ready-made "duration" field (see this issue's own notes).
 */
export function TripCard({
  id,
  title,
  status,
  destinationNames,
  durationDays,
  createdAt,
  isLoading = false,
}: TripCardProps) {
  if (isLoading) {
    return <TripCardSkeleton />
  }

  const destinationsText =
    destinationNames.length > 0
      ? destinationNames.join(', ')
      : NO_DESTINATIONS_FALLBACK

  return (
    <Link to={`/trips/${id}`} className={styles.link}>
      <Card className={styles.card}>
        <div className={styles.header}>
          <h3 className={styles.title}>{title}</h3>
          <span className={[styles.statusBadge, styles[status]].join(' ')}>
            {STATUS_LABELS[status]}
          </span>
        </div>
        <p className={styles.destinations}>{destinationsText}</p>
        <div className={styles.meta}>
          <span>{formatDuration(durationDays)}</span>
          <span>{formatCreatedAt(createdAt)}</span>
        </div>
      </Card>
    </Link>
  )
}
