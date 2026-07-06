# Research: Authentication & Collaboration User Experience

**Feature**: Authentication & Collaboration UX  
**Date**: 2026-07-06  
**Phase**: 0 (Research & Technology Decisions)

## Overview

This document captures technology decisions for implementing authentication, session management, rate limiting, subscription lifecycle, and collaboration workflows supporting Paid User (subscription-based) and Free User (no-payment collaborator) tiers.

## Technology Decisions

### Decision 1: JWT Token Storage — HTTP-only Cookies

**Problem**: Where and how to store JWT tokens on the client to balance security with usability?

**Options Evaluated**:
1. **localStorage**: Simple client-side access, but vulnerable to XSS attacks (any injected script can read tokens)
2. **sessionStorage**: Same as localStorage but cleared on tab close; still XSS-vulnerable
3. **HTTP-only Secure SameSite=Strict cookies**: Server-controlled, inaccessible to JavaScript, immune to XSS token theft
4. **In-memory only (no persistence)**: Most secure but poor UX (user logged out on page refresh)

**Decision**: HTTP-only Secure SameSite=Strict cookies

**Rationale**:
- XSS protection: Tokens inaccessible to JavaScript (even if XSS vulnerability exists elsewhere)
- CSRF protection: SameSite=Strict prevents cross-site cookie sending
- Secure flag: HTTPS-only transmission
- Automatic inclusion: Browser sends cookie with every request, no manual header management
- Aligns with constitution security requirements (V. Secure Configuration)

**Alternatives Rejected**:
- localStorage/sessionStorage: High XSS risk outweighs simplicity
- In-memory only: Poor UX (refresh = logout) for minimal security gain

**Implementation Notes**:
- Access token: 24-hour expiration, HTTP-only, Secure, SameSite=Strict
- Refresh token: 30-day expiration, HTTP-only, Secure, SameSite=Strict, separate cookie
- Backend sets cookies in `Set-Cookie` response headers
- Frontend axios instance configured with `withCredentials: true`

---

### Decision 2: JWT Algorithm — RS256 (Asymmetric)

**Problem**: Which JWT signing algorithm to use for access and refresh tokens?

**Options Evaluated**:
1. **HS256** (HMAC with SHA-256): Symmetric key, simple, single secret shared across services
2. **RS256** (RSA Signature with SHA-256): Asymmetric key pair, public key for validation, private key for signing
3. **ES256** (ECDSA with SHA-256): Asymmetric, smaller keys/signatures than RSA, but less widely supported

**Decision**: RS256 (RSA Signature with SHA-256)

**Rationale**:
- Key distribution: Public key can be safely shared with validation services without exposing signing capability
- Zero-downtime rotation: Multiple active public keys supported simultaneously (spec 004 requirement)
- Service separation: Frontend/validation services only need public key, signing isolated to auth service
- Industry standard: Widely used by Auth0, AWS Cognito, Okta
- Aligns with docs/security.md JWT Key Rotation strategy

**Alternatives Rejected**:
- HS256: Symmetric key means any service that validates must also be trusted to sign (poor security boundary)
- ES256: Smaller payload benefit minimal for web use case, less library/tooling support

**Implementation Notes**:
- RSA 2048-bit key pairs generated and stored in AWS Secrets Manager
- JWTSigningKey table tracks multiple active keys (key_id, public_key PEM, private_key_secret_arn, status)
- New tokens signed with current primary key (configured in app)
- Validation accepts any active key (enables rolling key rotation)
- Key rotation: Generate new key → store in Secrets Manager → insert DB row → update app config → old key remains active for grace period

---

### Decision 3: Password Hashing — bcrypt (cost 12)

**Problem**: How to securely hash passwords with appropriate computational cost?

**Options Evaluated**:
1. **bcrypt (cost 10-12)**: Battle-tested, adaptive work factor, designed for passwords
2. **argon2id**: Modern, memory-hard, OWASP recommendation, but less mature Go libraries
3. **PBKDF2**: Standards-compliant, but requires more iterations than bcrypt for equivalent security
4. **scrypt**: Memory-hard like argon2, but complex tuning parameters

**Decision**: bcrypt with cost factor 12

**Rationale**:
- Proven track record: 20+ years in production use, no significant vulnerabilities
- Adaptive cost: Work factor can be increased as hardware improves (currently cost 12 = 2^12 = 4096 iterations)
- Mature Go library: `golang.org/x/crypto/bcrypt` is well-maintained, part of Go extended standard library
- Cost 12 balance: ~250ms hashing time on modern hardware (acceptable for login, sufficient security)
- Aligns with docs/security.md password security requirements

**Alternatives Rejected**:
- argon2id: Technically superior but Go library less mature, requires more tuning (memory, parallelism parameters)
- PBKDF2: Requires 100k+ iterations for bcrypt-equivalent security, slower
- scrypt: Complex parameter tuning (N, r, p values) increases misconfiguration risk

**Implementation Notes**:
- `bcrypt.GenerateFromPassword(password, 12)` for registration/password change
- `bcrypt.CompareHashAndPassword(hash, password)` for login validation
- Never log passwords or hashes
- Password complexity: 8-72 chars, uppercase + lowercase + digit (validated before hashing)

---

### Decision 4: Rate Limiting Strategy — Progressive Delay with Account-Level Tracking

**Problem**: How to prevent brute-force login attacks without permanently locking legitimate users?

**Options Evaluated**:
1. **IP-based rate limiting only**: Simple (e.g., 10 attempts/min per IP), but vulnerable to distributed attacks and can block shared IPs
2. **Progressive delay with account-level tracking**: Track failed attempts per email, exponentially increase delay after threshold
3. **Account lockout after threshold**: Lock account for fixed duration (e.g., 30 min) after N failures; requires password reset to unlock
4. **CAPTCHA after threshold**: Show CAPTCHA after N failures; adds friction, requires third-party integration

**Decision**: Progressive delay with account-level tracking

**Rationale**:
- Email-based tracking: Prevents single-IP bypass (distributed attack) and doesn't penalize shared IPs (corporate networks)
- No permanent lockout: Legitimate users not locked out by typos or attacker-triggered lockouts
- Exponential backoff: After 5 failed attempts within 15 minutes → 1s, 2s, 4s, 8s, 16s delays make brute-force infeasible
- Simple UX: No CAPTCHA friction for legitimate users, clear feedback ("Too many attempts, try again in X seconds")
- Industry standard: Used by Auth0, AWS Cognito, GitHub
- Aligns with spec clarification decision (session 2026-07-06, Question 1)

**Alternatives Rejected**:
- IP-only limiting: Shared IPs (offices, VPNs) unfairly penalized, distributed attacks bypass
- Account lockout: Poor UX (legitimate user locked out), can be weaponized (attacker locks victim's account)
- CAPTCHA: Adds friction to every user after threshold, requires third-party dependency

**Implementation Notes**:
- Store failed attempt counts in-memory cache (Redis for multi-instance) keyed by email address
- Track: `{ email: string, count: int, first_attempt_at: timestamp }`
- On login failure: increment count, if count >= 5 within 15 min window → calculate delay `2^(count-5)` seconds
- On login success: reset count to 0
- Window resets after 15 minutes of no attempts

---

### Decision 5: Session Renewal — Automatic with Refresh Token Exchange

**Problem**: How to keep users logged in without forcing re-authentication when access tokens expire?

**Options Evaluated**:
1. **Manual re-login**: User must log in again when access token expires (poor UX)
2. **Long-lived access tokens**: 30-day expiration (reduces security: stolen token valid for weeks)
3. **Automatic refresh with dedicated refresh token**: Short-lived access token (24h) + long-lived refresh token (30d), auto-exchange on expiration
4. **Sliding session**: Extend access token expiration on every request (complex invalidation logic)

**Decision**: Automatic refresh with dedicated refresh token

**Rationale**:
- Security + UX balance: Access token expires in 24h (limits stolen token window), refresh token allows seamless renewal for 30 days
- Explicit revocation: Refresh tokens can be revoked (logout, password change) independently of access tokens
- Industry standard: OAuth2 pattern used by Google, GitHub, Auth0
- Aligns with spec FR-015 (detect expired tokens and auto-renew)

**Alternatives Rejected**:
- Manual re-login: Unacceptable UX (user interrupted every 24h)
- Long-lived access tokens: Security risk (stolen token valid for weeks, no revocation)
- Sliding session: Complex to implement correctly, difficult to reason about expiration

**Implementation Notes**:
- Access token: 24h expiration, contains user_id + has_subscription + iat/exp claims
- Refresh token: 30d expiration, stored in RefreshToken table (token_hash SHA-256, expires_at, revoked_at)
- Frontend detects 401 Unauthorized → calls `/auth/refresh` with refresh token cookie → receives new access token → retries original request
- Refresh token rotation: On exchange, optionally issue new refresh token and revoke old one (reduces replay window)

---

### Decision 6: Password Reset Token — Single-Use with Timestamp Marking

**Problem**: How to enforce that password reset tokens can only be used once to prevent replay attacks?

**Options Evaluated**:
1. **Delete token immediately after use**: Remove database row on redemption (loses audit trail)
2. **Mark token as used with timestamp**: Add `used_at` timestamp field, set on redemption, validation checks `used_at IS NULL`
3. **Time-based expiration only**: Rely on 1-hour expiration, no reuse prevention (vulnerable to replay if link intercepted)
4. **One-time code (6-digit numeric)**: Email numeric code, require manual entry (better security vs link interception, worse UX)

**Decision**: Mark token as used with timestamp

**Rationale**:
- Audit trail: Preserves when token was redeemed and by whom (IP address in security log)
- Single-use enforcement: Validation requires both `expires_at > now` AND `used_at IS NULL`
- Simple logic: One additional timestamp field, clear state machine (created → used or expired)
- Aligns with spec clarification (session 2026-07-06, Question 3)

**Alternatives Rejected**:
- Delete immediately: Loses audit trail of password resets (important security event)
- Expiration only: Vulnerable to replay attack within 1-hour window
- One-time code: Better security but worse UX (copy-paste friction, typo errors)

**Implementation Notes**:
- PasswordResetToken table: `id, user_id, token_hash (SHA-256), expires_at (now + 1h), used_at (NULL), created_at`
- Token generation: 32 bytes cryptographically random, hashed before storage
- Validation: `SELECT ... WHERE token_hash = ? AND expires_at > NOW() AND used_at IS NULL`
- On successful password reset: `UPDATE ... SET used_at = NOW() WHERE id = ?`

---

### Decision 7: Subscription Grace Period — 30 Days with Read-Only Access

**Problem**: What happens when a Paid User's subscription expires? Immediate access loss vs graceful degradation?

**Options Evaluated**:
1. **Immediate revocation**: Set `has_subscription=false`, trips become read-only immediately, harsh but simple
2. **Grace period with read-only access**: 30-day period where owned trips visible but read-only, collaborate as Free User, then archive
3. **Convert to Free User + archive owned trips immediately**: Instant archival, user continues as Free User (1 collaboration)
4. **Lock account until resubscription**: No access at all until payment (poor UX, no win-back opportunity)

**Decision**: Grace period with read-only access (30 days)

**Rationale**:
- Customer retention: Grace period provides window for win-back campaigns, users see their data before archival
- User expectation: Aligns with SaaS standard behavior (Spotify, Netflix, Dropbox show data before removal)
- Data protection: Trips not immediately deleted, easy to restore on resubscription
- Collaboration continuity: User can still collaborate as Free User on other trips during grace period
- Aligns with spec clarification (session 2026-07-06, Question 3)

**Alternatives Rejected**:
- Immediate revocation: Harsh UX, no reactivation opportunity, feels punitive
- Immediate archival: Same as immediate revocation, no grace period benefit
- Account lockout: Prevents any app usage, loses opportunity for organic reactivation (user sees value, resubscribes)

**Implementation Notes**:
- Subscription cancellation: Set `Subscription.status='cancelled'`, `grace_period_ends_at = NOW() + INTERVAL '30 days'`, `User.has_subscription=false`
- During grace period: Owned trips visible but API returns 403 Forbidden on PUT/DELETE with "Your subscription has expired. Renew to continue editing."
- Banner displayed: "Your subscription has expired. Renew to continue editing your trips." with "Renew Subscription" button
- After grace period: Scheduled job sets `Trip.archived=true` for all owned trips where `creator_id = user AND Subscription.grace_period_ends_at < NOW()`
- On resubscription: Set `User.has_subscription=true`, `Subscription.status='active'`, `grace_period_ends_at=NULL`, `Trip.archived=false` for all owned trips

---

### Decision 8: Free User Collaboration Limit — Single Trip with Leave-to-Accept Pattern

**Problem**: How to limit Free User collaboration to encourage upgrades without being overly restrictive?

**Options Evaluated**:
1. **No collaboration without payment**: Free Users cannot collaborate at all (blocks viral growth)
2. **Unlimited collaboration**: Free Users can join unlimited trips (no upgrade incentive)
3. **Single trip with leave-to-accept**: Free User can collaborate on 1 trip at a time, must leave current to accept new invitation
4. **Time-limited collaboration**: Free User gets 7-day collaboration access, then must upgrade (poor UX mid-trip)

**Decision**: Single trip with leave-to-accept pattern

**Rationale**:
- Viral growth enabler: Free Users can experience collaboration without payment barrier (invitations drive signups)
- Clear upgrade incentive: Multi-trip collaboration requires paid subscription, obvious value proposition
- Simple enforcement: Single database check (`SELECT COUNT(*) FROM collaborators WHERE user_id = ? AND status = 'active'`)
- Predictable UX: User knows they can only be on one trip, explicit "Leave Trip" action to accept new invitation
- Aligns with spec clarification (session 2026-07-06, Question 2)

**Alternatives Rejected**:
- No collaboration: Blocks viral growth, removes key differentiator from competitors
- Unlimited collaboration: No upgrade incentive, monetization failure
- Time-limited: Poor UX (collaboration interrupted mid-trip), confusing expiration tracking

**Implementation Notes**:
- Before accepting invitation: Check `SELECT COUNT(*) FROM collaborators WHERE user_id = ? AND status != 'rejected'`
- If count >= 1: Display "You are already collaborating on [Trip Name]. Leave that trip first to accept this invitation." with "Leave Current Trip" button
- Leave action: `DELETE FROM collaborators WHERE user_id = ? AND trip_id = ?`, display confirmation "You have left [Trip Name]. You can now accept new invitations."
- Paid Users: No collaboration limit check (can join unlimited trips)

---

### Decision 9: Free User Upgrade Flow — Preserve Existing Collaboration

**Problem**: When a Free User upgrades to Paid User mid-collaboration, what happens to their current collaboration?

**Options Evaluated**:
1. **Terminate collaboration on upgrade**: Auto-remove Collaborator record, user loses access (poor UX)
2. **Preserve existing collaboration**: Collaborator record unchanged, user gains trip creation capability, collaboration continues
3. **Promote to co-owner**: Upgrading Free User becomes co-owner with full edit rights (complex ownership model)
4. **Require manual decision**: Force user to choose "Keep collaboration or leave" (adds friction to upgrade flow)

**Decision**: Preserve existing collaboration

**Rationale**:
- Smooth upgrade experience: No disruption to ongoing trip planning, user simply gains additional capabilities
- Clear benefit: User can now create own trips WHILE still collaborating on original trip
- Simple implementation: No Collaborator record changes, just set `User.has_subscription=true` and remove one-trip limit
- Encourages upgrades: User sees immediate value (can create trips) without losing current access
- Aligns with spec clarification (session 2026-07-06, Question 4)

**Alternatives Rejected**:
- Terminate collaboration: Poor UX, punishes user for upgrading
- Promote to co-owner: Complex shared ownership model (who deletes? who approves other suggestions?)
- Manual decision: Adds friction to upgrade flow, complicates conversion funnel

**Implementation Notes**:
- On subscription purchase completion: Set `User.has_subscription=true`, `Subscription.status='active'`
- Leave Collaborator records unchanged (collaboration continues)
- Update UI badge from "Free User" to "Paid User" on shared trip
- Remove one-collaboration limit checks for user (can now accept multiple invitations)
- Enable "Create Trip" button on dashboard

---

### Decision 10: Security Event Logging — Structured JSON to CloudWatch

**Problem**: What authentication/authorization events should be logged, and in what format?

**Options Evaluated**:
1. **Minimal logging**: Only successful logins and explicit logouts (insufficient for security monitoring)
2. **Structured security events with correlation IDs**: Log all auth events (login, registration, password reset, token refresh, logout) with structured fields
3. **Detailed audit with request payloads**: Log full request/response bodies (comprehensive but risks logging sensitive data)
4. **Success-only logging**: Only log successful events to reduce volume (insufficient for brute-force detection)

**Decision**: Structured security events with correlation IDs

**Rationale**:
- Incident investigation: Failed login patterns identify brute-force attacks, password reset frequency detects account takeover attempts
- Correlation: Request ID traces user journey across multiple requests, enables debugging complex authentication flows
- Compliance: Audit trail for security events supports GDPR-aware design (user can request access log)
- No sensitive data: Excludes passwords, tokens, PII beyond user ID (aligns with constitution V. Secure Configuration)
- CloudWatch integration: Structured JSON enables CloudWatch Insights queries, alarms on anomalous patterns
- Aligns with spec clarification (session 2026-07-06, Question 2) and docs/security.md logging requirements

**Alternatives Rejected**:
- Minimal logging: Cannot detect brute-force, cannot debug failed authentication issues
- Request payload logging: Risks logging passwords, tokens, or other sensitive data
- Success-only: Cannot detect attacks (only see successful breaches, not attempts)

**Implementation Notes**:
- SecurityEvent table/log entry: `{correlation_id, event_type, user_id (or email for failures), severity, ip_address, user_agent, timestamp, details (JSONB)}`
- Event types: `auth_login_success`, `auth_login_failure`, `auth_registration`, `auth_password_reset_request`, `auth_password_reset_complete`, `auth_password_change`, `auth_token_refresh`, `auth_logout`
- CloudWatch Logs: Structured JSON, 30-day retention, alarms on: auth failures >100/min for 5 min, password reset rate >10/min
- Never log: passwords, password hashes, JWT tokens, refresh tokens, full request bodies

---

## Decision Summary Table

| Decision | Chosen Technology/Approach | Key Rationale |
|----------|---------------------------|---------------|
| JWT Storage | HTTP-only Secure SameSite=Strict cookies | XSS immunity, CSRF protection |
| JWT Algorithm | RS256 (asymmetric) | Public key distribution, zero-downtime rotation |
| Password Hashing | bcrypt cost 12 | Proven security, adaptive work factor, mature library |
| Rate Limiting | Progressive delay, account-level tracking | No permanent lockout, exponential backoff |
| Session Renewal | Auto-refresh with refresh token | Security + UX balance, explicit revocation |
| Password Reset | Single-use with used_at timestamp | Audit trail, replay prevention |
| Subscription Lapse | 30-day grace period, read-only access | Customer retention, data protection |
| Free User Limit | Single collaboration, leave-to-accept | Viral growth + upgrade incentive |
| Free User Upgrade | Preserve collaboration | Smooth UX, no disruption |
| Security Logging | Structured JSON to CloudWatch | Incident investigation, no sensitive data |

## Open Questions / Future Research

**None** - All technical decisions resolved for MVP scope. Post-MVP considerations:
- Redis for distributed rate limiting (if multi-instance scaling needed)
- Email service integration (replace stub email with SendGrid/SES)
- Real payment provider integration (replace stub with Stripe)
- MFA support (TOTP, WebAuthn)
- OAuth providers (Google, GitHub login)
