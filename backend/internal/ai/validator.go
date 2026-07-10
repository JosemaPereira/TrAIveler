package ai

import "context"

// ValidationRule is a single prompt-injection deny-list rule. Pattern is
// unused by the current stub implementation.
type ValidationRule struct {
	ID      string
	Pattern string
}

// PromptValidator screens user-supplied prompts before they are forwarded to
// an AI provider.
type PromptValidator struct {
	// rules will be loaded from config/prompt-rules.yml in a future
	// NFR-SEC-007 ticket (spec 002).
	rules []ValidationRule
}

// NewPromptValidator creates a PromptValidator with the given deny-list rules.
func NewPromptValidator(rules []ValidationRule) *PromptValidator {
	return &PromptValidator{rules: rules}
}

// Validate is a stub that always returns nil. Real prompt-injection deny-list
// rules (role switching, system prompt override, adversarial patterns) land
// in a future spec 002 (NFR) ticket — see docs/security.md "Prompt injection
// prevention".
func (v *PromptValidator) Validate(_ context.Context, _ string) error {
	return nil
}
