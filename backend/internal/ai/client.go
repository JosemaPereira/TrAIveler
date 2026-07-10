package ai

import "context"

// AIClient generates travel itineraries from a conversation history. It is
// the stable contract implementations must satisfy — currently OllamaClient
// (local development, see ollama_client.go); an Anthropic-backed
// implementation for staging/production lands in a future ticket
// (docs/roadmap.md 005-T112) without needing to change this interface.
//
// Name is spec-mandated (see docs/roadmap.md 005-T030) and pins the
// generated mock's filename/constructor (mocks.NewMockAIClient); renaming
// to avoid the ai.AIClient stutter would break that convention.
//
//nolint:revive // stutters (ai.AIClient) by design; see comment above.
type AIClient interface {
	// GenerateItinerary creates a complete trip itinerary from conversation history.
	GenerateItinerary(ctx context.Context, req ItineraryRequest) (*ItineraryResponse, error)

	// StreamItinerary generates an itinerary with streaming for real-time feedback.
	// The error channel receives at most one error and is then closed; the chunk channel
	// is closed when generation completes or fails.
	StreamItinerary(ctx context.Context, req ItineraryRequest) (<-chan StreamChunk, <-chan error)
}
