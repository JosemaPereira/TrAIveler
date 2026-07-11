# Feature Specification: API Documentation via OpenAPI/Swagger

**Feature Branch**: `009-api-documentation`

**Created**: 2026-07-10

**Status**: Draft

**Input**: User description: "API documentation via OpenAPI/Swagger: the project currently documents its API entirely through hand-written Markdown (each spec's contracts/api.md plus docs/api-design-standards.md for global conventions). There is no interactive API explorer and no machine-readable contract to validate requests/responses against. This spec should define generating and serving a machine-readable OpenAPI/Swagger specification for the backend API, along with an interactive documentation UI (e.g. Swagger UI or Redoc), so contributors and API consumers can browse and validate the real API surface instead of relying on manual Markdown that can drift from the implementation. Consider whether the spec is generated from Go doc-comment annotations (e.g. swaggo/swag) versus hand-maintained as an openapi.yaml, and decide during clarification. This complements (does not replace) docs/api-design-standards.md, which stays the source of truth for conventions."

## Clarifications

### Session 2026-07-10

- Q: Should the OpenAPI contract be generated from Go doc-comment annotations, hand-maintained as an `openapi.yaml`, or should the `openapi.yaml` be hand-written first and Go types/stubs generated from it? → A: `swaggo/swag` — Go doc-comment annotations colocated with handler code, regenerated on every build/CI run (code-first).
- Q: How should the interactive documentation UI be exposed across environments (dev/staging/production)? → A: Available in every environment, including production; in production it requires the same authentication (bearer token) already required by the API itself — no anonymous access, and no environment-based disabling.
- Q: Which interactive documentation UI should render the OpenAPI document? → A: Swagger UI — supports executing real "try it out" requests (required by FR-004/FR-005) and integrates directly with `swaggo/swag`-generated output.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Backend Developer Publishes an Always-Current Contract (Priority: P1)

As a backend developer adding or changing an endpoint, I need the machine-readable API contract to update automatically alongside my code change, so that the published documentation never drifts from what the API actually does.

**Why this priority**: This is the core problem being solved — hand-written Markdown drifts from the real API. If the contract isn't kept in lockstep with the code, the feature provides no more trust than the status quo.

**Independent Test**: Can be fully tested by adding a new endpoint (or changing an existing request/response shape) and confirming the published contract reflects the change the next time the API is built/deployed, with no manual documentation-editing step required.

**Acceptance Scenarios**:

1. **Given** a backend developer adds a new endpoint with its request/response types, **When** the project is built, **Then** the machine-readable contract includes that endpoint's path, method, parameters, request body, and response schema without any hand-written contract edit.
2. **Given** a backend developer changes an existing endpoint's response shape, **When** the project is built, **Then** the published contract reflects the new shape and a reviewer can see the change in the same code review as the implementation.
3. **Given** the machine-readable contract is regenerated, **When** it is compared against `docs/api-design-standards.md` conventions (error format, pagination, versioning), **Then** it is consistent with those documented conventions.

---

### User Story 2 - API Consumer Explores and Tries the API Interactively (Priority: P1)

As a frontend developer or external API consumer, I need an interactive documentation UI where I can browse every endpoint and try real requests against a running environment, so that I can understand and validate the API surface without reading Markdown or writing throwaway test scripts.

**Why this priority**: Interactive exploration is the main consumer-facing value of this feature — it is what turns a static contract into a tool people actually use day to day.

**Independent Test**: Can be fully tested by opening the documentation UI, selecting an endpoint, and executing a real request against a running backend instance, then confirming the displayed response matches what the API actually returned.

**Acceptance Scenarios**:

1. **Given** the backend is running, **When** a developer opens the documentation UI, **Then** they see every currently-implemented endpoint grouped by resource, each with its parameters, request body schema, and possible response codes.
2. **Given** an endpoint requires authentication, **When** a developer tries it from the documentation UI, **Then** the UI supports supplying the required credentials (e.g., a bearer token) and the request is sent with them.
3. **Given** a developer executes a request from the documentation UI, **When** the backend responds, **Then** the UI displays the actual status code, headers, and body returned — not a mocked example.

---

### User Story 3 - Technical Lead Gates Contract Drift in Review (Priority: P2)

As a technical lead reviewing a pull request, I need an automated check that fails when generated documentation is stale relative to the code, so that a contract/code mismatch cannot merge unnoticed.

**Why this priority**: Without an enforcement gate, the contract can still silently drift the same way the Markdown docs did — this priority protects the investment made in User Story 1.

**Independent Test**: Can be tested by introducing a code change that alters an endpoint's shape without regenerating the contract, opening a pull request, and confirming the relevant CI check fails.

**Acceptance Scenarios**:

1. **Given** a pull request changes an annotated endpoint's signature but does not regenerate the contract artifact, **When** CI runs, **Then** the documentation-drift check fails with a message indicating the contract is out of date.
2. **Given** a pull request regenerates the contract artifact consistently with its code changes, **When** CI runs, **Then** the documentation-drift check passes.

---

### Edge Cases

- What happens when an endpoint has no annotations/documentation yet (e.g., mid-development)? The contract generation MUST still succeed for the rest of the API, excluding or flagging the undocumented endpoint rather than failing the whole build.
- How does the system handle endpoints that are intentionally internal/undocumented (e.g., health checks used only by infrastructure)? These MUST be excludable from the published contract without being excluded from the API itself.
- What happens when the documentation UI is accessed in production? It MUST be reachable (not disabled), but MUST require the same bearer-token authentication as the API itself — no anonymous access, and no bypass of security to make endpoints easier to try.
- How does the system behave if contract generation fails during a build (e.g., malformed annotation)? The build MUST fail with a clear, actionable error rather than publishing a partial or invalid contract.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST produce a machine-readable OpenAPI (v3) document describing every implemented backend endpoint's path, HTTP method, parameters, request body schema, response schemas, and status codes.
- **FR-002**: The OpenAPI document MUST be regenerated from Go doc-comment annotations colocated with each handler (`swaggo/swag`-style) as part of the normal backend build/CI process, without a separate manual authoring step for endpoints that already carry these annotations.
- **FR-003**: System MUST serve Swagger UI, rendering the generated OpenAPI document and letting a user browse all endpoints grouped by resource.
- **FR-004**: The interactive documentation UI MUST allow executing real HTTP requests against a running backend instance and display the actual response received.
- **FR-005**: The interactive documentation UI MUST support supplying authentication credentials (e.g., a bearer token) so authenticated endpoints can be exercised.
- **FR-006**: System MUST allow specific endpoints (e.g., infrastructure-only health checks) to be excluded from the published OpenAPI document.
- **FR-007**: System MUST fail the build/CI with a clear error when contract generation encounters malformed or missing required documentation input, rather than publishing an invalid or partial contract.
- **FR-008**: CI MUST include an automated check that fails a pull request when the generated OpenAPI document would differ from the one checked into (or produced from) the current code, preventing undetected contract drift.
- **FR-009**: The OpenAPI document's error response schema, pagination parameters, and versioning (`/api/v1`) MUST be consistent with the conventions already documented in `docs/api-design-standards.md`.
- **FR-010**: The interactive documentation UI MUST be reachable in every environment (development, staging, production); in each environment it MUST require the same authentication already required by the API itself (e.g., a valid bearer token), with no anonymous access in production and no environment-based disabling of the UI.

### Key Entities

- **OpenAPI Document**: The machine-readable contract artifact (OpenAPI v3 schema) describing all published endpoints, their inputs, outputs, and status codes; generated from the backend's real route/handler/type definitions.
- **Documentation UI**: Swagger UI, the interactively-servable rendering of the OpenAPI Document that lets a human browse endpoints and execute real requests.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of currently-implemented, non-excluded backend endpoints appear in the generated OpenAPI document with accurate request/response schemas.
- **SC-002**: A developer can go from "I need to understand endpoint X" to "I have executed a real request against endpoint X and seen its actual response" in under 2 minutes using only the documentation UI.
- **SC-003**: A code change that alters an endpoint's request/response shape without updating the generated contract is caught by CI before merge, 100% of the time.
- **SC-004**: Zero manual Markdown edits are required to keep the machine-readable contract in sync with the API for endpoints that already carry code-level documentation annotations.

## Assumptions

- The backend (Go/Chi) is the system being documented; frontend-only or infrastructure-only surfaces are out of scope for this spec.
- Generation uses Go doc-comment annotations colocated with handler code (`swaggo/swag`-style), regenerated on every build/CI run — not a hand-maintained `openapi.yaml` and not a contract-first (`oapi-codegen`) approach (see Clarifications).
- The documentation UI is intended primarily for internal contributors and trusted API consumers (per the project's partner/admin roles), not as a public marketing-facing API portal.
- `docs/api-design-standards.md` remains the human-readable source of truth for conventions; this feature produces a derived, machine-readable, always-current artifact and an explorer for it — it does not replace or restate that document.
