import { beforeEach, describe, expect, it } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createMemoryRouter, RouterProvider } from 'react-router'

import { useAuthStore } from '@/stores/auth-store'
import { testUser } from '@/test/msw/handlers'
import { routes } from './index'

// RegisterPage/LoginPage use TanStack Query (via useRegister/useLogin), so
// this needs a QueryClientProvider ancestor just like the real composition
// root does (see App.tsx, which wraps RouterProvider in one). A fresh client
// per render keeps mutation state from leaking between tests.
function renderAt(path: string) {
  const memoryRouter = createMemoryRouter(routes, { initialEntries: [path] })
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={memoryRouter} />
    </QueryClientProvider>
  )
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
        screen.getByRole('heading', { name: 'Welcome Back' })
      ).toBeInTheDocument()
    })

    it('should render the register page', () => {
      renderAt('/register')

      expect(
        screen.getByRole('heading', { name: 'Create Your Account' })
      ).toBeInTheDocument()
    })

    it('should render the password-reset page', () => {
      renderAt('/password-reset')

      expect(
        screen.getByRole('heading', { name: 'Reset Password' })
      ).toBeInTheDocument()
    })
  })

  describe('when navigating to a guest-only route while authenticated', () => {
    it('should redirect /login to the dashboard', () => {
      useAuthStore.setState({ isAuthenticated: true })

      renderAt('/login')

      expect(
        screen.getByRole('heading', { name: /welcome back/i })
      ).toBeInTheDocument()
      expect(
        screen.queryByRole('heading', { name: 'Welcome Back' })
      ).not.toBeInTheDocument()
    })

    it('should redirect /register to the dashboard', () => {
      useAuthStore.setState({ isAuthenticated: true })

      renderAt('/register')

      expect(
        screen.getByRole('heading', { name: /welcome back/i })
      ).toBeInTheDocument()
      expect(
        screen.queryByRole('heading', { name: 'Create Your Account' })
      ).not.toBeInTheDocument()
    })
  })

  describe('when navigating to a protected route while unauthenticated', () => {
    it('should redirect /dashboard to the login page', () => {
      renderAt('/dashboard')

      expect(
        screen.getByRole('heading', { name: 'Welcome Back' })
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
        screen.getByRole('heading', { name: /welcome back/i })
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

  describe('when linking between the register and login pages', () => {
    it('should follow the login link from the register page', async () => {
      const user = userEvent.setup()
      renderAt('/register')

      await user.click(screen.getByRole('link', { name: /log in/i }))

      expect(
        screen.getByRole('heading', { name: 'Welcome Back' })
      ).toBeInTheDocument()
    })

    it('should follow the register link from the login page', async () => {
      const user = userEvent.setup()
      renderAt('/login')

      await user.click(screen.getByRole('link', { name: /sign up|register/i }))

      expect(
        screen.getByRole('heading', { name: 'Create Your Account' })
      ).toBeInTheDocument()
    })
  })

  describe('when registering successfully from the register page', () => {
    it('should navigate to the dashboard', async () => {
      const user = userEvent.setup()
      renderAt('/register')

      await user.type(
        screen.getByRole('textbox', { name: 'Email' }),
        'new-traveler@example.com'
      )
      await user.type(screen.getByLabelText('Password'), 'CorrectHorse1!')
      await user.type(
        screen.getByRole('textbox', { name: 'Full Name' }),
        'New Traveler'
      )
      await user.click(
        screen.getByRole('button', { name: 'Create Free Account' })
      )

      await waitFor(() => {
        expect(
          screen.getByRole('heading', { name: /welcome back/i })
        ).toBeInTheDocument()
      })
    })
  })

  describe('when logging in successfully from the login page', () => {
    it('should navigate to the dashboard when no redirect param is present', async () => {
      const user = userEvent.setup()
      renderAt('/login')

      await user.type(
        screen.getByRole('textbox', { name: 'Email' }),
        testUser.email
      )
      await user.type(screen.getByLabelText('Password'), 'CorrectHorse1!')
      await user.click(screen.getByRole('button', { name: 'Log In' }))

      await waitFor(() => {
        expect(
          screen.getByRole('heading', { name: /welcome back/i })
        ).toBeInTheDocument()
      })
    })

    it('should navigate to the sanitized redirect target when present', async () => {
      const user = userEvent.setup()
      renderAt('/login?redirect=%2Fsettings')

      await user.type(
        screen.getByRole('textbox', { name: 'Email' }),
        testUser.email
      )
      await user.type(screen.getByLabelText('Password'), 'CorrectHorse1!')
      await user.click(screen.getByRole('button', { name: 'Log In' }))

      await waitFor(() => {
        expect(
          screen.getByRole('heading', { name: 'Settings' })
        ).toBeInTheDocument()
      })
    })

    it('should ignore an off-site redirect target and fall back to the dashboard', async () => {
      const user = userEvent.setup()
      renderAt('/login?redirect=%2F%2Fevil.example.com')

      await user.type(
        screen.getByRole('textbox', { name: 'Email' }),
        testUser.email
      )
      await user.type(screen.getByLabelText('Password'), 'CorrectHorse1!')
      await user.click(screen.getByRole('button', { name: 'Log In' }))

      await waitFor(() => {
        expect(
          screen.getByRole('heading', { name: /welcome back/i })
        ).toBeInTheDocument()
      })
    })
  })
})
