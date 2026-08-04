import { describe, expect, it, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'

import { useAuthStore } from '@/stores/auth-store'
import { API_BASE_URL, testTrip, testUser } from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { createQueryWrapper } from '@/test/queryWrapper'
import { DashboardPage } from './DashboardPage'

function renderDashboard() {
  const { wrapper } = createQueryWrapper()
  return render(<DashboardPage />, { wrapper })
}

function mockTrips(trips: (typeof testTrip)[]) {
  server.use(
    http.get(`${API_BASE_URL}/trips`, () =>
      HttpResponse.json({ trips }, { status: 200 })
    )
  )
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
      mockTrips([])

      renderDashboard()

      expect(screen.getByRole('navigation')).toBeInTheDocument()
    })

    it("should render a welcome message with the user's name", () => {
      mockTrips([])

      renderDashboard()

      expect(
        screen.getByRole('heading', {
          level: 1,
          name: new RegExp(`welcome back.*${testUser.full_name}`, 'i'),
        })
      ).toBeInTheDocument()
    })
  })

  describe('while the trips list is loading', () => {
    it('should not render the empty state', () => {
      mockTrips([])

      renderDashboard()

      expect(
        screen.queryByRole('heading', { name: /no trips yet/i })
      ).not.toBeInTheDocument()
    })
  })

  describe('when the trip list resolves empty', () => {
    it('should render the empty state', async () => {
      mockTrips([])

      renderDashboard()

      await waitFor(() => {
        expect(
          screen.getByRole('heading', { name: /no trips yet/i })
        ).toBeInTheDocument()
      })
    })
  })

  describe('when the trip list resolves with trips', () => {
    it('should render a TripCard for each trip and not the empty state', async () => {
      mockTrips([testTrip])

      renderDashboard()

      await waitFor(() => {
        expect(
          screen.getByRole('heading', { name: testTrip.title })
        ).toBeInTheDocument()
      })
      expect(
        screen.queryByRole('heading', { name: /no trips yet/i })
      ).not.toBeInTheDocument()
    })
  })

  describe('when fetching trips fails', () => {
    it('should render an error message and not the empty state', async () => {
      server.use(
        http.get(`${API_BASE_URL}/trips`, () =>
          HttpResponse.json(
            {
              error: 'internal_error',
              message: 'Something went wrong',
              request_id: 'req_trips_500',
            },
            { status: 500 }
          )
        )
      )

      renderDashboard()

      await waitFor(() => {
        expect(screen.getByRole('alert')).toBeInTheDocument()
      })
      expect(
        screen.queryByRole('heading', { name: /no trips yet/i })
      ).not.toBeInTheDocument()
    })
  })

  describe('the New Trip CTA', () => {
    it('should be disabled when the user has no active subscription', () => {
      mockTrips([])
      useAuthStore.setState({
        user: { ...testUser, has_subscription: false },
      })

      renderDashboard()

      expect(screen.getByRole('button', { name: /new trip/i })).toBeDisabled()
    })

    it('should be enabled when the user has an active subscription', () => {
      mockTrips([])
      useAuthStore.setState({
        user: { ...testUser, has_subscription: true },
      })

      renderDashboard()

      expect(
        screen.getByRole('button', { name: /new trip/i })
      ).not.toBeDisabled()
    })
  })
})
