package trip

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JosemaPereira/TrAIveler/backend/internal/auth"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/subscription"
)

// maxTitleLength bounds Trip.Title per docs/data-model.md §Trip's [API] rule
// (1-200 characters).
const maxTitleLength = 200

// SubscriptionLookup is the persistence seam Update/Delete use to resolve
// the acting user's current subscription status. subscription.Repository
// satisfies this via its GetByUserID method.
type SubscriptionLookup interface {
	GetByUserID(ctx context.Context, userID string) (*subscription.Subscription, error)
}

// Service contains Trip business logic (001-T035): admin-only creation,
// ownership authorization (with a deliberate 404-vs-403 split between Get
// and Update/Delete — see each method's doc comment), forward-only status
// transitions, and subscription-gated writes. It has no knowledge of HTTP or
// SQL, delegating persistence to a Repository and subscription status to a
// SubscriptionLookup.
type Service struct {
	repo Repository
	subs SubscriptionLookup
}

// NewService builds a Service backed by repo and subs.
func NewService(repo Repository, subs SubscriptionLookup) *Service {
	return &Service{repo: repo, subs: subs}
}

// Create persists a new draft Trip owned by userID. Only an admin (role ==
// auth.RoleAdmin) may create a trip (docs/data-model.md §Trip's [Logic]
// rule); any other role is rejected with a domain Forbidden error before
// title validation runs.
func (s *Service) Create(ctx context.Context, userID, role, title string, description *string) (*Trip, error) {
	if role != auth.RoleAdmin {
		return nil, domainerrors.Forbidden("Only admin users can create trips")
	}
	if err := validateTitle(title); err != nil {
		return nil, err
	}

	tr := &Trip{
		ID:          uuid.NewString(),
		CreatorID:   userID,
		Title:       title,
		Description: description,
		Status:      TripStatusDraft,
	}
	if err := s.repo.CreateTrip(ctx, tr); err != nil {
		return nil, err
	}

	return tr, nil
}

// List returns every trip owned by userID.
func (s *Service) List(ctx context.Context, userID string) ([]*Trip, error) {
	return s.repo.ListTripsByUser(ctx, userID)
}

// Get returns the trip identified by tripID, provided userID owns it.
//
// A trip that exists but is owned by someone else is reported as the same
// domain NotFound a nonexistent trip would produce — deliberately not a
// Forbidden — so a non-owner cannot distinguish "doesn't exist" from
// "exists but isn't yours" via a 403-vs-404 split (the anti-ID-enumeration
// rule carried forward from superseded roadmap row 008-T105; see
// 001-T038's Notes in docs/roadmap.md). This is intentionally different
// from Update/Delete's ownership check, which does return Forbidden — see
// their doc comments.
func (s *Service) Get(ctx context.Context, userID, tripID string) (*Trip, error) {
	tr, err := s.repo.FindTripByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if tr.CreatorID != userID {
		return nil, domainerrors.NotFound("trip", tripID)
	}
	return tr, nil
}

// Update applies title/description/status to the trip identified by tripID
// using optimistic locking on expectedVersion, provided userID is the
// trip's creator and currently holds an active-or-grace-period
// subscription.
//
// Unlike Get, a non-owner is reported as Forbidden, not NotFound — this
// mirrors superseded roadmap row 008-T107's original behavior, which
// remains a deliberate, intentional split from Get's anti-enumeration 404.
//
// The subscription check accepts active OR grace-period subscriptions (via
// Subscription.IsActive), and treats "no subscription row at all" the same
// as "no active subscription" rather than surfacing a confusing
// subscription-not-found error for what should read as a trip-authorization
// failure.
//
// Status transitions are forward-only (docs/data-model.md §Trip "State
// Transitions"): draft->draft, draft->published, and published->published
// are allowed; published->draft is rejected as a validation error.
func (s *Service) Update(
	ctx context.Context,
	userID, tripID string,
	expectedVersion int,
	title string,
	description *string,
	status string,
) (*Trip, error) {
	tr, err := s.repo.FindTripByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if tr.CreatorID != userID {
		return nil, domainerrors.Forbidden("Only the trip creator can update this trip")
	}

	sub, err := s.resolveSubscription(ctx, userID)
	if err != nil {
		return nil, err
	}
	if sub == nil || !sub.IsActive(time.Now()) {
		return nil, domainerrors.Forbidden("An active subscription is required to update this trip")
	}

	if err := validateTitle(title); err != nil {
		return nil, err
	}
	if err := validateStatusTransition(tr.Status, status); err != nil {
		return nil, err
	}

	tr.Title = title
	tr.Description = description
	tr.Status = status
	tr.Version = expectedVersion

	if err := s.repo.UpdateTrip(ctx, tr); err != nil {
		return nil, err
	}

	return tr, nil
}

// Delete removes the trip identified by tripID, provided userID is the
// trip's creator and currently holds a strictly active (StatusActive)
// subscription — stricter than Update, which also accepts the grace period
// (008-T109's original semantics: deletion is not allowed during grace).
func (s *Service) Delete(ctx context.Context, userID, tripID string) error {
	tr, err := s.repo.FindTripByID(ctx, tripID)
	if err != nil {
		return err
	}
	if tr.CreatorID != userID {
		return domainerrors.Forbidden("Only the trip creator can delete this trip")
	}

	sub, err := s.resolveSubscription(ctx, userID)
	if err != nil {
		return err
	}
	if sub == nil || sub.Status != subscription.StatusActive {
		return domainerrors.Forbidden("An active subscription is required to delete this trip")
	}

	return s.repo.DeleteTrip(ctx, tripID)
}

// resolveSubscription returns userID's subscription, or nil if the user has
// none. A domain NotFound from the lookup is treated as "no subscription"
// rather than propagated, so Update/Delete can fold it into their own
// Forbidden response instead of surfacing a confusing
// subscription-not-found error; any other error propagates as-is.
func (s *Service) resolveSubscription(ctx context.Context, userID string) (*subscription.Subscription, error) {
	sub, err := s.subs.GetByUserID(ctx, userID)
	if err != nil {
		var domainErr *domainerrors.DomainError
		if errors.As(err, &domainErr) && domainErr.Code == "not_found" {
			return nil, nil
		}
		return nil, err
	}
	return sub, nil
}

// validateTitle enforces docs/data-model.md §Trip's [API] rule: title must
// be 1-200 characters after trimming surrounding whitespace.
func validateTitle(title string) *domainerrors.DomainError {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return domainerrors.Validation("One or more fields failed validation", domainerrors.ValidationError{
			Field: "title",
			Error: "Title is required",
		})
	}
	if len([]rune(trimmed)) > maxTitleLength {
		return domainerrors.Validation("One or more fields failed validation", domainerrors.ValidationError{
			Field: "title",
			Error: "Title must be at most 200 characters",
		})
	}
	return nil
}

// validateStatusTransition enforces the forward-only Trip status transition
// (docs/data-model.md §Trip "State Transitions"): draft->draft,
// draft->published, and published->published are allowed; published->draft
// is rejected.
func validateStatusTransition(current, next string) *domainerrors.DomainError {
	if next != TripStatusDraft && next != TripStatusPublished {
		return domainerrors.Validation("One or more fields failed validation", domainerrors.ValidationError{
			Field: "status",
			Error: "Status must be 'draft' or 'published'",
		})
	}
	if current == TripStatusPublished && next == TripStatusDraft {
		return domainerrors.Validation("One or more fields failed validation", domainerrors.ValidationError{
			Field: "status",
			Error: "A published trip cannot be moved back to draft",
		})
	}
	return nil
}
