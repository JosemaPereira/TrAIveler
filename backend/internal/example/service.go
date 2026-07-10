package example

import (
	"context"
	goerrors "errors"

	"github.com/google/uuid"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// CreateInput carries the fields a caller supplies when creating an
// Example; Status/Count/Version/timestamps are business-owned defaults set
// by the service, not accepted from the client.
type CreateInput struct {
	Name  string
	Email string
}

// UpdateInput carries the fields a caller supplies for a full-replacement
// update (PUT semantics, docs/api-design-standards.md §4). Email is
// deliberately not included: changing a unique identifying field is treated
// as a separate concern from a routine field update in this reference
// pattern, matching how many real systems gate email changes behind a
// dedicated, re-verified flow.
type UpdateInput struct {
	Name   string
	Status string
	Count  int
}

// ListOutput is a fully-normalized page of examples ready for the handler
// to serialize as the standard data+pagination envelope
// (docs/api-design-standards.md §6/§9) with no further computation needed.
type ListOutput struct {
	Items      []*Example
	Page       int
	PerPage    int
	Total      int
	TotalPages int
}

// Service contains Example's business logic (validation, uniqueness,
// optimistic-locking orchestration), delegating all persistence to a
// Repository. It intentionally has no knowledge of HTTP or SQL.
type Service interface {
	CreateExample(ctx context.Context, input CreateInput) (*Example, error)
	GetExample(ctx context.Context, id string) (*Example, error)
	UpdateExample(ctx context.Context, id string, expectedVersion int, input UpdateInput) (*Example, error)
	DeleteExample(ctx context.Context, id string) error
	ListExamples(ctx context.Context, filters ListFilters) (*ListOutput, error)
}

type service struct {
	repo Repository
}

// NewService builds a Service backed by repo.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// CreateExample validates input, rejects an already-registered email, then
// generates the entity's ID (google/uuid, rather than relying on the
// database column default — see repository.go's Create doc comment) and
// persists it as StatusActive with Count 0.
func (s *service) CreateExample(ctx context.Context, input CreateInput) (*Example, error) {
	ex := &Example{
		ID:     uuid.New().String(),
		Name:   input.Name,
		Email:  input.Email,
		Status: StatusActive,
		Count:  0,
	}

	if verr := ex.Validate(); verr != nil {
		return nil, verr
	}

	existing, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil && !isNotFound(err) {
		return nil, err
	}
	if existing != nil {
		return nil, domainerrors.Conflict("an example with this email already exists")
	}

	if err := s.repo.Create(ctx, ex); err != nil {
		return nil, err
	}

	return ex, nil
}

// GetExample returns the example with the given ID, or the repository's
// domain NotFound error unchanged.
func (s *service) GetExample(ctx context.Context, id string) (*Example, error) {
	return s.repo.FindByID(ctx, id)
}

// UpdateExample loads the current row, overlays input's fields, attaches
// the caller's expectedVersion for the repository's optimistic-locking
// check, validates the merged result, then persists it. A stale
// expectedVersion surfaces as the repository's domain Conflict.
func (s *service) UpdateExample(ctx context.Context, id string, expectedVersion int, input UpdateInput) (*Example, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing.Name = input.Name
	existing.Status = input.Status
	existing.Count = input.Count
	existing.Version = expectedVersion

	if verr := existing.Validate(); verr != nil {
		return nil, verr
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// DeleteExample removes the example with the given ID, or propagates the
// repository's domain NotFound error unchanged.
func (s *service) DeleteExample(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ListExamples normalizes pagination (defaults/bounds per
// docs/api-design-standards.md §9), delegates to the repository, and
// computes total_pages so the handler can serialize the response envelope
// without reimplementing pagination math.
func (s *service) ListExamples(ctx context.Context, filters ListFilters) (*ListOutput, error) {
	page, perPage, _, _ := normalizePagination(filters.Page, filters.PerPage)
	filters.Page, filters.PerPage = page, perPage

	result, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		Items:      result.Items,
		Page:       page,
		PerPage:    perPage,
		Total:      result.Total,
		TotalPages: totalPages(result.Total, perPage),
	}, nil
}

// isNotFound reports whether err is a domain NotFound error, letting
// CreateExample distinguish "email is available" (NotFound) from an
// unexpected repository failure (which must propagate, not be silently
// treated as availability).
func isNotFound(err error) bool {
	var domainErr *domainerrors.DomainError
	if goerrors.As(err, &domainErr) {
		return domainErr.Code == "not_found"
	}
	return false
}
