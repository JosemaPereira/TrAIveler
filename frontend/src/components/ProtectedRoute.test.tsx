import { render, screen, waitFor } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { beforeEach, describe, expect, it } from 'vitest'

import { useAuthStore } from '@/stores/auth-store'
import { ProtectedRoute } from './ProtectedRoute'

// Minimal route tree: a public /login target and a protected /dashboard behind
// ProtectedRoute, so the redirect can be observed by what renders.
const routes = [
  { path: '/login', element: <h1>Log In</h1> },
  {
    element: <ProtectedRoute />,
    children: [{ path: '/dashboard', element: <h1>Dashboard</h1> }],
  },
]

function renderAtDashboard() {
  const router = createMemoryRouter(routes, {
    initialEntries: ['/dashboard'],
  })

  render(<RouterProvider router={router} />)

  return router
}

// Zustand store is a module-singleton: reset before each test.
beforeEach(() => {
  useAuthStore.setState({
    isAuthenticated: false,
    user: null,
    isLoading: false,
  })
})

describe('<ProtectedRoute />', () => {
  describe('when the user is not authenticated', () => {
    it('should redirect to the login route instead of rendering the child', () => {
      renderAtDashboard()

      expect(
        screen.getByRole('heading', { name: 'Log In' })
      ).toBeInTheDocument()
      expect(
        screen.queryByRole('heading', { name: 'Dashboard' })
      ).not.toBeInTheDocument()
    })

    it('should carry the attempted path as a redirect query param', async () => {
      const router = renderAtDashboard()

      await waitFor(() => {
        expect(router.state.location.pathname).toBe('/login')
      })
      expect(router.state.location.search).toBe('?redirect=%2Fdashboard')
    })
  })

  describe('when the user is authenticated', () => {
    it('should render the protected child route', () => {
      useAuthStore.setState({ isAuthenticated: true })

      renderAtDashboard()

      expect(
        screen.getByRole('heading', { name: 'Dashboard' })
      ).toBeInTheDocument()
      expect(
        screen.queryByRole('heading', { name: 'Log In' })
      ).not.toBeInTheDocument()
    })
  })
})
