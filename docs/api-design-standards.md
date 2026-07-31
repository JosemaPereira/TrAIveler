# API Design Standards

<!-- PROMOTED:api-standards START -->
<!-- Generated from specs/007-api-design-standards/spec.md and contracts/api-design-standards.md -->
<!-- Last promoted: 2026-07-10 (added machine-readable contract cross-reference from spec 009) -->

**Version**: 1.0.0  
**Effective Date**: 2026-07-06  
**Status**: Active  
**Applies To**: All backend HTTP APIs exposed to frontend clients, external partners, or third-party integrations

---

## Purpose

This document establishes mandatory conventions for API design across the TrAIveler project. All backend developers, code reviewers, and API consumers should reference this document to ensure consistency, predictability, and maintainability.

**Audience**:
- Backend developers designing and implementing new endpoints
- Frontend developers integrating with backend APIs
- Technical leads and code reviewers
- Future external API consumers

---

## Table of Contents

1. [Resource Naming Conventions](#1-resource-naming-conventions)
2. [URL Structure and Nesting](#2-url-structure-and-nesting)
3. [API Versioning](#3-api-versioning)
4. [HTTP Methods](#4-http-methods)
5. [Request Format](#5-request-format)
6. [Response Format](#6-response-format)
7. [Error Responses](#7-error-responses)
8. [HTTP Status Codes](#8-http-status-codes)
9. [Pagination](#9-pagination)
10. [Filtering](#10-filtering)
11. [Sorting](#11-sorting)
12. [Rate Limiting](#12-rate-limiting)
13. [Authentication Headers](#13-authentication-headers)
14. [Timestamps and Dates](#14-timestamps-and-dates)
15. [Endpoint Patterns](#15-endpoint-patterns)
16. [Machine-Readable Contract](#16-machine-readable-contract)

---

## 1. Resource Naming Conventions

### Rules

✅ **DO**:
- Use **plural nouns** for collection resources: `/trips`, `/users`, `/activities`
- Use **lowercase** with **hyphens** for compound words: `/travel-styles`, `/conversation-messages`
- Use **verbs** only for actions that don't map to CRUD: `/trips/:id/generate`, `/suggestions/:id/approve`
- Keep resource names **concise** but **descriptive**: `/destinations` (not `/dest` or `/destination_locations`)

❌ **DON'T**:
- Use singular nouns for collections: `/trip`, `/user`
- Use file extensions: `/trips.json`, `/users.xml`
- Use camelCase or snake_case: `/travelStyles`, `/conversation_messages`
- Mix singular and plural arbitrarily

### Examples

```http
✅ GET /api/v1/trips
✅ POST /api/v1/users
✅ GET /api/v1/travel-styles
✅ POST /api/v1/trips/:id/generate

❌ GET /api/v1/trip
❌ GET /api/v1/travelStyles
❌ GET /api/v1/trips.json
```

---

## 2. URL Structure and Nesting

### Rules

✅ **DO**:
- Nest resources **1 level deep** for parent-child relationships: `/trips/:id/days`
- Provide **top-level endpoints** for resources that can be accessed independently: `/activities/:id`
- Use **query parameters** for filtering across hierarchies: `/activities?trip_id=:id`
- Keep URLs **readable** and **predictable**

❌ **DON'T**:
- Nest more than 2 levels deep: `/trips/:id/days/:day_id/activities/:activity_id/edit`
- Use deeply nested URLs for frequently accessed resources

### Maximum Nesting Depth: 2 Levels

**Pattern**:
```
/resource/:id/sub-resource
/resource/:id/sub-resource/:sub_id
```

**Examples**:

```http
✅ GET /api/v1/trips/:id/days
✅ GET /api/v1/trips/:id/days/:day_id
✅ GET /api/v1/trips/:id/collaborators
✅ GET /api/v1/activities/:id (top-level for direct access)
✅ GET /api/v1/activities?trip_id=:id (filter across hierarchy)

❌ GET /api/v1/trips/:id/days/:day_id/activities/:activity_id
```

---

## 3. API Versioning

### Strategy: URL Path Versioning

All API endpoints **MUST** include a version prefix in the URL path.

**Format**: `/api/v{MAJOR}`

**Current Version**: `/api/v1`

### Rules

✅ **DO**:
- Include version in **every endpoint**: `/api/v1/trips`
- Use **major version** only in URL (no minor/patch)
- Bump major version for **breaking changes**:
  - Removing endpoints or fields
  - Changing field types or semantics
  - Changing status codes or error formats
- Add **non-breaking changes** to existing version:
  - New optional request fields
  - New response fields (clients ignore unknown fields)
  - New endpoints

❌ **DON'T**:
- Omit version: `/api/trips`
- Use minor/patch in URL: `/api/v1.2.3/trips`
- Make breaking changes to existing versions

### Version Lifecycle

- **Active**: Receives new features and bug fixes
- **Deprecated**: Supported but read-only (no new features), minimum 6 months before removal
- **Removed**: Returns 410 Gone for all endpoints

### Example

```http
✅ /api/v1/trips
✅ /api/v2/trips (future breaking change)

❌ /api/trips
❌ /api/v1.2/trips
```

---

## 4. HTTP Methods

### Standard Methods

| Method | Semantics | Idempotent | Safe | Request Body | Response Body |
|--------|-----------|-----------|------|--------------|---------------|
| **GET** | Retrieve resource(s) | ✅ Yes | ✅ Yes | ❌ No | ✅ Yes |
| **POST** | Create new resource | ❌ No | ❌ No | ✅ Yes | ✅ Yes |
| **PUT** | Replace entire resource | ✅ Yes | ❌ No | ✅ Yes | ✅ Yes |
| **PATCH** | Partial update | ❌ No* | ❌ No | ✅ Yes | ✅ Yes |
| **DELETE** | Remove resource | ✅ Yes | ❌ No | ❌ No | ❌ No (204) |

*PATCH can be idempotent depending on implementation; prefer idempotent when possible

### Rules

✅ **DO**:
- Use **GET** for read-only operations (no side effects)
- Use **POST** for creating new resources
- Use **PUT** for full replacement (all fields required)
- Use **PATCH** for partial updates (only changed fields)
- Use **DELETE** for resource removal
- Make **PUT** and **DELETE** idempotent (multiple identical requests have same effect)

❌ **DON'T**:
- Use GET for mutations (creating, updating, deleting)
- Use POST for updates (use PUT or PATCH)
- Mix semantics arbitrarily

---

## 5. Request Format

### Content Type

All request bodies **MUST** use `Content-Type: application/json`

### Field Naming

Use **snake_case** for all JSON field names:

```json
{
  "start_date": "2026-08-01",
  "end_date": "2026-08-03",
  "travel_style_ids": ["abc", "def"]
}
```

### Required Fields

Document required vs. optional fields in endpoint specifications.

**Convention**: Omitted optional fields are treated as `null` (not as empty string or zero).

### Validation

- Server **MUST** validate all input (never trust client)
- Return **400 Bad Request** for malformed JSON
- Return **422 Unprocessable Entity** for validation failures

### Example

```http
POST /api/v1/trips
Content-Type: application/json

{
  "name": "Weekend in Barcelona",
  "start_date": "2026-08-01",
  "end_date": "2026-08-03",
  "destination_ids": ["dest_abc123"]
}
```

---

## 6. Response Format

### Single Resource Responses

Return resource as **flat JSON object** (no envelope):

```json
{
  "id": "trip_abc123",
  "name": "Weekend in Barcelona",
  "start_date": "2026-08-01",
  "end_date": "2026-08-03",
  "status": "draft",
  "created_at": "2026-07-06T10:00:00Z",
  "updated_at": "2026-07-06T10:00:00Z"
}
```

### List Responses

Return **data envelope** with pagination metadata:

```json
{
  "data": [
    {
      "id": "trip_abc123",
      "name": "Weekend in Barcelona",
      "start_date": "2026-08-01"
    },
    {
      "id": "trip_def456",
      "name": "Summer Road Trip",
      "start_date": "2026-07-15"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 42,
    "total_pages": 3,
    "has_next": true,
    "has_prev": false
  }
}
```

### Empty Lists

Return **empty array** (not 404):

```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 0,
    "total_pages": 0,
    "has_next": false,
    "has_prev": false
  }
}
```

### Field Types

- **IDs**: Strings (UUIDs formatted as `trip_abc123`, `user_xyz789`)
- **Dates**: ISO 8601 date strings (`"2026-08-01"`)
- **Timestamps**: ISO 8601 with timezone (`"2026-07-06T10:00:00Z"`)
- **Booleans**: `true` or `false` (not `1`/`0`)
- **Null**: Use `null` for absent optional values (not empty string)

---

## 7. Error Responses

### Standard Error Format

All errors **MUST** use this structure:

```json
{
  "error": "validation_failed",
  "message": "One or more fields failed validation",
  "request_id": "req_abc123xyz",
  "fields": [
    {
      "field": "email",
      "error": "Email address is already registered"
    },
    {
      "field": "start_date",
      "error": "Start date must be in the future"
    }
  ]
}
```

### Required Fields

- **error**: Machine-readable error code (snake_case, stable across versions)
- **message**: Human-readable English description for developers
- **request_id**: Correlation ID from `X-Request-ID` header

### Optional Fields

- **fields**: Array of field-level errors (for validation failures)
- **details**: Additional context object

### Error Codes Catalog

| Error Code | HTTP Status | Meaning | Use When |
|------------|-------------|---------|----------|
| `invalid_request` | 400 | Malformed request | JSON parse error, missing required params |
| `validation_failed` | 422 | Business rule validation failed | Email already exists, invalid date range |
| `authentication_required` | 401 | Missing/invalid auth token | No JWT, malformed/invalid JWT, bad signature, unknown key, wrong issuer |
| `token_expired` | 401 | Access token expired | Valid but expired access token; client should refresh |
| `forbidden` | 403 | Insufficient permissions | Partner trying to delete trip |
| `not_found` | 404 | Resource doesn't exist | `/trips/nonexistent` |
| `conflict` | 409 | Version mismatch | Optimistic locking conflict |
| `rate_limit_exceeded` | 429 | Too many requests | User exceeded rate limit |
| `internal_error` | 500 | Unexpected server error | Unhandled exception |

### Rules

✅ **DO**:
- Include `request_id` in all errors (enables log tracing)
- Use `fields` array for validation errors
- Keep `message` concise and actionable
- Log full error details server-side (with stack trace)

❌ **DON'T**:
- Expose stack traces to clients
- Include sensitive data in `message` (SQL queries, file paths, secrets)
- Return different error structures for different endpoints

### Example

```http
POST /api/v1/trips
Content-Type: application/json

{
  "name": "",
  "start_date": "2020-01-01"
}

HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json

{
  "error": "validation_failed",
  "message": "One or more fields failed validation",
  "request_id": "req_abc123xyz",
  "fields": [
    {
      "field": "name",
      "error": "Name cannot be empty"
    },
    {
      "field": "start_date",
      "error": "Start date must be in the future"
    }
  ]
}
```

---

## 8. HTTP Status Codes

### Success Codes (2xx)

| Code | Status | Use When | Response Body |
|------|--------|----------|---------------|
| 200 | OK | Successful GET, PUT, PATCH | Resource |
| 201 | Created | Successful POST | Created resource + `Location` header |
| 204 | No Content | Successful DELETE | Empty |

### Client Error Codes (4xx)

| Code | Status | Use When |
|------|--------|----------|
| 400 | Bad Request | Malformed JSON, invalid params, prompt injection |
| 401 | Unauthorized | Missing/invalid JWT |
| 403 | Forbidden | Valid auth but insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Optimistic locking version mismatch |
| 422 | Unprocessable Entity | Valid JSON but business validation failed |
| 429 | Too Many Requests | Rate limit exceeded |

### Server Error Codes (5xx)

| Code | Status | Use When |
|------|--------|----------|
| 500 | Internal Server Error | Unexpected backend error |
| 503 | Service Unavailable | Dependency down, graceful shutdown |

### Rules

✅ **DO**:
- Use semantic status codes (match HTTP standards)
- Include `Location` header with 201 responses
- Include `Retry-After` header with 503 responses
- Log 5xx errors with full context server-side

❌ **DON'T**:
- Return 200 with error payload (anti-pattern)
- Use custom 6xx codes (not in HTTP spec)
- Return 500 for validation errors (use 422)

---

## 9. Pagination

### Request Parameters

| Parameter | Type | Default | Max | Description |
|-----------|------|---------|-----|-------------|
| `page` | integer | 1 | - | Page number (1-indexed) |
| `per_page` | integer | 20 | 100 | Items per page |

### Response Envelope

```json
{
  "data": [...],
  "pagination": {
    "page": 2,
    "per_page": 20,
    "total": 157,
    "total_pages": 8,
    "has_next": true,
    "has_prev": true
  }
}
```

### Rules

✅ **DO**:
- Paginate **all list endpoints** (even if data set is small)
- Return **empty array** for page beyond `total_pages` (not 404)
- Document default and maximum `per_page` values
- Include `has_next` and `has_prev` for client convenience

❌ **DON'T**:
- Allow `per_page` > 100 (return 400 Bad Request)
- Omit pagination for "small" endpoints (data grows over time)

### Example

```http
GET /api/v1/trips?page=2&per_page=20

HTTP/1.1 200 OK
Content-Type: application/json

{
  "data": [...]
  "pagination": {
    "page": 2,
    "per_page": 20,
    "total": 157,
    "total_pages": 8,
    "has_next": true,
    "has_prev": true
  }
}
```

---

## 10. Filtering

### Query Parameter Syntax

**Simple Equality**:
```
?status=completed&user_id=abc123
```

**Range Queries** (date, numeric):
```
?start_date[gte]=2026-08-01&start_date[lte]=2026-08-31
?price[gt]=100&price[lt]=500
```

**Text Search**:
```
?q=barcelona
```

**Multiple Values** (OR logic):
```
?status=pending,approved,rejected
```

### Supported Operators

| Operator | Meaning | Example |
|----------|---------|---------|
| (none) or `[eq]` | Equal | `?status=completed` |
| `[ne]` | Not equal | `?status[ne]=deleted` |
| `[gt]` | Greater than | `?price[gt]=100` |
| `[gte]` | Greater than or equal | `?start_date[gte]=2026-08-01` |
| `[lt]` | Less than | `?price[lt]=500` |
| `[lte]` | Less than or equal | `?end_date[lte]=2026-08-31` |
| `[in]` | In list | `?status[in]=pending,approved` |
| `[like]` | Case-insensitive partial match | `?name[like]=barcelona` |

### Rules

✅ **DO**:
- Validate filter field names (return 400 for unknown fields)
- Validate operator syntax (return 400 for invalid operators)
- Return **empty array** for no matches (not 404)
- Document filterable fields per endpoint

❌ **DON'T**:
- Allow filtering on sensitive fields (passwords, tokens)
- Filter on non-indexed fields without performance testing

### Example

```http
GET /api/v1/trips?status=completed&start_date[gte]=2026-08-01&sort=-created_at

HTTP/1.1 200 OK

{
  "data": [...]
}
```

---

## 11. Sorting

### Query Parameter Syntax

**Single Field** (ascending):
```
?sort=created_at
```

**Single Field** (descending):
```
?sort=-created_at
```

**Multiple Fields** (priority left-to-right):
```
?sort=status,-created_at
```

### Rules

✅ **DO**:
- Use **minus prefix** (`-`) for descending sort
- Support **multi-field sorting** (comma-separated)
- Apply left-to-right priority (first field is primary sort)
- Validate sort field names (return 400 for unknown fields)

❌ **DON'T**:
- Allow sorting on unsortable fields (JSON blobs, text)
- Use separate `order` parameter (`?sort=name&order=desc` is verbose)

### Example

```http
GET /api/v1/trips?sort=status,-created_at

# Sorts by status (ascending), then created_at (descending)
```

---

## 12. Rate Limiting

### Rate Limits (MVP)

| Tier | Limit | Scope |
|------|-------|-------|
| Authenticated users | 100 req/min sustained, 10 req/sec burst | Per user ID |
| Unauthenticated | 10 req/min | Per IP address |
| AI generation endpoints | 5 req/min | Per user ID |

### Response Headers

**All responses** include:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1720270800
```

- `X-RateLimit-Limit`: Total requests allowed per window
- `X-RateLimit-Remaining`: Requests remaining in current window
- `X-RateLimit-Reset`: Unix timestamp when window resets

### Rate Limit Exceeded Response

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
Retry-After: 23

{
  "error": "rate_limit_exceeded",
  "message": "You have exceeded the rate limit. Please try again in 23 seconds.",
  "request_id": "req_xyz789",
  "retry_after": 23
}
```

### Rules

✅ **DO**:
- Include rate limit headers in **all responses** (not just 429)
- Use **user ID** for authenticated requests (not IP)
- Use **IP address** for unauthenticated requests
- Include `Retry-After` header in 429 responses

❌ **DON'T**:
- Count 429 responses against rate limit (prevent compounding)
- Block users permanently (rate limits reset every minute)

---

## 13. Authentication Headers

### Authorization Header

```http
Authorization: Bearer <jwt_token>
```

**Note**: In MVP, JWT is stored in HTTP-only cookie, so this header may not be used. Include here for future external API consumers.

**As-built (2026-07-23, issue #179)**: `middleware.Authenticate` accepts the `access_token` cookie
**only** — no `Authorization: Bearer` code path exists yet. The generated Swagger 2.0 contract
therefore declares its security definition as `CookieAuth` (`type: apiKey`, `in: header`,
`name: Cookie`), not a bearer scheme; see `backend/cmd/api/docs.go`. Revisit when the external-API
consumer case above is actually built.

### Request ID Header

All requests **SHOULD** include (client-generated):

```http
X-Request-ID: req_abc123xyz
```

If omitted, server generates one.

### Custom Headers

Avoid custom headers when possible. Use standard HTTP headers or query parameters.

---

## 14. Timestamps and Dates

### Format

- **Dates**: ISO 8601 date (`"2026-08-01"`)
- **Timestamps**: ISO 8601 with UTC timezone (`"2026-07-06T10:00:00Z"`)

### Rules

✅ **DO**:
- Always use **UTC timezone** (`Z` suffix)
- Use consistent format across all endpoints
- Parse client-provided timestamps leniently (accept multiple ISO formats)
- Return timestamps in consistent format (always `Z` suffix)

❌ **DON'T**:
- Use Unix timestamps (not human-readable)
- Use local timezones without offset
- Mix date formats (`2026-08-01` vs `08/01/2026`)

### Example

```json
{
  "start_date": "2026-08-01",
  "created_at": "2026-07-06T10:00:00Z",
  "updated_at": "2026-07-06T14:30:15Z"
}
```

---

## 15. Endpoint Patterns

### Pattern: List Resources

```http
GET /api/v1/{resource}?page=1&per_page=20&sort=-created_at

200 OK
{
  "data": [...],
  "pagination": {...}
}
```

### Pattern: Get Single Resource

```http
GET /api/v1/{resource}/:id

200 OK
{
  "id": "...",
  ...
}

404 Not Found
{
  "error": "not_found",
  "message": "Resource not found",
  "request_id": "..."
}
```

### Pattern: Create Resource

```http
POST /api/v1/{resource}
Content-Type: application/json

{
  "field1": "value1",
  ...
}

201 Created
Location: /api/v1/{resource}/:id
{
  "id": "...",
  ...
}

422 Unprocessable Entity
{
  "error": "validation_failed",
  "message": "...",
  "request_id": "...",
  "fields": [...]
}
```

### Pattern: Update Resource (Full Replacement)

```http
PUT /api/v1/{resource}/:id
Content-Type: application/json

{
  "field1": "new_value1",
  ...
}

200 OK
{
  "id": "...",
  ...
}

409 Conflict
{
  "error": "conflict",
  "message": "Resource was modified by another request",
  "request_id": "...",
  "current_version": {...}
}
```

### Pattern: Update Resource (Partial)

```http
PATCH /api/v1/{resource}/:id
Content-Type: application/json

{
  "field1": "new_value1"
}

200 OK
{
  "id": "...",
  ...
}
```

### Pattern: Delete Resource

```http
DELETE /api/v1/{resource}/:id

204 No Content
```

### Pattern: Action on Resource

```http
POST /api/v1/{resource}/:id/{action}
Content-Type: application/json

{
  "param1": "value1"
}

200 OK
{
  "result": "..."
}
```

**Example**: `POST /api/v1/trips/:id/generate`, `POST /api/v1/suggestions/:id/approve`

---

## 16. Machine-Readable Contract

This document remains the human-readable source of truth for API conventions. A derived,
always-current **machine-readable** OpenAPI v3 contract is generated directly from Go doc-comment
annotations on each handler (`swaggo/swag`) and served at:

- **`GET /swagger/doc.json`** — the generated OpenAPI v3 document
- **`GET /swagger/index.html`** — Swagger UI, an interactive explorer rendering the document above

The generated contract is regenerated and drift-checked in CI on every backend change, so it can
never fall out of sync with the implementation (see `docs/testing-guidelines.md` for the CI
drift-check gate). It does not restate or replace the conventions above — see
`specs/009-api-documentation/spec.md` for the full feature definition.

---

## Compliance

### How to Validate Compliance

1. **Code Review Checklist**: Reference this document in PR reviews
2. **Integration Tests**: Validate error format, headers, status codes
3. **API Documentation**: Ensure all endpoints document adherence to these standards
4. **Linting**: Where possible, automate checks (URL structure, field naming)

### Exceptions

If an endpoint **must** deviate from these standards (rare), document the exception with:
- **Why**: Technical or business reason
- **Alternative**: What pattern is used instead
- **Approval**: Technical lead sign-off

### Updates

This document is versioned (SemVer). Breaking changes to standards require:
- Constitution update (if mandated stack/patterns change)
- Cross-functional review (backend, frontend, PM)
- Migration plan for existing endpoints

---

## References

- **Research**: `specs/007-api-design-standards/research.md` — Detailed rationale for each decision
- **Existing Contracts**: `specs/001-product-vision-scope/contracts/api.md` — Examples of endpoints following these standards
- **Architecture**: `docs/architecture.md` — REST API component overview
- **Security**: `docs/security.md` — Authentication and authorization context
- **NFRs**: `docs/nfrs.md` — NFR-API-001, NFR-API-003, NFR-REL-003
- **Machine-Readable Contract**: `specs/009-api-documentation/spec.md` — OpenAPI v3 generation
  (`swaggo/swag`) and Swagger UI (`/swagger/*`)

<!-- PROMOTED:api-standards END -->

---

**Document Version**: 1.0.0  
**Last Updated**: 2026-07-06  
**Next Review**: After first implementation sprint
