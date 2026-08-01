import { useEffect, useState } from 'react'
import type { ChangeEvent, SyntheticEvent } from 'react'
import { Link } from 'react-router'

import { Button } from '../../../components/primitives/Button'
import { Input } from '../../../components/primitives/Input'
import { getErrorMessage, isAPIError } from '../../../lib/query-client'
import { useLogin } from '../hooks/useLogin'
import { isValidEmail } from '../validation'
import styles from './LoginForm.module.css'

export interface LoginFormProps {
  /** Invoked once the login mutation resolves successfully. */
  onSuccess?: () => void
}

// The backend deliberately answers identically for "unknown email" and
// "wrong password" (anti-enumeration, docs/security.md) — this form honors
// that by never rendering the server's own message for a 401, only this
// single generic copy, so no wording difference could leak which case it was.
const GENERIC_AUTH_ERROR = 'Invalid email or password.'

interface FieldValues {
  email: string
  password: string
}

function validateFields(values: FieldValues): Record<string, string> {
  const errors: Record<string, string> = {}

  if (!isValidEmail(values.email)) {
    errors.email = 'Enter a valid email address.'
  }
  if (values.password.length === 0) {
    errors.password = 'Password is required.'
  }

  return errors
}

function countdownLabel(seconds: number): string {
  const unit = seconds === 1 ? 'second' : 'seconds'
  return `Too many attempts. Try again in ${String(seconds)} ${unit}.`
}

/**
 * Login form for `POST /auth/login`. Client-side validation (email shape,
 * non-empty password — no strength check here, that only applies at
 * registration) blocks submission before any server round trip.
 *
 * Renders the rate-limit countdown `useLogin` seeds via `retryAfterSeconds`:
 * a real ticking countdown (not just the static starting number), disabling
 * submit while it is above zero, and restarting whenever a *new* 429 arrives
 * — keyed on the error object itself so two consecutive rate-limit hits with
 * an identical wait still restart the clock.
 */
export function LoginForm({ onSuccess }: LoginFormProps) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [clientErrors, setClientErrors] = useState<Record<string, string>>({})
  const [countdown, setCountdown] = useState<number | undefined>(undefined)

  const login = useLogin()

  useEffect(() => {
    if (login.retryAfterSeconds !== undefined) {
      setCountdown(login.retryAfterSeconds)
    }
    // Intentionally keyed on the error object, not on retryAfterSeconds: a
    // second 429 with the same wait must still restart the countdown, and
    // `login.error` is a fresh object per failed attempt (see api-client.ts).
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [login.error])

  useEffect(() => {
    if (!countdown) {
      return
    }
    const timeoutId = setTimeout(() => {
      setCountdown((current) =>
        current === undefined ? undefined : current - 1
      )
    }, 1000)
    return () => {
      clearTimeout(timeoutId)
    }
  }, [countdown])

  const isRateLimited = countdown !== undefined && countdown > 0
  const isSubmitting = login.isPending
  const isDisabled = isSubmitting || isRateLimited

  const is401 = isAPIError(login.error) && login.error.status === 401
  const is429 = isAPIError(login.error) && login.error.status === 429
  const formError =
    login.isError && !is429
      ? is401
        ? GENERIC_AUTH_ERROR
        : getErrorMessage(login.error)
      : null

  async function submit() {
    const values: FieldValues = { email, password }
    const errors = validateFields(values)
    if (Object.keys(errors).length > 0) {
      setClientErrors(errors)
      return
    }
    setClientErrors({})

    try {
      await login.mutateAsync(values)
      onSuccess?.()
    } catch {
      // Surfaced above via login.error/isError — nothing further to do.
    }
  }

  function handleSubmit(event: SyntheticEvent<HTMLFormElement>) {
    event.preventDefault()
    void submit()
  }

  return (
    <form
      className={styles.form}
      noValidate
      aria-busy={isSubmitting}
      onSubmit={handleSubmit}
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
        error={clientErrors.email}
        disabled={isDisabled}
      />
      <Input
        label="Password"
        type="password"
        name="password"
        autoComplete="current-password"
        value={password}
        onChange={(event: ChangeEvent<HTMLInputElement>) => {
          setPassword(event.target.value)
        }}
        error={clientErrors.password}
        disabled={isDisabled}
      />

      <p className={styles.forgotPassword}>
        <Link to="/password-reset">Forgot Password?</Link>
      </p>

      {isRateLimited && (
        <p role="status" aria-live="polite" className={styles.countdown}>
          {countdownLabel(countdown)}
        </p>
      )}

      <Button type="submit" disabled={isDisabled} aria-disabled={isDisabled}>
        {isSubmitting ? 'Logging in…' : 'Log In'}
      </Button>
    </form>
  )
}
