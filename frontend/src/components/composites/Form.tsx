import { useState } from 'react'
import type { InputHTMLAttributes, SyntheticEvent } from 'react'
import { Button } from '../primitives/Button'
import { Input } from '../primitives/Input'
import styles from './Form.module.css'

export interface FormFieldConfig extends Omit<
  InputHTMLAttributes<HTMLInputElement>,
  'name'
> {
  name: string
  label: string
}

export interface FormSubmitError extends Error {
  fields?: Record<string, string>
}

export interface FormProps {
  fields: FormFieldConfig[]
  onSubmit: (data: Record<string, string>) => Promise<void>
  submitLabel?: string
}

const DEFAULT_SUBMIT_LABEL = 'Submit'
const SUBMITTING_LABEL = 'Submitting…'
const GENERIC_ERROR_MESSAGE = 'Something went wrong. Please try again.'

/**
 * Structural check for a submission rejection that carries per-field
 * validation errors (FormSubmitError), as opposed to a generic Error that
 * should be surfaced as a single form-level message.
 */
function hasFieldErrors(err: unknown): err is FormSubmitError {
  return (
    err instanceof Error &&
    'fields' in err &&
    Boolean((err as FormSubmitError).fields)
  )
}

/**
 * Config-driven form composite built on the Button and Input primitives.
 * Owns submission state (loading, form-level error, field-level errors) so
 * callers only need to supply field config and an async onSubmit handler.
 * Double-submit is prevented by disabling the fields and the submit button
 * (and reflecting it via aria-busy) for the duration of onSubmit.
 */
export function Form({
  fields,
  onSubmit,
  submitLabel = DEFAULT_SUBMIT_LABEL,
}: FormProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})

  async function handleSubmit(event: SyntheticEvent<HTMLFormElement>) {
    event.preventDefault()
    // Capture currentTarget synchronously: React nulls it out (matching
    // native DOM behavior) once this handler returns, so it would be stale
    // by the time the `await onSubmit(...)` below resolves.
    const form = event.currentTarget

    setFormError(null)
    setFieldErrors({})
    setIsSubmitting(true)

    const data = Object.fromEntries(new FormData(form).entries()) as Record<
      string,
      string
    >

    try {
      await onSubmit(data)
      form.reset()
      setFormError(null)
      setFieldErrors({})
    } catch (err) {
      if (hasFieldErrors(err)) {
        setFieldErrors(err.fields ?? {})
      } else {
        const message =
          err instanceof Error && err.message
            ? err.message
            : GENERIC_ERROR_MESSAGE
        setFormError(message)
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form
      className={styles.form}
      noValidate
      aria-busy={isSubmitting}
      onSubmit={(event) => {
        void handleSubmit(event)
      }}
    >
      {formError && (
        <p role="alert" className={styles.formError}>
          {formError}
        </p>
      )}
      {fields.map(({ name, label, disabled, ...rest }) => (
        <Input
          key={name}
          name={name}
          label={label}
          error={fieldErrors[name]}
          disabled={isSubmitting || disabled}
          {...rest}
        />
      ))}
      <Button type="submit" disabled={isSubmitting}>
        {isSubmitting ? SUBMITTING_LABEL : submitLabel}
      </Button>
    </form>
  )
}
