import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Input } from './Input'

describe('Input', () => {
  it('renders a labeled text field', () => {
    render(<Input label="Destination" />)

    expect(
      screen.getByRole('textbox', { name: 'Destination' })
    ).toBeInTheDocument()
  })

  it('derives an id from the label when none is provided, so clicking the label focuses the input', async () => {
    const user = userEvent.setup()
    render(<Input label="Trip name" />)

    await user.click(screen.getByText('Trip name'))

    expect(screen.getByRole('textbox', { name: 'Trip name' })).toHaveFocus()
  })

  it('falls back to a generated id when the label slugifies to an empty string', () => {
    render(<Input label="!!!" />)

    const input = screen.getByRole('textbox', { name: '!!!' })
    expect(input.id).not.toBe('')
  })

  it('uses an explicit id when provided instead of deriving one', () => {
    render(<Input label="Email" id="custom-email-id" />)

    expect(screen.getByRole('textbox', { name: 'Email' })).toHaveAttribute(
      'id',
      'custom-email-id'
    )
  })

  it('forwards the name and required attributes', () => {
    render(<Input label="Destination" name="destination" required />)

    const input = screen.getByRole('textbox', { name: 'Destination' })
    expect(input).toHaveAttribute('name', 'destination')
    expect(input).toBeRequired()
  })

  it('does not show an error state by default', () => {
    render(<Input label="Destination" />)

    const input = screen.getByRole('textbox', { name: 'Destination' })
    expect(input).not.toHaveAttribute('aria-invalid')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('shows an accessible error message and marks the field as invalid', () => {
    render(<Input label="Destination" error="Destination is required" />)

    const input = screen.getByRole('textbox', { name: 'Destination' })
    const alert = screen.getByRole('alert')

    expect(input).toHaveAttribute('aria-invalid', 'true')
    expect(alert).toHaveTextContent('Destination is required')
    expect(input).toHaveAttribute('aria-describedby', alert.id)
  })

  it('allows typing into the field', async () => {
    const user = userEvent.setup()
    render(<Input label="Destination" />)

    const input = screen.getByRole('textbox', { name: 'Destination' })
    await user.type(input, 'Kyoto')

    expect(input).toHaveValue('Kyoto')
  })
})
