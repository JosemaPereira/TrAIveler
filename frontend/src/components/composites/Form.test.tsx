import { describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Form } from './Form'
import type { FormSubmitError } from './Form'

/**
 * Creates a promise whose resolution is controlled externally, so tests can
 * assert on the in-flight (pending) state of an async onSubmit handler
 * before deciding how it settles.
 */
function createDeferred<T = void>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

describe('<Form />', () => {
  describe('when rendered with field configs', () => {
    it('should render an Input for each field with the correct label', () => {
      render(
        <Form
          fields={[
            { name: 'destination', label: 'Destination' },
            { name: 'email', label: 'Email' },
          ]}
          onSubmit={vi.fn()}
        />
      )

      expect(
        screen.getByRole('textbox', { name: 'Destination' })
      ).toBeInTheDocument()
      expect(screen.getByRole('textbox', { name: 'Email' })).toBeInTheDocument()
    })
  })

  describe('when the form is submitted successfully', () => {
    it('should call onSubmit with a plain object built from the form values', async () => {
      const user = userEvent.setup()
      const handleSubmit = vi.fn().mockResolvedValue(undefined)
      render(
        <Form
          fields={[
            { name: 'destination', label: 'Destination' },
            { name: 'email', label: 'Email', type: 'email' },
          ]}
          onSubmit={handleSubmit}
        />
      )

      await user.type(
        screen.getByRole('textbox', { name: 'Destination' }),
        'Kyoto'
      )
      await user.type(
        screen.getByRole('textbox', { name: 'Email' }),
        'kyoto@example.com'
      )
      await user.click(screen.getByRole('button', { name: 'Submit' }))

      expect(handleSubmit).toHaveBeenCalledWith({
        destination: 'Kyoto',
        email: 'kyoto@example.com',
      })
    })

    it('should reset the form fields afterwards', async () => {
      const user = userEvent.setup()
      const handleSubmit = vi.fn().mockResolvedValue(undefined)
      render(
        <Form
          fields={[{ name: 'destination', label: 'Destination' }]}
          onSubmit={handleSubmit}
        />
      )
      const input = screen.getByRole('textbox', { name: 'Destination' })

      await user.type(input, 'Kyoto')
      await user.click(screen.getByRole('button', { name: 'Submit' }))

      await waitFor(() => expect(input).toHaveValue(''))
    })
  })

  describe('when a submission is in flight', () => {
    it('should disable the submit button and update its label', async () => {
      const user = userEvent.setup()
      const deferred = createDeferred()
      const handleSubmit = vi.fn().mockReturnValue(deferred.promise)
      render(
        <Form
          fields={[{ name: 'destination', label: 'Destination' }]}
          onSubmit={handleSubmit}
        />
      )

      await user.type(
        screen.getByRole('textbox', { name: 'Destination' }),
        'Kyoto'
      )
      await user.click(screen.getByRole('button', { name: 'Submit' }))

      expect(screen.getByRole('button', { name: 'Submitting…' })).toBeDisabled()

      deferred.resolve()
      await waitFor(() =>
        expect(
          screen.getByRole('button', { name: 'Submit' })
        ).not.toBeDisabled()
      )
    })

    it('should disable all Inputs, preventing edits mid-submit', async () => {
      const user = userEvent.setup()
      const deferred = createDeferred()
      const handleSubmit = vi.fn().mockReturnValue(deferred.promise)
      render(
        <Form
          fields={[{ name: 'destination', label: 'Destination' }]}
          onSubmit={handleSubmit}
        />
      )
      const input = screen.getByRole('textbox', { name: 'Destination' })

      await user.type(input, 'Kyoto')
      await user.click(screen.getByRole('button', { name: 'Submit' }))

      expect(input).toBeDisabled()

      deferred.resolve()
      await waitFor(() => expect(input).not.toBeDisabled())
    })

    it('should mark the form as aria-busy', async () => {
      const user = userEvent.setup()
      const deferred = createDeferred()
      const handleSubmit = vi.fn().mockReturnValue(deferred.promise)
      const { container } = render(
        <Form
          fields={[{ name: 'destination', label: 'Destination' }]}
          onSubmit={handleSubmit}
        />
      )

      await user.type(
        screen.getByRole('textbox', { name: 'Destination' }),
        'Kyoto'
      )
      await user.click(screen.getByRole('button', { name: 'Submit' }))

      expect(container.querySelector('form')).toHaveAttribute(
        'aria-busy',
        'true'
      )

      deferred.resolve()
      await waitFor(() =>
        expect(container.querySelector('form')).toHaveAttribute(
          'aria-busy',
          'false'
        )
      )
    })
  })

  describe('when the submission fails', () => {
    it('should show a form-level alert when the rejection is a plain Error without a fields map', async () => {
      const user = userEvent.setup()
      const handleSubmit = vi
        .fn()
        .mockRejectedValue(new Error('Network failure'))
      render(
        <Form
          fields={[{ name: 'destination', label: 'Destination' }]}
          onSubmit={handleSubmit}
        />
      )

      await user.type(
        screen.getByRole('textbox', { name: 'Destination' }),
        'Kyoto'
      )
      await user.click(screen.getByRole('button', { name: 'Submit' }))

      expect(await screen.findByRole('alert')).toHaveTextContent(
        'Network failure'
      )
    })

    it('should surface field-level errors next to the correct Input without a generic alert', async () => {
      const user = userEvent.setup()
      const fieldError: FormSubmitError = Object.assign(
        new Error('Validation failed'),
        { fields: { destination: 'Destination is required' } }
      )
      const handleSubmit = vi.fn().mockRejectedValue(fieldError)
      render(
        <Form
          fields={[
            { name: 'destination', label: 'Destination' },
            { name: 'email', label: 'Email' },
          ]}
          onSubmit={handleSubmit}
        />
      )

      await user.type(
        screen.getByRole('textbox', { name: 'Destination' }),
        'Kyoto'
      )
      await user.click(screen.getByRole('button', { name: 'Submit' }))

      const alerts = await screen.findAllByRole('alert')

      expect(alerts).toHaveLength(1)
      expect(alerts[0]).toHaveTextContent('Destination is required')
    })
  })

  describe('when a new submission starts after a failure', () => {
    it('should clear previously shown errors as soon as it starts', async () => {
      const user = userEvent.setup()
      const deferred = createDeferred()
      const handleSubmit = vi
        .fn()
        .mockRejectedValueOnce(new Error('First failure'))
        .mockReturnValueOnce(deferred.promise)
      render(
        <Form
          fields={[{ name: 'destination', label: 'Destination' }]}
          onSubmit={handleSubmit}
        />
      )
      await user.type(
        screen.getByRole('textbox', { name: 'Destination' }),
        'Kyoto'
      )
      await user.click(screen.getByRole('button', { name: 'Submit' }))
      expect(await screen.findByRole('alert')).toHaveTextContent(
        'First failure'
      )

      await user.click(screen.getByRole('button', { name: 'Submit' }))

      expect(screen.queryByRole('alert')).not.toBeInTheDocument()

      deferred.resolve()
      await waitFor(() => {
        expect(handleSubmit).toHaveBeenCalledTimes(2)
      })
    })
  })
})
