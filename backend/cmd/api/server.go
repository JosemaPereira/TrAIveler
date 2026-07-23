package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/JosemaPereira/TrAIveler/backend/config"
	"github.com/JosemaPereira/TrAIveler/backend/internal/auth"
	"github.com/JosemaPereira/TrAIveler/backend/internal/auth/jwt"
	"github.com/JosemaPereira/TrAIveler/backend/internal/database"
	"github.com/JosemaPereira/TrAIveler/backend/internal/example"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
)

// healthzTimeout bounds the database ping issued by /healthz.
const healthzTimeout = 2 * time.Second

// version is the semantic version of the running binary. It is the single
// source of truth for both the startup log (main.go) and the /healthz
// response, so it can later be overridden at build time via
// `-ldflags "-X main.version=..."` without touching call sites.
var version = "0.1.0"

// HTTPServer wires the Chi router, database client, configuration, and
// structured logger together and exposes the API's operational HTTP
// endpoints. Domain routes are registered the same way in routes.go.
type HTTPServer struct {
	router         *chi.Mux
	db             database.Client
	cfg            *config.Config
	logger         *slog.Logger
	exampleHandler *example.Handler
	authHandler    *auth.Handler
	// authKeyProvider is retained for auth-activation (008-T207, issue #179): the
	// jwt.Validator behind the Authenticate gate must use the key set these
	// handlers sign with.
	authKeyProvider jwt.KeyProvider
	startTime       time.Time
}

// NewHTTPServer builds an HTTPServer with the standard middleware chain
// registered, in order: RequestID -> Logger -> Recovery -> [RateLimit] ->
// CORS -> BodySize. The order is deliberate: RequestID runs first so every
// later middleware (and the access log) can correlate by request ID; Recovery
// wraps everything downstream so a panic still yields a clean response; and
// RateLimit sits just inside Recovery so throttled requests are still logged
// and correlated but are rejected before any CORS/body handling work. The
// global rate limit is a coarse per-IP abuse guard, engaged only when
// cfg.RateLimit.Requests > 0 (disabled by default); stricter per-endpoint
// limits are added as route-level middleware in later work.
func NewHTTPServer(db database.Client, cfg *config.Config, logger *slog.Logger) (*HTTPServer, error) {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	if cfg.RateLimit.Requests > 0 {
		router.Use(middleware.RateLimit(cfg.RateLimit.Requests, cfg.RateLimit.Window))
	}
	router.Use(middleware.CORS(cfg.Server.AllowedCORS))
	router.Use(middleware.BodySize)

	// internal/example is the canonical layered-pattern reference (model ->
	// repository -> service -> handler); wiring it here is what makes it a
	// runnable demo instead of just unit-tested code in isolation. Future
	// domain packages (Trip, User, ...) follow the same three-line
	// construction and get mounted the same way in registerRoutes below.
	exampleRepo := example.NewPostgresRepository(db)
	exampleService := example.NewService(exampleRepo)
	exampleHandler := example.NewHandler(exampleService)

	// The auth vertical (register/login/logout) is composed in buildAuthComponents
	// (cmd/api/auth.go), which can fail on fatal misconfiguration.
	authComps, err := buildAuthComponents(cfg, db, logger)
	if err != nil {
		return nil, fmt.Errorf("build auth components: %w", err)
	}

	s := &HTTPServer{
		router:          router,
		db:              db,
		cfg:             cfg,
		logger:          logger,
		exampleHandler:  exampleHandler,
		authHandler:     authComps.handler,
		authKeyProvider: authComps.keyProvider,
		startTime:       time.Now(),
	}

	s.registerRoutes()

	return s, nil
}

// Router returns the underlying Chi router as an http.Handler, ready to be
// used as an *http.Server's Handler.
func (s *HTTPServer) Router() http.Handler {
	return s.router
}

// HealthCheckResponse is the JSON body returned by GET /healthz, matching
// spec 002's HealthCheckResponse contract (see
// specs/002-nfr-system-constraints/data-model.md).
type HealthCheckResponse struct {
	Status        string  `json:"status"`
	Version       string  `json:"version"`
	UptimeSeconds float64 `json:"uptime_seconds"`
}

// handleHealthz reports service health by pinging the database with a
// bounded timeout. Per spec 002's validation rules, the HTTP status code is
// always 200 OK — load balancers use the HTTP code for routing, so a failed
// or timed-out ping is only signaled via status:"degraded" in the body, not
// via a non-2xx HTTP status. The ping failure is still logged so it isn't a
// silent failure mode for operators.
func (s *HTTPServer) handleHealthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), healthzTimeout)
	defer cancel()

	status := "ok"
	if err := s.db.Ping(ctx); err != nil {
		s.logger.Warn("healthz database ping failed", "error", err)
		status = "degraded"
	}

	writeJSON(w, http.StatusOK, HealthCheckResponse{
		Status:        status,
		Version:       version,
		UptimeSeconds: time.Since(s.startTime).Seconds(),
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Default().Error("failed to write JSON response", "error", err)
	}
}
