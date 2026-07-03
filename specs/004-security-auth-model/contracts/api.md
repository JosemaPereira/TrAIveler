# API Contract: Security & Authentication/Authorization

**Feature**: 004-security-auth-model | **Date**: 2026-07-03

**Purpose**: Define RESTful API endpoints for authentication, authorization, and security operations.

---

## Base URL

- **Staging**: `https://api-staging.traivelr.example.com`
- **Production**: `https://api.traivelr.example.com`

**Protocol**: HTTPS only (TLS 1.2+)

---

## Authentication Endpoints

### POST /auth/register

Register a new user account.

**Authentication**: None (public endpoint)

**Request Headers**:
```
Content-Type: application/json
```

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "role": "admin"
}
```

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `email` | string | Yes | RFC 5322 email format; unique |
| `password` | string | Yes | 8-72 chars; ≥1 uppercase, ≥1 lowercase, ≥1 digit |
| `role` | string | No | 'admin' or 'partner'; defaults to 'admin' |

**Success Response** (201 Created):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "role": "admin",
  "created_at": "2026-07-03T10:00:00Z"
}
```

**Error Responses**:

| Status | Code | Message | Details |
|--------|------|---------|---------|
| 400 | `invalid_email` | Email format is invalid | `{"field": "email", "value": "not-an-email"}` |
| 400 | `weak_password` | Password does not meet requirements | `{"missing": ["uppercase", "digit"]}` |
| 409 | `email_exists` | Email already registered | `{"email": "user@example.com"}` |
| 500 | `internal_error` | Internal server error | `{"correlation_id": "..."}` |

**Side Effects**:
- Creates `users` table row with bcrypt-hashed password (cost factor 12)
- Logs `auth_register_success` security event (severity: info)

---

### POST /auth/login

Authenticate user and issue JWT tokens.

**Authentication**: None (public endpoint)

**Request Headers**:
```
Content-Type: application/json
```

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `email` | string | Yes | Valid email |
| `password` | string | Yes | Non-empty |

**Success Response** (200 OK):
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "role": "admin",
    "last_login_at": "2026-07-03T10:00:00Z"
  }
}
```

**Response Cookies** (Set-Cookie headers):
```
access_token=<jwt>; HttpOnly; Secure; SameSite=Strict; Max-Age=86400; Path=/
refresh_token=<jwt>; HttpOnly; Secure; SameSite=Strict; Max-Age=2592000; Path=/auth/refresh
```

**Error Responses**:

| Status | Code | Message | Details |
|--------|------|---------|---------|
| 400 | `missing_credentials` | Email and password required | `{"missing": ["email"]}` |
| 401 | `invalid_credentials` | Invalid email or password | `{}` (no details to prevent email enumeration) |
| 429 | `rate_limit_exceeded` | Too many login attempts | `{"retry_after": 60}` (if rate limiting implemented) |
| 500 | `internal_error` | Internal server error | `{"correlation_id": "..."}` |

**Side Effects**:
- Updates `users.last_login_at` timestamp
- Creates `refresh_tokens` table row (hashed token, 30-day expiration)
- Logs `auth_login_success` or `auth_login_failure` security event

---

### POST /auth/refresh

Issue a new access token using a valid refresh token.

**Authentication**: Refresh token (from cookie)

**Request Headers**:
```
Cookie: refresh_token=<jwt>
```

**Request Body**: None

**Success Response** (200 OK):
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "role": "admin"
  }
}
```

**Response Cookies**:
```
access_token=<new_jwt>; HttpOnly; Secure; SameSite=Strict; Max-Age=86400; Path=/
```

**Error Responses**:

| Status | Code | Message | Details |
|--------|------|---------|---------|
| 401 | `missing_token` | Refresh token required | `{}` |
| 401 | `invalid_token` | Refresh token is invalid or expired | `{}` |
| 401 | `revoked_token` | Refresh token has been revoked | `{}` |
| 500 | `internal_error` | Internal server error | `{"correlation_id": "..."}` |

**Side Effects**:
- Logs `auth_token_refresh` security event (severity: info)

---

### POST /auth/logout

Invalidate the current refresh token and clear cookies.

**Authentication**: Access token (from cookie)

**Request Headers**:
```
Cookie: access_token=<jwt>; refresh_token=<jwt>
```

**Request Body**: None

**Success Response** (204 No Content):
- Empty body
- Clears `access_token` and `refresh_token` cookies (`Set-Cookie` with `Max-Age=0`)

**Error Responses**:

| Status | Code | Message | Details |
|--------|------|---------|---------|
| 401 | `unauthorized` | Authentication required | `{}` |
| 500 | `internal_error` | Internal server error | `{"correlation_id": "..."}` |

**Side Effects**:
- Updates `refresh_tokens.revoked_at` for current refresh token
- Logs `auth_logout` security event (severity: info)

---

### PUT /auth/password

Change the authenticated user's password.

**Authentication**: Access token (from cookie)

**Request Headers**:
```
Content-Type: application/json
Cookie: access_token=<jwt>
```

**Request Body**:
```json
{
  "current_password": "OldPass123!",
  "new_password": "NewPass456!",
  "invalidate_all_sessions": false
}
```

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `current_password` | string | Yes | Must match stored password hash |
| `new_password` | string | Yes | 8-72 chars; ≥1 uppercase, ≥1 lowercase, ≥1 digit |
| `invalidate_all_sessions` | boolean | No | Default: false; if true, revokes all refresh tokens |

**Success Response** (200 OK):
```json
{
  "message": "Password changed successfully",
  "sessions_invalidated": false
}
```

**Error Responses**:

| Status | Code | Message | Details |
|--------|------|---------|---------|
| 400 | `weak_password` | New password does not meet requirements | `{"missing": ["digit"]}` |
| 401 | `unauthorized` | Authentication required | `{}` |
| 401 | `incorrect_password` | Current password is incorrect | `{}` |
| 500 | `internal_error` | Internal server error | `{"correlation_id": "..."}` |

**Side Effects**:
- Updates `users.password_hash` with bcrypt hash of new password
- If `invalidate_all_sessions=true`, updates `refresh_tokens.revoked_at` for all user's tokens
- Logs `auth_password_change` security event with `invalidate_all_sessions` flag

---

### DELETE /auth/account

Delete the authenticated user's account and all associated data.

**Authentication**: Access token (from cookie)

**Request Headers**:
```
Cookie: access_token=<jwt>
```

**Request Body**: None

**Success Response** (202 Accepted):
```json
{
  "message": "Account deletion initiated",
  "pii_removal_completion_date": "2026-08-02T10:00:00Z"
}
```

**Error Responses**:

| Status | Code | Message | Details |
|--------|------|---------|---------|
| 401 | `unauthorized` | Authentication required | `{}` |
| 500 | `internal_error` | Internal server error | `{"correlation_id": "..."}` |

**Side Effects**:
- Immediately revokes all user's refresh tokens (`revoked_at = NOW()`)
- Marks user account for deletion (soft delete); PII removed within 30 days
- Anonymizes user ID in shared trips (replaces with "deleted_user")
- Logs `auth_account_deletion` security event (severity: info)

---

## Authorization Middleware

### Authentication Check (All Protected Endpoints)

**Middleware**: `AuthenticationMiddleware`

**Applied To**: All endpoints except `/auth/register`, `/auth/login`, public static assets

**Behavior**:
1. Extract `access_token` from HTTP-only cookie
2. Validate JWT signature against all active signing keys (from `jwt_signing_keys` table)
3. Verify JWT expiration (`exp` claim ≤ NOW())
4. Extract `user_id` and `role` from JWT claims
5. Attach user context to request for downstream handlers

**Error Response** (401 Unauthorized):
```json
{
  "error": {
    "code": "unauthorized",
    "message": "Authentication required"
  }
}
```

**Logs**:
- If token missing: Logs `auth_missing_token` (severity: warning)
- If token invalid/expired: Logs `auth_invalid_token` (severity: warning)

---

### Role-Based Authorization (Admin-Only Endpoints)

**Middleware**: `RequireRole("admin")`

**Applied To**: Admin-only operations (trip CRUD, approval/rejection)

**Behavior**:
1. Check authenticated user's role from request context
2. If role != "admin", reject with 403 Forbidden

**Error Response** (403 Forbidden):
```json
{
  "error": {
    "code": "forbidden",
    "message": "Insufficient permissions",
    "details": {
      "required_role": "admin",
      "user_role": "partner"
    }
  }
}
```

**Logs**:
- Logs `authz_denied` security event (severity: warning) with user ID, role, operation, resource ID

---

## Input Validation Endpoints

### POST /validate/prompt

Validate user input destined for AI provider (internal endpoint, not exposed publicly).

**Authentication**: Service-to-service (internal call from AI integration layer)

**Request Headers**:
```
Content-Type: application/json
X-Correlation-ID: <uuid>
```

**Request Body**:
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "prompt": "Help me plan a 3-day trip to Paris"
}
```

**Success Response** (200 OK):
```json
{
  "valid": true
}
```

**Rejection Response** (400 Bad Request):
```json
{
  "valid": false,
  "error": {
    "code": "prompt_injection_detected",
    "message": "Prompt contains forbidden patterns",
    "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
    "details": {
      "pattern": "instruction_override",
      "matched": "ignore previous instructions"
    }
  }
}
```

**Validation Rules**:
- Reject instruction-override patterns: `/(ignore|disregard|forget).*(previous|prior|above|instructions)/i`
- Reject system prompt extraction: `/(print|show|reveal).*(instructions|system prompt|configuration)/i`
- Reject off-topic keywords: finance, stock, medical, diagnose, code, SQL, programming

**Side Effects**:
- If rejected, logs `validation_prompt_injection` security event (severity: error)

---

## Security Event Endpoints

### GET /admin/security-events

Retrieve security events for monitoring and incident investigation (admin-only).

**Authentication**: Access token (admin role required)

**Request Headers**:
```
Cookie: access_token=<jwt>
```

**Query Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `event_type` | string | No | Filter by event type (e.g., `auth_login_failure`) |
| `user_id` | UUID | No | Filter by user ID |
| `start_date` | ISO8601 | No | Start of time range (default: 24 hours ago) |
| `end_date` | ISO8601 | No | End of time range (default: now) |
| `limit` | integer | No | Max results (default: 100, max: 1000) |

**Success Response** (200 OK):
```json
{
  "events": [
    {
      "id": "...",
      "correlation_id": "...",
      "event_type": "auth_login_failure",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "severity": "warning",
      "ip_address": "203.0.113.42",
      "timestamp": "2026-07-03T10:00:00Z",
      "details": {
        "user_email": "user@example.com",
        "failure_reason": "invalid_password"
      }
    }
  ],
  "total": 42,
  "limit": 100
}
```

**Error Responses**:

| Status | Code | Message | Details |
|--------|------|---------|---------|
| 401 | `unauthorized` | Authentication required | `{}` |
| 403 | `forbidden` | Admin role required | `{"required_role": "admin"}` |
| 400 | `invalid_date_range` | start_date must be before end_date | `{}` |

---

## Optimistic Locking Headers

### Conditional Update Headers

All update operations (PUT, PATCH, DELETE) on versioned resources (trips, itinerary items) require version checking.

**Request Headers**:
```
If-Match: <version_number>
```

**Example**:
```
PUT /trips/550e8400-e29b-41d4-a716-446655440000
If-Match: 5
Content-Type: application/json

{
  "name": "Paris Trip Updated"
}
```

**Success Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Paris Trip Updated",
  "version": 6
}
```

**Conflict Response** (409 Conflict):
```json
{
  "error": {
    "code": "version_conflict",
    "message": "Resource was modified by another request",
    "current": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Paris Trip Different",
      "version": 6
    }
  }
}
```

**Side Effects**:
- Logs `concurrency_conflict` security event (severity: warning) with resource type, ID, and attempted version

---

## Error Response Format

All error responses follow a consistent JSON structure:

```json
{
  "error": {
    "code": "error_code_snake_case",
    "message": "Human-readable error description",
    "details": {
      "field": "value"
    },
    "correlation_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Error codes never expose sensitive information** (e.g., "email not found" vs "invalid credentials" to prevent email enumeration).

---

## Security Headers

All API responses include security headers:

```
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'
```

---

## Rate Limiting

**MVP**: No rate limiting implemented. CloudWatch alarms monitor authentication failure rate (FR-053).

**Post-MVP**: Rate limiting headers will follow RFC 6585:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1625097600
```

---

## CORS Configuration

**Allowed Origins** (staging):
- `https://app-staging.traivelr.example.com`

**Allowed Origins** (production):
- `https://app.traivelr.example.com`

**Allowed Methods**: `GET, POST, PUT, DELETE, OPTIONS`

**Allowed Headers**: `Content-Type, Authorization, X-Correlation-ID`

**Allow Credentials**: `true` (required for HTTP-only cookies)

**Max Age**: `86400` (24 hours)

---

## API Versioning

**Strategy**: No versioning for MVP (single version). Future versions will use URL path versioning (`/v2/auth/login`).

**Breaking Changes**: Will be communicated 30 days in advance with deprecation warnings in response headers:
```
Deprecation: Sun, 01 Aug 2026 00:00:00 GMT
Sunset: Sun, 01 Sep 2026 00:00:00 GMT
```

---

## Testing Requirements

### Integration Tests (Required)

- [ ] `POST /auth/register` with valid data → 201 Created
- [ ] `POST /auth/register` with existing email → 409 Conflict
- [ ] `POST /auth/login` with valid credentials → 200 OK + cookies
- [ ] `POST /auth/login` with invalid password → 401 Unauthorized
- [ ] `POST /auth/refresh` with valid refresh token → 200 OK + new access token
- [ ] `POST /auth/refresh` with expired token → 401 Unauthorized
- [ ] `POST /auth/logout` with valid token → 204 No Content + cleared cookies
- [ ] `PUT /auth/password` with `invalidate_all_sessions=true` → all refresh tokens revoked
- [ ] `DELETE /auth/account` → user marked for deletion, tokens revoked
- [ ] `GET /trips/:id` without authentication → 401 Unauthorized
- [ ] `PUT /trips/:id` as partner role → 403 Forbidden (admin-only operation)
- [ ] `PUT /trips/:id` with wrong version → 409 Conflict with current version
- [ ] `POST /validate/prompt` with prompt injection → 400 Bad Request

### Security Tests (Required)

- [ ] SQL injection payloads in `/auth/login` email field → rejected
- [ ] XSS payloads in trip name field → sanitized before storage
- [ ] Prompt injection patterns in AI input → blocked by validation layer
- [ ] Expired JWT token → 401 Unauthorized
- [ ] Tampered JWT token (modified signature) → 401 Unauthorized
- [ ] Missing `If-Match` header on update → 400 Bad Request
- [ ] CORS preflight from unauthorized origin → rejected

---

## Observability

### Metrics (CloudWatch)

- `auth.login.success` (count) — Successful logins
- `auth.login.failure` (count) — Failed login attempts
- `auth.token.refresh` (count) — Token refresh requests
- `authz.denied` (count) — Authorization denials (403)
- `validation.prompt_injection` (count) — Prompt injection detections
- `concurrency.conflict` (count) — Optimistic locking conflicts (409)
- `api.error_rate` (count) — 5xx responses

### Logs (CloudWatch Logs)

All endpoints log structured JSON with:
- `correlation_id` (UUID, propagated from `X-Correlation-ID` header or generated)
- `method` (HTTP method)
- `path` (request path)
- `status` (response status code)
- `duration_ms` (request processing time)
- `user_id` (if authenticated)
- `error` (if status ≥400)

**30-day retention policy** (FR-051a).

---

## Contract Changelog

| Date | Version | Change |
|------|---------|--------|
| 2026-07-03 | 1.0 | Initial security & authentication API contract |
