package jwt

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// fakeStore is an in-memory RefreshTokenStore for tests, keyed by token hash.
type fakeStore struct {
	byHash   map[string]RefreshTokenRecord
	created  []NewRefreshToken
	revoked  []uuid.UUID
	failNext error // if set, the next store call returns this error
}

func newFakeStore() *fakeStore {
	return &fakeStore{byHash: map[string]RefreshTokenRecord{}}
}

func (s *fakeStore) ByHash(_ context.Context, hash string) (RefreshTokenRecord, error) {
	if s.failNext != nil {
		return RefreshTokenRecord{}, s.failNext
	}
	rec, ok := s.byHash[hash]
	if !ok {
		return RefreshTokenRecord{}, ErrRefreshTokenNotFound
	}
	return rec, nil
}

func (s *fakeStore) Revoke(_ context.Context, id uuid.UUID) error {
	if s.failNext != nil {
		return s.failNext
	}
	s.revoked = append(s.revoked, id)
	// Reflect the revocation in the stored record so a reuse attempt sees it.
	for hash, rec := range s.byHash {
		if rec.ID == id {
			now := time.Now()
			rec.RevokedAt = &now
			s.byHash[hash] = rec
		}
	}
	return nil
}

func (s *fakeStore) Create(_ context.Context, token NewRefreshToken) error {
	if s.failNext != nil {
		return s.failNext
	}
	s.created = append(s.created, token)
	s.byHash[token.TokenHash] = RefreshTokenRecord{
		ID:        uuid.New(),
		UserID:    token.UserID,
		ExpiresAt: token.ExpiresAt,
	}
	return nil
}

// stubSubs is a fixed SubscriptionResolver.
type stubSubs struct {
	has bool
	err error
}

func (s stubSubs) HasActiveSubscription(context.Context, uuid.UUID) (bool, error) {
	return s.has, s.err
}

// seedToken inserts a live refresh token into the store and returns its
// plaintext value for the test to present.
func seedToken(t *testing.T, store *fakeStore, userID uuid.UUID, expiresAt time.Time) string {
	t.Helper()
	raw, hash, err := generateRefreshToken()
	require.NoError(t, err)
	store.byHash[hash] = RefreshTokenRecord{ID: uuid.New(), UserID: userID, ExpiresAt: expiresAt}
	return raw
}

// newTestRefresher wires a Refresher over a real Generator with a live key.
func newTestRefresher(t *testing.T, store RefreshTokenStore, subs SubscriptionResolver) *Refresher {
	t.Helper()
	key := testKey(t, "key-1")
	provider, err := NewStaticKeyProvider("key-1", key)
	require.NoError(t, err)
	return NewRefresher(store, NewGenerator(provider, time.Hour), subs, 30*24*time.Hour)
}

func TestRefresher_RefreshToken_Success(t *testing.T) {
	store := newFakeStore()
	userID := uuid.New()
	oldRaw := seedToken(t, store, userID, time.Now().Add(24*time.Hour))
	oldHash := HashRefreshToken(oldRaw)

	refresher := newTestRefresher(t, store, stubSubs{has: true})
	pair, err := refresher.RefreshToken(context.Background(), oldRaw)
	require.NoError(t, err)

	// A new access token and a new, different refresh token were issued.
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.NotEqual(t, oldRaw, pair.RefreshToken)
	assert.True(t, pair.AccessExpiresAt.After(time.Now()))
	assert.True(t, pair.RefreshExpiresAt.After(time.Now()))

	// The old token was revoked and exactly one new token persisted.
	assert.Len(t, store.revoked, 1)
	require.Len(t, store.created, 1)
	assert.Equal(t, userID, store.created[0].UserID)
	assert.NotEqual(t, oldHash, store.created[0].TokenHash, "new token hash must differ from the old one")

	// The freshly minted access token carries the resolved subscription claim.
	claims, err := NewValidator(mustProvider(t, refresher)).ValidateToken(context.Background(), pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims.Subject)
	assert.True(t, claims.HasSubscription)
}

// mustProvider returns the KeyProvider behind a Refresher's generator.
func mustProvider(t *testing.T, r *Refresher) KeyProvider {
	t.Helper()
	return r.generator.keys
}

func TestRefresher_RefreshToken_RevokedTokenCannotBeReused(t *testing.T) {
	store := newFakeStore()
	oldRaw := seedToken(t, store, uuid.New(), time.Now().Add(24*time.Hour))

	refresher := newTestRefresher(t, store, stubSubs{})
	_, err := refresher.RefreshToken(context.Background(), oldRaw)
	require.NoError(t, err)

	// Presenting the same token again is rejected: it is now revoked.
	_, err = refresher.RefreshToken(context.Background(), oldRaw)
	assertUnauthorized(t, err)
}

func TestRefresher_RefreshToken_Rejections(t *testing.T) {
	t.Run("when the token is unknown", func(t *testing.T) {
		refresher := newTestRefresher(t, newFakeStore(), stubSubs{})
		_, err := refresher.RefreshToken(context.Background(), "does-not-exist")
		assertUnauthorized(t, err)
	})

	t.Run("when the token is empty", func(t *testing.T) {
		refresher := newTestRefresher(t, newFakeStore(), stubSubs{})
		_, err := refresher.RefreshToken(context.Background(), "")
		assertUnauthorized(t, err)
	})

	t.Run("when the token is expired", func(t *testing.T) {
		store := newFakeStore()
		expiredRaw := seedToken(t, store, uuid.New(), time.Now().Add(-time.Minute))
		refresher := newTestRefresher(t, store, stubSubs{})
		_, err := refresher.RefreshToken(context.Background(), expiredRaw)
		assertUnauthorized(t, err)
		assert.Empty(t, store.revoked, "an expired token must not be revoked or rotated")
	})

	t.Run("when the token is already revoked", func(t *testing.T) {
		store := newFakeStore()
		raw, hash, err := generateRefreshToken()
		require.NoError(t, err)
		revokedAt := time.Now()
		store.byHash[hash] = RefreshTokenRecord{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			RevokedAt: &revokedAt,
		}
		refresher := newTestRefresher(t, store, stubSubs{})
		_, err = refresher.RefreshToken(context.Background(), raw)
		assertUnauthorized(t, err)
	})
}

func TestRefresher_RefreshToken_StoreErrorIsNotUnauthorized(t *testing.T) {
	store := newFakeStore()
	store.failNext = errors.New("db is down")

	refresher := newTestRefresher(t, store, stubSubs{})
	_, err := refresher.RefreshToken(context.Background(), "anything")

	require.Error(t, err)
	// A transient storage failure must surface as an infrastructure error, not
	// be masked as an authentication rejection.
	var domainErr *domainerrors.DomainError
	assert.False(t, errors.As(err, &domainErr), "storage failure should not become a domain auth error")
}

func TestRefresher_RefreshToken_SubscriptionErrorPropagates(t *testing.T) {
	store := newFakeStore()
	raw := seedToken(t, store, uuid.New(), time.Now().Add(24*time.Hour))

	refresher := newTestRefresher(t, store, stubSubs{err: errors.New("subs unavailable")})
	_, err := refresher.RefreshToken(context.Background(), raw)
	assert.ErrorContains(t, err, "subs unavailable")
}
