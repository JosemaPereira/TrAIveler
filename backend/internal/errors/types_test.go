package errors_test

import (
	"errors"
	"fmt"
	"testing"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotFound_SetsCodeAndMessage(t *testing.T) {
	err := domainerrors.NotFound("trip", "trip-123")

	assert.Equal(t, "not_found", err.Code)
	assert.Contains(t, err.Message, "trip")
	assert.Contains(t, err.Message, "trip-123")
}

func TestValidation_SetsCodeMessageAndFields(t *testing.T) {
	fields := []domainerrors.ValidationError{
		{Field: "email", Error: "Email address is already registered"},
		{Field: "start_date", Error: "Start date must be in the future"},
	}

	err := domainerrors.Validation("One or more fields failed validation", fields...)

	assert.Equal(t, "validation_failed", err.Code)
	assert.Equal(t, "One or more fields failed validation", err.Message)
	require.Len(t, err.Fields, 2)
	assert.Equal(t, "email", err.Fields[0].Field)
	assert.Equal(t, "Email address is already registered", err.Fields[0].Error)
}

func TestValidation_NoFields_FieldsIsEmpty(t *testing.T) {
	err := domainerrors.Validation("One or more fields failed validation")

	assert.Empty(t, err.Fields)
}

func TestUnauthorized_SetsCode(t *testing.T) {
	err := domainerrors.Unauthorized("missing or invalid token")

	assert.Equal(t, "authentication_required", err.Code)
	assert.Equal(t, "missing or invalid token", err.Message)
}

func TestForbidden_SetsCode(t *testing.T) {
	err := domainerrors.Forbidden("insufficient permissions")

	assert.Equal(t, "forbidden", err.Code)
	assert.Equal(t, "insufficient permissions", err.Message)
}

func TestConflict_SetsCode(t *testing.T) {
	err := domainerrors.Conflict("version mismatch")

	assert.Equal(t, "conflict", err.Code)
	assert.Equal(t, "version mismatch", err.Message)
}

func TestDomainError_Error_ReturnsCodeAndMessage(t *testing.T) {
	err := domainerrors.NotFound("trip", "trip-123")

	assert.Contains(t, err.Error(), "not_found")
	assert.Contains(t, err.Error(), err.Message)
}

func TestDomainError_Error_IncludesWrappedCause(t *testing.T) {
	cause := errors.New("connection refused")
	err := &domainerrors.DomainError{
		Code:    "internal_error",
		Message: "database unavailable",
		Err:     cause,
	}

	assert.Contains(t, err.Error(), "connection refused")
}

func TestDomainError_Unwrap_ReturnsWrappedCause(t *testing.T) {
	cause := errors.New("connection refused")
	err := &domainerrors.DomainError{
		Code:    "internal_error",
		Message: "database unavailable",
		Err:     cause,
	}

	assert.Same(t, cause, err.Unwrap())
}

func TestErrorsAs_UnwrapsDomainErrorThroughStandardWrapping(t *testing.T) {
	original := domainerrors.NotFound("trip", "trip-123")
	wrapped := fmt.Errorf("handler failed: %w", original)

	var target *domainerrors.DomainError
	require.True(t, errors.As(wrapped, &target))
	assert.Same(t, original, target)
}

func TestErrorsIs_MatchesSentinelThroughDomainErrorCause(t *testing.T) {
	sentinel := errors.New("record not found in store")
	domainErr := &domainerrors.DomainError{
		Code:    "not_found",
		Message: "trip not found",
		Err:     sentinel,
	}
	wrapped := fmt.Errorf("lookup failed: %w", domainErr)

	assert.True(t, errors.Is(wrapped, sentinel))
}
