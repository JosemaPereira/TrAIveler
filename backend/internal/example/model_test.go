package example

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnitValidate_ValidExample_ReturnsNil verifies that a fully valid Example
// passes validation with no error.
func TestUnitValidate_ValidExample_ReturnsNil(t *testing.T) {
	ex := &Example{
		Name:   "Ada Lovelace",
		Email:  "ada@example.com",
		Status: StatusActive,
	}

	err := ex.Validate()

	assert.Nil(t, err)
}

// TestUnitValidate_MissingName_ReturnsValidationError verifies that an empty
// or whitespace-only name is rejected with a field-level error.
func TestUnitValidate_MissingName_ReturnsValidationError(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ex := &Example{
				Name:   tt.in,
				Email:  "ada@example.com",
				Status: StatusActive,
			}

			err := ex.Validate()

			require.NotNil(t, err)
			assert.Equal(t, "validation_failed", err.Code)
			require.Len(t, err.Fields, 1)
			assert.Equal(t, "name", err.Fields[0].Field)
		})
	}
}

// TestUnitValidate_MissingEmail_ReturnsValidationError verifies that an empty
// or whitespace-only email is rejected with a field-level error.
func TestUnitValidate_MissingEmail_ReturnsValidationError(t *testing.T) {
	ex := &Example{
		Name:   "Ada Lovelace",
		Email:  "",
		Status: StatusActive,
	}

	err := ex.Validate()

	require.NotNil(t, err)
	assert.Equal(t, "validation_failed", err.Code)
	require.Len(t, err.Fields, 1)
	assert.Equal(t, "email", err.Fields[0].Field)
}

// TestUnitValidate_InvalidStatus_ReturnsValidationError verifies that a
// status outside StatusActive/StatusInactive is rejected.
func TestUnitValidate_InvalidStatus_ReturnsValidationError(t *testing.T) {
	ex := &Example{
		Name:   "Ada Lovelace",
		Email:  "ada@example.com",
		Status: "archived",
	}

	err := ex.Validate()

	require.NotNil(t, err)
	assert.Equal(t, "validation_failed", err.Code)
	require.Len(t, err.Fields, 1)
	assert.Equal(t, "status", err.Fields[0].Field)
}

// TestUnitValidate_MultipleFailures_ReturnsAllFieldErrors verifies that all
// failing fields are reported together, not just the first one found.
func TestUnitValidate_MultipleFailures_ReturnsAllFieldErrors(t *testing.T) {
	ex := &Example{
		Name:   "",
		Email:  "",
		Status: "bogus",
	}

	err := ex.Validate()

	require.NotNil(t, err)
	assert.Len(t, err.Fields, 3)
}
