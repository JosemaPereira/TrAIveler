import { createContext } from 'react'

/**
 * Unused placeholder left over from early Spec 004 scaffolding. Session state
 * ended up living in `stores/auth-store.ts` (Zustand) rather than React
 * context, so this file has no remaining imports.
 */
export const AuthContext = createContext<undefined>(undefined)
