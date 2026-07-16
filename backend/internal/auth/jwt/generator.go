package jwt

import (
	"context"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// defaultAccessTTL is the 24h access-token lifetime from docs/security.md.
const defaultAccessTTL = 24 * time.Hour

// Generator issues signed RS256 access tokens using the KeyProvider's primary
// key, stamping that key's id into the `kid` header for validation during
// rotation.
type Generator struct {
	keys      KeyProvider
	accessTTL time.Duration
	now       func() time.Time // injectable for tests
}

// NewGenerator builds a Generator. A non-positive accessTTL falls back to 24h.
func NewGenerator(provider KeyProvider, accessTTL time.Duration) *Generator {
	if accessTTL <= 0 {
		accessTTL = defaultAccessTTL
	}
	return &Generator{keys: provider, accessTTL: accessTTL, now: time.Now}
}

// GenerateAccessToken issues a signed access token for userID with the
// hasSubscription claim and the standard registered claims (sub, iss, iat, nbf,
// exp, jti).
func (g *Generator) GenerateAccessToken(ctx context.Context, userID uuid.UUID, hasSubscription bool) (string, error) {
	if userID == uuid.Nil {
		return "", fmt.Errorf("jwt: user id must not be empty")
	}

	signingKey, err := g.keys.SigningKey(ctx)
	if err != nil {
		return "", fmt.Errorf("jwt: load signing key: %w", err)
	}

	now := g.now()
	claims := Claims{
		HasSubscription: hasSubscription,
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    issuer,
			IssuedAt:  gojwt.NewNumericDate(now),
			NotBefore: gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(g.accessTTL)),
			ID:        uuid.NewString(),
		},
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims)
	token.Header["kid"] = signingKey.KeyID

	signed, err := token.SignedString(signingKey.Private)
	if err != nil {
		return "", fmt.Errorf("jwt: sign access token: %w", err)
	}
	return signed, nil
}
