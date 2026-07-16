package jwt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testKey generates a fresh RSA key pair and wraps it as an active ManagedKey.
// 2048 bits keeps generation fast while meeting the RS256 minimum.
func testKey(t *testing.T, keyID string) ManagedKey {
	t.Helper()
	private, err := rsa.GenerateKey(rand.Reader, minRSABits)
	require.NoError(t, err)
	return ManagedKey{KeyID: keyID, Private: private, Public: &private.PublicKey}
}

// pkcs1PEM encodes an RSA private key as a PKCS#1 PEM string.
func pkcs1PEM(t *testing.T, key *rsa.PrivateKey) string {
	t.Helper()
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	return string(pem.EncodeToMemory(block))
}

// pkcs8PEM encodes an RSA private key as a PKCS#8 PEM string.
func pkcs8PEM(t *testing.T, key *rsa.PrivateKey) string {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	block := &pem.Block{Type: "PRIVATE KEY", Bytes: der}
	return string(pem.EncodeToMemory(block))
}

func TestNewStaticKeyProvider_Errors(t *testing.T) {
	active := testKey(t, "key-a")
	retired := testKey(t, "key-b")
	retired.Retired = true

	noPrivate := testKey(t, "key-c")
	noPrivate.Private = nil

	tests := []struct {
		name    string
		primary string
		keys    []ManagedKey
	}{
		{
			name:    "when no keys are supplied",
			primary: "key-a",
			keys:    nil,
		},
		{
			name:    "when the primary key id is not in the set",
			primary: "missing",
			keys:    []ManagedKey{active},
		},
		{
			name:    "when the primary key is retired",
			primary: "key-b",
			keys:    []ManagedKey{retired},
		},
		{
			name:    "when the primary key has no private half",
			primary: "key-c",
			keys:    []ManagedKey{noPrivate},
		},
		{
			name:    "when two keys share an id",
			primary: "key-a",
			keys:    []ManagedKey{active, active},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewStaticKeyProvider(tt.primary, tt.keys...)
			assert.Error(t, err)
		})
	}
}

func TestStaticKeyProvider_SigningKey_ReturnsPrimary(t *testing.T) {
	primary := testKey(t, "primary")
	secondary := testKey(t, "secondary")

	provider, err := NewStaticKeyProvider("primary", primary, secondary)
	require.NoError(t, err)

	signingKey, err := provider.SigningKey(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "primary", signingKey.KeyID)
	assert.Equal(t, primary.Private, signingKey.Private)
}

func TestStaticKeyProvider_VerificationKey(t *testing.T) {
	active := testKey(t, "active")
	retired := testKey(t, "retired")
	retired.Retired = true

	provider, err := NewStaticKeyProvider("active", active, retired)
	require.NoError(t, err)

	t.Run("when the key id is active it returns the public key", func(t *testing.T) {
		pub, err := provider.VerificationKey(context.Background(), "active")
		require.NoError(t, err)
		assert.Equal(t, active.Public, pub)
	})

	t.Run("when the key id is unknown it returns ErrUnknownKeyID", func(t *testing.T) {
		_, err := provider.VerificationKey(context.Background(), "nope")
		assert.ErrorIs(t, err, ErrUnknownKeyID)
	})

	t.Run("when the key id is retired it returns ErrKeyRetired", func(t *testing.T) {
		_, err := provider.VerificationKey(context.Background(), "retired")
		assert.ErrorIs(t, err, ErrKeyRetired)
	})
}

func TestLoadKeyFromPEM(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, minRSABits)
	require.NoError(t, err)

	t.Run("when given a PKCS#1 PEM it derives the public key", func(t *testing.T) {
		managed, err := LoadKeyFromPEM("key-1", pkcs1PEM(t, key))
		require.NoError(t, err)
		assert.Equal(t, "key-1", managed.KeyID)
		assert.Equal(t, key.PublicKey, *managed.Public)
	})

	t.Run("when given a PKCS#8 PEM it derives the public key", func(t *testing.T) {
		managed, err := LoadKeyFromPEM("key-8", pkcs8PEM(t, key))
		require.NoError(t, err)
		assert.Equal(t, key.PublicKey, *managed.Public)
	})

	t.Run("when the key id is empty it errors", func(t *testing.T) {
		_, err := LoadKeyFromPEM("", pkcs1PEM(t, key))
		assert.Error(t, err)
	})

	t.Run("when the PEM is malformed it errors", func(t *testing.T) {
		_, err := LoadKeyFromPEM("key", "not a pem")
		assert.Error(t, err)
	})

	t.Run("when the RSA key is below the 2048-bit minimum it errors", func(t *testing.T) {
		weak, err := rsa.GenerateKey(rand.Reader, 1024)
		require.NoError(t, err)
		_, err = LoadKeyFromPEM("weak", pkcs1PEM(t, weak))
		assert.ErrorContains(t, err, "minimum")
	})
}
