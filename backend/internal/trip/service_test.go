//go:build test

package trip_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/internal/auth"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/subscription"
	"github.com/JosemaPereira/TrAIveler/backend/internal/trip"
	tripmocks "github.com/JosemaPereira/TrAIveler/backend/internal/trip/mocks"
)

const (
	testUserID  = "9f1c1a1e-2b3d-4c5e-8f6a-1234567890ab"
	testOtherID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	testTripID  = "11111111-2222-3333-4444-555555555555"
)

// fakeSubscriptionLookup hand-fakes trip.SubscriptionLookup, the small
// consumer-owned port `make mocks` would otherwise auto-generate (see
// docs/mock-standards.md and .github/memory/patterns-discovered.md's
// "mockery --all Regenerates a Mock for Every Interface..." entry) — its
// generated mock is deliberately not committed.
type fakeSubscriptionLookup struct {
	sub *subscription.Subscription
	err error
}

func (f *fakeSubscriptionLookup) GetByUserID(_ context.Context, _ string) (*subscription.Subscription, error) {
	return f.sub, f.err
}

// ownedTrip returns a draft Trip owned by testUserID, the baseline fixture
// for Update/Delete tests.
func ownedTrip() *trip.Trip {
	return &trip.Trip{ID: testTripID, CreatorID: testUserID, Status: trip.TripStatusDraft, Version: 3}
}

// --- Create ---

func TestUnitCreate_NonAdminRole_ReturnsForbidden(t *testing.T) {
	repo := tripmocks.NewMockRepository(t) // no .EXPECT(): CreateTrip must not be called
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	got, err := svc.Create(context.Background(), testUserID, auth.RolePartner, "My Trip", nil)

	require.Error(t, err)
	assert.Nil(t, got)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "forbidden", domainErr.Code)
}

func TestUnitCreate_AdminRoleValidTitle_PersistsDraftTrip(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().
		CreateTrip(mock.Anything, mock.MatchedBy(func(tr *trip.Trip) bool {
			return tr.ID != "" &&
				tr.CreatorID == testUserID &&
				tr.Title == "My Trip" &&
				tr.Status == trip.TripStatusDraft
		})).
		Return(nil).
		Once()
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	got, err := svc.Create(context.Background(), testUserID, auth.RoleAdmin, "My Trip", nil)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.NotEmpty(t, got.ID)
	assert.Equal(t, trip.TripStatusDraft, got.Status)
}

func TestUnitCreate_EmptyTitle_ReturnsValidationError(t *testing.T) {
	repo := tripmocks.NewMockRepository(t) // no .EXPECT(): CreateTrip must not be called
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	got, err := svc.Create(context.Background(), testUserID, auth.RoleAdmin, "   ", nil)

	require.Error(t, err)
	assert.Nil(t, got)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "validation_failed", domainErr.Code)
}

func TestUnitCreate_TitleOver200Characters_ReturnsValidationError(t *testing.T) {
	repo := tripmocks.NewMockRepository(t) // no .EXPECT(): CreateTrip must not be called
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})
	longTitle := strings.Repeat("a", 201)

	got, err := svc.Create(context.Background(), testUserID, auth.RoleAdmin, longTitle, nil)

	require.Error(t, err)
	assert.Nil(t, got)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "validation_failed", domainErr.Code)
}

func TestUnitCreate_RepositoryFails_PropagatesError(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	wantErr := errors.New("connection refused")
	repo.EXPECT().CreateTrip(mock.Anything, mock.Anything).Return(wantErr).Once()
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	got, err := svc.Create(context.Background(), testUserID, auth.RoleAdmin, "My Trip", nil)

	require.ErrorIs(t, err, wantErr)
	assert.Nil(t, got)
}

// --- List ---

func TestUnitList_DelegatesToRepository(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	want := []*trip.Trip{{ID: testTripID, CreatorID: testUserID}}
	repo.EXPECT().ListTripsByUser(mock.Anything, testUserID).Return(want, nil).Once()
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	got, err := svc.List(context.Background(), testUserID)

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

// TestUnitList_ArchivedTrips_ExcludedFromResults covers the 008-T103
// carry-over cited by 001-T038: Trip.Archived exists as a column but nothing
// yet sets it true, so this is the first behavioral test of the filter.
// Non-archived order is preserved, matching the repository's own ordering.
func TestUnitList_ArchivedTrips_ExcludedFromResults(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	visible := &trip.Trip{ID: testTripID, CreatorID: testUserID, Archived: false}
	archived := &trip.Trip{ID: testOtherID, CreatorID: testUserID, Archived: true}
	repo.EXPECT().ListTripsByUser(mock.Anything, testUserID).
		Return([]*trip.Trip{visible, archived}, nil).Once()
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	got, err := svc.List(context.Background(), testUserID)

	require.NoError(t, err)
	assert.Equal(t, []*trip.Trip{visible}, got)
}

// --- Get ---

func TestUnitGet_OwnedTrip_ReturnsTrip(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	want := &trip.Trip{ID: testTripID, CreatorID: testUserID}
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(want, nil).Once()
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	got, err := svc.Get(context.Background(), testUserID, testTripID)

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestUnitGet_TripNotFound_PropagatesNotFound(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).
		Return(nil, domainerrors.NotFound("trip", testTripID)).Once()
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	got, err := svc.Get(context.Background(), testUserID, testTripID)

	require.Error(t, err)
	assert.Nil(t, got)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestUnitGet_NotOwner_ReturnsNotFoundNotForbidden(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).
		Return(&trip.Trip{ID: testTripID, CreatorID: testOtherID}, nil).Once()
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	got, err := svc.Get(context.Background(), testUserID, testTripID)

	require.Error(t, err)
	assert.Nil(t, got)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code,
		"a non-owner must see the same 404 a nonexistent trip would produce, never a 403 (anti-enumeration)")
}

// --- Update ---

func TestUnitUpdate_NotOwner_ReturnsForbidden(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).
		Return(&trip.Trip{ID: testTripID, CreatorID: testOtherID}, nil).Once()
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	got, err := svc.Update(context.Background(), testUserID, testTripID, 1, "New Title", nil, trip.TripStatusDraft)

	require.Error(t, err)
	assert.Nil(t, got)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "forbidden", domainErr.Code,
		"unlike Get, a non-owner update attempt must surface 403, not 404")
}

func TestUnitUpdate_TripNotFound_PropagatesNotFound(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).
		Return(nil, domainerrors.NotFound("trip", testTripID)).Once()
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	_, err := svc.Update(context.Background(), testUserID, testTripID, 1, "New Title", nil, trip.TripStatusDraft)

	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestUnitUpdate_NoSubscription_ReturnsForbidden(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(ownedTrip(), nil).Once()
	subs := &fakeSubscriptionLookup{err: domainerrors.NotFound("subscription", testUserID)}
	svc := trip.NewService(repo, subs)

	_, err := svc.Update(context.Background(), testUserID, testTripID, 3, "New Title", nil, trip.TripStatusDraft)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "forbidden", domainErr.Code,
		"a NotFound subscription lookup must fold into Forbidden, not surface as a confusing 404")
}

func TestUnitUpdate_SubscriptionLookupFails_PropagatesError(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(ownedTrip(), nil).Once()
	wantErr := errors.New("connection refused")
	subs := &fakeSubscriptionLookup{err: wantErr}
	svc := trip.NewService(repo, subs)

	_, err := svc.Update(context.Background(), testUserID, testTripID, 3, "New Title", nil, trip.TripStatusDraft)

	require.ErrorIs(t, err, wantErr)
}

func TestUnitUpdate_ExpiredSubscription_ReturnsForbidden(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(ownedTrip(), nil).Once()
	subs := &fakeSubscriptionLookup{sub: &subscription.Subscription{Status: subscription.StatusExpired}}
	svc := trip.NewService(repo, subs)

	_, err := svc.Update(context.Background(), testUserID, testTripID, 3, "New Title", nil, trip.TripStatusDraft)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "forbidden", domainErr.Code)
}

func TestUnitUpdate_GracePeriodSubscription_Succeeds(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(ownedTrip(), nil).Once()
	repo.EXPECT().UpdateTrip(mock.Anything, mock.Anything).Return(nil).Once()
	graceEnd := time.Now().Add(time.Hour)
	subs := &fakeSubscriptionLookup{sub: &subscription.Subscription{
		Status:            subscription.StatusCancelled,
		GracePeriodEndsAt: &graceEnd,
	}}
	svc := trip.NewService(repo, subs)

	got, err := svc.Update(context.Background(), testUserID, testTripID, 3, "New Title", nil, trip.TripStatusDraft)

	require.NoError(t, err, "Update must accept a still-in-grace-period subscription")
	require.NotNil(t, got)
	assert.Equal(t, "New Title", got.Title)
}

func TestUnitUpdate_ActiveSubscriptionValidTransition_UpdatesTrip(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(ownedTrip(), nil).Once()
	repo.EXPECT().
		UpdateTrip(mock.Anything, mock.MatchedBy(func(tr *trip.Trip) bool {
			return tr.ID == testTripID &&
				tr.Title == "New Title" &&
				tr.Status == trip.TripStatusPublished &&
				tr.Version == 3
		})).
		Return(nil).
		Once()
	subs := &fakeSubscriptionLookup{sub: &subscription.Subscription{Status: subscription.StatusActive}}
	svc := trip.NewService(repo, subs)

	got, err := svc.Update(context.Background(), testUserID, testTripID, 3, "New Title", nil, trip.TripStatusPublished)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, trip.TripStatusPublished, got.Status)
}

func TestUnitUpdate_PublishedToDraft_ReturnsValidationError(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	published := ownedTrip()
	published.Status = trip.TripStatusPublished
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(published, nil).Once()
	subs := &fakeSubscriptionLookup{sub: &subscription.Subscription{Status: subscription.StatusActive}}
	svc := trip.NewService(repo, subs) // no UpdateTrip expectation: must not be called

	_, err := svc.Update(context.Background(), testUserID, testTripID, 3, "New Title", nil, trip.TripStatusDraft)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "validation_failed", domainErr.Code,
		"published->draft must be rejected: status transitions are forward-only")
}

func TestUnitUpdate_EmptyTitle_ReturnsValidationError(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(ownedTrip(), nil).Once()
	subs := &fakeSubscriptionLookup{sub: &subscription.Subscription{Status: subscription.StatusActive}}
	svc := trip.NewService(repo, subs) // no UpdateTrip expectation: must not be called

	_, err := svc.Update(context.Background(), testUserID, testTripID, 3, "   ", nil, trip.TripStatusDraft)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "validation_failed", domainErr.Code)
}

func TestUnitUpdate_InvalidStatusValue_ReturnsValidationError(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(ownedTrip(), nil).Once()
	subs := &fakeSubscriptionLookup{sub: &subscription.Subscription{Status: subscription.StatusActive}}
	svc := trip.NewService(repo, subs) // no UpdateTrip expectation: must not be called

	_, err := svc.Update(context.Background(), testUserID, testTripID, 3, "New Title", nil, "archived")

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "validation_failed", domainErr.Code)
}

func TestUnitUpdate_RepositoryConflict_PropagatesConflict(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(ownedTrip(), nil).Once()
	repo.EXPECT().UpdateTrip(mock.Anything, mock.Anything).
		Return(domainerrors.Conflict("trip was modified concurrently")).Once()
	subs := &fakeSubscriptionLookup{sub: &subscription.Subscription{Status: subscription.StatusActive}}
	svc := trip.NewService(repo, subs)

	_, err := svc.Update(context.Background(), testUserID, testTripID, 3, "New Title", nil, trip.TripStatusDraft)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "conflict", domainErr.Code)
}

// --- Delete ---

func TestUnitDelete_NotOwner_ReturnsForbidden(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).
		Return(&trip.Trip{ID: testTripID, CreatorID: testOtherID}, nil).Once()
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	err := svc.Delete(context.Background(), testUserID, testTripID)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "forbidden", domainErr.Code)
}

func TestUnitDelete_TripNotFound_PropagatesNotFound(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).
		Return(nil, domainerrors.NotFound("trip", testTripID)).Once()
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	err := svc.Delete(context.Background(), testUserID, testTripID)

	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

func TestUnitDelete_GracePeriodSubscription_ReturnsForbidden(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(ownedTrip(), nil).Once()
	graceEnd := time.Now().Add(time.Hour)
	subs := &fakeSubscriptionLookup{sub: &subscription.Subscription{
		Status:            subscription.StatusCancelled,
		GracePeriodEndsAt: &graceEnd,
	}}
	svc := trip.NewService(repo, subs) // no DeleteTrip expectation: must not be called

	err := svc.Delete(context.Background(), testUserID, testTripID)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "forbidden", domainErr.Code,
		"Delete requires a strictly active subscription; unlike Update, the grace period is not enough")
}

func TestUnitDelete_NoSubscription_ReturnsForbidden(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(ownedTrip(), nil).Once()
	subs := &fakeSubscriptionLookup{err: domainerrors.NotFound("subscription", testUserID)}
	svc := trip.NewService(repo, subs)

	err := svc.Delete(context.Background(), testUserID, testTripID)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "forbidden", domainErr.Code)
}

func TestUnitDelete_SubscriptionLookupFails_PropagatesError(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(ownedTrip(), nil).Once()
	wantErr := errors.New("connection refused")
	subs := &fakeSubscriptionLookup{err: wantErr}
	svc := trip.NewService(repo, subs)

	err := svc.Delete(context.Background(), testUserID, testTripID)

	require.ErrorIs(t, err, wantErr)
}

func TestUnitDelete_ActiveSubscription_DeletesTrip(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(ownedTrip(), nil).Once()
	repo.EXPECT().DeleteTrip(mock.Anything, testTripID).Return(nil).Once()
	subs := &fakeSubscriptionLookup{sub: &subscription.Subscription{Status: subscription.StatusActive}}
	svc := trip.NewService(repo, subs)

	err := svc.Delete(context.Background(), testUserID, testTripID)

	require.NoError(t, err)
}

// --- GetItinerary ---

const (
	testDayOneID      = "22222222-2222-2222-2222-222222222222"
	testDayTwoID      = "33333333-3333-3333-3333-333333333333"
	testDestinationID = "44444444-4444-4444-4444-444444444444"
)

func TestUnitGetItinerary_OwnedTrip_AssemblesNestedDaysDestinationsAndActivities(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(ownedTrip(), nil).Once()

	label := "Tokyo Day 1"
	dayOne := &trip.Day{ID: testDayOneID, TripID: testTripID, DayNumber: 1, Label: &label, DestinationID: strPtr(testDestinationID)}
	dayTwo := &trip.Day{ID: testDayTwoID, TripID: testTripID, DayNumber: 2} // no destination, no label
	repo.EXPECT().ListDaysByTrip(mock.Anything, testTripID).Return([]*trip.Day{dayOne, dayTwo}, nil).Once()

	activityOne := &trip.Activity{ID: "act-1", DayID: testDayOneID, Title: "Visit Museum", SequenceOrder: 1}
	repo.EXPECT().
		ListActivitiesByDayIDs(mock.Anything, mock.MatchedBy(func(ids []string) bool {
			return len(ids) == 2 && ids[0] == testDayOneID && ids[1] == testDayTwoID
		})).
		Return([]*trip.Activity{activityOne}, nil).Once()

	destination := &trip.Destination{ID: testDestinationID, Name: "Tokyo", Country: "JP"}
	repo.EXPECT().
		ListDestinationsByIDs(mock.Anything, []string{testDestinationID}).
		Return([]*trip.Destination{destination}, nil).Once()

	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	got, err := svc.GetItinerary(context.Background(), testUserID, testTripID)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, testTripID, got.TripID)
	require.Len(t, got.Days, 2)

	first := got.Days[0]
	assert.Equal(t, testDayOneID, first.ID)
	assert.Equal(t, 1, first.DayNumber)
	require.NotNil(t, first.Label)
	assert.Equal(t, label, *first.Label)
	require.NotNil(t, first.Destination)
	assert.Equal(t, testDestinationID, first.Destination.ID)
	require.Len(t, first.Activities, 1)
	assert.Equal(t, "Visit Museum", first.Activities[0].Title)

	second := got.Days[1]
	assert.Equal(t, testDayTwoID, second.ID)
	assert.Nil(t, second.Label, "a day with no label must produce a nil Label, not an empty string")
	assert.Nil(t, second.Destination, "a day with a nil DestinationID must produce a nil Destination")
	assert.Empty(t, second.Activities, "a day with no activities must produce an empty, non-nil slice")
}

func TestUnitGetItinerary_TripHasNoDays_ReturnsEmptyDaysSlice(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).Return(ownedTrip(), nil).Once()
	repo.EXPECT().ListDaysByTrip(mock.Anything, testTripID).Return([]*trip.Day{}, nil).Once()
	repo.EXPECT().ListActivitiesByDayIDs(mock.Anything, []string{}).Return([]*trip.Activity{}, nil).Once()
	repo.EXPECT().ListDestinationsByIDs(mock.Anything, []string{}).Return([]*trip.Destination{}, nil).Once()
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	got, err := svc.GetItinerary(context.Background(), testUserID, testTripID)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Empty(t, got.Days)
}

func TestUnitGetItinerary_NotOwner_ReturnsNotFoundWithoutFurtherRepositoryCalls(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	// No ListDaysByTrip/ListActivitiesByDayIDs/ListDestinationsByIDs .EXPECT():
	// the mock's default "must not be called" behavior proves Get's ownership
	// check short-circuits before any itinerary data is fetched.
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).
		Return(&trip.Trip{ID: testTripID, CreatorID: testOtherID}, nil).Once()
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	got, err := svc.GetItinerary(context.Background(), testUserID, testTripID)

	require.Error(t, err)
	assert.Nil(t, got)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code,
		"a non-owner must see the same 404 Get would produce (anti-enumeration)")
}

func TestUnitGetItinerary_TripNotFound_PropagatesNotFound(t *testing.T) {
	repo := tripmocks.NewMockRepository(t)
	repo.EXPECT().FindTripByID(mock.Anything, testTripID).
		Return(nil, domainerrors.NotFound("trip", testTripID)).Once()
	svc := trip.NewService(repo, &fakeSubscriptionLookup{})

	got, err := svc.GetItinerary(context.Background(), testUserID, testTripID)

	require.Error(t, err)
	assert.Nil(t, got)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}

// strPtr returns a pointer to s, a small helper for building nullable-field
// fixtures inline.
func strPtr(s string) *string { return &s }
