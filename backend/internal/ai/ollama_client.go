package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// retryBackoffUnit is the base delay between retry attempts; the actual
// delay grows linearly with the attempt number to give a transient failure
// (e.g. the local Ollama server still starting up) a little more room.
const retryBackoffUnit = 50 * time.Millisecond

// OllamaClient implements AIClient against a local Ollama server's native
// /api/chat HTTP endpoint (not the Anthropic SDK). It is the concrete
// backend used for local development and MVP testing, per product decision;
// AnthropicClient (anthropic.go) is the equivalent implementation used in
// staging/production — see NewAIClient in client.go for how a caller picks
// between the two via config.AIConfig.Provider.
type OllamaClient struct {
	httpClient *http.Client
	host       string
	model      string
	maxRetries int
}

// ollamaMessage is the wire shape of a single chat message in Ollama's API.
type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ollamaChatRequest is the wire shape of a POST /api/chat request body.
type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Format   string          `json:"format,omitempty"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

// ollamaChatResponse is the wire shape of a single /api/chat response
// object — the full body for a non-streaming call, or one line of a
// streaming call's newline-delimited body.
type ollamaChatResponse struct {
	Model           string        `json:"model"`
	Message         ollamaMessage `json:"message"`
	Done            bool          `json:"done"`
	PromptEvalCount int           `json:"prompt_eval_count"`
	EvalCount       int           `json:"eval_count"`
}

// NewOllamaClient creates an AIClient backed by a local Ollama server.
// host is the server's base URL (e.g. "http://localhost:11434"); model is
// the Ollama model tag to request (e.g. "gemma3:4b"). If httpClient is nil,
// a default client scoped to timeout is used. maxRetries is the number of
// additional attempts made after the first, on transient errors (connection
// failures, 5xx responses).
func NewOllamaClient(host, model string, timeout time.Duration, maxRetries int, httpClient *http.Client) *OllamaClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	return &OllamaClient{
		host:       host,
		model:      model,
		maxRetries: maxRetries,
		httpClient: httpClient,
	}
}

// buildMessages converts the request's conversation history into Ollama
// chat messages, prepending req.SystemPrompt as a leading role:"system"
// message when non-empty — Ollama's /api/chat endpoint accepts a system
// message natively, unlike Anthropic's separate top-level System field (see
// AnthropicClient.buildSystem). Beyond that, this is a pure forwarding step
// — itinerary-specific prompt content, tool-calling, and trip-generation
// business logic are out of scope here and belong to internal/itinerary
// (001-T037/issue #236).
func (c *OllamaClient) buildMessages(req ItineraryRequest) []ollamaMessage {
	messages := make([]ollamaMessage, 0, len(req.ConversationHistory)+1)
	if req.SystemPrompt != "" {
		messages = append(messages, ollamaMessage{Role: "system", Content: req.SystemPrompt})
	}
	for _, m := range req.ConversationHistory {
		messages = append(messages, ollamaMessage(m))
	}
	return messages
}

// closeBody closes an HTTP response body, logging (rather than dropping)
// any close error — the request has already succeeded or failed by this
// point, so a close failure is never actionable by the caller.
func closeBody(resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		slog.Warn("ollama: failed to close response body", "error", err)
	}
}

// doChatRequest issues a POST /api/chat request, retrying transient
// failures (connection errors, 5xx responses) up to c.maxRetries additional
// times with a linear backoff. It returns the raw HTTP response on a
// non-error, non-5xx status, leaving the caller responsible for closing the
// response body.
func (c *OllamaClient) doChatRequest(ctx context.Context, req ItineraryRequest, stream bool) (*http.Response, error) {
	body, err := json.Marshal(ollamaChatRequest{
		Model:    c.model,
		Format:   "json",
		Messages: c.buildMessages(req),
		Stream:   stream,
	})
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal request body: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * retryBackoffUnit
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("ollama: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		resp, err := c.sendChatRequest(ctx, body)
		if err != nil {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("ollama: %w", ctx.Err())
			}
			lastErr = err
			continue
		}

		if resp.StatusCode >= http.StatusInternalServerError {
			lastErr = fmt.Errorf("ollama: server error: status %d", resp.StatusCode)
			closeBody(resp)
			continue
		}

		if resp.StatusCode >= http.StatusBadRequest {
			closeBody(resp)
			return nil, fmt.Errorf("ollama: unexpected status %d", resp.StatusCode)
		}

		return resp, nil
	}

	return nil, fmt.Errorf("ollama: request failed after %d attempt(s): %w", c.maxRetries+1, lastErr)
}

// sendChatRequest performs a single HTTP round trip to the Ollama /api/chat
// endpoint, respecting ctx cancellation.
func (c *OllamaClient) sendChatRequest(ctx context.Context, body []byte) (*http.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.host+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(httpReq)
}

// GenerateItinerary creates a complete trip itinerary from conversation
// history via a single (non-streaming) Ollama /api/chat call. Ollama's
// format: "json" field forces the model to return valid JSON, which is then
// unmarshaled into an ItineraryResponse.
func (c *OllamaClient) GenerateItinerary(ctx context.Context, req ItineraryRequest) (*ItineraryResponse, error) {
	resp, err := c.doChatRequest(ctx, req, false)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	var envelope ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("ollama: decode response envelope: %w", err)
	}

	var itinerary ItineraryResponse
	if err := json.Unmarshal([]byte(envelope.Message.Content), &itinerary); err != nil {
		return nil, fmt.Errorf("ollama: parse itinerary JSON content: %w", err)
	}

	itinerary.Metadata = ResponseMetadata{
		Model:      c.model,
		TokensUsed: envelope.PromptEvalCount + envelope.EvalCount,
		Provider:   "ollama",
	}

	return &itinerary, nil
}

// StreamItinerary generates an itinerary with streaming for real-time
// feedback. It runs the request in a goroutine so the channels are returned
// immediately; Ollama streams newline-delimited JSON objects, each decoded
// and forwarded as a StreamChunk until a "done": true line is seen.
func (c *OllamaClient) StreamItinerary(ctx context.Context, req ItineraryRequest) (<-chan StreamChunk, <-chan error) {
	chunkCh := make(chan StreamChunk)
	errCh := make(chan error, 1)

	go func() {
		defer close(chunkCh)
		defer close(errCh)

		resp, err := c.doChatRequest(ctx, req, true)
		if err != nil {
			errCh <- err
			return
		}
		defer closeBody(resp)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var envelope ollamaChatResponse
			if err := json.Unmarshal(line, &envelope); err != nil {
				errCh <- fmt.Errorf("ollama: decode stream chunk: %w", err)
				return
			}

			if envelope.Done {
				chunkCh <- StreamChunk{Done: true}
				return
			}

			chunkCh <- StreamChunk{Content: envelope.Message.Content, Done: false}
		}

		if err := scanner.Err(); err != nil {
			errCh <- fmt.Errorf("ollama: read stream: %w", err)
		}
	}()

	return chunkCh, errCh
}
