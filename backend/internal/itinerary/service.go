package itinerary

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JosemaPereira/TrAIveler/backend/internal/ai"
	"github.com/JosemaPereira/TrAIveler/backend/internal/conversation"
	"github.com/JosemaPereira/TrAIveler/backend/internal/trip"
)

// Persister is the narrow, consumer-owned persistence port Service needs to
// turn a ready AI response into Destination/Day/Activity rows. It is
// satisfied structurally by trip.Repository (no wiring required) and is
// deliberately not registered in backend/.mockery.yaml — like
// conversation.ItineraryGenerator and trip.SubscriptionLookup, this is a
// small internal port hand-faked in tests rather than mockery-generated (see
// .github/memory/patterns-discovered.md's "mockery --all Regenerates a Mock
// for Every Interface..." entry).
type Persister interface {
	CreateDestination(ctx context.Context, destination *trip.Destination) error
	UpsertDay(ctx context.Context, day *trip.Day) error
	UpsertActivity(ctx context.Context, activity *trip.Activity) error
}

// Service implements conversation.ItineraryGenerator (001-T037): it drives
// the AI provider through a trip-planning conversation via Continue,
// persisting the resulting itinerary once the AI reports it is ready.
type Service struct {
	ai   ai.AIClient
	repo Persister
}

// NewService builds a Service backed by client (the AI provider) and repo
// (the persistence port).
func NewService(client ai.AIClient, repo Persister) *Service {
	return &Service{ai: client, repo: repo}
}

// compile-time assurance that Service satisfies conversation.ItineraryGenerator.
var _ conversation.ItineraryGenerator = (*Service)(nil)

// destinationKey identifies a Destination for within-call deduplication (see
// resolveDestination): the same (name, country) pair appearing across
// multiple days of one Continue call must resolve to a single persisted row.
type destinationKey struct {
	name    string
	country string
}

// Continue implements conversation.ItineraryGenerator: it streams the AI
// provider's next conversational turn for tripID given the conversation so
// far, and either returns a clarifying question (ready=false) or persists
// the completed itinerary and returns a completion message (ready=true).
//
// Preferences is deliberately left zero-valued on the outgoing request: this
// method's signature carries no structured preferences parameter, so the
// conversational flow relies entirely on the AI extracting trip constraints
// from the free-text history — see SystemPrompt.
func (s *Service) Continue(ctx context.Context, tripID string, history []ai.Message) (string, bool, error) {
	req := ai.ItineraryRequest{
		TripID:              tripID,
		ConversationHistory: history,
		SystemPrompt:        SystemPrompt,
	}

	content, err := s.accumulateStream(ctx, req)
	if err != nil {
		return "", false, ai.TranslateError(err)
	}

	var resp ai.ItineraryResponse
	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		return "", false, fmt.Errorf("itinerary: parse AI response JSON: %w", err)
	}

	if !resp.Ready {
		return resp.Reply, false, nil
	}

	if err := s.persist(ctx, tripID, resp); err != nil {
		return "", false, err
	}

	return resp.Reply, true, nil
}

// accumulateStream drives req through s.ai.StreamItinerary and accumulates
// StreamChunk.Content until a Done chunk (or the chunk channel closing
// without one), watching the error channel throughout. Any error received
// from the error channel aborts accumulation immediately and is returned
// as-is (translation into this codebase's domain error vocabulary is the
// caller's job — see Continue). Each channel is set to nil once observed
// closed so the select loop stops considering it, avoiding a busy-loop
// against an already-closed, always-ready channel.
func (s *Service) accumulateStream(ctx context.Context, req ai.ItineraryRequest) (string, error) {
	chunkCh, errCh := s.ai.StreamItinerary(ctx, req)

	var content strings.Builder
	for chunkCh != nil || errCh != nil {
		select {
		case chunk, ok := <-chunkCh:
			if !ok {
				chunkCh = nil
				continue
			}
			if chunk.Done {
				return content.String(), nil
			}
			content.WriteString(chunk.Content)
		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			if err != nil {
				return "", err
			}
		}
	}

	return content.String(), nil
}

// persist writes resp's days and activities as Destination/Day/Activity rows
// for tripID. It only reads resp.Days[].Destination and
// resp.Days[].Activities — the top-level resp.Destinations/resp.Activities
// fields are a separate, pre-existing duplication of the same data and are
// deliberately left unread (out of this ticket's scope to clean up).
//
// Destinations are deduped by (name, country) within this single call: a
// multi-day stay in the same city resolves to one CreateDestination call,
// not one per day (see destinationKey) — the database's lack of a
// uniqueness constraint on destinations tolerates cross-operation variation,
// not same-operation duplication.
//
// If persistence fails partway, the error is returned as-is: no rollback
// mechanism exists at this layer, consistent with the rest of this
// codebase's repositories (single-statement or caller-orchestrated, no
// explicit transactions).
func (s *Service) persist(ctx context.Context, tripID string, resp ai.ItineraryResponse) error {
	destinationIDs := make(map[destinationKey]string)

	for _, day := range resp.Days {
		var destinationID *string
		if day.Destination.Name != "" {
			id, err := s.resolveDestination(ctx, destinationIDs, day.Destination)
			if err != nil {
				return err
			}
			destinationID = &id
		}

		dayRow := &trip.Day{
			TripID:        tripID,
			DestinationID: destinationID,
			DayNumber:     day.DayNumber,
		}
		if err := s.repo.UpsertDay(ctx, dayRow); err != nil {
			return fmt.Errorf("itinerary: upsert day %d: %w", day.DayNumber, err)
		}

		if err := s.persistActivities(ctx, dayRow.ID, day.Activities); err != nil {
			return err
		}
	}

	return nil
}

// persistActivities inserts each of a day's activities, marking them
// IsAIGenerated and leaving ID empty so UpsertActivity takes the insert path
// (see trip.Repository.UpsertActivity's doc comment).
func (s *Service) persistActivities(ctx context.Context, dayID string, activities []ai.Activity) error {
	for _, activity := range activities {
		activityRow := &trip.Activity{
			DayID:         dayID,
			Title:         activity.Title,
			Type:          activity.Type,
			SequenceOrder: activity.SequenceOrder,
			Description:   stringPtrOrNil(activity.Description),
			IsAIGenerated: true,
		}
		if err := s.repo.UpsertActivity(ctx, activityRow); err != nil {
			return fmt.Errorf("itinerary: upsert activity %q: %w", activity.Title, err)
		}
	}
	return nil
}

// resolveDestination returns the persisted ID for dest, creating it via
// CreateDestination on first sight and caching the result in ids so a
// repeated (name, country) pair within the same persist call reuses the same
// row instead of inserting a duplicate.
func (s *Service) resolveDestination(ctx context.Context, ids map[destinationKey]string, dest ai.Destination) (string, error) {
	key := destinationKey{name: dest.Name, country: dest.Country}
	if id, ok := ids[key]; ok {
		return id, nil
	}

	row := &trip.Destination{
		Name:      dest.Name,
		Country:   dest.Country,
		Region:    stringPtrOrNil(dest.Region),
		Latitude:  dest.Latitude,
		Longitude: dest.Longitude,
	}
	if err := s.repo.CreateDestination(ctx, row); err != nil {
		return "", fmt.Errorf("itinerary: create destination %q: %w", dest.Name, err)
	}

	ids[key] = row.ID
	return row.ID, nil
}

// stringPtrOrNil returns a pointer to s, or nil when s is empty — used to
// populate trip's nullable string fields (Region, Description) from the AI
// response's plain (always-present) string fields.
func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
