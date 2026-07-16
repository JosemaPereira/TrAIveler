import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen } from '@testing-library/react'
import { EmptyState } from './EmptyState'

describe('<EmptyState />', () => {
  describe('when rendered with only a title', () => {
    it('should render the title', () => {
      render(<EmptyState title="No trips yet" />)

      expect(screen.getByText('No trips yet')).toBeInTheDocument()
    })

    it('should not render a message paragraph', () => {
      render(<EmptyState title="No trips yet" />)

      expect(
        screen.queryByText('Create your first one.')
      ).not.toBeInTheDocument()
    })

    it('should not render an action button', () => {
      render(<EmptyState title="No trips yet" />)

      expect(screen.queryByRole('button')).not.toBeInTheDocument()
    })
  })

  describe('when a message is provided', () => {
    it('should render the message', () => {
      render(
        <EmptyState title="No trips yet" message="Create your first one." />
      )

      expect(screen.getByText('Create your first one.')).toBeInTheDocument()
    })
  })

  describe('when an icon is provided', () => {
    it('should render the icon hidden from assistive technology', () => {
      render(
        <EmptyState title="No trips yet" icon={<svg data-testid="icon" />} />
      )

      const icon = screen.getByTestId('icon')

      expect(icon).toBeInTheDocument()
      expect(icon.parentElement).toHaveAttribute('aria-hidden', 'true')
    })
  })

  describe('when an action is provided', () => {
    it('should render an action button with the given label', () => {
      render(
        <EmptyState
          title="No trips yet"
          action={{ label: 'Create a trip', onClick: () => {} }}
        />
      )

      expect(
        screen.getByRole('button', { name: 'Create a trip' })
      ).toBeInTheDocument()
    })

    it('should call action.onClick when the action button is clicked', () => {
      const onClick = vi.fn()
      render(
        <EmptyState
          title="No trips yet"
          action={{ label: 'Create a trip', onClick }}
        />
      )

      fireEvent.click(screen.getByRole('button', { name: 'Create a trip' }))

      expect(onClick).toHaveBeenCalledTimes(1)
    })
  })
})
