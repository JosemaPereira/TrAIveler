// Package payment defines the payment-processing seam used by the
// subscription/registration flow, plus a demo stub implementation. Callers
// depend on the PaymentProvider interface, so the MVP's StubPaymentProvider can
// be swapped for a real gateway (Stripe, etc.) without touching them
// (Dependency-Injection-with-Interface-First; see internal/database.Client).
package payment

import "context"

// PaymentProvider is the dependency-injection seam for charging a payment method
// while creating a subscription. Implementations must be concurrency-safe. It is
// deliberately minimal — the MVP only needs to attempt a charge and record an
// opaque reference; a real gateway can wrap richer semantics behind the same
// contract.
//
//nolint:revive // name mandated by issue #168 (008-T042); the payment.PaymentProvider stutter is deliberate.
type PaymentProvider interface {
	// ProcessPayment charges the tokenized payment method (token) for planID. On
	// success it returns an opaque reference the caller persists on the
	// subscription (Subscription.StubPaymentRef) for later reconciliation; on
	// failure it returns an error and the caller must not create the subscription.
	ProcessPayment(ctx context.Context, token, planID string) (string, error)
}
