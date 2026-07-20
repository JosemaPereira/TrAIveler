//go:build test

package subscription_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/subscription"
	submocks "github.com/JosemaPereira/TrAIveler/backend/internal/subscription/mocks"
)

func TestUnitHasActiveSubscription_ActiveRow_ReturnsTrue(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := submocks.NewMockRepository(t)
	repo.EXPECT().GetByUserID(ctx, userID.String()).
		Return(&subscription.Subscription{Status: subscription.StatusActive}, nil).Once()

	got, err := subscription.NewResolver(repo).HasActiveSubscription(ctx, userID)

	require.NoError(t, err)
	assert.True(t, got)
}

func TestUnitHasActiveSubscription_GracePeriodRow_ReturnsTrue(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	graceEnd := time.Now().Add(time.Hour)
	repo := submocks.NewMockRepository(t)
	repo.EXPECT().GetByUserID(ctx, userID.String()).
		Return(&subscription.Subscription{
			Status:            subscription.StatusCancelled,
			GracePeriodEndsAt: &graceEnd,
		}, nil).Once()

	got, err := subscription.NewResolver(repo).HasActiveSubscription(ctx, userID)

	require.NoError(t, err)
	assert.True(t, got)
}

func TestUnitHasActiveSubscription_ExpiredGraceRow_ReturnsFalse(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	graceEnd := time.Now().Add(-time.Hour)
	repo := submocks.NewMockRepository(t)
	repo.EXPECT().GetByUserID(ctx, userID.String()).
		Return(&subscription.Subscription{
			Status:            subscription.StatusCancelled,
			GracePeriodEndsAt: &graceEnd,
		}, nil).Once()

	got, err := subscription.NewResolver(repo).HasActiveSubscription(ctx, userID)

	require.NoError(t, err)
	assert.False(t, got)
}

func TestUnitHasActiveSubscription_NoSubscription_ReturnsFalseNoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := submocks.NewMockRepository(t)
	repo.EXPECT().GetByUserID(ctx, userID.String()).
		Return(nil, domainerrors.NotFound("subscription", userID.String())).Once()

	got, err := subscription.NewResolver(repo).HasActiveSubscription(ctx, userID)

	require.NoError(t, err)
	assert.False(t, got)
}

func TestUnitHasActiveSubscription_RepositoryError_PropagatesError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	wantErr := errors.New("connection refused")
	repo := submocks.NewMockRepository(t)
	repo.EXPECT().GetByUserID(ctx, userID.String()).Return(nil, wantErr).Once()

	got, err := subscription.NewResolver(repo).HasActiveSubscription(ctx, userID)

	require.ErrorIs(t, err, wantErr)
	assert.False(t, got)
}
