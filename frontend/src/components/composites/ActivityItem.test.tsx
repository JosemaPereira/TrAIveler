import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'

import type { ActivityData } from './ActivityItem'
import { ActivityItem } from './ActivityItem'

const baseActivity: ActivityData = {
  id: 'activity-1',
  title: 'Visit the Eiffel Tower',
  type: 'visit',
  description: 'Iconic Parisian landmark with panoramic city views.',
  isAIGenerated: false,
}

describe('<ActivityItem />', () => {
  describe('when rendered with a visit activity', () => {
    it('should render the activity title', () => {
      render(<ActivityItem {...baseActivity} />)

      expect(
        screen.getByRole('heading', { name: 'Visit the Eiffel Tower' })
      ).toBeInTheDocument()
    })

    it('should render the description', () => {
      render(<ActivityItem {...baseActivity} />)

      expect(
        screen.getByText('Iconic Parisian landmark with panoramic city views.')
      ).toBeInTheDocument()
    })

    it('should render a visible type label conveying the activity type as text', () => {
      render(<ActivityItem {...baseActivity} />)

      expect(screen.getByText('Visit')).toBeInTheDocument()
    })

    it('should not show the AI-generated indicator', () => {
      render(<ActivityItem {...baseActivity} />)

      expect(screen.queryByText(/ai-generated/i)).not.toBeInTheDocument()
    })
  })

  describe('having no description', () => {
    it('should render without throwing and omit the description text', () => {
      const withoutDescription: ActivityData = {
        id: baseActivity.id,
        title: baseActivity.title,
        type: baseActivity.type,
        isAIGenerated: baseActivity.isAIGenerated,
      }

      render(<ActivityItem {...withoutDescription} />)

      expect(
        screen.getByRole('heading', { name: 'Visit the Eiffel Tower' })
      ).toBeInTheDocument()
    })
  })

  describe.each([
    ['food', 'Food'],
    ['logistics', 'Logistics'],
    ['transfer', 'Transfer'],
  ] as const)('when the activity type is %s', (type, label) => {
    it(`should render the ${label} type label`, () => {
      render(<ActivityItem {...baseActivity} type={type} />)

      expect(screen.getByText(label)).toBeInTheDocument()
    })
  })

  describe('when the activity is AI-generated', () => {
    it('should show a non-color-only AI-generated indicator', () => {
      render(<ActivityItem {...baseActivity} isAIGenerated />)

      expect(screen.getByText(/ai-generated/i)).toBeInTheDocument()
    })
  })

  describe('when isLoading is true', () => {
    it('should render a decorative skeleton instead of the activity content', () => {
      render(<ActivityItem {...baseActivity} isLoading />)

      expect(
        screen.queryByRole('heading', { name: 'Visit the Eiffel Tower' })
      ).not.toBeInTheDocument()
    })
  })
})
