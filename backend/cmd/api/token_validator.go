package main

import (
	"context"
	"errors"
	"fmt"

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
// as a failure rather than yielding empty claims. Either way Authenticate
// collapses the result into the uniform 401 envelope, so the cause only ever
// reaches the logs.
func (a *jwtTokenValidator) ValidateToken(ctx context.Context, token string) (middleware.AuthClaims, error) {
	claims, err := a.validator.ValidateToken(ctx, token)
	if err != nil {
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
