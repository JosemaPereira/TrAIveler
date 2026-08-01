package testdb_test

import (
	"testing"

	"github.com/JosemaPereira/TrAIveler/backend/internal/testdb"
	"github.com/stretchr/testify/assert"
)

// TestPostgresImage_MatchesRuntimePin guards the single shared constant that
// keeps every testcontainer-backed integration test pinned to the same
// PostgreSQL image as docker-compose.yml / the planned RDS engine_version
// (see docs/architecture.md and patterns-discovered.md's "A Version Pin Can
// Be Deliberate Architecture, Not Drift"). If this ever needs to change, it
// should change here once, not independently in five test files.
func TestPostgresImage_MatchesRuntimePin(t *testing.T) {
	assert.Equal(t, "postgres:15.4-alpine", testdb.PostgresImage)
}
