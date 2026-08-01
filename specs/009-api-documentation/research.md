# Phase 0 Research: API Documentation via OpenAPI/Swagger

All Technical Context unknowns were resolved during `/speckit-clarify` (generation mechanism,
UI tool, environment exposure). This phase resolves the remaining implementation-level
decisions needed before design.

## Decision: Contract generation library

- **Decision**: `github.com/swaggo/swag` (CLI `swag init`), parsing Go doc-comment annotations
  on each Chi handler function into a Swagger 2.0 (OpenAPI 2.0) document (`swagger.json`/
  `swagger.yaml` + a generated `docs.go` Go package for embedding). `swag init` has no OpenAPI
  v3 output mode (confirmed against the installed CLI, `swag init --help`) — Swagger 2.0 is
  the tool's native, non-negotiable output format, not a configuration choice.
- **Rationale**: Confirmed by the Sprint 3 planning clarification (see spec's `## Clarifications`).
  It is the most widely-adopted Go OpenAPI generator, has first-class Chi support via
  `swaggo/http-swagger`, requires zero new abstraction (annotations live directly above the
  handler they describe, the same place a doc comment would already go per
  `docs/coding-guidelines.md`), and satisfies FR-002 (no separate manual authoring step).
- **Alternatives considered**:
  - Hand-maintained `openapi.yaml` — rejected: reintroduces the exact manual-sync risk (Markdown
    drift) this feature exists to remove.
  - `oapi-codegen` (contract-first: hand-write `openapi.yaml`, generate Go types/server
    interfaces from it) — rejected: inverts the project's existing code-first handler pattern
    (`internal/example/handler.go` already defines Go request/response types directly), would
    require rewriting existing handlers to conform to codegen'd interfaces for no added value
    at this project's scale.

## Decision: Documentation UI

- **Decision**: Swagger UI, served via `github.com/swaggo/http-swagger/v2`
  (`httpSwagger.Handler(...)` mounted as a Chi route).
- **Rationale**: Confirmed by clarification — Swagger UI supports "Try it out" (execute real
  requests), which Redoc does not; `http-swagger` consumes the `swag`-generated artifact
  directly with no extra conversion step.
- **Alternatives considered**: Redoc (read-only, rejected per clarification); a custom-built UI
  (rejected — violates Simplicity/KISS, no reason to build what `swaggo/http-swagger` already
  provides).

## Decision: Where global OpenAPI metadata lives

- **Decision**: A new `backend/cmd/api/docs.go` file holds only the package-level `swag`
  general-API annotations (`@title`, `@version`, `@description`, `@BasePath /api/v1`,
  `@securityDefinitions.apikey BearerAuth`, `@in header`, `@name Authorization`) — no executable
  code. Endpoint-level annotations live directly above each handler function, per `swag`
  convention.
- **Rationale**: Keeps global contract metadata in exactly one discoverable place (mirrors how
  `cmd/api/routes.go` is already "the single place new domain endpoints get added", per its own
  doc comment) rather than scattering `@title`/`@version` across handler files.
- **Alternatives considered**: Embedding general annotations in `main.go` — rejected, `main.go`
  should stay focused on process bootstrap, not OpenAPI metadata.

## Decision: Excluding intentionally-undocumented endpoints (FR-006)

- **Decision**: No exclusion mechanism is built. `swag` only includes handlers that carry its
  doc-comment annotations — `/healthz` (and any future infra-only endpoint) is excluded simply
  by never annotating it. This is already how `internal/errors` and infra endpoints are written
  today (undocumented by choice), so no behavior change is needed there.
- **Rationale**: Satisfies FR-006 with zero added configuration — the opt-in nature of `swag`
  annotations is itself the exclusion mechanism. Avoids inventing an allow/deny-list that would
  need independent maintenance.
- **Alternatives considered**: An explicit `// +swagger:exclude` marker or config allow-list —
  rejected as unnecessary complexity (Principle II) given the opt-in default already does this.

## Decision: CI drift-check mechanism (FR-008)

- **Decision**: `backend/docs/` (the generated `docs.go`, `swagger.json`, `swagger.yaml`) is
  committed to the repository, like other generated artifacts in this codebase (e.g., mocks
  under `docs/mock-standards.md`). A new `backend-ci.yml` step runs `make swagger` (which wraps
  `swag init`) and then `git diff --exit-code -- backend/docs`, failing the job if regeneration
  produces a diff.
- **Rationale**: Committing the artifact is what makes User Story 1's acceptance scenario 2
  possible ("a reviewer can see the change in the same code review as the implementation") —
  an artifact regenerated only at runtime and never diffed leaves no PR-visible trail and no
  way to gate merges (FR-008 explicitly requires a PR-blocking check). Matches the existing
  `dorny/paths-filter`-gated, always-running required-check pattern already used for
  `backend-ci.yml`/`frontend-ci.yml`/`infra-plan.yml` (see `patterns-discovered.md`, "Required
  Status Checks Must Always Run").
- **Alternatives considered**: Regenerate on every request/build without committing anything —
  rejected, provides no reviewable diff and no way to implement FR-008's PR-blocking check.

## Decision: Auth gating for Swagger UI ahead of Sprint 5

- **Decision**: The Swagger UI route is mounted inside the same Chi route group as `/api/v1`
  domain endpoints (not a standalone unauthenticated group), so it automatically picks up
  whatever authentication middleware Sprint 5 adds to that group, with zero changes needed to
  the mount point itself when that middleware lands. Until then, a
  `// TODO(sprint-5): remove once JWT middleware is wired` comment documents the gap — no
  interim bespoke auth (e.g., a hardcoded API key) is invented, since production is currently
  documented as dormant (`docs/cloud-and-environments.md`) and no other endpoint has auth yet
  either.
- **Rationale**: Building a one-off auth mechanism just for Swagger UI would violate Simplicity
  (Principle II) and pre-empt Sprint 5's actual JWT design. Mounting inside the shared route
  group means the feature is auth-complete the moment Sprint 5 ships, with no follow-up PR
  required against this feature's own code.
- **Alternatives considered**: An `ENABLE_SWAGGER_UI` environment flag defaulting to `false` in
  production — rejected: FR-010 explicitly requires the UI to be reachable (not disabled) in
  every environment; an env-gated disable would contradict that clarification.

### Implementation Status Note (2026-08-01, issue #192 / PR #204)

The mount point described above is unchanged — Swagger UI still lives inside the same route
registration path as `/api/v1`. However, the *gating mechanism* deviated from this decision once
implemented: `Authenticate` is now applied to `/swagger/*` only when `Config.IsProduction()` is
true (new `Config.Environment`/`Config.IsProduction()`, commit `6215738`), not uniformly across
every environment as decided above. This was a deliberate revisit, not an oversight: the
"reachable in every environment" property this section optimized for was actually motivated by
FR-010's concern that Swagger must never be silently disabled in production — that property is
still preserved (production remains gated, dev/staging are reachable). What changed is that local
and staging onboarding friction from requiring a JWT to open Swagger UI was judged to outweigh
keeping the gating mechanism uniform, so dev/staging were carved out while production stays
gated. The previously-rejected "environment flag" alternative above is conceptually close to what
shipped, except the switch is derived from `Config.IsProduction()` rather than a standalone
`ENABLE_SWAGGER_UI` flag, and it disables the *auth requirement* in non-production rather than
disabling the UI itself — so FR-010's actual requirement (never disabled in production) is not
contradicted.

## Open items carried to implementation

- None — all NEEDS CLARIFICATION items from Technical Context are resolved above.
