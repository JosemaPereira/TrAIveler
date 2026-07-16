// Package jwt implements TrAIveler's JWT token primitives: an RS256 access-token
// Generator, a Validator supporting zero-downtime multi-key rotation, and a
// Refresher that rotates refresh tokens.
//
// The Validator selects the verification key from each token's `kid` header, so
// a token signed by a previous key keeps validating until that key is retired —
// this is what makes rotation zero-downtime. See docs/security.md and
// specs/004-security-auth-model/data-model.md.
//
// Key material is sourced through KeyProvider and refresh-token persistence
// through RefreshTokenStore. Only in-memory/static implementations ship here;
// the Secrets Manager key loader and PostgreSQL store land with later flows
// (Spec 008 Sprint 6-7, Spec 004 Phase 3).
package jwt
