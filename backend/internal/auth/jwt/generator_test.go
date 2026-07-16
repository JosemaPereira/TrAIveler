package jwt

import (
	"context"
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_GenerateAccessToken(t *testing.T) {
	key := testKey(t, "key-primary")
	provider, err := NewStaticKeyProvider("key-primary", key)
	require.NoError(t, err)

	generator := NewGenerator(provider, time.Hour)
	issuedAt := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	generator.now = func() time.Time { return issuedAt }

	userID := uuid.New()
	tokenString, err := generator.GenerateAccessToken(context.Background(), userID, true)
	require.NoError(t, err)

	// Parse the raw token with the signing public key to inspect its contents.
	// The parser's clock is pinned to issuedAt so expiry validation does not
	// depend on the wall clock at test time.
	claims := &Claims{}
	parsed, err := gojwt.NewParser(gojwt.WithTimeFunc(func() time.Time { return issuedAt })).
		ParseWithClaims(tokenString, claims, func(*gojwt.Token) (any, error) {
			return key.Public, nil
		})
	require.NoError(t, err)

	assert.Equal(t, gojwt.SigningMethodRS256.Alg(), parsed.Method.Alg())
	assert.Equal(t, "key-primary", parsed.Header["kid"])
	assert.Equal(t, userID.String(), claims.Subject)
	assert.Equal(t, issuer, claims.Issuer)
	assert.True(t, claims.HasSubscription)
	assert.NotEmpty(t, claims.ID, "jti should be set")
	assert.True(t, issuedAt.Equal(claims.IssuedAt.Time), "iat should equal the pinned issue time")
	assert.True(t, issuedAt.Add(time.Hour).Equal(claims.ExpiresAt.Time), "exp should be iat + TTL")
}

func TestGenerator_GenerateAccessToken_DefaultTTLIs24h(t *testing.T) {
	key := testKey(t, "key-primary")
	provider, err := NewStaticKeyProvider("key-primary", key)
	require.NoError(t, err)

	// A non-positive TTL falls back to the docs/security.md 24h default.
	generator := NewGenerator(provider, 0)
	issuedAt := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	generator.now = func() time.Time { return issuedAt }

	tokenString, err := generator.GenerateAccessToken(context.Background(), uuid.New(), false)
	require.NoError(t, err)

	claims := &Claims{}
	_, err = gojwt.NewParser(gojwt.WithTimeFunc(func() time.Time { return issuedAt })).
		ParseWithClaims(tokenString, claims, func(*gojwt.Token) (any, error) {
			return key.Public, nil
		})
	require.NoError(t, err)
	assert.True(t, issuedAt.Add(24*time.Hour).Equal(claims.ExpiresAt.Time), "default TTL should be 24h")
	assert.False(t, claims.HasSubscription)
}

func TestGenerator_GenerateAccessToken_RejectsNilUserID(t *testing.T) {
	key := testKey(t, "key-primary")
	provider, err := NewStaticKeyProvider("key-primary", key)
	require.NoError(t, err)

	_, err = NewGenerator(provider, time.Hour).GenerateAccessToken(context.Background(), uuid.Nil, false)
	assert.Error(t, err)
}

// failingKeyProvider always fails to produce a signing key, exercising the
// generator's error path when key material is unavailable.
type failingKeyProvider struct{ KeyProvider }

func (failingKeyProvider) SigningKey(context.Context) (SigningKey, error) {
	return SigningKey{}, ErrNoSigningKey
}

func TestGenerator_GenerateAccessToken_PropagatesSigningKeyError(t *testing.T) {
	generator := NewGenerator(failingKeyProvider{}, time.Hour)
	_, err := generator.GenerateAccessToken(context.Background(), uuid.New(), false)
	assert.ErrorIs(t, err, ErrNoSigningKey)
}
