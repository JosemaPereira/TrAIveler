//go:build test

package ai_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/internal/ai"
	mocks "github.com/JosemaPereira/TrAIveler/backend/internal/ai/mocks"
)

// TestItineraryResponse_JSONRoundTrip_PreservesFields exercises the JSON
// tags on ItineraryResponse and its nested types, since these types will
// eventually cross an HTTP boundary in handlers built on this package.
func TestItineraryResponse_JSONRoundTrip_PreservesFields(t *testing.T) {
	original := ai.ItineraryResponse{
		Destinations: []ai.Destination{{Name: "Lisbon", Country: "PT"}},
		Days: []ai.Day{
			{
				DayNumber:   1,
				Destination: ai.Destination{Name: "Lisbon", Country: "PT"},
				Activities: []ai.Activity{
					{Title: "Belém Tower", Type: "visit", Description: "Historic tower", SequenceOrder: 1},
				},
			},
		},
		Activities: []ai.Activity{
			{Title: "Belém Tower", Type: "visit", Description: "Historic tower", SequenceOrder: 1},
		},
		Metadata: ai.ResponseMetadata{Model: "gemma3:4b", TokensUsed: 42, Provider: "ollama"},
	}

	raw, err := json.Marshal(original)
	require.NoError(t, err)

	var roundTripped ai.ItineraryResponse
	require.NoError(t, json.Unmarshal(raw, &roundTripped))

	assert.Equal(t, original, roundTripped)
}

// TestAIClient_MockContract_GenerateAndStreamItinerary drives both AIClient
// methods through the generated mock to prove the interface, request, and
// response types compile and wire together as expected.
func TestAIClient_MockContract_GenerateAndStreamItinerary(t *testing.T) {
	client := mocks.NewMockAIClient(t)
	ctx := context.Background()
	req := ai.ItineraryRequest{
		TripID: "trip-123",
		ConversationHistory: []ai.Message{
			{Role: "user", Content: "Plan a 3-day trip to Lisbon"},
		},
		Preferences: ai.TravelPreferences{
			Budget:       "moderate",
			Pace:         "relaxed",
			GroupSize:    2,
			TravelStyles: []string{"gastronomy", "museums-and-art"},
			StartDate:    "2026-08-01",
			EndDate:      "2026-08-03",
		},
	}
	wantResponse := &ai.ItineraryResponse{
		Destinations: []ai.Destination{{Name: "Lisbon", Country: "PT"}},
		Metadata:     ai.ResponseMetadata{Model: "gemma3:4b", Provider: "ollama"},
	}
	client.EXPECT().GenerateItinerary(ctx, req).Return(wantResponse, nil).Once()

	gotResponse, err := client.GenerateItinerary(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, wantResponse, gotResponse)

	chunkCh := make(chan ai.StreamChunk, 1)
	errCh := make(chan error, 1)
	chunkCh <- ai.StreamChunk{Content: "Day 1: ", Done: false}
	close(chunkCh)
	close(errCh)
	client.EXPECT().StreamItinerary(ctx, req).Return(chunkCh, errCh).Once()

	gotChunks, gotErrs := client.StreamItinerary(ctx, req)
	var collected []ai.StreamChunk
	for chunk := range gotChunks {
		collected = append(collected, chunk)
	}
	for err := range gotErrs {
		t.Fatalf("unexpected error from stream: %v", err)
	}

	require.Len(t, collected, 1)
	assert.Equal(t, "Day 1: ", collected[0].Content)
}
