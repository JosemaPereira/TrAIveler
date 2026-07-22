package subscription

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/JosemaPereira/TrAIveler/backend/internal/subscription/payment"
)

// DefaultPlanID is the single seeded MVP 'basic' plan
// (migrations/006_create_plans_table.sql). Registration uses it because the MVP
// offers exactly one paid plan; a real multi-plan flow would take the plan from
// the request instead.
const DefaultPlanID = "00000000-0000-0000-0000-000000000001"

// Service contains Subscription business logic: it charges the payment provider
// and, only on success, persists an active subscription. It has no knowledge of
// HTTP or SQL, delegating persistence to a Repository and payment to a
// payment.PaymentProvider.
type Service struct {
	provider payment.PaymentProvider
	repo     Repository
}

// NewService builds a Service backed by provider and repo.
func NewService(provider payment.PaymentProvider, repo Repository) *Service {
	return &Service{provider: provider, repo: repo}
}

// CreateSubscription charges the tokenized payment method for planID and, only
// if the charge succeeds, persists a new active subscription for userID stamped
// with the provider's opaque reference. A failed charge short-circuits: no
// subscription is created and the payment error is returned. The subscription ID
// is generated here (google/uuid) so it is known before persistence, mirroring
// the repository's contract.
//
// Per the deliberate migration-007 deferral (see models.go), no
// current_period_* fields are set — those columns do not exist yet.
func (s *Service) CreateSubscription(ctx context.Context, userID, planID, paymentToken string) (*Subscription, error) {
	ref, err := s.provider.ProcessPayment(ctx, paymentToken, planID)
	if err != nil {
		return nil, fmt.Errorf("process payment: %w", err)
	}

	sub := &Subscription{
		ID:             uuid.NewString(),
		UserID:         userID,
		PlanID:         planID,
		Status:         StatusActive,
		StubPaymentRef: &ref,
	}

	if validationErr := sub.Validate(); validationErr != nil {
		return nil, validationErr
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}
