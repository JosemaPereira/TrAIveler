package jwt

import gojwt "github.com/golang-jwt/jwt/v5"

// issuer is the `iss` claim stamped on every access token and required by the
// validator. A fixed, non-secret string that ties tokens to this application.
const issuer = "traiveler"

// Claims are the decoded contents of a TrAIveler access token. The user's ID is
// carried in the standard Subject (`sub`) claim; HasSubscription is the single
// application-specific claim, letting the auth middleware gate subscription-only
// features without a database lookup on every request.
//
// Only non-sensitive authorization facts belong here: the token is signed, not
// encrypted, so its payload is readable by anyone holding it.
type Claims struct {
	// HasSubscription reports whether the user had an active subscription at
	// the moment the token was issued. It can go stale within the token's
	// lifetime; flows that must react immediately to a subscription change
	// revoke the user's refresh tokens rather than trusting this claim.
	HasSubscription bool `json:"has_subscription"`

	gojwt.RegisteredClaims
}
