import { describe, expect, it, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'

import { useAuthStore } from '@/stores/auth-store'
import { testUser } from '@/test/msw/handlers'
import { createQueryWrapper } from '@/test/queryWrapper'
import { DashboardPage } from './DashboardPage'

function renderDashboard() {
  const { wrapper } = createQueryWrapper()
  return render(<DashboardPage />, { wrapper })
}

beforeEach(() => {
  useAuthStore.setState({
    isAuthenticated: true,
    user: testUser,
    isLoading: false,
  })
})

describe('<DashboardPage />', () => {
  describe('when rendered for an authenticated user', () => {
    it('should render the Navigation', () => {
      renderDashboard()

      expect(screen.getByRole('navigation')).toBeInTheDocument()
    })

    it("should render a welcome message with the user's name", () => {
      renderDashboard()

      expect(
        screen.getByRole('heading', {
          level: 1,
          name: new RegExp(`welcome back.*${testUser.full_name}`, 'i'),
        })
      ).toBeInTheDocument()
    })

    it('should render the empty state for trips', () => {
      renderDashboard()

      expect(
        screen.getByRole('heading', { name: /no trips yet/i })
      ).toBeInTheDocument()
    })
  })
})
