//go:build test

package subscription_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/subscription"
	submocks "github.com/JosemaPereira/TrAIveler/backend/internal/subscription/mocks"
	paymentmocks "github.com/JosemaPereira/TrAIveler/backend/internal/subscription/payment/mocks"
)

const (
	testUserID = "9f1c1a1e-2b3d-4c5e-8f6a-1234567890ab"
	testToken  = "tok_visa_demo"
)

// TestUnitCreateSubscription_PaymentSucceeds_PersistsActiveSubscription verifies
// the happy path: the payment provider is charged first, its opaque reference is
// stamped on a new active subscription, and that subscription is persisted.
func TestUnitCreateSubscription_PaymentSucceeds_PersistsActiveSubscription(t *testing.T) {
	provider := paymentmocks.NewMockPaymentProvider(t)
	repo := submocks.NewMockRepository(t)

	const paymentRef = "stub_ref_123"
	provider.EXPECT().
		ProcessPayment(mock.Anything, testToken, subscription.DefaultPlanID).
		Return(paymentRef, nil).
		Once()
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(sub *subscription.Subscription) bool {
			return sub.ID != "" &&
				sub.UserID == testUserID &&
				sub.PlanID == subscription.DefaultPlanID &&
				sub.Status == subscription.StatusActive &&
				sub.StubPaymentRef != nil && *sub.StubPaymentRef == paymentRef
		})).
		Return(nil).
		Once()

	svc := subscription.NewService(provider, repo)
	sub, err := svc.CreateSubscription(context.Background(), testUserID, subscription.DefaultPlanID, testToken)

	require.NoError(t, err)
	require.NotNil(t, sub)
	assert.NotEmpty(t, sub.ID)
	assert.Equal(t, subscription.StatusActive, sub.Status)
	require.NotNil(t, sub.StubPaymentRef)
	assert.Equal(t, paymentRef, *sub.StubPaymentRef)
}

// TestUnitCreateSubscription_PaymentFails_DoesNotPersist verifies that a failed
// charge short-circuits: the repository is never touched and the payment error
// propagates.
func TestUnitCreateSubscription_PaymentFails_DoesNotPersist(t *testing.T) {
	provider := paymentmocks.NewMockPaymentProvider(t)
	repo := submocks.NewMockRepository(t) // no .EXPECT(): Create must not be called

	paymentErr := errors.New("card declined")
	provider.EXPECT().
		ProcessPayment(mock.Anything, testToken, subscription.DefaultPlanID).
		Return("", paymentErr).
		Once()

	svc := subscription.NewService(provider, repo)
	sub, err := svc.CreateSubscription(context.Background(), testUserID, subscription.DefaultPlanID, testToken)

	require.Error(t, err)
	assert.Nil(t, sub)
	assert.ErrorIs(t, err, paymentErr)
}

// TestUnitCreateSubscription_RepositoryFails_PropagatesError verifies that a
// persistence failure after a successful charge surfaces to the caller.
func TestUnitCreateSubscription_RepositoryFails_PropagatesError(t *testing.T) {
	provider := paymentmocks.NewMockPaymentProvider(t)
	repo := submocks.NewMockRepository(t)

	provider.EXPECT().
		ProcessPayment(mock.Anything, testToken, subscription.DefaultPlanID).
		Return("stub_ref_456", nil).
		Once()
	repoErr := domainerrors.Conflict("a subscription for user already exists")
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(repoErr).Once()

	svc := subscription.NewService(provider, repo)
	sub, err := svc.CreateSubscription(context.Background(), testUserID, subscription.DefaultPlanID, testToken)

	require.Error(t, err)
	assert.Nil(t, sub)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "conflict", domainErr.Code)
}

// TestUnitDefaultPlanID_MatchesSeededBasicPlan pins the DefaultPlanID constant to
// the single seeded MVP 'basic' plan (migration 006), which registration relies on.
func TestUnitDefaultPlanID_MatchesSeededBasicPlan(t *testing.T) {
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", subscription.DefaultPlanID)
}
