package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword_CorrectPassword_ComparePasswordSucceeds(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	err = ComparePassword(hash, "correct-horse-battery-staple")
	assert.NoError(t, err)
}

func TestHashPassword_WrongPassword_ComparePasswordFails(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	require.NoError(t, err)

	err = ComparePassword(hash, "wrong-password")
	assert.Error(t, err)
}

func TestHashPassword_ProducesHashWithBcryptCostTwelve(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	require.NoError(t, err)

	cost, err := bcrypt.Cost([]byte(hash))
	require.NoError(t, err)
	assert.Equal(t, 12, cost)
}

func TestHashPassword_EmptyPassword_StillProducesUsableHash(t *testing.T) {
	hash, err := HashPassword("")
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	assert.NoError(t, ComparePassword(hash, ""))
	assert.Error(t, ComparePassword(hash, "not-empty"))
}
