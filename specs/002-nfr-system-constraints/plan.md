# Implementation Plan: Non-Functional Requirements and System Constraints

**Branch**: `002-nfr-system-constraints` | **Date**: 2026-07-02 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/002-nfr-system-constraints/spec.md`

## Summary

Implement the cross-cutting quality infrastructure that enforces the 26 NFRs defined in the
specification. The primary deliverables are: a structured-logging middleware (slog) with
correlation IDs, a `/healthz` health-check endpoint, a server-side prompt-validation layer
(pattern-based deny-list), an AI output sanitisation utility (bluemonday), a privacy-policy
page, and the GitHub Actions CI pipeline stages that enforce coverage floors, lint gates,
secret scanning, security scanning, accessibility audits (axe-core + Lighthouse CI), and a
manual-trigger load-test gate (k6, 500 VUs) before any release to `main`.

See [research.md](research.md) for all tool and library decisions.

## Technical Context

**Language/Version**: Go 1.24 (backend), TypeScript 5.x + React 19 (frontend)

**Primary Dependencies**:
- **Backend**: `github.com/go-chi/chi/v5` (HTTP router), `log/slog` (structured logging, stdlib),
  `github.com/google/uuid` (correlation ID generation),
  `github.com/microcosm-cc/bluemonday` (AI output HTML sanitisation),
  `github.com/golangci/golangci-lint` (lint), `gosec` (SAST), `govulncheck` (dep audit)
- **Frontend**: `@axe-core/playwright` (accessibility scanning in E2E), `@lhci/cli` (Lighthouse
  CI for Core Web Vitals), `npm audit` (dependency vulnerability check), ESLint + Prettier
- **CI**: GitHub Actions — `ci.yml` (lint + coverage + security), `accessibility.yml`
  (axe-core + LHCI), `load-test.yml` (k6 500 VU, manual trigger)
- **Load test**: `k6` (open-source, runs locally and in CI)
- **Secret scanning**: `gitleaks` (integrated as a CI step and pre-commit hook)

**Storage**: PostgreSQL 16 (confirmed in spec 001 plan); daily pg_dump or managed-service
automated backup satisfies RPO ≤ 24 h (infrastructure spec decision).

**Testing**: `go test -coverprofile` (backend ≥ 80% business logic), Vitest `@vitest/coverage-v8`
(frontend ≥ 80% shared components/hooks), Playwright (E2E + accessibility), k6 (load)

**Target Platform**: Linux container (backend), modern browsers (frontend, English-only UI at MVP)

**Project Type**: Web application — RESTful JSON API (Go) + Single-Page Application (React)

**Performance Goals**: p95 ≤ 500 ms for non-AI endpoints at 100 RPS; LCP ≤ 2.5 s;
AI acknowledgement ≤ 3 s; 500 concurrent VUs for 10 minutes without degradation

**Constraints**: Stateless backend services; WCAG 2.1 AA; TLS 1.2+ at ingress (infra layer);
RPO ≤ 24 h; English-only UI; no i18n infrastructure at MVP; AI system prompt never exposed to client

**Scale/Scope**: POC baseline — 500 concurrent virtual users; single Docker instance

## Constitution Check

*Re-evaluated post-Phase 1 design: no changes from pre-research check.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | ✅ PASS | Every deliverable (middleware, handler, validator, sanitiser) has a corresponding test file. Coverage gates block merge if floors drop below 80%. |
| II. Simplicity — KISS & DRY | ✅ PASS | `log/slog` (stdlib) over zap/zerolog; pattern-based deny-list over LLM classifier; bluemonday for sanitisation (5-line integration); no speculative abstractions. |
| III. Code Quality & Consistency | ✅ PASS | This spec defines the quality gates themselves. golangci-lint + ESLint enforced in `ci.yml` on every PR. |
| IV. Accessible & Token-Driven UI | ✅ PASS | Privacy policy page uses the design token system. axe-core scan runs on every PR and blocks on WCAG 2.1 AA violations. |
| V. Secure Configuration | ✅ PASS | AI system prompt stored in env var, never hard-coded. Prompt validation config is application config (non-secret, versioned). gitleaks blocks secrets in every commit. |

**No violations — Complexity Tracking table not required.**

## Project Structure

### Documentation (this feature)

```text
specs/002-nfr-system-constraints/
├── plan.md              # This file
├── research.md          # Phase 0 — library and tool decisions
├── data-model.md        # Phase 1 — log entry, health check, prompt validation schemas
├── quickstart.md        # Phase 1 — how to validate each NFR category locally and in CI
├── contracts/
│   └── api.md           # Phase 1 — /healthz contract, X-Request-ID convention, rejection schema
└── tasks.md             # Phase 2 output (/speckit.tasks — NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   └── server/
│       └── main.go                        # register /healthz route and middleware chain
├── internal/
│   └── middleware/
│       ├── requestid.go                   # generate / propagate X-Request-ID (NFR-OBS-003)
│       ├── requestid_test.go
│       ├── logger.go                      # slog HTTP structured-log middleware (NFR-OBS-001)
│       └── logger_test.go
├── internal/
│   └── ai/
│       ├── validator/
│       │   ├── prompt_validator.go        # deny-list pattern validation (NFR-SEC-007)
│       │   └── prompt_validator_test.go
│       └── sanitizer/
│           ├── output_sanitizer.go        # bluemonday HTML sanitisation (NFR-SEC-008)
│           └── output_sanitizer_test.go
└── pkg/
    └── health/
        ├── handler.go                     # GET /healthz response (NFR-OBS-002)
        └── handler_test.go

backend/config/
└── prompt-rules.yml                       # versioned deny-list patterns (NFR-SEC-007)

frontend/
└── src/
    ├── pages/
    │   ├── PrivacyPolicyPage.tsx          # static privacy policy page (NFR-PRIV-003)
    │   └── PrivacyPolicyPage.test.tsx
    └── components/
        └── atoms/
            ├── PrivacyPolicyLink.tsx      # link shown on registration form (NFR-PRIV-003)
            └── PrivacyPolicyLink.test.tsx

.github/
└── workflows/
    ├── ci.yml                             # lint + coverage + secret scan + security scan
    ├── accessibility.yml                  # axe-core (Playwright) + Lighthouse CI
    └── load-test.yml                      # k6 500 VU load test (manual trigger)

tests/
└── load/
    └── k6/
        ├── baseline.js                    # 500 VU sustained test (NFR-SCALE-001)
        └── scenarios/
            └── api_latency.js             # p95 ≤ 500 ms assertion (NFR-PERF-001)
```

**Structure Decision**: Web application layout (backend + frontend + CI). The NFR
infrastructure is woven into the existing domain layout from spec 001. Middleware
and AI utilities live inside `backend/internal/` per the constitution's domain-based
Go layout. The load-test scripts live in `tests/load/` at the repo root, separate
from Go unit/integration tests. CI workflows are in `.github/workflows/` as separate
files to keep failure attribution clean.
