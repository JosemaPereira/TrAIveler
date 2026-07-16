package jwt

import gojwt "github.com/golang-jwt/jwt/v5"

// issuer is the `iss` claim stamped on every token and required by the validator.
const issuer = "traiveler"

// Claims are the decoded contents of an access token. The user ID rides in the
// standard Subject claim; HasSubscription is the only app-specific claim, so the
// auth middleware can gate subscription features without a database lookup.
//
// The token is signed, not encrypted: only non-sensitive authorization facts
// belong here.
type Claims struct {
	// HasSubscription reflects the user's subscription at issue time and can go
	// stale within the token's lifetime; flows needing an immediate effect
	// revoke refresh tokens rather than trusting this claim.
	HasSubscription bool `json:"has_subscription"`

	gojwt.RegisteredClaims
}
