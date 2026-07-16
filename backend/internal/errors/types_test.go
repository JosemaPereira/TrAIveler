package errors_test

import (
	"errors"
	"fmt"
	"testing"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainErrorConstructors(t *testing.T) {
	tests := []struct {
		name            string
		err             *domainerrors.DomainError
		wantCode        string
		wantMessage     string
		wantMsgContains []string
	}{
		{
			name:            "when NotFound is built it should set the not_found code and name the resource",
			err:             domainerrors.NotFound("trip", "trip-123"),
			wantCode:        "not_found",
			wantMsgContains: []string{"trip", "trip-123"},
		},
		{
			name:        "when Unauthorized is built it should set the authentication_required code",
			err:         domainerrors.Unauthorized("missing or invalid token"),
			wantCode:    "authentication_required",
			wantMessage: "missing or invalid token",
		},
		{
			name:        "when Forbidden is built it should set the forbidden code",
			err:         domainerrors.Forbidden("insufficient permissions"),
			wantCode:    "forbidden",
			wantMessage: "insufficient permissions",
		},
		{
			name:        "when Conflict is built it should set the conflict code",
			err:         domainerrors.Conflict("version mismatch"),
			wantCode:    "conflict",
			wantMessage: "version mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCode, tt.err.Code)
			if tt.wantMessage != "" {
				assert.Equal(t, tt.wantMessage, tt.err.Message)
			}
			for _, fragment := range tt.wantMsgContains {
				assert.Contains(t, tt.err.Message, fragment)
			}
		})
	}
}

func TestValidation(t *testing.T) {
	t.Run("when field errors are provided it should set the code, message, and fields", func(t *testing.T) {
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
	})

	t.Run("when no field errors are provided it should leave fields empty", func(t *testing.T) {
		err := domainerrors.Validation("One or more fields failed validation")

		assert.Empty(t, err.Fields)
	})
}

func TestServiceUnavailable_SetsCodeMessageAndRetryAfterDetail(t *testing.T) {
	err := domainerrors.ServiceUnavailable(30)

	assert.Equal(t, "service_unavailable", err.Code)
	assert.NotEmpty(t, err.Message)
	require.NotNil(t, err.Details)
	assert.Equal(t, 30, err.Details["retry_after_seconds"])
}

func TestDomainError_Error(t *testing.T) {
	t.Run("when the error has no cause it should return the code and message", func(t *testing.T) {
		err := domainerrors.NotFound("trip", "trip-123")

		assert.Contains(t, err.Error(), "not_found")
		assert.Contains(t, err.Error(), err.Message)
	})

	t.Run("when the error wraps a cause it should include the cause text", func(t *testing.T) {
		cause := errors.New("connection refused")
		err := &domainerrors.DomainError{
			Code:    "internal_error",
			Message: "database unavailable",
			Err:     cause,
		}

		assert.Contains(t, err.Error(), "connection refused")
	})
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
