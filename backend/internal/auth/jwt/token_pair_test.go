package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests pin TokenPair.UserID on both paths that mint a pair — initial
// issuance and rotation — so neither can quietly stop populating it. See
// TokenPair's doc in refresher.go for why the pair has to carry it.

func TestIssue_ValidUser_TokenPairCarriesTheOwningUserID(t *testing.T) {
	// Arrange
	store := newFakeStore()
	issuer, _ := newTestIssuer(t, store, 30*24*time.Hour)
	userID := uuid.New()

	// Act
	pair, err := issuer.Issue(context.Background(), userID, false)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, userID, pair.UserID)
}

func TestRefreshToken_ValidToken_TokenPairCarriesTheOwningUserID(t *testing.T) {
	// Arrange
	store := newFakeStore()
	refresher := newTestRefresher(t, store, stubSubs{has: true})
	userID := uuid.New()
	raw := seedToken(t, store, userID, time.Now().Add(24*time.Hour))

	// Act
	pair, err := refresher.RefreshToken(context.Background(), raw)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, userID, pair.UserID,
		"the refreshed pair must identify the token's owner so the caller can log auth_token_refresh")
}
