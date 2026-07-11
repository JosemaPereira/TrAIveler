import styles from './LoadingSpinner.module.css'

export interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg'
  label?: string
}

/**
 * Loading-state primitive for async operations. Announces itself to
 * screen readers via role="status"/aria-live while keeping the spinning
 * graphic itself hidden from assistive technology (it is purely decorative;
 * the visually-hidden label is the actual announcement).
 */
export function LoadingSpinner({
  size = 'md',
  label = 'Loading...',
}: LoadingSpinnerProps) {
  const spinnerClassNames = [styles.spinner, styles[size]]
    .filter(Boolean)
    .join(' ')

  return (
    <div role="status" aria-live="polite" className={styles.status}>
      <div className={spinnerClassNames} aria-hidden="true" />
      <span className={styles.srOnly}>{label}</span>
    </div>
  )
}
