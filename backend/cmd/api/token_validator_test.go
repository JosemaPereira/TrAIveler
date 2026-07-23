//go:build test

package main

import (
	"context"
	"errors"
	"testing"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/internal/auth/jwt"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
)

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
}
