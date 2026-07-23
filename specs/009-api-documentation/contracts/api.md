# Contract: API Documentation Endpoints

This feature exposes two new HTTP surfaces on the existing backend service. Both are
documentation/tooling endpoints, not domain resources — they follow
`docs/api-design-standards.md` where applicable (versioning, error format) and are called out
here where they deliberately diverge (e.g., living outside `/api/v1`, like `/healthz` already
does).

## `GET /swagger/doc.json`

Serves the generated Swagger 2.0 (OpenAPI 2.0) document as JSON.

- **Auth**: Required. Since issue #179 (008-T208) `/swagger/*` is registered with the same
  `middleware.Authenticate` chain as the authenticated `/api/v1` group, so it needs a valid
  `access_token` cookie — see `research.md`'s "Auth gating for Swagger UI ahead of Sprint 5"
  decision, which this delivers.
- **Response 200**: `application/json` body containing the full Swagger 2.0 (OpenAPI 2.0)
  document (see `data-model.md` for the shape).
- **Response 401**: The standard `authentication_required` envelope per
  `docs/api-design-standards.md` §7, identical to every other gated endpoint.

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
// @Failure     401 {object} errors.ErrorResponse
// @Security    CookieAuth
// @Router      /examples [post]
func (h *Handler) CreateExample(w http.ResponseWriter, r *http.Request) { ... }
```

`@Security CookieAuth` names the single security definition declared in `backend/cmd/api/docs.go`.
It was originally written `BearerAuth` (`in: header`, `name: Authorization`); since issue #179
(008-T208) mounted `middleware.Authenticate`, the definition is `CookieAuth` (`in: header`,
`name: Cookie`) because the server authenticates **only** on the HTTP-only `access_token` cookie
and has no `Authorization: Bearer` code path — see that file's comment for why Swagger 2.0 forces
the `Cookie`-header apiKey encoding. Add `@Security` to a handler if and only if it is mounted in
the authenticated route group, and pair it with `@Failure 401 {object} errors.ErrorResponse`; the
public auth entry points (`/auth/register`, `/auth/login`, `/auth/refresh`) carry neither.

A handler with no such annotations is silently excluded from the contract (see `research.md`'s
"Excluding intentionally-undocumented endpoints" decision) — this is intentional and requires
no extra step for endpoints like `/healthz`.

## CI contract

`backend-ci.yml` gains a step, gated the same way as the rest of the required checks
(`dorny/paths-filter`, only runs when `backend/**` changes):

1. Run `make swagger` (wraps `swag init -g cmd/api/docs.go -o docs`).
2. Run `git diff --exit-code -- backend/docs`.
3. Fail the job (and therefore the PR's required check) if step 2 reports any difference.
