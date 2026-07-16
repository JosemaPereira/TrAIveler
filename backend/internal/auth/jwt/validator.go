package jwt

import (
	"context"
	"errors"
	"fmt"

	gojwt "github.com/golang-jwt/jwt/v5"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// Validator verifies access tokens against a KeyProvider's active keys. It
// supports multi-key rotation: the signing key is chosen per-token from the
// `kid` header, so a token minted by a previous key keeps validating until that
// key is retired.
type Validator struct {
	keys   KeyProvider
	parser *gojwt.Parser
}

// NewValidator builds a Validator backed by provider.
//
// The parser is locked to RS256 only (WithValidMethods) to defend against
// algorithm-confusion attacks — most importantly a forged token declaring
// `alg: none` or a symmetric algorithm that would otherwise be verified with
// the RSA public key as an HMAC secret. Expiration and issuer are both required
// and checked.
func NewValidator(provider KeyProvider) *Validator {
	parser := gojwt.NewParser(
		gojwt.WithValidMethods([]string{gojwt.SigningMethodRS256.Alg()}),
		gojwt.WithExpirationRequired(),
		gojwt.WithIssuer(issuer),
	)
	return &Validator{keys: provider, parser: parser}
}

// ValidateToken verifies tokenString's signature and standard claims and, on
// success, returns its decoded Claims.
//
// Every failure — bad signature, expired/malformed token, unknown or retired
// signing key — collapses into a single generic "authentication_required"
// domain error so no key-management or parsing detail leaks to the caller. The
// underlying cause is wrapped for server-side logging via errors.Unwrap.
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

// invalidToken wraps the parse failure in an authentication_required domain
// error, keeping the cause available to errors.Is/As (and thus to logging)
// while presenting a uniform message to clients.
func invalidToken(cause error) error {
	domainErr := domainerrors.Unauthorized("invalid or expired token")
	domainErr.Err = fmt.Errorf("jwt: validate token: %w", cause)
	return domainErr
}
