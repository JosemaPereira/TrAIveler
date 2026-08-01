import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Button } from './Button'
import styles from './Button.module.css'
import { cssClass } from '@/test/cssModule'

describe('<Button />', () => {
  describe('when rendered with default props', () => {
    it('should render its children as a button', () => {
      render(<Button>Save trip</Button>)

      expect(
        screen.getByRole('button', { name: 'Save trip' })
      ).toBeInTheDocument()
    })

    it('should apply the primary variant', () => {
      render(<Button>Confirm</Button>)

      expect(screen.getByRole('button', { name: 'Confirm' })).toHaveClass(
        cssClass(styles, 'primary')
      )
    })

    it('should apply the md size', () => {
      render(<Button>Default size</Button>)

      expect(screen.getByRole('button', { name: 'Default size' })).toHaveClass(
        cssClass(styles, 'md')
      )
    })
  })

  describe('when a variant is requested', () => {
    it.each([
      ['primary', cssClass(styles, 'primary')],
      ['secondary', cssClass(styles, 'secondary')],
      ['danger', cssClass(styles, 'danger')],
    ] as const)(
      'should apply the %s variant class',
      (variant, expectedClass) => {
        render(<Button variant={variant}>Action</Button>)

        expect(screen.getByRole('button', { name: 'Action' })).toHaveClass(
          expectedClass
        )
      }
    )
  })

  describe('when a size is requested', () => {
    it.each([
      ['sm', cssClass(styles, 'sm')],
      ['md', cssClass(styles, 'md')],
      ['lg', cssClass(styles, 'lg')],
    ] as const)('should apply the %s size class', (size, expectedClass) => {
      render(<Button size={size}>Action</Button>)

      expect(screen.getByRole('button', { name: 'Action' })).toHaveClass(
        expectedClass
      )
    })
  })

  describe('when clicked', () => {
    it('should invoke the onClick handler', async () => {
      const user = userEvent.setup()
      const handleClick = vi.fn()
      render(<Button onClick={handleClick}>Click me</Button>)

      await user.click(screen.getByRole('button', { name: 'Click me' }))

      expect(handleClick).toHaveBeenCalledTimes(1)
    })
  })

  describe('when disabled', () => {
    it('should render as disabled and not invoke onClick', async () => {
      const user = userEvent.setup()
      const handleClick = vi.fn()
      render(
        <Button disabled onClick={handleClick}>
          Unavailable
        </Button>
      )
      const button = screen.getByRole('button', { name: 'Unavailable' })

      await user.click(button)

      expect(button).toBeDisabled()
      expect(handleClick).not.toHaveBeenCalled()
    })
  })

  describe('when accessibility and styling props are provided', () => {
    it('should support an accessible label via aria-label', () => {
      render(<Button aria-label="Close dialog">X</Button>)

      expect(
        screen.getByRole('button', { name: 'Close dialog' })
      ).toBeInTheDocument()
    })

    it('should forward additional className values alongside variant styling', () => {
      render(<Button className="custom-class">Styled</Button>)

      const button = screen.getByRole('button', { name: 'Styled' })

      expect(button).toHaveClass('custom-class')
      expect(button).toHaveClass(cssClass(styles, 'primary'))
    })
  })
})
