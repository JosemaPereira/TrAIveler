package main

import (
	crand "crypto/rand"
	"crypto/rsa"
	"fmt"
	"log/slog"

	"github.com/JosemaPereira/TrAIveler/backend/config"
	"github.com/JosemaPereira/TrAIveler/backend/internal/auth"
	"github.com/JosemaPereira/TrAIveler/backend/internal/auth/jwt"
	"github.com/JosemaPereira/TrAIveler/backend/internal/auth/ratelimit"
	"github.com/JosemaPereira/TrAIveler/backend/internal/database"
	"github.com/JosemaPereira/TrAIveler/backend/internal/subscription"
	"github.com/JosemaPereira/TrAIveler/backend/internal/subscription/payment"
)

const (
	// primaryKeyID labels the single configured JWT signing key. The multi-key
	// rotation machinery (docs/security.md) is future work; today there is one
	// primary key, whose id is stamped into every token's `kid` header and looked
	// up again for verification.
	primaryKeyID = "primary"
	// devEphemeralKeyID labels the throwaway key generated when no JWT_SIGNING_KEY
	// is configured — a development-only convenience (see buildKeyProvider).
	devEphemeralKeyID = "dev-ephemeral"
	// ephemeralKeyBits matches jwt's 2048-bit RS256 minimum.
	ephemeralKeyBits = 2048
)

// authComponents bundles the composed authentication surface. keyProvider is
// exposed so the auth-activation work (008-T207, issue #179) can build the
// jwt.Validator on the same key set these handlers sign with, rather than
// constructing a second provider.
type authComponents struct {
	handler     *auth.Handler
	keyProvider jwt.KeyProvider
}

// buildAuthComponents wires the authentication vertical from configuration and
// the database client: the JWT key provider and generator, the initial-session
// issuer over the refresh-token store, the subscription service (for paid
// registration), and the auth service and HTTP handler. It returns an error only
// for a genuinely fatal misconfiguration (e.g. an unparseable JWT_SIGNING_KEY).
func buildAuthComponents(cfg *config.Config, db database.Client, logger *slog.Logger) (*authComponents, error) {
	keyProvider, err := buildKeyProvider(cfg, logger)
	if err != nil {
		return nil, err
	}

	generator := jwt.NewGenerator(keyProvider, cfg.Auth.JWTExpiration)

	userRepo := auth.NewPostgresUserRepository(db)
	refreshRepo := auth.NewPostgresRefreshTokenRepository(db)
	issuer := jwt.NewIssuer(auth.NewRefreshStore(refreshRepo), generator, cfg.Auth.RefreshExpiration)

	subscriptionService := subscription.NewService(
		payment.NewStubPaymentProvider(logger),
		subscription.NewPostgresRepository(db),
	)

	authService := auth.NewService(userRepo, subscriptionService, ratelimit.New())

	handler := auth.NewHandler(authService, auth.NewJWTTokenIssuer(issuer), refreshRepo, auth.CookieConfig{
		Domain: cfg.Auth.CookieDomain,
		Secure: cfg.Auth.CookieSecure,
	})

	return &authComponents{handler: handler, keyProvider: keyProvider}, nil
}

// buildKeyProvider builds the JWT key provider from configuration. With a
// configured JWT_SIGNING_KEY it loads that raw PEM (the "raw now, ARN later"
// secrets model, issue #143). With none set — tolerated only outside production,
// where config.Load already requires it — it generates an ephemeral in-memory RSA
// key so local development can mint and verify tokens; that key resets on every
// restart, invalidating existing local sessions, which is acceptable for dev.
func buildKeyProvider(cfg *config.Config, logger *slog.Logger) (jwt.KeyProvider, error) {
	if cfg.Auth.JWTSigningKey != "" {
		key, err := jwt.LoadKeyFromPEM(primaryKeyID, cfg.Auth.JWTSigningKey)
		if err != nil {
			return nil, fmt.Errorf("load JWT signing key: %w", err)
		}
		return jwt.NewStaticKeyProvider(primaryKeyID, key)
	}

	logger.Warn("JWT_SIGNING_KEY not set — generating an ephemeral in-memory RSA key (development only; " +
		"tokens do not survive a restart)")
	private, err := rsa.GenerateKey(crand.Reader, ephemeralKeyBits)
	if err != nil {
		return nil, fmt.Errorf("generate ephemeral JWT key: %w", err)
	}
	return jwt.NewStaticKeyProvider(devEphemeralKeyID, jwt.ManagedKey{
		KeyID:   devEphemeralKeyID,
		Private: private,
		Public:  &private.PublicKey,
	})
}
