# Quickstart: API Design Standards Validation

**Feature**: API Design Standards

**Date**: 2026-07-06

**Purpose**: Runnable validation scenarios to verify that API endpoints comply with documented standards

---

## Prerequisites

- Backend API running locally or in staging environment
- HTTP client (curl, Postman, HTTPie, or integration test suite)
- Valid JWT token for authenticated endpoints
- Sample data seeded in database

---

## Validation Scenarios

### Scenario 1: Resource Naming Conventions

**Objective**: Verify that all endpoints use plural nouns, lowercase, hyphenated compound words

**Steps**:

1. List all API routes:
   ```bash
   curl http://localhost:8080/api/v1/routes
   # Or inspect router configuration in code
   ```

2. Verify resource naming patterns:
   ```bash
   # ✅ PASS: Plural nouns
   GET /api/v1/trips
   GET /api/v1/users
   GET /api/v1/activities
   
   # ✅ PASS: Hyphenated compounds
   GET /api/v1/travel-styles
   GET /api/v1/conversation-messages
   
   # ❌ FAIL: Would be violations
   GET /api/v1/trip          # singular
   GET /api/v1/travelStyles  # camelCase
   GET /api/v1/trips.json    # file extension
   ```

**Expected Outcome**: All endpoints follow plural noun + lowercase + hyphenated conventions

**Validation Method**: Manual code review + automated linter rule

---

### Scenario 2: URL Nesting Depth

**Objective**: Verify that URL nesting does not exceed 2 levels

**Steps**:

1. Test parent-child relationships (1 level nesting - PASS):
   ```bash
   GET /api/v1/trips/:id/days
   GET /api/v1/trips/:id/collaborators
   GET /api/v1/users/:id/subscriptions
   ```

2. Test grandchild resources (2 levels maximum - PASS):
   ```bash
   GET /api/v1/trips/:id/days/:day_id
   GET /api/v1/trips/:id/days/:day_id/activities
   ```

3. Verify top-level endpoints exist for deep resources (PASS):
   ```bash
   GET /api/v1/activities/:id
   # Can access activity directly without traversing trip → day → activity
   ```

4. Verify deep nesting is avoided (FAIL example):
   ```bash
   # ❌ This would violate standards:
   GET /api/v1/trips/:id/days/:day_id/activities/:activity_id/edit
   ```

**Expected Outcome**: No URLs exceed 2 levels of nesting; deep resources have top-level endpoints

**Validation Method**: Route inspection + integration tests

---

### Scenario 3: API Versioning

**Objective**: Verify that all endpoints include `/api/v1` prefix

**Steps**:

1. Test that versioned endpoints work:
   ```bash
   curl http://localhost:8080/api/v1/trips
   # Expected: 200 OK
   ```

2. Test that non-versioned endpoints return 404:
   ```bash
   curl http://localhost:8080/api/trips
   # Expected: 404 Not Found
   ```

3. Verify version appears in all endpoint URLs:
   ```bash
   GET /api/v1/auth/login
   GET /api/v1/trips
   POST /api/v1/trips/:id/generate
   ```

**Expected Outcome**: All endpoints include `/api/v1` prefix; unversioned URLs are not routed

**Validation Method**: Integration tests + manual inspection

---

### Scenario 4: Request/Response Format

**Objective**: Verify that single resources return flat JSON, lists return data envelope

**Steps**:

1. Test single resource response (GET, POST, PUT):
   ```bash
   curl -X GET http://localhost:8080/api/v1/trips/:id \
     -H "Authorization: Bearer $JWT"
   
   # Expected (flat JSON):
   {
     "id": "trip_abc123",
     "name": "Weekend in Barcelona",
     "start_date": "2026-08-01",
     "created_at": "2026-07-06T10:00:00Z"
   }
   ```

2. Test list response:
   ```bash
   curl -X GET http://localhost:8080/api/v1/trips \
     -H "Authorization: Bearer $JWT"
   
   # Expected (envelope with pagination):
   {
     "data": [
       { "id": "trip_abc123", ... },
       { "id": "trip_def456", ... }
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

3. Test empty list response:
   ```bash
   curl -X GET "http://localhost:8080/api/v1/trips?status=nonexistent" \
     -H "Authorization: Bearer $JWT"
   
   # Expected (empty data array, not 404):
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

4. Verify field naming (snake_case):
   ```json
   {
     "start_date": "2026-08-01",     # ✅ snake_case
     "travel_style_ids": [...],      # ✅ snake_case
     "created_at": "...",            # ✅ snake_case
     
     # ❌ Would be violations:
     "startDate": "...",             # camelCase
     "TravelStyleIds": [...],        # PascalCase
   }
   ```

**Expected Outcome**: Single resources are flat, lists have envelope, fields use snake_case

**Validation Method**: Integration tests asserting response structure

---

### Scenario 5: Error Response Format

**Objective**: Verify that all errors follow standardized format with error code, message, request_id

**Steps**:

1. Test validation error (422):
   ```bash
   curl -X POST http://localhost:8080/api/v1/trips \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer $JWT" \
     -d '{
       "name": "",
       "start_date": "2020-01-01"
     }'
   
   # Expected:
   HTTP/1.1 422 Unprocessable Entity
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

2. Test authentication error (401):
   ```bash
   curl -X GET http://localhost:8080/api/v1/trips
   # No Authorization header
   
   # Expected:
   HTTP/1.1 401 Unauthorized
   {
     "error": "authentication_required",
     "message": "Authentication token is missing or invalid",
     "request_id": "req_def456uvw"
   }
   ```

3. Test not found error (404):
   ```bash
   curl -X GET http://localhost:8080/api/v1/trips/nonexistent \
     -H "Authorization: Bearer $JWT"
   
   # Expected:
   HTTP/1.1 404 Not Found
   {
     "error": "not_found",
     "message": "Trip not found",
     "request_id": "req_ghi789rst"
   }
   ```

4. Test forbidden error (403):
   ```bash
   curl -X DELETE http://localhost:8080/api/v1/trips/:id \
     -H "Authorization: Bearer $PARTNER_JWT"
   # Partner role trying to delete (only admin can delete)
   
   # Expected:
   HTTP/1.1 403 Forbidden
   {
     "error": "forbidden",
     "message": "You do not have permission to perform this action",
     "request_id": "req_jkl012mno"
   }
   ```

5. Verify `request_id` correlation:
   ```bash
   # Send request with custom request ID:
   curl -X GET http://localhost:8080/api/v1/trips/nonexistent \
     -H "X-Request-ID: my_custom_id_123" \
     -H "Authorization: Bearer $JWT"
   
   # Expected: Error response includes same request_id
   {
     "error": "not_found",
     "message": "Trip not found",
     "request_id": "my_custom_id_123"
   }
   ```

**Expected Outcome**: All errors include `error` (code), `message`, `request_id`; validation errors include `fields` array

**Validation Method**: Integration tests for each error scenario

---

### Scenario 6: HTTP Status Code Usage

**Objective**: Verify that endpoints return semantically correct HTTP status codes

**Steps**:

1. Test success codes:
   ```bash
   # 200 OK for GET, PUT, PATCH
   GET /api/v1/trips/:id → 200
   PUT /api/v1/trips/:id → 200
   PATCH /api/v1/trips/:id → 200
   
   # 201 Created for POST
   POST /api/v1/trips → 201 + Location header
   
   # 204 No Content for DELETE
   DELETE /api/v1/trips/:id → 204 (empty body)
   ```

2. Test client error codes:
   ```bash
   # 400 for malformed JSON
   POST /api/v1/trips with invalid JSON → 400
   
   # 401 for missing auth
   GET /api/v1/trips without token → 401
   
   # 403 for insufficient permissions
   DELETE /api/v1/trips/:id as partner → 403
   
   # 404 for nonexistent resource
   GET /api/v1/trips/nonexistent → 404
   
   # 409 for version conflict
   PUT /api/v1/trips/:id with stale version → 409
   
   # 422 for validation failure
   POST /api/v1/trips with empty name → 422
   
   # 429 for rate limit exceeded
   100+ requests in 1 minute → 429
   ```

3. Verify NO anti-patterns:
   ```bash
   # ❌ FAIL: 200 with error payload
   POST /api/v1/trips with validation error → 200 { "success": false, "error": "..." }
   # This violates standards; should be 422
   ```

**Expected Outcome**: Status codes match HTTP semantics; no 200 with error payloads

**Validation Method**: Integration tests asserting status codes for each scenario

---

### Scenario 7: Pagination

**Objective**: Verify that list endpoints support pagination with correct parameters and metadata

**Steps**:

1. Test default pagination:
   ```bash
   curl -X GET http://localhost:8080/api/v1/trips \
     -H "Authorization: Bearer $JWT"
   
   # Expected: page=1, per_page=20 (defaults)
   {
     "data": [...],
     "pagination": {
       "page": 1,
       "per_page": 20,
       "total": 157,
       "total_pages": 8,
       "has_next": true,
       "has_prev": false
     }
   }
   ```

2. Test custom pagination:
   ```bash
   curl -X GET "http://localhost:8080/api/v1/trips?page=3&per_page=10" \
     -H "Authorization: Bearer $JWT"
   
   # Expected: page=3, per_page=10
   {
     "data": [...],
     "pagination": {
       "page": 3,
       "per_page": 10,
       "total": 157,
       "total_pages": 16,
       "has_next": true,
       "has_prev": true
     }
   }
   ```

3. Test maximum per_page validation:
   ```bash
   curl -X GET "http://localhost:8080/api/v1/trips?per_page=200" \
     -H "Authorization: Bearer $JWT"
   
   # Expected: 400 Bad Request
   {
     "error": "invalid_request",
     "message": "per_page cannot exceed 100",
     "request_id": "..."
   }
   ```

4. Test page beyond total_pages:
   ```bash
   curl -X GET "http://localhost:8080/api/v1/trips?page=999" \
     -H "Authorization: Bearer $JWT"
   
   # Expected: 200 OK with empty data array (not 404)
   {
     "data": [],
     "pagination": {
       "page": 999,
       "per_page": 20,
       "total": 157,
       "total_pages": 8,
       "has_next": false,
       "has_prev": true
     }
   }
   ```

**Expected Outcome**: All list endpoints paginate; defaults apply; max 100 per_page; beyond range returns empty array

**Validation Method**: Integration tests for each list endpoint

---

### Scenario 8: Filtering and Sorting

**Objective**: Verify that list endpoints support filtering and sorting with documented syntax

**Steps**:

1. Test simple filtering:
   ```bash
   curl -X GET "http://localhost:8080/api/v1/trips?status=completed" \
     -H "Authorization: Bearer $JWT"
   
   # Expected: Only trips with status=completed
   ```

2. Test range filtering:
   ```bash
   curl -X GET "http://localhost:8080/api/v1/trips?start_date[gte]=2026-08-01&start_date[lte]=2026-08-31" \
     -H "Authorization: Bearer $JWT"
   
   # Expected: Only trips starting in August 2026
   ```

3. Test text search:
   ```bash
   curl -X GET "http://localhost:8080/api/v1/trips?q=barcelona" \
     -H "Authorization: Bearer $JWT"
   
   # Expected: Trips matching "barcelona" in name or description
   ```

4. Test multiple values (OR):
   ```bash
   curl -X GET "http://localhost:8080/api/v1/trips?status=pending,approved" \
     -H "Authorization: Bearer $JWT"
   
   # Expected: Trips with status pending OR approved
   ```

5. Test sorting (ascending):
   ```bash
   curl -X GET "http://localhost:8080/api/v1/trips?sort=created_at" \
     -H "Authorization: Bearer $JWT"
   
   # Expected: Trips sorted by created_at ascending
   ```

6. Test sorting (descending):
   ```bash
   curl -X GET "http://localhost:8080/api/v1/trips?sort=-created_at" \
     -H "Authorization: Bearer $JWT"
   
   # Expected: Trips sorted by created_at descending (newest first)
   ```

7. Test multi-field sorting:
   ```bash
   curl -X GET "http://localhost:8080/api/v1/trips?sort=status,-created_at" \
     -H "Authorization: Bearer $JWT"
   
   # Expected: Sorted by status ascending, then created_at descending
   ```

8. Test invalid filter field:
   ```bash
   curl -X GET "http://localhost:8080/api/v1/trips?nonexistent_field=value" \
     -H "Authorization: Bearer $JWT"
   
   # Expected: 400 Bad Request
   {
     "error": "invalid_request",
     "message": "Invalid filter field: nonexistent_field",
     "request_id": "..."
   }
   ```

**Expected Outcome**: Filtering and sorting work as documented; invalid fields return 400

**Validation Method**: Integration tests for each filter/sort scenario

---

### Scenario 9: Rate Limiting

**Objective**: Verify that rate limiting is enforced and communicated via headers

**Steps**:

1. Test rate limit headers in normal response:
   ```bash
   curl -X GET http://localhost:8080/api/v1/trips \
     -H "Authorization: Bearer $JWT" \
     -v
   
   # Expected headers:
   X-RateLimit-Limit: 100
   X-RateLimit-Remaining: 99
   X-RateLimit-Reset: 1720270800
   ```

2. Test rate limit exceeded:
   ```bash
   # Send 101 requests in 60 seconds
   for i in {1..101}; do
     curl -X GET http://localhost:8080/api/v1/trips \
       -H "Authorization: Bearer $JWT"
   done
   
   # Expected on 101st request:
   HTTP/1.1 429 Too Many Requests
   Retry-After: 23
   
   {
     "error": "rate_limit_exceeded",
     "message": "You have exceeded the rate limit. Please try again in 23 seconds.",
     "request_id": "...",
     "retry_after": 23
   }
   ```

3. Test rate limit headers after exceeding:
   ```bash
   curl -X GET http://localhost:8080/api/v1/trips \
     -H "Authorization: Bearer $JWT" \
     -v
   # (after rate limit exceeded)
   
   # Expected headers:
   X-RateLimit-Limit: 100
   X-RateLimit-Remaining: 0
   X-RateLimit-Reset: 1720270800
   ```

4. Test rate limit reset:
   ```bash
   # Wait until X-RateLimit-Reset timestamp
   sleep 60
   
   curl -X GET http://localhost:8080/api/v1/trips \
     -H "Authorization: Bearer $JWT"
   
   # Expected: 200 OK with refreshed rate limit
   X-RateLimit-Limit: 100
   X-RateLimit-Remaining: 99
   X-RateLimit-Reset: 1720270860
   ```

**Expected Outcome**: All responses include rate limit headers; 101st request returns 429; limits reset after window

**Validation Method**: Integration tests simulating high request volume

---

### Scenario 10: Timestamps and Dates

**Objective**: Verify that all timestamps use ISO 8601 format with UTC timezone

**Steps**:

1. Test date format:
   ```bash
   curl -X GET http://localhost:8080/api/v1/trips/:id \
     -H "Authorization: Bearer $JWT"
   
   # Expected date format:
   {
     "start_date": "2026-08-01",      # ✅ ISO 8601 date
     "end_date": "2026-08-03"         # ✅ ISO 8601 date
   }
   ```

2. Test timestamp format:
   ```bash
   # Expected timestamp format:
   {
     "created_at": "2026-07-06T10:00:00Z",    # ✅ ISO 8601 with UTC (Z)
     "updated_at": "2026-07-06T14:30:15Z"     # ✅ ISO 8601 with UTC (Z)
   }
   ```

3. Test that non-ISO formats are rejected:
   ```bash
   curl -X POST http://localhost:8080/api/v1/trips \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer $JWT" \
     -d '{
       "name": "Test Trip",
       "start_date": "08/01/2026"   # ❌ US date format
     }'
   
   # Expected: 422 Unprocessable Entity
   {
     "error": "validation_failed",
     "message": "Invalid date format. Use ISO 8601 (YYYY-MM-DD)",
     "request_id": "...",
     "fields": [
       {
         "field": "start_date",
         "error": "Invalid date format. Expected YYYY-MM-DD"
       }
     ]
   }
   ```

**Expected Outcome**: All dates use YYYY-MM-DD, all timestamps use ISO 8601 with Z suffix

**Validation Method**: Integration tests asserting response formats + request validation tests

---

## Summary of Validation Coverage

| Standard | Validation Method | Scenario |
|----------|------------------|----------|
| Resource Naming | Manual review + linter | Scenario 1 |
| URL Nesting | Route inspection + tests | Scenario 2 |
| API Versioning | Integration tests | Scenario 3 |
| Request/Response Format | Integration tests | Scenario 4 |
| Error Format | Integration tests (all error codes) | Scenario 5 |
| HTTP Status Codes | Integration tests (success + errors) | Scenario 6 |
| Pagination | Integration tests (all list endpoints) | Scenario 7 |
| Filtering/Sorting | Integration tests (valid + invalid) | Scenario 8 |
| Rate Limiting | Load tests | Scenario 9 |
| Timestamps | Integration tests (format validation) | Scenario 10 |

---

## Automation

### Integration Test Suite

Create integration tests for each scenario:

```go
// backend/tests/integration/api_standards_test.go

func TestAPIStandards_ErrorFormat(t *testing.T) {
    // Scenario 5: Error Response Format
    resp := makeRequest(t, "POST", "/api/v1/trips", invalidPayload)
    
    assert.Equal(t, 422, resp.StatusCode)
    
    var errResp ErrorResponse
    json.Unmarshal(resp.Body, &errResp)
    
    assert.NotEmpty(t, errResp.Error)        // machine-readable code
    assert.NotEmpty(t, errResp.Message)      // human-readable
    assert.NotEmpty(t, errResp.RequestID)    // correlation ID
    assert.NotEmpty(t, errResp.Fields)       // field-level errors
}

func TestAPIStandards_Pagination(t *testing.T) {
    // Scenario 7: Pagination
    resp := makeRequest(t, "GET", "/api/v1/trips?page=2&per_page=10")
    
    var listResp ListResponse
    json.Unmarshal(resp.Body, &listResp)
    
    assert.NotNil(t, listResp.Pagination)
    assert.Equal(t, 2, listResp.Pagination.Page)
    assert.Equal(t, 10, listResp.Pagination.PerPage)
    assert.NotZero(t, listResp.Pagination.Total)
}
```

### Linter Rules

Where possible, automate checks:

- URL structure: Custom linter rule checking route definitions
- Field naming: JSON tag validation (snake_case)
- Error format: Linter ensuring all error returns use standard format

---

## Completion Criteria

✅ All 10 validation scenarios pass  
✅ Integration test suite covers all standards  
✅ Existing endpoints (from specs 001, 004) comply or are documented as exceptions  
✅ PR template includes API standards checklist  
✅ Code review process references this document

---

## Next Steps After Validation

1. Promote standards to `docs/api-design-standards.md` (make it project-wide reference)
2. Update PR template to include API standards compliance section
3. Add integration tests to CI pipeline (enforce standards automatically)
4. Create linter rules for automatable checks
5. Conduct team training session on API standards

---

**Document Version**: 1.0.0  
**Last Updated**: 2026-07-06
