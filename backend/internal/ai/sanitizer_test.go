//go:build test

package ai_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/internal/ai"
)

// TestSanitize_ScriptTag_PassesThrough documents the current stub behavior:
// real HTML/script stripping (bluemonday UGCPolicy, deny <script>/<iframe>/
// <object>, strip javascript: URLs) lands in a future spec 002 (NFR-SEC-008)
// ticket — see docs/security.md.
func TestSanitize_ScriptTag_PassesThrough(t *testing.T) {
	sanitizer := ai.NewOutputSanitizer()
	input := `<script>alert('XSS')</script>`

	output, err := sanitizer.Sanitize(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, input, output)
}

// TestSanitize_EmptyString_ReturnsEmptyString covers the empty-input edge case.
func TestSanitize_EmptyString_ReturnsEmptyString(t *testing.T) {
	sanitizer := ai.NewOutputSanitizer()

	output, err := sanitizer.Sanitize(context.Background(), "")

	require.NoError(t, err)
	assert.Equal(t, "", output)
}

// TestSanitize_PlainText_PassesThrough covers the everyday, non-adversarial path.
func TestSanitize_PlainText_PassesThrough(t *testing.T) {
	sanitizer := ai.NewOutputSanitizer()
	input := "Visit the Belém Tower on day 1."

	output, err := sanitizer.Sanitize(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, input, output)
}
