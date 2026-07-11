import { Outlet } from 'react-router'

/**
 * Root layout shared by every route. Future sprints will add persistent
 * chrome here (header/navigation, footer); for now it only provides the
 * <Outlet /> mount point for nested routes.
 */
export function RootLayout() {
  return (
    <div>
      <Outlet />
    </div>
  )
}
