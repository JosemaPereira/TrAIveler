import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen } from '@testing-library/react'
import { EmptyState } from './EmptyState'

describe('EmptyState', () => {
  it('renders the title', () => {
    render(<EmptyState title="No trips yet" />)

    expect(screen.getByText('No trips yet')).toBeInTheDocument()
  })

  it('renders the message when provided', () => {
    render(<EmptyState title="No trips yet" message="Create your first one." />)

    expect(screen.getByText('Create your first one.')).toBeInTheDocument()
  })

  it('does not render a message paragraph when message is omitted', () => {
    render(<EmptyState title="No trips yet" />)

    expect(screen.queryByText('Create your first one.')).not.toBeInTheDocument()
  })

  it('renders the icon and hides it from assistive technology', () => {
    render(
      <EmptyState title="No trips yet" icon={<svg data-testid="icon" />} />
    )

    const icon = screen.getByTestId('icon')
    expect(icon).toBeInTheDocument()
    expect(icon.parentElement).toHaveAttribute('aria-hidden', 'true')
  })

  it('does not render an action button when action is not provided', () => {
    render(<EmptyState title="No trips yet" />)

    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('renders an action button with the given label when action is provided', () => {
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

  it('calls action.onClick when the action button is clicked', () => {
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
