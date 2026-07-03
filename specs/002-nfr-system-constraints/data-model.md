# Data Model: Non-Functional Requirements and System Constraints

**Date**: 2026-07-02 | **Plan**: [plan.md](plan.md) | **Spec**: [spec.md](spec.md)

Entities derived from the NFR infrastructure requirements. These are not application domain
entities (users, itineraries) — those are defined in spec 001. The entities here are the
*operational schemas* that the NFR infrastructure must produce and consume consistently
across the entire application.

---

## Entity Map

```
HTTP Request
    │
    ▼
RequestIDMiddleware ──► StructuredLogEntry (emitted per request, NFR-OBS-001/003)
    │
    ▼
PromptValidationLayer
    ├── [PASS] ──► AI Provider
    └── [FAIL] ──► PromptRejectionResponse (NFR-SEC-007)

GET /healthz ──► HealthCheckResponse (NFR-OBS-002)

AI Provider Response
    └── OutputSanitizer ──► sanitised string (stored in DB, NFR-SEC-008)

PrivacyPolicyPage (static, linked from registration, NFR-PRIV-003)
```

---

## Entities

### StructuredLogEntry

Every HTTP request processed by the backend produces exactly one JSON log line on stdout.
This is the canonical format for all server-side observability. Log aggregation tools (e.g.,
CloudWatch Logs, Loki) ingest these lines and index the fields.

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `timestamp` | `string` (RFC 3339) | ✅ | UTC; emitted by `slog.NewJSONHandler` as `"time"` key |
| `level` | `string` | ✅ | `"INFO"`, `"WARN"`, `"ERROR"` — standard slog levels |
| `service` | `string` | ✅ | Fixed value `"traiveler-api"`; injected at logger initialisation |
| `request_id` | `string` (UUID v4) | ✅ | Correlation ID; matches `X-Request-ID` response header |
| `method` | `string` | ✅ | HTTP method: `"GET"`, `"POST"`, etc. |
| `path` | `string` | ✅ | URL path without query string, e.g., `"/api/v1/trips"` |
| `status` | `int` | ✅ | HTTP response status code |
| `duration_ms` | `float64` | ✅ | Request processing time in milliseconds |
| `msg` | `string` | ✅ | Human-readable summary, e.g., `"request completed"` |
| `error` | `string` | ❌ | Present only on `WARN`/`ERROR` entries; Go error message |
| `user_id` | `string` (UUID v4) | ❌ | Present when the request is authenticated; never logged for unauthenticated requests |

**Example log line** (pretty-printed for readability; actual output is single-line JSON):

```json
{
  "time": "2026-07-02T14:22:05.123Z",
  "level": "INFO",
  "service": "traiveler-api",
  "request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "method": "POST",
  "path": "/api/v1/trips/generate",
  "status": 202,
  "duration_ms": 187.4,
  "msg": "request completed"
}
```

**Validation rules**:
- `request_id` MUST be a valid UUID v4 string.
- `duration_ms` MUST be ≥ 0.
- `user_id` MUST NOT appear in log entries for unauthenticated requests (privacy constraint, NFR-PRIV-002).
- Stack traces and raw error messages MUST NOT appear in `"ERROR"` log entries delivered to end-users; they are logged server-side only and never surfaced in API responses.

---

### HealthCheckResponse

The response body returned by `GET /healthz`. Consumed by load balancers, uptime monitors, and
Kubernetes-style readiness probes.

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `status` | `string` | ✅ | `"ok"` when the service is healthy; `"degraded"` if a non-critical dependency is unavailable |
| `version` | `string` | ✅ | Semantic version of the running binary, e.g., `"0.1.0"` |
| `uptime_seconds` | `float64` | ✅ | Seconds since the process started |

**Example response**:

```json
{
  "status": "ok",
  "version": "0.1.0",
  "uptime_seconds": 3742.1
}
```

**Validation rules**:
- The endpoint MUST respond within 100 ms (NFR-OBS-002).
- HTTP status code is always `200 OK` even when `status` is `"degraded"` — load balancers use
  the HTTP code to determine routing; `"degraded"` is an informational signal for operators only.
- `uptime_seconds` MUST be ≥ 0.

---

### PromptValidationRule

A single entry in the server-side deny-list configuration file
(`backend/config/prompt-rules.yml`). The validation layer loads all rules at startup and
evaluates every incoming AI prompt against each rule before forwarding to the provider.

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `id` | `string` | ✅ | Unique kebab-case identifier, e.g., `"instruction-override"` |
| `description` | `string` | ✅ | Human-readable explanation of what this rule blocks |
| `pattern` | `string` | ✅ | Case-insensitive substring or regex pattern matched against the normalised prompt text |
| `match_type` | `string` | ✅ | `"substring"` or `"regex"` |
| `enabled` | `bool` | ✅ | If `false`, the rule is loaded but not evaluated (for staged rollout) |

**Example config snippet** (`prompt-rules.yml`):

```yaml
rules:
  - id: instruction-override
    description: Blocks attempts to override the system prompt via user input
    pattern: "ignore (previous|prior|all) instructions"
    match_type: regex
    enabled: true

  - id: system-prompt-extraction
    description: Blocks attempts to read the system prompt
    pattern: "repeat your (system|initial) prompt"
    match_type: regex
    enabled: true

  - id: role-switching
    description: Blocks attempts to reassign the assistant's role
    pattern: "you are now"
    match_type: substring
    enabled: true
```

**Validation rules**:
- `id` values MUST be unique within the file.
- `match_type` MUST be one of `"substring"` or `"regex"`.
- If `match_type` is `"regex"`, the pattern MUST be a valid Go `regexp` expression (validated
  at startup; invalid patterns cause the server to refuse to start).

---

### PromptRejectionResponse

The JSON body returned with HTTP `400 Bad Request` when a prompt fails validation
(NFR-SEC-007). Designed to be informative to the user without disclosing which specific rule
was triggered (to prevent adversarial rule-learning).

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `error` | `string` | ✅ | Fixed value: `"invalid_prompt"` |
| `message` | `string` | ✅ | User-facing message; MUST NOT reveal rule details or matched pattern |
| `request_id` | `string` (UUID v4) | ✅ | Correlation ID for incident investigation |

**Example response**:

```json
{
  "error": "invalid_prompt",
  "message": "Your request could not be processed. Please describe your travel plans and try again.",
  "request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

**Validation rules**:
- `message` MUST NOT contain the matched pattern, rule ID, or any information that reveals
  which specific check was triggered.
- The rejection event MUST produce a `WARN`-level `StructuredLogEntry` with an `error` field
  containing the matched rule ID (logged server-side, never returned to the client).

---

## State Transitions

### Prompt Lifecycle

```
User Input
    │
    ▼
[Prompt Validation Layer]
    │
    ├── RULE MATCH → PromptRejectionResponse (400) + WARN log entry
    │
    └── NO MATCH → [AI Provider Request]
                        │
                        ▼
                   [AI Response Received]
                        │
                        ▼
                   [Output Sanitiser (bluemonday)]
                        │
                        ▼
                   Sanitised Content → [Database Storage] + [API Response]
```

### Request Log Lifecycle

```
Incoming HTTP Request
    │
    ▼
[RequestID Middleware] — generates UUID v4 if no X-Request-ID header present
    │
    ▼
[Logger Middleware] — starts timer, attaches logger-with-request-id to context
    │
    ▼
[Handler / Downstream Middleware]
    │
    ▼
[Logger Middleware] — records status code, duration; emits StructuredLogEntry
    │
    ▼
Response written — X-Request-ID header set on response
```
