package migrations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDir_ResolvesToRealMigrationsDirectory(t *testing.T) {
	require.True(t, filepath.IsAbs(Dir), "Dir should be an absolute path, not a caller-relative string")

	info, err := os.Stat(Dir)
	require.NoError(t, err, "expected Dir to point at an existing directory")
	assert.True(t, info.IsDir(), "expected Dir to be a directory")

	// Confirms Dir resolved to the real migrations directory, not just any existing one.
	_, err = os.Stat(filepath.Join(Dir, "001_create_users_table.sql"))
	assert.NoError(t, err, "expected 001_create_users_table.sql to exist under Dir")
}

func TestSetDialect_ConfiguresPostgresDialect_ReturnsNoError(t *testing.T) {
	err := SetDialect()
	assert.NoError(t, err, "SetDialect should configure the postgres dialect without error")
}
