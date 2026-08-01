import { useState } from 'react'
import type { ChangeEvent, SyntheticEvent } from 'react'

import { Button } from '../../../components/primitives/Button'
import { Input } from '../../../components/primitives/Input'
import { Label } from '../../../components/primitives/Label'
import { getErrorMessage, getFieldErrors } from '../../../lib/query-client'
import { useRegister } from '../hooks/useRegister'
import {
  isValidEmail,
  isValidFullName,
  validatePasswordStrength,
} from '../validation'
import styles from './RegisterForm.module.css'

export interface RegisterFormProps {
  /** Invoked once registration succeeds, on either submit path. */
  onSuccess?: () => void
}

// The whole payment provider is a labeled stub (`StubPaymentProvider` always
// succeeds — see backend/internal/subscription/payment) — there is no real
// checkout page to collect a token from (that is a separate, still-Backlog
// ticket), so this is a fixed client-side stand-in, consistent with the
// backend's own "[DEMO]" stub labeling. Exported so tests can assert on the
// exact value rather than merely "some string".
export const DEMO_PAYMENT_TOKEN = 'demo-payment-token'

interface FieldValues {
  email: string
  password: string
  full_name: string
}

function validateFields(values: FieldValues): Record<string, string> {
  const errors: Record<string, string> = {}

  if (!isValidEmail(values.email)) {
    errors.email = 'Enter a valid email address.'
  }

  const passwordErrors = validatePasswordStrength(values.password)
  if (passwordErrors.length > 0) {
    errors.password = passwordErrors.join(' ')
  }

  if (!isValidFullName(values.full_name)) {
    errors.full_name = 'Full name is required.'
  }

  return errors
}

/**
 * Registration form with two submit outcomes that both call `useRegister`
 * (a single `POST /auth/register` request either way):
 *
 * - "Create Free Account": registers with no `payment_method_token`.
 * - "Continue to Payment [DEMO]" (revealed by the demo checkbox): registers
 *   with a fixed demo token, since the real checkout page is out of scope
 *   here (see `DEMO_PAYMENT_TOKEN`).
 *
 * Client-side validation (email shape, backend-matching password strength,
 * non-empty full name) blocks submission and shows inline errors without a
 * server round trip. Server-side errors are rendered via
 * `getErrorMessage`/`getFieldErrors`: a 409 as a form-level duplicate-email
 * message, a 422 as per-field messages.
 */
export function RegisterForm({ onSuccess }: RegisterFormProps) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [fullName, setFullName] = useState('')
  const [wantsPayment, setWantsPayment] = useState(false)
  const [clientErrors, setClientErrors] = useState<Record<string, string>>({})

  const register = useRegister()
  const isSubmitting = register.isPending

  const serverFieldErrors = getFieldErrors(register.error)
  const hasClientErrors = Object.keys(clientErrors).length > 0
  const fieldErrors = hasClientErrors ? clientErrors : serverFieldErrors
  const formError =
    !hasClientErrors &&
    register.isError &&
    Object.keys(serverFieldErrors).length === 0
      ? getErrorMessage(register.error)
      : null

  async function submit(withPayment: boolean) {
    const values: FieldValues = { email, password, full_name: fullName }
    const errors = validateFields(values)
    if (Object.keys(errors).length > 0) {
      setClientErrors(errors)
      return
    }
    setClientErrors({})

    try {
      await register.mutateAsync({
        ...values,
        ...(withPayment ? { payment_method_token: DEMO_PAYMENT_TOKEN } : {}),
      })
      onSuccess?.()
    } catch {
      // Surfaced above via register.error/isError — nothing further to do.
    }
  }

  function handleFreeSubmit(event: SyntheticEvent<HTMLFormElement>) {
    event.preventDefault()
    void submit(false)
  }

  function handlePaymentClick() {
    void submit(true)
  }

  return (
    <form
      className={styles.form}
      noValidate
      aria-busy={isSubmitting}
      onSubmit={handleFreeSubmit}
    >
      {formError && (
        <p role="alert" className={styles.formError}>
          {formError}
        </p>
      )}

      <Input
        label="Email"
        type="email"
        name="email"
        autoComplete="email"
        value={email}
        onChange={(event: ChangeEvent<HTMLInputElement>) => {
          setEmail(event.target.value)
        }}
        error={fieldErrors.email}
        disabled={isSubmitting}
      />
      <Input
        label="Password"
        type="password"
        name="password"
        autoComplete="new-password"
        value={password}
        onChange={(event: ChangeEvent<HTMLInputElement>) => {
          setPassword(event.target.value)
        }}
        error={fieldErrors.password}
        disabled={isSubmitting}
      />
      <Input
        label="Full Name"
        type="text"
        name="full_name"
        autoComplete="name"
        value={fullName}
        onChange={(event: ChangeEvent<HTMLInputElement>) => {
          setFullName(event.target.value)
        }}
        error={fieldErrors.full_name}
        disabled={isSubmitting}
      />

      <div className={styles.paymentOption}>
        <input
          type="checkbox"
          id="wants-payment"
          className={styles.checkbox}
          checked={wantsPayment}
          onChange={(event: ChangeEvent<HTMLInputElement>) => {
            setWantsPayment(event.target.checked)
          }}
          disabled={isSubmitting}
        />
        <Label htmlFor="wants-payment">
          [DEMO] Add a payment method now
        </Label>
      </div>

      <div className={styles.actions}>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Creating account…' : 'Create Free Account'}
        </Button>
        {wantsPayment && (
          <Button
            type="button"
            variant="secondary"
            disabled={isSubmitting}
            onClick={handlePaymentClick}
          >
            {isSubmitting ? 'Processing…' : 'Continue to Payment [DEMO]'}
          </Button>
        )}
      </div>
    </form>
  )
}
