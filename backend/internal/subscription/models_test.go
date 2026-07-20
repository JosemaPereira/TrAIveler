package subscription

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitValidate_ValidSubscription_ReturnsNil(t *testing.T) {
	sub := &Subscription{
		UserID: "11111111-1111-1111-1111-111111111111",
		PlanID: "00000000-0000-0000-0000-000000000001",
		Status: StatusActive,
	}

	err := sub.Validate()

	assert.Nil(t, err)
}

func TestUnitValidate_MissingUserID_ReturnsValidationError(t *testing.T) {
	sub := &Subscription{
		UserID: "",
		PlanID: "00000000-0000-0000-0000-000000000001",
		Status: StatusActive,
	}

	err := sub.Validate()

	require.NotNil(t, err)
	assert.Equal(t, "validation_failed", err.Code)
	require.Len(t, err.Fields, 1)
	assert.Equal(t, "user_id", err.Fields[0].Field)
}

func TestUnitValidate_MissingPlanID_ReturnsValidationError(t *testing.T) {
	sub := &Subscription{
		UserID: "11111111-1111-1111-1111-111111111111",
		PlanID: "",
		Status: StatusActive,
	}

	err := sub.Validate()

	require.NotNil(t, err)
	require.Len(t, err.Fields, 1)
	assert.Equal(t, "plan_id", err.Fields[0].Field)
}

func TestUnitValidate_InvalidStatus_ReturnsValidationError(t *testing.T) {
	sub := &Subscription{
		UserID: "11111111-1111-1111-1111-111111111111",
		PlanID: "00000000-0000-0000-0000-000000000001",
		Status: "past_due",
	}

	err := sub.Validate()

	require.NotNil(t, err)
	require.Len(t, err.Fields, 1)
	assert.Equal(t, "status", err.Fields[0].Field)
}

func TestUnitValidate_MultipleFailures_ReturnsAllFieldErrors(t *testing.T) {
	sub := &Subscription{Status: "bogus"}

	err := sub.Validate()

	require.NotNil(t, err)
	assert.Len(t, err.Fields, 3)
}

func TestUnitIsActive_StatusScenarios_ReportsBenefitEligibility(t *testing.T) {
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name string
		sub  Subscription
		want bool
	}{
		{
			name: "when the subscription is active it should grant benefits",
			sub:  Subscription{Status: StatusActive},
			want: true,
		},
		{
			name: "when cancelled but inside the grace period it should still grant benefits",
			sub:  Subscription{Status: StatusCancelled, GracePeriodEndsAt: &future},
			want: true,
		},
		{
			name: "when cancelled and past the grace period it should not grant benefits",
			sub:  Subscription{Status: StatusCancelled, GracePeriodEndsAt: &past},
			want: false,
		},
		{
			name: "when cancelled with no grace period set it should not grant benefits",
			sub:  Subscription{Status: StatusCancelled, GracePeriodEndsAt: nil},
			want: false,
		},
		{
			name: "when the subscription is stub_pending it should not grant benefits",
			sub:  Subscription{Status: StatusStubPending},
			want: false,
		},
		{
			name: "when the subscription is expired it should not grant benefits",
			sub:  Subscription{Status: StatusExpired, GracePeriodEndsAt: &past},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.sub.IsActive(now))
		})
	}
}
