package payment

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// discardLogger builds a slog.Logger whose output is thrown away, for cases
// that do not assert on log content.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestUnitStubPaymentProviderProcessPayment(t *testing.T) {
	ctx := context.Background()

	t.Run("when charging a valid payment token", func(t *testing.T) {
		t.Run("should return a non-empty reference and no error", func(t *testing.T) {
			// Arrange
			provider := NewStubPaymentProvider(discardLogger())

			// Act
			ref, err := provider.ProcessPayment(ctx, "tok_visa", "plan_basic")

			// Assert
			require.NoError(t, err)
			assert.NotEmpty(t, ref)
		})

		t.Run("should log the demo payment line including the token", func(t *testing.T) {
			// Arrange
			var buf bytes.Buffer
			provider := NewStubPaymentProvider(slog.New(slog.NewJSONHandler(&buf, nil)))

			// Act
			_, err := provider.ProcessPayment(ctx, "tok_visa", "plan_basic")

			// Assert
			require.NoError(t, err)
			assert.Contains(t, buf.String(), "[DEMO] Payment processed: tok_visa")
		})
	})

	t.Run("having a nil logger", func(t *testing.T) {
		t.Run("should fall back to the default logger without panicking", func(t *testing.T) {
			// Arrange
			provider := NewStubPaymentProvider(nil)

			// Act
			ref, err := provider.ProcessPayment(ctx, "tok_amex", "plan_basic")

			// Assert
			require.NoError(t, err)
			assert.NotEmpty(t, ref)
		})
	})

	t.Run("having multiple charges", func(t *testing.T) {
		t.Run("should return a unique reference each time", func(t *testing.T) {
			// Arrange
			provider := NewStubPaymentProvider(discardLogger())

			// Act
			ref1, err1 := provider.ProcessPayment(ctx, "tok_a", "plan_basic")
			ref2, err2 := provider.ProcessPayment(ctx, "tok_b", "plan_basic")

			// Assert
			require.NoError(t, err1)
			require.NoError(t, err2)
			assert.NotEqual(t, ref1, ref2, "each stub charge should yield a distinct reference")
		})
	})
}
