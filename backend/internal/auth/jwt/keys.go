package jwt

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// minRSABits is the minimum RSA modulus size accepted for signing keys. RS256
// requires 2048 bits per specs/004-security-auth-model/data-model.md's
// JWTSigningKey validation rules.
const minRSABits = 2048

// Errors returned by KeyProvider implementations. Callers (the Validator)
// translate these into a generic "invalid token" response so no key-management
// detail leaks to clients.
var (
	// ErrUnknownKeyID is returned when a token's `kid` matches no key the
	// provider knows about.
	ErrUnknownKeyID = errors.New("jwt: unknown signing key id")
	// ErrKeyRetired is returned when a token's `kid` matches a key that has
	// been retired and is therefore no longer accepted for validation.
	ErrKeyRetired = errors.New("jwt: signing key is retired")
	// ErrNoSigningKey is returned when no primary key is available to sign new
	// tokens.
	ErrNoSigningKey = errors.New("jwt: no active signing key configured")
)

// SigningKey is the private half of a key pair used to sign new access tokens,
// paired with its KeyID so the generator can stamp the token's `kid` header.
type SigningKey struct {
	KeyID   string
	Private *rsa.PrivateKey
}

// ManagedKey is one entry in a KeyProvider's key set: an RSA key pair (the
// private half may be absent for verify-only keys) plus the rotation status
// that determines whether it may still validate tokens.
type ManagedKey struct {
	KeyID string
	// Private is the signing key; nil for keys retained only to validate
	// tokens that others signed (e.g. a public key imported for verification).
	Private *rsa.PrivateKey
	Public  *rsa.PublicKey
	// Retired mirrors jwt_signing_keys.status = 'retired': the key is kept for
	// audit but no longer accepted for validation. See docs/security.md.
	Retired bool
}

// KeyProvider supplies the key material the Generator and Validator need. It is
// the seam between this package's token logic and however keys are actually
// stored — a static in-memory set today (StaticKeyProvider), an AWS Secrets
// Manager-backed loader later (Spec 004 Phase 3). Implementations must be safe
// for concurrent use.
type KeyProvider interface {
	// SigningKey returns the current primary key used to sign new tokens, or
	// ErrNoSigningKey if none is available.
	SigningKey(ctx context.Context) (SigningKey, error)

	// VerificationKey returns the public key for the given `kid` if that key is
	// currently accepted for validation. It returns ErrUnknownKeyID for an
	// unrecognized id and ErrKeyRetired for a known-but-retired key.
	VerificationKey(ctx context.Context, keyID string) (*rsa.PublicKey, error)
}

// StaticKeyProvider is an in-memory KeyProvider backed by a fixed set of keys,
// with one designated as the primary signing key. It is the MVP/local-dev and
// test implementation; production key loading (Secrets Manager + the
// jwt_signing_keys table) will provide its own KeyProvider without any change
// to the Generator/Validator/Refresher.
//
// A StaticKeyProvider is immutable after construction and therefore safe for
// concurrent use.
type StaticKeyProvider struct {
	keys      map[string]ManagedKey
	primaryID string
}

// NewStaticKeyProvider builds a StaticKeyProvider from the given keys, treating
// primaryKeyID as the signing key for new tokens. Every non-retired key remains
// accepted for validation, which is what supports zero-downtime rotation: sign
// with the new primary while the previous key is still active.
//
// It errors if keys is empty, contains duplicate IDs, or if primaryKeyID does
// not name a usable (present, non-retired, private-key-bearing) signing key.
func NewStaticKeyProvider(primaryKeyID string, keys ...ManagedKey) (*StaticKeyProvider, error) {
	if len(keys) == 0 {
		return nil, errors.New("jwt: at least one key is required")
	}

	indexed := make(map[string]ManagedKey, len(keys))
	for _, k := range keys {
		if k.KeyID == "" {
			return nil, errors.New("jwt: key id must not be empty")
		}
		if _, dup := indexed[k.KeyID]; dup {
			return nil, fmt.Errorf("jwt: duplicate key id %q", k.KeyID)
		}
		if k.Public == nil {
			return nil, fmt.Errorf("jwt: key %q is missing its public key", k.KeyID)
		}
		indexed[k.KeyID] = k
	}

	primary, ok := indexed[primaryKeyID]
	switch {
	case !ok:
		return nil, fmt.Errorf("jwt: primary key %q not found in key set", primaryKeyID)
	case primary.Retired:
		return nil, fmt.Errorf("jwt: primary key %q is retired", primaryKeyID)
	case primary.Private == nil:
		return nil, fmt.Errorf("jwt: primary key %q has no private key to sign with", primaryKeyID)
	}

	return &StaticKeyProvider{keys: indexed, primaryID: primaryKeyID}, nil
}

// SigningKey returns the configured primary key.
func (p *StaticKeyProvider) SigningKey(_ context.Context) (SigningKey, error) {
	primary, ok := p.keys[p.primaryID]
	if !ok || primary.Private == nil {
		return SigningKey{}, ErrNoSigningKey
	}
	return SigningKey{KeyID: primary.KeyID, Private: primary.Private}, nil
}

// VerificationKey returns the public key for keyID when it is active.
func (p *StaticKeyProvider) VerificationKey(_ context.Context, keyID string) (*rsa.PublicKey, error) {
	key, ok := p.keys[keyID]
	if !ok {
		return nil, ErrUnknownKeyID
	}
	if key.Retired {
		return nil, ErrKeyRetired
	}
	return key.Public, nil
}

// LoadKeyFromPEM builds a ManagedKey from a PEM-encoded RSA private key,
// deriving the public half from it. It enforces the RS256 2048-bit minimum.
//
// This is the raw-value path for the current secrets model (issue #143): the
// private key arrives as a raw PEM string via configuration (JWT_SIGNING_KEY),
// whether supplied locally or injected by ECS from Secrets Manager. The ARN-
// resolving loader that reads jwt_signing_keys.private_key_secret_arn at runtime
// is a later task (Spec 004 Phase 3).
func LoadKeyFromPEM(keyID, privateKeyPEM string) (ManagedKey, error) {
	if keyID == "" {
		return ManagedKey{}, errors.New("jwt: key id must not be empty")
	}

	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return ManagedKey{}, errors.New("jwt: no PEM block found in private key")
	}

	private, err := parseRSAPrivateKey(block)
	if err != nil {
		return ManagedKey{}, err
	}
	if bits := private.N.BitLen(); bits < minRSABits {
		return ManagedKey{}, fmt.Errorf("jwt: RSA key is %d bits, minimum is %d", bits, minRSABits)
	}

	return ManagedKey{
		KeyID:   keyID,
		Private: private,
		Public:  &private.PublicKey,
	}, nil
}

// parseRSAPrivateKey decodes either PKCS#1 ("RSA PRIVATE KEY") or PKCS#8
// ("PRIVATE KEY") PEM blocks, the two formats openssl emits, and rejects
// anything that is not an RSA key.
func parseRSAPrivateKey(block *pem.Block) (*rsa.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("jwt: parse RSA private key: %w", err)
	}
	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("jwt: private key is %T, want *rsa.PrivateKey", parsed)
	}
	return rsaKey, nil
}
