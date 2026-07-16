package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStaticKeyProvider_RejectsKeyMissingPublicHalf(t *testing.T) {
	key := testKey(t, "key-1")
	key.Public = nil
	_, err := NewStaticKeyProvider("key-1", key)
	assert.ErrorContains(t, err, "public key")
}

func TestStaticKeyProvider_SigningKey_ErrorWhenPrimaryGone(t *testing.T) {
	// A provider whose primaryID points at no key (only constructible directly)
	// must report ErrNoSigningKey rather than panic.
	provider := &StaticKeyProvider{keys: map[string]ManagedKey{}, primaryID: "gone"}
	_, err := provider.SigningKey(context.Background())
	assert.ErrorIs(t, err, ErrNoSigningKey)
}

func TestNewRefresher_DefaultRefreshTTLIs30Days(t *testing.T) {
	store := newFakeStore()
	raw := seedToken(t, store, uuid.New(), time.Now().Add(24*time.Hour))

	key := testKey(t, "key-1")
	provider, err := NewStaticKeyProvider("key-1", key)
	require.NoError(t, err)
	// A non-positive refreshTTL falls back to the 30-day default.
	refresher := NewRefresher(store, NewGenerator(provider, time.Hour), stubSubs{}, 0)

	before := time.Now()
	pair, err := refresher.RefreshToken(context.Background(), raw)
	require.NoError(t, err)

	wantMin := before.Add(30 * 24 * time.Hour)
	assert.False(t, pair.RefreshExpiresAt.Before(wantMin), "refresh token should live ~30 days")
}

func TestRefresher_RefreshToken_CreateFailureIsInfraError(t *testing.T) {
	store := &createFailStore{fakeStore: newFakeStore()}
	raw := seedToken(t, store.fakeStore, uuid.New(), time.Now().Add(24*time.Hour))

	refresher := newTestRefresher(t, store, stubSubs{})
	_, err := refresher.RefreshToken(context.Background(), raw)
	assert.ErrorContains(t, err, "persist refresh token")
}

// createFailStore behaves like fakeStore but fails only on Create, so the
// revoke-then-create ordering can be exercised past the revoke step.
type createFailStore struct {
	*fakeStore
}

func (s *createFailStore) Create(context.Context, NewRefreshToken) error {
	return assertErr
}

var assertErr = &stubError{"create failed"}

type stubError struct{ msg string }

func (e *stubError) Error() string { return e.msg }
