import { afterEach, describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { ErrorBoundary } from './ErrorBoundary'

const ERROR_MESSAGE = 'Test error'

function ThrowError(): never {
  throw new Error(ERROR_MESSAGE)
}

afterEach(() => {
  vi.unstubAllEnvs()
  vi.unstubAllGlobals()
  vi.restoreAllMocks()
})

describe('<ErrorBoundary />', () => {
  describe('when no error is thrown', () => {
    it('should render children normally', () => {
      render(
        <ErrorBoundary>
          <p>Safe content</p>
        </ErrorBoundary>
      )

      expect(screen.getByText('Safe content')).toBeInTheDocument()
    })
  })

  describe('when a child throws during render', () => {
    it('should render the fallback UI instead of children', () => {
      vi.spyOn(console, 'error').mockImplementation(() => {})

      render(
        <ErrorBoundary>
          <ThrowError />
        </ErrorBoundary>
      )

      expect(screen.queryByText('Safe content')).not.toBeInTheDocument()
      expect(
        screen.getByRole('heading', { name: 'Something went wrong' })
      ).toBeInTheDocument()
    })

    it('should show a "Go Home" actionable element in the fallback UI', () => {
      vi.spyOn(console, 'error').mockImplementation(() => {})

      render(
        <ErrorBoundary>
          <ThrowError />
        </ErrorBoundary>
      )

      expect(
        screen.getByRole('button', { name: 'Go Home' })
      ).toBeInTheDocument()
    })

    it('should navigate to / when "Go Home" is clicked', async () => {
      const user = userEvent.setup()
      const assign = vi.fn()
      vi.stubGlobal('location', { assign })
      vi.spyOn(console, 'error').mockImplementation(() => {})
      render(
        <ErrorBoundary>
          <ThrowError />
        </ErrorBoundary>
      )

      await user.click(screen.getByRole('button', { name: 'Go Home' }))

      expect(assign).toHaveBeenCalledWith('/')
    })

    it('should log the caught error via console.error', () => {
      const consoleErrorSpy = vi
        .spyOn(console, 'error')
        .mockImplementation(() => {})

      render(
        <ErrorBoundary>
          <ThrowError />
        </ErrorBoundary>
      )

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        'ErrorBoundary caught an error:',
        expect.any(Error),
        expect.anything()
      )
    })
  })

  describe('having a development build', () => {
    it('should show the raw error message', () => {
      vi.stubEnv('DEV', true)
      vi.spyOn(console, 'error').mockImplementation(() => {})

      render(
        <ErrorBoundary>
          <ThrowError />
        </ErrorBoundary>
      )

      expect(screen.getByText(ERROR_MESSAGE)).toBeInTheDocument()
    })
  })

  describe('having a non-development build', () => {
    it('should hide the raw error message', () => {
      vi.stubEnv('DEV', false)
      vi.spyOn(console, 'error').mockImplementation(() => {})

      render(
        <ErrorBoundary>
          <ThrowError />
        </ErrorBoundary>
      )

      expect(screen.queryByText(ERROR_MESSAGE)).not.toBeInTheDocument()
    })
  })
})
