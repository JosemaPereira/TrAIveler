# Non-Functional Requirements

<!-- PROMOTED:nfrs START -->
<!-- Generated from specs/002-nfr-system-constraints/spec.md -->
<!-- Last promoted: 2026-07-03 -->

## Overview

These are system-wide constraints that every feature and architectural decision must satisfy. They do not describe what the system does — they describe how well it does it and under what conditions it remains trustworthy, usable, and maintainable.

These requirements are binding across all future specs. Any proposed feature or architectural component that would violate an NFR defined here requires an explicit constitution amendment or an approved NFR revision before proceeding.

## Performance (NFR-PERF)

| ID | Requirement | Measurable Target | Validation Method |
|----|-------------|-------------------|-------------------|
| **NFR-PERF-001** | Non-AI API endpoints MUST respond within a p95 latency threshold under normal load | p95 latency ≤ 500 ms at 100 RPS | Load test (k6) on staging; p95 reported in CI artifact |
| **NFR-PERF-002** | AI itinerary generation requests MUST be acknowledged promptly, with results delivered asynchronously | Initial acknowledgement ≤ 3 s; full result delivery ≤ 30 s for typical requests | Integration test with timeout assertions; manual timing on representative prompts |
| **NFR-PERF-003** | The web application MUST meet Core Web Vitals "Good" thresholds | LCP ≤ 2.5 s; CLS ≤ 0.1; INP ≤ 200 ms on simulated 4G connection | Lighthouse CI audit in pull-request pipeline |
| **NFR-PERF-004** | The server-side error rate MUST remain low under normal operating conditions | Error rate (5xx) < 0.5% of requests under normal load | Load test error-rate assertion; production monitoring dashboard |

## Scalability (NFR-SCALE)

| ID | Requirement | Measurable Target | Validation Method |
|----|-------------|-------------------|-------------------|
| **NFR-SCALE-001** | The system MUST support concurrent users at MVP launch without degradation | 500 concurrent virtual users sustained for 10 minutes without exceeding NFR-PERF targets | Load test (k6) on staging with VU ramp-up to 500 |
| **NFR-SCALE-002** | Backend services MUST be stateless to allow horizontal scaling | A second backend instance can be added without any shared in-process state | Architectural review checklist; integration test running two instances simultaneously |
| **NFR-SCALE-003** | The system MUST shed load gracefully when capacity is exceeded | Returns `503 Service Unavailable` with `Retry-After` header when load exceeds capacity | Load test driving requests beyond capacity; assert 503 + header |

## Availability and Reliability (NFR-AVAIL)

| ID | Requirement | Measurable Target | Validation Method |
|----|-------------|-------------------|-------------------|
| **NFR-AVAIL-001** | The application MUST meet a minimum uptime target for the MVP | ≥ 99.5% monthly uptime (≤ 3.6 hours downtime/month) | Uptime monitoring (synthetic ping every 60 s); monthly SLA report |
| **NFR-AVAIL-002** | Saved user itineraries MUST not be lost due to a single infrastructure failure | RPO ≤ 24 hours: automated daily backup | Disaster-recovery drill before first release |
| **NFR-AVAIL-003** | The system MUST recover from a restart within a defined time window | Cold-start to ready-to-serve ≤ 60 s | Health-check polling; assert `/healthz` returns `200 OK` within 60 s |
| **NFR-AVAIL-004** | Mean Time To Recovery for P1 incidents MUST be bounded | MTTR ≤ 30 minutes from alert to service restoration | Incident retrospective tracking; simulated chaos drill before first release |

## Accessibility (NFR-A11Y)

| ID | Requirement | Measurable Target | Validation Method |
|----|-------------|-------------------|-------------------|
| **NFR-A11Y-001** | Every user-facing page MUST comply with WCAG 2.1 Level AA | Zero WCAG 2.1 AA violations on all application pages | Automated `axe-core` scan integrated into E2E test suite; run on every PR |
| **NFR-A11Y-002** | All interactive elements MUST be keyboard-operable with visible focus indicators | 100% of interactive elements reachable and activatable via keyboard | Manual keyboard walkthrough before each milestone release |
| **NFR-A11Y-003** | Text and UI components MUST meet minimum color contrast requirements | 4.5:1 contrast ratio for normal text; 3:1 for large text and UI components | Automated contrast check via axe-core |
| **NFR-A11Y-004** | The application MUST achieve a minimum Lighthouse accessibility score | Lighthouse accessibility score ≥ 90 on all primary pages | Lighthouse CI audit in pull-request pipeline |

## Security (NFR-SEC)

| ID | Requirement | Measurable Target | Validation Method |
|----|-------------|-------------------|-------------------|
| **NFR-SEC-001** | No secrets, API keys, or credentials MUST be committed to the repository | Zero secrets detected by automated secret-scanning on every commit | Secret-scanning tool (`gitleaks`) integrated in CI; pre-commit hook and PR gate |
| **NFR-SEC-002** | All data in transit MUST be encrypted | TLS 1.2 or higher required for all client-server and service-to-service communication | TLS configuration audit; integration test asserting HTTPS enforced |
| **NFR-SEC-003** | The backend MUST have no high-severity static-analysis security findings | `gosec` reports zero high-severity findings on every PR | `gosec` integrated into CI lint gate |
| **NFR-SEC-004** | Known vulnerable dependencies MUST not be shipped | Zero critical or high CVEs in direct and transitive dependencies at release | `govulncheck` (Go) and `npm audit --audit-level=high` (frontend) in CI |
| **NFR-SEC-005** | All API inputs MUST be validated and sanitized before use | Automated injection tests (SQLi, XSS, path traversal) return 400 for malicious payloads | Security-focused integration test suite; OWASP ZAP baseline scan on staging |
| **NFR-SEC-006** | Authentication tokens MUST expire within a bounded window | Access tokens expire within 24 hours; refresh tokens expire within 30 days | Unit test asserting token TTL values; integration test confirming expired tokens rejected with 401 |
| **NFR-SEC-007** | All user input sent to the AI provider MUST be validated and scoped to travel-planning purpose | Prompt-validation layer rejects instruction-override patterns, prompt extraction attempts, and off-topic prompts with 400 response | Unit tests covering known injection patterns; integration test asserting 400 for malicious prompts |
| **NFR-SEC-008** | All AI-generated content MUST be sanitized before rendering or persisting | HTML/script tags and executable content stripped or escaped before storage and rendering | Unit tests asserting XSS payloads in mocked AI responses are neutralized |

## Maintainability (NFR-MAINT)

| ID | Requirement | Measurable Target | Validation Method |
|----|-------------|-------------------|-------------------|
| **NFR-MAINT-001** | Backend business-logic code MUST maintain minimum test coverage | ≥ 80% line coverage for Go packages in `internal/` | `go test -coverprofile` in CI; coverage gate fails if below 80% |
| **NFR-MAINT-002** | Frontend shared components and hooks MUST maintain minimum test coverage | ≥ 80% line coverage for files in `src/components/` and `src/hooks/` | Vitest coverage report in CI; gate fails if below 80% |
| **NFR-MAINT-003** | All code MUST pass automated linting with zero errors before merge | `golangci-lint` (Go) and ESLint (TypeScript) report zero errors on every PR | Lint gates in CI; PRs cannot be merged while lint gates failing |
| **NFR-MAINT-004** | All public API contracts MUST be documented before shipping | Every REST endpoint has entry in `specs/*/contracts/api.md` with request/response schema | Contract-review checklist item on every PR adding/modifying endpoint |
| **NFR-MAINT-005** | Dead code and commented-out code MUST not be committed | Zero instances of commented-out code blocks or unreachable code detected by linters | ESLint `no-unused-vars`, `no-unreachable`; `staticcheck` (Go) |

## Data Privacy (NFR-PRIV)

| ID | Requirement | Measurable Target | Validation Method |
|----|-------------|-------------------|-------------------|
| **NFR-PRIV-001** | The system MUST support user right-to-deletion | Deleting user account removes all PII within 30 days | Integration test asserting zero records for deleted user; manual audit before each milestone |
| **NFR-PRIV-002** | The system MUST store only minimum PII necessary | No PII field persisted unless directly required by functional requirement | Data model review checklist on every PR adding entity attribute with personal data |
| **NFR-PRIV-003** | A privacy policy page MUST be available and linked from registration/login flow | Privacy policy page reachable from sign-up page | E2E test asserting privacy policy link present and navigable on registration page |

## Observability (NFR-OBS)

| ID | Requirement | Measurable Target | Validation Method |
|----|-------------|-------------------|-------------------|
| **NFR-OBS-001** | Every HTTP request processed by backend MUST produce structured log entry | 100% of requests emit JSON log line with: timestamp, correlation ID, method, path, status code, duration, service name | Integration test asserting log output structure on sample requests |
| **NFR-OBS-002** | A health-check endpoint MUST be available on all backend services | `GET /healthz` returns `200 OK` with `{"status":"ok"}` within 100 ms | Automated health-check test in integration suite; monitored in production every 60 s |
| **NFR-OBS-003** | All request/response correlation MUST be traceable via unique ID | Every HTTP response includes `X-Request-ID` header matching correlation ID in server log | Integration test asserting header presence and log/header correlation ID match |
| **NFR-OBS-004** | An alerting rule MUST fire when error rate exceeds defined threshold | Alert triggers when 5xx error rate > 1% over 5-minute rolling window | Alerting rule definition reviewed; simulated error-injection test validates alert fires |

<!-- PROMOTED:nfrs END -->
