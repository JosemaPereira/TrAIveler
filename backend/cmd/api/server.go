package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/JosemaPereira/TrAIveler/backend/config"
	"github.com/JosemaPereira/TrAIveler/backend/internal/database"
	"github.com/JosemaPereira/TrAIveler/backend/internal/example"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
)

// healthzTimeout bounds the database ping issued by /healthz.
const healthzTimeout = 2 * time.Second

// HTTPServer wires the Chi router, database client, configuration, and
// structured logger together and exposes the API's operational HTTP
// endpoints. Domain routes are registered the same way in routes.go.
type HTTPServer struct {
	router         *chi.Mux
	db             database.Client
	cfg            *config.Config
	logger         *slog.Logger
	exampleHandler *example.Handler
}

// NewHTTPServer builds an HTTPServer with the standard middleware chain
// registered, in order: RequestID -> Logger -> Recovery -> CORS -> BodySize.
// The order is deliberate: RequestID runs first so every later middleware
// (and the access log) can correlate by request ID, and Recovery wraps
// CORS/BodySize/handlers so a downstream panic still yields a clean
// response instead of crashing the server.
func NewHTTPServer(db database.Client, cfg *config.Config, logger *slog.Logger) *HTTPServer {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
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

	s := &HTTPServer{
		router:         router,
		db:             db,
		cfg:            cfg,
		logger:         logger,
		exampleHandler: exampleHandler,
	}

	s.registerRoutes()

	return s
}

// Router returns the underlying Chi router as an http.Handler, ready to be
// used as an *http.Server's Handler.
func (s *HTTPServer) Router() http.Handler {
	return s.router
}

// healthzResponse is the JSON body returned by GET /healthz.
type healthzResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Error    string `json:"error,omitempty"`
}

// handleHealthz reports service health by pinging the database with a
// bounded timeout: 200 + {"status":"healthy","database":"connected"} when
// the ping succeeds, 503 + {"status":"unhealthy","database":"disconnected",
// "error":"<message>"} when it fails or times out.
func (s *HTTPServer) handleHealthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), healthzTimeout)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, healthzResponse{
			Status:   "unhealthy",
			Database: "disconnected",
			Error:    err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, healthzResponse{
		Status:   "healthy",
		Database: "connected",
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Default().Error("failed to write JSON response", "error", err)
	}
}
