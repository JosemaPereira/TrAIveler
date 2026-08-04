import { createBrowserRouter } from 'react-router'
import type { RouteObject } from 'react-router'

import { GuestRoute } from '@/components/GuestRoute'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { DashboardPage } from './DashboardPage'
import { GeneratePage } from './GeneratePage'
import { HomePage } from './HomePage'
import { LoginPage } from './LoginPage'
import { PasswordResetPage } from './PasswordResetPage'
import { RegisterPage } from './RegisterPage'
import { RootLayout } from './RootLayout'
import { SettingsPage } from './SettingsPage'
import { TripDetailPage } from './TripDetailPage'

export const routes: RouteObject[] = [
  {
    path: '/',
    element: <RootLayout />,
    children: [
      { index: true, element: <HomePage /> },

      // Public route (reachable regardless of session state).
      { path: 'password-reset', element: <PasswordResetPage /> },

      // Guest-only routes: GuestRoute redirects to /dashboard when the auth
      // store reports an authenticated session.
      {
        element: <GuestRoute />,
        children: [
          { path: 'login', element: <LoginPage /> },
          { path: 'register', element: <RegisterPage /> },
        ],
      },

      // Protected routes: ProtectedRoute redirects to /login when the auth
      // store reports no authenticated session.
      {
        element: <ProtectedRoute />,
        children: [
          { path: 'dashboard', element: <DashboardPage /> },
          { path: 'generate', element: <GeneratePage /> },
          { path: 'trips/:id', element: <TripDetailPage /> },
          { path: 'settings', element: <SettingsPage /> },
        ],
      },
    ],
  },
]

export const router = createBrowserRouter(routes)
