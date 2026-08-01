import { Link, useNavigate } from 'react-router'

import { RegisterForm } from '../features/auth/components/RegisterForm'

/**
 * Registration page (Spec 008, 008-T060/T061/T092). Renders `RegisterForm`
 * and, on a successful registration (either submit path — the form's hook
 * already logged the user into the auth store), navigates to `/dashboard`.
 * `replace: true` keeps the register form out of the back-button history so a
 * signed-in user cannot navigate back into it.
 */
export function RegisterPage() {
  const navigate = useNavigate()

  return (
    <div>
      <h1>Create Your Account</h1>
      <RegisterForm
        onSuccess={() => {
          void navigate('/dashboard', { replace: true })
        }}
      />
      <p>
        Already have an account? <Link to="/login">Log In</Link>
      </p>
    </div>
  )
}
