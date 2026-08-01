import { Navigation } from '@/components/composites/Navigation'
import { EmptyState } from '@/components/primitives/EmptyState'
import { useUser } from '@/stores/auth-store'

/**
 * EmptyState below is a placeholder; the real trip list lands with
 * `TripDashboard`/`useTrips` in Sprint 8.
 */
export function DashboardPage() {
  const user = useUser()

  return (
    <div>
      <Navigation />
      <h1>Welcome back{user ? `, ${user.full_name}` : ''}</h1>
      <EmptyState title="No trips yet" />
    </div>
  )
}
