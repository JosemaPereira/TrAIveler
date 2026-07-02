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

### Error Handling
- Never ignore errors. Always handle or propagate them explicitly.
- Wrap errors with context using `fmt.Errorf("context: %w", err)` to preserve the error chain.
- Return errors to callers; do not use `log.Fatal` or `os.Exit` outside of `main`.

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
