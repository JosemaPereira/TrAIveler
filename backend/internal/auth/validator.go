package auth

import (
	"fmt"
	"unicode"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

const (
	minPasswordLength = 8
	// maxPasswordLength matches bcrypt's own 72-byte input limit.
	maxPasswordLength = 72
)

// ValidatePassword enforces docs/security.md's password strength rules:
// 8-72 characters, at least one uppercase letter, one lowercase letter, and
// one digit. Failing rules are collected into a single combined error.
func ValidatePassword(password string) error {
	var fields []domainerrors.ValidationError

	if length := len([]rune(password)); length < minPasswordLength || length > maxPasswordLength {
		fields = append(fields, domainerrors.ValidationError{
			Field: "password",
			Error: fmt.Sprintf("Password must be between %d and %d characters", minPasswordLength, maxPasswordLength),
		})
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasUpper {
		fields = append(fields, domainerrors.ValidationError{
			Field: "password",
			Error: "Password must contain at least one uppercase letter",
		})
	}
	if !hasLower {
		fields = append(fields, domainerrors.ValidationError{
			Field: "password",
			Error: "Password must contain at least one lowercase letter",
		})
	}
	if !hasDigit {
		fields = append(fields, domainerrors.ValidationError{
			Field: "password",
			Error: "Password must contain at least one digit",
		})
	}

	if len(fields) > 0 {
		return domainerrors.Validation("Password does not meet strength requirements", fields...)
	}

	return nil
}
