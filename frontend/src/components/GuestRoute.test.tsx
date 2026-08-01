import { render, screen, waitFor } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { beforeEach, describe, expect, it } from 'vitest'

import { useAuthStore } from '@/stores/auth-store'
import { GuestRoute } from './GuestRoute'

// Minimal route tree: a protected /dashboard target and a guest-only /login
// behind GuestRoute, so the redirect can be observed by what renders.
const routes = [
  { path: '/dashboard', element: <h1>Dashboard</h1> },
  {
    element: <GuestRoute />,
    children: [{ path: '/login', element: <h1>Log In</h1> }],
  },
]

function renderAtLogin() {
  const router = createMemoryRouter(routes, {
    initialEntries: ['/login'],
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

describe('<GuestRoute />', () => {
  describe('when the user is authenticated', () => {
    it('should redirect to the dashboard instead of rendering the guest child', () => {
      useAuthStore.setState({ isAuthenticated: true })

      renderAtLogin()

      expect(
        screen.getByRole('heading', { name: 'Dashboard' })
      ).toBeInTheDocument()
      expect(
        screen.queryByRole('heading', { name: 'Log In' })
      ).not.toBeInTheDocument()
    })

    it('should navigate to the dashboard route', async () => {
      useAuthStore.setState({ isAuthenticated: true })

      const router = renderAtLogin()

      await waitFor(() => {
        expect(router.state.location.pathname).toBe('/dashboard')
      })
    })
  })

  describe('when the user is not authenticated', () => {
    it('should render the guest-only child route', () => {
      renderAtLogin()

      expect(
        screen.getByRole('heading', { name: 'Log In' })
      ).toBeInTheDocument()
      expect(
        screen.queryByRole('heading', { name: 'Dashboard' })
      ).not.toBeInTheDocument()
    })
  })
})
