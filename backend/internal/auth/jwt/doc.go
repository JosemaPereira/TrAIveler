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
// through RefreshTokenStore — both interfaces, implemented outside this
// package. KeyProvider only has an in-memory StaticKeyProvider today; the
// Secrets Manager key loader is still pending (roadmap tasks
// 004-T070–004-T072, gated on 003-T045). RefreshTokenStore is already backed
// by a real PostgreSQL implementation (internal/auth.RefreshStore over
// PostgresRefreshTokenRepository, wired in cmd/api/auth.go) — that part
// landed in Sprint 6 (issue #175).
package jwt
