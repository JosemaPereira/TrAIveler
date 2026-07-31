// This file holds only swag general-API annotations — it intentionally
// contains no executable code. `swag init -g cmd/api/docs.go -o docs` (the
// `make swagger` target) parses the annotations below into the package-level
// OpenAPI metadata (title, version, description, base path, and the shared
// security definition) for the generated contract in backend/docs/.
// Endpoint-level annotations live directly above each handler function, per
// swag convention — see internal/example/handler.go, the canonical reference
// for the annotation shape.
// The package doc comment for `main` itself lives in cmd/api/main.go.
//
// The security definition is declared as an apiKey carried in the `Cookie`
// request header because that is literally how this API authenticates:
// middleware.Authenticate reads the HTTP-only `access_token` cookie and
// nothing else — there is no `Authorization: Bearer` code path at all
// (docs/security.md, "Token Storage"). Swagger 2.0 has no cookie security
// scheme (that arrived with OpenAPI 3's `in: cookie`), so an apiKey in the
// `Cookie` header is the closest valid Swagger 2.0 encoding of the real
// credential. Do not "correct" it back to a bearer scheme.
//
// @title TrAIveler API
// @version 1.0
// @description Machine-readable OpenAPI contract for the TrAIveler backend API, generated from Go doc-comment annotations colocated with each handler (see specs/009-api-documentation/).
// @BasePath /api/v1
// @securityDefinitions.apikey CookieAuth
// @in header
// @name Cookie
// @description Session authentication via the HTTP-only `access_token` cookie set by `POST /api/v1/auth/login`, `POST /api/v1/auth/register`, and `POST /api/v1/auth/refresh`. Send it as `Cookie: access_token=<jwt>`. The "Authorize" box cannot supply this credential — the cookie is HttpOnly and `Cookie` is a forbidden header for browser-issued requests — but Swagger UI is served from the same origin behind the same gate, so a browser that reached this page already holds the cookie and sends it automatically on every "Try it out" request.
package main
