// This file holds only swag general-API annotations — it intentionally
// contains no executable code. `swag init -g cmd/api/docs.go -o docs` (the
// `make swagger` target) parses the annotations below into the package-level
// OpenAPI metadata (title, version, description, base path, and the shared
// bearer-token security definition) for the generated contract in
// backend/docs/. Endpoint-level annotations live directly above each handler
// function, per swag convention — see internal/example/handler.go once it
// gains its own annotations in a later phase of specs/009-api-documentation/.
// The package doc comment for `main` itself lives in cmd/api/main.go.
//
// @title TrAIveler API
// @version 1.0
// @description Machine-readable OpenAPI contract for the TrAIveler backend API, generated from Go doc-comment annotations colocated with each handler (see specs/009-api-documentation/).
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main
