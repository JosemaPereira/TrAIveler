package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Issuer mints a brand-new session (an access token plus a first refresh token)
// for a user who just proved their identity. It is the initial-issuance
// counterpart to Refresher, which rotates an existing token; both persist only
// the refresh token's SHA-256 hash, never the plaintext.
type Issuer struct {
	store      RefreshTokenStore
	generator  *Generator
	refreshTTL time.Duration
	now        func() time.Time // injectable for tests
}

// NewIssuer wires an Issuer. A non-positive refreshTTL falls back to 30 days
// (defaultRefreshTTL, docs/security.md).
func NewIssuer(store RefreshTokenStore, generator *Generator, refreshTTL time.Duration) *Issuer {
	if refreshTTL <= 0 {
		refreshTTL = defaultRefreshTTL
	}
	return &Issuer{store: store, generator: generator, refreshTTL: refreshTTL, now: time.Now}
}

// Issue mints an access token (carrying hasSubscription) and a fresh refresh
// token for userID, persisting only the refresh token's hash. The access token is
// minted first so a signing failure short-circuits before any refresh token is
// stored.
func (i *Issuer) Issue(ctx context.Context, userID uuid.UUID, hasSubscription bool) (TokenPair, error) {
	accessToken, err := i.generator.GenerateAccessToken(ctx, userID, hasSubscription)
	if err != nil {
		return TokenPair{}, fmt.Errorf("jwt: issue access token: %w", err)
	}

	rawRefresh, refreshHash, err := generateRefreshToken()
	if err != nil {
		return TokenPair{}, fmt.Errorf("jwt: generate refresh token: %w", err)
	}

	now := i.now()
	refreshExpiresAt := now.Add(i.refreshTTL)
	if err := i.store.Create(ctx, NewRefreshToken{
		UserID:    userID,
		TokenHash: refreshHash,
		ExpiresAt: refreshExpiresAt,
	}); err != nil {
		return TokenPair{}, fmt.Errorf("jwt: persist refresh token: %w", err)
	}

	return TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     rawRefresh,
		AccessExpiresAt:  now.Add(i.generator.accessTTL),
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}
