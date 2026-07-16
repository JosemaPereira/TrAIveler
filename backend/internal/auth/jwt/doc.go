// Package jwt implements TrAIveler's JWT token primitives: an access-token
// Generator (RS256), a multi-key Validator that supports zero-downtime signing
// key rotation, and a Refresher that exchanges a refresh token for a fresh
// access/refresh token pair.
//
// The design follows docs/security.md (JWT RS256, multi-key rotation) and
// specs/004-security-auth-model/data-model.md (JWTSigningKey / RefreshToken
// entities). Key selection during validation is driven by the token's `kid`
// header: any key that a KeyProvider still reports as active is accepted, so a
// token signed by a previous key keeps validating until that key is retired —
// this is what makes rotation zero-downtime.
//
// Key material sourcing is abstracted behind KeyProvider and the refresh-token
// persistence behind RefreshTokenStore. This package deliberately ships only
// in-memory/static implementations (StaticKeyProvider) plus these interfaces;
// the AWS Secrets Manager-backed key loader and the PostgreSQL-backed refresh
// token store are wired by the later login/registration flows (Spec 008 Sprint
// 6-7) and Spec 004 Phase 3, which own the runtime key/store machinery.
//
// This is the sibling of the flat internal/auth package (password hashing and
// validation); JWT lives in its own subpackage because it is a cohesive
// sub-domain with several exported types and to avoid a filename clash with the
// password validator.go.
package jwt
