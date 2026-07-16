package jwt

import (
	"context"
	"errors"
	"fmt"

	gojwt "github.com/golang-jwt/jwt/v5"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// Validator verifies access tokens against a KeyProvider's active keys,
// selecting the key per-token from the `kid` header so tokens from a still-
// active previous key keep validating during rotation.
type Validator struct {
	keys   KeyProvider
	parser *gojwt.Parser
}

// NewValidator builds a Validator locked to RS256 (WithValidMethods) to block
// algorithm-confusion attacks — chiefly a forged `alg: none` or symmetric token
// verified with the RSA public key as an HMAC secret. Expiration and issuer are
// required.
func NewValidator(provider KeyProvider) *Validator {
	parser := gojwt.NewParser(
		gojwt.WithValidMethods([]string{gojwt.SigningMethodRS256.Alg()}),
		gojwt.WithExpirationRequired(),
		gojwt.WithIssuer(issuer),
	)
	return &Validator{keys: provider, parser: parser}
}

// ValidateToken verifies tokenString and returns its Claims. Every failure —
// bad signature, expired/malformed token, unknown or retired key — collapses
// into a generic authentication_required error; the cause is wrapped for logs.
func (v *Validator) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	claims := &Claims{}

	_, err := v.parser.ParseWithClaims(tokenString, claims, func(token *gojwt.Token) (any, error) {
		keyID, ok := token.Header["kid"].(string)
		if !ok || keyID == "" {
			return nil, errors.New("jwt: token is missing its kid header")
		}
		return v.keys.VerificationKey(ctx, keyID)
	})
	if err != nil {
		return nil, invalidToken(err)
	}

	return claims, nil
}

// invalidToken wraps a parse failure in an authentication_required error,
// keeping the cause visible to errors.Is/As and logs.
func invalidToken(cause error) error {
	domainErr := domainerrors.Unauthorized("invalid or expired token")
	domainErr.Err = fmt.Errorf("jwt: validate token: %w", cause)
	return domainErr
}
