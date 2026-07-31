//go:build test

package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/internal/auth/jwt"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
)

// jwtIssuer mirrors the unexported issuer constant in internal/auth/jwt: the
// real *jwt.Validator requires the `iss` claim to match, so the expired token
// this test signs by hand must carry it.
const jwtIssuer = "traiveler"

// fakeClaimsValidator is a hand fake for the one-method claimsValidator port
// (preferred over a generated mock for internal ports, per
// docs/mock-standards.md). It keeps the adapter test deterministic: no RSA keys,
// no signing, no clock.
type fakeClaimsValidator struct {
	claims    *jwt.Claims
	err       error
	gotToken  string
	callCount int
}

func (f *fakeClaimsValidator) ValidateToken(_ context.Context, token string) (*jwt.Claims, error) {
	f.callCount++
	f.gotToken = token
	return f.claims, f.err
}

func claimsFor(subject string, hasSubscription bool) *jwt.Claims {
	return &jwt.Claims{
		HasSubscription:  hasSubscription,
		RegisteredClaims: gojwt.RegisteredClaims{Subject: subject},
	}
}

func TestJWTTokenValidator_ValidateToken_MapsClaimsOntoAuthClaims(t *testing.T) {
	errInvalidToken := errors.New("invalid or expired token")

	testCases := []struct {
		name        string
		claims      *jwt.Claims
		validateErr error
		wantClaims  middleware.AuthClaims
		wantErr     bool
	}{
		{
			name:       "when the token is valid and the user has a subscription",
			claims:     claimsFor("11111111-1111-1111-1111-111111111111", true),
			wantClaims: middleware.AuthClaims{UserID: "11111111-1111-1111-1111-111111111111", HasSubscription: true},
		},
		{
			name:       "when the token is valid and the user has no subscription",
			claims:     claimsFor("22222222-2222-2222-2222-222222222222", false),
			wantClaims: middleware.AuthClaims{UserID: "22222222-2222-2222-2222-222222222222", HasSubscription: false},
		},
		{
			name:        "when the underlying validator rejects the token",
			validateErr: errInvalidToken,
			wantErr:     true,
		},
		{
			name:    "having a validator that returns no claims and no error",
			claims:  nil,
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			fake := &fakeClaimsValidator{claims: testCase.claims, err: testCase.validateErr}
			adapter := newJWTTokenValidator(fake)

			// Act
			got, err := adapter.ValidateToken(context.Background(), "raw.access.token")

			// Assert
			assert.Equal(t, 1, fake.callCount, "the adapter must delegate to the wrapped validator")
			assert.Equal(t, "raw.access.token", fake.gotToken, "the raw token must be forwarded unchanged")

			if testCase.wantErr {
				require.Error(t, err)
				assert.Equal(t, middleware.AuthClaims{}, got, "a failed validation must not leak partial claims")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.wantClaims, got)
		})
	}
}

func TestJWTTokenValidator_ValidateToken_PropagatesTheUnderlyingError(t *testing.T) {
	// Arrange
	sentinel := errors.New("boom")
	adapter := newJWTTokenValidator(&fakeClaimsValidator{err: sentinel})

	// Act
	_, err := adapter.ValidateToken(context.Background(), "raw.access.token")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel, "the cause must stay inspectable for logs and errors.Is")
	assert.NotErrorIs(t, err, middleware.ErrTokenExpired,
		"a generic validation failure must NOT be reported as an expiry (stays authentication_required)")
}

// TestJWTTokenValidator_ValidateToken_ExpiryMapping proves the adapter translates
// only the JWT library's expiry error (gojwt.ErrTokenExpired), however deeply
// wrapped, into middleware.ErrTokenExpired — the signal the Authenticate gate uses
// to emit token_expired. Every other error passes through and stays
// authentication_required.
func TestJWTTokenValidator_ValidateToken_ExpiryMapping(t *testing.T) {
	testCases := []struct {
		name        string
		validateErr error
		wantExpired bool
	}{
		{
			name:        "when the validator error wraps gojwt.ErrTokenExpired it maps to ErrTokenExpired",
			validateErr: fmt.Errorf("jwt: validate token: %w", gojwt.ErrTokenExpired),
			wantExpired: true,
		},
		{
			name:        "when the validator error is a bare gojwt.ErrTokenExpired it maps to ErrTokenExpired",
			validateErr: gojwt.ErrTokenExpired,
			wantExpired: true,
		},
		{
			name:        "when the validator error is any other failure it does not map to ErrTokenExpired",
			validateErr: errors.New("invalid signature"),
			wantExpired: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			adapter := newJWTTokenValidator(&fakeClaimsValidator{err: testCase.validateErr})

			// Act
			_, err := adapter.ValidateToken(context.Background(), "raw.access.token")

			// Assert
			require.Error(t, err)
			if testCase.wantExpired {
				assert.ErrorIs(t, err, middleware.ErrTokenExpired)
			} else {
				assert.NotErrorIs(t, err, middleware.ErrTokenExpired)
			}
		})
	}
}

// TestJWTTokenValidator_ValidateToken_RealValidatorExpiredToken is the load-bearing
// integration point: it drives the *real* *jwt.Validator with a genuinely-expired
// but validly-signed token and asserts the adapter reports middleware.ErrTokenExpired.
// This proves the whole error chain is errors.Is-detectable for expiry —
// DomainError.Unwrap -> fmt %w -> gojwt's joined validation error -> gojwt.ErrTokenExpired —
// rather than trusting a hand-rolled wrapper. A valid signature is required for the
// parser to even reach expiry validation, which is exactly why this leaks nothing:
// only a token we truly signed can qualify.
func TestJWTTokenValidator_ValidateToken_RealValidatorExpiredToken(t *testing.T) {
	// Arrange: a real key, a real validator, and a hand-signed expired token.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	const keyID = "key-1"
	provider, err := jwt.NewStaticKeyProvider(keyID, jwt.ManagedKey{
		KeyID:   keyID,
		Private: privateKey,
		Public:  &privateKey.PublicKey,
	})
	require.NoError(t, err)

	now := time.Now()
	claims := jwt.Claims{
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   "44444444-4444-4444-4444-444444444444",
			Issuer:    jwtIssuer,
			IssuedAt:  gojwt.NewNumericDate(now.Add(-2 * time.Hour)),
			NotBefore: gojwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: gojwt.NewNumericDate(now.Add(-time.Hour)), // already expired
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyID
	signed, err := token.SignedString(privateKey)
	require.NoError(t, err)

	adapter := newJWTTokenValidator(jwt.NewValidator(provider))

	// Act
	_, err = adapter.ValidateToken(context.Background(), signed)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, middleware.ErrTokenExpired,
		"a validly-signed but expired token must surface as middleware.ErrTokenExpired through the real validator")
}
