// Package security is a reserved, currently-empty namespace.
//
// The HTTP security concerns it was originally scaffolded to hold (Spec 008)
// were deliberately implemented in existing packages rather than here, to avoid
// a redundant layer:
//   - JWT access-token cookie validation → internal/middleware.Authenticate
//   - Per-IP request throttling → internal/middleware.RateLimit
//   - Login progressive-delay rate limiting → internal/auth/ratelimit
//   - Structured SecurityEvent logging → internal/observability.LogSecurityEvent
//
// This package holds no code and imports nothing; it is kept only as a namespace
// placeholder. Add code here only for a genuinely new security concern that fits
// none of the packages above.
package security
