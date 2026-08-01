import { Link, useNavigate, useSearchParams } from 'react-router'

import { LoginForm } from '../features/auth/components/LoginForm'
import { resolveLoginRedirect } from './login-redirect'

/**
 * Login page. Renders `LoginForm` and, on a successful login, navigates to
 * the sanitized `?redirect=` target `features/auth/session-expiry.ts`
 * attaches when it bounces an expired session here (`resolveLoginRedirect`
 * rejects anything that is not a same-origin relative path), defaulting to
 * `/dashboard` otherwise. `replace: true` keeps the login form out of the
 * back-button history.
 */
export function LoginPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  return (
    <div>
      <h1>Welcome Back</h1>
      <LoginForm
        onSuccess={() => {
          const target = resolveLoginRedirect(searchParams.get('redirect'))
          void navigate(target, { replace: true })
        }}
      />
      <p>
        Need an account? <Link to="/register">Sign Up</Link>
      </p>
    </div>
  )
}
