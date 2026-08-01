import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { Label } from './Label'
import styles from './Label.module.css'
import { cssClass } from '@/test/cssModule'

describe('<Label />', () => {
  describe('when rendered with a control id', () => {
    it('should associate itself with the matching control', async () => {
      const user = userEvent.setup()
      render(
        <>
          <Label htmlFor="email">Email</Label>
          <input id="email" type="text" />
        </>
      )

      await user.click(screen.getByText('Email'))

      expect(screen.getByRole('textbox', { name: 'Email' })).toHaveFocus()
    })

    it('should render its children as the label text', () => {
      render(<Label htmlFor="email">Email</Label>)

      expect(screen.getByText('Email')).toBeInTheDocument()
    })
  })

  describe('when the field is not required', () => {
    it('should not render a required indicator', () => {
      render(<Label htmlFor="email">Email</Label>)

      expect(screen.queryByText('*')).not.toBeInTheDocument()
      expect(screen.queryByText('(required)')).not.toBeInTheDocument()
    })
  })

  describe('when the field is required', () => {
    it('should render a decorative asterisk that assistive technology ignores', () => {
      render(
        <Label htmlFor="email" required>
          Email
        </Label>
      )

      expect(screen.getByText('*')).toHaveAttribute('aria-hidden', 'true')
    })

    it('should expose the requirement to screen readers as text', () => {
      render(
        <Label htmlFor="email" required>
          Email
        </Label>
      )

      const srText = screen.getByText('(required)')

      expect(srText).toBeInTheDocument()
      expect(srText).not.toHaveAttribute('aria-hidden')
      expect(srText).toHaveClass(cssClass(styles, 'visuallyHidden'))
    })
  })

  describe('when a className is provided', () => {
    it('should merge it with the base label class', () => {
      render(
        <Label htmlFor="email" className="custom-label">
          Email
        </Label>
      )

      const label = screen.getByText('Email')

      expect(label).toHaveClass(cssClass(styles, 'label'))
      expect(label).toHaveClass('custom-label')
    })
  })

  describe('when native label attributes are provided', () => {
    it('should forward them to the underlying element', () => {
      render(
        <Label htmlFor="email" id="email-label" title="Your email address">
          Email
        </Label>
      )

      const label = screen.getByText('Email')

      expect(label).toHaveAttribute('id', 'email-label')
      expect(label).toHaveAttribute('title', 'Your email address')
    })
  })
})
