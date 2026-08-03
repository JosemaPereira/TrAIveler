//go:build test

package itinerary_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/internal/ai"
	aimocks "github.com/JosemaPereira/TrAIveler/backend/internal/ai/mocks"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/itinerary"
	"github.com/JosemaPereira/TrAIveler/backend/internal/trip"
)

const testTripID = "11111111-2222-3333-4444-555555555555"

var testHistory = []ai.Message{
	{Role: "user", Content: "I have 10 days in Japan. I'm an anime fan, my partner loves metal music."},
}

// fakePersister hand-fakes itinerary.Persister, the small consumer-owned port
// `make mocks` would otherwise auto-generate (see docs/mock-standards.md and
// .github/memory/patterns-discovered.md's "mockery --all Regenerates a Mock
// for Every Interface..." entry) — its generated mock is deliberately not
// committed, mirroring conversation.ItineraryGenerator/trip.SubscriptionLookup's
// own hand-faked test doubles.
type fakePersister struct {
	createDestinationErr error
	upsertDayErr         error
	upsertActivityErr    error

	createDestinationCalls []trip.Destination
	upsertDayCalls         []trip.Day
	upsertActivityCalls    []trip.Activity
}

func (f *fakePersister) CreateDestination(_ context.Context, destination *trip.Destination) error {
	f.createDestinationCalls = append(f.createDestinationCalls, *destination)
	if f.createDestinationErr != nil {
		return f.createDestinationErr
	}
	destination.ID = fmt.Sprintf("dest-%d", len(f.createDestinationCalls))
	return nil
}

func (f *fakePersister) UpsertDay(_ context.Context, day *trip.Day) error {
	f.upsertDayCalls = append(f.upsertDayCalls, *day)
	if f.upsertDayErr != nil {
		return f.upsertDayErr
	}
	day.ID = fmt.Sprintf("day-%d", len(f.upsertDayCalls))
	return nil
}

func (f *fakePersister) UpsertActivity(_ context.Context, activity *trip.Activity) error {
	f.upsertActivityCalls = append(f.upsertActivityCalls, *activity)
	return f.upsertActivityErr
}

// closedStream returns a chunk channel pre-loaded with content followed by a
// Done chunk, and an already-closed (empty) error channel — mirrors what a
// successful AIClient.StreamItinerary call produces once fully drained.
func closedStream(content string) (<-chan ai.StreamChunk, <-chan error) {
	chunkCh := make(chan ai.StreamChunk, 2)
	chunkCh <- ai.StreamChunk{Content: content}
	chunkCh <- ai.StreamChunk{Done: true}
	close(chunkCh)

	errCh := make(chan error, 1)
	close(errCh)

	return chunkCh, errCh
}

// erroredStream returns an already-closed (empty) chunk channel and an error
// channel pre-loaded with err — mirrors what a failing AIClient.StreamItinerary
// call produces once fully drained.
func erroredStream(err error) (<-chan ai.StreamChunk, <-chan error) {
	chunkCh := make(chan ai.StreamChunk)
	close(chunkCh)

	errCh := make(chan error, 1)
	errCh <- err
	close(errCh)

	return chunkCh, errCh
}

// matchesRequest builds a mock.MatchedBy predicate asserting Continue built
// the ai.ItineraryRequest correctly: the given tripID/history forwarded
// as-is, Preferences left zero-valued (the conversational MVP flow relies on
// the AI extracting constraints from free text, not a structured field), and
// a non-empty SystemPrompt matching itinerary.SystemPrompt.
func matchesRequest(tripID string, history []ai.Message) func(ai.ItineraryRequest) bool {
	return func(req ai.ItineraryRequest) bool {
		return req.TripID == tripID &&
			assert.ObjectsAreEqual(history, req.ConversationHistory) &&
			assert.ObjectsAreEqual(ai.TravelPreferences{}, req.Preferences) &&
			req.SystemPrompt == itinerary.SystemPrompt
	}
}

// --- Continue: not-ready (clarifying question) turns ---

func TestUnitContinue_NotReady_ReturnsReplyWithoutPersisting(t *testing.T) {
	client := aimocks.NewMockAIClient(t)
	chunkCh, errCh := closedStream(`{"ready": false, "reply": "Where would you like to go?"}`)
	client.EXPECT().
		StreamItinerary(mock.Anything, mock.MatchedBy(matchesRequest(testTripID, testHistory))).
		Return(chunkCh, errCh).
		Once()
	persister := &fakePersister{}
	svc := itinerary.NewService(client, persister)

	reply, ready, err := svc.Continue(context.Background(), testTripID, testHistory)

	require.NoError(t, err)
	assert.False(t, ready)
	assert.Equal(t, "Where would you like to go?", reply)
	assert.Empty(t, persister.createDestinationCalls)
	assert.Empty(t, persister.upsertDayCalls)
	assert.Empty(t, persister.upsertActivityCalls)
}

// --- Continue: ready turns, persistence ---

func TestUnitContinue_ReadySingleDaySingleDestination_PersistsAndReturnsReady(t *testing.T) {
	const readyJSON = `{
		"ready": true,
		"reply": "Your itinerary is ready!",
		"days": [
			{
				"day_number": 1,
				"destination": {"name": "Lisbon", "country": "PT", "region": "Lisbon District", "latitude": 38.7223, "longitude": -9.1393},
				"activities": [
					{"title": "Belem Tower", "type": "visit", "description": "Historic tower", "sequence_order": 1},
					{"title": "Dinner at Time Out Market", "type": "food", "sequence_order": 2}
				]
			}
		]
	}`
	client := aimocks.NewMockAIClient(t)
	chunkCh, errCh := closedStream(readyJSON)
	client.EXPECT().
		StreamItinerary(mock.Anything, mock.MatchedBy(matchesRequest(testTripID, testHistory))).
		Return(chunkCh, errCh).
		Once()
	persister := &fakePersister{}
	svc := itinerary.NewService(client, persister)

	reply, ready, err := svc.Continue(context.Background(), testTripID, testHistory)

	require.NoError(t, err)
	assert.True(t, ready)
	assert.Equal(t, "Your itinerary is ready!", reply)

	require.Len(t, persister.createDestinationCalls, 1)
	dest := persister.createDestinationCalls[0]
	assert.Equal(t, "Lisbon", dest.Name)
	assert.Equal(t, "PT", dest.Country)
	require.NotNil(t, dest.Region)
	assert.Equal(t, "Lisbon District", *dest.Region)
	assert.InDelta(t, 38.7223, dest.Latitude, 0.0001)
	assert.InDelta(t, -9.1393, dest.Longitude, 0.0001)

	require.Len(t, persister.upsertDayCalls, 1)
	day := persister.upsertDayCalls[0]
	assert.Equal(t, testTripID, day.TripID)
	require.NotNil(t, day.DestinationID)
	assert.Equal(t, "dest-1", *day.DestinationID)
	assert.Equal(t, 1, day.DayNumber)

	require.Len(t, persister.upsertActivityCalls, 2)
	act0 := persister.upsertActivityCalls[0]
	assert.Equal(t, "day-1", act0.DayID)
	assert.Equal(t, "Belem Tower", act0.Title)
	assert.Equal(t, "visit", act0.Type)
	assert.Equal(t, 1, act0.SequenceOrder)
	require.NotNil(t, act0.Description)
	assert.Equal(t, "Historic tower", *act0.Description)
	assert.True(t, act0.IsAIGenerated)
	assert.Empty(t, act0.ID, "empty ID must be left for the repository to take the insert path")

	act1 := persister.upsertActivityCalls[1]
	assert.Equal(t, "Dinner at Time Out Market", act1.Title)
	assert.Equal(t, 2, act1.SequenceOrder)
	assert.Nil(t, act1.Description, "an empty AI-supplied description must map to a nil pointer")
}

func TestUnitContinue_ReadyMultiDaySameDestination_DedupesDestinationCreation(t *testing.T) {
	const readyJSON = `{
		"ready": true,
		"reply": "Your itinerary is ready!",
		"days": [
			{
				"day_number": 1,
				"destination": {"name": "Lisbon", "country": "PT", "region": "Lisbon District", "latitude": 38.72, "longitude": -9.13},
				"activities": [{"title": "Alfama walk", "type": "visit", "sequence_order": 1}]
			},
			{
				"day_number": 2,
				"destination": {"name": "Lisbon", "country": "PT", "region": "Lisbon District", "latitude": 38.72, "longitude": -9.13},
				"activities": [{"title": "Belem Tower", "type": "visit", "sequence_order": 1}]
			}
		]
	}`
	client := aimocks.NewMockAIClient(t)
	chunkCh, errCh := closedStream(readyJSON)
	client.EXPECT().StreamItinerary(mock.Anything, mock.Anything).Return(chunkCh, errCh).Once()
	persister := &fakePersister{}
	svc := itinerary.NewService(client, persister)

	_, ready, err := svc.Continue(context.Background(), testTripID, testHistory)

	require.NoError(t, err)
	assert.True(t, ready)
	require.Len(t, persister.createDestinationCalls, 1,
		"a multi-day stay in the same city must dedupe to a single CreateDestination call")

	require.Len(t, persister.upsertDayCalls, 2)
	require.NotNil(t, persister.upsertDayCalls[0].DestinationID)
	require.NotNil(t, persister.upsertDayCalls[1].DestinationID)
	assert.Equal(t, *persister.upsertDayCalls[0].DestinationID, *persister.upsertDayCalls[1].DestinationID,
		"both days must reference the same deduped destination ID")
}

func TestUnitContinue_ReadyDayWithNoDestination_UpsertsDayWithNilDestinationID(t *testing.T) {
	const readyJSON = `{
		"ready": true,
		"reply": "Your itinerary is ready!",
		"days": [
			{
				"day_number": 1,
				"destination": {"name": "", "country": ""},
				"activities": [{"title": "Flight to Lisbon", "type": "transfer", "sequence_order": 1}]
			}
		]
	}`
	client := aimocks.NewMockAIClient(t)
	chunkCh, errCh := closedStream(readyJSON)
	client.EXPECT().StreamItinerary(mock.Anything, mock.Anything).Return(chunkCh, errCh).Once()
	persister := &fakePersister{}
	svc := itinerary.NewService(client, persister)

	_, ready, err := svc.Continue(context.Background(), testTripID, testHistory)

	require.NoError(t, err)
	assert.True(t, ready)
	assert.Empty(t, persister.createDestinationCalls,
		"an arrival/transfer day with no destination must not call CreateDestination")
	require.Len(t, persister.upsertDayCalls, 1)
	assert.Nil(t, persister.upsertDayCalls[0].DestinationID)
}

// --- Continue: error paths ---

func TestUnitContinue_StreamError_ReturnsTranslatedError(t *testing.T) {
	client := aimocks.NewMockAIClient(t)
	wantErr := &ai.ProviderUnavailableError{RetryAfter: 20 * time.Second, Err: errors.New("boom")}
	chunkCh, errCh := erroredStream(wantErr)
	client.EXPECT().StreamItinerary(mock.Anything, mock.Anything).Return(chunkCh, errCh).Once()
	persister := &fakePersister{}
	svc := itinerary.NewService(client, persister)

	reply, ready, err := svc.Continue(context.Background(), testTripID, testHistory)

	require.Error(t, err)
	assert.Empty(t, reply)
	assert.False(t, ready)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr, "the AIClient error must round-trip through ai.TranslateError")
	assert.Equal(t, "service_unavailable", domainErr.Code)
	assert.Empty(t, persister.createDestinationCalls)
	assert.Empty(t, persister.upsertDayCalls)
}

func TestUnitContinue_MalformedJSON_ReturnsWrappedErrorNoPersistence(t *testing.T) {
	client := aimocks.NewMockAIClient(t)
	chunkCh, errCh := closedStream("not valid json")
	client.EXPECT().StreamItinerary(mock.Anything, mock.Anything).Return(chunkCh, errCh).Once()
	persister := &fakePersister{}
	svc := itinerary.NewService(client, persister)

	reply, ready, err := svc.Continue(context.Background(), testTripID, testHistory)

	require.Error(t, err)
	assert.Empty(t, reply)
	assert.False(t, ready)
	var domainErr *domainerrors.DomainError
	assert.False(t, errors.As(err, &domainErr), "a JSON parse failure must be a plain wrapped error, not a domain error")
	assert.Empty(t, persister.createDestinationCalls)
	assert.Empty(t, persister.upsertDayCalls)
	assert.Empty(t, persister.upsertActivityCalls)
}

func TestUnitContinue_PersistenceFails_PropagatesErrorNoCompletion(t *testing.T) {
	const readyJSON = `{
		"ready": true,
		"reply": "Your itinerary is ready!",
		"days": [
			{
				"day_number": 1,
				"destination": {"name": "Lisbon", "country": "PT", "region": "Lisbon District", "latitude": 38.72, "longitude": -9.13},
				"activities": [{"title": "Belem Tower", "type": "visit", "sequence_order": 1}]
			}
		]
	}`

	tests := []struct {
		name      string
		persister func(wantErr error) *fakePersister
	}{
		{
			name:      "when CreateDestination fails",
			persister: func(wantErr error) *fakePersister { return &fakePersister{createDestinationErr: wantErr} },
		},
		{
			name:      "when UpsertDay fails",
			persister: func(wantErr error) *fakePersister { return &fakePersister{upsertDayErr: wantErr} },
		},
		{
			name:      "when UpsertActivity fails",
			persister: func(wantErr error) *fakePersister { return &fakePersister{upsertActivityErr: wantErr} },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := aimocks.NewMockAIClient(t)
			chunkCh, errCh := closedStream(readyJSON)
			client.EXPECT().StreamItinerary(mock.Anything, mock.Anything).Return(chunkCh, errCh).Once()
			wantErr := errors.New("write failed")
			persister := tt.persister(wantErr)
			svc := itinerary.NewService(client, persister)

			reply, ready, err := svc.Continue(context.Background(), testTripID, testHistory)

			require.ErrorIs(t, err, wantErr)
			assert.False(t, ready)
			assert.Empty(t, reply)
		})
	}
}
