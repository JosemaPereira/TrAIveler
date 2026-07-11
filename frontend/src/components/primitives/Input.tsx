import { useId } from 'react'
import type { InputHTMLAttributes } from 'react'
import styles from './Input.module.css'

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string
  error?: string
}

/**
 * Converts a label into a URL/DOM-safe slug, used as a fallback id so the
 * rendered <label> can be associated with its <input> without requiring
 * every caller to supply an explicit id.
 */
function slugify(value: string): string {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

/**
 * Core text input primitive. Always renders an associated <label> for
 * accessibility, and surfaces validation errors via aria-invalid /
 * aria-describedby pointing at a role="alert" error message.
 */
export function Input({ label, error, id, className, ...rest }: InputProps) {
  const reactId = useId()
  const inputId = id ?? (slugify(label) || reactId)
  const errorId = `${inputId}-error`
  const inputClassNames = [styles.input, error && styles.inputError, className]
    .filter(Boolean)
    .join(' ')

  return (
    <div className={styles.field}>
      <label htmlFor={inputId} className={styles.label}>
        {label}
      </label>
      <input
        id={inputId}
        className={inputClassNames}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? errorId : undefined}
        {...rest}
      />
      {error && (
        <p id={errorId} role="alert" className={styles.error}>
          {error}
        </p>
      )}
    </div>
  )
}
