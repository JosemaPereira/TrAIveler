import { createBrowserRouter } from 'react-router'
import type { RouteObject } from 'react-router'

import { ProtectedRoute } from '@/components/ProtectedRoute'
import { DashboardPage } from './DashboardPage'
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

      // Public routes (reachable without an authenticated session).
      { path: 'login', element: <LoginPage /> },
      { path: 'register', element: <RegisterPage /> },
      { path: 'password-reset', element: <PasswordResetPage /> },

      // Protected routes: ProtectedRoute redirects to /login when the auth
      // store reports no authenticated session.
      {
        element: <ProtectedRoute />,
        children: [
          { path: 'dashboard', element: <DashboardPage /> },
          { path: 'trips/:id', element: <TripDetailPage /> },
          { path: 'settings', element: <SettingsPage /> },
        ],
      },
    ],
  },
]

export const router = createBrowserRouter(routes)
