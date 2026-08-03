//go:build test

package conversation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/internal/ai"
	"github.com/JosemaPereira/TrAIveler/backend/internal/conversation"
	convmocks "github.com/JosemaPereira/TrAIveler/backend/internal/conversation/mocks"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

const (
	testTripID    = "11111111-2222-3333-4444-555555555555"
	testSessionID = "99999999-8888-7777-6666-555555555555"
)

// fakeItineraryGenerator hand-fakes conversation.ItineraryGenerator, the
// small consumer-owned port `make mocks` would otherwise auto-generate (see
// docs/mock-standards.md and .github/memory/patterns-discovered.md's
// "mockery --all Regenerates a Mock for Every Interface..." entry) — its
// generated mock is deliberately not committed. No real implementation
// exists yet; 001-T037 (the future Itinerary service) is expected to
// provide one.
type fakeItineraryGenerator struct {
	reply      string
	ready      bool
	err        error
	gotHistory []ai.Message
}

func (f *fakeItineraryGenerator) Continue(_ context.Context, _ string, history []ai.Message) (string, bool, error) {
	f.gotHistory = history
	return f.reply, f.ready, f.err
}

// --- SendMessage: session reuse vs. creation ---

func TestUnitSendMessage_NoExistingSession_StartsNewInProgressSession(t *testing.T) {
	repo := convmocks.NewMockRepository(t)
	repo.EXPECT().GetSessionByTrip(mock.Anything, testTripID).
		Return(nil, domainerrors.NotFound("conversation session for trip", testTripID)).Once()
	repo.EXPECT().
		CreateSession(mock.Anything, mock.MatchedBy(func(s *conversation.Session) bool {
			return s.ID != "" && s.TripID == testTripID && s.Status == conversation.StatusInProgress
		})).
		Return(nil).Once()
	repo.EXPECT().AppendMessage(mock.Anything, mock.Anything).Return(nil).Twice()
	repo.EXPECT().ListMessages(mock.Anything, mock.Anything).Return([]*conversation.Message{}, nil).Once()
	gen := &fakeItineraryGenerator{reply: "Sure, where to?", ready: false}
	svc := conversation.NewService(repo, gen)

	assistantMsg, ready, err := svc.SendMessage(context.Background(), testTripID, "Plan me a trip")

	require.NoError(t, err)
	assert.False(t, ready)
	require.NotNil(t, assistantMsg)
	assert.Equal(t, "Sure, where to?", assistantMsg.Content)
	assert.Equal(t, conversation.RoleAssistant, assistantMsg.Role)
}

func TestUnitSendMessage_ExistingInProgressSession_ReusesSessionAndThreadsMessages(t *testing.T) {
	repo := convmocks.NewMockRepository(t) // no .EXPECT() CreateSession: must not be called
	existing := &conversation.Session{ID: testSessionID, TripID: testTripID, Status: conversation.StatusInProgress}
	repo.EXPECT().GetSessionByTrip(mock.Anything, testTripID).Return(existing, nil).Once()
	repo.EXPECT().
		AppendMessage(mock.Anything, mock.MatchedBy(func(m *conversation.Message) bool {
			return m.SessionID == testSessionID && m.Role == conversation.RoleUser && m.Content == "More details"
		})).
		Return(nil).Once()
	repo.EXPECT().ListMessages(mock.Anything, testSessionID).
		Return([]*conversation.Message{{SessionID: testSessionID, Role: conversation.RoleUser, Content: "More details"}}, nil).
		Once()
	repo.EXPECT().
		AppendMessage(mock.Anything, mock.MatchedBy(func(m *conversation.Message) bool {
			return m.SessionID == testSessionID && m.Role == conversation.RoleAssistant
		})).
		Return(nil).Once()
	gen := &fakeItineraryGenerator{reply: "Continuing...", ready: false}
	svc := conversation.NewService(repo, gen)

	_, _, err := svc.SendMessage(context.Background(), testTripID, "More details")

	require.NoError(t, err)
	require.Len(t, gen.gotHistory, 1,
		"the history sent to the generator must include the just-appended user turn")
	assert.Equal(t, ai.Message{Role: conversation.RoleUser, Content: "More details"}, gen.gotHistory[0])
}

func TestUnitSendMessage_MostRecentSessionNoLongerInProgress_StartsFreshSession(t *testing.T) {
	repo := convmocks.NewMockRepository(t)
	completed := &conversation.Session{ID: "old-session", TripID: testTripID, Status: conversation.StatusCompleted}
	repo.EXPECT().GetSessionByTrip(mock.Anything, testTripID).Return(completed, nil).Once()
	repo.EXPECT().
		CreateSession(mock.Anything, mock.MatchedBy(func(s *conversation.Session) bool {
			return s.ID != "" && s.ID != "old-session" &&
				s.TripID == testTripID && s.Status == conversation.StatusInProgress
		})).
		Return(nil).Once()
	repo.EXPECT().AppendMessage(mock.Anything, mock.Anything).Return(nil).Twice()
	repo.EXPECT().ListMessages(mock.Anything, mock.Anything).Return([]*conversation.Message{}, nil).Once()
	gen := &fakeItineraryGenerator{reply: "Let's start planning again", ready: false}
	svc := conversation.NewService(repo, gen)

	_, _, err := svc.SendMessage(context.Background(), testTripID, "Start over")

	require.NoError(t, err,
		"a trip whose most recent session already completed/abandoned must start a fresh session")
}

func TestUnitSendMessage_CreateSessionFails_PropagatesError(t *testing.T) {
	repo := convmocks.NewMockRepository(t) // no .EXPECT() AppendMessage: must not be reached
	repo.EXPECT().GetSessionByTrip(mock.Anything, testTripID).
		Return(nil, domainerrors.NotFound("conversation session for trip", testTripID)).Once()
	wantErr := errors.New("write failed")
	repo.EXPECT().CreateSession(mock.Anything, mock.Anything).Return(wantErr).Once()
	svc := conversation.NewService(repo, &fakeItineraryGenerator{})

	_, _, err := svc.SendMessage(context.Background(), testTripID, "Hello")

	require.ErrorIs(t, err, wantErr)
}

func TestUnitSendMessage_GetSessionByTripFailsWithNonNotFoundError_PropagatesError(t *testing.T) {
	repo := convmocks.NewMockRepository(t)
	wantErr := errors.New("connection refused")
	repo.EXPECT().GetSessionByTrip(mock.Anything, testTripID).Return(nil, wantErr).Once()
	svc := conversation.NewService(repo, &fakeItineraryGenerator{})

	_, _, err := svc.SendMessage(context.Background(), testTripID, "Hello")

	require.ErrorIs(t, err, wantErr)
}

// --- SendMessage: propagation of repository/generator failures ---

func TestUnitSendMessage_AppendUserMessageFails_PropagatesError(t *testing.T) {
	repo := convmocks.NewMockRepository(t)
	existing := &conversation.Session{ID: testSessionID, TripID: testTripID, Status: conversation.StatusInProgress}
	repo.EXPECT().GetSessionByTrip(mock.Anything, testTripID).Return(existing, nil).Once()
	wantErr := errors.New("write failed")
	repo.EXPECT().AppendMessage(mock.Anything, mock.Anything).Return(wantErr).Once()
	svc := conversation.NewService(repo, &fakeItineraryGenerator{})

	_, _, err := svc.SendMessage(context.Background(), testTripID, "Hello")

	require.ErrorIs(t, err, wantErr)
}

func TestUnitSendMessage_ListMessagesFails_PropagatesError(t *testing.T) {
	repo := convmocks.NewMockRepository(t)
	existing := &conversation.Session{ID: testSessionID, TripID: testTripID, Status: conversation.StatusInProgress}
	repo.EXPECT().GetSessionByTrip(mock.Anything, testTripID).Return(existing, nil).Once()
	repo.EXPECT().AppendMessage(mock.Anything, mock.Anything).Return(nil).Once()
	wantErr := errors.New("read failed")
	repo.EXPECT().ListMessages(mock.Anything, testSessionID).Return(nil, wantErr).Once()
	svc := conversation.NewService(repo, &fakeItineraryGenerator{})

	_, _, err := svc.SendMessage(context.Background(), testTripID, "Hello")

	require.ErrorIs(t, err, wantErr)
}

func TestUnitSendMessage_GeneratorFails_PropagatesError(t *testing.T) {
	repo := convmocks.NewMockRepository(t)
	existing := &conversation.Session{ID: testSessionID, TripID: testTripID, Status: conversation.StatusInProgress}
	repo.EXPECT().GetSessionByTrip(mock.Anything, testTripID).Return(existing, nil).Once()
	repo.EXPECT().AppendMessage(mock.Anything, mock.Anything).Return(nil).Once()
	repo.EXPECT().ListMessages(mock.Anything, testSessionID).Return([]*conversation.Message{}, nil).Once()
	wantErr := errors.New("ai provider unavailable")
	gen := &fakeItineraryGenerator{err: wantErr}
	svc := conversation.NewService(repo, gen)

	_, _, err := svc.SendMessage(context.Background(), testTripID, "Hello")

	require.ErrorIs(t, err, wantErr)
}

func TestUnitSendMessage_AppendAssistantMessageFails_PropagatesError(t *testing.T) {
	repo := convmocks.NewMockRepository(t)
	existing := &conversation.Session{ID: testSessionID, TripID: testTripID, Status: conversation.StatusInProgress}
	repo.EXPECT().GetSessionByTrip(mock.Anything, testTripID).Return(existing, nil).Once()
	repo.EXPECT().
		AppendMessage(mock.Anything, mock.MatchedBy(func(m *conversation.Message) bool {
			return m.Role == conversation.RoleUser
		})).
		Return(nil).Once()
	repo.EXPECT().ListMessages(mock.Anything, testSessionID).Return([]*conversation.Message{}, nil).Once()
	wantErr := errors.New("write failed")
	repo.EXPECT().
		AppendMessage(mock.Anything, mock.MatchedBy(func(m *conversation.Message) bool {
			return m.Role == conversation.RoleAssistant
		})).
		Return(wantErr).Once()
	gen := &fakeItineraryGenerator{reply: "Here you go"}
	svc := conversation.NewService(repo, gen)

	_, _, err := svc.SendMessage(context.Background(), testTripID, "Hello")

	require.ErrorIs(t, err, wantErr)
}

// --- SendMessage: itinerary_ready -> CompleteSession ---

func TestUnitSendMessage_GeneratorReportsReady_CompletesSession(t *testing.T) {
	repo := convmocks.NewMockRepository(t)
	existing := &conversation.Session{ID: testSessionID, TripID: testTripID, Status: conversation.StatusInProgress}
	repo.EXPECT().GetSessionByTrip(mock.Anything, testTripID).Return(existing, nil).Once()
	repo.EXPECT().AppendMessage(mock.Anything, mock.Anything).Return(nil).Twice()
	repo.EXPECT().ListMessages(mock.Anything, testSessionID).Return([]*conversation.Message{}, nil).Once()
	repo.EXPECT().CompleteSession(mock.Anything, testSessionID).Return(nil).Once()
	gen := &fakeItineraryGenerator{reply: "Here is your final itinerary", ready: true}
	svc := conversation.NewService(repo, gen)

	_, ready, err := svc.SendMessage(context.Background(), testTripID, "Looks good, accept it")

	require.NoError(t, err)
	assert.True(t, ready)
}

func TestUnitSendMessage_GeneratorNotReady_DoesNotCompleteSession(t *testing.T) {
	repo := convmocks.NewMockRepository(t) // no .EXPECT() CompleteSession: must not be called
	existing := &conversation.Session{ID: testSessionID, TripID: testTripID, Status: conversation.StatusInProgress}
	repo.EXPECT().GetSessionByTrip(mock.Anything, testTripID).Return(existing, nil).Once()
	repo.EXPECT().AppendMessage(mock.Anything, mock.Anything).Return(nil).Twice()
	repo.EXPECT().ListMessages(mock.Anything, testSessionID).Return([]*conversation.Message{}, nil).Once()
	gen := &fakeItineraryGenerator{reply: "Tell me more", ready: false}
	svc := conversation.NewService(repo, gen)

	_, ready, err := svc.SendMessage(context.Background(), testTripID, "I want to go to Tokyo")

	require.NoError(t, err)
	assert.False(t, ready)
}

func TestUnitSendMessage_CompleteSessionFails_PropagatesError(t *testing.T) {
	repo := convmocks.NewMockRepository(t)
	existing := &conversation.Session{ID: testSessionID, TripID: testTripID, Status: conversation.StatusInProgress}
	repo.EXPECT().GetSessionByTrip(mock.Anything, testTripID).Return(existing, nil).Once()
	repo.EXPECT().AppendMessage(mock.Anything, mock.Anything).Return(nil).Twice()
	repo.EXPECT().ListMessages(mock.Anything, testSessionID).Return([]*conversation.Message{}, nil).Once()
	wantErr := errors.New("update failed")
	repo.EXPECT().CompleteSession(mock.Anything, testSessionID).Return(wantErr).Once()
	gen := &fakeItineraryGenerator{reply: "Final itinerary", ready: true}
	svc := conversation.NewService(repo, gen)

	_, _, err := svc.SendMessage(context.Background(), testTripID, "Accept it")

	require.ErrorIs(t, err, wantErr)
}

// --- GetHistory ---

func TestUnitGetHistory_ExistingSession_ReturnsSessionAndMessages(t *testing.T) {
	repo := convmocks.NewMockRepository(t)
	session := &conversation.Session{ID: testSessionID, TripID: testTripID, Status: conversation.StatusInProgress}
	messages := []*conversation.Message{{SessionID: testSessionID, Role: conversation.RoleUser, Content: "Hi"}}
	repo.EXPECT().GetSessionByTrip(mock.Anything, testTripID).Return(session, nil).Once()
	repo.EXPECT().ListMessages(mock.Anything, testSessionID).Return(messages, nil).Once()
	svc := conversation.NewService(repo, &fakeItineraryGenerator{})

	gotSession, gotMessages, err := svc.GetHistory(context.Background(), testTripID)

	require.NoError(t, err)
	assert.Equal(t, session, gotSession)
	assert.Equal(t, messages, gotMessages)
}

func TestUnitGetHistory_ListMessagesFails_PropagatesError(t *testing.T) {
	repo := convmocks.NewMockRepository(t)
	session := &conversation.Session{ID: testSessionID, TripID: testTripID, Status: conversation.StatusInProgress}
	repo.EXPECT().GetSessionByTrip(mock.Anything, testTripID).Return(session, nil).Once()
	wantErr := errors.New("read failed")
	repo.EXPECT().ListMessages(mock.Anything, testSessionID).Return(nil, wantErr).Once()
	svc := conversation.NewService(repo, &fakeItineraryGenerator{})

	_, _, err := svc.GetHistory(context.Background(), testTripID)

	require.ErrorIs(t, err, wantErr)
}

func TestUnitGetHistory_NoSessionForTrip_PropagatesNotFound(t *testing.T) {
	repo := convmocks.NewMockRepository(t) // no .EXPECT() ListMessages: must not be called
	repo.EXPECT().GetSessionByTrip(mock.Anything, testTripID).
		Return(nil, domainerrors.NotFound("conversation session for trip", testTripID)).Once()
	svc := conversation.NewService(repo, &fakeItineraryGenerator{})

	_, _, err := svc.GetHistory(context.Background(), testTripID)

	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "not_found", domainErr.Code)
}
