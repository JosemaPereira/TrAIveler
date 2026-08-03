package main

import (
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	// Blank-imported so its init() registers the generated Swagger spec
	// with swaggo/swag's global spec registry (default instance name
	// "swagger") before httpSwagger.Handler looks it up. See
	// backend/docs/docs.go.
	_ "github.com/JosemaPereira/TrAIveler/backend/docs"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
)

// registerRoutes is the single place new domain endpoints (trips, auth,
// itinerary, ...) get added as the API grows. /healthz stays unversioned,
// public, and outside both groups; every domain route is mounted under /api/v1
// per docs/api-design-standards.md §3.
//
// The /api/v1 subtree is split in two (008-T208): a public group for the auth
// entry points that cannot require an access token — register, login, and
// refresh (008-T149: "no auth middleware, uses refresh token from cookie") —
// and an authenticated group carrying middleware.Authenticate for everything
// else. Trip and Conversation endpoints (001-T038/001-T039/001-T040) are
// fully authenticated (specs/001-product-vision-scope/contracts/api.md line
// 195: "All trip endpoints require authentication"), so both mount inside
// the authenticated group, with no public sub-routes of their own — unlike
// auth's public/protected split.
//
// /swagger/* is gated too, but only in production, which is how this
// delivers specs/009-api-documentation/research.md's "Auth gating for
// Swagger UI ahead of Sprint 5" decision without the onboarding friction it
// causes locally: a fresh checkout hitting /swagger/index.html got a 401
// before any login was even possible (issue #192 Task C). Outside
// production the route is mounted with no middleware at all.
//
// Every group here registers full paths (e.g. `/auth/login`, `/trips`, not
// a nested r.Route("/auth", ...) / r.Route("/trips", ...)) because Chi
// panics when the same pattern is routed twice on one tree, which sibling
// subtrees sharing a prefix (auth's public/protected split; trip.Handler and
// conversation.Handler both owning parts of /trips) would do.
func (s *HTTPServer) registerRoutes() {
	s.router.Get("/healthz", s.handleHealthz)

	authenticate := middleware.Authenticate(s.tokenValidator)

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Group(s.authHandler.RegisterPublicRoutes)

		r.Group(func(r chi.Router) {
			r.Use(authenticate)

			s.authHandler.RegisterProtectedRoutes(r)
			s.tripHandler.RegisterRoutes(r)
			s.conversationHandler.RegisterRoutes(r)
		})
	})

	// /swagger/* is a top-level path (not nested under /api/v1's URL prefix).
	// It only carries the same gate as the authenticated group above in
	// production; see the doc comment above for why.
	swaggerRoute := s.router.With()
	if s.cfg.IsProduction() {
		swaggerRoute = s.router.With(authenticate)
	}
	swaggerRoute.Get("/swagger/*", httpSwagger.Handler())
}
