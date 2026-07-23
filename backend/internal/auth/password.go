package auth

import "golang.org/x/crypto/bcrypt"

// DefaultBcryptCost is the bcrypt cost mandated by docs/security.md (deliberately
// above bcrypt.DefaultCost of 10). It is also the fallback HashPassword uses when
// a caller supplies a cost outside bcrypt's valid range.
const DefaultBcryptCost = 12

// HashPassword hashes password with bcrypt at the given cost. A cost outside
// bcrypt's supported range [bcrypt.MinCost, bcrypt.MaxCost] falls back to
// DefaultBcryptCost, so a misconfigured BCRYPT_COST can neither error nor silently
// produce a weaker-than-intended hash.
func HashPassword(password string, cost int) (string, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = DefaultBcryptCost
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ComparePassword reports whether password matches hash.
func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
