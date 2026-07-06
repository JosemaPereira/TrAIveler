# Data Model: API Design Standards

**Feature**: API Design Standards

**Date**: 2026-07-06

**Note**: This feature is documentation-only and does not introduce new database entities. Instead, this document defines the logical structure of API design conventions.

---

## Logical Model

This feature establishes a **convention catalog** — a structured set of rules that all API endpoints must follow. The catalog is not stored in a database but exists as documentation that developers reference during implementation and code review.

### Entity: API Standard

An API Standard is a documented rule or pattern that applies to API endpoints.

**Attributes**:
- **Category**: The aspect of API design this standard governs (Resource Naming, URL Structure, Versioning, Request Format, Response Format, Error Handling, Status Codes, Pagination, Filtering, Sorting, Rate Limiting)
- **Rule**: The specific convention (e.g., "Use plural nouns for collection resources")
- **Rationale**: Why this rule exists (consistency, industry best practice, technical constraint)
- **Examples**: Code samples demonstrating correct usage
- **Counter-Examples**: Anti-patterns to avoid
- **Enforcement**: How the rule is validated (manual code review, linter, integration tests)

### Entity: Endpoint Pattern

A reusable template for common API operations (e.g., "List Resource", "Get Single Item", "Create Item").

**Attributes**:
- **Pattern Name**: Descriptive identifier (e.g., "List Resources with Pagination")
- **HTTP Method**: GET, POST, PUT, PATCH, DELETE
- **URL Template**: Pattern with placeholders (e.g., `/api/v1/{resource}`)
- **Request Parameters**: Query params, path params, headers, body schema
- **Response Structure**: Success response format with examples
- **Error Responses**: Possible error codes and formats
- **Status Codes**: Expected HTTP status codes for success/error cases

### Entity: Error Response Format

The standardized structure for error responses across all endpoints.

**Attributes** (documented in research.md Decision 5):
- **error**: Machine-readable error code (snake_case)
- **message**: Human-readable description (English)
- **request_id**: Correlation ID from `X-Request-ID` header
- **fields** (optional): Array of field-level validation errors
- **details** (optional): Additional context object

**Validation Rules**:
- `error` code MUST be documented in the error codes catalog
- `message` MUST NOT contain sensitive information (stack traces, SQL queries, internal paths)
- `request_id` MUST match the `X-Request-ID` header in the request
- `fields` array MUST only appear for validation errors (400, 422 status codes)

### Entity: Rate Limit Policy

Configuration for API rate limiting by user tier.

**Attributes** (documented in research.md Decision 9):
- **User Tier**: authenticated, unauthenticated, AI endpoints
- **Sustained Rate**: Requests per minute (e.g., 100/min)
- **Burst Rate**: Requests per second (e.g., 10/sec)
- **Scope**: per-user, per-IP, per-endpoint
- **Response Headers**: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`
- **Error Code**: `rate_limit_exceeded` (429 status)

---

## Relationships

```
API Standard
  └─ applies to → Endpoint Pattern
       └─ may return → Error Response Format
       └─ enforces → Rate Limit Policy
```

---

## Validation Rules

### Standard Consistency

- [API] All endpoints within the same API version MUST follow identical conventions
- [API] Breaking a documented standard requires either fixing the endpoint OR amending the standard (with constitution update)
- [API] New standards MUST NOT contradict existing standards (resolve conflicts before adding)

### Pattern Completeness

- [Logic] Every Endpoint Pattern MUST document all possible status codes
- [Logic] Every Endpoint Pattern MUST include at least one example request and response
- [Logic] Endpoint Pattern examples MUST be valid according to documented Request/Response Format standards

### Error Format Consistency

- [DB] Error codes MUST be unique (no duplicate `error` values with different meanings)
- [Logic] Error messages MUST be English, concise (max 200 characters), and actionable
- [API] 4xx errors MUST include `request_id` for client-side error reporting

### Rate Limit Transparency

- [API] All responses MUST include `X-RateLimit-*` headers (not just 429 responses)
- [API] 429 responses MUST include `Retry-After` header (seconds until reset)
- [Logic] Rate limit counters MUST NOT increment for 429 responses (prevent compounding)

---

## Invariants

1. **URL Immutability**: Once an endpoint is deployed in a given version (e.g., `/api/v1/trips`), its URL path CANNOT change within that version. Breaking changes require a new version (`/api/v2/trips`).

2. **Error Code Stability**: `error` codes in responses are part of the API contract. Clients depend on them for error handling logic. Changing an error code is a breaking change.

3. **Pagination Backward Compatibility**: Changing pagination from offset-based to cursor-based MUST support both mechanisms during transition period (minimum 6 months).

4. **Rate Limit Fairness**: Authenticated users MUST be rate-limited by user ID (not IP address) to prevent shared-network users from being unfairly penalized.

5. **Status Code Semantics**: HTTP status codes MUST match RFC 7231 definitions. Using `200 OK` with an error payload is forbidden.

---

## Performance Considerations

- **Pagination**: Offset-based pagination performance degrades with deep pages (OFFSET 10000 is slow). For MVP scale (< 50 users, hundreds of records), this is acceptable. Consider cursor-based migration for endpoints with >10k records.

- **Filtering**: Complex filter queries with multiple conditions can generate slow SQL. Indexes MUST be created for commonly filtered fields (see `docs/data-model.md`).

- **Rate Limiting**: Rate limit state MUST be stored in fast key-value store (Redis) to avoid database bottleneck on every request. Cache TTL matches rate limit window (60 seconds).

---

## Migration Strategy

**Phase 1: Document Standards** (This Spec)
- Create `docs/api-design-standards.md` with all conventions
- Add examples for each pattern
- Define error code catalog

**Phase 2: Validate Existing Endpoints** (Code Review)
- Audit all endpoints in spec 001, 004 contracts against new standards
- Identify any deviations (grandfather in OR fix)
- Document exceptions with rationale

**Phase 3: Enforcement**
- Add API standards section to PR template (`.github/pull_request_template.md`)
- Create integration tests that validate error format, headers, status codes
- Add linter rules where possible (URL structure, field naming)

**Phase 4: New Endpoint Checklist**
- All new endpoints MUST include standards compliance checklist in PR
- Code reviewers MUST verify compliance before approval

---

## Cross-References

- **Related Specs**: 001 (existing contracts), 002 (NFR-API-001, NFR-API-003, NFR-REL-003), 004 (authentication)
- **Documentation**: `docs/api-design-standards.md` (output of this spec), `docs/architecture.md` (REST API component)
- **Testing**: Integration tests validate standards compliance (status codes, headers, error format)
