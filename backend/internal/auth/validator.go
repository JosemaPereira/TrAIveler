package auth

import (
	"fmt"
	"unicode"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

const (
	// minPasswordLength is the mandated minimum password length
	// (docs/security.md, "Password Security").
	minPasswordLength = 8
	// maxPasswordLength matches bcrypt's own 72-byte input limit, the same
	// bound HashPassword (password.go) relies on implicitly.
	maxPasswordLength = 72
)

// ValidatePassword checks password against this project's strength rules
// (docs/security.md, "Password Security"): length between 8 and 72
// characters inclusive, at least one uppercase letter, one lowercase
// letter, and one digit. It returns nil when password satisfies every
// rule.
//
// All failing rules are collected into a single combined error rather than
// stopping at the first one, mirroring internal/example's Validate()
// convention (see internal/example/model.go) so a caller can report every
// issue to the user in one round trip.
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
