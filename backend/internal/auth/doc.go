// Package auth provides authentication primitives: bcrypt password hashing and
// comparison (password.go), password-strength validation (validator.go), and
// the auth domain models plus registration/login request/response DTOs
// (models.go). JWT issuance, validation, and multi-key rotation live in the
// jwt subpackage. See docs/security.md and specs/004-security-auth-model/.
package auth
