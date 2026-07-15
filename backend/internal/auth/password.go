package auth

import "golang.org/x/crypto/bcrypt"

// bcryptCost is mandated by docs/security.md; deliberately not bcrypt.DefaultCost (10).
const bcryptCost = 12

// HashPassword hashes password with bcrypt at bcryptCost.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ComparePassword reports whether password matches hash.
func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
