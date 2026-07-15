package auth

import "golang.org/x/crypto/bcrypt"

// bcryptCost is the mandated work factor for password hashing, per
// docs/security.md ("Bcrypt hashing with cost factor 12 (2^12 = 4096
// iterations)"). It is intentionally not bcrypt.DefaultCost (10).
const bcryptCost = 12

// HashPassword hashes the given plaintext password with bcrypt at the
// mandated cost factor (see bcryptCost). The returned string is the full
// bcrypt-encoded hash (algorithm, cost, salt, and digest), safe to store
// directly in the users table's password_hash column.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ComparePassword reports whether the given plaintext password matches the
// bcrypt hash produced by HashPassword. It returns a non-nil error when the
// password does not match or the hash is malformed.
func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
