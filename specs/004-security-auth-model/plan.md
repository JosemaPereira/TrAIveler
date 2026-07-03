# Implementation Plan: Security & Authentication/Authorization Model

**Branch**: `004-security-auth-model` | **Date**: 2026-07-03 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/004-security-auth-model/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

This plan defines the security foundation for TrAIveler: JWT-based authentication with multi-key rotation support, role-based authorization (admin/partner), input validation and sanitization (SQL injection, XSS, prompt injection), output sanitization for AI-generated content, data protection (TLS 1.2+, AES-256 at-rest, AWS Secrets Manager), optimistic locking for concurrent modifications, and security logging with 30-day retention in CloudWatch. This model provides backend and frontend engineers with clear, enforceable security contracts for all future endpoints and UI components.

## Technical Context

**Language/Version**: Go 1.24+ (backend), React 19+ with TypeScript strict mode (frontend)

**Primary Dependencies**: 
- Backend: Chi router, pgx/v5 (PostgreSQL), goose/v3 (migrations), Anthropic SDK, slog (structured logging), JWT library (e.g., golang-jwt/jwt/v5), bcrypt for password hashing
- Frontend: TanStack Query, Zustand, React Router v7, Lucide React, Vitest, React Testing Library, MSW

**Storage**: PostgreSQL 15.4 on Amazon RDS (Multi-AZ production, Single-AZ staging)

**Testing**: Go testing + testify/assert + testify/require (backend unit/integration), Vitest + React Testing Library (frontend unit/integration), MSW (API mocking), Playwright (E2E), OWASP ZAP baseline scan (security)

**Target Platform**: AWS ECS Fargate ARM64 Graviton2 containers (backend), S3 + CloudFront CDN (frontend static assets), AWS Secrets Manager (secrets), CloudWatch Logs + Metrics (observability)

**Project Type**: Web application (RESTful API backend + React SPA frontend)

**Performance Goals**: <200ms p95 response time for API endpoints (NFR-PERF-001), <2s initial page load (NFR-PERF-002)

**Constraints**: JWT tokens HTTP-only cookies (XSS prevention), TLS 1.2+ required (data in transit), bcrypt cost factor 12+ (password hashing), 30-day log retention (cost optimization), no automated blocking on security alarms (manual review for MVP)

**Scale/Scope**: Staging environment active ($200/month budget), basic plan limits (1 admin + 1 partner per trip), MVP security foundation (no MFA, no OAuth, no WAF)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development
✅ **PASS** - FR-003, FR-004 require 100% security test coverage. Integration tests for authentication, authorization, prompt injection, and XSS/SQL injection blocking will be written before implementation (Red-Green-Refactor). Test suites defined in quickstart.md.

### II. Simplicity — KISS & DRY
✅ **PASS** - The security model extracts common patterns (JWT validation, prompt validation, output sanitization) into reusable utilities. No speculative abstractions: all requirements (FR-001 through FR-053+) are directly traceable to user stories and success criteria.

### III. Code Quality & Consistency
✅ **PASS** - All Go code will pass gofmt, golangci-lint (errcheck, govet, staticcheck, revive, gosec). All TypeScript will pass ESLint + Prettier. Security utilities will follow coding-guidelines.md (import order, error wrapping, naming conventions).

### IV. Accessible & Token-Driven UI
✅ **PASS** - Authentication UI (login, registration) is out of scope for this spec (assumption documented). When implemented separately, those components will follow WCAG 2.1 AA (visible focus, keyboard operability, form labels) and use design tokens per ui-guidelines.md.

### V. Secure Configuration
✅ **PASS** - FR-040 mandates AWS Secrets Manager for all secrets (DB passwords, API keys, JWT signing keys). FR-041 forbids secrets in code or Git. FR-042 requires runtime secret retrieval with periodic refresh. Constitution security rules (prompt injection prevention, output sanitization, authorization enforcement) are embedded in FR-027 through FR-037.

### Technology Stack Compliance
✅ **PASS** - Backend: Go 1.24+, Chi router, PostgreSQL. Frontend: React 19+, TypeScript strict. Infrastructure: AWS (ECS Fargate, RDS, Secrets Manager, CloudWatch). All mandated technologies from constitution are used.

**Verdict**: ✅ **ALL GATES PASS** — Proceed to Phase 0.

---

## Phase 1 Re-Evaluation (Post-Design)

### I. Test-First Development
✅ **PASS** - quickstart.md defines 9 validation scenarios with concrete test commands and expected outputs. Integration test requirements specified in contracts/api.md (13 required tests). E2E tests defined for authentication flows. All tests can be written before implementation.

### II. Simplicity — KISS & DRY
✅ **PASS** - data-model.md defines 5 entities (User, RefreshToken, JWTSigningKey, SecurityEvent, TripVersion augmentation) without unnecessary abstractions. Security utilities (auth, authorization, validation, concurrency, observability) are separated by domain. No speculative complexity introduced.

### III. Code Quality & Consistency
✅ **PASS** - Project structure in plan.md follows coding-guidelines.md (domain-based internal/ layout). All Go packages use consistent naming (jwt.go, password.go, rbac.go). Test files colocated with source (_test.go suffix).

### IV. Accessible & Token-Driven UI
✅ **PASS** - Frontend structure includes auth context, role hooks, and secure rendering utilities. Security headers defined in contracts/api.md. Authentication UI implementation deferred per assumptions.

### V. Secure Configuration
✅ **PASS** - data-model.md specifies JWT signing keys stored in AWS Secrets Manager with private key never in DB. contracts/api.md enforces HTTPS-only, secure cookies, and security headers. Prompt validation and output sanitization fully specified.

**Post-Design Verdict**: ✅ **ALL GATES PASS** — Design is constitution-compliant. Ready for task generation (/speckit.tasks).

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── auth/                    # Authentication domain
│   │   ├── jwt.go               # JWT token generation, validation, multi-key rotation
│   │   ├── jwt_test.go
│   │   ├── password.go          # Bcrypt password hashing, validation
│   │   ├── password_test.go
│   │   ├── handler.go           # HTTP handlers: login, logout, refresh, password change
│   │   ├── handler_test.go
│   │   ├── middleware.go        # Authentication middleware (JWT validation)
│   │   └── middleware_test.go
│   ├── authorization/           # Authorization domain
│   │   ├── rbac.go              # Role-based access control (admin/partner permissions)
│   │   ├── rbac_test.go
│   │   ├── middleware.go        # Authorization middleware (role enforcement)
│   │   └── middleware_test.go
│   ├── validation/              # Input validation & sanitization
│   │   ├── schema.go            # JSON schema validation
│   │   ├── injection.go         # SQL injection, XSS, path traversal detection
│   │   ├── prompt.go            # Prompt injection validation for AI inputs
│   │   ├── prompt_test.go
│   │   └── sanitize.go          # Output sanitization for AI responses
│   ├── concurrency/             # Optimistic locking
│   │   ├── versioning.go        # Version/timestamp-based conflict detection
│   │   └── versioning_test.go
│   └── observability/           # Security logging & monitoring
│       ├── logger.go            # Structured security event logging
│       ├── metrics.go           # CloudWatch metrics emission
│       └── correlation.go       # Correlation ID generation/propagation
├── tests/
│   ├── integration/
│   │   ├── auth_test.go         # End-to-end authentication flows
│   │   ├── rbac_test.go         # Role-based authorization tests
│   │   ├── injection_test.go    # SQL/XSS/prompt injection blocking tests
│   │   └── concurrency_test.go  # Optimistic locking conflict tests
│   └── security/
│       ├── owasp_vectors_test.go # OWASP Top 10 payload tests
│       └── llm_prompts_test.go   # OWASP LLM Top 10 prompt tests

frontend/
├── src/
│   ├── lib/
│   │   ├── auth.ts              # Auth state management, token refresh logic
│   │   ├── authContext.tsx      # React context for authentication state
│   │   └── secureRender.ts      # Safe rendering utilities (prevent dangerouslySetInnerHTML with AI content)
│   ├── hooks/
│   │   ├── useAuth.ts           # Authentication hook (login, logout, token refresh)
│   │   ├── useRole.ts           # Role-based feature toggling (admin/partner)
│   │   └── useSecureContent.ts  # Hook for safely rendering AI content
│   └── components/
│       └── ProtectedRoute.tsx   # Route guard for authenticated-only pages
├── tests/
│   ├── integration/
│   │   ├── authFlow.test.ts     # Login/logout/refresh integration tests
│   │   └── roleUI.test.ts       # Role-based UI visibility tests
│   └── unit/
│       ├── authContext.test.tsx
│       └── secureRender.test.ts

e2e/
├── tests/
│   ├── auth.spec.ts             # E2E authentication flows (login, logout, token expiration)
│   ├── authorization.spec.ts    # E2E role-based access tests (admin vs partner operations)
│   └── security.spec.ts         # E2E security: XSS attempts, prompt injection UI blocking
```

**Structure Decision**: Web application structure (backend + frontend + e2e). Security utilities are organized by domain: `auth/` for authentication, `authorization/` for RBAC, `validation/` for input/output sanitization, `concurrency/` for optimistic locking, `observability/` for security logging. Frontend mirrors backend concerns with auth state management, role hooks, and secure rendering utilities.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No constitution violations — this section intentionally left empty.
