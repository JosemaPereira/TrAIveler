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

### Package Naming
- Package names must be lowercase, single words, with no underscores or camelCase: `itinerary`, `handler`, `store`.
- File names use snake_case: `itinerary_service.go`, `trip_handler.go`.

### Struct Field Ordering
- **Order struct fields by size (largest to smallest) to optimize memory alignment**.
- Group fields logically when alignment permits, but prioritize alignment to avoid padding waste.
- Place strings and pointers before smaller types (int64, int32, int, bool).
- Use `fieldalignment` linter to detect suboptimal struct layouts.

**Example:**
```go
// ❌ Poor alignment (40 bytes with padding)
type ServerConfig struct {
    Port         int           // 8 bytes
    ReadTimeout  time.Duration // 8 bytes  
    WriteTimeout time.Duration // 8 bytes
    IdleTimeout  time.Duration // 8 bytes
    AllowedCORS  string        // 16 bytes (pointer + len)
}

// ✅ Optimal alignment (8 bytes with proper ordering)
type ServerConfig struct {
    AllowedCORS  string        // 16 bytes (string header)
    ReadTimeout  time.Duration // 8 bytes
    WriteTimeout time.Duration // 8 bytes
    IdleTimeout  time.Duration // 8 bytes
    Port         int           // 8 bytes
}
```

**Rule**: Strings first, then time.Duration, then int, then bool. This minimizes padding.

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
- Key enabled linters: `errcheck`, `govet`, `staticcheck`, `revive`, `gosec`.

### Project Structure (Backend)
```
backend/
  cmd/
    server/         # main entrypoint
  internal/
    <domain>/       # one package per domain concept (itinerary, trip, user, etc.)
      handler.go
      service.go
      repository.go
  pkg/              # packages safe to import from outside internal/
  config/
```

---

## React — Frontend

### Formatting
- All code must be formatted with **Prettier** using the project's `.prettierrc` configuration.
- **ESLint** must report zero errors before committing. Warnings should be resolved, not suppressed.

### TypeScript
- The project uses **TypeScript**. Avoid `any`; prefer explicit types or `unknown` with type guards.
- Enable strict mode in `tsconfig.json`.

### Import Organization
Imports must be grouped in this order, separated by blank lines:

```ts
// 1. React and framework
import { useState, useEffect } from 'react';

// 2. Third-party libraries
import { useQuery } from '@tanstack/react-query';

// 3. Internal — absolute paths (via path aliases)
import { ItineraryCard } from '@/components/ItineraryCard';
import { useTrip } from '@/hooks/useTrip';

// 4. Styles / assets
import styles from './TripView.module.css';
```

### Component Conventions
- One component per file. The file name matches the component name in PascalCase: `TripCard.tsx`.
- Co-locate the component's styles, tests, and types in the same folder when they are only used by that component.
- Prefer function components with hooks. Do not use class components.

### Hooks
- Custom hooks live in `src/hooks/`. Hook files and function names start with `use`: `useItinerary.ts`.
- A hook must have a single, clearly named responsibility.

### Project Structure (Frontend)
```
frontend/
  src/
    components/     # shared/reusable UI components
    features/       # feature-scoped components and logic
      itinerary/
      trip/
    hooks/          # shared custom hooks
    pages/          # top-level route components
    services/       # API client functions
    types/          # shared TypeScript types and interfaces
    utils/          # pure utility functions
```

---

## Shared Rules

- All variable, function, type, and file names must be in English.
- Secrets and environment-specific values must be read from environment variables. Never hard-code them.
- Remove dead code and commented-out code before merging a pull request.
