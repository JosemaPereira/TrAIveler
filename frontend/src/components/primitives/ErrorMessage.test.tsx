import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen } from '@testing-library/react'
import { ErrorMessage } from './ErrorMessage'

describe('ErrorMessage', () => {
  it('renders as an alert with the default title', () => {
    render(<ErrorMessage message="Something went wrong." />)

    const alert = screen.getByRole('alert')
    expect(alert).toBeInTheDocument()
    expect(screen.getByText('Error')).toBeInTheDocument()
  })

  it('renders a custom title when provided', () => {
    render(<ErrorMessage title="Trip failed to load" message="Try again." />)

    expect(screen.getByText('Trip failed to load')).toBeInTheDocument()
  })

  it('renders the message text', () => {
    render(<ErrorMessage message="Network request failed." />)

    expect(screen.getByText('Network request failed.')).toBeInTheDocument()
  })

  it('hides the alert icon from assistive technology', () => {
    render(<ErrorMessage message="Something went wrong." />)

    const alert = screen.getByRole('alert')
    const icon = alert.querySelector('[aria-hidden="true"]')
    expect(icon).not.toBeNull()
  })

  it('does not render a retry button when onRetry is not provided', () => {
    render(<ErrorMessage message="Something went wrong." />)

    expect(
      screen.queryByRole('button', { name: 'Try Again' })
    ).not.toBeInTheDocument()
  })

  it('renders a retry button when onRetry is provided', () => {
    render(<ErrorMessage message="Something went wrong." onRetry={() => {}} />)

    expect(
      screen.getByRole('button', { name: 'Try Again' })
    ).toBeInTheDocument()
  })

  it('calls onRetry when the retry button is clicked', () => {
    const onRetry = vi.fn()
    render(<ErrorMessage message="Something went wrong." onRetry={onRetry} />)

    fireEvent.click(screen.getByRole('button', { name: 'Try Again' }))

    expect(onRetry).toHaveBeenCalledTimes(1)
  })
})
