# API Contracts: Non-Functional Requirements and System Constraints

**Date**: 2026-07-02 | **Plan**: [plan.md](../plan.md) | **Data Model**: [data-model.md](../data-model.md)

This file documents the contracts introduced by the NFR infrastructure layer. These are
*cross-cutting* contracts that apply to every endpoint in the system, plus the dedicated
health-check endpoint. Application domain endpoints (trips, itineraries, auth) are
documented in their respective feature spec contracts.

---

## Cross-Cutting Headers

The following headers are present on **every** HTTP request/response processed by the backend.

### Request Header — `X-Request-ID`

| Property | Value |
|----------|-------|
| **Header name** | `X-Request-ID` |
| **Direction** | Request (optional from client) → Response (always present) |
| **Format** | UUID v4 string, e.g., `a1b2c3d4-e5f6-7890-abcd-ef1234567890` |
| **Behaviour** | If the client sends `X-Request-ID`, the server propagates it unchanged. If absent, the server generates a new UUID v4 and uses it as the correlation ID. The correlation ID is always echoed in the response header and embedded in the structured log entry for the request. |
| **NFR** | NFR-OBS-003 |

**Example**:

```
# Client may optionally send:
X-Request-ID: a1b2c3d4-e5f6-7890-abcd-ef1234567890

# Server always returns (generated or echoed):
X-Request-ID: a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

---

## Endpoints

### `GET /healthz`

Returns the health status of the backend service. Consumed by load balancers, uptime monitors,
container orchestrators, and the CI health-check test.

**Auth required**: No — unauthenticated endpoint.

**Request**:

```
GET /healthz HTTP/1.1
Host: <service-host>
```

No request body. No required query parameters.

**Response — 200 OK (healthy)**:

```json
{
  "status": "ok",
  "version": "0.1.0",
  "uptime_seconds": 3742.1
}
```

**Response — 200 OK (degraded, non-critical dependency unavailable)**:

```json
{
  "status": "degraded",
  "version": "0.1.0",
  "uptime_seconds": 3742.1
}
```

> **Note**: The HTTP status code is `200 OK` in both cases. A non-`2xx` code is reserved for
> cases where the service itself cannot respond (e.g., process crash). The `"degraded"` value
> in the body signals operators without causing load-balancer routing changes.

**Response headers**:

| Header | Value |
|--------|-------|
| `Content-Type` | `application/json` |
| `X-Request-ID` | Correlation ID (generated for this request) |
| `Cache-Control` | `no-store` |

**Performance requirement**: Response MUST be returned within **100 ms** (NFR-OBS-002).

**NFRs enforced**: NFR-OBS-002.

---

### Prompt Rejection — `400 Bad Request`

Returned by any AI-generation endpoint when the user's input fails the server-side prompt
validation layer. This is not a standalone endpoint — it is the error response schema
emitted by the prompt-validation middleware when applied to any route that accepts AI input.

**Trigger**: The incoming prompt matches one or more rules in `backend/config/prompt-rules.yml`
(NFR-SEC-007).

**Response body**:

```json
{
  "error": "invalid_prompt",
  "message": "Your request could not be processed. Please describe your travel plans and try again.",
  "request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

| Field | Type | Notes |
|-------|------|-------|
| `error` | `string` | Fixed machine-readable code: `"invalid_prompt"` |
| `message` | `string` | User-facing message; MUST NOT reveal which rule was matched or the matched pattern |
| `request_id` | `string` | UUID v4 correlation ID; matches `X-Request-ID` response header |

**Response headers**:

| Header | Value |
|--------|-------|
| `Content-Type` | `application/json` |
| `X-Request-ID` | Correlation ID for incident investigation |

**Server-side log entry** (not visible to client, WARN level):

```json
{
  "time": "2026-07-02T14:30:00.000Z",
  "level": "WARN",
  "service": "traiveler-api",
  "request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "method": "POST",
  "path": "/api/v1/trips/generate",
  "status": 400,
  "duration_ms": 2.1,
  "msg": "prompt rejected by validation layer",
  "error": "rule matched: instruction-override"
}
```

**NFRs enforced**: NFR-SEC-007.

---

### Rate-Limit Fast-Fail — `503 Service Unavailable`

Returned when the AI provider returns a rate-limit (429) response. The backend does not queue
the request; it fast-fails immediately with a user-actionable message.

**Response body**:

```json
{
  "error": "ai_unavailable",
  "message": "The AI service is busy. Please try again in a moment.",
  "request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

**Response headers**:

| Header | Value |
|--------|-------|
| `Content-Type` | `application/json` |
| `X-Request-ID` | Correlation ID |
| `Retry-After` | Number of seconds the client should wait before retrying (sourced from the AI provider's response, or defaulting to `30`) |

**NFRs enforced**: NFR-SCALE-003 (graceful load shedding), NFR-PERF-002 (AI acknowledgement ≤ 3 s).

---

## Structured Log Contract

The structured log format is an *inter-service observability contract*. Any log aggregation
tool, alerting rule, or ops runbook that references log fields MUST use the field names
defined in the [StructuredLogEntry](../data-model.md#structuredlogentry) schema.

**Field name stability guarantee**: Field names in `StructuredLogEntry` MUST NOT be renamed
without a corresponding update to all alerting rules and runbooks that reference them.
Renaming constitutes a breaking change to the observability contract.

**Alerting rule reference** — error rate threshold (NFR-OBS-004):

```
ALERT condition: count(status >= 500) / count(all requests) > 0.01
  OVER rolling 5-minute window
  USING field: "status" from StructuredLogEntry
  ROUTE TO: on-call engineer
```

---

## Accessibility Contract

All HTML pages served by the frontend MUST satisfy the following contract before any PR
is merged:

| Check | Tool | Threshold | Enforcement |
|-------|------|-----------|-------------|
| WCAG 2.1 AA violations | `@axe-core/playwright` | Zero violations | `accessibility.yml` CI gate blocks PR merge |
| Lighthouse accessibility score | `@lhci/cli` | ≥ 90 | `accessibility.yml` CI gate blocks PR merge |
| Lighthouse LCP | `@lhci/cli` | ≤ 2.5 s | `accessibility.yml` CI gate blocks PR merge |
| Lighthouse CLS | `@lhci/cli` | ≤ 0.1 | `accessibility.yml` CI gate blocks PR merge |
| Lighthouse INP | `@lhci/cli` | ≤ 200 ms | `accessibility.yml` CI gate blocks PR merge |

**Configuration reference**: Lighthouse CI thresholds are defined in `lighthouserc.yml` at
the repo root. The accessibility Playwright helper is defined in
`frontend/tests/helpers/a11y.ts`.
