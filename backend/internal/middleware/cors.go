package middleware

import (
	"net/http"
	"strings"
)

const (
	allowedMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	allowedHeaders = "Content-Type, Authorization, X-Request-ID"
)

// CORS returns middleware that enforces a cross-origin resource sharing
// allow-list built from allowedOrigins, a comma-separated list of exact
// origins (typically sourced from config.Config.Server.AllowedCORS by the
// caller). Requests without an Origin header, or with a disallowed one, pass
// through unmodified so same-origin and non-browser clients are unaffected.
func CORS(allowedOrigins string) func(http.Handler) http.Handler {
	origins := parseAllowedOrigins(allowedOrigins)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && origins[origin] {
				setCORSHeaders(w, origin)
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func setCORSHeaders(w http.ResponseWriter, origin string) {
	h := w.Header()
	h.Set("Access-Control-Allow-Origin", origin)
	h.Set("Access-Control-Allow-Methods", allowedMethods)
	h.Set("Access-Control-Allow-Headers", allowedHeaders)
	h.Set("Access-Control-Allow-Credentials", "true")
}

func parseAllowedOrigins(raw string) map[string]bool {
	origins := make(map[string]bool)
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins[origin] = true
		}
	}
	return origins
}
