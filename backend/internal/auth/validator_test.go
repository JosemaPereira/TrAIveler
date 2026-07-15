package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

func TestValidatePassword_ValidPassword_ReturnsNil(t *testing.T) {
	err := ValidatePassword("Correct1Horse")

	assert.NoError(t, err)
}

func TestValidatePassword_SevenChars_ReturnsError(t *testing.T) {
	// "Abcdef1" is 7 chars: below the 8-char minimum.
	err := ValidatePassword("Abcdef1")

	assert.Error(t, err)
}

func TestValidatePassword_EightChars_ReturnsNil(t *testing.T) {
	// Exactly at the 8-char minimum boundary; all other rules satisfied.
	pw := "Abcdefg1"
	require.Len(t, pw, 8)

	err := ValidatePassword(pw)

	assert.NoError(t, err)
}

func TestValidatePassword_SeventyTwoChars_ReturnsNil(t *testing.T) {
	// Exactly at the 72-char maximum boundary (bcrypt's own input limit);
	// all other rules satisfied.
	pw := "Aa1" + strings.Repeat("b", 69)
	require.Len(t, pw, 72)

	err := ValidatePassword(pw)

	assert.NoError(t, err)
}

func TestValidatePassword_SeventyThreeChars_ReturnsError(t *testing.T) {
	// One char past the 72-char maximum boundary.
	pw := "Aa1" + strings.Repeat("b", 70)
	require.Len(t, pw, 73)

	err := ValidatePassword(pw)

	assert.Error(t, err)
}

func TestValidatePassword_MissingUppercase_ReturnsError(t *testing.T) {
	err := ValidatePassword("lowercase1")

	assert.Error(t, err)
}

func TestValidatePassword_MissingLowercase_ReturnsError(t *testing.T) {
	err := ValidatePassword("UPPERCASE1")

	assert.Error(t, err)
}

func TestValidatePassword_MissingDigit_ReturnsError(t *testing.T) {
	err := ValidatePassword("NoDigitsHere")

	assert.Error(t, err)
}

// TestValidatePassword_MultipleFailures_ReturnsCombinedError verifies that
// all failing rules are reported together, not just the first one found,
// mirroring internal/example's Validate() collect-all convention.
func TestValidatePassword_MultipleFailures_ReturnsCombinedError(t *testing.T) {
	// Too short, no uppercase, no digit: three rules fail at once.
	err := ValidatePassword("abc")

	require.Error(t, err)
	domainErr, ok := err.(*domainerrors.DomainError)
	require.True(t, ok, "expected *domainerrors.DomainError, got %T", err)
	assert.Equal(t, "validation_failed", domainErr.Code)
	assert.Len(t, domainErr.Fields, 3)
}
