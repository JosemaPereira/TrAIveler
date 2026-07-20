package subscription

import (
	"time"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// Status values accepted by Subscription.Status. These mirror the CHECK
// constraint on the subscriptions table (see
// migrations/007_create_subscriptions_table.sql) so callers don't hardcode
// the wire strings.
const (
	// StatusStubPending is the initial state after registration with a payment
	// token, before the (stub) payment provider confirms it.
	StatusStubPending = "stub_pending"
	// StatusActive means the subscription currently grants paid benefits.
	StatusActive = "active"
	// StatusCancelled means the user cancelled but is still within the 30-day
	// grace period (owned trips remain visible but read-only).
	StatusCancelled = "cancelled"
	// StatusExpired means the grace period has elapsed; owned trips are archived.
	StatusExpired = "expired"
)

// Subscription is the billing relationship between a User and a Plan.
//
// The field set intentionally matches the live subscriptions table
// (migrations/007_create_subscriptions_table.sql), which is the source of
// truth: that migration deliberately omitted current_period_start/end and a
// version column ("in no design doc" at the time). The API contract's
// register response (specs/008-auth-collaboration-ux/contracts/api.md) does
// reference current_period_start/end; surfacing those requires a follow-up
// schema migration and is tracked separately (see docs/roadmap.md 008-T039
// Notes) rather than modeled here as columns that don't exist.
//
// db tags drive repository row scanning; json tags drive the API wire format.
// stub_payment_ref is server-internal and never serialized (json:"-").
type Subscription struct {
	ID                string     `json:"id" db:"id"`
	UserID            string     `json:"user_id" db:"user_id"`
	PlanID            string     `json:"plan_id" db:"plan_id"`
	Status            string     `json:"status" db:"status"`
	StubPaymentRef    *string    `json:"-" db:"stub_payment_ref"`
	GracePeriodEndsAt *time.Time `json:"grace_period_ends_at,omitempty" db:"grace_period_ends_at"`
	CancelledAt       *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// IsActive reports whether the subscription currently grants paid benefits at
// time now: an 'active' subscription, or a 'cancelled' one still inside its
// 30-day grace period. 'stub_pending' and 'expired' always return false.
// This is the single source of truth for the has_subscription JWT claim
// (consumed via the SubscriptionResolver adapter in repository.go).
func (s *Subscription) IsActive(now time.Time) bool {
	switch s.Status {
	case StatusActive:
		return true
	case StatusCancelled:
		return s.GracePeriodEndsAt != nil && now.Before(*s.GracePeriodEndsAt)
	default:
		return false
	}
}

// Validate performs business-rule validation, collecting every failing field
// rather than stopping at the first, so a caller can fix all issues in one
// round trip. Returns nil when s is valid.
func (s *Subscription) Validate() *domainerrors.DomainError {
	var fields []domainerrors.ValidationError

	if s.UserID == "" {
		fields = append(fields, domainerrors.ValidationError{
			Field: "user_id",
			Error: "User ID is required",
		})
	}

	if s.PlanID == "" {
		fields = append(fields, domainerrors.ValidationError{
			Field: "plan_id",
			Error: "Plan ID is required",
		})
	}

	if !isValidStatus(s.Status) {
		fields = append(fields, domainerrors.ValidationError{
			Field: "status",
			Error: "Status must be one of 'stub_pending', 'active', 'cancelled', 'expired'",
		})
	}

	if len(fields) > 0 {
		return domainerrors.Validation("One or more fields failed validation", fields...)
	}

	return nil
}

// isValidStatus reports whether status is one of the allowed subscription
// states, matching the subscriptions table CHECK constraint.
func isValidStatus(status string) bool {
	switch status {
	case StatusStubPending, StatusActive, StatusCancelled, StatusExpired:
		return true
	default:
		return false
	}
}
