import { describe, expect, it, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { useAuthStore } from '@/stores/auth-store'
import { testUser } from '@/test/msw/handlers'
import { createQueryWrapper } from '@/test/queryWrapper'
import { Navigation } from './Navigation'

function renderNavigation() {
  const { wrapper } = createQueryWrapper()
  return render(<Navigation />, { wrapper })
}

beforeEach(() => {
  useAuthStore.setState({
    isAuthenticated: true,
    user: testUser,
    isLoading: false,
  })
})

describe('<Navigation />', () => {
  describe('when rendered for an authenticated user', () => {
    it("should display the user's full name", () => {
      renderNavigation()

      expect(screen.getByText(testUser.full_name)).toBeInTheDocument()
    })

    it('should render a nav landmark', () => {
      renderNavigation()

      expect(screen.getByRole('navigation')).toBeInTheDocument()
    })
  })

  describe('having a user with an active subscription', () => {
    it('should show the subscription badge', () => {
      useAuthStore.setState({
        user: { ...testUser, has_subscription: true },
      })

      renderNavigation()

      expect(screen.getByText(/subscri/i)).toBeInTheDocument()
    })
  })

  describe('having a user without an active subscription', () => {
    it('should not show the subscription badge', () => {
      useAuthStore.setState({
        user: { ...testUser, has_subscription: false },
      })

      renderNavigation()

      expect(screen.queryByText(/subscri/i)).not.toBeInTheDocument()
    })
  })

  describe('when the Log Out button is clicked', () => {
    it('should trigger the logout mutation and clear the session', async () => {
      const user = userEvent.setup()
      renderNavigation()

      await user.click(screen.getByRole('button', { name: /log out/i }))

      await waitFor(() => {
        expect(useAuthStore.getState().isAuthenticated).toBe(false)
      })
      expect(useAuthStore.getState().user).toBeNull()
    })
  })

  describe('having the default (unopened) mobile menu state', () => {
    it('should render the mobile menu toggle as closed', () => {
      renderNavigation()

      expect(
        screen.getByRole('button', { name: /menu/i })
      ).toHaveAttribute('aria-expanded', 'false')
    })
  })

  describe('when the mobile menu toggle is clicked', () => {
    it('should open the menu', async () => {
      const user = userEvent.setup()
      renderNavigation()

      const toggle = screen.getByRole('button', { name: /menu/i })
      await user.click(toggle)

      expect(toggle).toHaveAttribute('aria-expanded', 'true')
    })

    it('should close the menu when clicked again', async () => {
      const user = userEvent.setup()
      renderNavigation()

      const toggle = screen.getByRole('button', { name: /menu/i })
      await user.click(toggle)
      await user.click(toggle)

      expect(toggle).toHaveAttribute('aria-expanded', 'false')
    })
  })
})
