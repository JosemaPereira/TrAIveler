package jwt

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// minRSABits is the RS256 minimum modulus size (data-model.md, JWTSigningKey).
const minRSABits = 2048

// KeyProvider errors. The Validator maps these to a generic "invalid token" so
// no key-management detail leaks to clients.
var (
	ErrUnknownKeyID = errors.New("jwt: unknown signing key id")
	ErrKeyRetired   = errors.New("jwt: signing key is retired")
	ErrNoSigningKey = errors.New("jwt: no active signing key configured")
)

// SigningKey is a private key plus its KeyID, used to sign new tokens and stamp
// the `kid` header.
type SigningKey struct {
	KeyID   string
	Private *rsa.PrivateKey
}

// ManagedKey is one entry in a KeyProvider's key set.
type ManagedKey struct {
	KeyID string
	// Private is nil for verify-only keys.
	Private *rsa.PrivateKey
	Public  *rsa.PublicKey
	// Retired mirrors jwt_signing_keys.status='retired': kept for audit, no
	// longer accepted for validation.
	Retired bool
}

// KeyProvider supplies key material to the Generator and Validator — the seam
// between token logic and key storage. Implementations must be concurrency-safe.
type KeyProvider interface {
	// SigningKey returns the current primary key, or ErrNoSigningKey.
	SigningKey(ctx context.Context) (SigningKey, error)
	// VerificationKey returns the public key for keyID, or ErrUnknownKeyID /
	// ErrKeyRetired if it is unknown or retired.
	VerificationKey(ctx context.Context, keyID string) (*rsa.PublicKey, error)
}

// StaticKeyProvider is an immutable, concurrency-safe in-memory KeyProvider with
// one designated primary key. It is the MVP/test implementation; production key
// loading provides its own KeyProvider without touching the token logic.
type StaticKeyProvider struct {
	keys      map[string]ManagedKey
	primaryID string
}

// NewStaticKeyProvider builds a provider from keys, using primaryKeyID to sign
// new tokens. Every non-retired key stays valid for verification, which is what
// enables rotation. It errors on an empty set, duplicate IDs, or a primary that
// is missing, retired, or lacks a private key.
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

// LoadKeyFromPEM builds a ManagedKey from a PEM-encoded RSA private key, deriving
// the public half and enforcing the 2048-bit minimum.
//
// This is the raw-value path for the current secrets model (issue #143): the key
// arrives as a raw PEM string via config, supplied locally or injected by ECS
// from Secrets Manager. The ARN-resolving loader is a later task.
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

	return ManagedKey{KeyID: keyID, Private: private, Public: &private.PublicKey}, nil
}

// parseRSAPrivateKey decodes a PKCS#1 or PKCS#8 RSA key, rejecting non-RSA keys.
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
