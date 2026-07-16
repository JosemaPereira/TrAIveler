import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { LoadingSpinner } from './LoadingSpinner'
import styles from './LoadingSpinner.module.css'
import { cssClass } from '../../test/cssModule'

describe('<LoadingSpinner />', () => {
  describe('when rendered with default props', () => {
    it('should render a status role with polite live announcements', () => {
      render(<LoadingSpinner />)

      const status = screen.getByRole('status')

      expect(status).toHaveAttribute('aria-live', 'polite')
    })

    it('should render the default label as visually-hidden text for screen readers', () => {
      render(<LoadingSpinner />)

      expect(screen.getByText('Loading...')).toBeInTheDocument()
    })

    it('should hide the spinning element from assistive technology', () => {
      render(<LoadingSpinner />)

      const status = screen.getByRole('status')

      expect(status.querySelector('[aria-hidden="true"]')).not.toBeNull()
    })

    it('should apply the md size', () => {
      render(<LoadingSpinner />)

      const status = screen.getByRole('status')
      const spinner = status.querySelector(`.${cssClass(styles, 'spinner')}`)

      expect(spinner).toHaveClass(cssClass(styles, 'md'))
    })
  })

  describe('when a custom label is provided', () => {
    it('should render the custom label', () => {
      render(<LoadingSpinner label="Fetching your trip..." />)

      expect(screen.getByText('Fetching your trip...')).toBeInTheDocument()
    })
  })

  describe('when a size is requested', () => {
    it.each([
      ['sm', cssClass(styles, 'sm')],
      ['md', cssClass(styles, 'md')],
      ['lg', cssClass(styles, 'lg')],
    ] as const)('should apply the %s size class', (size, expectedClass) => {
      render(<LoadingSpinner size={size} />)

      const status = screen.getByRole('status')
      const spinner = status.querySelector(`.${cssClass(styles, 'spinner')}`)

      expect(spinner).toHaveClass(expectedClass)
    })
  })
})
