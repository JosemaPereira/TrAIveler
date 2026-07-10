//go:build test

package example_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/example"
	examplemocks "github.com/JosemaPereira/TrAIveler/backend/internal/example/mocks"
)

// TestUnitCreateExample_ValidInput_ReturnsCreatedExample verifies the happy
// path: email is free (repo.FindByEmail returns NotFound), so the service
// generates an ID, defaults Status to active, and persists via repo.Create.
func TestUnitCreateExample_ValidInput_ReturnsCreatedExample(t *testing.T) {
	repo := examplemocks.NewMockRepository(t)
	repo.EXPECT().
		FindByEmail(mock.Anything, "ada@example.com").
		Return(nil, domainerrors.NotFound("example", "ada@example.com")).
		Once()
	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(ex *example.Example) bool {
			return ex.ID != "" && ex.Name == "Ada Lovelace" && ex.Email == "ada@example.com" &&
				ex.Status == example.StatusActive && ex.Count == 0
		})).
		Return(nil).
		Once()

	svc := example.NewService(repo)
	created, err := svc.CreateExample(context.Background(), example.CreateInput{
		Name:  "Ada Lovelace",
		Email: "ada@example.com",
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, example.StatusActive, created.Status)
}

// TestUnitCreateExample_DuplicateEmail_ReturnsConflict verifies that an
// email already present (repo.FindByEmail succeeds) is rejected before
// repo.Create is ever called.
func TestUnitCreateExample_DuplicateEmail_ReturnsConflict(t *testing.T) {
	repo := examplemocks.NewMockRepository(t)
	repo.EXPECT().
		FindByEmail(mock.Anything, "ada@example.com").
		Return(&example.Example{ID: "existing-id", Email: "ada@example.com"}, nil).
		Once()

	svc := example.NewService(repo)
	created, err := svc.CreateExample(context.Background(), example.CreateInput{
		Name:  "Ada Lovelace",
		Email: "ada@example.com",
	})

	require.Error(t, err)
	assert.Nil(t, created)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "conflict", domainErr.Code)
}

// TestUnitCreateExample_MissingName_ReturnsValidationErrorWithoutTouchingRepo
// verifies that model validation runs before any repository call.
func TestUnitCreateExample_MissingName_ReturnsValidationErrorWithoutTouchingRepo(t *testing.T) {
	repo := examplemocks.NewMockRepository(t) // no .EXPECT() calls: repo must not be touched

	svc := example.NewService(repo)
	created, err := svc.CreateExample(context.Background(), example.CreateInput{
		Name:  "",
		Email: "ada@example.com",
	})

	require.Error(t, err)
	assert.Nil(t, created)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "validation_failed", domainErr.Code)
}

// TestUnitCreateExample_RepositoryLookupFails_PropagatesUnexpectedError
// verifies that a non-NotFound error from FindByEmail (an actual DB
// failure) is propagated rather than silently treated as "email available".
func TestUnitCreateExample_RepositoryLookupFails_PropagatesUnexpectedError(t *testing.T) {
	repo := examplemocks.NewMockRepository(t)
	dbErr := errors.New("connection reset")
	repo.EXPECT().
		FindByEmail(mock.Anything, "ada@example.com").
		Return(nil, dbErr).
		Once()

	svc := example.NewService(repo)
	created, err := svc.CreateExample(context.Background(), example.CreateInput{
		Name:  "Ada Lovelace",
		Email: "ada@example.com",
	})

	require.Error(t, err)
	assert.Nil(t, created)
	assert.ErrorIs(t, err, dbErr)
}

// TestUnitGetExample_ExistingID_ReturnsExample verifies the simple
// pass-through to the repository.
func TestUnitGetExample_ExistingID_ReturnsExample(t *testing.T) {
	repo := examplemocks.NewMockRepository(t)
	want := &example.Example{ID: "abc-123", Name: "Ada Lovelace"}
	repo.EXPECT().FindByID(mock.Anything, "abc-123").Return(want, nil).Once()

	svc := example.NewService(repo)
	got, err := svc.GetExample(context.Background(), "abc-123")

	require.NoError(t, err)
	assert.Same(t, want, got)
}

// TestUnitGetExample_MissingID_PropagatesNotFound verifies the repository's
// domain NotFound error surfaces unchanged to the caller.
func TestUnitGetExample_MissingID_PropagatesNotFound(t *testing.T) {
	repo := examplemocks.NewMockRepository(t)
	repo.EXPECT().
		FindByID(mock.Anything, "missing").
		Return(nil, domainerrors.NotFound("example", "missing")).
		Once()

	svc := example.NewService(repo)
	got, err := svc.GetExample(context.Background(), "missing")

	require.Error(t, err)
	assert.Nil(t, got)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

// TestUnitUpdateExample_ValidInput_AppliesChangesAndPersists verifies the
// happy path: existing row is loaded, fields overwritten, expected version
// attached, then persisted via repo.Update.
func TestUnitUpdateExample_ValidInput_AppliesChangesAndPersists(t *testing.T) {
	repo := examplemocks.NewMockRepository(t)
	existing := &example.Example{
		ID: "abc-123", Name: "Old Name", Email: "ada@example.com",
		Status: example.StatusActive, Count: 1, Version: 3,
	}
	repo.EXPECT().FindByID(mock.Anything, "abc-123").Return(existing, nil).Once()
	repo.EXPECT().
		Update(mock.Anything, mock.MatchedBy(func(ex *example.Example) bool {
			return ex.ID == "abc-123" && ex.Name == "New Name" && ex.Count == 5 && ex.Version == 3
		})).
		Return(nil).
		Once()

	svc := example.NewService(repo)
	updated, err := svc.UpdateExample(context.Background(), "abc-123", 3, example.UpdateInput{
		Name:   "New Name",
		Status: example.StatusActive,
		Count:  5,
	})

	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, 5, updated.Count)
}

// TestUnitUpdateExample_MissingID_PropagatesNotFoundWithoutCallingUpdate
// verifies that a missing row short-circuits before repo.Update is called.
func TestUnitUpdateExample_MissingID_PropagatesNotFoundWithoutCallingUpdate(t *testing.T) {
	repo := examplemocks.NewMockRepository(t)
	repo.EXPECT().
		FindByID(mock.Anything, "missing").
		Return(nil, domainerrors.NotFound("example", "missing")).
		Once()
	// No .EXPECT() for Update: it must not be called.

	svc := example.NewService(repo)
	updated, err := svc.UpdateExample(context.Background(), "missing", 1, example.UpdateInput{
		Name:   "New Name",
		Status: example.StatusActive,
	})

	require.Error(t, err)
	assert.Nil(t, updated)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

// TestUnitUpdateExample_InvalidStatus_ReturnsValidationErrorWithoutCallingUpdate
// verifies that model validation runs on the merged entity before
// repo.Update is called.
func TestUnitUpdateExample_InvalidStatus_ReturnsValidationErrorWithoutCallingUpdate(t *testing.T) {
	repo := examplemocks.NewMockRepository(t)
	existing := &example.Example{ID: "abc-123", Name: "Old Name", Email: "ada@example.com", Status: example.StatusActive, Version: 1}
	repo.EXPECT().FindByID(mock.Anything, "abc-123").Return(existing, nil).Once()
	// No .EXPECT() for Update: it must not be called.

	svc := example.NewService(repo)
	updated, err := svc.UpdateExample(context.Background(), "abc-123", 1, example.UpdateInput{
		Name:   "New Name",
		Status: "bogus",
	})

	require.Error(t, err)
	assert.Nil(t, updated)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "validation_failed", domainErr.Code)
}

// TestUnitUpdateExample_VersionConflict_PropagatesConflict verifies that a
// repo-level optimistic-locking conflict surfaces unchanged.
func TestUnitUpdateExample_VersionConflict_PropagatesConflict(t *testing.T) {
	repo := examplemocks.NewMockRepository(t)
	existing := &example.Example{ID: "abc-123", Name: "Old Name", Email: "ada@example.com", Status: example.StatusActive, Version: 3}
	repo.EXPECT().FindByID(mock.Anything, "abc-123").Return(existing, nil).Once()
	repo.EXPECT().
		Update(mock.Anything, mock.Anything).
		Return(domainerrors.Conflict("example was modified by another request")).
		Once()

	svc := example.NewService(repo)
	updated, err := svc.UpdateExample(context.Background(), "abc-123", 3, example.UpdateInput{
		Name:   "New Name",
		Status: example.StatusActive,
	})

	require.Error(t, err)
	assert.Nil(t, updated)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "conflict", domainErr.Code)
}

// TestUnitDeleteExample_ExistingID_Succeeds verifies the pass-through to
// the repository.
func TestUnitDeleteExample_ExistingID_Succeeds(t *testing.T) {
	repo := examplemocks.NewMockRepository(t)
	repo.EXPECT().Delete(mock.Anything, "abc-123").Return(nil).Once()

	svc := example.NewService(repo)
	err := svc.DeleteExample(context.Background(), "abc-123")

	assert.NoError(t, err)
}

// TestUnitDeleteExample_MissingID_PropagatesNotFound verifies the
// repository's domain NotFound error surfaces unchanged.
func TestUnitDeleteExample_MissingID_PropagatesNotFound(t *testing.T) {
	repo := examplemocks.NewMockRepository(t)
	repo.EXPECT().
		Delete(mock.Anything, "missing").
		Return(domainerrors.NotFound("example", "missing")).
		Once()

	svc := example.NewService(repo)
	err := svc.DeleteExample(context.Background(), "missing")

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

// TestUnitListExamples_NormalizesPaginationAndComputesTotalPages verifies
// that the service normalizes raw page/per_page before delegating to the
// repository, and computes total_pages for the caller (handler) to build
// the standard list envelope (docs/api-design-standards.md §9).
func TestUnitListExamples_NormalizesPaginationAndComputesTotalPages(t *testing.T) {
	repo := examplemocks.NewMockRepository(t)
	items := []*example.Example{{ID: "1"}, {ID: "2"}}
	repo.EXPECT().
		List(mock.Anything, mock.MatchedBy(func(f example.ListFilters) bool {
			return f.Page == 1 && f.PerPage == 20
		})).
		Return(&example.ListResult{Items: items, Total: 41}, nil).
		Once()

	svc := example.NewService(repo)
	out, err := svc.ListExamples(context.Background(), example.ListFilters{Page: 0, PerPage: 0})

	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, 1, out.Page)
	assert.Equal(t, 20, out.PerPage)
	assert.Equal(t, 41, out.Total)
	assert.Equal(t, 3, out.TotalPages)
	assert.Len(t, out.Items, 2)
}
