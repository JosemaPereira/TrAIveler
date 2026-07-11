import { AlertCircle } from 'lucide-react'
import { Button } from './Button'
import styles from './ErrorMessage.module.css'

export interface ErrorMessageProps {
  title?: string
  message: string
  onRetry?: () => void
}

const RETRY_LABEL = 'Try Again'

/**
 * Error-state primitive for failed data-dependent operations. Announces
 * itself immediately via role="alert"; the retry button is opt-in so
 * callers without a meaningful retry action (e.g. a permanent 403) can omit
 * it entirely.
 */
export function ErrorMessage({
  title = 'Error',
  message,
  onRetry,
}: ErrorMessageProps) {
  return (
    <div role="alert" className={styles.errorMessage}>
      <AlertCircle aria-hidden="true" className={styles.icon} />
      <div className={styles.content}>
        <h3 className={styles.title}>{title}</h3>
        <p className={styles.message}>{message}</p>
        {onRetry && (
          <Button variant="secondary" size="sm" onClick={onRetry}>
            {RETRY_LABEL}
          </Button>
        )}
      </div>
    </div>
  )
}
