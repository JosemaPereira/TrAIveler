import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen } from '@testing-library/react'
import { ErrorMessage } from './ErrorMessage'

describe('<ErrorMessage />', () => {
  describe('when rendered with only a message', () => {
    it('should render as an alert with the default title', () => {
      render(<ErrorMessage message="Something went wrong." />)

      expect(screen.getByRole('alert')).toBeInTheDocument()
      expect(screen.getByText('Error')).toBeInTheDocument()
    })

    it('should render the message text', () => {
      render(<ErrorMessage message="Network request failed." />)

      expect(screen.getByText('Network request failed.')).toBeInTheDocument()
    })

    it('should hide the alert icon from assistive technology', () => {
      render(<ErrorMessage message="Something went wrong." />)

      const alert = screen.getByRole('alert')

      expect(alert.querySelector('[aria-hidden="true"]')).not.toBeNull()
    })

    it('should not render a retry button', () => {
      render(<ErrorMessage message="Something went wrong." />)

      expect(
        screen.queryByRole('button', { name: 'Try Again' })
      ).not.toBeInTheDocument()
    })

    it('should not render request id content', () => {
      render(<ErrorMessage message="Something went wrong." />)

      expect(screen.queryByText(/req-/)).not.toBeInTheDocument()
    })
  })

  describe('when a custom title is provided', () => {
    it('should render the custom title', () => {
      render(<ErrorMessage title="Trip failed to load" message="Try again." />)

      expect(screen.getByText('Trip failed to load')).toBeInTheDocument()
    })
  })

  describe('when onRetry is provided', () => {
    it('should render a retry button', () => {
      render(
        <ErrorMessage message="Something went wrong." onRetry={() => {}} />
      )

      expect(
        screen.getByRole('button', { name: 'Try Again' })
      ).toBeInTheDocument()
    })

    it('should call onRetry when the retry button is clicked', () => {
      const onRetry = vi.fn()
      render(<ErrorMessage message="Something went wrong." onRetry={onRetry} />)

      fireEvent.click(screen.getByRole('button', { name: 'Try Again' }))

      expect(onRetry).toHaveBeenCalledTimes(1)
    })
  })

  describe('when a request id is provided', () => {
    it('should render the request id', () => {
      render(
        <ErrorMessage message="Something went wrong." requestId="req-abc-123" />
      )

      expect(screen.getByText(/req-abc-123/)).toBeInTheDocument()
    })

    it('should not render request id content when it is an empty string', () => {
      render(<ErrorMessage message="Something went wrong." requestId="" />)

      expect(screen.queryByTestId('error-request-id')).not.toBeInTheDocument()
    })

    it('should still render and call the retry button when onRetry is also provided', () => {
      const onRetry = vi.fn()
      render(
        <ErrorMessage
          message="Something went wrong."
          requestId="req-xyz-789"
          onRetry={onRetry}
        />
      )

      fireEvent.click(screen.getByRole('button', { name: 'Try Again' }))

      expect(screen.getByText(/req-xyz-789/)).toBeInTheDocument()
      expect(onRetry).toHaveBeenCalledTimes(1)
    })
  })
})
