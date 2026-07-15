// Package auth provides authentication primitives: JWT issuance/validation
// and password hashing. Scaffolding for Spec 004; see
// specs/004-security-auth-model/.
package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// MinimumSupportedBcryptCost forces bcrypt into a direct dependency ahead of
// HashPassword's implementation (password.go is the source of truth for the
// actual mandated cost).
const MinimumSupportedBcryptCost = bcrypt.MinCost

// SigningMethod forces golang-jwt/jwt/v5 into a direct dependency ahead of
// the JWT issuance/validation implementation. Mandated by docs/security.md.
var SigningMethod = jwt.SigningMethodRS256
