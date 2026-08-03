import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'

import type { TripCardProps } from './TripCard'
import { TripCard } from './TripCard'

function renderTripCard(overrides: Partial<TripCardProps> = {}) {
  const props: TripCardProps = {
    id: 'trip-1',
    title: 'Two Weeks in Japan',
    status: 'draft',
    destinationNames: ['Tokyo', 'Kyoto'],
    durationDays: 14,
    createdAt: '2026-01-15T10:00:00.000Z',
    ...overrides,
  }
  return render(<TripCard {...props} />, { wrapper: MemoryRouter })
}

describe('<TripCard />', () => {
  describe('when rendered with trip data', () => {
    it('should render the trip title', () => {
      renderTripCard()

      expect(
        screen.getByRole('heading', { name: 'Two Weeks in Japan' })
      ).toBeInTheDocument()
    })

    it('should link to the trip detail page', () => {
      renderTripCard({ id: 'trip-42' })

      expect(screen.getByRole('link')).toHaveAttribute(
        'href',
        '/trips/trip-42'
      )
    })

    it('should render the joined destination names', () => {
      renderTripCard({ destinationNames: ['Tokyo', 'Kyoto'] })

      expect(screen.getByText('Tokyo, Kyoto')).toBeInTheDocument()
    })

    it('should render the duration in days', () => {
      renderTripCard({ durationDays: 5 })

      expect(screen.getByText('5 days')).toBeInTheDocument()
    })

    it('should render a singular duration for one day', () => {
      renderTripCard({ durationDays: 1 })

      expect(screen.getByText('1 day')).toBeInTheDocument()
    })

    it('should render a readable created date', () => {
      renderTripCard({ createdAt: '2026-01-15T10:00:00.000Z' })

      expect(screen.getByText(/january 15, 2026/i)).toBeInTheDocument()
    })

    it('should render the draft status', () => {
      renderTripCard({ status: 'draft' })

      expect(screen.getByText('Draft')).toBeInTheDocument()
    })

    it('should render the published status', () => {
      renderTripCard({ status: 'published' })

      expect(screen.getByText('Published')).toBeInTheDocument()
    })
  })

  describe('when destinationNames is empty', () => {
    it('should render a fallback message', () => {
      renderTripCard({ destinationNames: [] })

      expect(screen.getByText('Destinations coming soon')).toBeInTheDocument()
    })
  })

  describe('when isLoading is true', () => {
    it('should render a skeleton instead of the trip content', () => {
      renderTripCard({ isLoading: true })

      expect(
        screen.queryByRole('heading', { name: 'Two Weeks in Japan' })
      ).not.toBeInTheDocument()
      expect(screen.queryByRole('link')).not.toBeInTheDocument()
    })
  })
})
