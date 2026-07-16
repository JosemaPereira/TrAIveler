package jwt

import (
	"context"
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// signWith builds a token with an explicit method, kid header, and signing key,
// bypassing the Generator so tests can craft adversarial tokens.
func signWith(t *testing.T, method gojwt.SigningMethod, kid string, signingKey any, claims Claims) string {
	t.Helper()
	token := gojwt.NewWithClaims(method, claims)
	if kid != "" {
		token.Header["kid"] = kid
	}
	signed, err := token.SignedString(signingKey)
	require.NoError(t, err)
	return signed
}

// validClaims returns registered claims that pass every non-signature check,
// expiring at exp.
func validClaims(exp time.Time) Claims {
	now := time.Now()
	return Claims{
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   uuid.NewString(),
			Issuer:    issuer,
			IssuedAt:  gojwt.NewNumericDate(now),
			NotBefore: gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(exp),
		},
	}
}

func TestValidator_ValidateToken_RoundTrip(t *testing.T) {
	key := testKey(t, "key-1")
	provider, err := NewStaticKeyProvider("key-1", key)
	require.NoError(t, err)

	generator := NewGenerator(provider, time.Hour)
	userID := uuid.New()
	tokenString, err := generator.GenerateAccessToken(context.Background(), userID, true)
	require.NoError(t, err)

	claims, err := NewValidator(provider).ValidateToken(context.Background(), tokenString)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims.Subject)
	assert.True(t, claims.HasSubscription)
}

// TestValidator_ValidateToken_MultiKeyRotation is the core rotation scenario:
// while an old key is still active, tokens it signed keep validating even
// though a new primary key now signs fresh tokens.
func TestValidator_ValidateToken_MultiKeyRotation(t *testing.T) {
	oldKey := testKey(t, "key-old")
	newKey := testKey(t, "key-new")

	// After rotation the new key is primary; the old key is still active.
	provider, err := NewStaticKeyProvider("key-new", oldKey, newKey)
	require.NoError(t, err)
	validator := NewValidator(provider)

	tokenFromOldKey := signWith(t, gojwt.SigningMethodRS256, "key-old", oldKey.Private, validClaims(time.Now().Add(time.Hour)))
	tokenFromNewKey := signWith(t, gojwt.SigningMethodRS256, "key-new", newKey.Private, validClaims(time.Now().Add(time.Hour)))

	_, err = validator.ValidateToken(context.Background(), tokenFromOldKey)
	assert.NoError(t, err, "token from the still-active old key must validate")

	_, err = validator.ValidateToken(context.Background(), tokenFromNewKey)
	assert.NoError(t, err, "token from the new primary key must validate")
}

// TestValidator_ValidateToken_RejectsRetiredKey confirms retirement ends a
// key's validity: once the old key is retired, tokens it signed are rejected.
func TestValidator_ValidateToken_RejectsRetiredKey(t *testing.T) {
	oldKey := testKey(t, "key-old")
	newKey := testKey(t, "key-new")

	tokenFromOldKey := signWith(t, gojwt.SigningMethodRS256, "key-old", oldKey.Private, validClaims(time.Now().Add(time.Hour)))

	retiredOld := oldKey
	retiredOld.Retired = true
	provider, err := NewStaticKeyProvider("key-new", retiredOld, newKey)
	require.NoError(t, err)

	_, err = NewValidator(provider).ValidateToken(context.Background(), tokenFromOldKey)
	assertUnauthorized(t, err)
}

func TestValidator_ValidateToken_Rejections(t *testing.T) {
	key := testKey(t, "key-1")
	otherKey := testKey(t, "key-other")
	provider, err := NewStaticKeyProvider("key-1", key)
	require.NoError(t, err)
	validator := NewValidator(provider)

	t.Run("when the token is expired", func(t *testing.T) {
		expired := signWith(t, gojwt.SigningMethodRS256, "key-1", key.Private, validClaims(time.Now().Add(-time.Minute)))
		_, err := validator.ValidateToken(context.Background(), expired)
		assertUnauthorized(t, err)
	})

	t.Run("when the signature does not match the named key", func(t *testing.T) {
		// Signed by otherKey but claims to be key-1: signature check fails.
		forged := signWith(t, gojwt.SigningMethodRS256, "key-1", otherKey.Private, validClaims(time.Now().Add(time.Hour)))
		_, err := validator.ValidateToken(context.Background(), forged)
		assertUnauthorized(t, err)
	})

	t.Run("when the kid is unknown", func(t *testing.T) {
		unknown := signWith(t, gojwt.SigningMethodRS256, "ghost", key.Private, validClaims(time.Now().Add(time.Hour)))
		_, err := validator.ValidateToken(context.Background(), unknown)
		assertUnauthorized(t, err)
	})

	t.Run("when the kid header is missing", func(t *testing.T) {
		noKid := signWith(t, gojwt.SigningMethodRS256, "", key.Private, validClaims(time.Now().Add(time.Hour)))
		_, err := validator.ValidateToken(context.Background(), noKid)
		assertUnauthorized(t, err)
	})

	t.Run("when the issuer is wrong", func(t *testing.T) {
		claims := validClaims(time.Now().Add(time.Hour))
		claims.Issuer = "attacker"
		wrongIssuer := signWith(t, gojwt.SigningMethodRS256, "key-1", key.Private, claims)
		_, err := validator.ValidateToken(context.Background(), wrongIssuer)
		assertUnauthorized(t, err)
	})

	t.Run("when the algorithm is none (alg-confusion attack)", func(t *testing.T) {
		unsigned := signWith(t, gojwt.SigningMethodNone, "key-1", gojwt.UnsafeAllowNoneSignatureType, validClaims(time.Now().Add(time.Hour)))
		_, err := validator.ValidateToken(context.Background(), unsigned)
		assertUnauthorized(t, err)
	})

	t.Run("when the token is malformed", func(t *testing.T) {
		_, err := validator.ValidateToken(context.Background(), "not.a.jwt")
		assertUnauthorized(t, err)
	})
}

// assertUnauthorized asserts err is the uniform authentication_required domain
// error that ValidateToken returns for every rejection.
func assertUnauthorized(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var domainErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "authentication_required", domainErr.Code)
}
