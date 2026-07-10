import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Button } from './Button'
import styles from './Button.module.css'
import { cssClass } from '../../test/cssModule'

describe('Button', () => {
  it('renders its children as a button', () => {
    render(<Button>Save trip</Button>)

    expect(
      screen.getByRole('button', { name: 'Save trip' })
    ).toBeInTheDocument()
  })

  it('applies the primary variant by default', () => {
    render(<Button>Confirm</Button>)

    expect(screen.getByRole('button', { name: 'Confirm' })).toHaveClass(
      cssClass(styles, 'primary')
    )
  })

  it.each([
    ['primary', cssClass(styles, 'primary')],
    ['secondary', cssClass(styles, 'secondary')],
    ['danger', cssClass(styles, 'danger')],
  ] as const)(
    'applies the %s variant class when requested',
    (variant, expectedClass) => {
      render(<Button variant={variant}>Action</Button>)

      expect(screen.getByRole('button', { name: 'Action' })).toHaveClass(
        expectedClass
      )
    }
  )

  it('applies the md size by default', () => {
    render(<Button>Default size</Button>)

    expect(screen.getByRole('button', { name: 'Default size' })).toHaveClass(
      cssClass(styles, 'md')
    )
  })

  it.each([
    ['sm', cssClass(styles, 'sm')],
    ['md', cssClass(styles, 'md')],
    ['lg', cssClass(styles, 'lg')],
  ] as const)(
    'applies the %s size class when requested',
    (size, expectedClass) => {
      render(<Button size={size}>Action</Button>)

      expect(screen.getByRole('button', { name: 'Action' })).toHaveClass(
        expectedClass
      )
    }
  )

  it('invokes the onClick handler when clicked', async () => {
    const user = userEvent.setup()
    const handleClick = vi.fn()
    render(<Button onClick={handleClick}>Click me</Button>)

    await user.click(screen.getByRole('button', { name: 'Click me' }))

    expect(handleClick).toHaveBeenCalledTimes(1)
  })

  it('renders as disabled and does not invoke onClick when disabled', async () => {
    const user = userEvent.setup()
    const handleClick = vi.fn()
    render(
      <Button disabled onClick={handleClick}>
        Unavailable
      </Button>
    )

    const button = screen.getByRole('button', { name: 'Unavailable' })
    expect(button).toBeDisabled()

    await user.click(button)

    expect(handleClick).not.toHaveBeenCalled()
  })

  it('supports an accessible label via aria-label', () => {
    render(<Button aria-label="Close dialog">X</Button>)

    expect(
      screen.getByRole('button', { name: 'Close dialog' })
    ).toBeInTheDocument()
  })

  it('forwards additional className values alongside variant styling', () => {
    render(<Button className="custom-class">Styled</Button>)

    const button = screen.getByRole('button', { name: 'Styled' })
    expect(button).toHaveClass('custom-class')
    expect(button).toHaveClass(cssClass(styles, 'primary'))
  })
})
