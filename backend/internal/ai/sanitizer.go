package ai

import (
	"context"

	"github.com/microcosm-cc/bluemonday"
)

// OutputSanitizer strips unsafe HTML/script content from AI-generated
// output before it reaches a client.
type OutputSanitizer struct {
	// policy is unused by the current stub; the real policy is wired in a
	// future spec 002 ticket.
	policy *bluemonday.Policy
}

// NewOutputSanitizer creates an OutputSanitizer. The policy is pre-built
// (bluemonday.StrictPolicy) so the field is real and ready to use once
// Sanitize gains its real implementation; it is not yet applied below.
func NewOutputSanitizer() *OutputSanitizer {
	return &OutputSanitizer{policy: bluemonday.StrictPolicy()}
}

// Sanitize is a pass-through stub. Real HTML/script stripping (bluemonday
// UGCPolicy, deny <script>/<iframe>/<object>, strip javascript: URLs) lands
// in a future spec 002 (NFR-SEC-008) ticket — see docs/security.md.
func (s *OutputSanitizer) Sanitize(_ context.Context, output string) (string, error) {
	return output, nil
}
