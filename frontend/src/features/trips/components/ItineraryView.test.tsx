import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import type { DaySectionProps } from '@/components/composites/DaySection'
import type { ItineraryViewProps } from './ItineraryView'
import { ItineraryView } from './ItineraryView'

const days: Omit<DaySectionProps, 'isLoading'>[] = [
  {
    dayNumber: 2,
    activities: [
      {
        id: 'activity-2',
        title: 'Fushimi Inari Shrine',
        type: 'visit',
        isAIGenerated: false,
      },
    ],
  },
  {
    dayNumber: 1,
    label: 'Arrival',
    activities: [
      {
        id: 'activity-1',
        title: 'Airport transfer',
        type: 'transfer',
        isAIGenerated: false,
      },
    ],
  },
]

function renderItineraryView(overrides: Partial<ItineraryViewProps> = {}) {
  const props: ItineraryViewProps = {
    days,
    ...overrides,
  }
  return render(<ItineraryView {...props} />)
}

describe('<ItineraryView />', () => {
  describe('when rendered with days', () => {
    it('should render day sections sorted by day number', () => {
      renderItineraryView()

      const dayHeadings = screen.getAllByRole('heading', { level: 2 })
      expect(dayHeadings[0]).toHaveTextContent(/day 1/i)
      expect(dayHeadings[1]).toHaveTextContent(/day 2/i)
    })

    it('should render the activities for each day', () => {
      renderItineraryView()

      expect(screen.getByText('Airport transfer')).toBeInTheDocument()
      expect(screen.getByText('Fushimi Inari Shrine')).toBeInTheDocument()
    })
  })

  describe('when isLoading is true', () => {
    it('should render skeleton day sections instead of real content', () => {
      renderItineraryView({ isLoading: true })

      expect(screen.queryByText('Airport transfer')).not.toBeInTheDocument()
      expect(
        screen.queryByRole('heading', { name: /day 1/i })
      ).not.toBeInTheDocument()
    })

    it('should take priority over an error', () => {
      renderItineraryView({ isLoading: true, error: 'Something failed' })

      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    })
  })

  describe('when error is set', () => {
    it('should render the error message', () => {
      renderItineraryView({ error: 'Failed to load itinerary' })

      expect(screen.getByRole('alert')).toHaveTextContent(
        'Failed to load itinerary'
      )
    })

    it('should call onRetry when the retry action is used', async () => {
      const user = userEvent.setup()
      const onRetry = vi.fn()
      renderItineraryView({ error: 'Failed to load itinerary', onRetry })

      await user.click(screen.getByRole('button', { name: /try again/i }))

      expect(onRetry).toHaveBeenCalledTimes(1)
    })

    it('should take priority over an empty days array', () => {
      renderItineraryView({ error: 'Failed to load itinerary', days: [] })

      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
  })

  describe('when days is empty and there is no error', () => {
    it('should render an empty-state message', () => {
      renderItineraryView({ days: [] })

      expect(screen.getByText(/itinerary not yet generated/i)).toBeInTheDocument()
    })
  })
})
