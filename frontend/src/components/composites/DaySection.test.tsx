import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'

import type { ActivityData } from './ActivityItem'
import type { DaySectionProps } from './DaySection'
import { DaySection } from './DaySection'

const activities: ActivityData[] = [
  {
    id: 'activity-1',
    title: 'Visit the Louvre',
    type: 'visit',
    isAIGenerated: false,
  },
  {
    id: 'activity-2',
    title: 'Dinner at Le Comptoir',
    type: 'food',
    isAIGenerated: true,
  },
]

function renderDaySection(overrides: Partial<DaySectionProps> = {}) {
  const props: DaySectionProps = {
    dayNumber: 1,
    activities,
    ...overrides,
  }
  return render(<DaySection {...props} />)
}

describe('<DaySection />', () => {
  describe('when rendered with a day number and no label', () => {
    it('should show the plain day heading', () => {
      renderDaySection({ label: undefined })

      expect(screen.getByRole('heading', { name: 'Day 1' })).toBeInTheDocument()
    })
  })

  describe('having a label', () => {
    it('should show the day heading including the label', () => {
      renderDaySection({ label: 'Arrival in Paris' })

      expect(screen.getByRole('heading', { name: /day 1/i })).toHaveTextContent(
        'Arrival in Paris'
      )
    })
  })

  describe('when there are activities', () => {
    it('should render each activity inside an ordered list item', () => {
      renderDaySection()

      const list = screen.getByRole('list')
      expect(list.tagName).toBe('OL')
      expect(screen.getAllByRole('listitem')).toHaveLength(2)
      expect(screen.getByText('Visit the Louvre')).toBeInTheDocument()
      expect(screen.getByText('Dinner at Le Comptoir')).toBeInTheDocument()
    })
  })

  describe('when activities is empty', () => {
    it('should show an inline empty message instead of a list', () => {
      renderDaySection({ activities: [] })

      expect(
        screen.getByText('No activities planned for this day.')
      ).toBeInTheDocument()
      expect(screen.queryByRole('list')).not.toBeInTheDocument()
    })
  })

  describe('when isLoading is true', () => {
    it('should render a section-shaped skeleton instead of activities', () => {
      renderDaySection({ isLoading: true })

      expect(
        screen.queryByRole('heading', { name: /day 1/i })
      ).not.toBeInTheDocument()
      expect(screen.queryByRole('list')).not.toBeInTheDocument()
      expect(
        screen.queryByText('No activities planned for this day.')
      ).not.toBeInTheDocument()
    })
  })
})
