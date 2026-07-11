import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { LoadingSpinner } from './LoadingSpinner'
import styles from './LoadingSpinner.module.css'
import { cssClass } from '../../test/cssModule'

describe('LoadingSpinner', () => {
  it('renders a status role with polite live announcements', () => {
    render(<LoadingSpinner />)

    const status = screen.getByRole('status')
    expect(status).toHaveAttribute('aria-live', 'polite')
  })

  it('renders the default label as visually-hidden text for screen readers', () => {
    render(<LoadingSpinner />)

    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('renders a custom label when provided', () => {
    render(<LoadingSpinner label="Fetching your trip..." />)

    expect(screen.getByText('Fetching your trip...')).toBeInTheDocument()
  })

  it('hides the spinning element from assistive technology', () => {
    render(<LoadingSpinner />)

    const status = screen.getByRole('status')
    const spinner = status.querySelector('[aria-hidden="true"]')
    expect(spinner).not.toBeNull()
  })

  it('applies the md size by default', () => {
    render(<LoadingSpinner />)

    const status = screen.getByRole('status')
    const spinner = status.querySelector(`.${cssClass(styles, 'spinner')}`)
    expect(spinner).toHaveClass(cssClass(styles, 'md'))
  })

  it.each([
    ['sm', cssClass(styles, 'sm')],
    ['md', cssClass(styles, 'md')],
    ['lg', cssClass(styles, 'lg')],
  ] as const)(
    'applies the %s size class when requested',
    (size, expectedClass) => {
      render(<LoadingSpinner size={size} />)

      const status = screen.getByRole('status')
      const spinner = status.querySelector(`.${cssClass(styles, 'spinner')}`)
      expect(spinner).toHaveClass(expectedClass)
    }
  )
})
