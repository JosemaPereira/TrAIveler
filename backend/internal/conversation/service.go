package conversation

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/JosemaPereira/TrAIveler/backend/internal/ai"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// ItineraryGenerator produces the AI's next conversational turn for a trip's
// planning session and reports whether the conversation has produced a
// ready-to-persist itinerary. Implemented by the future Itinerary service
// (001-T037, docs/roadmap.md) — no implementation exists yet; this
// package's tests use a hand-faked stub. Mirrors the RefreshTokenStore/
// SubscriptionResolver interface-first pattern in
// backend/internal/auth/jwt/refresher.go.
type ItineraryGenerator interface {
	Continue(ctx context.Context, tripID string, history []ai.Message) (reply string, ready bool, err error)
}

// Service contains Conversation business logic (001-T036): reusing or
// starting a trip's in-progress planning session, appending user/assistant
// turns, delegating the AI's next turn to an ItineraryGenerator, and
// completing the session once the conversation produces a
// ready-to-persist itinerary. It has no knowledge of HTTP or SQL,
// delegating persistence to a Repository and AI generation to an
// ItineraryGenerator.
//
// Deliberately performs no ownership/role check: Trip ownership is already
// enforced by trip.Service.Get's anti-enumeration check, and the future
// Conversation HTTP handler (001-T039) is expected to call that first
// before reaching this service — the same way nested REST resources
// inherit their parent's access control.
type Service struct {
	repo Repository
	gen  ItineraryGenerator
}

// NewService builds a Service backed by repo and gen.
func NewService(repo Repository, gen ItineraryGenerator) *Service {
	return &Service{repo: repo, gen: gen}
}

// SendMessage appends content as a user turn to tripID's active planning
// session (reusing the current in_progress session, or starting a new one
// if none exists or the most recent one already completed/abandoned — see
// resolveActiveSession), asks the ItineraryGenerator for the AI's reply,
// appends that reply as an assistant turn, and completes the session when
// the generator reports the conversation is ready to be turned into an
// itinerary.
func (s *Service) SendMessage(ctx context.Context, tripID, content string) (*Message, bool, error) {
	session, err := s.resolveActiveSession(ctx, tripID)
	if err != nil {
		return nil, false, err
	}

	if err := s.repo.AppendMessage(ctx, &Message{
		SessionID: session.ID,
		Role:      RoleUser,
		Content:   content,
	}); err != nil {
		return nil, false, err
	}

	history, err := s.repo.ListMessages(ctx, session.ID)
	if err != nil {
		return nil, false, err
	}

	reply, ready, err := s.gen.Continue(ctx, tripID, toAIHistory(history))
	if err != nil {
		return nil, false, err
	}

	assistantMessage := &Message{
		SessionID: session.ID,
		Role:      RoleAssistant,
		Content:   reply,
	}
	if err := s.repo.AppendMessage(ctx, assistantMessage); err != nil {
		return nil, false, err
	}

	if ready {
		if err := s.repo.CompleteSession(ctx, session.ID); err != nil {
			return nil, false, err
		}
	}

	return assistantMessage, ready, nil
}

// GetHistory returns tripID's most recent conversation session and its
// messages in chronological order. A trip with no conversation started yet
// surfaces the repository's domain NotFound as-is.
func (s *Service) GetHistory(ctx context.Context, tripID string) (*Session, []*Message, error) {
	session, err := s.repo.GetSessionByTrip(ctx, tripID)
	if err != nil {
		return nil, nil, err
	}

	messages, err := s.repo.ListMessages(ctx, session.ID)
	if err != nil {
		return nil, nil, err
	}

	return session, messages, nil
}

// resolveActiveSession returns tripID's in_progress session, starting a
// fresh one when the trip has no session yet or its most recent session is
// no longer in_progress. This is what satisfies docs/data-model.md
// §ConversationSession's "only one active session per trip at a time"
// [Logic] rule: reuse rather than a separate uniqueness check.
func (s *Service) resolveActiveSession(ctx context.Context, tripID string) (*Session, error) {
	session, err := s.repo.GetSessionByTrip(ctx, tripID)
	if err != nil {
		if !isNotFound(err) {
			return nil, err
		}
		return s.startSession(ctx, tripID)
	}
	if session.Status != StatusInProgress {
		return s.startSession(ctx, tripID)
	}
	return session, nil
}

// startSession creates and persists a fresh in_progress session for tripID.
func (s *Service) startSession(ctx context.Context, tripID string) (*Session, error) {
	session := &Session{
		ID:     uuid.NewString(),
		TripID: tripID,
		Status: StatusInProgress,
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

// toAIHistory maps this package's Message rows to the ai.Message shape the
// ItineraryGenerator consumes.
func toAIHistory(messages []*Message) []ai.Message {
	history := make([]ai.Message, 0, len(messages))
	for _, m := range messages {
		history = append(history, ai.Message{Role: m.Role, Content: m.Content})
	}
	return history
}

// isNotFound reports whether err is (or wraps) a domain NotFound error.
func isNotFound(err error) bool {
	var domainErr *domainerrors.DomainError
	return errors.As(err, &domainErr) && domainErr.Code == "not_found"
}
