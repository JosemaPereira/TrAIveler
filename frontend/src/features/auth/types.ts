import type { User } from '@/stores/auth-store'

/**
 * Wire types for the auth surface (`POST /api/v1/auth/*`).
 *
 * Field names are snake_case because this codebase has no case-conversion
 * layer — see `lib/api-client.ts`. Shapes mirror the Go structs that actually
 * serialize them (`backend/internal/auth/models.go`,
 * `backend/internal/subscription/models.go`), not the spec prose, which is
 * older than the implementation.
 */

/** Subscription status values, mirroring the subscriptions table CHECK constraint. */
export type SubscriptionStatus =
  'stub_pending' | 'active' | 'cancelled' | 'expired'

export interface Subscription {
  id: string
  user_id: string
  plan_id: string
  status: SubscriptionStatus
  /** Only present while a cancelled subscription is inside its 30-day grace period. */
  grace_period_ends_at?: string
  cancelled_at?: string
  created_at: string
}

export interface RegisterRequest {
  email: string
  password: string
  full_name: string
  /**
   * Opaque token from the (stubbed) payment provider. Omitted for a free
   * account, in which case the response carries no subscription.
   */
  payment_method_token?: string
}

export interface RegisterResponse {
  user: User
  subscription?: Subscription
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  user: User
}

/**
 * `GET /auth/me` body. Same `user` envelope as register and login, so all three
 * decode identically rather than this one being special-cased.
 */
export interface CurrentUserResponse {
  user: User
}
