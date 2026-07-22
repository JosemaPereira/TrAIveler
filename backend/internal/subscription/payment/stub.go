package payment

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

// stubReferencePrefix marks payment references minted by the stub so they are
// obviously non-production when they appear on a persisted subscription.
const stubReferencePrefix = "stub_"

// StubPaymentProvider is a no-op PaymentProvider that always succeeds and emits
// a visible "[DEMO]" log line, so the registration flow is demonstrable
// end-to-end before a real gateway is integrated (docs/product-vision.md:
// "visible payment stub"). It never contacts an external system.
type StubPaymentProvider struct {
	logger *slog.Logger
}

// compile-time assurance that StubPaymentProvider satisfies PaymentProvider.
var _ PaymentProvider = (*StubPaymentProvider)(nil)

// NewStubPaymentProvider returns a PaymentProvider backed by the stub (interface
// return type so callers wire against the seam). A nil logger falls back to
// slog.Default().
func NewStubPaymentProvider(logger *slog.Logger) PaymentProvider {
	if logger == nil {
		logger = slog.Default()
	}
	return &StubPaymentProvider{logger: logger}
}

// ProcessPayment always succeeds, logging the demo charge and returning a fresh
// stub payment reference to persist. The token is echoed into the log so the
// demo charge is visible; a real provider must never log payment credentials.
func (p *StubPaymentProvider) ProcessPayment(ctx context.Context, token, planID string) (string, error) {
	p.logger.InfoContext(ctx, fmt.Sprintf("[DEMO] Payment processed: %s", token),
		slog.String("plan_id", planID),
	)
	return stubReferencePrefix + uuid.NewString(), nil
}
