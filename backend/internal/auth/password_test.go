package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword_CorrectPassword_ComparePasswordSucceeds(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple", DefaultBcryptCost)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	err = ComparePassword(hash, "correct-horse-battery-staple")
	assert.NoError(t, err)
}

func TestHashPassword_WrongPassword_ComparePasswordFails(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple", DefaultBcryptCost)
	require.NoError(t, err)

	err = ComparePassword(hash, "wrong-password")
	assert.Error(t, err)
}

func TestHashPassword_UsesGivenCost(t *testing.T) {
	// A non-default (but valid) cost must be reflected in the produced hash, proving
	// the configured BCRYPT_COST reaches bcrypt rather than a hardcoded constant.
	const customCost = 6
	hash, err := HashPassword("correct-horse-battery-staple", customCost)
	require.NoError(t, err)

	cost, err := bcrypt.Cost([]byte(hash))
	require.NoError(t, err)
	assert.Equal(t, customCost, cost)
}

func TestHashPassword_InvalidCost_FallsBackToDefault(t *testing.T) {
	for _, invalid := range []int{0, bcrypt.MinCost - 1, bcrypt.MaxCost + 1} {
		hash, err := HashPassword("correct-horse-battery-staple", invalid)
		require.NoError(t, err, "invalid cost %d must not error", invalid)

		cost, err := bcrypt.Cost([]byte(hash))
		require.NoError(t, err)
		assert.Equal(t, DefaultBcryptCost, cost, "invalid cost %d must fall back to the default", invalid)
	}
}

func TestHashPassword_EmptyPassword_StillProducesUsableHash(t *testing.T) {
	hash, err := HashPassword("", DefaultBcryptCost)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	assert.NoError(t, ComparePassword(hash, ""))
	assert.Error(t, ComparePassword(hash, "not-empty"))
}
