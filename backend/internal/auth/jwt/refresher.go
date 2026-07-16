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
	// defaultRefreshTTL is the refresh-token lifetime. docs/security.md and
	// specs/004-security-auth-model/data-model.md both cap this at 30 days.
	defaultRefreshTTL = 30 * 24 * time.Hour
	// refreshTokenBytes is the entropy of an opaque refresh token before
	// encoding. data-model.md requires "32 bytes minimum".
	refreshTokenBytes = 32
)

// ErrRefreshTokenNotFound is the sentinel a RefreshTokenStore returns from
// ByHash when no row matches the given hash. The Refresher maps it — like an
// expired or revoked token — to a generic authentication error.
var ErrRefreshTokenNotFound = errors.New("jwt: refresh token not found")

// RefreshTokenRecord is a stored refresh token as the Refresher needs to see
// it. It mirrors the non-secret columns of the refresh_tokens table; the token
// value itself is never stored or returned, only its hash is persisted.
type RefreshTokenRecord struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	// RevokedAt is nil for a live token; non-nil once logout, password change,
	// or a prior refresh has revoked it.
	RevokedAt *time.Time
}

// NewRefreshToken is the data needed to persist a freshly issued refresh token.
// Only the SHA-256 hash is stored, never the plaintext token.
type NewRefreshToken struct {
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
}

// RefreshTokenStore is the persistence seam for refresh tokens. The MVP has no
// implementation here; the PostgreSQL-backed store lands with the login flow
// (Spec 008 Sprint 6-7) against the refresh_tokens table. Implementations must
// be safe for concurrent use.
type RefreshTokenStore interface {
	// ByHash looks a token up by its SHA-256 hash, returning
	// ErrRefreshTokenNotFound if none matches.
	ByHash(ctx context.Context, tokenHash string) (RefreshTokenRecord, error)
	// Revoke marks the token with the given id revoked. Revoking an already-
	// revoked token is a no-op.
	Revoke(ctx context.Context, id uuid.UUID) error
	// Create persists a newly issued refresh token.
	Create(ctx context.Context, token NewRefreshToken) error
}

// SubscriptionResolver reports whether a user currently has an active
// subscription, so a refreshed access token carries an up-to-date
// has_subscription claim rather than copying the stale one from the old token.
// Its implementation lands with the subscription domain.
type SubscriptionResolver interface {
	HasActiveSubscription(ctx context.Context, userID uuid.UUID) (bool, error)
}

// TokenPair is the result of a successful refresh: a new signed access token
// and a new opaque refresh token, each with its expiry. RefreshToken is the
// plaintext value handed to the client; only its hash is persisted.
type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

// Refresher exchanges a valid refresh token for a fresh access/refresh pair,
// rotating the refresh token on every use: the presented token is revoked and a
// brand-new one issued. This one-time-use property means a stolen-and-replayed
// refresh token is detected (its single use is already spent) and limits the
// window a leaked token is useful.
type Refresher struct {
	store      RefreshTokenStore
	generator  *Generator
	subs       SubscriptionResolver
	refreshTTL time.Duration
	now        func() time.Time
}

// NewRefresher wires a Refresher from its collaborators. A non-positive
// refreshTTL falls back to the 30-day default from docs/security.md.
func NewRefresher(store RefreshTokenStore, generator *Generator, subs SubscriptionResolver, refreshTTL time.Duration) *Refresher {
	if refreshTTL <= 0 {
		refreshTTL = defaultRefreshTTL
	}
	return &Refresher{
		store:      store,
		generator:  generator,
		subs:       subs,
		refreshTTL: refreshTTL,
		now:        time.Now,
	}
}

// RefreshToken validates rawRefreshToken, revokes it, and issues a new
// access/refresh token pair for its owner.
//
// A missing, expired, or already-revoked token yields a generic
// authentication_required domain error (no distinction is exposed to the
// client). Storage or signing failures surface as wrapped errors for the
// caller to log and translate. The old token is revoked before the new pair is
// minted: if a later step fails the user simply re-authenticates, and the
// presented token can never be reused.
func (r *Refresher) RefreshToken(ctx context.Context, rawRefreshToken string) (TokenPair, error) {
	if rawRefreshToken == "" {
		return TokenPair{}, unauthorizedRefresh(errors.New("empty refresh token"))
	}

	hash := hashRefreshToken(rawRefreshToken)
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
		AccessToken:      accessToken,
		RefreshToken:     rawNext,
		AccessExpiresAt:  now.Add(r.generator.accessTTL),
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// generateRefreshToken mints a cryptographically random opaque refresh token,
// returning both the plaintext (for the client) and its SHA-256 hash (for
// storage).
func generateRefreshToken() (raw, hash string, err error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("jwt: read random bytes: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	return raw, hashRefreshToken(raw), nil
}

// hashRefreshToken returns the hex-encoded SHA-256 hash used as the stored
// lookup key. SHA-256 (not bcrypt) is appropriate here: the token is already
// high-entropy random, so the only requirement is a fast, deterministic,
// preimage-resistant digest — matching data-model.md's token_hash column.
func hashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// unauthorizedRefresh builds the uniform authentication error for any invalid
// refresh token, wrapping the specific cause for server-side logging.
func unauthorizedRefresh(cause error) error {
	domainErr := domainerrors.Unauthorized("invalid or expired refresh token")
	domainErr.Err = fmt.Errorf("jwt: refresh: %w", cause)
	return domainErr
}
