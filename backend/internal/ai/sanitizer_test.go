//go:build test

package ai_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/internal/ai"
)

// TestSanitize documents the current stub behavior: real HTML/script
// stripping (bluemonday UGCPolicy, deny <script>/<iframe>/<object>, strip
// javascript: URLs) lands in a future spec 002 (NFR-SEC-008) ticket — see
// docs/security.md. Until then every input passes through unchanged.
func TestSanitize(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "when the output contains a script tag it should currently pass through unchanged (stub)",
			input: `<script>alert('XSS')</script>`,
		},
		{
			name:  "when the output is an empty string it should return an empty string",
			input: "",
		},
		{
			name:  "when the output is plain text it should pass through unchanged",
			input: "Visit the Belém Tower on day 1.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizer := ai.NewOutputSanitizer()

			output, err := sanitizer.Sanitize(context.Background(), tt.input)

			require.NoError(t, err)
			assert.Equal(t, tt.input, output)
		})
	}
}
