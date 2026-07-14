// Package auth provides authentication primitives: JWT issuance/validation
// (with multi-key rotation) and password hashing, plus the HTTP handlers
// and middleware built on top of them. It is currently scaffolding for
// Spec 004 (Security & Authentication/Authorization Model, see
// specs/004-security-auth-model/) — the JWT, password hashing, handler, and
// middleware implementations land in a later Spec 004 issue
// (specs/004-security-auth-model/tasks.md, T012+).
package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// MinimumSupportedBcryptCost re-exports bcrypt's own floor for the work
// factor. It is declared here, ahead of the password hashing implementation
// (specs/004-security-auth-model/tasks.md T012), purely so this package
// pulls in golang.org/x/crypto/bcrypt as a genuine direct dependency
// instead of an indirect, transitive one. The mandated work factor itself
// is 12 (see docs/security.md, "Password Security") — T012's
// HashPassword/ComparePassword implementation is the source of truth for
// that value, not this constant.
const MinimumSupportedBcryptCost = bcrypt.MinCost

// SigningMethod is the JWT signing algorithm mandated by docs/security.md
// ("JWT RS256 with multi-key rotation"). It is declared here, ahead of the
// JWT issuance/validation implementation (specs/004-security-auth-model/
// tasks.md, Foundational phase), purely so this package pulls in
// github.com/golang-jwt/jwt/v5 as a genuine direct dependency instead of an
// indirect, transitive one.
var SigningMethod = jwt.SigningMethodRS256
