import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Card } from './Card'
import styles from './Card.module.css'
import { cssClass } from '@/test/cssModule'

describe('<Card />', () => {
  describe('when rendered with default props', () => {
    it('should render its children', () => {
      render(
        <Card>
          <p>Trip summary</p>
        </Card>
      )

      expect(screen.getByText('Trip summary')).toBeInTheDocument()
    })

    it('should apply the md padding', () => {
      render(<Card>Content</Card>)

      expect(screen.getByText('Content')).toHaveClass(cssClass(styles, 'md'))
    })
  })

  describe('when a padding size is requested', () => {
    it.each([
      ['sm', cssClass(styles, 'sm')],
      ['md', cssClass(styles, 'md')],
      ['lg', cssClass(styles, 'lg')],
    ] as const)(
      'should apply the %s padding class',
      (padding, expectedClass) => {
        render(<Card padding={padding}>Content</Card>)

        expect(screen.getByText('Content')).toHaveClass(expectedClass)
      }
    )
  })

  describe('when an additional className is provided', () => {
    it('should forward it alongside the padding styling', () => {
      render(<Card className="custom-card">Content</Card>)

      const card = screen.getByText('Content')

      expect(card).toHaveClass('custom-card')
      expect(card).toHaveClass(cssClass(styles, 'card'))
    })
  })
})
