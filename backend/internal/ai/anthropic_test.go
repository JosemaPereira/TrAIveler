//go:build test

package ai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/JosemaPereira/TrAIveler/backend/internal/ai"
)

// anthropicMessageWireRequest mirrors the wire shape AnthropicClient sends,
// used by the fake server to decode incoming requests.
type anthropicMessageWireRequest struct {
	Model  string `json:"model"`
	System []struct {
		Text string `json:"text"`
	} `json:"system"`
	Messages []struct {
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"messages"`
}

// writeAnthropicError writes a valid Anthropic-shaped error envelope, since
// the SDK requires one to construct a distinguishable *anthropic.Error
// rather than failing to unmarshal the response body. Retry-After: 0
// collapses the SDK's built-in exponential backoff between attempts to
// (near) zero, keeping retry tests fast and deterministic.
func writeAnthropicError(w http.ResponseWriter, status int, errType string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "0")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type": "error",
		"error": map[string]string{
			"type":    errType,
			"message": "fake server error",
		},
	})
}

// writeAnthropicSuccess writes a valid Anthropic Messages API success body
// whose sole text content block is itineraryJSON.
func writeAnthropicSuccess(w http.ResponseWriter, model, itineraryJSON string, inputTokens, outputTokens int) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":   "msg_fake123",
		"type": "message",
		"role": "assistant",
		"content": []map[string]any{
			{"type": "text", "text": itineraryJSON},
		},
		"model":         model,
		"stop_reason":   "end_turn",
		"stop_sequence": nil,
		"usage": map[string]any{
			"input_tokens":  inputTokens,
			"output_tokens": outputTokens,
		},
	})
}

func TestAnthropicGenerateItinerary_Success_ReturnsParsedItinerary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wireReq anthropicMessageWireRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&wireReq))
		assert.Equal(t, "claude-3-5-sonnet-20241022", wireReq.Model)
		require.Len(t, wireReq.Messages, 1)
		assert.Equal(t, "user", wireReq.Messages[0].Role)

		content := `{"destinations":[{"name":"Lisbon","country":"PT"}],"days":[],"activities":[],"metadata":{"model":"","provider":"","tokens_used":0}}`
		writeAnthropicSuccess(w, "claude-3-5-sonnet-20241022", content, 10, 20)
	}))
	defer server.Close()

	client := ai.NewAnthropicClient("test-key", "claude-3-5-sonnet-20241022", 5*time.Second, 3,
		option.WithBaseURL(server.URL), option.WithHTTPClient(server.Client()))

	resp, err := client.GenerateItinerary(context.Background(), newTestRequest(t))

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Destinations, 1)
	assert.Equal(t, "Lisbon", resp.Destinations[0].Name)
	assert.Equal(t, "PT", resp.Destinations[0].Country)
	assert.Equal(t, "claude-3-5-sonnet-20241022", resp.Metadata.Model)
	assert.Equal(t, "anthropic", resp.Metadata.Provider)
	assert.Equal(t, 30, resp.Metadata.TokensUsed)
}

func TestAnthropicGenerateItinerary_RateLimitedThenSucceeds_RetriesAndReturnsItinerary(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempt := atomic.AddInt32(&callCount, 1)
		if attempt <= 2 {
			writeAnthropicError(w, http.StatusTooManyRequests, "rate_limit_error")
			return
		}

		content := `{"destinations":[],"days":[],"activities":[],"metadata":{"model":"","provider":"","tokens_used":0}}`
		writeAnthropicSuccess(w, "claude-3-5-sonnet-20241022", content, 5, 5)
	}))
	defer server.Close()

	client := ai.NewAnthropicClient("test-key", "claude-3-5-sonnet-20241022", 5*time.Second, 3,
		option.WithBaseURL(server.URL), option.WithHTTPClient(server.Client()))

	resp, err := client.GenerateItinerary(context.Background(), newTestRequest(t))

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int32(3), atomic.LoadInt32(&callCount))
}

func TestAnthropicGenerateItinerary_ServiceUnavailableExhaustsRetries_ReturnsProviderUnavailableError(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&callCount, 1)
		writeAnthropicError(w, http.StatusServiceUnavailable, "overloaded_error")
	}))
	defer server.Close()

	client := ai.NewAnthropicClient("test-key", "claude-3-5-sonnet-20241022", 5*time.Second, 2,
		option.WithBaseURL(server.URL), option.WithHTTPClient(server.Client()))

	resp, err := client.GenerateItinerary(context.Background(), newTestRequest(t))

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ai.ErrProviderUnavailable)

	var unavailableErr *ai.ProviderUnavailableError
	require.ErrorAs(t, err, &unavailableErr)

	// maxRetries=2 means 1 initial attempt + 2 retries = 3 total calls,
	// mirroring OllamaClient's retry-exhaustion test convention.
	assert.Equal(t, int32(3), atomic.LoadInt32(&callCount))
}

func TestAnthropicGenerateItinerary_ClientTimeoutExceeded_ReturnsPromptContextDeadlineError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(300 * time.Millisecond):
		}
	}))
	defer server.Close()

	// maxRetries=0 and a short client-level request timeout: once that
	// per-attempt timeout fires, the SDK returns ctx.Err() immediately
	// without retrying (see anthropic-sdk-go's internal/requestconfig),
	// so this also asserts the call doesn't hang waiting on retries.
	client := ai.NewAnthropicClient("test-key", "claude-3-5-sonnet-20241022", 30*time.Millisecond, 0,
		option.WithBaseURL(server.URL), option.WithHTTPClient(server.Client()))

	done := make(chan struct{})
	var resp *ai.ItineraryResponse
	var err error
	go func() {
		resp, err = client.GenerateItinerary(context.Background(), newTestRequest(t))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("GenerateItinerary did not return promptly after client timeout — likely hanging")
	}

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestAnthropicGenerateItinerary_WithSystemPrompt_ForwardsSystemField(t *testing.T) {
	var gotSystem []struct {
		Text string `json:"text"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wireReq anthropicMessageWireRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&wireReq))
		gotSystem = wireReq.System

		content := `{"destinations":[],"days":[],"activities":[],"metadata":{"model":"","provider":"","tokens_used":0}}`
		writeAnthropicSuccess(w, "claude-3-5-sonnet-20241022", content, 1, 1)
	}))
	defer server.Close()

	client := ai.NewAnthropicClient("test-key", "claude-3-5-sonnet-20241022", 5*time.Second, 3,
		option.WithBaseURL(server.URL), option.WithHTTPClient(server.Client()))

	req := newTestRequest(t)
	req.SystemPrompt = "You are a travel-planning assistant."

	_, err := client.GenerateItinerary(context.Background(), req)

	require.NoError(t, err)
	require.Len(t, gotSystem, 1, "the system prompt must be forwarded as a single system text block")
	assert.Equal(t, "You are a travel-planning assistant.", gotSystem[0].Text)
}

func TestAnthropicGenerateItinerary_EmptySystemPrompt_OmitsSystemField(t *testing.T) {
	var gotSystem []struct {
		Text string `json:"text"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wireReq anthropicMessageWireRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&wireReq))
		gotSystem = wireReq.System

		content := `{"destinations":[],"days":[],"activities":[],"metadata":{"model":"","provider":"","tokens_used":0}}`
		writeAnthropicSuccess(w, "claude-3-5-sonnet-latest", content, 1, 1)
	}))
	defer server.Close()

	client := ai.NewAnthropicClient("test-key", "claude-3-5-sonnet-20241022", 5*time.Second, 3,
		option.WithBaseURL(server.URL), option.WithHTTPClient(server.Client()))

	_, err := client.GenerateItinerary(context.Background(), newTestRequest(t))

	require.NoError(t, err)
	assert.Empty(t, gotSystem, "an empty SystemPrompt must not add a system field, preserving existing behavior")
}

func TestAnthropicStreamItinerary_WithSystemPrompt_ForwardsSystemField(t *testing.T) {
	var gotSystem []struct {
		Text string `json:"text"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wireReq anthropicMessageWireRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&wireReq))
		gotSystem = wireReq.System

		flusher, ok := w.(http.Flusher)
		require.True(t, ok)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"))
		flusher.Flush()
	}))
	defer server.Close()

	client := ai.NewAnthropicClient("test-key", "claude-3-5-sonnet-20241022", 5*time.Second, 3,
		option.WithBaseURL(server.URL), option.WithHTTPClient(server.Client()))

	req := newTestRequest(t)
	req.SystemPrompt = "You are a travel-planning assistant."

	chunkCh, errCh := client.StreamItinerary(context.Background(), req)
	//nolint:revive // intentional drain loop: block until the channel closes, body has nothing to do
	for range chunkCh {
	}
	for err := range errCh {
		t.Fatalf("unexpected error from stream: %v", err)
	}

	require.Len(t, gotSystem, 1, "the system prompt must be forwarded as a single system text block")
	assert.Equal(t, "You are a travel-planning assistant.", gotSystem[0].Text)
}

func TestAnthropicStreamItinerary_HappyPath_CollectsChunksUntilDone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		events := []struct {
			eventType string
			data      string
		}{
			{"message_start", `{"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","model":"claude-3-5-sonnet-20241022","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":5,"output_tokens":0}}}`},
			{"content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`},
			{"content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Day 1: "}}`},
			{"content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"visit Belém."}}`},
			{"content_block_stop", `{"type":"content_block_stop","index":0}`},
			{"message_delta", `{"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":7}}`},
			{"message_stop", `{"type":"message_stop"}`},
		}
		for _, event := range events {
			_, _ = w.Write([]byte("event: " + event.eventType + "\n"))
			_, _ = w.Write([]byte("data: " + event.data + "\n\n"))
			flusher.Flush()
		}
	}))
	defer server.Close()

	client := ai.NewAnthropicClient("test-key", "claude-3-5-sonnet-20241022", 5*time.Second, 3,
		option.WithBaseURL(server.URL), option.WithHTTPClient(server.Client()))

	chunkCh, errCh := client.StreamItinerary(context.Background(), newTestRequest(t))

	var content string
	var sawDone bool
	for chunk := range chunkCh {
		if chunk.Done {
			sawDone = true
			continue
		}
		content += chunk.Content
	}

	for err := range errCh {
		t.Fatalf("unexpected error from stream: %v", err)
	}

	assert.True(t, sawDone, "expected a final Done chunk")
	assert.Equal(t, "Day 1: visit Belém.", content)
}
