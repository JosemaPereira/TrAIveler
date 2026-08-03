package main

import (
	"fmt"

	"github.com/JosemaPereira/TrAIveler/backend/config"
	"github.com/JosemaPereira/TrAIveler/backend/internal/ai"
	"github.com/JosemaPereira/TrAIveler/backend/internal/auth"
	"github.com/JosemaPereira/TrAIveler/backend/internal/conversation"
	"github.com/JosemaPereira/TrAIveler/backend/internal/database"
	"github.com/JosemaPereira/TrAIveler/backend/internal/itinerary"
	"github.com/JosemaPereira/TrAIveler/backend/internal/subscription"
	"github.com/JosemaPereira/TrAIveler/backend/internal/trip"
)

// tripComponents bundles the composed Trip/Conversation HTTP surface
// (001-T038/001-T039/001-T040).
type tripComponents struct {
	tripHandler         *trip.Handler
	conversationHandler *conversation.Handler
}

// buildTripComponents wires the Trip and Conversation verticals from config
// and the database client: trip/conversation repositories, a second
// subscription repository (stateless, safe to construct independently of
// buildAuthComponents' own instance), the AI client (this is its first real
// caller — see internal/ai.NewAIClient's doc comment), the Itinerary
// service (satisfies conversation.ItineraryGenerator, compile-time asserted
// in internal/itinerary/service.go), and the Trip/Conversation HTTP
// handlers. Mirrors buildAuthComponents' shape (cmd/api/auth.go). Errors
// only on genuinely fatal misconfiguration (e.g. an unknown AI_PROVIDER,
// already validated by config.Load in practice).
func buildTripComponents(cfg *config.Config, db database.Client) (*tripComponents, error) {
	tripRepo := trip.NewPostgresRepository(db)
	subscriptionRepo := subscription.NewPostgresRepository(db)
	userRepo := auth.NewPostgresUserRepository(db)

	aiClient, err := ai.NewAIClient(cfg.AI)
	if err != nil {
		return nil, fmt.Errorf("build AI client: %w", err)
	}

	itineraryService := itinerary.NewService(aiClient, tripRepo)

	conversationRepo := conversation.NewPostgresRepository(db)
	conversationService := conversation.NewService(conversationRepo, itineraryService)

	tripService := trip.NewService(tripRepo, subscriptionRepo)

	return &tripComponents{
		tripHandler:         trip.NewHandler(tripService, userRepo),
		conversationHandler: conversation.NewHandler(conversationService, tripService),
	}, nil
}
