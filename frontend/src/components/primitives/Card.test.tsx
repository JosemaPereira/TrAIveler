import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Card } from './Card'
import styles from './Card.module.css'
import { cssClass } from '../../test/cssModule'

describe('Card', () => {
  it('renders its children', () => {
    render(
      <Card>
        <p>Trip summary</p>
      </Card>
    )

    expect(screen.getByText('Trip summary')).toBeInTheDocument()
  })

  it('applies the md padding by default', () => {
    render(<Card>Content</Card>)

    expect(screen.getByText('Content')).toHaveClass(cssClass(styles, 'md'))
  })

  it.each([
    ['sm', cssClass(styles, 'sm')],
    ['md', cssClass(styles, 'md')],
    ['lg', cssClass(styles, 'lg')],
  ] as const)(
    'applies the %s padding class when requested',
    (padding, expectedClass) => {
      render(<Card padding={padding}>Content</Card>)

      expect(screen.getByText('Content')).toHaveClass(expectedClass)
    }
  )

  it('forwards an additional className alongside padding styling', () => {
    render(<Card className="custom-card">Content</Card>)

    const card = screen.getByText('Content')
    expect(card).toHaveClass('custom-card')
    expect(card).toHaveClass(cssClass(styles, 'card'))
  })
})
