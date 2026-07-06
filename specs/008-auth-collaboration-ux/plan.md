# Implementation Plan: Authentication & Collaboration User Experience

**Branch**: `008-auth-collaboration-ux` | **Date**: 2026-07-06 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/008-auth-collaboration-ux/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Implement complete authentication and collaboration user experience supporting two user types: Paid Users (subscription-based trip creators) and Free Users (no-payment collaborators with single-trip limit). Includes dual registration paths (with/without payment), session management with automatic renewal, progressive rate limiting for brute-force protection, password reset with single-use tokens, collaboration invitation workflow with suggest-then-approve pattern, subscription lifecycle management (30-day grace period, trip archival, reactivation), Free User upgrade flow preserving existing collaborations, and minimal design system with accessibility compliance. Technical approach uses JWT RS256 tokens (HTTP-only cookies), bcrypt password hashing (cost 12), structured security event logging to CloudWatch, and optimistic locking for concurrent modifications.

## Technical Context

**Language/Version**: Go 1.24+ (backend), React 19+ with TypeScript strict mode (frontend)

**Primary Dependencies**: 
- Backend: Chi (HTTP router), pgx/v5 (PostgreSQL driver), goose (migrations), golang-jwt/jwt v5 (JWT RS256 tokens), bcrypt (password hashing cost 12)
- Frontend: Vite (build tool), TanStack Query v5 (server state), Zustand (client state), React Router v7 (routing), Vitest + React Testing Library (testing), Playwright (E2E)
- Shared: AWS SDK for CloudWatch Logs (security event logging)

**Storage**: PostgreSQL 15.4 on Amazon RDS with optimistic locking (version numbers on User, Trip, Subscription, Collaborator tables)

**Testing**: Go: `go test` with testify assertions; React: Vitest (unit/integration), Playwright (E2E); 80% coverage target for business logic and shared components

**Target Platform**: Web application - ECS Fargate ARM64 (backend), S3+CloudFront (frontend static assets), modern browsers (Chrome/Firefox/Safari/Edge last 2 versions), responsive mobile-first design

**Project Type**: Full-stack web application with RESTful API (JSON), dual user tiers (Paid/Free), subscription-based SaaS model

**Performance Goals**: 
- Registration → first trip creation < 5 minutes (SC-001)
- Login → dashboard render < 10 seconds (SC-002)
- Critical actions (login, create trip, submit suggestion) feedback < 100ms (SC-008)
- 95% first-try authentication success rate (SC-003)

**Constraints**:
- WCAG 2.1 AA accessibility compliance (4.5:1 contrast, keyboard navigation, visible focus indicators)
- HTTP-only Secure SameSite=Strict cookies for token storage (XSS/CSRF protection)
- Generic error messages for authentication failures (prevent account enumeration)
- Progressive delay rate limiting (5 failed attempts → 1s, 2s, 4s, 8s exponential backoff)
- Structured security event logging (no passwords, tokens, or PII beyond user ID in logs)
- Single-use password reset tokens (1-hour expiration, used_at timestamp prevents replay)
- Free User single-collaboration limit (must leave trip to accept new invitation)
- 30-day subscription grace period with read-only access before trip archival

**Scale/Scope**: 
- MVP: < 50 concurrent users, single ECS instance, best-effort uptime
- User types: 2 (Paid User, Free User)
- Registration paths: 2 (with payment, without payment)
- User stories: 6 (P1: Paid User onboarding, Free User onboarding, Login; P2: Collaboration, Password management; P3: Design system)
- Functional requirements: 27
- Key entities: 4 new (PasswordResetToken, UINotification, Subscription.grace_period_ends_at, Trip.archived)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Requirement | Status | Notes |
|-----------|-------------|--------|-------|
| **I. Test-First Development** | TDD Red-Green-Refactor mandatory | ✅ PASS | Plan includes test-first tasks for all auth flows (login, registration, password reset, rate limiting, session renewal); unit tests before handlers, integration tests before E2E |
| **II. Simplicity (KISS/DRY)** | Simplest correct solution; no speculative abstraction | ✅ PASS | No premature abstraction; shared logic extracted only when pattern appears 3+ times (rate limiter, security event logger, token validator reused across endpoints) |
| **III. Code Quality** | Zero lint errors, formatted code, English naming | ✅ PASS | golangci-lint + gofmt (backend), ESLint + Prettier (frontend), all identifiers/comments in English, no dead code |
| **IV. Accessible UI** | WCAG 2.1 AA, design tokens, Atomic Design, Loading/Error/Empty states | ✅ PASS | Spec defines FR-021 (design tokens), SC-004 (keyboard accessibility), SC-005 (zero raw errors), explicit Loading/Error/Empty state requirements (US5) |
| **V. Secure Configuration** | Secrets in env vars, no hardcoded credentials, user-friendly errors | ✅ PASS | Database passwords, JWT signing keys, Anthropic API key in AWS Secrets Manager; generic auth error messages (FR-005); no stack traces to users (FR-019) |
| **V. Prompt Injection Prevention** | Server-side validation, reject override attempts | ✅ PASS | Not directly applicable to auth feature; inherited from spec 002 NFR-SEC-007 for AI endpoints |
| **V. Output Sanitization** | Sanitize before render/persist | ✅ PASS | Not directly applicable to auth feature; inherited from spec 002 NFR-SEC-008 for AI content |
| **V. Authorization Enforcement** | API-layer role checks, client-side advisory only | ✅ PASS | Paid User vs Free User checks enforced at API layer (FR-002a collaboration limit, FR-006a trip creation block, FR-009 suggestion-only access); optimistic locking for concurrent modifications (FR-011, FR-022) |
| **Mandated Stack - Backend** | Go 1.24+, Chi, pgx/v5, goose, PostgreSQL 15.4 | ✅ PASS | All mandated backend technologies used per constitution |
| **Mandated Stack - Frontend** | React 19+, TypeScript strict, Vitest, React Testing Library, Playwright | ✅ PASS | All mandated frontend technologies used per constitution |
| **Mandated Stack - Infrastructure** | AWS, ECS Fargate ARM64, RDS PostgreSQL, CloudWatch, Secrets Manager | ✅ PASS | All mandated infrastructure technologies used per constitution; structured JSON logs to CloudWatch (FR-020) |
| **Mandated Stack - Styling** | CSS Modules + CSS custom properties, no CSS-in-JS | ✅ PASS | Design system uses CSS custom properties (FR-021), CSS Modules for scoping |

**Result**: ✅ **ALL GATES PASS** - No constitution violations. Feature aligns with all core principles and mandated stack.

## Project Structure

### Documentation (this feature)

```text
specs/008-auth-collaboration-ux/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── api.md          # REST API endpoint contracts
├── checklists/
│   └── requirements.md  # Spec quality validation (completed)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   └── api/
│       └── main.go                      # HTTP server entry point
├── internal/
│   ├── auth/                           # Authentication domain
│   │   ├── handler.go                  # Auth HTTP handlers (register, login, logout, password reset)
│   │   ├── service.go                  # Auth business logic
│   │   ├── repository.go               # User, PasswordResetToken data access
│   │   ├── models.go                   # Domain models (LoginRequest, RegisterRequest, etc.)
│   │   ├── jwt/                        # JWT token management
│   │   │   ├── generator.go           # JWT creation (RS256)
│   │   │   ├── validator.go           # JWT validation (multi-key support)
│   │   │   └── refresher.go           # Refresh token exchange
│   │   ├── password/                   # Password management
│   │   │   ├── hasher.go              # bcrypt hashing (cost 12)
│   │   │   └── validator.go           # Password complexity rules
│   │   └── ratelimit/                  # Rate limiting
│   │       ├── limiter.go             # Progressive delay tracker
│   │       └── store.go               # Failed attempt storage (in-memory or Redis)
│   ├── subscription/                   # Subscription & billing domain
│   │   ├── handler.go                  # Subscription HTTP handlers (create, cancel, renew)
│   │   ├── service.go                  # Subscription lifecycle logic (grace period, archival)
│   │   ├── repository.go               # Subscription, Plan data access
│   │   ├── models.go                   # Domain models
│   │   └── payment/
│   │       ├── provider.go            # PaymentProvider interface
│   │       └── stub.go                # Stub implementation (always succeeds)
│   ├── collaboration/                  # Trip collaboration domain
│   │   ├── handler.go                  # Collaboration HTTP handlers (invite, leave, suggestions)
│   │   ├── service.go                  # Collaboration business logic (Free User limits)
│   │   ├── repository.go               # Collaborator, Suggestion data access
│   │   └── models.go                   # Domain models
│   ├── security/                       # Cross-cutting security
│   │   ├── logger.go                   # SecurityEvent structured logging to CloudWatch
│   │   └── middleware.go               # Request ID, auth check, rate limit middleware
│   └── pkg/
│       ├── database/
│       │   ├── connection.go          # PostgreSQL connection pooling (pgx/v5)
│       │   └── migrations/            # goose migration files
│       │       ├── 001_create_users.sql
│       │       ├── 002_create_password_reset_tokens.sql
│       │       ├── 003_add_subscription_grace_period.sql
│       │       └── 004_add_trip_archived.sql
│       └── config/
│           └── config.go              # Environment variable loading
└── tests/
    ├── unit/                           # Unit tests (80% coverage target)
    ├── integration/                    # Integration tests (API + DB)
    └── contract/                       # Contract tests (API spec validation)

frontend/
├── src/
│   ├── features/
│   │   ├── auth/                       # Authentication feature
│   │   │   ├── components/            # LoginForm, RegisterForm, PasswordResetForm
│   │   │   ├── hooks/                 # useAuth, useLogin, useRegister, usePasswordReset
│   │   │   ├── pages/                 # LoginPage, RegisterPage, PasswordResetPage
│   │   │   └── services/              # authApi (TanStack Query hooks)
│   │   ├── subscription/               # Subscription management feature
│   │   │   ├── components/            # SubscriptionCheckout, GracePeriodBanner, UpgradePrompt
│   │   │   ├── hooks/                 # useSubscription, useUpgrade, useCancel
│   │   │   └── services/              # subscriptionApi
│   │   ├── collaboration/              # Collaboration feature
│   │   │   ├── components/            # InviteCollaborator, SuggestionCard, LeaveTrip Button
│   │   │   ├── hooks/                 # useCollaborators, useSuggestions, useInvite
│   │   │   └── services/              # collaborationApi
│   │   └── trips/                      # Trip dashboard & management
│   │       ├── components/            # TripCard, TripDashboard, EmptyState
│   │       ├── hooks/                 # useTrips, useTripDetail
│   │       └── pages/                 # DashboardPage, TripDetailPage
│   ├── components/                     # Shared/Atomic components
│   │   ├── primitives/                # Button, Input, Label, ErrorMessage (Atomic Design)
│   │   ├── composites/                # Form, Modal, Card, Banner
│   │   └── features/                  # LoadingSpinner, EmptyState, ErrorBoundary
│   ├── stores/                         # Zustand global state
│   │   └── authStore.ts               # Auth state (user, has_subscription, isAuthenticated)
│   ├── styles/
│   │   ├── tokens.css                 # CSS custom properties (colors, spacing, typography, focus)
│   │   └── global.css                 # Global resets & base styles
│   ├── services/
│   │   ├── api.ts                     # Axios instance with interceptors (auth, retry, error)
│   │   └── errorHandler.ts            # User-friendly error message mapper
│   ├── routes/
│   │   └── router.tsx                 # React Router v7 routes with protected route wrapper
│   └── App.tsx                         # Root component
└── tests/
    ├── unit/                           # Vitest unit tests (80% coverage for hooks/components)
    ├── integration/                    # MSW mocked API tests
    └── e2e/                            # Playwright E2E tests (primary user flows)

e2e/
└── specs/
    ├── auth/
    │   ├── paid-user-registration.spec.ts     # US1: Paid User onboarding
    │   ├── free-user-registration.spec.ts     # US2: Free User onboarding
    │   ├── login-and-dashboard.spec.ts        # US3: Login & trip management
    │   └── password-management.spec.ts        # US4: Password reset & change
    ├── collaboration/
    │   └── invite-and-suggestions.spec.ts     # US5: Collaboration workflow
    └── accessibility/
        └── wcag-compliance.spec.ts            # WCAG 2.1 AA validation

infra/
└── terraform/
    └── modules/
        └── secrets/
            └── jwt-keys.tf                    # JWTSigningKey rotation infrastructure
```

**Structure Decision**: Web application with backend + frontend + e2e separation (constitution Option 2). Backend uses Go domain-driven structure (`internal/<domain>/`), frontend uses Atomic Design feature modules, E2E tests organized by user story.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

**N/A** - All constitution gates pass. No violations to track.

---

## Phase 0: Research & Technology Decisions

**Output Artifact**: `research.md`

### Research Tasks Completed

1. **JWT Token Storage Strategy**: Evaluated localStorage vs sessionStorage vs HTTP-only cookies → Chose HTTP-only Secure SameSite=Strict cookies (XSS immunity, CSRF protection)
2. **JWT Signing Algorithm**: Evaluated HS256 vs RS256 vs ES256 → Chose RS256 asymmetric (public key distribution, zero-downtime rotation)
3. **Password Hashing**: Evaluated bcrypt vs argon2id vs PBKDF2 vs scrypt → Chose bcrypt cost 12 (proven security, mature Go library, 250ms hashing time)
4. **Rate Limiting Strategy**: Evaluated IP-based vs progressive delay vs account lockout vs CAPTCHA → Chose progressive delay with account-level tracking (no permanent lockout, exponential backoff after 5 failures)
5. **Session Renewal**: Evaluated manual re-login vs long-lived tokens vs automatic refresh → Chose automatic refresh with dedicated refresh token (24h access token + 30d refresh token)
6. **Password Reset Token Security**: Evaluated delete-on-use vs mark-as-used vs expiration-only → Chose mark-as-used with timestamp (audit trail, replay prevention)
7. **Subscription Grace Period**: Evaluated immediate revocation vs grace period vs immediate archival → Chose 30-day grace period with read-only access (customer retention, data protection)
8. **Free User Collaboration Limit**: Evaluated no-collaboration vs unlimited vs single-trip vs time-limited → Chose single trip with leave-to-accept pattern (viral growth + upgrade incentive)
9. **Free User Upgrade Flow**: Evaluated terminate vs preserve vs promote-to-owner → Chose preserve existing collaboration (smooth UX, no disruption)
10. **Security Event Logging**: Evaluated minimal vs structured vs detailed-audit → Chose structured JSON to CloudWatch with correlation IDs (incident investigation, no sensitive data)

**Key Findings**:
- All NEEDS CLARIFICATION items from Technical Context resolved
- All decisions align with docs/security.md, docs/data-model.md, and constitution requirements
- No blocking technical unknowns remaining

**See**: `specs/008-auth-collaboration-ux/research.md` for complete decision rationale, alternatives evaluated, and implementation notes.

---

## Phase 1: Design & Contracts

**Output Artifacts**: `data-model.md`, `contracts/api.md`, `quickstart.md`

### Data Model (data-model.md)

**New Entities**:
- **PasswordResetToken**: Single-use tokens for password reset (token_hash SHA-256, expires_at +1h, used_at timestamp, CASCADE delete on user deletion)
- **SecurityEvent**: Structured authentication/authorization event logging (9 event types, correlation_id for tracing, JSONB details field with no sensitive data, CloudWatch integration)

**Modified Entities**:
- **User**: +4 fields (failed_login_attempts, last_failed_login_at, email_verified, email_verification_token_hash) for rate limiting and future MFA
- **Subscription**: +2 fields (grace_period_ends_at, cancelled_at) for 30-day grace period lifecycle
- **Trip**: +1 field (archived) for subscription lapse archival
- **Collaborator**: Updated validation rules for Free User single-collaboration limit (max 1 active collaboration)

**Indexes Added**: 8 new indexes for high-frequency queries (login, rate limiting, password reset validation, collaboration limit checks, grace period queries)

**Migration Strategy**: 6 new forward-only migrations (003-008) for password reset tokens, security events, user rate limiting fields, subscription grace period, trip archival

**See**: `specs/008-auth-collaboration-ux/data-model.md` for full entity definitions, validation rules ([DB], [Logic], [API] tags), state transitions, business rules, GDPR compliance strategy, and performance optimizations.

---

### API Contracts (contracts/api.md)

**Endpoint Count**: 16 total
- **Authentication**: 7 endpoints (register, login, refresh, logout, password-reset-request, password-reset-complete, password-change)
- **Subscription**: 3 endpoints (create, cancel, renew)
- **Collaboration**: 6 endpoints (invite, leave, create-suggestion, approve-suggestion, reject-suggestion)

**Authentication Pattern**:
- JWT tokens in HTTP-only Secure SameSite=Strict cookies (access_token 24h, refresh_token 30d)
- Automatic token refresh on 401 (frontend calls POST /auth/refresh)
- Progressive delay rate limiting: 5 failed attempts → exponential backoff (1s, 2s, 4s, 8s, 16s)

**Error Format**: Consistent with docs/api-design-standards.md (error code, message, request_id, optional fields array)

**Rate Limiting**:
- Public endpoints (register, login, password-reset-request): 10/hour per IP
- Authenticated endpoints: 100/minute per user
- All responses include X-RateLimit-* headers

**Key Flows**:
- Registration: POST /auth/register (optional payment_method_token for Paid User) → 201 + cookies → redirect /dashboard
- Login: POST /auth/login → 200 + cookies → redirect /dashboard
- Session renewal: 401 Unauthorized → POST /auth/refresh → new access_token → retry original request
- Password reset: POST /auth/password-reset-request (email link) → POST /auth/password-reset-complete (token + new password) → redirect /login
- Collaboration: POST /trips/:id/collaborators (invite) → POST /trips/:id/suggestions (Free User suggests) → PATCH /suggestions/:id/approve (creator approves)
- Subscription lifecycle: POST /subscriptions (upgrade) → DELETE /subscriptions/:id (cancel, start grace period) → POST /subscriptions/:id/renew (reactivate)

**See**: `specs/008-auth-collaboration-ux/contracts/api.md` for complete request/response schemas, validation rules, error codes, side effects, and security event logging.

---

### Quickstart Validation (quickstart.md)

**Scenario Count**: 6 (mapping to 6 user stories in spec.md)

1. **Paid User Registration → First Trip Creation** (US1): Register with payment, verify subscription, create trip
2. **Free User Registration → Accept Invitation** (US2): Register without payment, accept collaboration, verify single-trip limit
3. **Login → Dashboard Access** (US3): Authenticate, access dashboard, verify session persistence, test rate limiting
4. **Password Reset → Password Change** (US4): Request reset, complete via email link, verify token single-use, change password while authenticated
5. **Collaboration → Suggestion → Approval** (US5): Free User suggests, Paid User approves, verify trip updated
6. **Free User Upgrade → Collaboration Preserved** (US2 extension): Upgrade to Paid, verify existing collaboration continues, multi-collaboration enabled, trip creation enabled

**Validation Methods**:
- Browser: Manual UI testing via frontend app (http://localhost:3000)
- cURL: Direct API testing with request/response validation
- SQL: Database queries to verify data integrity and business rules

**Prerequisites**: Backend (localhost:8080), frontend (localhost:3000), PostgreSQL, email stub, payment stub

**Expected Execution Time**: ~30 minutes for all scenarios (manual testing)

**Automation Path**: All scenarios convertible to Playwright E2E tests in `e2e/specs/auth/` and `e2e/specs/collaboration/`

**See**: `specs/008-auth-collaboration-ux/quickstart.md` for step-by-step instructions, cURL commands, SQL validation queries, and expected results.

---

## Constitution Check (Post-Design Re-Evaluation)

*GATE: Re-check after Phase 1 design to ensure no violations introduced by design artifacts.*

| Gate | Pass/Fail | Rationale |
|------|-----------|-----------|
| **I. Test-Driven Development** | ✅ PASS | Quickstart.md defines 6 runnable validation scenarios mapping to user stories. E2E test structure in plan (e2e/specs/auth/, e2e/specs/collaboration/). TDD workflow ready. |
| **II. Simplicity** | ✅ PASS | Data model adds only 2 new entities (PasswordResetToken, SecurityEvent) and 7 fields across existing entities. API design uses standard REST patterns. No unnecessary complexity. |
| **III. Code Quality** | ✅ PASS | API contracts define clear validation rules ([DB], [Logic], [API] tags). Error handling follows docs/api-design-standards.md. Security event logging structured. |
| **IV. Accessibility** | ✅ PASS | UI components follow Atomic Design (docs/ui-guidelines.md). Quickstart scenarios include keyboard navigation validation (implicit in E2E tests). |
| **V. Secure Configuration** | ✅ PASS | JWT keys stored in AWS Secrets Manager (research.md Decision 2). No secrets in code. HTTP-only cookies. bcrypt cost 12. Structured security logging with no sensitive data. |
| **VI. Projects** | ✅ PASS | Backend + frontend + e2e structure maintained (plan.md Project Structure). No new projects added. |
| **VII. Source Code** | ✅ PASS | Go (backend), React (frontend), Playwright (e2e) per constitution. No new languages. |
| **VIII. Language Versions** | ✅ PASS | Go 1.24+, React 19+ per docs. No version changes. |
| **IX. Databases** | ✅ PASS | PostgreSQL (docs/data-model.md). New entities use existing RDS instance. No new databases. |
| **X. Linters** | ✅ PASS | No new linters required. Existing golangci-lint (Go), ESLint (React). |
| **XI. Frameworks** | ✅ PASS | Chi (Go HTTP), React Router (frontend), TanStack Query (state) per constitution. No new frameworks. |
| **XII. Package Dependencies** | ✅ PASS | New: golang.org/x/crypto/bcrypt (password hashing), golang-jwt/jwt/v5 (JWT tokens). Both well-established, security-focused. Aligns with docs/security.md requirements. |

**Result**: All gates pass. Design artifacts introduce no constitution violations. Ready for Phase 2 (tasks generation via `/speckit.tasks`).

**Complexity Tracking Update**: No new violations (still N/A).

---

## Planning Complete

**Branch**: `008-auth-collaboration-ux`  
**Implementation Plan**: `specs/008-auth-collaboration-ux/plan.md` (this file)  
**Generated Artifacts**:
- ✅ `research.md`: 10 technology decisions with rationale, alternatives, implementation notes
- ✅ `data-model.md`: 2 new entities, 7 new fields, 8 new indexes, 6 migrations, validation rules, GDPR compliance
- ✅ `contracts/api.md`: 16 REST endpoints with request/response schemas, error codes, rate limits, security logging
- ✅ `quickstart.md`: 6 runnable validation scenarios with browser/cURL/SQL validation methods

**Next Step**: Run `/speckit.tasks` to generate dependency-ordered task list (tasks.md) for implementation.
