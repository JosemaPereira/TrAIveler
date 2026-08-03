//go:build test

package ai_test

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/config"
	"github.com/JosemaPereira/TrAIveler/backend/internal/ai"
	mocks "github.com/JosemaPereira/TrAIveler/backend/internal/ai/mocks"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// TestItineraryResponse_JSONRoundTrip_PreservesFields exercises the JSON
// tags on ItineraryResponse and its nested types, since these types will
// eventually cross an HTTP boundary in handlers built on this package.
func TestItineraryResponse_JSONRoundTrip_PreservesFields(t *testing.T) {
	original := ai.ItineraryResponse{
		Ready:        true,
		Reply:        "Your itinerary is ready!",
		Destinations: []ai.Destination{{Name: "Lisbon", Country: "PT", Region: "Lisbon District", Latitude: 38.7223, Longitude: -9.1393}},
		Days: []ai.Day{
			{
				DayNumber:   1,
				Destination: ai.Destination{Name: "Lisbon", Country: "PT", Region: "Lisbon District", Latitude: 38.7223, Longitude: -9.1393},
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

func TestNewAIClient(t *testing.T) {
	tests := []struct {
		name       string
		cfg        config.AIConfig
		wantClient ai.AIClient
		wantErrMsg string
	}{
		{
			name: "when the provider is ollama it should return an Ollama-backed client",
			cfg: config.AIConfig{
				Provider:   "ollama",
				Timeout:    5 * time.Second,
				MaxRetries: 3,
				Ollama:     config.OllamaConfig{Host: "http://localhost:11434", Model: "gemma3:4b"},
			},
			wantClient: &ai.OllamaClient{},
		},
		{
			name: "when the provider is anthropic it should return an Anthropic-backed client",
			cfg: config.AIConfig{
				Provider:   "anthropic",
				Timeout:    5 * time.Second,
				MaxRetries: 3,
				Anthropic:  config.AnthropicConfig{APIKey: "test-key", Model: "claude-3-5-sonnet-20241022"},
			},
			wantClient: &ai.AnthropicClient{},
		},
		{
			name:       "when the provider is unknown it should return an error naming the provider",
			cfg:        config.AIConfig{Provider: "unknown-provider"},
			wantErrMsg: "unknown-provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := ai.NewAIClient(tt.cfg)

			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.Nil(t, client)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, client)
			assert.IsType(t, tt.wantClient, client)
		})
	}
}

func TestTranslateError(t *testing.T) {
	t.Run("when the error is ProviderUnavailableError it should translate to a service_unavailable domain error", func(t *testing.T) {
		original := &ai.ProviderUnavailableError{RetryAfter: 45 * time.Second, Err: stderrors.New("boom")}

		translated := ai.TranslateError(original)

		var domainErr *domainerrors.DomainError
		require.ErrorAs(t, translated, &domainErr)
		assert.Equal(t, "service_unavailable", domainErr.Code)
		assert.Equal(t, 45, domainErr.Details["retry_after_seconds"])
	})

	t.Run("when ProviderUnavailableError has no RetryAfter it should use a positive default", func(t *testing.T) {
		original := &ai.ProviderUnavailableError{Err: stderrors.New("boom")}

		translated := ai.TranslateError(original)

		var domainErr *domainerrors.DomainError
		require.ErrorAs(t, translated, &domainErr)
		assert.Equal(t, "service_unavailable", domainErr.Code)
		assert.Greater(t, domainErr.Details["retry_after_seconds"], 0)
	})

	t.Run("when the error is any other error it should return it unchanged", func(t *testing.T) {
		original := stderrors.New("some other failure")

		translated := ai.TranslateError(original)

		assert.Same(t, original, translated)
	})

	t.Run("when the error is nil it should return nil", func(t *testing.T) {
		assert.NoError(t, ai.TranslateError(nil))
	})
}
