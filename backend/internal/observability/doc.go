// Package observability provides security-event logging and correlation ID
// primitives for the auth vertical (Spec 004, see
// specs/004-security-auth-model/):
//   - LogSecurityEvent (logger.go) emits one structured JSON log line per
//     SecurityEvent. It is already live — internal/auth's service and handler
//     call it on every registration, successful/failed login, token refresh,
//     and logout. It does not persist to the security_events table (Phase 3,
//     out of scope here) and does not sanitize its details map, so callers
//     must never pass passwords, tokens, or full request payloads.
//   - GenerateCorrelationID (correlation.go) mints the UUID v4 threaded
//     through those log lines and echoed as the request's correlation ID.
//
// CloudWatch metrics emission — part of this package's originally planned
// scope — is not implemented yet; no metrics.go exists here. That remains a
// later Spec 004 item.
package observability
