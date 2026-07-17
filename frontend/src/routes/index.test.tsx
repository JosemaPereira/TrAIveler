import { beforeEach, describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'

import { useAuthStore } from '../stores/auth-store'
import { routes } from './index'

function renderAt(path: string) {
  const memoryRouter = createMemoryRouter(routes, { initialEntries: [path] })

  render(<RouterProvider router={memoryRouter} />)
}

// Zustand store is a module-singleton: reset before each test so the
// protected-route assertions don't leak an authenticated session.
beforeEach(() => {
  useAuthStore.setState({
    isAuthenticated: false,
    user: null,
    isLoading: false,
  })
})

describe('routes', () => {
  describe('when navigating to the root route', () => {
    it('should render the TrAIveler placeholder heading', () => {
      renderAt('/')

      expect(
        screen.getByRole('heading', { name: 'TrAIveler' })
      ).toBeInTheDocument()
    })
  })

  describe('when navigating to a public route', () => {
    it('should render the login page', () => {
      renderAt('/login')

      expect(
        screen.getByRole('heading', { name: 'Log In' })
      ).toBeInTheDocument()
    })

    it('should render the register page', () => {
      renderAt('/register')

      expect(
        screen.getByRole('heading', { name: 'Create Account' })
      ).toBeInTheDocument()
    })

    it('should render the password-reset page', () => {
      renderAt('/password-reset')

      expect(
        screen.getByRole('heading', { name: 'Reset Password' })
      ).toBeInTheDocument()
    })
  })

  describe('when navigating to a protected route while unauthenticated', () => {
    it('should redirect /dashboard to the login page', () => {
      renderAt('/dashboard')

      expect(
        screen.getByRole('heading', { name: 'Log In' })
      ).toBeInTheDocument()
      expect(
        screen.queryByRole('heading', { name: 'Dashboard' })
      ).not.toBeInTheDocument()
    })
  })

  describe('when navigating to a protected route while authenticated', () => {
    it('should render the dashboard page', () => {
      useAuthStore.setState({ isAuthenticated: true })

      renderAt('/dashboard')

      expect(
        screen.getByRole('heading', { name: 'Dashboard' })
      ).toBeInTheDocument()
    })

    it('should render the trip-detail page for /trips/:id', () => {
      useAuthStore.setState({ isAuthenticated: true })

      renderAt('/trips/trip_1')

      expect(screen.getByRole('heading', { name: 'Trip' })).toBeInTheDocument()
    })

    it('should render the settings page', () => {
      useAuthStore.setState({ isAuthenticated: true })

      renderAt('/settings')

      expect(
        screen.getByRole('heading', { name: 'Settings' })
      ).toBeInTheDocument()
    })
  })
})
