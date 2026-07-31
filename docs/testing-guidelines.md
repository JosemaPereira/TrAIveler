# Testing Guidelines

TrAIveler follows a **Test-Driven Development** workflow: write a failing test (Red), make it pass with minimal code (Green), then refactor. All tests must pass before a pull request can be merged.

---

## Testing Strategy

The project uses a three-layer testing strategy:

```mermaid
%%{init: {'flowchart': {'curve': 'linear'}}}%%
graph TD
    subgraph "Layer 1: Unit Tests"
        U1[Individual functions<br/>Services<br/>Components in isolation]
        U2[Go testing + testify<br/>Vitest + React Testing Library]
        U1 --> U2
    end

    subgraph "Layer 2: Integration Tests"
        I1[Interaction between layers<br/>HTTP handlers ↔ services<br/>Components ↔ API]
        I2[Go net/http/httptest<br/>Vitest + MSW]
        I1 --> I2
    end

    subgraph "Layer 3: E2E Tests"
        E1[Full user flows<br/>through the browser]
        E2[Playwright]
        E1 --> E2
    end

    U2 --> I1
    I2 --> E1

    style U1 fill:#e8f5e9
    style I1 fill:#fff4e6
    style E1 fill:#e1f5ff
```

| Layer | Scope | Tools |
|---|---|---|
| Unit | Individual functions, services, and components in isolation | Go `testing` + `testify` / Vitest + React Testing Library |
| Integration | Interaction between layers (HTTP handlers ↔ services, components ↔ API) | Go `net/http/httptest` / Vitest + MSW |
| E2E | Full user flows through the browser | Playwright |

---

## Layer 1: Unit Tests

### Backend (Go)

- Test files live next to the source file they test: `itinerary_service_test.go` alongside `itinerary_service.go`.
- Use the standard `testing` package for all tests. Use `github.com/stretchr/testify/assert` and `testify/require` for assertions.
- Use interfaces and dependency injection to keep units testable without real databases or external services.
- Test function naming: `TestUnitName_Condition_ExpectedResult`.

```go
// Good
func TestGenerateItinerary_ValidInput_ReturnsItinerary(t *testing.T) { ... }
func TestGenerateItinerary_EmptyDestination_ReturnsError(t *testing.T) { ... }
```

#### Mocking Dependencies with Mockery

TrAIveler uses **[vektra/mockery](https://github.com/vektra/mockery)** to automatically generate type-safe mocks for interfaces.

**Why Mockery:**
- ✅ Type-safe: Compiler catches interface changes immediately
- ✅ Consistent: All mocks follow the same pattern
- ✅ Fluent API: Readable and expressive test setup
- ✅ Auto-cleanup: Automatic assertion verification
- ✅ Low maintenance: Regenerate when interfaces change

**Mock Location Standard:**
- **All mocks must be in `/mocks` subdirectory**
- Pattern: `<package>/mocks/<interface>_mock.go`
- Example: `internal/database/client.go` → `internal/database/mocks/client_mock.go`

**Generating Mocks:**
```bash
# From backend/ directory
make mocks

# Or directly
mockery --config .mockery.yaml --all
```

**Using Generated Mocks:**
```go
import (
    "testing"
    "github.com/stretchr/testify/mock"
    "github.com/yourproject/internal/database"
    dbmocks "github.com/yourproject/internal/database/mocks"
)

func TestService_ProcessData_Success(t *testing.T) {
    // Create mock from mocks subdirectory
    mockDB := dbmocks.NewMockClient(t)
    
    // Setup expectations using fluent API
    mockDB.EXPECT().Ping(mock.Anything).Return(nil).Once()
    mockDB.EXPECT().Close().Return(nil).Once()
    
    // Use mock in service
    service := NewDataService(mockDB)
    err := service.ProcessData(context.Background())
    
    // Assert results
    assert.NoError(t, err)
    // mockDB.AssertExpectations(t) called automatically on cleanup
}

// Test with specific argument matching
func TestService_ProcessData_WithSpecificContext(t *testing.T) {
    mockDB := dbmocks.NewMockClient(t)
    
    // Match exact context value
    ctx := context.WithValue(context.Background(), "request_id", "123")
    mockDB.On("Ping", ctx).Return(nil).Once()
    
    service := NewDataService(mockDB)
    err := service.ProcessData(ctx)
    
    assert.NoError(t, err)
}

// Test with custom behavior
func TestService_ProcessData_WithRunFunction(t *testing.T) {
    mockDB := dbmocks.NewMockClient(t)
    
    callCount := 0
    mockDB.EXPECT().Ping(mock.Anything).Run(func(ctx context.Context) {
        callCount++
        // Custom logic during method call
    }).Return(nil).Times(3)
    
    service := NewDataService(mockDB)
    // ... test logic ...
    
    assert.Equal(t, 3, callCount)
}
```

**Mock Expectations API:**
- `.Return(value)` — specify return values
- `.Once()`, `.Twice()`, `.Times(n)` — expected call count
- `.Run(func(...) { })` — execute custom logic when called
- `.Maybe()` — optional call (won't fail if not called)
- `mock.Anything` — match any argument value
- Specific values — match exact argument

**Best Practices:**
- Always use `NewMockClient(t)` to get automatic assertion checking
- Prefer `.EXPECT()` fluent API over `.On()` for better type safety
- Use `mock.Anything` for arguments you don't care about
- Use specific values when testing argument passing matters
- Regenerate mocks after interface changes: `make mocks`
- Commit generated mocks to git for consistency

**See Also:**
- [Mock Standards](mock-standards.md) — Comprehensive mock generation reference
- `backend/internal/database/client_mock_example_test.go` — Complete mockery usage examples
- `backend/internal/database/mocks/client_mock.go` — Generated mock example
- [Coding Guidelines](coding-guidelines.md) — Mock Generation section

#### Repository Integration Tests with Optional Testcontainers

A repository's DB-backed methods (`Create`/`FindByID`/`FindByEmail`/`Update`/`Delete`/`List` in
`backend/internal/example/repository.go`) aren't exercised by the service-layer mock-based tests
above — they need a real PostgreSQL connection. Rather than requiring Docker for every local test
run, these tests are gated behind Go's `testing.Short()`:

- **`make test`** (default, CI's fast lane) — `go test -tags=test ./... -short`, skips
  testcontainer-gated tests entirely. No Docker/Colima required.
- **`make test-all`** — `go test -tags=test ./...` (no `-short`), runs the full suite including
  testcontainers. Requires [Colima](https://github.com/abiosoft/colima) (`colima start --cpu 2
  --memory 4`) or another Docker-compatible runtime running locally.
- **`make test-coverage`** — the actual CI coverage target; same `-short` flag as `make test`, so
  it does **not** exercise `repository.go`'s DB-backed methods. This is expected, not a bug: see
  `specs/005-system-architecture/validation-results.md` (005-T128) for the concrete before/after
  coverage numbers this produces (64.1% short-mode vs. 89.1% full-suite for
  `internal/example`'s combined service+repository code).

`backend/internal/example/repository_integration_test.go` is the reference implementation: it
starts a real `postgres:16-alpine` container via `testcontainers-go`, applies the goose migrations
from `backend/migrations/`, and exercises the repository against it — see
`setupRepositoryTestDB(t, ctx)` in that file, which mirrors `backend/internal/database/client_test.go`'s
container-setup helper. See also `backend/TESTING.md` for the full command reference.

### Frontend (React / TypeScript)

- Test files live next to the source file they test: `TripCard.test.tsx` alongside `TripCard.tsx`.
- Use **Vitest** as the test runner and **React Testing Library** for component tests.
- Unit test naming follows: `describe` block names the unit, `it`/`test` block describes the behavior.

```ts
describe('TripCard', () => {
  it('renders the trip destination name', () => { ... });
  it('shows a placeholder when no image is provided', () => { ... });
});
```

- Test what the user sees and does, not implementation details. Query by accessible role, label, or text — not by CSS class or internal state.

---

## Layer 2: Integration Tests

### Backend (Go)

- Integration tests live in an `integration/` subdirectory within the relevant package, or in a top-level `tests/integration/` folder.
- Use `net/http/httptest` to spin up a real HTTP server and test handler-to-service interactions.
- Use a test database or an in-memory substitute (e.g., SQLite) — never run integration tests against a production or shared database.
- File naming: `<feature>_integration_test.go`.

```go
func TestCreateTrip_ValidPayload_Returns201(t *testing.T) { ... }
```

### Frontend (React / TypeScript)

- Integration tests verify that a feature's components, hooks, and API calls work together correctly.
- Use **Mock Service Worker (MSW)** to intercept HTTP requests instead of mocking modules directly.
- These tests live under `src/features/<feature>/__tests__/` or alongside the feature entry component.

**Current state (as of Sprint 6, issue #180)**: MSW is **wired up**. The shared server lives in
`frontend/src/test/msw/server.ts` with default handlers in `handlers.ts`, and its
`listen`/`resetHandlers`/`close` lifecycle is owned by `frontend/src/test/setup.ts`, so every test
file gets it for free — import `server` only to add per-test overrides via `server.use(...)`. The
server runs with `onUnhandledRequest: 'error'`: a request no handler accounts for is a hard failure,
never a silent real network call. Handler URLs are built from the same `VITE_API_BASE_URL` the
client reads, so the two cannot drift. `frontend/src/test/queryWrapper.tsx` provides the
`QueryClientProvider` + `MemoryRouter` wrapper such tests need. First consumers:
`src/features/auth/`.

**MSW and a stubbed global `fetch` cannot share a test file.** `vi.stubGlobal('fetch', …)` replaces
the very function MSW's Node interceptor patches, so a handler in the same file is silently never
consulted — the stub answers and the assertion quietly tests nothing. `src/lib/api-client.test.ts`
keeps its direct-`fetch` stubs because it is the unit test for the client itself, **not** a
precedent for feature-level tests; anything exercising a real request *sequence* (e.g. the
401 → refresh → retry flow) goes in a separate MSW-backed file such as `api-client.refresh.test.ts`.

---

## Layer 3: E2E Tests

- E2E tests are written with **Playwright** and live in the top-level `e2e/` directory.
- Each file covers one user flow: `generate-itinerary.spec.ts`, `share-trip.spec.ts`.
- E2E tests run against a fully deployed (or locally running) instance of the application.
- Do not use hard-coded timeouts. Use Playwright's built-in auto-wait and `expect` assertions.
- Accessibility scanning (`@axe-core/playwright`, WCAG 2.1 AA) is a dedicated CI gate
  (the `Accessibility Audit` job in `.github/workflows/frontend-ci.yml`, issue #93) rather than a
fourth testing layer — it reuses
  Playwright specs tagged `@accessibility` in `e2e/` plus a Lighthouse CI audit. See
  [`e2e/README.md`](../e2e/README.md#accessibility-testing) for current wiring status and
  `docs/nfrs.md` (NFR-A11Y) for the validated thresholds.

```ts
test('user can generate a trip itinerary', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('Destination').fill('Tokyo, Japan');
  await page.getByRole('button', { name: 'Generate Itinerary' }).click();
  await expect(page.getByRole('heading', { name: 'Day 1' })).toBeVisible();
});
```

---

## Folder Structure Summary

```
backend/
  internal/
    itinerary/
      itinerary_service.go
      itinerary_service_test.go       # unit
    handler/
      trip_handler.go
      trip_handler_test.go            # unit
  tests/
    integration/
      trip_integration_test.go        # integration

frontend/
  src/
    components/
      TripCard/
        TripCard.tsx
        TripCard.test.tsx             # unit
    features/
      itinerary/
        ItineraryView.tsx
        __tests__/
          ItineraryView.test.tsx      # integration

e2e/
  generate-itinerary.spec.ts          # E2E
  share-trip.spec.ts                  # E2E
```

---

## Coverage Targets

| Layer | Minimum Coverage |
|---|---|
| Unit (backend) | 90% of business logic |
| Unit (frontend) | 90% of shared components and hooks |
| Integration | All public API endpoints |
| E2E | All primary user flows defined in functional requirements |

Coverage is a floor, not a goal. Prioritize meaningful tests over achieving a percentage.

**Observed numbers (Spec 005 validation sweep, 2026-07-13 — see
`specs/005-system-architecture/validation-results.md` for full detail)**: backend
`internal/example` (service + repository combined) measured **89.1%** with the full,
Colima-backed test run — marginally under the 90% floor adopted on 2026-07-16 (the floor was 80%
at the time of that sweep), so the next backend change in that package should close the ~1-point
gap. The CI/`make test-coverage` lane alone (short mode, no testcontainers) reads a lower
**64.1%** by design, since `repository.go`'s DB-backed methods are only exercised by the
testcontainer-gated test (see "Repository Integration Tests with Optional Testcontainers" above)
— not a real shortfall. Frontend coverage **is enforced at 90%** (statements, branches, functions,
lines) via `thresholds` in `vitest.config.ts`, so `npm run test:coverage` and the CI coverage step
fail below that floor (this supersedes the enforcement gap formerly tracked as roadmap task
`002-T041`); 136/136 frontend tests pass across 15 files as of the same sweep.

---

<!-- PROMOTED:validation-gates START -->
<!-- Generated from specs/009-api-documentation/spec.md, plan.md, research.md -->
<!-- Last promoted: 2026-07-10 -->

## Validation Gates Beyond the Three Testing Layers

Not every required CI gate is a test in the unit/integration/E2E sense above. These gates validate
that a generated artifact matches the code it was generated from, and block merge on drift the same
way a failing test would — but they are not substitutes for the testing layers, and the testing
layers are not substitutes for them.

### `swagger-drift` (backend CI)

- **What it checks**: That the committed Swagger 2.0 (OpenAPI 2.0) contract (`backend/docs/`, generated by
  `swag init` from Go doc-comment annotations) is identical to what regenerating it right now would
  produce.
- **How**: `.github/workflows/backend-ci.yml` runs `make swagger` (wraps `swag init -g
  cmd/api/docs.go -o docs`) and then `git diff --exit-code -- backend/docs`, failing the job if any
  difference is reported.
- **Why it is not a "test"**: It does not exercise application behavior; it only proves the
  machine-readable contract was regenerated and committed alongside the code change that produced
  it (see `docs/api-design-standards.md` §16 and `specs/009-api-documentation/research.md`).
- **Scope**: Gated the same way as the rest of the required backend checks (`dorny/paths-filter`,
  only runs when `backend/**` changes).

<!-- PROMOTED:validation-gates END -->

---

## Rules

- Tests must be deterministic. Avoid relying on real clocks, random values, or external services in unit and integration tests.
- A failing test must never be committed unless it is intentionally in the Red phase of TDD and tracked in the current branch.
- Test names must be written in English and must describe the expected behavior, not the implementation.
