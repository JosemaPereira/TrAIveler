// Package security provides cross-cutting HTTP security middleware and
// security event logging: JWT cookie validation, request ID generation,
// rate limiting, and structured SecurityEvent logging to CloudWatch. It is
// currently scaffolding for Spec 008 (Authentication & Collaboration UX,
// see specs/008-auth-collaboration-ux/) — the logger and middleware
// implementations land in later Spec 008 issues
// (specs/008-auth-collaboration-ux/tasks.md, T026+). Distinct from the
// existing internal/authorization (RBAC) and internal/observability
// (metrics/correlation) packages scaffolded for Spec 004.
package security
