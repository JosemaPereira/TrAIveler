package jwt

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

const (
	// defaultRefreshTTL is the 30-day refresh-token lifetime (docs/security.md).
	defaultRefreshTTL = 30 * 24 * time.Hour
	// refreshTokenBytes is the token entropy before encoding (data-model.md: 32 min).
	refreshTokenBytes = 32
)

// RefreshFailureMessage is the single client-facing message for every failed
// refresh, per specs/008-auth-collaboration-ux/contracts/api.md.
//
// It is exported because the HTTP handler rejects a missing refresh cookie
// itself, without ever reaching the Refresher, and the two 401 bodies must be
// byte-identical: absent, unknown, expired, and revoked tokens have to be
// indistinguishable from the outside, or the uniform error stops preventing
// the enumeration it exists to prevent. Two separate literals would silently
// drift apart, so both paths read this constant. (Same reasoning as
// HashRefreshToken being exported for the logout path.)
const RefreshFailureMessage = "Session expired. Please log in again."

// ErrRefreshTokenNotFound is returned by RefreshTokenStore.ByHash when no row
// matches; the Refresher treats it as an authentication failure.
var ErrRefreshTokenNotFound = errors.New("jwt: refresh token not found")

// RefreshTokenRecord is a stored refresh token's non-secret fields, mirroring
// the refresh_tokens table. Only the hash is ever persisted, never the token.
type RefreshTokenRecord struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	RevokedAt *time.Time // nil while live
}

// NewRefreshToken is the data to persist a freshly issued refresh token.
type NewRefreshToken struct {
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
}

// RefreshTokenStore is the persistence seam for refresh tokens, implemented in
// production by auth.RefreshStore over the refresh_tokens table.
// Implementations must be concurrency-safe.
type RefreshTokenStore interface {
	// ByHash looks up a token by SHA-256 hash, returning ErrRefreshTokenNotFound.
	ByHash(ctx context.Context, tokenHash string) (RefreshTokenRecord, error)
	// Revoke marks the token revoked; revoking twice is a no-op.
	Revoke(ctx context.Context, id uuid.UUID) error
	// Create persists a newly issued refresh token.
	Create(ctx context.Context, token NewRefreshToken) error
}

// SubscriptionResolver reports a user's current subscription status so a
// refreshed token carries an up-to-date has_subscription claim.
type SubscriptionResolver interface {
	HasActiveSubscription(ctx context.Context, userID uuid.UUID) (bool, error)
}

// TokenPair is a freshly minted session (issued or refreshed). RefreshToken is
// the plaintext handed to the client; only its hash is stored.
//
// UserID identifies the session's owner. The pair carries it because
// POST /auth/refresh is a public route: its handler has no authenticated
// request context to attribute the auth_token_refresh security event from, and
// re-parsing the access token just to recover the subject would be wasteful.
type TokenPair struct {
	UserID           uuid.UUID
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

// Refresher exchanges a valid refresh token for a fresh access/refresh pair,
// rotating the refresh token on every use: the presented token is revoked and a
// new one issued. This makes tokens one-time-use, so a replayed stolen token is
// already spent.
type Refresher struct {
	store      RefreshTokenStore
	generator  *Generator
	subs       SubscriptionResolver
	refreshTTL time.Duration
	now        func() time.Time
}

// NewRefresher wires a Refresher. A non-positive refreshTTL falls back to 30 days.
func NewRefresher(store RefreshTokenStore, generator *Generator, subs SubscriptionResolver, refreshTTL time.Duration) *Refresher {
	if refreshTTL <= 0 {
		refreshTTL = defaultRefreshTTL
	}
	return &Refresher{store: store, generator: generator, subs: subs, refreshTTL: refreshTTL, now: time.Now}
}

// RefreshToken validates rawRefreshToken, revokes it, and issues a new
// access/refresh pair for its owner.
//
// A missing, expired, or revoked token yields a generic authentication error;
// storage or signing failures surface wrapped for the caller to log. The old
// token is revoked before minting the new pair, so it can never be reused even
// if a later step fails.
func (r *Refresher) RefreshToken(ctx context.Context, rawRefreshToken string) (TokenPair, error) {
	if rawRefreshToken == "" {
		return TokenPair{}, unauthorizedRefresh(errors.New("empty refresh token"))
	}

	hash := HashRefreshToken(rawRefreshToken)
	record, err := r.store.ByHash(ctx, hash)
	if errors.Is(err, ErrRefreshTokenNotFound) {
		return TokenPair{}, unauthorizedRefresh(err)
	}
	if err != nil {
		return TokenPair{}, fmt.Errorf("jwt: look up refresh token: %w", err)
	}

	now := r.now()
	if record.RevokedAt != nil {
		return TokenPair{}, unauthorizedRefresh(errors.New("refresh token already revoked"))
	}
	if !record.ExpiresAt.After(now) {
		return TokenPair{}, unauthorizedRefresh(errors.New("refresh token expired"))
	}

	if err := r.store.Revoke(ctx, record.ID); err != nil {
		return TokenPair{}, fmt.Errorf("jwt: revoke refresh token: %w", err)
	}

	hasSubscription, err := r.subs.HasActiveSubscription(ctx, record.UserID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("jwt: resolve subscription: %w", err)
	}

	accessToken, err := r.generator.GenerateAccessToken(ctx, record.UserID, hasSubscription)
	if err != nil {
		return TokenPair{}, fmt.Errorf("jwt: issue access token: %w", err)
	}

	rawNext, hashNext, err := generateRefreshToken()
	if err != nil {
		return TokenPair{}, fmt.Errorf("jwt: generate refresh token: %w", err)
	}

	refreshExpiresAt := now.Add(r.refreshTTL)
	if err := r.store.Create(ctx, NewRefreshToken{
		UserID:    record.UserID,
		TokenHash: hashNext,
		ExpiresAt: refreshExpiresAt,
	}); err != nil {
		return TokenPair{}, fmt.Errorf("jwt: persist refresh token: %w", err)
	}

	return TokenPair{
		UserID:           record.UserID,
		AccessToken:      accessToken,
		RefreshToken:     rawNext,
		AccessExpiresAt:  now.Add(r.generator.accessTTL),
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// generateRefreshToken mints a random opaque token, returning the plaintext (for
// the client) and its SHA-256 hash (for storage).
func generateRefreshToken() (raw, hash string, err error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("jwt: read random bytes: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	return raw, HashRefreshToken(raw), nil
}

// HashRefreshToken returns the hex SHA-256 hash used as the stored lookup key.
// SHA-256 (not bcrypt) suffices since the token is already high-entropy random.
// Exported so a caller holding a presented token (e.g. logout) can compute the
// same lookup key without re-deriving the scheme.
func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// unauthorizedRefresh builds the uniform auth error for any invalid refresh
// token, wrapping the cause for logs only. The client-facing message is
// RefreshFailureMessage — see its doc for why the handler shares this exact
// constant.
func unauthorizedRefresh(cause error) error {
	domainErr := domainerrors.Unauthorized(RefreshFailureMessage)
	domainErr.Err = fmt.Errorf("jwt: refresh: %w", cause)
	return domainErr
}
