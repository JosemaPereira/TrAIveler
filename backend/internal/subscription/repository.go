package subscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	authjwt "github.com/JosemaPereira/TrAIveler/backend/internal/auth/jwt"
	"github.com/JosemaPereira/TrAIveler/backend/internal/database"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// gracePeriodDays is the fixed 30-day window between cancellation and
// expiry, during which owned trips stay visible but read-only (BR-001,
// specs/008-auth-collaboration-ux/data-model.md).
const gracePeriodDays = 30

// uniqueViolationCode is the Postgres error code for a UNIQUE constraint
// violation (23505), used to translate a racing duplicate subscription
// (idx_subscriptions_user_id is UNIQUE) into a domain Conflict.
const uniqueViolationCode = "23505"

// Repository defines data access for Subscription. PostgresRepository is the
// only implementation; tests use the generated mock (mocks/repository_mock.go).
//
// Method names are the idiomatic short forms (Create/GetByUserID/Update/
// Cancel) rather than tasks.md's literal CreateSubscription/... to avoid
// package-name stutter (subscription.Create reads cleanly); see the example
// package for the same convention.
type Repository interface {
	Create(ctx context.Context, sub *Subscription) error
	GetByUserID(ctx context.Context, userID string) (*Subscription, error)
	Update(ctx context.Context, sub *Subscription) error
	Cancel(ctx context.Context, id string) error
}

// PostgresRepository implements Repository using the shared pgx pool exposed
// by database.Client.
type PostgresRepository struct {
	db database.Client
}

// NewPostgresRepository builds a Repository backed by PostgreSQL.
func NewPostgresRepository(db database.Client) Repository {
	return &PostgresRepository{db: db}
}

// Create inserts sub, which must already have ID/UserID/PlanID/Status
// populated by the caller (the service layer generates the ID via
// google/uuid rather than relying on the table's gen_random_uuid() default,
// so the ID is known before commit). Populates CreatedAt from the database
// default. A racing duplicate for the same user (one subscription per user)
// is reported as a domain Conflict rather than a raw pgconn error.
func (r *PostgresRepository) Create(ctx context.Context, sub *Subscription) error {
	const query = `
		INSERT INTO subscriptions (id, user_id, plan_id, status, stub_payment_ref)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`

	err := r.db.Pool().QueryRow(ctx, query, sub.ID, sub.UserID, sub.PlanID, sub.Status, sub.StubPaymentRef).
		Scan(&sub.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return domainerrors.Conflict(fmt.Sprintf("a subscription for user %q already exists", sub.UserID))
		}
		return fmt.Errorf("create subscription: %w", err)
	}

	return nil
}

// GetByUserID returns the subscription owned by userID, or a domain NotFound
// error if the user has none. Callers treat NotFound as "no subscription"
// (e.g. the SubscriptionResolver maps it to has_subscription=false).
func (r *PostgresRepository) GetByUserID(ctx context.Context, userID string) (*Subscription, error) {
	const query = `
		SELECT id, user_id, plan_id, status, stub_payment_ref, grace_period_ends_at, cancelled_at, created_at
		FROM subscriptions
		WHERE user_id = $1`

	var sub Subscription
	err := r.db.Pool().QueryRow(ctx, query, userID).Scan(
		&sub.ID, &sub.UserID, &sub.PlanID, &sub.Status, &sub.StubPaymentRef,
		&sub.GracePeriodEndsAt, &sub.CancelledAt, &sub.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NotFound("subscription", userID)
		}
		return nil, fmt.Errorf("find subscription: %w", err)
	}

	return &sub, nil
}

// Update writes the mutable subscription fields (status and the
// cancellation/grace-period timestamps) for the row identified by sub.ID.
// The subscriptions table carries no version column, so this is a plain
// update, not an optimistic-locking one. Updating an ID that doesn't exist
// is reported as a domain NotFound.
func (r *PostgresRepository) Update(ctx context.Context, sub *Subscription) error {
	const query = `
		UPDATE subscriptions
		SET status = $1, stub_payment_ref = $2, grace_period_ends_at = $3, cancelled_at = $4
		WHERE id = $5`

	tag, err := r.db.Pool().Exec(ctx, query,
		sub.Status, sub.StubPaymentRef, sub.GracePeriodEndsAt, sub.CancelledAt, sub.ID,
	)
	if err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.NotFound("subscription", sub.ID)
	}

	return nil
}

// Cancel transitions an active subscription to 'cancelled', stamping
// cancelled_at=NOW() and grace_period_ends_at=NOW()+30 days (BR-001). The
// timestamps are computed in SQL so the database clock is authoritative.
// It only affects a currently-active row; if no active subscription with
// that ID exists (already cancelled/expired, or unknown ID) it returns a
// domain NotFound rather than silently resetting an existing grace period.
func (r *PostgresRepository) Cancel(ctx context.Context, id string) error {
	const query = `
		UPDATE subscriptions
		SET status = $1,
		    cancelled_at = NOW(),
		    grace_period_ends_at = NOW() + make_interval(days => $2)
		WHERE id = $3 AND status = $4`

	tag, err := r.db.Pool().Exec(ctx, query, StatusCancelled, gracePeriodDays, id, StatusActive)
	if err != nil {
		return fmt.Errorf("cancel subscription: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.NotFound("active subscription", id)
	}

	return nil
}

// Resolver adapts a Repository to jwt.SubscriptionResolver so each token
// refresh (jwt.Refresher) can stamp a fresh has_subscription claim from the
// live subscription row rather than a stale one (FR-003/FR-022).
type Resolver struct {
	repo Repository
}

// compile-time assurance that Resolver satisfies the jwt contract.
var _ authjwt.SubscriptionResolver = (*Resolver)(nil)

// NewResolver builds a Resolver over repo.
func NewResolver(repo Repository) *Resolver {
	return &Resolver{repo: repo}
}

// HasActiveSubscription reports whether userID currently has a subscription
// granting paid benefits (active, or cancelled-within-grace, evaluated at the
// current wall-clock time via Subscription.IsActive). A user with no
// subscription row is not an error: it resolves to false, has_subscription=false.
func (r *Resolver) HasActiveSubscription(ctx context.Context, userID uuid.UUID) (bool, error) {
	sub, err := r.repo.GetByUserID(ctx, userID.String())
	if err != nil {
		var domainErr *domainerrors.DomainError
		if errors.As(err, &domainErr) && domainErr.Code == "not_found" {
			return false, nil
		}
		return false, err
	}

	return sub.IsActive(time.Now()), nil
}
