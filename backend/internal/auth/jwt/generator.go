package jwt

import (
	"context"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// defaultAccessTTL is the access-token lifetime mandated by docs/security.md
// ("Access tokens: 24 hours maximum lifetime").
const defaultAccessTTL = 24 * time.Hour

// Generator issues signed RS256 access tokens. It signs with whichever key the
// KeyProvider currently reports as primary and stamps that key's id into the
// token's `kid` header so the Validator can select the matching public key
// during rotation.
type Generator struct {
	keys      KeyProvider
	accessTTL time.Duration
	// now is injectable so tests can pin token timestamps; production code
	// leaves it as time.Now.
	now func() time.Time
}

// NewGenerator builds a Generator that signs tokens with keys from provider.
// A non-positive accessTTL falls back to the 24h default from docs/security.md.
func NewGenerator(provider KeyProvider, accessTTL time.Duration) *Generator {
	if accessTTL <= 0 {
		accessTTL = defaultAccessTTL
	}
	return &Generator{
		keys:      provider,
		accessTTL: accessTTL,
		now:       time.Now,
	}
}

// GenerateAccessToken issues a signed access token for userID, embedding the
// hasSubscription authorization claim. The token carries the standard
// registered claims (sub, iss, iat, nbf, exp, jti) and expires after the
// generator's configured TTL.
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
