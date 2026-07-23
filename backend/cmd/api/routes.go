package main

import (
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	// Blank-imported so its init() registers the generated Swagger spec
	// with swaggo/swag's global spec registry (default instance name
	// "swagger") before httpSwagger.Handler looks it up. See
	// backend/docs/docs.go.
	_ "github.com/JosemaPereira/TrAIveler/backend/docs"
)

// registerRoutes is the single place new domain endpoints (trips, auth,
// itinerary, ...) get added as the API grows, keeping route wiring
// centralized. /healthz stays unversioned and outside the /api/v1 group;
// every domain route is mounted under /api/v1 per
// docs/api-design-standards.md §3.
func (s *HTTPServer) registerRoutes() {
	s.router.Get("/healthz", s.handleHealthz)

	// /swagger/* is mounted as a top-level path (not nested under
	// /api/v1's URL prefix) but inside the same Chi route Group as the
	// /api/v1 domain routes, per specs/009-api-documentation/research.md's
	// "Auth gating for Swagger UI ahead of Sprint 5" decision: any
	// authentication middleware Sprint 5 adds to this shared group will
	// automatically cover /swagger/* too, with no change needed here.
	//
	// TODO(sprint-5): remove once JWT middleware is wired — until then,
	// /swagger/* is intentionally unauthenticated. No interim bespoke auth
	// is added because production is currently documented as dormant
	// (docs/cloud-and-environments.md) and no other endpoint has auth yet
	// either.
	s.router.Group(func(r chi.Router) {
		r.Route("/api/v1", func(r chi.Router) {
			s.exampleHandler.RegisterRoutes(r)
			// Auth entry points (register/login public; logout needs a valid access
			// token). The Authenticate gate isn't mounted yet — 008-T208 (issue #179)
			// will split this into public and authenticated groups, moving logout
			// (and /swagger/*) behind it. Until then all /api/v1 routes are ungated,
			// matching the TODO(sprint-5) below.
			s.authHandler.RegisterRoutes(r)
		})

		r.Get("/swagger/*", httpSwagger.Handler())
	})
}
