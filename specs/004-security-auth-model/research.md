# Research: Security & Authentication/Authorization Model

**Feature**: 004-security-auth-model | **Date**: 2026-07-03

**Purpose**: Document architectural decisions, rationale, and alternatives considered for the security foundation.

---

## Decision 1: JWT Signing Algorithm — RS256 (Asymmetric)

**Decision**: Use RS256 (RSA Signature with SHA-256) for JWT token signing.

**Rationale**:
- Public key can be shared with frontend and other services for token validation without exposing signing capability
- Private key remains secure in backend only; only the backend can issue tokens
- Industry standard for distributed systems where token verification happens in multiple locations
- Required by constitution (inherited from docs/security.md)

**Alternatives Considered**:
- **HS256 (HMAC with SHA-256)**: Simpler symmetric algorithm but requires sharing the secret key with all validators, increasing attack surface. If the secret leaks from any service, an attacker can forge tokens. Rejected due to security risk in distributed architecture.
- **ES256 (ECDSA with P-256)**: Smaller key size and faster verification than RSA, but less widely supported in Go JWT libraries and requires careful curve selection. Rejected in favor of RS256's broader ecosystem support and proven track record.

---

## Decision 2: JWT Key Rotation Strategy — Multi-Key Validation

**Decision**: Support multiple active signing keys simultaneously; new tokens use the current primary key, validation accepts any active key within its validity period.

**Rationale**:
- Enables zero-downtime key rotation without forcing mass user logout
- Industry standard approach (Auth0, Okta, AWS Cognito all use this pattern)
- Addresses FR-042 requirement for secret rotation without redeployment
- Balances security (regular key rotation) with availability (users remain authenticated)

**Clarification Source**: Clarification session 2026-07-03, Question 1 → Option B selected

**Alternatives Considered**:
- **Single active key with immediate rotation**: Simple but causes mass logout on rotation. Rejected due to poor user experience and operational risk.
- **Key version in token claims**: More complex versioning logic; requires storing key version in every token. Rejected in favor of implicit validation against all active keys (simpler and equally secure).

**Implementation Notes**:
- Store keys in AWS Secrets Manager with metadata: `{key_id, public_key, private_key, status: "active"|"retired", created_at, retire_at}`
- Validation loop iterates through all active keys until one successfully verifies the token
- Retired keys remain in storage for audit but are excluded from validation

---

## Decision 3: Token Storage — HTTP-Only Cookies

**Decision**: Store JWT access and refresh tokens in HTTP-only, Secure, SameSite=Strict cookies.

**Rationale**:
- HTTP-only flag prevents JavaScript access → XSS attacks cannot steal tokens
- Secure flag enforces HTTPS-only transmission → prevents man-in-the-middle token theft
- SameSite=Strict prevents CSRF attacks by blocking cross-origin requests with cookies
- Required by FR-002 (HTTP-only cookies to prevent XSS-based token theft)

**Alternatives Considered**:
- **LocalStorage/SessionStorage**: Accessible to any JavaScript running on the page, including malicious scripts injected via XSS. Rejected due to high XSS vulnerability.
- **Memory-only (React state)**: Secure but tokens lost on page refresh, requiring re-authentication on every browser reload. Rejected due to poor user experience.

---

## Decision 4: Token Lifetime — 24h Access + 30d Refresh

**Decision**: Access tokens expire after 24 hours max, refresh tokens after 30 days max.

**Rationale**:
- 24-hour access tokens balance security (limited exposure window if compromised) with UX (users not frequently interrupted)
- 30-day refresh tokens allow "remember me" convenience without indefinite access
- If access token is compromised, exposure window is <24 hours before natural expiration
- Aligns with industry standards for web applications (e.g., GitHub, GitLab use similar lifetimes)

**Alternatives Considered**:
- **Shorter access tokens (15 min)**: Higher security but requires frequent refresh requests, increasing API load and user-visible latency. Rejected as over-engineering for MVP threat model.
- **Longer refresh tokens (90+ days)**: Reduces re-authentication frequency but increases risk if refresh token is stolen (compromised device remains authenticated for months). Rejected due to security risk exceeding MVP risk tolerance.

---

## Decision 5: Password Hashing — bcrypt Cost Factor 12

**Decision**: Use bcrypt with cost factor 12 (2^12 = 4096 iterations) for password hashing.

**Rationale**:
- Bcrypt is resistant to brute-force attacks due to intentional computational cost
- Cost factor 12 provides strong security (OWASP recommendation for 2026) while maintaining <200ms hashing time on modern hardware
- Adaptive algorithm: cost factor can be increased in future without changing codebase
- Required by FR-046 (bcrypt cost factor 12 or higher)

**Alternatives Considered**:
- **Argon2id**: Newer algorithm with better resistance to GPU/ASIC attacks. However, less mature ecosystem in Go compared to bcrypt. Deferred for post-MVP evaluation.
- **PBKDF2**: Older standard with configurable iterations, but more vulnerable to GPU acceleration than bcrypt. Rejected in favor of bcrypt's proven track record.

**Implementation Notes**:
- On user registration: `bcrypt.GenerateFromPassword(password, 12)`
- On login: `bcrypt.CompareHashAndPassword(storedHash, providedPassword)`
- Hashing time on staging ECS Fargate (ARM64): ~150ms avg (within p95 <200ms budget)

---

## Decision 6: Optimistic Locking Strategy — Version Numbers

**Decision**: Implement optimistic locking for all admin trip modification operations using monotonically incrementing version numbers (integer column in database).

**Rationale**:
- Prevents lost updates when admin approves a suggestion while another admin edits the same trip
- Version number approach is simpler than timestamp-based locking (no clock skew issues)
- Database-native integer increment is atomic and reliable
- Required by FR-011a (optimistic locking for admin trip modifications)

**Clarification Source**: Clarification session 2026-07-03, Question 2 → All admin trip modifications

**Alternatives Considered**:
- **Last-write-wins (no locking)**: Simplest but loses data when concurrent modifications occur. Rejected due to risk of overwriting partner suggestions or admin edits.
- **Pessimistic locking (row locks)**: Prevents conflicts but requires holding database locks during user think time, reducing concurrency. Rejected as over-engineering for MVP scale (1-2 concurrent editors per trip).
- **Timestamp-based optimistic locking**: Similar to version numbers but vulnerable to clock skew between servers. Rejected in favor of version number simplicity.

**Implementation Notes**:
- Add `version` column (BIGINT) to `trips` and `itinerary_items` tables
- On update: `UPDATE trips SET ..., version = version + 1 WHERE id = $1 AND version = $2`
- If no rows updated (version mismatch), return 409 Conflict with current resource state

---

## Decision 7: Password Change Session Invalidation — User Choice

**Decision**: When users change their password, provide an explicit checkbox option to invalidate all active sessions ("Log out all other devices").

**Rationale**:
- Balances security (compromised credentials can be fully revoked) with UX (user may be changing password on multiple trusted devices)
- Follows industry best practice (Google, Apple, Microsoft all offer this option)
- Educates users about session management through explicit choice
- Required by FR-008a (user choice during password change)

**Clarification Source**: Clarification session 2026-07-03, Question 3 → Option C (user choice)

**Alternatives Considered**:
- **Always invalidate all sessions**: Maximum security but frustrating UX if user changes password from one device and must re-login on all others immediately. Rejected due to poor UX for legitimate use cases.
- **Never invalidate sessions**: Preserves UX but leaves compromised credentials usable until natural token expiration (up to 30 days). Rejected due to security risk.

**Implementation Notes**:
- Password change endpoint accepts optional boolean parameter: `invalidate_all_sessions`
- If true, delete all refresh tokens from database (access tokens expire naturally within 24h)
- Frontend presents checkbox (default unchecked) with helper text: "Recommended if you suspect your account was compromised"

---

## Decision 8: Prompt Injection Detection — Pattern Matching + Keyword Filtering

**Decision**: Implement prompt validation layer using pattern matching for known instruction-override patterns and keyword filtering for off-topic prompts.

**Rationale**:
- Deterministic validation (no ML models) ensures predictable behavior and low latency (<10ms overhead)
- Pattern matching catches known attack vectors (OWASP LLM01: "ignore previous", "new system prompt", etc.)
- Keyword filtering prevents off-topic abuse (financial advice, medical guidance, code generation)
- Required by FR-027 through FR-032 (prompt validation layer with rejection of instruction override, system prompt extraction, and off-topic prompts)

**Alternatives Considered**:
- **ML-based classifier**: Higher accuracy but introduces latency (50-200ms inference), model training/maintenance overhead, and unpredictable false positives. Deferred for post-MVP evaluation.
- **AI provider's built-in filtering**: Anthropic Claude has content moderation, but doesn't block all prompt injection patterns and adds API round-trip latency. Rejected as insufficient; our validation is first-line defense.

**Implementation Notes**:
- Reject patterns: `/(ignore|disregard|forget).*(previous|prior|above|instructions|prompt)/i`, `/new system (prompt|message|instructions)/i`, `/(print|show|reveal).*(instructions|system prompt|configuration)/i`
- Off-topic keywords: finance, stock, invest, medical, diagnose, prescription, code, programming, SQL
- Return 400 Bad Request with correlation ID; log rejected prompt (first 100 chars) for security review

---

## Decision 9: Output Sanitization — HTML Stripping + Safe Rendering

**Decision**: Sanitize all AI-generated content by stripping dangerous HTML tags and event handlers before persisting; render using React's default XSS-safe rendering (never `dangerouslySetInnerHTML` with AI content).

**Rationale**:
- Defense-in-depth: sanitize at ingress (before DB) and at display (React rendering)
- HTML stripping removes `<script>`, `<iframe>`, `<object>`, event handlers (`onclick`, etc.), `javascript:` URLs
- React's default rendering escapes all special characters, preventing XSS even if sanitization is bypassed
- Required by FR-033 through FR-037 (sanitize before persist/render; strip HTML/scripts; use React's XSS-safe rendering)

**Alternatives Considered**:
- **Markdown-only output from AI**: Force AI to return Markdown instead of HTML. However, Anthropic Claude occasionally returns HTML despite instructions. Rejected as unreliable without enforcement layer.
- **Content Security Policy (CSP) only**: CSP is defense-in-depth but doesn't prevent stored XSS in database. Rejected as insufficient; CSP should be added but not as sole mitigation.

**Implementation Notes**:
- Use Go library like `bluemonday` or `github.com/microcosm-cc/bluemonday` for HTML sanitization (strict policy: allow only plain text, strip all tags)
- Frontend: render AI content with `{aiContent}` (React escapes automatically), never `<div dangerouslySetInnerHTML={{__html: aiContent}} />`

---

## Decision 10: Security Alarm Response — Manual Review (No Automated Blocking)

**Decision**: CloudWatch alarms for high authentication failure rates and prompt injection detection trigger notifications only; no automated IP blocking or account suspension for MVP.

**Rationale**:
- Automated blocking risks false positives that could lock out legitimate users or entire office networks
- MVP threat model prioritizes availability over automated defense (staging environment, pre-revenue)
- Manual review allows security team to assess patterns before implementing automated responses
- Required by FR-053 (alert for manual review; no automated blocking)

**Clarification Source**: Clarification session 2026-07-03, Question 4 → Option B (manual review)

**Alternatives Considered**:
- **Automated IP blocking**: Blocks attackers faster but risks collateral damage (NAT-ed offices, shared IPs, VPN users). Deferred until production with rate-limiting infrastructure (AWS WAF).
- **Automated PagerDuty escalation**: Creates alert fatigue if detection is too sensitive. Rejected in favor of CloudWatch SNS email notifications for MVP (lower urgency, manual triage).

**Implementation Notes**:
- CloudWatch alarm: authentication failure rate >100/min for 5 consecutive minutes → SNS email to security team
- CloudWatch alarm: prompt injection detection rate >10/min for 5 consecutive minutes → SNS email to security team
- Alarms include link to CloudWatch Logs Insights query for recent events

---

## Decision 11: Log Retention Period — 30 Days

**Decision**: Configure CloudWatch Logs retention period of 30 days for all security event logs.

**Rationale**:
- 30 days is sufficient for incident response (most security incidents detected within 7-14 days)
- Cost-optimized for MVP staging budget ($200/month includes compute, storage, logs)
- Meets immediate forensic investigation needs without enterprise compliance overhead
- Required by FR-051a (30-day retention for cost optimization)

**Clarification Source**: Clarification session 2026-07-03, Question 5 → Option A (30 days)

**Alternatives Considered**:
- **90 days retention**: Better for delayed incident discovery but 3x log storage cost. Deferred until production with higher security budget.
- **1 year retention**: Required for some compliance frameworks (PCI DSS, SOC 2) but exceeds MVP scope. Deferred until compliance certifications are pursued.

**Implementation Notes**:
- Terraform configuration: `aws_cloudwatch_log_group.security_events` with `retention_in_days = 30`
- Export logs to S3 for long-term archival if compliance requirements change post-MVP

---

## Decision 12: Secret Rotation Frequency — Manual On-Demand (MVP)

**Decision**: JWT signing keys, database passwords, and API keys are rotated manually on-demand using AWS Secrets Manager; no automated rotation schedule for MVP.

**Rationale**:
- Multi-key JWT rotation strategy (Decision 2) enables zero-downtime key rotation when needed
- Manual rotation is sufficient for MVP threat model (staging only, no production traffic yet)
- Automated rotation adds operational complexity (rotation lambdas, alerting, rollback procedures) that exceeds MVP scope
- FR-042 requires rotation support without redeployment (✓ achieved via runtime secret fetch), not automated scheduling

**Alternatives Considered**:
- **30-day automated rotation**: Industry best practice but requires Lambda rotation functions, testing rotation rollback, and operational runbooks. Deferred for production with dedicated security operations team.
- **90-day automated rotation**: Longer interval reduces rotation risk but still requires automation infrastructure. Deferred for same reasons.

**Implementation Notes**:
- Document manual rotation procedure in operations runbook (add new key to Secrets Manager, mark old key as retired after 30-day grace period, delete retired key)
- Backend fetches active keys from Secrets Manager on startup and refreshes every 5 minutes
- Monitor `SecretRotation` CloudWatch metric for manual rotation events

---

## Summary of Research Findings

| Decision | Technology/Pattern | Rationale |
|----------|-------------------|-----------|
| JWT Algorithm | RS256 (asymmetric) | Public key validation, distributed architecture, industry standard |
| Key Rotation | Multi-key validation | Zero-downtime rotation without mass logout |
| Token Storage | HTTP-only cookies (Secure, SameSite=Strict) | XSS prevention, CSRF prevention |
| Token Lifetime | 24h access, 30d refresh | Security/UX balance |
| Password Hashing | bcrypt cost factor 12 | OWASP 2026 recommendation, <200ms hashing time |
| Optimistic Locking | Version numbers (integer) | Prevents lost updates, atomic DB increment |
| Password Change | User choice to invalidate all sessions | Security/UX balance, user education |
| Prompt Injection | Pattern matching + keyword filtering | Deterministic, low latency, known attack vectors |
| Output Sanitization | HTML stripping + React safe rendering | Defense-in-depth, XSS prevention |
| Alarm Response | Manual review (no automated blocking) | Avoid false positives, MVP availability priority |
| Log Retention | 30 days CloudWatch | Cost-optimized, sufficient for incident response |
| Secret Rotation | Manual on-demand | MVP scope, zero-downtime capable via multi-key |

**All research findings are directly traceable to functional requirements and clarification decisions. No speculative or "future-proofing" architectures included.**
