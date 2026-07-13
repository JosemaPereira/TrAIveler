package ai

import (
	"context"
	"errors"
	"fmt"

	"github.com/JosemaPereira/TrAIveler/backend/config"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// AIClient generates travel itineraries from a conversation history. It is
// the stable contract implementations must satisfy: OllamaClient (local
// development, see ollama_client.go) and AnthropicClient (staging/production,
// see anthropic.go) both implement it — see NewAIClient below for how a
// caller picks between them.
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

// defaultRetryAfterSeconds is the Retry-After hint TranslateError uses when
// a *ProviderUnavailableError didn't carry a provider-suggested delay
// (e.g. no Retry-After header was present on the response that finally
// exhausted retries) — a reasonable "try again shortly" default rather
// than omitting the header entirely.
const defaultRetryAfterSeconds = 30

// NewAIClient selects and constructs the concrete AIClient implementation
// for cfg.Provider ("ollama" | "anthropic", already validated by
// config.Load — see config.go's validate), mirroring the exhaustive
// switch-with-error-default style that function itself uses. This is the
// only place in the codebase that should choose between providers: callers
// depend on the AIClient interface, never on OllamaClient/AnthropicClient
// directly, so switching AI_PROVIDER never requires touching caller code.
// Per the local-dev-is-Ollama, staging/production-is-Anthropic split
// documented in ollama_client.go/backend/README.md, neither branch is
// treated as the "default" case — both are equally first-class.
func NewAIClient(cfg config.AIConfig) (AIClient, error) {
	switch cfg.Provider {
	case "ollama":
		return NewOllamaClient(cfg.Ollama.Host, cfg.Ollama.Model, cfg.Timeout, cfg.MaxRetries, nil), nil
	case "anthropic":
		return NewAnthropicClient(cfg.Anthropic.APIKey, cfg.Anthropic.Model, cfg.Timeout, cfg.MaxRetries), nil
	default:
		return nil, fmt.Errorf("ai: unknown AI_PROVIDER %q", cfg.Provider)
	}
}

// TranslateError maps an AIClient failure into this codebase's HTTP error
// vocabulary: a *ProviderUnavailableError (see anthropic.go — retries
// exhausted against a 429/503 response) becomes a "service_unavailable"
// *errors.DomainError carrying a Retry-After hint, ready to pass straight
// to errors.HandleError. Any other error (including nil) is returned
// unchanged. AnthropicClient/OllamaClient deliberately do not import
// internal/errors themselves, to stay transport-agnostic (neither is
// HTTP-specific — see their own doc comments); this factory-adjacent
// helper is the seam where an AIClient error crosses into the HTTP error
// vocabulary, per docs/roadmap.md 005-T113. Callers are expected to run
// every AIClient error through this before handing it to
// errors.HandleError, the same way they'd use the errors package's other
// constructors (NotFound, Conflict, etc.) directly for their own failures.
func TranslateError(err error) error {
	var unavailable *ProviderUnavailableError
	if errors.As(err, &unavailable) {
		retryAfterSeconds := int(unavailable.RetryAfter.Seconds())
		if retryAfterSeconds <= 0 {
			retryAfterSeconds = defaultRetryAfterSeconds
		}
		return domainerrors.ServiceUnavailable(retryAfterSeconds)
	}
	return err
}
