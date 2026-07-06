# Research: API Design Standards

**Feature**: API Design Standards

**Date**: 2026-07-06

**Purpose**: Document research findings and decisions for establishing project-wide API design conventions

## Research Tasks Completed

1. REST resource naming conventions analysis
2. URL structure and nesting depth best practices
3. API versioning strategies evaluation
4. Request/response format standards
5. Error response format patterns
6. HTTP status code conventions
7. Pagination approaches
8. Filtering and sorting strategies
9. Rate limiting implementation patterns

---

## Decision 1: Resource Naming Conventions

**Research Question**: Should resources use singular or plural nouns in URLs? What casing? How to handle compound words?

**Options Evaluated**:
- **Plural nouns** (REST best practice: `/trips`, `/users`)
- **Singular nouns** (RPC-style: `/trip`, `/user`)
- **Mixed approach** (inconsistent)

**Decision**: **Plural nouns for collection resources**

**Rationale**:
- Industry standard (Google, GitHub, Stripe, AWS APIs)
- Semantically correct: GET `/trips` returns multiple trips
- Consistent with existing contracts in spec 001 (`/trips`, `/plans`, `/destinations`)
- URL readability: `/trips/:id/days` reads naturally as "days within a trip"

**Conventions Established**:
- Collection endpoints: plural nouns (`/trips`, `/users`, `/activities`)
- Casing: lowercase with hyphens for compound words (`/travel-styles`, `/conversation-messages`)
- No file extensions in URLs (`.json`, `.xml` — use Accept headers instead)
- Action endpoints (non-CRUD): verbs after resource (`/trips/:id/generate`, `/suggestions/:id/approve`)

**Alternatives Considered**:
- Singular nouns rejected: conflicts with REST semantics, inconsistent with industry
- CamelCase/snake_case rejected: hyphens are URL-native and most readable

**Source**: REST API Design Rulebook (O'Reilly), Google API Design Guide

---

## Decision 2: URL Structure and Nesting Depth

**Research Question**: How deep should nested resources go? When should resources be top-level vs. nested?

**Options Evaluated**:
- **Shallow nesting** (max 2 levels: `/trips/:id/days`)
- **Deep nesting** (reflect full hierarchy: `/trips/:id/days/:day_id/activities/:activity_id`)
- **Flat structure** (all top-level with query params: `/activities?trip_id=:id`)

**Decision**: **Maximum 2 levels of nesting; use top-level endpoints for deeply nested resources**

**Rationale**:
- Deep nesting creates brittle URLs that break when hierarchies change
- Long URLs are hard to read, type, and remember
- A resource that can be accessed independently should have a top-level endpoint
- Existing pattern in spec 001: `/trips/:id/days` (2 levels), but also `/destinations` (top-level)

**Conventions Established**:
- Parent-child relationships: 1 level nesting (`/trips/:id/days`, `/trips/:id/collaborators`)
- Grandchild resources: provide both nested and top-level access:
  - Nested (context): `/trips/:id/days/:day_id/activities` (when listing within a day)
  - Top-level (direct): `/activities/:id` (when accessing a specific activity)
- Query parameters for filtering across hierarchies: `/activities?trip_id=:id`

**Alternatives Considered**:
- Deep nesting rejected: URLs like `/trips/:id/days/:day_id/activities/:activity_id/edit` are unmanageable
- Fully flat rejected: loses semantic relationship between parent and child resources

**Source**: Stripe API (uses 2-level nesting max), Heroku API Design Guide

---

## Decision 3: API Versioning Strategy

**Research Question**: How should the API be versioned? URL path, header, query parameter?

**Options Evaluated**:
- **URL path versioning** (`/api/v1/trips`)
- **Header versioning** (`Accept: application/vnd.traiveler.v1+json`)
- **Query parameter** (`/api/trips?version=1`)
- **No versioning** (breaking changes force client updates)

**Decision**: **URL path versioning with `/api/v1` prefix**

**Rationale**:
- Already established in spec 001 contracts (`/api/v1/...`)
- Most visible and discoverable (appears in browser, logs, documentation)
- Easy to route and cache at CDN/load balancer level
- Simple for frontend developers (no custom headers)
- Version is explicit in every request (no ambiguity)

**Conventions Established**:
- Current version: `/api/v1` (major version only, no minor/patch in URL)
- All endpoints MUST include version prefix
- Breaking changes require new major version (`/api/v2`)
- Non-breaking changes (new optional fields, new endpoints) added to existing version
- Old versions supported for minimum 6 months after new version release

**Alternatives Considered**:
- Header versioning rejected: harder to test in browser, invisible in logs, requires client-side header management
- Query parameter rejected: pollutes every endpoint, easy to omit accidentally
- No versioning rejected: forces all clients to upgrade simultaneously (impossible for public APIs)

**Source**: AWS API Gateway best practices, Azure API Management versioning guide

---

## Decision 4: Request and Response Format

**Research Question**: What should be the standard structure for request/response payloads?

**Options Evaluated**:
- **Flat JSON** (fields at top level: `{"name":"Trip","start_date":"..."}`)
- **Wrapped JSON** (data in envelope: `{"data":{"name":"Trip"},"meta":{}}`)
- **JSON:API specification** (strict envelope with `type`, `attributes`, `relationships`)

**Decision**: **Flat JSON for simple resources, consistent envelope for lists and errors**

**Rationale**:
- Matches existing spec 001 contracts (flat objects for single resources)
- Simpler for clients: no unwrapping required for single-item responses
- Lists need metadata (pagination): use consistent envelope structure
- Errors need structure (code, message, fields): use consistent envelope

**Conventions Established**:

**Single Resource Responses** (GET, POST, PUT, PATCH):
```json
{
  "id": "uuid",
  "name": "Weekend in Barcelona",
  "start_date": "2026-08-01",
  "end_date": "2026-08-03",
  "created_at": "2026-07-06T10:00:00Z",
  "updated_at": "2026-07-06T10:00:00Z"
}
```

**List Responses** (GET collections):
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 42,
    "total_pages": 3
  }
}
```

**Empty Lists**: Return `{"data":[], "pagination":{...}}` (not 404)

**Field Naming**:
- `snake_case` for all JSON fields (Go convention, matches PostgreSQL)
- ISO 8601 for all timestamps (`2026-07-06T10:00:00Z`)
- UUIDs as strings (not binary)
- Booleans as true/false (not 1/0)

**Alternatives Considered**:
- Wrapped JSON everywhere rejected: adds unnecessary complexity for single resources
- CamelCase rejected: inconsistent with backend language (Go) and database (PostgreSQL)
- JSON:API rejected: over-engineered for MVP scope, steep learning curve

**Source**: Existing spec 001 contracts, Google JSON Style Guide

---

## Decision 5: Standardized Error Response Format

**Research Question**: What structure should error responses follow to provide actionable information?

**Decision**: **Consistent error envelope with machine-readable codes and field-level details**

**Conventions Established**:

**Error Response Structure**:
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

**Required Fields**:
- `error`: Machine-readable error code (snake_case, stable across versions)
- `message`: Human-readable English description for developers
- `request_id`: Correlation ID from `X-Request-ID` header for log tracing

**Optional Fields**:
- `fields`: Array of field-level errors (for validation failures)
- `details`: Additional context (e.g., `{"max_length": 255, "provided_length": 300}`)

**Error Codes** (machine-readable, documented in API standards):
- `invalid_request` — malformed JSON, missing required parameters
- `validation_failed` — business rule validation failed
- `authentication_required` — missing or invalid JWT
- `forbidden` — user lacks permission for this action
- `not_found` — resource does not exist
- `conflict` — optimistic locking version mismatch
- `rate_limit_exceeded` — client has exceeded rate limit
- `internal_error` — unexpected server error (details hidden from client)

**Rationale**:
- `request_id` enables support teams to trace errors in logs
- Machine-readable `error` codes allow frontend to implement specific error handling
- Field-level errors enable inline validation feedback in forms
- Consistent structure across all endpoints reduces client-side error handling code

**Alternatives Considered**:
- HTTP status code only rejected: not enough detail for actionable user feedback
- Plain text messages rejected: requires string parsing, breaks with i18n
- Stack traces in production rejected: security risk (information disclosure)

**Source**: Stripe API error responses, Twilio API errors, RFC 7807 (Problem Details for HTTP APIs)

---

## Decision 6: HTTP Status Code Usage

**Research Question**: Which HTTP status codes should be used for common scenarios?

**Decision**: **Strict adherence to semantic HTTP status codes**

**Conventions Established**:

**2xx Success**:
- `200 OK` — Successful GET, PUT, PATCH (returns resource)
- `201 Created` — Successful POST (returns created resource + `Location` header)
- `204 No Content` — Successful DELETE (no response body)

**4xx Client Errors**:
- `400 Bad Request` — Malformed JSON, invalid parameters, prompt injection detected
- `401 Unauthorized` — Missing or invalid JWT token
- `403 Forbidden` — Valid auth but insufficient permissions (e.g., partner cannot delete trips)
- `404 Not Found` — Resource does not exist (or user lacks permission to know it exists)
- `409 Conflict` — Optimistic locking version mismatch, duplicate resource creation
- `422 Unprocessable Entity` — Valid JSON structure but business rule validation failed
- `429 Too Many Requests` — Rate limit exceeded

**5xx Server Errors**:
- `500 Internal Server Error` — Unexpected backend error (log with request ID, hide details from client)
- `503 Service Unavailable` — Dependency unavailable (database, AI provider) or graceful shutdown in progress

**Never Use**:
- `200 OK` with error payload (anti-pattern: breaks HTTP semantics)
- `418 I'm a teapot` (RFC 2324 joke, not for production)

**Rationale**:
- Standard HTTP semantics enable middleware (proxies, CDNs) to handle responses correctly
- Client libraries (axios, fetch) can distinguish success/error by status code range
- Monitoring systems can alert on 5xx without parsing response bodies
- Matches existing spec 001 contracts and spec 002 NFR-API-001 (status code correctness)

**Alternatives Considered**:
- Always returning 200 with error payload rejected: breaks REST, confuses monitoring, violates HTTP standards

**Source**: RFC 7231 (HTTP/1.1 Semantics), Mozilla MDN HTTP Status Reference

---

## Decision 7: Pagination Conventions

**Research Question**: What pagination approach should list endpoints use?

**Options Evaluated**:
- **Offset-based** (`?page=2&per_page=20`)
- **Cursor-based** (`?cursor=abc123&limit=20`)
- **Keyset pagination** (using last record's ID/timestamp)

**Decision**: **Offset-based pagination for MVP with cursor-based migration path**

**Rationale**:
- Simpler to implement and understand (matches most SQL `LIMIT/OFFSET`)
- Frontend can show "Page X of Y" and jump to arbitrary pages (better UX for trips list)
- MVP scale (< 50 users, hundreds of trips) won't hit offset performance issues
- Existing spec 001 doesn't specify pagination strategy → choose simplest for MVP
- Cursor-based can be added later for high-volume endpoints without breaking existing clients

**Conventions Established**:

**Request Parameters**:
- `page`: Page number (1-indexed, default: 1)
- `per_page`: Items per page (default: 20, max: 100)

**Response Envelope**:
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

**Rules**:
- All list endpoints MUST support pagination (even if data set is small)
- `per_page` values above 100 return 400 Bad Request
- `page` beyond `total_pages` returns empty array (not 404)
- `total` and `total_pages` MAY be omitted for cursor-based pagination (future)

**Alternatives Considered**:
- Cursor-based rejected for MVP: harder to implement, no "jump to page N" UX, overkill for MVP scale
- No pagination rejected: violates NFR-API-003 (pagination required), breaks on large data sets

**Source**: GitHub REST API (uses offset + cursor hybrid), Slack API pagination guide

---

## Decision 8: Filtering and Sorting Conventions

**Research Question**: How should clients filter and sort list results?

**Decision**: **Query parameter-based filtering with multi-field sort support**

**Conventions Established**:

**Filtering**:
- Simple equality: `?status=completed&user_id=abc123`
- Date ranges: `?start_date[gte]=2026-08-01&start_date[lte]=2026-08-31`
- Text search: `?q=barcelona` (full-text search on searchable fields)
- Multiple values (OR): `?status=pending,approved` (comma-separated)

**Operators** (for range queries):
- `[eq]` — Equal (default, can be omitted)
- `[ne]` — Not equal
- `[gt]` — Greater than
- `[gte]` — Greater than or equal
- `[lt]` — Less than
- `[lte]` — Less than or equal
- `[in]` — In list (comma-separated values)
- `[like]` — Case-insensitive partial match (SQL `ILIKE %term%`)

**Sorting**:
- Single field: `?sort=created_at` (ascending by default)
- Descending: `?sort=-created_at` (prefix with `-`)
- Multiple fields: `?sort=status,-created_at` (comma-separated, left-to-right priority)

**Example**:
```
GET /api/v1/trips?status=pending&start_date[gte]=2026-08-01&sort=-created_at&page=1&per_page=20
```

**Rules**:
- Filtering on non-existent fields returns 400 Bad Request
- Invalid operator syntax returns 400 Bad Request
- Empty filter results return empty array (not 404)
- Sorting on non-sortable fields (e.g., JSON blobs) returns 400 Bad Request

**Rationale**:
- Query parameters are RESTful and easy to construct in JavaScript (`new URLSearchParams()`)
- Operator syntax `[gte]`, `[lte]` is intuitive and used by Stripe, Laravel, MongoDB
- Minus prefix for descending sort is concise and used by JSON:API, Django REST Framework
- Multi-field sort enables "sort by status, then by date" UX

**Alternatives Considered**:
- JSON in query string rejected: hard to construct, URL-encoding issues, not RESTful
- POST with filter payload rejected: breaks HTTP semantics (POST implies mutation), not cacheable

**Source**: Stripe API filtering, Laravel query builder, JSON:API specification

---

## Decision 9: Rate Limiting Strategy

**Research Question**: How should rate limiting be implemented and communicated to clients?

**Options Evaluated**:
- **Fixed window** (100 req/min, resets at :00)
- **Sliding window** (100 req in rolling 60-second window)
- **Token bucket** (burst allowed, refills over time)
- **Per-user vs per-IP** (authenticated vs anonymous)

**Decision**: **Sliding window per authenticated user with burst tolerance**

**Rationale**:
- Sliding window prevents "double-dip" at window boundary (more fair than fixed window)
- Per-user limits (not IP) prevent one abusive user from blocking others behind shared NAT
- Burst tolerance (e.g., 10 req/sec burst, 100 req/min sustained) allows legitimate client retries
- Existing spec 002 NFR-REL-003 specifies rate limiting as non-functional requirement

**Conventions Established**:

**Rate Limits** (MVP):
- Authenticated users: 100 requests/minute sustained, 10 requests/second burst
- Unauthenticated endpoints (`/auth/register`, `/auth/login`): 10 requests/minute per IP
- AI generation endpoints: 5 requests/minute per user (AI provider rate limit protection)

**Response Headers** (included in all responses):
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1720270800
```

**429 Response** (when limit exceeded):
```json
{
  "error": "rate_limit_exceeded",
  "message": "You have exceeded the rate limit. Please try again in 23 seconds.",
  "request_id": "req_xyz789",
  "retry_after": 23
}
```
Header: `Retry-After: 23` (seconds until reset)

**Rules**:
- Rate limit state stored in Redis (shared across ECS tasks)
- Anonymous requests use IP address as key (from `X-Forwarded-For` header, validated)
- Authenticated requests use user ID as key (more accurate, prevents IP sharing issues)
- 429 responses do NOT count against rate limit (prevent compounding)

**Rationale for Headers**:
- `X-RateLimit-*` headers are industry standard (GitHub, Twitter, Stripe)
- `Retry-After` enables clients to implement exponential backoff correctly
- Including headers in all responses (not just 429) allows clients to throttle proactively

**Alternatives Considered**:
- Fixed window rejected: "double dip" exploit at window boundaries
- Token bucket rejected: more complex to implement, harder to explain to users
- Per-IP for authenticated users rejected: shared office/VPN users unfairly penalized

**Source**: GitHub Rate Limiting, Stripe Rate Limits, IETF Draft on RateLimit Headers

---

## Summary of All Decisions

| Convention | Decision | Rationale |
|------------|----------|-----------|
| **Resource Naming** | Plural nouns, lowercase, hyphenated compounds | REST best practice, matches spec 001 |
| **URL Nesting** | Max 2 levels; deep resources get top-level endpoints | Readability, flexibility, Stripe pattern |
| **Versioning** | URL path `/api/v1` | Visibility, caching, already established |
| **Request/Response** | Flat JSON for singles, envelope for lists | Simplicity + metadata when needed |
| **Field Naming** | `snake_case`, ISO 8601 timestamps, UUIDs as strings | Backend/DB consistency |
| **Error Format** | `{error, message, request_id, fields?}` | Traceability, actionable feedback |
| **Status Codes** | Semantic HTTP (200/201/204, 400/401/403/404/409/422, 500/503) | HTTP standards, monitoring |
| **Pagination** | Offset-based (`page`, `per_page`) with metadata envelope | MVP simplicity, UX flexibility |
| **Filtering** | Query params with operators (`[gte]`, `[like]`) | RESTful, intuitive, Stripe-like |
| **Sorting** | Query param `sort=field,-field2` (minus = descending) | Concise, JSON:API pattern |
| **Rate Limiting** | Sliding window per user, 100/min, headers in all responses | Fairness, transparency, GitHub pattern |

---

## Cross-References

- **Existing Contracts**: `specs/001-product-vision-scope/contracts/api.md` — Already uses `/api/v1`, plural nouns, flat JSON
- **Architecture**: `docs/architecture.md` — RESTful API backend on ECS Fargate
- **Security**: `docs/security.md` — JWT authentication (affects rate limiting per-user), prompt validation (400 status code)
- **NFRs**: `docs/nfrs.md` — NFR-API-001 (status codes), NFR-API-003 (pagination), NFR-REL-003 (rate limiting)

---

## Open Questions

None — all technical clarifications resolved through existing project documentation and industry best practices research.
