# API Contracts: Authentication & Collaboration UX

**Feature**: Authentication & Collaboration UX  
**Date**: 2026-07-06  
**Phase**: 1 (Design)

## Overview

This document defines the REST API endpoint contracts for authentication, session management, subscription lifecycle, and collaboration workflows. All endpoints follow the standards in docs/api-design-standards.md.

**Base URL**: `/api/v1`  
**Content-Type**: `application/json`  
**Authentication**: JWT token in HTTP-only cookie (except registration, login, password-reset endpoints)

---

## Authentication Endpoints

### POST /auth/register

Register a new user account (Paid User or Free User). Paid Users optionally provide payment method for immediate subscription activation.

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecureP@ss123",
  "full_name": "Jane Doe",
  "payment_method_token": "tok_visa_test_123" // Optional, only for Paid User registration
}
```

**Request Validation**:
- `email`: Valid email format, max 255 chars, unique (409 if exists)
- `password`: 8-72 chars, at least one uppercase, one lowercase, one digit
- `full_name`: 1-100 chars, non-empty
- `payment_method_token`: Optional, if provided triggers subscription creation

**Success Response** (201 Created):
```json
{
  "user": {
    "id": "uuid-here",
    "email": "user@example.com",
    "full_name": "Jane Doe",
    "has_subscription": true,  // true if payment_method_token provided, false otherwise
    "created_at": "2026-07-06T10:00:00Z"
  },
  "subscription": {  // Only present if payment_method_token provided
    "id": "uuid-here",
    "plan_id": "uuid-here",
    "status": "active",
    "current_period_start": "2026-07-06T10:00:00Z",
    "current_period_end": "2026-08-06T10:00:00Z"
  }
}
```
**Cookies Set**:
- `access_token`: JWT token (24h expiration, HTTP-only, Secure, SameSite=Strict)
- `refresh_token`: JWT token (30d expiration, HTTP-only, Secure, SameSite=Strict)

**Error Responses**:
- `400 Bad Request`: Invalid request body (weak password, invalid email format)
  ```json
  {
    "error": "VALIDATION_ERROR",
    "message": "Password must contain at least one uppercase letter, one lowercase letter, and one digit",
    "request_id": "uuid-here",
    "fields": [{"field": "password", "message": "Invalid format"}]
  }
  ```
- `409 Conflict`: Email already registered
  ```json
  {
    "error": "EMAIL_EXISTS",
    "message": "Email already registered",
    "request_id": "uuid-here"
  }
  ```
- `503 Service Unavailable`: Database unavailable
  ```json
  {
    "error": "SERVICE_UNAVAILABLE",
    "message": "Service temporarily unavailable. Please try again later.",
    "request_id": "uuid-here",
    "retry_after": 30  // seconds
  }
  ```

**Rate Limit**: 10 requests/minute per IP (header: `X-RateLimit-Remaining: 9`)

**Security Event Logged**: `auth_registration` (severity: info, user_id: new user ID, ip_address, user_agent)

---

### POST /auth/login

Authenticate user with email and password. Returns JWT tokens in HTTP-only cookies.

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecureP@ss123"
}
```

**Request Validation**:
- `email`: Valid email format, non-empty
- `password`: Non-empty

**Success Response** (200 OK):
```json
{
  "user": {
    "id": "uuid-here",
    "email": "user@example.com",
    "full_name": "Jane Doe",
    "has_subscription": true,
    "created_at": "2026-07-06T10:00:00Z"
  }
}
```
**Cookies Set**:
- `access_token`: JWT token (24h expiration, HTTP-only, Secure, SameSite=Strict)
- `refresh_token`: JWT token (30d expiration, HTTP-only, Secure, SameSite=Strict)

**Error Responses**:
- `400 Bad Request`: Missing required fields
- `401 Unauthorized`: Invalid credentials (incorrect password or nonexistent user)
  ```json
  {
    "error": "INVALID_CREDENTIALS",
    "message": "Invalid email or password",
    "request_id": "uuid-here"
  }
  ```
- `429 Too Many Requests`: Rate limit exceeded (progressive delay after 5 failed attempts)
  ```json
  {
    "error": "RATE_LIMIT_EXCEEDED",
    "message": "Too many failed login attempts. Please try again in 8 seconds.",
    "request_id": "uuid-here",
    "retry_after": 8
  }
  ```

**Rate Limit**: Progressive delay after 5 failed attempts within 15 minutes (exponential backoff: 1s, 2s, 4s, 8s, 16s)

**Security Events Logged**:
- Success: `auth_login_success` (severity: info, user_id, ip_address, user_agent)
- Failure: `auth_login_failure` (severity: warning, email, ip_address, user_agent, details: {reason: "invalid_password" | "user_not_found"})

---

### POST /auth/refresh

Exchange refresh token for new access token (automatic session renewal). Frontend calls this when access token expires (detected via 401 response).

**Authentication**: Refresh token cookie required

**Request Body**: Empty

**Success Response** (200 OK):
```json
{
  "message": "Token refreshed successfully"
}
```
**Cookies Set**:
- `access_token`: New JWT token (24h expiration, HTTP-only, Secure, SameSite=Strict)
- `refresh_token`: Optionally rotated (new 30d token, old one revoked)

**Error Responses**:
- `401 Unauthorized`: Invalid or expired refresh token
  ```json
  {
    "error": "INVALID_REFRESH_TOKEN",
    "message": "Session expired. Please log in again.",
    "request_id": "uuid-here"
  }
  ```
- `503 Service Unavailable`: Database unavailable

**Rate Limit**: 100 requests/minute per user (standard authenticated limit)

**Security Event Logged**: `auth_token_refresh` (severity: info, user_id, correlation_id)

---

### POST /auth/logout

Logout user by revoking refresh token and clearing cookies.

**Authentication**: Access token cookie required

**Request Body**: Empty

**Success Response** (204 No Content)

**Cookies Cleared**:
- `access_token`: Deleted
- `refresh_token`: Deleted

**Error Responses**:
- `401 Unauthorized`: No valid access token (already logged out)

**Rate Limit**: 100 requests/minute per user

**Security Event Logged**: `auth_logout` (severity: info, user_id, correlation_id)

---

### POST /auth/password-reset-request

Request password reset email with single-use token link.

**Request Body**:
```json
{
  "email": "user@example.com"
}
```

**Request Validation**:
- `email`: Valid email format, non-empty

**Success Response** (200 OK):
```json
{
  "message": "If an account with that email exists, a password reset link has been sent."
}
```
**Note**: Always returns success (even if email doesn't exist) to prevent user enumeration.

**Side Effects**:
- If user exists: Generate reset token, store hash in `password_reset_tokens`, send email with link
- If user doesn't exist: No email sent, log security event for suspicious activity detection

**Error Responses**:
- `400 Bad Request`: Invalid email format
- `429 Too Many Requests`: Rate limit exceeded (10 requests/hour per IP)
  ```json
  {
    "error": "RATE_LIMIT_EXCEEDED",
    "message": "Too many password reset requests. Please try again later.",
    "request_id": "uuid-here",
    "retry_after": 3600
  }
  ```

**Rate Limit**: 10 requests/hour per IP (prevents abuse)

**Security Event Logged**: `auth_password_reset_request` (severity: info, email, ip_address, details: {user_exists: true/false})

---

### POST /auth/password-reset-complete

Complete password reset using token from email link.

**Request Body**:
```json
{
  "token": "base64url-encoded-token-here",
  "new_password": "NewSecureP@ss456"
}
```

**Request Validation**:
- `token`: Non-empty, 43 chars (32 bytes base64url-encoded)
- `new_password`: 8-72 chars, at least one uppercase, one lowercase, one digit

**Success Response** (200 OK):
```json
{
  "message": "Password reset successfully. You can now log in with your new password."
}
```

**Side Effects**:
- Update user password hash (bcrypt cost 12)
- Mark token as used (`used_at = NOW()`)
- Revoke all refresh tokens for user (force re-login on all devices)

**Error Responses**:
- `400 Bad Request`: Invalid request body (weak password, missing token)
- `400 Bad Request`: Invalid or expired token
  ```json
  {
    "error": "INVALID_TOKEN",
    "message": "This password reset link is invalid or has expired. Please request a new one.",
    "request_id": "uuid-here"
  }
  ```
- `400 Bad Request`: Token already used
  ```json
  {
    "error": "TOKEN_ALREADY_USED",
    "message": "This password reset link has already been used. Please request a new one if needed.",
    "request_id": "uuid-here"
  }
  ```

**Rate Limit**: 10 requests/hour per IP

**Security Event Logged**: `auth_password_reset_complete` (severity: info, user_id, ip_address)

---

### POST /auth/password-change

Change password for authenticated user (requires current password).

**Authentication**: Access token cookie required

**Request Body**:
```json
{
  "current_password": "SecureP@ss123",
  "new_password": "NewSecureP@ss456"
}
```

**Request Validation**:
- `current_password`: Non-empty
- `new_password`: 8-72 chars, at least one uppercase, one lowercase, one digit, must differ from current password

**Success Response** (200 OK):
```json
{
  "message": "Password changed successfully."
}
```

**Side Effects**:
- Update user password hash (bcrypt cost 12)
- Revoke all refresh tokens except current session (force re-login on other devices)

**Error Responses**:
- `400 Bad Request`: Weak new password or same as current password
- `401 Unauthorized`: Incorrect current password
  ```json
  {
    "error": "INVALID_PASSWORD",
    "message": "Current password is incorrect",
    "request_id": "uuid-here"
  }
  ```

**Rate Limit**: 100 requests/minute per user

**Security Event Logged**: `auth_password_change` (severity: info, user_id, correlation_id)

---

## Subscription Endpoints

### POST /subscriptions

Create new subscription (upgrade Free User to Paid User or resubscribe after cancellation).

**Authentication**: Access token cookie required

**Request Body**:
```json
{
  "plan_id": "uuid-here",
  "payment_method_token": "tok_visa_test_123"
}
```

**Request Validation**:
- `plan_id`: Valid UUID, must exist in plans table
- `payment_method_token`: Non-empty, valid payment token (validated by stub provider)

**Success Response** (201 Created):
```json
{
  "subscription": {
    "id": "uuid-here",
    "user_id": "uuid-here",
    "plan_id": "uuid-here",
    "status": "active",
    "current_period_start": "2026-07-06T10:00:00Z",
    "current_period_end": "2026-08-06T10:00:00Z",
    "grace_period_ends_at": null,
    "cancelled_at": null
  }
}
```

**Side Effects**:
- Set `User.has_subscription = true`
- If resubscribing after cancellation: Unarchive owned trips (`Trip.archived = false`)
- Clear grace period fields (`grace_period_ends_at = NULL`, `cancelled_at = NULL`)

**Error Responses**:
- `400 Bad Request`: Invalid plan_id or payment_method_token
- `409 Conflict`: User already has active subscription
  ```json
  {
    "error": "SUBSCRIPTION_EXISTS",
    "message": "You already have an active subscription",
    "request_id": "uuid-here"
  }
  ```
- `402 Payment Required`: Payment processing failed (stub always succeeds, real provider may fail)
  ```json
  {
    "error": "PAYMENT_FAILED",
    "message": "Payment processing failed. Please check your payment method.",
    "request_id": "uuid-here"
  }
  ```

**Rate Limit**: 100 requests/minute per user

---

### DELETE /subscriptions/:id

Cancel subscription (start 30-day grace period).

**Authentication**: Access token cookie required

**Path Parameters**:
- `id`: Subscription UUID

**Request Body**: Empty

**Success Response** (200 OK):
```json
{
  "subscription": {
    "id": "uuid-here",
    "status": "cancelled",
    "grace_period_ends_at": "2026-08-05T10:00:00Z",
    "cancelled_at": "2026-07-06T10:00:00Z"
  },
  "message": "Subscription cancelled. You will have read-only access to your trips until 2026-08-05. Renew anytime to restore full access."
}
```

**Side Effects**:
- Set `Subscription.status = 'cancelled'`, `cancelled_at = NOW()`, `grace_period_ends_at = NOW() + INTERVAL '30 days'`
- Set `User.has_subscription = false`
- During grace period: Owned trips become read-only (403 on PUT/DELETE)

**Error Responses**:
- `404 Not Found`: Subscription not found or not owned by user
- `409 Conflict`: Subscription already cancelled

**Rate Limit**: 100 requests/minute per user

---

### POST /subscriptions/:id/renew

Renew cancelled subscription during grace period.

**Authentication**: Access token cookie required

**Path Parameters**:
- `id`: Subscription UUID

**Request Body**: Empty (uses existing payment method)

**Success Response** (200 OK):
```json
{
  "subscription": {
    "id": "uuid-here",
    "status": "active",
    "grace_period_ends_at": null,
    "cancelled_at": null,
    "current_period_end": "2026-08-06T10:00:00Z"
  },
  "message": "Subscription renewed successfully."
}
```

**Side Effects**:
- Set `Subscription.status = 'active'`, `cancelled_at = NULL`, `grace_period_ends_at = NULL`
- Set `User.has_subscription = true`
- Unarchive owned trips (`Trip.archived = false`)

**Error Responses**:
- `404 Not Found`: Subscription not found or not owned by user
- `400 Bad Request`: Subscription not cancelled (only cancelled subscriptions can be renewed via this endpoint)

**Rate Limit**: 100 requests/minute per user

---

## Collaboration Endpoints

### POST /trips/:trip_id/collaborators

Invite collaborator to trip (email-based invitation).

**Authentication**: Access token cookie required

**Path Parameters**:
- `trip_id`: Trip UUID

**Request Body**:
```json
{
  "email": "collaborator@example.com"
}
```

**Request Validation**:
- `email`: Valid email format, non-empty
- User must be trip creator or existing collaborator with `can_invite=true` permission

**Success Response** (201 Created):
```json
{
  "collaborator": {
    "id": "uuid-here",
    "trip_id": "uuid-here",
    "user_id": "uuid-here-if-registered",  // null if invited user not registered yet
    "email": "collaborator@example.com",
    "status": "pending",
    "invited_at": "2026-07-06T10:00:00Z"
  },
  "message": "Invitation sent to collaborator@example.com"
}
```

**Side Effects**:
- If invited user exists: Create Collaborator record with `user_id` and `status='pending'`, send email notification
- If invited user doesn't exist: Create Collaborator record with `email` only and `status='pending'`, send signup invitation email

**Error Responses**:
- `400 Bad Request`: Invalid email format
- `403 Forbidden`: User not authorized to invite (not creator, not collaborator with invite permission)
- `404 Not Found`: Trip not found
- `409 Conflict`: User already invited or collaborating on this trip

**Rate Limit**: 100 requests/minute per user

---

### DELETE /trips/:trip_id/collaborators/leave

Leave trip as collaborator (for Free Users to accept new invitation).

**Authentication**: Access token cookie required

**Path Parameters**:
- `trip_id`: Trip UUID

**Request Body**: Empty

**Success Response** (204 No Content)

**Side Effects**:
- Delete Collaborator record for user on this trip
- For Free Users: Enables accepting new invitation (collaboration count decreases)

**Error Responses**:
- `404 Not Found`: Trip not found or user not a collaborator
- `403 Forbidden`: User is trip creator (cannot leave own trip)

**Rate Limit**: 100 requests/minute per user

---

### POST /trips/:trip_id/suggestions

Create suggestion for trip (collaborator proposes change).

**Authentication**: Access token cookie required

**Path Parameters**:
- `trip_id`: Trip UUID

**Request Body**:
```json
{
  "suggestion_type": "activity",  // or "destination", "accommodation", etc.
  "content": "Visit the Louvre Museum",
  "details": {
    "day_number": 2,
    "estimated_duration": "3 hours"
  }
}
```

**Request Validation**:
- `suggestion_type`: Non-empty, max 50 chars
- `content`: Non-empty, max 500 chars
- `details`: Optional JSONB object
- User must be collaborator on this trip (not creator)

**Success Response** (201 Created):
```json
{
  "suggestion": {
    "id": "uuid-here",
    "trip_id": "uuid-here",
    "collaborator_id": "uuid-here",
    "suggestion_type": "activity",
    "content": "Visit the Louvre Museum",
    "details": {"day_number": 2, "estimated_duration": "3 hours"},
    "status": "pending",
    "created_at": "2026-07-06T10:00:00Z"
  }
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body
- `403 Forbidden`: User is trip creator (creators edit directly, don't suggest) or not a collaborator
- `404 Not Found`: Trip not found

**Rate Limit**: 100 requests/minute per user

---

### PATCH /suggestions/:id/approve

Approve suggestion (trip creator only).

**Authentication**: Access token cookie required

**Path Parameters**:
- `id`: Suggestion UUID

**Request Body**: Empty

**Success Response** (200 OK):
```json
{
  "suggestion": {
    "id": "uuid-here",
    "status": "approved",
    "approved_at": "2026-07-06T10:05:00Z"
  },
  "message": "Suggestion approved. The trip has been updated."
}
```

**Side Effects**:
- Set `Suggestion.status = 'approved'`, `approved_at = NOW()`
- Apply suggestion to trip (implementation-specific: create Activity, update Destination, etc.)

**Error Responses**:
- `403 Forbidden`: User is not trip creator
- `404 Not Found`: Suggestion not found
- `409 Conflict`: Suggestion already approved or rejected

**Rate Limit**: 100 requests/minute per user

---

### PATCH /suggestions/:id/reject

Reject suggestion (trip creator only).

**Authentication**: Access token cookie required

**Path Parameters**:
- `id`: Suggestion UUID

**Request Body**: Empty

**Success Response** (200 OK):
```json
{
  "suggestion": {
    "id": "uuid-here",
    "status": "rejected",
    "rejected_at": "2026-07-06T10:05:00Z"
  },
  "message": "Suggestion rejected."
}
```

**Side Effects**:
- Set `Suggestion.status = 'rejected'`, `rejected_at = NOW()`

**Error Responses**:
- `403 Forbidden`: User is not trip creator
- `404 Not Found`: Suggestion not found
- `409 Conflict`: Suggestion already approved or rejected

**Rate Limit**: 100 requests/minute per user

---

## Common Response Headers

All responses include:
- `X-RateLimit-Limit`: Total requests allowed in window
- `X-RateLimit-Remaining`: Requests remaining in current window
- `X-RateLimit-Reset`: Unix timestamp when window resets
- `X-Request-ID`: Correlation ID for request tracing

---

## Error Response Format

All errors follow docs/api-design-standards.md format:

```json
{
  "error": "ERROR_CODE",
  "message": "Human-readable error message",
  "request_id": "uuid-here",
  "fields": [  // Optional, for validation errors
    {"field": "email", "message": "Invalid email format"}
  ]
}
```

---

## Authentication Flow Summary

```
1. Register: POST /auth/register → 201 + cookies
2. Login: POST /auth/login → 200 + cookies
3. Access API: Include cookies in all requests
4. Token expires: API returns 401 → Frontend calls POST /auth/refresh → New access token
5. Logout: POST /auth/logout → 204, cookies cleared
```

---

## Summary

**Total Endpoints**: 16
- Authentication: 7 endpoints (register, login, refresh, logout, password-reset-request, password-reset-complete, password-change)
- Subscription: 3 endpoints (create, cancel, renew)
- Collaboration: 6 endpoints (invite, leave, create-suggestion, approve-suggestion, reject-suggestion)

All endpoints use JWT tokens in HTTP-only cookies (except public endpoints: register, login, password-reset-request). All endpoints follow API design standards with consistent error format, rate limiting headers, and correlation IDs.
