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
	// primaryKeyID labels the single configured JWT signing key; multi-key rotation
	// (docs/security.md) is future work. Its id is stamped into each token's `kid`.
	primaryKeyID = "primary"
	// devEphemeralKeyID labels the throwaway key generated when no JWT_SIGNING_KEY
	// is configured — a development-only convenience (see buildKeyProvider).
	devEphemeralKeyID = "dev-ephemeral"
	// ephemeralKeyBits matches jwt's 2048-bit RS256 minimum.
	ephemeralKeyBits = 2048
)

// authComponents bundles the composed authentication surface. keyProvider is
// exposed so NewHTTPServer can build the Authenticate gate's jwt.Validator on
// the same key set these handlers sign with, rather than a second one
// (008-T207).
type authComponents struct {
	handler     *auth.Handler
	keyProvider jwt.KeyProvider
}

// buildAuthComponents wires the authentication vertical from config and the
// database client (JWT key provider and generator, session issuer and
// refresher, subscription service, auth service, and HTTP handler). It errors
// only on genuinely fatal misconfiguration (e.g. an unparseable
// JWT_SIGNING_KEY).
func buildAuthComponents(cfg *config.Config, db database.Client, logger *slog.Logger) (*authComponents, error) {
	keyProvider, err := buildKeyProvider(cfg, logger)
	if err != nil {
		return nil, err
	}

	generator := jwt.NewGenerator(keyProvider, cfg.Auth.JWTExpiration)

	userRepo := auth.NewPostgresUserRepository(db)
	refreshRepo := auth.NewPostgresRefreshTokenRepository(db)
	refreshStore := auth.NewRefreshStore(refreshRepo)
	issuer := jwt.NewIssuer(refreshStore, generator, cfg.Auth.RefreshExpiration)

	// One subscription repository serves both the registration flow and the
	// refresh flow's resolver, so a refreshed token re-reads the same rows the
	// service writes (008-T211/T212).
	subscriptionRepo := subscription.NewPostgresRepository(db)
	subscriptionService := subscription.NewService(payment.NewStubPaymentProvider(logger), subscriptionRepo)

	// The Refresher rotates refresh tokens and re-resolves has_subscription on
	// every use, so a lapsed subscription takes effect within one access-token
	// lifetime (008-T212, FR-022).
	refresher := jwt.NewRefresher(
		refreshStore, generator, subscription.NewResolver(subscriptionRepo), cfg.Auth.RefreshExpiration,
	)

	authService := auth.NewService(userRepo, subscriptionService, ratelimit.New(), cfg.Auth.BcryptCost)

	handler := auth.NewHandler(
		authService,
		auth.NewJWTTokenIssuer(issuer),
		auth.NewJWTTokenRefresher(refresher),
		refreshRepo,
		auth.CookieConfig{
			Domain: cfg.Auth.CookieDomain,
			Secure: cfg.Auth.CookieSecure,
		},
	)

	return &authComponents{handler: handler, keyProvider: keyProvider}, nil
}

// buildKeyProvider builds the JWT key provider from config. With JWT_SIGNING_KEY
// set it loads that raw PEM ("raw now, ARN later", issue #143). With none — only
// tolerated outside production, where config.Load already requires it — it
// generates an ephemeral in-memory RSA key so local dev can mint/verify tokens;
// that key resets on each restart, invalidating local sessions (acceptable).
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
