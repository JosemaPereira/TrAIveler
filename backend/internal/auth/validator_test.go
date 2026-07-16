package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "when the password satisfies every rule",
			password: "Correct1Horse",
			wantErr:  false,
		},
		{
			name: "when the password is below the 8-char minimum",
			// "Abcdef1" is 7 chars: below the 8-char minimum.
			password: "Abcdef1",
			wantErr:  true,
		},
		{
			name: "when the password is exactly at the 8-char minimum",
			// All other rules satisfied at the lower boundary.
			password: "Abcdefg1",
			wantErr:  false,
		},
		{
			name: "when the password is exactly at the 72-char maximum",
			// bcrypt's own input limit; all other rules satisfied.
			password: "Aa1" + strings.Repeat("b", 69),
			wantErr:  false,
		},
		{
			name:     "when the password is one char past the 72-char maximum",
			password: "Aa1" + strings.Repeat("b", 70),
			wantErr:  true,
		},
		{
			name:     "when the password is missing an uppercase letter",
			password: "lowercase1",
			wantErr:  true,
		},
		{
			name:     "when the password is missing a lowercase letter",
			password: "UPPERCASE1",
			wantErr:  true,
		},
		{
			name:     "when the password is missing a digit",
			password: "NoDigitsHere",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)

			assertValidationResult(t, err, tt.wantErr)
		})
	}
}

func TestValidatePassword_MultipleFailures_ReturnsCombinedError(t *testing.T) {
	// Too short, no uppercase, no digit: three rules fail at once.
	err := ValidatePassword("abc")

	require.Error(t, err)
	domainErr, ok := err.(*domainerrors.DomainError)
	require.True(t, ok, "expected *domainerrors.DomainError, got %T", err)
	assert.Equal(t, "validation_failed", domainErr.Code)
	assert.Len(t, domainErr.Fields, 3)
}

func assertValidationResult(t *testing.T, err error, wantErr bool) {
	t.Helper()

	if wantErr {
		assert.Error(t, err)
		return
	}
	assert.NoError(t, err)
}
