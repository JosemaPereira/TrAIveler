package main

import (
	"context"
	"errors"
	"fmt"

	gojwt "github.com/golang-jwt/jwt/v5"

	"github.com/JosemaPereira/TrAIveler/backend/internal/auth/jwt"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
)

// claimsValidator is the narrow view of *jwt.Validator this adapter needs. It
// exists so the adapter is unit-testable with a hand fake (no RSA keys, no
// clock); production always passes the real *jwt.Validator.
type claimsValidator interface {
	ValidateToken(ctx context.Context, token string) (*jwt.Claims, error)
}

// The real validator must always satisfy the port the adapter wraps.
var _ claimsValidator = (*jwt.Validator)(nil)

// jwtTokenValidator adapts the auth/jwt validator to middleware.TokenValidator
// (008-T207). It has to live in the composition root: internal/middleware cannot
// import internal/auth/jwt, because auth/jwt -> internal/errors ->
// internal/middleware is a real import cycle. The middleware-local AuthClaims
// port and TokenValidator interface exist precisely to break it, and this is the
// single place the two concrete types are joined.
type jwtTokenValidator struct {
	validator claimsValidator
}

// newJWTTokenValidator wires a middleware.TokenValidator backed by the JWT
// validator, so the Authenticate gate verifies tokens against the same key set
// the auth handlers sign with.
func newJWTTokenValidator(validator claimsValidator) middleware.TokenValidator {
	return &jwtTokenValidator{validator: validator}
}

// ValidateToken verifies the raw access token and maps *jwt.Claims onto the
// middleware's local AuthClaims. The user id rides in the standard Subject
// claim. Failures are wrapped for context but keep the cause inspectable via
// errors.Is/As; a validator that returns neither claims nor an error is treated
// as a failure rather than yielding empty claims.
//
// Expiry is singled out: a validly-signed-but-expired token surfaces from the JWT
// layer as an error wrapping gojwt.ErrTokenExpired (the auth/jwt validator wraps
// its cause with %w, and internal/errors.DomainError.Unwrap exposes it), so this
// is the one composition-root seam where auth/jwt and internal/middleware meet and
// the two can be joined. We translate it into middleware.ErrTokenExpired, which
// Authenticate answers with 401 token_expired (008-T150). Every other failure
// passes through unchanged and Authenticate collapses it into the uniform
// authentication_required envelope, so the cause only ever reaches the logs.
func (a *jwtTokenValidator) ValidateToken(ctx context.Context, token string) (middleware.AuthClaims, error) {
	claims, err := a.validator.ValidateToken(ctx, token)
	if err != nil {
		if errors.Is(err, gojwt.ErrTokenExpired) {
			return middleware.AuthClaims{}, fmt.Errorf("validate access token: %w", middleware.ErrTokenExpired)
		}
		return middleware.AuthClaims{}, fmt.Errorf("validate access token: %w", err)
	}
	if claims == nil {
		return middleware.AuthClaims{}, errors.New("validate access token: validator returned no claims")
	}

	return middleware.AuthClaims{
		UserID:          claims.Subject,
		HasSubscription: claims.HasSubscription,
	}, nil
}
