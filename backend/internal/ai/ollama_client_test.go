//go:build test

package ai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/internal/ai"
)

// ollamaChatWireRequest mirrors the wire shape OllamaClient sends, used by
// the fake server to decode incoming requests.
type ollamaChatWireRequest struct {
	Model    string `json:"model"`
	Format   string `json:"format"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	Stream bool `json:"stream"`
}

func newTestRequest() ai.ItineraryRequest {
	return ai.ItineraryRequest{
		TripID: "trip-123",
		ConversationHistory: []ai.Message{
			{Role: "user", Content: "Plan a 3-day trip to Lisbon"},
		},
		Preferences: ai.TravelPreferences{Budget: "moderate", GroupSize: 2},
	}
}

func TestGenerateItinerary_Success_ReturnsParsedItinerary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wireReq ollamaChatWireRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&wireReq))
		assert.Equal(t, "gemma3:4b", wireReq.Model)
		assert.Equal(t, "json", wireReq.Format)
		assert.False(t, wireReq.Stream)
		require.Len(t, wireReq.Messages, 1)
		assert.Equal(t, "user", wireReq.Messages[0].Role)

		content := `{"destinations":[{"name":"Lisbon","country":"PT"}],"days":[],"activities":[],"metadata":{"model":"","provider":"","tokens_used":0}}`
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model":             "gemma3:4b",
			"message":           map[string]string{"role": "assistant", "content": content},
			"done":              true,
			"prompt_eval_count": 10,
			"eval_count":        20,
		})
	}))
	defer server.Close()

	client := ai.NewOllamaClient(server.URL, "gemma3:4b", 5*time.Second, 3, nil)

	resp, err := client.GenerateItinerary(context.Background(), newTestRequest())

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Destinations, 1)
	assert.Equal(t, "Lisbon", resp.Destinations[0].Name)
	assert.Equal(t, "PT", resp.Destinations[0].Country)
	assert.Equal(t, "gemma3:4b", resp.Metadata.Model)
	assert.Equal(t, "ollama", resp.Metadata.Provider)
	assert.Equal(t, 30, resp.Metadata.TokensUsed)
}

func TestGenerateItinerary_MalformedContentJSON_ReturnsWrappedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model":             "gemma3:4b",
			"message":           map[string]string{"role": "assistant", "content": "not valid json"},
			"done":              true,
			"prompt_eval_count": 1,
			"eval_count":        1,
		})
	}))
	defer server.Close()

	client := ai.NewOllamaClient(server.URL, "gemma3:4b", 5*time.Second, 3, nil)

	resp, err := client.GenerateItinerary(context.Background(), newTestRequest())

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "ollama:")
}

func TestGenerateItinerary_ServerErrorThenSuccess_RetriesAndReturnsItinerary(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := atomic.AddInt32(&callCount, 1)
		if attempt <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		content := `{"destinations":[],"days":[],"activities":[],"metadata":{"model":"","provider":"","tokens_used":0}}`
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model":             "gemma3:4b",
			"message":           map[string]string{"role": "assistant", "content": content},
			"done":              true,
			"prompt_eval_count": 5,
			"eval_count":        5,
		})
	}))
	defer server.Close()

	client := ai.NewOllamaClient(server.URL, "gemma3:4b", 5*time.Second, 3, nil)

	resp, err := client.GenerateItinerary(context.Background(), newTestRequest())

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int32(3), atomic.LoadInt32(&callCount))
}

func TestGenerateItinerary_ServerErrorExhaustsRetries_ReturnsWrappedError(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := ai.NewOllamaClient(server.URL, "gemma3:4b", 5*time.Second, 2, nil)

	resp, err := client.GenerateItinerary(context.Background(), newTestRequest())

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "ollama:")
	// maxRetries=2 means 1 initial attempt + 2 retries = 3 total calls.
	assert.Equal(t, int32(3), atomic.LoadInt32(&callCount))
}

func TestGenerateItinerary_ConnectionRefused_ExhaustsRetriesReturnsWrappedError(t *testing.T) {
	// A closed server guarantees connection-refused style network errors.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	client := ai.NewOllamaClient(server.URL, "gemma3:4b", 2*time.Second, 1, nil)

	done := make(chan struct{})
	var resp *ai.ItineraryResponse
	var err error
	go func() {
		resp, err = client.GenerateItinerary(context.Background(), newTestRequest())
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("GenerateItinerary did not return in time — likely hanging")
	}

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "ollama:")
}

func TestGenerateItinerary_ContextCanceled_ReturnsPromptlyWithContextError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(300 * time.Millisecond):
		}
	}))
	defer server.Close()

	client := ai.NewOllamaClient(server.URL, "gemma3:4b", 5*time.Second, 3, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	var err error
	go func() {
		_, err = client.GenerateItinerary(ctx, newTestRequest())
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("GenerateItinerary did not return promptly after context cancellation — likely hanging")
	}

	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestStreamItinerary_HappyPath_CollectsChunksUntilDone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		lines := []string{
			`{"message":{"role":"assistant","content":"Day 1: "},"done":false}`,
			`{"message":{"role":"assistant","content":"visit Belém."},"done":false}`,
			`{"done":true,"prompt_eval_count":3,"eval_count":7}`,
		}
		for _, line := range lines {
			_, _ = fmt.Fprintln(w, line)
			flusher.Flush()
		}
	}))
	defer server.Close()

	client := ai.NewOllamaClient(server.URL, "gemma3:4b", 5*time.Second, 3, nil)

	chunkCh, errCh := client.StreamItinerary(context.Background(), newTestRequest())

	var content strings.Builder
	var sawDone bool
	for chunk := range chunkCh {
		if chunk.Done {
			sawDone = true
			continue
		}
		content.WriteString(chunk.Content)
	}

	for err := range errCh {
		t.Fatalf("unexpected error from stream: %v", err)
	}

	assert.True(t, sawDone, "expected a final Done chunk")
	assert.Equal(t, "Day 1: visit Belém.", content.String())
}

func TestStreamItinerary_MidStreamError_DeliversErrorAndClosesChannels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		_, _ = fmt.Fprintln(w, `{"message":{"role":"assistant","content":"Day 1: "},"done":false}`)
		flusher.Flush()
		_, _ = fmt.Fprintln(w, `not valid json`)
		flusher.Flush()
	}))
	defer server.Close()

	client := ai.NewOllamaClient(server.URL, "gemma3:4b", 5*time.Second, 3, nil)

	chunkCh, errCh := client.StreamItinerary(context.Background(), newTestRequest())

	var chunks []ai.StreamChunk
	for chunk := range chunkCh {
		chunks = append(chunks, chunk)
	}

	var gotErr error
	for err := range errCh {
		gotErr = err
	}

	require.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "ollama:")
	require.Len(t, chunks, 1)
	assert.Equal(t, "Day 1: ", chunks[0].Content)

	// Channels must actually be closed (range above already proves this,
	// but assert explicitly for readability of intent).
	_, chunkOpen := <-chunkCh
	_, errOpen := <-errCh
	assert.False(t, chunkOpen)
	assert.False(t, errOpen)
}

func TestStreamItinerary_ContextCanceled_ClosesChannelsPromptly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)
		_, _ = fmt.Fprintln(w, `{"message":{"role":"assistant","content":"Day 1: "},"done":false}`)
		flusher.Flush()

		select {
		case <-r.Context().Done():
		case <-time.After(300 * time.Millisecond):
		}
	}))
	defer server.Close()

	client := ai.NewOllamaClient(server.URL, "gemma3:4b", 5*time.Second, 3, nil)

	ctx, cancel := context.WithCancel(context.Background())
	chunkCh, errCh := client.StreamItinerary(ctx, newTestRequest())

	// Read the first chunk, then cancel and make sure both channels close promptly.
	<-chunkCh
	cancel()

	done := make(chan struct{})
	go func() {
		for range chunkCh {
		}
		for range errCh {
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("stream channels did not close promptly after context cancellation")
	}
}
