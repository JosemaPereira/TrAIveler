import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'

import { Badge } from './Badge'
import styles from './Badge.module.css'
import { cssClass } from '@/test/cssModule'

describe('<Badge />', () => {
  describe('when rendered with the free variant', () => {
    it('should apply the free variant class', () => {
      render(<Badge variant="free">Free</Badge>)

      expect(screen.getByText('Free')).toHaveClass(cssClass(styles, 'free'))
    })
  })

  describe('when rendered with the paid variant', () => {
    it('should apply the paid variant class', () => {
      render(<Badge variant="paid">Paid</Badge>)

      expect(screen.getByText('Paid')).toHaveClass(cssClass(styles, 'paid'))
    })
  })

  describe('when rendered with the pending variant', () => {
    it('should apply the pending variant class', () => {
      render(<Badge variant="pending">Pending</Badge>)

      expect(screen.getByText('Pending')).toHaveClass(
        cssClass(styles, 'pending')
      )
    })
  })

  describe('when rendered', () => {
    it('should render its children as the visible label text', () => {
      render(<Badge variant="paid">Paid</Badge>)

      expect(screen.getByText('Paid')).toBeInTheDocument()
    })

    it('should expose the label as accessible text without extra ARIA', () => {
      render(<Badge variant="paid">Paid</Badge>)

      const badge = screen.getByText('Paid')

      expect(badge).not.toHaveAttribute('role')
      expect(badge).not.toHaveAttribute('aria-label')
    })

    it('should apply the base badge class', () => {
      render(<Badge variant="free">Free</Badge>)

      expect(screen.getByText('Free')).toHaveClass(cssClass(styles, 'badge'))
    })
  })

  describe('when a className is provided', () => {
    it('should merge it with the base and variant classes', () => {
      render(
        <Badge variant="free" className="custom-badge">
          Free
        </Badge>
      )

      const badge = screen.getByText('Free')

      expect(badge).toHaveClass(cssClass(styles, 'badge'))
      expect(badge).toHaveClass(cssClass(styles, 'free'))
      expect(badge).toHaveClass('custom-badge')
    })
  })

  describe('when native span attributes are provided', () => {
    it('should forward them to the underlying element', () => {
      render(
        <Badge variant="free" id="tier-badge" title="Subscription tier">
          Free
        </Badge>
      )

      const badge = screen.getByText('Free')

      expect(badge).toHaveAttribute('id', 'tier-badge')
      expect(badge).toHaveAttribute('title', 'Subscription tier')
    })
  })
})
