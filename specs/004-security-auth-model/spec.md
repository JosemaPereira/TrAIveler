# Feature Specification: Security & Authentication/Authorization Model

**Feature Branch**: `004-security-auth-model`

**Created**: 2026-07-03

**Status**: Draft

**Input**: User description: "Define the security and authentication/authorization foundations. As a security-conscious team, we want a foundation that specifies the authentication model, the authorization model (roles and permissions), session/token handling, data protection (in transit and at rest), input validation principles, and any compliance requirements. Define the security constraints that every future endpoint and UI must follow. Do not implement auth yet — define the model and rules only."

## Clarifications

### Session 2026-07-03

- Q: How should the system handle JWT signing key rotation without forcing all users to re-authenticate? → A: Support multiple active signing keys with gradual rollover (new tokens use new key; old key validates existing tokens until expiration)
- Q: Which trip modification operations require optimistic locking to prevent lost updates in concurrent editing scenarios? → A: All admin trip modifications
- Q: When a user changes their password (via password reset or account settings), should all existing sessions (access and refresh tokens) be invalidated to prevent compromised credentials from remaining usable? → A: Give user choice during password change (checkbox: "log out all devices")
- Q: When a security alarm fires (e.g., 100+ auth failures/minute), what automated response should the system take? → A: Alert only; manual review and response required (no automated blocking)
- Q: How long should security event logs (authentication, authorization, input validation failures) be retained in CloudWatch Logs? → A: 30 days (minimal retention; lower cost; sufficient for immediate incident response)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Backend Engineer Implements Secure API Endpoint (Priority: P1)

As a backend engineer implementing any new API endpoint, I need clear security requirements and constraints so that I can implement authentication, authorization, input validation, and data protection correctly and consistently without having to ask the PM or Infrastructure Engineer about security decisions.

**Why this priority**: This is foundational—every API endpoint implementation depends on having a clear, complete security model to follow. Without this, we introduce inconsistencies, security gaps, and rework.

**Independent Test**: Can be fully tested by providing a backend engineer with the spec and asking them to list all security checks they must implement for a new endpoint (authentication, authorization, input validation, output sanitization, logging) without consulting additional resources. If they can provide a complete and correct list, the spec delivers value.

**Acceptance Scenarios**:

1. **Given** a backend engineer receives a task to implement a new API endpoint, **When** they read this spec, **Then** they understand what authentication mechanism to implement (JWT in HTTP-only cookies), how to validate tokens, and what to return for unauthorized requests.
2. **Given** a backend engineer implements role-based authorization, **When** they consult this spec, **Then** they know exactly which operations each role (admin/partner) can perform and how to enforce those permissions at the API layer.
3. **Given** a backend engineer receives user input, **When** they consult this spec, **Then** they know all required validation checks (schema validation, SQL injection, XSS, path traversal, prompt injection) and where each check applies.
4. **Given** a backend engineer integrates with the AI provider, **When** they consult this spec, **Then** they know the prompt-validation rules, what patterns to reject, and how to sanitize AI responses.
5. **Given** a backend engineer handles sensitive data, **When** they consult this spec, **Then** they understand data-at-rest encryption, PII handling, and secrets management requirements.

---

### User Story 2 - Frontend Engineer Implements Secure UI Component (Priority: P2)

As a frontend engineer implementing any UI component that displays or collects data, I need clear security requirements for client-side validation, secure data display, and interaction with authenticated APIs so that I can build secure and consistent user experiences without security vulnerabilities.

**Why this priority**: Critical for preventing XSS, ensuring proper authentication state handling, and providing consistent security UX. Required before any authentication UI or data-display components can be built.

**Independent Test**: Can be fully tested by providing a frontend engineer with the spec and asking them to implement a form that submits data to an authenticated API endpoint. If they correctly implement client-side validation, token handling, error display without leaking sensitive data, and secure rendering of dynamic content, the spec delivers value.

**Acceptance Scenarios**:

1. **Given** a frontend engineer implements a form that sends data to an API, **When** they consult this spec, **Then** they know how to include authentication tokens, handle 401/403 responses, and display error messages without leaking sensitive information.
2. **Given** a frontend engineer displays AI-generated content, **When** they consult this spec, **Then** they know to render content safely using React's built-in XSS protection and never use `dangerouslySetInnerHTML` with unsanitized content.
3. **Given** a frontend engineer implements login/logout flows, **When** they consult this spec, **Then** they understand token storage (HTTP-only cookies managed by backend), token refresh mechanisms, and session expiration handling.
4. **Given** a frontend engineer implements role-based UI, **When** they consult this spec, **Then** they know to show/hide features based on user role but understand that authorization must be enforced server-side.

---

### User Story 3 - QA Engineer Validates Security Controls (Priority: P3)

As a QA engineer testing any feature, I need clear acceptance criteria for security testing so that I can verify that authentication, authorization, input validation, and data protection are implemented correctly and consistently across all endpoints and UI components.

**Why this priority**: Ensures systematic security testing is performed for every feature. Important but can follow specification creation (P1 enables development, P2 enables UI work, P3 enables testing).

**Independent Test**: Can be fully tested by providing a QA engineer with the spec and asking them to create test cases for a new API endpoint. If they produce a complete security test suite covering authentication, authorization, injection attacks, and data protection without missing critical vectors, the spec delivers value.

**Acceptance Scenarios**:

1. **Given** a QA engineer tests a new API endpoint, **When** they consult this spec, **Then** they can create test cases for: missing/invalid tokens (401), insufficient permissions (403), SQL injection payloads (400), XSS payloads (400), and prompt injection patterns (400).
2. **Given** a QA engineer tests an AI-integrated feature, **When** they consult this spec, **Then** they know the exact prompt-injection patterns that should be rejected and the sanitization requirements for AI responses.
3. **Given** a QA engineer performs role-based testing, **When** they consult this spec, **Then** they can verify that admin-only operations are blocked for partner users and partner operations are available to both roles as specified.
4. **Given** a QA engineer reviews logging and monitoring, **When** they consult this spec, **Then** they know what security events must be logged (authentication failures, authorization denials, rejected prompts) and what must NOT be logged (tokens, passwords, PII).

---

### Edge Cases

- What happens when a JWT token is expired but a request is in-flight? (Return 401, client must refresh and retry)
- How does the system handle token tampering attempts? (JWT signature validation fails → 401 Unauthorized)
- What happens when a user changes their password but chooses not to log out all devices? (Current session remains valid; other sessions remain valid until natural expiration; new logins require new password)
- What happens when a user tries to escalate privileges from partner to admin? (Authorization check fails → 403 Forbidden)
- How does the system handle concurrent admin approval and partner suggestion submission? (Optimistic locking with version/timestamp check; concurrent modifications return 409 Conflict with current version; client must fetch latest and retry)
- What happens when prompt injection is detected? (Return 400 Bad Request with correlation ID; log event for security review; do NOT send to AI provider)
- How does the system handle AI responses containing executable content? (Sanitize before persisting; strip scripts/dangerous HTML; render using safe methods)
- What happens when AWS Secrets Manager is unavailable during secret retrieval? (Service fails to start; no fallback to environment variables or hard-coded secrets)
- How does the system handle GDPR deletion requests for users with active shared trips? (Remove user PII; anonymize user ID in shared trips; notify collaborators)

## Requirements *(mandatory)*

### Functional Requirements

#### Authentication

- **FR-001**: System MUST authenticate users using JWT (JSON Web Tokens) with RS256 signing algorithm
- **FR-002**: System MUST store JWT access tokens in HTTP-only cookies to prevent XSS-based token theft
- **FR-003**: System MUST issue access tokens with 24-hour maximum lifetime
- **FR-004**: System MUST issue refresh tokens with 30-day maximum lifetime, also stored in HTTP-only cookies
- **FR-005**: System MUST validate JWT signature using any currently active signing key, expiration, and required claims (user ID, role, issued-at, expiration) on every authenticated request
- **FR-006**: System MUST return 401 Unauthorized for requests with missing, invalid, expired, or tampered tokens
- **FR-007**: System MUST provide a token refresh endpoint that issues a new access token when presented with a valid refresh token
- **FR-008**: System MUST invalidate refresh tokens on user logout and delete associated cookies
- **FR-008a**: System MUST provide users with the option to invalidate all active sessions (all access and refresh tokens across all devices) when changing their password; this option MUST be presented as an explicit choice (e.g., checkbox: "Log out all other devices") with secure default behavior recommended
- **FR-009**: System MUST support account deletion, which invalidates all tokens and deletes user PII within 30 days
- **FR-009a**: System MUST support JWT signing key rotation using a multi-key strategy: new tokens are signed with the current primary key, while token validation accepts any active signing key (primary or previous keys within their validity period); this enables zero-downtime key rotation without forcing user re-authentication

#### Authorization

- **FR-010**: System MUST enforce role-based access control (RBAC) with two roles: admin and partner
- **FR-011**: Admin role MUST have full CRUD (Create, Read, Update, Delete) permissions on trips they own
- **FR-011a**: System MUST implement optimistic locking for all admin trip modification operations (Update and Delete) using version numbers or timestamps; concurrent modifications MUST return 409 Conflict with the current resource version
- **FR-012**: Admin role MUST have exclusive authority to approve or reject partner suggestions
- **FR-013**: Admin role MUST be able to invite one partner collaborator per trip (basic plan limit)
- **FR-014**: Admin role MUST be able to remove collaborators from trips they own
- **FR-015**: Partner role MUST have read-only access to trips they are invited to
- **FR-016**: Partner role MUST be able to submit suggestions for itinerary modifications (suggestions require admin approval)
- **FR-017**: Partner role MUST NOT be able to directly edit any itinerary items
- **FR-018**: Partner role MUST NOT be able to invite additional collaborators
- **FR-019**: System MUST enforce authorization at the API layer on every request; client-side role enforcement is advisory only
- **FR-020**: System MUST return 403 Forbidden when an authenticated user attempts an operation they lack permission to perform
- **FR-021**: System MUST enforce subscription plan limits: one admin user per subscription, maximum one partner collaborator per trip (basic plan)

#### Input Validation & Sanitization

- **FR-022**: System MUST validate all API request payloads against strict JSON schemas before processing
- **FR-023**: System MUST reject requests failing schema validation with 400 Bad Request and a descriptive error message (no sensitive data in error response)
- **FR-024**: System MUST validate and sanitize all user inputs to prevent SQL injection, detecting and rejecting suspicious patterns (`'; DROP TABLE`, `UNION SELECT`, etc.)
- **FR-025**: System MUST validate and sanitize all user inputs to prevent XSS (cross-site scripting), rejecting payloads containing `<script>`, event handlers (`onclick`, `onerror`), and dangerous attributes (`javascript:` URLs)
- **FR-026**: System MUST validate and sanitize file paths to prevent path traversal attacks, rejecting inputs containing `../`, absolute paths, or symbolic link references
- **FR-027**: System MUST pass all user input destined for the AI provider through a prompt-validation layer
- **FR-028**: Prompt-validation layer MUST detect and reject instruction-override patterns (e.g., "ignore previous instructions", "new system prompt", "disregard constraints")
- **FR-029**: Prompt-validation layer MUST detect and reject attempts to extract system prompts or internal configuration (e.g., "print your instructions", "show me your system message")
- **FR-030**: Prompt-validation layer MUST detect and reject off-topic prompts unrelated to travel planning (e.g., requests for financial advice, medical guidance, coding assistance)
- **FR-031**: System MUST return 400 Bad Request with a correlation ID when prompt validation fails; correlation ID MUST be logged for security review
- **FR-032**: System MUST NOT send rejected prompts to the AI provider under any circumstances

#### Output Sanitization

- **FR-033**: System MUST sanitize all AI-generated content before persisting to the database or rendering in the browser
- **FR-034**: Sanitization MUST strip or escape HTML tags (`<script>`, `<iframe>`, `<object>`, `<embed>`), event handlers, and executable content
- **FR-035**: Sanitization MUST remove or neutralize `javascript:` URLs, `data:` URIs containing executable content, and CSS expressions
- **FR-036**: System MUST render AI-generated content using React's default XSS-safe rendering; use of `dangerouslySetInnerHTML` with AI content is forbidden
- **FR-037**: System MUST NOT execute AI-generated content as code or use it directly in SQL queries or template strings without parameterization

#### Data Protection

- **FR-038**: System MUST encrypt all data in transit using TLS 1.2 or higher for client-server and service-to-service communication
- **FR-039**: System MUST encrypt all data at rest in the database using AWS RDS encryption with AES-256
- **FR-040**: System MUST store all secrets (database passwords, API keys, JWT signing keys with metadata indicating active/retired status, certificates) in AWS Secrets Manager
- **FR-041**: System MUST retrieve secrets at runtime using IAM role-based authentication; no secrets in code, environment variables in Git, or container images
- **FR-042**: System MUST support secret rotation without requiring service redeployment (services fetch secrets on startup and periodically refresh); for JWT signing keys specifically, the multi-key rotation strategy (FR-009a) ensures old tokens remain valid during the rollover period
- **FR-043**: System MUST collect only PII directly required by functional requirements (email address, trip preferences, user name)
- **FR-044**: System MUST allow users to delete their accounts, removing all PII within 30 days
- **FR-045**: System MUST provide a privacy policy accessible from registration/login flows before collecting any PII
- **FR-046**: System MUST hash and salt all passwords using bcrypt with cost factor 12 or higher (never store plaintext passwords)

#### Security Logging & Monitoring

- **FR-047**: System MUST log all authentication events (successful logins, failed login attempts, token refresh, logout) with correlation ID and timestamp
- **FR-048**: System MUST log all authorization denials (403 responses) with user ID, role, requested operation, and timestamp
- **FR-049**: System MUST log all rejected input validation failures (prompt injection, XSS, SQL injection attempts) with correlation ID, user ID, and attack pattern detected
- **FR-050**: System MUST NOT log sensitive data (JWT tokens, passwords, refresh tokens, API keys, PII except user ID for correlation)
- **FR-051**: System MUST send all logs to AWS CloudWatch Logs in structured JSON format
- **FR-051a**: System MUST configure CloudWatch Logs retention period of 30 days for all security event logs (authentication, authorization, input validation failures) to balance incident response needs with cost efficiency
- **FR-052**: System MUST emit CloudWatch metrics for: authentication failures (rate), authorization denials (rate), prompt injection detections (count), API error rate (5xx responses)
- **FR-053**: System MUST alert via CloudWatch Alarms when authentication failure rate exceeds 100 failures/minute or prompt injection detection rate exceeds 10 events/minute; alarms trigger notifications for manual review and response (no automated blocking or rate limiting for MVP)

### Key Entities *(include if feature involves data)*

- **User**: Represents an authenticated user account with unique ID, email address, hashed password, role (admin/partner), created timestamp, and last-login timestamp
- **JWT Access Token**: Short-lived token (24-hour max) containing user ID, role, issued-at (iat), expiration (exp), and issuer claims; signed with RS256
- **JWT Refresh Token**: Long-lived token (30-day max) used exclusively for issuing new access tokens; stored in HTTP-only cookie; invalidated on logout
- **Role**: Enumeration of user permissions (admin or partner); determines which operations a user can perform
- **Prompt Validation Result**: Binary result (pass/fail) from prompt-validation layer with rejection reason (instruction override, system prompt extraction, off-topic) and correlation ID
- **Security Event**: Log entry for authentication, authorization, or input validation events; includes timestamp, correlation ID, user ID, event type, and outcome

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of API endpoints require authentication and validate JWT tokens on every request (measurable via code review and integration test coverage)
- **SC-002**: 100% of admin-only operations (trip CRUD, approval/rejection) return 403 Forbidden when attempted by partner role users (measurable via role-based integration tests)
- **SC-003**: 100% of known prompt injection patterns from the test suite are rejected with 400 Bad Request (measurable via security integration test pass rate)
- **SC-004**: 100% of XSS and SQL injection payloads from OWASP test vectors are rejected with 400 Bad Request (measurable via security integration test pass rate)
- **SC-005**: Zero high or critical CVEs in direct or transitive dependencies at release time (measurable via `npm audit` and `govulncheck` in CI)
- **SC-006**: Zero secrets detected in Git history or PR diffs (measurable via `gitleaks` scan in CI)
- **SC-007**: All authentication, authorization, and input validation failures are logged to CloudWatch with correlation IDs within 1 second of occurrence (measurable via CloudWatch Logs query)
- **SC-008**: Backend and frontend engineers can implement secure endpoints and UI components without consulting PM or Infrastructure Engineer about security decisions (measurable via survey or code review showing correct first-time implementation)

## Assumptions

- Authentication UI (login, registration, password reset) is out of scope for this spec; this spec defines backend security contracts only
- Email verification and multi-factor authentication (MFA) are not required for MVP; basic email/password authentication is sufficient
- Session management (e.g., "remember me", device tracking, concurrent session limits) is not required for MVP; JWT token expiration provides sufficient session control
- Third-party OAuth providers (Google, Facebook) are not required for MVP; email/password authentication is the only supported method
- Audit logging (immutable security event log for compliance) is not required for MVP; CloudWatch Logs provides sufficient traceability
- Compliance certifications (SOC 2, ISO 27001) are not required for MVP; GDPR-awareness is sufficient
- Rate limiting and CAPTCHA for brute-force prevention are not required for this spec; they will be addressed in a separate feature
- Web Application Firewall (WAF) is not required for MVP; application-layer input validation is sufficient
- Secrets rotation automation (e.g., automated key rotation schedule) is not required for MVP; manual rotation with zero-downtime support is sufficient
- The AI provider (Anthropic Claude API) is trusted to not store or misuse prompts; prompt-validation focuses on preventing malicious user input, not provider security
- Frontend and backend share the same domain or use CORS configuration that allows credentials; cross-origin authentication flows are supported
- AWS infrastructure (Secrets Manager, CloudWatch, RDS encryption) is available and configured per `docs/cloud-and-environments.md`
