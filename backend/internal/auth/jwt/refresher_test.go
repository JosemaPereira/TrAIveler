package jwt

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// fakeConnectivityError stands in for a pgx connection-level failure (dial
// refused, timeout, DNS failure, ...), reused here (rather than importing
// the errors package's test-only type) so this package's tests stay free of
// a test-to-test dependency. pgx's own such errors are unexported and only
// constructible by actually dialing, so this minimal type satisfies the same
// duck-typed `interface{ SafeToRetry() bool }` that pgconn.SafeToRetry
// checks for via errors.As — see errors.ServiceUnavailableFromDB.
type fakeConnectivityError struct {
	cause error
}

func (e *fakeConnectivityError) Error() string     { return fmt.Sprintf("dial: %s", e.cause) }
func (e *fakeConnectivityError) Unwrap() error     { return e.cause }
func (e *fakeConnectivityError) SafeToRetry() bool { return true }

// newConnRefusedError builds a connection-level failure, e.g. a database
// outage during POST /auth/refresh (issue #192 Task A).
func newConnRefusedError() error {
	return &fakeConnectivityError{cause: errors.New("connection refused")}
}

// fakeStore is an in-memory RefreshTokenStore for tests, keyed by token hash.
// Each storage-touching method has its own fail hook so a test can target a
// single call site (e.g. Revoke succeeding but Create failing) instead of
// every call failing identically.
type fakeStore struct {
	byHash     map[string]RefreshTokenRecord
	created    []NewRefreshToken
	revoked    []uuid.UUID
	failByHash error // if set, ByHash returns this error instead of looking up
	failRevoke error // if set, Revoke returns this error instead of revoking
	failCreate error // if set, Create returns this error instead of persisting
}

func newFakeStore() *fakeStore {
	return &fakeStore{byHash: map[string]RefreshTokenRecord{}}
}

func (s *fakeStore) ByHash(_ context.Context, hash string) (RefreshTokenRecord, error) {
	if s.failByHash != nil {
		return RefreshTokenRecord{}, s.failByHash
	}
	rec, ok := s.byHash[hash]
	if !ok {
		return RefreshTokenRecord{}, ErrRefreshTokenNotFound
	}
	return rec, nil
}

func (s *fakeStore) Revoke(_ context.Context, id uuid.UUID) error {
	if s.failRevoke != nil {
		return s.failRevoke
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
	if s.failCreate != nil {
		return s.failCreate
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
	// A plain, non-connectivity query-level error (not something
	// pgconn.SafeToRetry would recognize) — contrast with
	// TestRefresher_RefreshToken_DBConnectivityFailure_MapsToServiceUnavailable
	// below, which uses a connection-level failure instead.
	store.failByHash = errors.New("db is down")

	refresher := newTestRefresher(t, store, stubSubs{})
	_, err := refresher.RefreshToken(context.Background(), "anything")

	require.Error(t, err)
	// A transient, non-connectivity storage failure must surface as an
	// unmapped infrastructure error (500 via errors.HandleError), not be
	// masked as an authentication rejection nor misclassified as
	// service_unavailable.
	var domainErr *domainerrors.DomainError
	assert.False(t, errors.As(err, &domainErr), "storage failure should not become a domain error")
}

// TestRefresher_RefreshToken_DBConnectivityFailure_MapsToServiceUnavailable
// is issue #192 Task A: specs/008-auth-collaboration-ux/contracts/api.md
// requires 503 for "Database unavailable" on refresh, so a connection-level
// failure at any of the three storage-touching points, or at the also
// DB-backed subscription-resolver step, must map to service_unavailable
// (503 + Retry-After) rather than the generic 500 a plain wrapped error
// would produce.
func TestRefresher_RefreshToken_DBConnectivityFailure_MapsToServiceUnavailable(t *testing.T) {
	assertServiceUnavailable := func(t *testing.T, err error) {
		t.Helper()
		require.Error(t, err)
		var domainErr *domainerrors.DomainError
		require.True(t, errors.As(err, &domainErr), "expected a *errors.DomainError")
		assert.Equal(t, "service_unavailable", domainErr.Code)
		assert.Equal(t, 30, domainErr.Details["retry_after_seconds"])
	}

	t.Run("when looking up the presented token fails", func(t *testing.T) {
		store := newFakeStore()
		store.failByHash = newConnRefusedError()

		refresher := newTestRefresher(t, store, stubSubs{})
		_, err := refresher.RefreshToken(context.Background(), "anything")

		assertServiceUnavailable(t, err)
	})

	t.Run("when revoking the old token fails", func(t *testing.T) {
		store := newFakeStore()
		raw := seedToken(t, store, uuid.New(), time.Now().Add(24*time.Hour))
		store.failRevoke = newConnRefusedError()

		refresher := newTestRefresher(t, store, stubSubs{})
		_, err := refresher.RefreshToken(context.Background(), raw)

		assertServiceUnavailable(t, err)
	})

	t.Run("when persisting the rotated token fails", func(t *testing.T) {
		store := newFakeStore()
		raw := seedToken(t, store, uuid.New(), time.Now().Add(24*time.Hour))
		store.failCreate = newConnRefusedError()

		refresher := newTestRefresher(t, store, stubSubs{has: true})
		_, err := refresher.RefreshToken(context.Background(), raw)

		assertServiceUnavailable(t, err)
	})

	t.Run("when resolving the subscription fails", func(t *testing.T) {
		store := newFakeStore()
		raw := seedToken(t, store, uuid.New(), time.Now().Add(24*time.Hour))

		refresher := newTestRefresher(t, store, stubSubs{err: newConnRefusedError()})
		_, err := refresher.RefreshToken(context.Background(), raw)

		assertServiceUnavailable(t, err)
	})
}

func TestRefresher_RefreshToken_SubscriptionErrorPropagates(t *testing.T) {
	store := newFakeStore()
	raw := seedToken(t, store, uuid.New(), time.Now().Add(24*time.Hour))

	refresher := newTestRefresher(t, store, stubSubs{err: errors.New("subs unavailable")})
	_, err := refresher.RefreshToken(context.Background(), raw)
	assert.ErrorContains(t, err, "subs unavailable")
}
