package observability

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCorrelationID_ReturnsParseableUUID(t *testing.T) {
	id := GenerateCorrelationID()

	_, err := uuid.Parse(id)
	require.NoError(t, err)
}

func TestGenerateCorrelationID_TwoCalls_ReturnDifferentValues(t *testing.T) {
	first := GenerateCorrelationID()
	second := GenerateCorrelationID()

	assert.NotEqual(t, first, second)
}
