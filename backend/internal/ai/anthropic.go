package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
)

// anthropicMaxTokens bounds the length of a single non-streaming itinerary
// generation response. This is a first-pass fixed value (not yet
// config-driven); tuning it against real itinerary sizes, and any
// structured-output/tool-forcing to *guarantee* schema-valid JSON (today's
// schema compliance is prompt-instructed only — see internal/itinerary's
// SystemPrompt — not enforced by the API), remain unimplemented follow-up
// work, not yet assigned to a ticket.
const anthropicMaxTokens = 4096

// ErrProviderUnavailable is the sentinel identifying an AIClient failure
// caused by the provider staying transiently unavailable (HTTP 429 "rate
// limited" or 503 "service unavailable") after every configured retry
// attempt has been exhausted. Callers use errors.Is against this sentinel
// (or errors.As against *ProviderUnavailableError, to also recover the
// suggested retry delay) to distinguish this from other, non-retryable
// AIClient failures — see the wrapping helper alongside NewAIClient in
// client.go (docs/roadmap.md 005-T113), which turns it into a 503 response
// with a Retry-After header.
var ErrProviderUnavailable = errors.New("ai: provider unavailable after exhausting retries")

// ProviderUnavailableError wraps ErrProviderUnavailable with the provider's
// suggested retry delay, when the failing response carried a Retry-After
// header; RetryAfter is zero when none was present, leaving the caller free
// to fall back to its own default.
type ProviderUnavailableError struct {
	RetryAfter time.Duration
	Err        error
}

// Error implements the error interface.
func (e *ProviderUnavailableError) Error() string {
	return fmt.Sprintf("%s: %v", ErrProviderUnavailable, e.Err)
}

// Unwrap exposes the underlying transport/API error for logging and
// errors.As, without making that error part of the ErrProviderUnavailable
// identity check (see Is below).
func (e *ProviderUnavailableError) Unwrap() error {
	return e.Err
}

// Is lets errors.Is(err, ErrProviderUnavailable) succeed for a
// *ProviderUnavailableError without requiring Err itself to wrap the
// sentinel.
func (e *ProviderUnavailableError) Is(target error) bool {
	return target == ErrProviderUnavailable
}

// AnthropicClient implements AIClient against the official Anthropic Claude
// API via github.com/anthropics/anthropic-sdk-go. It is the concrete
// backend used in staging/production, per product decision
// (AI_PROVIDER=anthropic); local development and MVP testing use the free
// local OllamaClient instead (ollama_client.go) — see that file's doc
// comment and backend/README.md for the split.
type AnthropicClient struct {
	client anthropic.Client
	model  string
}

// NewAnthropicClient creates an AIClient backed by the Anthropic Claude API.
// apiKey/model come from config.AIConfig.Anthropic; timeout bounds each
// individual request attempt (config.AIConfig.Timeout, 60s by default).
// maxRetries is the number of additional attempts made after the first, on
// transient errors (429/5xx) — mirrors NewOllamaClient's parameter, but the
// actual retrying/backoff is delegated to the SDK's own built-in support
// (option.WithMaxRetries) rather than a hand-rolled loop, since that is the
// idiomatic mechanism for this SDK (its internal/requestconfig package
// already implements exponential backoff and honors a Retry-After response
// header). opts, if given, are appended after the defaults, letting tests
// inject option.WithBaseURL/option.WithHTTPClient to point the client at a
// fake server — the equivalent, in this SDK's idiomatic construction style,
// of the optional *http.Client parameter NewOllamaClient accepts.
func NewAnthropicClient(apiKey, model string, timeout time.Duration, maxRetries int, opts ...option.RequestOption) *AnthropicClient {
	clientOpts := append([]option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithRequestTimeout(timeout),
		option.WithMaxRetries(maxRetries),
	}, opts...)

	return &AnthropicClient{
		client: anthropic.NewClient(clientOpts...),
		model:  model,
	}
}

// buildMessages converts the request's conversation history into Anthropic
// message params. Anthropic's Messages API only recognizes "user" and
// "assistant" turns (a system prompt is a separate top-level field, not a
// message role — see buildSystem, which handles req.SystemPrompt) — any
// history role other than "assistant" is sent as a user turn.
// Itinerary-specific system prompt *content*, tool-calling, and
// trip-generation business logic are out of scope here and belong to
// internal/itinerary (001-T037/issue #236), mirroring OllamaClient.buildMessages'
// same scope limit.
func (c *AnthropicClient) buildMessages(req ItineraryRequest) []anthropic.MessageParam {
	messages := make([]anthropic.MessageParam, 0, len(req.ConversationHistory))
	for _, m := range req.ConversationHistory {
		block := anthropic.NewTextBlock(m.Content)
		if m.Role == "assistant" {
			messages = append(messages, anthropic.NewAssistantMessage(block))
		} else {
			messages = append(messages, anthropic.NewUserMessage(block))
		}
	}
	return messages
}

// buildSystem returns the system prompt param for req, or nil when
// req.SystemPrompt is empty — Anthropic's Messages API takes the system
// prompt as a separate top-level field (MessageNewParams.System), not a
// message role, unlike buildMessages' user/assistant turns.
func (c *AnthropicClient) buildSystem(req ItineraryRequest) []anthropic.TextBlockParam {
	if req.SystemPrompt == "" {
		return nil
	}
	return []anthropic.TextBlockParam{{Text: req.SystemPrompt}}
}

// wrapError translates a raw SDK error into this package's error
// vocabulary: a *ProviderUnavailableError (wrapping ErrProviderUnavailable)
// for a 429/503 API error that survived every retry attempt, or an
// otherwise-opaque wrapped error for everything else — including context
// cancellation/timeouts, which errors.Is(err, context.DeadlineExceeded)
// still sees through the %w wrap below.
func (c *AnthropicClient) wrapError(err error) error {
	var apiErr *anthropic.Error
	if errors.As(err, &apiErr) &&
		(apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode == http.StatusServiceUnavailable) {
		return &ProviderUnavailableError{
			RetryAfter: retryAfter(apiErr.Response),
			Err:        err,
		}
	}
	return fmt.Errorf("anthropic: %w", err)
}

// closeStream closes a streaming response, logging (rather than dropping)
// any close error — mirrors closeBody's rationale in ollama_client.go: the
// stream has already succeeded or failed by this point, so a close failure
// is never actionable by the caller.
func closeStream(stream *ssestream.Stream[anthropic.MessageStreamEventUnion]) {
	if err := stream.Close(); err != nil {
		slog.Warn("anthropic: failed to close stream", "error", err)
	}
}

// retryAfter reads a Retry-After response header (seconds form only — the
// HTTP-date form is rare for this API and not needed yet) and returns the
// suggested delay, or zero if absent/unparseable.
func retryAfter(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}
	seconds, err := strconv.Atoi(resp.Header.Get("Retry-After"))
	if err != nil || seconds < 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

// GenerateItinerary creates a complete trip itinerary from conversation
// history via a single (non-streaming) Anthropic Messages API call. Like
// OllamaClient.GenerateItinerary, the model's response text is parsed
// directly as itinerary JSON — the response shape is instructed via
// req.SystemPrompt (see internal/itinerary), not enforced by the API's own
// structured-output/tool-forcing features (see anthropicMaxTokens' comment).
func (c *AnthropicClient) GenerateItinerary(ctx context.Context, req ItineraryRequest) (*ItineraryResponse, error) {
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: anthropicMaxTokens,
		System:    c.buildSystem(req),
		Messages:  c.buildMessages(req),
	})
	if err != nil {
		return nil, c.wrapError(err)
	}

	if len(message.Content) == 0 || message.Content[0].Text == "" {
		return nil, fmt.Errorf("anthropic: empty response content")
	}

	var itinerary ItineraryResponse
	if err := json.Unmarshal([]byte(message.Content[0].Text), &itinerary); err != nil {
		return nil, fmt.Errorf("anthropic: parse itinerary JSON content: %w", err)
	}

	itinerary.Metadata = ResponseMetadata{
		Model:      message.Model,
		Provider:   "anthropic",
		TokensUsed: int(message.Usage.InputTokens + message.Usage.OutputTokens),
	}

	return &itinerary, nil
}

// StreamItinerary generates an itinerary with streaming for real-time
// feedback. It runs the request in a goroutine so the channels are returned
// immediately, mirroring OllamaClient.StreamItinerary; Anthropic streams
// Server-Sent Events, of which only content_block_delta text deltas and the
// terminal message_stop are relevant to this interface's StreamChunk shape.
func (c *AnthropicClient) StreamItinerary(ctx context.Context, req ItineraryRequest) (<-chan StreamChunk, <-chan error) {
	chunkCh := make(chan StreamChunk)
	errCh := make(chan error, 1)

	go func() {
		defer close(chunkCh)
		defer close(errCh)

		stream := c.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
			Model:     c.model,
			MaxTokens: anthropicMaxTokens,
			System:    c.buildSystem(req),
			Messages:  c.buildMessages(req),
		})
		defer closeStream(stream)

		for stream.Next() {
			event := stream.Current()
			switch event.Type {
			case "content_block_delta":
				if event.Delta.Text != "" {
					chunkCh <- StreamChunk{Content: event.Delta.Text, Done: false}
				}
			case "message_stop":
				chunkCh <- StreamChunk{Done: true}
				return
			}
		}

		if err := stream.Err(); err != nil {
			errCh <- c.wrapError(err)
		}
	}()

	return chunkCh, errCh
}
