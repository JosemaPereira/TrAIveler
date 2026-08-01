# Coding Guidelines

These guidelines apply to all source code in TrAIveler. They cover both the Go backend and the React frontend. All code, identifiers, comments, and documentation must be written in English.

---

## General Principles

### KISS — Keep It Simple, Stupid
- Prefer the simplest solution that correctly solves the problem.
- Avoid speculative abstractions: do not build for requirements that do not exist yet.
- A function or component that is hard to explain is a signal that it needs to be simplified.

### DRY — Don't Repeat Yourself
- Extract shared logic into a reusable function, hook, or package only when the same logic appears in **three or more** places.
- Do not prematurely abstract two similar-looking pieces of code — wait until the pattern is clear.
- Shared constants (strings, numbers, config keys) must be declared once and referenced everywhere else.

---

## Go — Backend

### Formatting
- All Go code must be formatted with `gofmt` before committing. CI will reject unformatted code.
- Maximum line length is **120 characters**. Prefer shorter lines when readability allows.

### Import Organization
Imports must be grouped in this exact order, separated by blank lines:

```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "net/http"

    // 2. Third-party packages
    "github.com/some/dependency"

    // 3. Internal packages
    "github.com/traivelr/internal/itinerary"
)
```

Use `goimports` (or an equivalent tool) to keep imports sorted and remove unused imports automatically.

**Real example** (`backend/internal/example/handler.go`), showing all three groups:

```go
import (
    "encoding/json"
    "fmt"
    "log/slog"
    "net/http"
    "strconv"
    "strings"

    "github.com/go-chi/chi/v5"

    domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
    "github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
)
```

**Enforcement note**: `goimports` is enabled as a `golangci-lint` linter (`backend/.golangci.yml`),
but that config does **not** set goimports' `local-prefixes` option — so the linter enforces that
each import block is gofmt/goimports-formatted (correctly sorted within its group, no unused
imports), not the specific three-way stdlib/external/internal split itself. The split is a project
convention, preserved by keeping the blank lines between groups when you add or remove an import —
`goimports` won't merge or re-split existing groups on its own without that flag configured. Follow
the convention by hand; don't rely on the linter to catch a missing blank line here.

### Package Naming
- Package names must be lowercase, single words, with no underscores or camelCase: `itinerary`, `handler`, `store`.
- File names use snake_case: `itinerary_service.go`, `trip_handler.go`. Matches the real codebase,
  e.g. `backend/internal/middleware/request_id.go`, `backend/internal/middleware/body_size.go`,
  `backend/internal/ai/ollama_client.go`.

### Struct Field Ordering
- **Prioritize logical grouping and readability over memory alignment optimization**.
- Group related fields together to make the struct's purpose clear.
- Order fields by their importance to the struct's functionality, not by size.
- Place configuration fields logically (e.g., timeouts together, connection settings together).
- **Rationale**: The `fieldalignment` linter is disabled project-wide because:
  - Memory savings are typically negligible (few bytes per struct instance)
  - Modern hardware makes alignment differences minimal
  - Readability and maintainability matter more than micro-optimizations
  - Forced size-based ordering makes code harder to understand
  - Time spent on alignment could be better used elsewhere

**Example:**
```go
// ✅ Good - Logical grouping
type ServerConfig struct {
    // Server behavior
    Port         int
    AllowedCORS  string
    
    // Timeouts (grouped together)
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
}

// ❌ Avoid - Scattered by size without clear logic
type ServerConfig struct {
    AllowedCORS  string        // 16 bytes
    ReadTimeout  time.Duration // 8 bytes
    WriteTimeout time.Duration // 8 bytes
    IdleTimeout  time.Duration // 8 bytes
    Port         int           // 8 bytes
}
```

### Function Parameters
- **Context must always be the first parameter** when present (after receiver for methods).
- Testing parameter `*testing.T` comes before context in test helpers (Go convention).

**Example:**
```go
// ✅ Correct - context first
func ProcessData(ctx context.Context, userID string, data []byte) error

// ✅ Correct - test helper (t before ctx is acceptable)
func setupTestDB(t *testing.T, ctx context.Context) *DB

// ❌ Wrong - context not first
func ProcessData(userID string, ctx context.Context, data []byte) error
```

### Spelling
- Use **US English spelling** in all code, comments, and documentation.
- Common corrections: "canceled" (not "cancelled"), "color" (not "colour"), "optimize" (not "optimise").
- The `misspell` linter enforces US English spelling.

### Error Handling
- Never ignore errors. Always handle or propagate them explicitly.
- Wrap errors with context using `fmt.Errorf("context: %w", err)` to preserve the error chain.
- Return errors to callers; do not use `log.Fatal` or `os.Exit` outside of `main`.
- `_ = someCall()` does **not** satisfy this project's `errcheck` config (`check-blank: true` in
  `backend/.golangci.yml` flags blank-assigned errors too, not just unchecked ones). If an error
  genuinely cannot be usefully propagated (e.g. a best-effort write to an already-committed HTTP
  response), check it and log it — don't discard it with `_ =`.

**Example:**
```go
// ❌ Still flagged by errcheck (check-blank: true)
_ = json.NewEncoder(w).Encode(payload)

// ✅ Correct - checked and logged
if err := json.NewEncoder(w).Encode(payload); err != nil {
    slog.Default().Error("failed to write response", "error", err)
}
```

- **Log or Return — Never Both**: an error must either be handled (logged and mitigated/rendered)
  or propagated (wrapped and returned to the caller), but never both. If you wrap and return an
  error, do not also log it at that level — doing so produces duplicate noise in production logs,
  since the caller (or its caller) will log it again. Log errors only at the application boundary,
  where they stop propagating: the HTTP handler (via `errors.HandleError`), the background worker
  runner, or `main`.

**Example:**

```go
// ❌ Wrong - logs AND returns; the boundary will log it again
func (s *service) GetTrip(ctx context.Context, id string) (*Trip, error) {
    trip, err := s.repo.FindByID(ctx, id)
    if err != nil {
        slog.Default().Error("failed to find trip", "error", err)
        return nil, fmt.Errorf("get trip %s: %w", id, err)
    }
    return trip, nil
}

// ✅ Correct - wrap and return only; the HTTP handler logs it once
func (s *service) GetTrip(ctx context.Context, id string) (*Trip, error) {
    trip, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get trip %s: %w", id, err)
    }
    return trip, nil
}
```

### Unused Parameters

- `revive`'s `unused-parameter` rule is enabled repo-wide. In test doubles (e.g. an inline
  `http.HandlerFunc` standing in for the next handler in a middleware test), name only the
  parameters the closure body actually references and use `_` for the rest.

**Example:**
```go
// ❌ Flagged: neither w nor r is used
next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    nextCalled = true
})

// ✅ Correct
next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
    nextCalled = true
})
```

### Code Complexity
- **Keep cognitive complexity at 15 or below** to avoid triggering SonarQube warnings (rule `go:S3776`).
- Cognitive complexity measures how difficult code is to understand based on nested control flow, not just lines of code.
- If a function exceeds complexity 15, refactor by:
  - **Extracting helper functions**: Pull out nested logic into separate, well-named functions with `t.Helper()` for test helpers
  - **Early returns**: Use guard clauses to reduce nesting (return early on error conditions)
  - **Simplifying conditionals**: Replace complex if/else chains with switch statements or lookup tables
  - **Table-driven tests**: Use test tables instead of multiple similar test cases with duplicated assertion logic

**Example — Reducing Test Complexity:**
```go
// ❌ High complexity (17) - nested conditionals in loop
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        result, err := Process(tt.input)
        
        if tt.wantErr {
            if err == nil {
                t.Error("expected error, got nil")
            }
            if !strings.Contains(err.Error(), tt.errMsg) {
                t.Errorf("error = %v, want %s", err, tt.errMsg)
            }
        } else {
            if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
            if result != tt.want {
                t.Errorf("got %v, want %v", result, tt.want)
            }
        }
    })
}

// ✅ Low complexity - extracted helper
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        result, err := Process(tt.input)
        assertResult(t, result, err, tt.want, tt.wantErr, tt.errMsg)
    })
}

func assertResult(t *testing.T, result, err, want interface{}, wantErr bool, errMsg string) {
    t.Helper()
    if wantErr {
        if err == nil {
            t.Error("expected error, got nil")
            return
        }
        if !strings.Contains(err.Error(), errMsg) {
            t.Errorf("error = %v, want %s", err, errMsg)
        }
        return
    }
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
    if result != want {
        t.Errorf("got %v, want %v", result, want)
    }
}
```

**Benefits:**
- Code passes SonarQube quality gates without warnings
- Functions are easier to understand and maintain
- Helper functions can be reused across multiple tests
- Reduced cognitive load for code reviewers

### Dependency Injection (Interface-First Design)
- **Define interfaces before implementations** to enable easy mocking, testing, and component swapping.
- Interfaces should be small and focused (1-5 methods). Prefer multiple small interfaces over large ones.
- **Accept interfaces, return concrete types** in most cases. Exception: factories return interfaces when multiple implementations exist.
- Place the interface in the package that uses it (consumer), not where it's implemented (producer).
- Use descriptive interface names ending in `-er` when appropriate: `Reader`, `Writer`, `Closer`, `Client`.

**Example — Database Client:**
```go
// Define interface first
type Client interface {
    Ping(ctx context.Context) error
    Close() error
    Pool() *pgxpool.Pool
}

// Concrete implementation (private)
type pgxClient struct {
    pool *pgxpool.Pool
}

// Factory returns interface
func NewClient(ctx context.Context, url string) (Client, error) {
    return &pgxClient{pool: ...}, nil
}

// Usage in main.go
var dbClient database.Client
dbClient, err = database.NewClient(ctx, cfg.Database.URL)
```

**Benefits:**
- Easy to mock in tests: create `type mockClient struct{}` that implements `Client`
- Swap implementations without changing consumers (e.g., in-memory vs PostgreSQL)
- Clear contract: interface documents what operations are available
- Encourages loose coupling between components

### Mock Generation with Mockery
- Use **[vektra/mockery](https://github.com/vektra/mockery)** to automatically generate type-safe mocks for interfaces.
- **Standard location**: Mocks are generated in a `/mocks` subdirectory relative to the interface.
  - Example: `internal/database/client.go` → `internal/database/mocks/client_mock.go`
- All mock files are excluded from production builds using `//go:build test` build tag.

**Configuration:**
- Mockery configuration is defined in `backend/.mockery.yaml`
- Settings: `with-expecter: true` enables fluent API for setting expectations
- Settings: `dir: "{{.InterfaceDir}}/mocks"` generates mocks in /mocks subdirectory
- Settings: `filename: "{{.InterfaceName | snakecase}}_mock.go"` uses snake_case naming

**Generating Mocks:**
```bash
# From backend/ directory

# Generate all mocks defined in .mockery.yaml
make mocks

# Or use mockery directly
mockery --config .mockery.yaml --all
```

**Using Mocks in Tests:**
```go
import (
    "testing"
    "github.com/stretchr/testify/mock"
    "github.com/yourproject/internal/database"
    dbmocks "github.com/yourproject/internal/database/mocks"
)

func TestMyService_Success(t *testing.T) {
    // Create mock from mocks subdirectory
    mockDB := dbmocks.NewMockClient(t)
    
    // Setup expectation using fluent API
    mockDB.EXPECT().Ping(mock.Anything).Return(nil).Once()
    
    // Use mock in service
    service := NewMyService(mockDB)
    err := service.DoSomething(context.Background())
    
    assert.NoError(t, err)
    // mockDB.AssertExpectations(t) is automatically called on cleanup
}
```

**Mock Expectations API:**
- `.Return(value)` — specify return values
- `.Once()`, `.Twice()`, `.Times(n)` — number of calls expected
- `.Run(func(...) { })` — execute custom logic when method is called
- `.Maybe()` — call is optional (won't fail if not called)
- `mock.Anything` — match any argument value
- Specific values — match exact argument (e.g., `Ping(ctx)` expects that exact context)

**Best Practices:**
- Generate mocks for all interfaces in `internal/` packages
- **Always place mocks in `/mocks` subdirectory** (e.g., `internal/database/mocks/`)
- Import mocks with alias: `dbmocks "path/to/package/mocks"`
- Commit generated mock files to git for consistency across team
- Regenerate mocks after interface changes: `make mocks`
- Use `mock.Anything` for arguments you don't care about
- Use specific values when testing argument passing
- Prefer `.EXPECT()` fluent API over `.On()` for better type safety

**See Also:** [Mock Standards](mock-standards.md) for complete reference

### Linting
- The project uses `golangci-lint`. All lint checks must pass before a pull request can be merged.
- Key enabled linters: `errcheck`, `govet`, `staticcheck`, `revive`, `gosec`, `gofmt`, `goimports`, `misspell`, `unparam`, `unconvert`, `goconst`, `gocyclo`, `gosimple`, `ineffassign`, `unused`.
- **govet configuration**: `enable-all: true` with `shadow` and `fieldalignment` disabled.
  - `shadow` - Variable shadowing warnings are too noisy for practical use
  - `fieldalignment` - Struct field ordering micro-optimization sacrifices readability
- **Test file exemptions**: The following linters are disabled for `*_test.go` files:
  - `gocyclo` - Cyclomatic complexity (test helpers can be complex)
  - `errcheck` - Error checking (some test errors are intentionally ignored)
  - `gosec` - Security checks (tests don't need production security)
  - `goconst` - Constant detection (test data repetition is acceptable)

**Running lints locally:**
```bash
# From backend/ directory
make lint              # Run all configured linters
make fmt               # Format code with gofmt + go mod tidy
make vet               # Run go vet only
```

### Project Structure (Backend)

Actual current structure (see [`backend/README.md`](../backend/README.md#project-structure) for
the full, up-to-date tree with per-file descriptions):

```
backend/
  cmd/
    api/            # main entrypoint (main.go, server.go, routes.go) — not cmd/server/
  internal/
    middleware/      # RequestID, Logger, Recovery, CORS, BodySize, Authenticate, RateLimit
    database/         # pgxpool client (pool size from DB_MIN/MAX_CONNECTIONS)
      migrations/     # shared goose config (Dir, SetDialect) — the .sql files do NOT live here
    errors/           # DomainError + HandleError
    ai/               # AIClient interface + OllamaClient/AnthropicClient
    auth/             # flat helpers: password.go (bcrypt cost 12), validator.go
      jwt/            # Generator/Validator/Refresher, KeyProvider, Claims (RS256, multi-key rotation)
      ratelimit/      # progressive-delay login-attempt throttle (Limiter + in-memory TTL store)
    observability/    # GenerateCorrelationID, LogSecurityEvent (structured JSON)
    subscription/     # doc.go scaffold only (Spec 008, unbuilt)
    collaboration/    # doc.go scaffold only (Spec 008, unbuilt)
    security/         # doc.go scaffold only (Spec 008, unbuilt)
    example/          # throwaway model→repository→service→handler reference pattern —
                       # copy this layering for a real <domain>/ package (handler.go,
                       # service.go, repository.go, model.go), then delete example/
  pkg/                # deliberately empty (.gitkeep only) — dead by convention; new shared
                       # backend code goes under internal/, even when a spec/task literally
                       # names a pkg/... path
  config/
  migrations/         # goose migrations, flat and shared across specs (001-004 security
                       # tables + the throwaway timestamp-versioned examples migration)
```

`internal/example/` is the reference for what a real `internal/<domain>/` package should look
like — `handler.go`, `service.go`, `repository.go` (plus `model.go`) — until the first real domain
(Trip) ships and it is deleted.

**Subpackage vs. flat convention (the "why subfolders" rule)**: introduce a subpackage
(`internal/auth/jwt/`, `internal/auth/ratelimit/`) for a cohesive sub-domain that owns multiple
types/interfaces and a boundary of its own (jwt: `Generator`/`Validator`/`Refresher`,
`KeyProvider`, `Claims`; ratelimit: `Limiter` + its store); keep small, self-contained helpers
flat in the parent package (`password.go`, `validator.go`). The asymmetry is intentional:
`plan.md` originally envisioned an `internal/auth/password/` subpackage, but it was deliberately
kept flat because it is small — so a stale `internal/auth/password/` path in task text should
resolve to the flat helpers, not a new subpackage.

### Testing Conventions (Backend)

These are code-level conventions; the three-layer strategy (unit/integration/E2E), folder
structure, and coverage targets live in [`testing-guidelines.md`](testing-guidelines.md).

- **Table-Driven Tests (TDT)**: for complex or conditional logic (input validation, routing, state
  mapping, error scenarios), ALWAYS prefer table-driven tests using structs and descriptive subtest
  names via `t.Run`. Combine with an extracted assertion helper to keep cognitive complexity at 15
  or below (see the [Code Complexity](#code-complexity) example above).
- **Test Isolation & Hermeticity**: each database integration test must run in its own transaction
  or use an isolated database schema within the container runtime, so tests can execute in parallel
  safely and never depend on another test's leftover state.
- **Clean Test APIs & Lifecycle**: prefer `t.Cleanup(func() { ... })` over deferred calls (`defer`)
  inside test setup helpers when the lifecycle of a resource is bound to the helper itself —
  cleanup then runs when the *test* finishes, not when the helper returns.
- **Helper Parameters**: test helpers must accept `*testing.T` as the first parameter, before the
  context parameter (e.g., `setupTestDB(t *testing.T, ctx context.Context)`), consistent with the
  [Function Parameters](#function-parameters) rule above. Call `t.Helper()` at the top of every
  helper so failures are reported at the caller's line, keeping test tracebacks clear and cognitive
  complexity low.

---

## React — Frontend

### Formatting
- All code must be formatted with **Prettier** using the project's `.prettierrc` configuration.
- **ESLint** must report zero errors before committing. Warnings should be resolved, not suppressed.

### TypeScript
- The project uses **TypeScript**. Avoid `any`; prefer explicit types or `unknown` with type guards.
- Enable strict mode in `tsconfig.json`.

### Import Organization
Imports must be grouped in this order, separated by blank lines: React/framework → third-party →
internal absolute (`@/...`) → styles/assets.

Any import that crosses a directory boundary (i.e. would otherwise need a `../`) **must** use the
`@/` path alias (mapped to `./src` in `vite.config.ts` / `tsconfig.json`'s
`"paths": { "@/*": ["./src/*"] }`), regardless of depth — this includes single-level `../` imports,
not just deep `../../../` chains. Only same-directory imports (`./Foo.module.css`, a co-located
test/type file) stay relative.

```ts
// 1. React and framework
import { useState, useEffect } from 'react';

// 2. Third-party libraries
import { useQuery } from '@tanstack/react-query';

// 3. Internal — absolute paths (via the '@' path alias)
import { ItineraryCard } from '@/components/ItineraryCard';
import { useTrip } from '@/hooks/useTrip';
import { isAPIError } from '@/lib/query-client';

// 4. Styles / assets — same-directory relative imports are fine here
import styles from './TripView.module.css';
```

**Enforcement**: the `@/`-over-`../` rule is enforced automatically by ESLint's built-in
`no-restricted-imports` (`frontend/eslint.config.js`), which errors on any `../*` import —
`npm run lint` will catch it. Group *ordering* (React → third-party → internal → styles) remains a
documented convention only (also stated in
[`frontend/README.md`](../frontend/README.md#code-standards)); no `eslint-plugin-import`/`import/order`
plugin is installed, so ordering itself is not auto-checked and relies on code review.

### Component Conventions
- One component per file. The file name matches the component name in PascalCase: `TripCard.tsx`.
- Co-locate the component's styles, tests, and types in the same folder when they are only used by that component.
- Prefer function components with hooks. Do not use class components.

### Hooks
- Custom hooks live in `src/hooks/`. Hook files and function names start with `use`: `useItinerary.ts`.
- A hook must have a single, clearly named responsibility.

### Project Structure (Frontend)

Actual current structure (see [`frontend/README.md`](../frontend/README.md#project-structure) for
the full, up-to-date tree with per-file descriptions):

```
frontend/
  src/
    components/
      primitives/     # Button, Input, Card, LoadingSpinner, ErrorMessage, EmptyState
      composites/     # Form (composes Button + Input)
      ErrorBoundary.tsx  # top-level, sits outside the primitives/composites/features layers
    routes/           # React Router v7 config (index.tsx, RootLayout, HomePage)
    stores/           # Zustand stores (auth-store.ts)
    features/         # feature-scoped components and logic (empty placeholder today)
    hooks/            # shared custom hooks, e.g. useErrorHandler.ts
    lib/               # framework/infra wiring: api-client.ts, query-client.ts
    styles/            # tokens.css, global.css
    test/              # Vitest setup helpers
```

Note: there is no `pages/`, `services/`, `types/`, or `utils/` directory in the real codebase —
route-level components live in `routes/`, API client functions in `lib/`, and shared types are
currently colocated with the module that defines them (e.g. `APIError` in `lib/api-client.ts`).
Introduce `types/`/`utils/` only once a genuinely cross-cutting type or pure helper needs one,
per this file's DRY principle above.

### File Naming Conventions

| Kind | Convention | Example |
|------|------------|---------|
| Component | `PascalCase.tsx`, matching the exported component name | `Button.tsx`, `ErrorBoundary.tsx` |
| Component test | `PascalCase.test.tsx`, co-located | `Button.test.tsx` |
| Component styles | `PascalCase.module.css`, co-located | `Button.module.css` |
| Hook | `camelCase.ts`, `use`-prefixed | `useErrorHandler.ts` |
| Library / infra module (`lib/`, `stores/`) | `kebab-case.ts` | `api-client.ts`, `query-client.ts`, `auth-store.ts` |
| Route component | `PascalCase.tsx` under `routes/` | `HomePage.tsx`, `RootLayout.tsx` |

### Testing Conventions & UI Selectors (Frontend)

These are code-level conventions; the three-layer strategy, folder structure, and coverage targets
live in [`testing-guidelines.md`](testing-guidelines.md).

- **Behavior-First Testing**: test components based on user interactions and accessibility markers
  rather than internal state, props, or mock-heavy React internals. If a test breaks when the
  implementation is refactored but the behavior is unchanged, it was testing the wrong thing.
- **AAA Pattern**: structure each test internally in three distinct phases separated by a single
  blank line: **Arrange** (render and mock setup), **Act** (user interactions via `userEvent`),
  and **Assert** (expectations).
- **Hierarchical Contexts**: use nested `describe` blocks to represent behavior and context
  cleanly (`describe('when [context]', ...)` or `describe('having [precondition]', ...)`),
  followed by leaf-level `it('should [outcome]', ...)` assertions.
- **A11y-First & E2E Preparation**: always write UI code that lets tests query elements via
  standard ARIA roles (e.g., `<button>` instead of a styled generic `<div>` with an `onClick`) —
  the same queries our upcoming Playwright E2E suite will rely on. If a unique element cannot be
  safely queried by its role or accessible name, add a dedicated data attribute:
  `data-testid="element-name"`. Never use styling classes (`.active-card`) or brittle DOM paths
  for test selection.

**Example:**

```tsx
describe('when the trip form is submitted', () => {
  it('should disable the submit button while the request is in flight', async () => {
    const user = userEvent.setup();
    render(<TripForm onSubmit={slowSubmit} />);

    await user.click(screen.getByRole('button', { name: /create trip/i }));

    expect(screen.getByRole('button', { name: /create trip/i })).toBeDisabled();
  });
});
```

---

## Shared Rules

- All variable, function, type, and file names must be in English.
- Secrets and environment-specific values must be read from environment variables. Never hard-code them.
- Remove dead code and commented-out code before merging a pull request.
