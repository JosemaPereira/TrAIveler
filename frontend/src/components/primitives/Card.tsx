import type { ReactNode } from 'react'
import styles from './Card.module.css'

export interface CardProps {
  children: ReactNode
  padding?: 'sm' | 'md' | 'lg'
  className?: string
}

/**
 * Surface primitive used to group related content. Styling (surface color,
 * shadow, and padding scale) comes entirely from design tokens.
 */
export function Card({ children, padding = 'md', className }: CardProps) {
  const classNames = [styles.card, styles[padding], className]
    .filter(Boolean)
    .join(' ')

  return <div className={classNames}>{children}</div>
}
