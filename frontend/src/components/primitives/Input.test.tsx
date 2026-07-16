import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Input } from './Input'

describe('<Input />', () => {
  describe('when rendered with a label', () => {
    it('should render a labeled text field', () => {
      render(<Input label="Destination" />)

      expect(
        screen.getByRole('textbox', { name: 'Destination' })
      ).toBeInTheDocument()
    })

    it('should not show an error state by default', () => {
      render(<Input label="Destination" />)

      const input = screen.getByRole('textbox', { name: 'Destination' })

      expect(input).not.toHaveAttribute('aria-invalid')
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    })
  })

  describe('when no id is provided', () => {
    it('should derive an id from the label so clicking the label focuses the input', async () => {
      const user = userEvent.setup()
      render(<Input label="Trip name" />)

      await user.click(screen.getByText('Trip name'))

      expect(screen.getByRole('textbox', { name: 'Trip name' })).toHaveFocus()
    })

    it('should fall back to a generated id when the label slugifies to an empty string', () => {
      render(<Input label="!!!" />)

      const input = screen.getByRole('textbox', { name: '!!!' })

      expect(input.id).not.toBe('')
    })
  })

  describe('when an explicit id is provided', () => {
    it('should use it instead of deriving one', () => {
      render(<Input label="Email" id="custom-email-id" />)

      expect(screen.getByRole('textbox', { name: 'Email' })).toHaveAttribute(
        'id',
        'custom-email-id'
      )
    })
  })

  describe('when native attributes are provided', () => {
    it('should forward the name and required attributes', () => {
      render(<Input label="Destination" name="destination" required />)

      const input = screen.getByRole('textbox', { name: 'Destination' })

      expect(input).toHaveAttribute('name', 'destination')
      expect(input).toBeRequired()
    })
  })

  describe('when an error is provided', () => {
    it('should show an accessible error message and mark the field as invalid', () => {
      render(<Input label="Destination" error="Destination is required" />)

      const input = screen.getByRole('textbox', { name: 'Destination' })
      const alert = screen.getByRole('alert')

      expect(input).toHaveAttribute('aria-invalid', 'true')
      expect(alert).toHaveTextContent('Destination is required')
      expect(input).toHaveAttribute('aria-describedby', alert.id)
    })
  })

  describe('when the user types into the field', () => {
    it('should update the field value', async () => {
      const user = userEvent.setup()
      render(<Input label="Destination" />)
      const input = screen.getByRole('textbox', { name: 'Destination' })

      await user.type(input, 'Kyoto')

      expect(input).toHaveValue('Kyoto')
    })
  })
})
