# Contract: API Documentation Endpoints

This feature exposes two new HTTP surfaces on the existing backend service. Both are
documentation/tooling endpoints, not domain resources — they follow
`docs/api-design-standards.md` where applicable (versioning, error format) and are called out
here where they deliberately diverge (e.g., living outside `/api/v1`, like `/healthz` already
does).

## `GET /swagger/doc.json`

Serves the generated OpenAPI v3 document as JSON.

- **Auth**: Same as any other route mounted inside the shared route group (none today; JWT
  once Sprint 5 lands — see `research.md`).
- **Response 200**: `application/json` body containing the full OpenAPI v3 document (see
  `data-model.md` for the shape).
- **Response 401/403**: Once auth exists, same error format as every other protected endpoint
  per `docs/api-design-standards.md` §7 (standardized error envelope).

## `GET /swagger/index.html` (and other `/swagger/*` static assets)

Serves the Swagger UI single-page application, pre-configured to load `doc.json` from the
route above.

- **Auth**: Same as `/swagger/doc.json`.
- **Response 200**: `text/html` (and associated JS/CSS assets under `/swagger/*`).

## Annotation contract (developer-facing)

Every handler that should appear in the published contract MUST carry `swag` doc-comment
annotations directly above its function definition, following this shape (mirrors
`internal/example/handler.go` once updated):

```go
// CreateExample godoc
// @Summary     Create a new example
// @Description Creates an example resource for demonstration purposes
// @Tags        examples
// @Accept      json
// @Produce     json
// @Param       body body CreateExampleRequest true "Example payload"
// @Success     201 {object} ExampleResponse
// @Failure     400 {object} errors.ErrorResponse
// @Security    BearerAuth
// @Router      /examples [post]
func (h *Handler) CreateExample(w http.ResponseWriter, r *http.Request) { ... }
```

A handler with no such annotations is silently excluded from the contract (see `research.md`'s
"Excluding intentionally-undocumented endpoints" decision) — this is intentional and requires
no extra step for endpoints like `/healthz`.

## CI contract

`backend-ci.yml` gains a step, gated the same way as the rest of the required checks
(`dorny/paths-filter`, only runs when `backend/**` changes):

1. Run `make swagger` (wraps `swag init -g cmd/api/docs.go -o docs`).
2. Run `git diff --exit-code -- backend/docs`.
3. Fail the job (and therefore the PR's required check) if step 2 reports any difference.
