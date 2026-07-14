import { createContext } from 'react'

/**
 * React context for authentication state.
 *
 * Scaffolding placeholder for Spec 004 (Security & Authentication/
 * Authorization Model, see specs/004-security-auth-model/). The real
 * context value (current user, tokens, loading state), its `AuthProvider`,
 * and the `useAuth` consumer hook are implemented in a later Spec 004
 * issue, once `lib/auth.ts` and the backend auth endpoints exist.
 */
export const AuthContext = createContext<undefined>(undefined)
