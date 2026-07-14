package example

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitValidate_ValidExample_ReturnsNil(t *testing.T) {
	ex := &Example{
		Name:   "Ada Lovelace",
		Email:  "ada@example.com",
		Status: StatusActive,
	}

	err := ex.Validate()

	assert.Nil(t, err)
}

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
