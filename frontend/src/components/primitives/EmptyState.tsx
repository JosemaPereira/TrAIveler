import type { ReactNode } from 'react'
import { Button } from './Button'
import styles from './EmptyState.module.css'

export interface EmptyStateAction {
  label: string
  onClick: () => void
}

export interface EmptyStateProps {
  icon?: ReactNode
  title: string
  message?: string
  action?: EmptyStateAction
}

/**
 * Empty-state primitive for data-dependent views with nothing to show
 * (e.g. "No trips yet"). The icon is purely decorative — hidden from
 * assistive technology — since the heading already conveys the state.
 */
export function EmptyState({ icon, title, message, action }: EmptyStateProps) {
  return (
    <div className={styles.emptyState}>
      {icon && (
        <div aria-hidden="true" className={styles.icon}>
          {icon}
        </div>
      )}
      <h3 className={styles.title}>{title}</h3>
      {message && <p className={styles.message}>{message}</p>}
      {action && <Button onClick={action.onClick}>{action.label}</Button>}
    </div>
  )
}
