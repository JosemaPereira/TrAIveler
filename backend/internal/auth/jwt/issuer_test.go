package jwt

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestIssuer wires an Issuer over a real Generator with a live key, also
// returning the provider so parseIssuedAccessToken can validate a minted token
// against the same key that signed it.
func newTestIssuer(t *testing.T, store RefreshTokenStore, refreshTTL time.Duration) (*Issuer, *StaticKeyProvider) {
	t.Helper()
	key := testKey(t, "key-1")
	provider, err := NewStaticKeyProvider("key-1", key)
	require.NoError(t, err)
	return NewIssuer(store, NewGenerator(provider, time.Hour), refreshTTL), provider
}

func TestIssue_ValidUser_ReturnsSignedPairAndPersistsRefreshHash(t *testing.T) {
	store := newFakeStore()
	issuer, provider := newTestIssuer(t, store, 30*24*time.Hour)
	userID := uuid.New()

	pair, err := issuer.Issue(context.Background(), userID, true)

	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.True(t, pair.AccessExpiresAt.After(time.Now()))
	assert.True(t, pair.RefreshExpiresAt.After(pair.AccessExpiresAt))

	// The refresh token is persisted only as its SHA-256 hash, never the plaintext.
	require.Len(t, store.created, 1)
	assert.Equal(t, userID, store.created[0].UserID)
	assert.Equal(t, HashRefreshToken(pair.RefreshToken), store.created[0].TokenHash)
	assert.NotEqual(t, pair.RefreshToken, store.created[0].TokenHash)

	// The access token embeds the subscription claim for the user.
	claims := parseIssuedAccessToken(t, provider, pair.AccessToken)
	assert.Equal(t, userID.String(), claims.Subject)
	assert.True(t, claims.HasSubscription)
}

func TestIssue_NilUserID_ReturnsErrorAndPersistsNothing(t *testing.T) {
	store := newFakeStore()
	issuer, _ := newTestIssuer(t, store, 30*24*time.Hour)

	_, err := issuer.Issue(context.Background(), uuid.Nil, false)

	require.Error(t, err)
	assert.Empty(t, store.created, "no refresh token must be persisted when access minting fails")
}

func TestIssue_StorePersistFails_ReturnsWrappedError(t *testing.T) {
	store := newFakeStore()
	store.failNext = errors.New("db down")
	issuer, _ := newTestIssuer(t, store, 30*24*time.Hour)

	_, err := issuer.Issue(context.Background(), uuid.New(), false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
}

func TestNewIssuer_NonPositiveTTL_FallsBackToDefault(t *testing.T) {
	store := newFakeStore()
	issuer, _ := newTestIssuer(t, store, 0)
	require.Equal(t, defaultRefreshTTL, issuer.refreshTTL)

	pair, err := issuer.Issue(context.Background(), uuid.New(), false)
	require.NoError(t, err)
	// ~30 days out (defaultRefreshTTL), allowing a small execution window.
	assert.WithinDuration(t, time.Now().Add(defaultRefreshTTL), pair.RefreshExpiresAt, time.Minute)
}

// parseIssuedAccessToken decodes an access token minted in these tests without
// re-validating the signature (a Validator has its own tests); it just exposes
// the claims for assertions.
func parseIssuedAccessToken(t *testing.T, provider *StaticKeyProvider, token string) Claims {
	t.Helper()
	claims, err := NewValidator(provider).ValidateToken(context.Background(), token)
	require.NoError(t, err)
	return *claims
}
