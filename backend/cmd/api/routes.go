package main

import "github.com/go-chi/chi/v5"

// registerRoutes is the single place new domain endpoints (trips, auth,
// itinerary, ...) get added as the API grows, keeping route wiring
// centralized. /healthz stays unversioned and outside the /api/v1 group;
// every domain route is mounted under /api/v1 per
// docs/api-design-standards.md §3.
func (s *HTTPServer) registerRoutes() {
	s.router.Get("/healthz", s.handleHealthz)

	s.router.Route("/api/v1", func(r chi.Router) {
		s.exampleHandler.RegisterRoutes(r)
	})
}
