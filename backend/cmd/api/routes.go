package main

// registerRoutes is the single place new domain endpoints (trips, auth,
// itinerary, ...) get added as the API grows, keeping route wiring
// centralized.
func (s *HTTPServer) registerRoutes() {
	s.router.Get("/healthz", s.handleHealthz)
}
