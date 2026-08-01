/**
 * Client-side validation helpers shared by RegisterForm and LoginForm.
 *
 * Password strength mirrors `backend/internal/auth/validator.go` exactly (8-72
 * characters, at least one uppercase letter, one lowercase letter, one digit)
 * so a locally-rejected password never diverges from what the server would
 * reject — the point is to fail fast without a round trip, not to invent a
 * separate policy.
 */

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

const MIN_PASSWORD_LENGTH = 8
const MAX_PASSWORD_LENGTH = 72

/** Basic shape check — good enough to catch typos before a server round trip. */
export function isValidEmail(email: string): boolean {
  return EMAIL_PATTERN.test(email.trim())
}

/**
 * Returns the list of failing password rules (empty when the password is
 * strong enough), in the same order the backend validator checks them.
 */
export function validatePasswordStrength(password: string): string[] {
  const errors: string[] = []

  if (
    password.length < MIN_PASSWORD_LENGTH ||
    password.length > MAX_PASSWORD_LENGTH
  ) {
    errors.push(
      `Password must be between ${String(MIN_PASSWORD_LENGTH)} and ${String(MAX_PASSWORD_LENGTH)} characters`
    )
  }
  if (!/[A-Z]/.test(password)) {
    errors.push('Password must contain at least one uppercase letter')
  }
  if (!/[a-z]/.test(password)) {
    errors.push('Password must contain at least one lowercase letter')
  }
  if (!/[0-9]/.test(password)) {
    errors.push('Password must contain at least one digit')
  }

  return errors
}

/** Required, non-empty after trimming whitespace. */
export function isValidFullName(fullName: string): boolean {
  return fullName.trim().length > 0
}
