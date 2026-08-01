import type { HTMLAttributes, ReactNode } from 'react'

import styles from './Badge.module.css'

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  /**
   * Status this badge represents. Drives the visual variant only — the
   * visible label comes from `children`, so copy (and its translation, if
   * ever needed) stays out of this primitive.
   */
  variant: 'free' | 'paid' | 'pending'
  children: ReactNode
}

/**
 * Status indicator primitive for subscription tiers (free/paid) and
 * transitional states (pending).
 *
 * Renders as plain text inside a <span>: a static status label needs no
 * ARIA role or label — its text content already conveys the state to
 * assistive technology (docs/ui-guidelines.md: use ARIA only when a native
 * element cannot express the semantics). It is not interactive, so it is not
 * part of the keyboard tab order.
 */
export function Badge({ variant, className, children, ...rest }: BadgeProps) {
  const badgeClassNames = [styles.badge, styles[variant], className]
    .filter(Boolean)
    .join(' ')

  return (
    <span className={badgeClassNames} {...rest}>
      {children}
    </span>
  )
}
