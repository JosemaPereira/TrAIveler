# Security & Authorization

<!-- PROMOTED:security START -->
<!-- Generated from specs/001-product-vision-scope/spec.md, specs/002-nfr-system-constraints/spec.md, and specs/004-security-auth-model/spec.md -->
<!-- Last promoted: 2026-07-03 -->

## Authentication Model

All trip creation, editing, deletion, and sharing operations require authenticated user accounts. Authentication is implemented using JWT tokens stored in HTTP-only cookies.

**JWT Algorithm**: RS256 (RSA Signature with SHA-256) — asymmetric signing allows public key distribution for validation without exposing signing capability.

**Token Expiration**:
- Access tokens: 24 hours maximum lifetime
- Refresh tokens: 30 days maximum lifetime

**Token Storage**: HTTP-only, Secure, SameSite=Strict cookies to prevent XSS-based token theft and CSRF attacks.

**Password Security**: Bcrypt hashing with cost factor 12 (2^12 = 4096 iterations); never store plaintext passwords.

### JWT Key Rotation

The system supports zero-downtime JWT signing key rotation using a multi-key strategy:
- New tokens are signed with the current primary key
- Token validation accepts any active signing key (primary or previous keys within their validity period)
- Keys are stored in AWS Secrets Manager with metadata indicating active/retired status
- Rotation does not force user re-authentication; old tokens remain valid until natural expiration

### Password Change & Session Management

When users change their password, they can optionally invalidate all active sessions (all access and refresh tokens across all devices). This is presented as an explicit choice (e.g., checkbox: "Log out all other devices") to balance security with user experience.

## Authorization & Roles

The system enforces two distinct roles with specific permissions:

```mermaid
graph TB
    User[User]
    
    subgraph "Admin Role"
        A1[Full CRUD on own trips]
        A2[Approve/reject suggestions]
        A3[Invite 1 partner max]
        A4[Remove collaborators]
    end

    subgraph "Partner Role"
        P1[View shared trips]
        P2[Submit suggestions]
        P3[❌ Cannot edit directly]
        P4[❌ Cannot invite others]
    end

    User --> |Subscription owner| A1
    User --> |Invited collaborator| P1

    style A1 fill:#e8f5e9
    style A2 fill:#e8f5e9
    style A3 fill:#e8f5e9
    style A4 fill:#e8f5e9
    style P1 fill:#fff4e6
    style P2 fill:#fff4e6
    style P3 fill:#ffebee
    style P4 fill:#ffebee
```

### Admin Role

- Full CRUD (Create, Read, Update, Delete) operations on their own trips
- Exclusive authority to approve or reject partner suggestions
- Can invite up to one partner collaborator (basic plan limit)
- Can remove collaborators from trips

### Partner Role

- Can view trips they are invited to
- Can submit suggestions for modifications to shared itineraries
- **Cannot** directly apply changes to any itinerary item (suggestions must be approved by admin)
- Cannot invite additional collaborators

## Concurrency Control

**Optimistic Locking**: All admin trip modification operations (Update and Delete) use version numbers or timestamps to detect concurrent modifications. When a conflict is detected, the system returns `409 Conflict` with the current resource version, requiring the client to fetch the latest state and retry.

This prevents lost updates when multiple users (or an admin approving a suggestion while another admin edits) modify the same trip simultaneously.

## Subscription Plan Limits

**Basic Plan** (MVP scope):
- One admin user per subscription
- Maximum one partner collaborator per trip
- Full trip management features

## Data Protection

### Secrets Management

- All secrets (database passwords, API keys, tokens, certificates) MUST be stored in AWS Secrets Manager
- Secrets are never committed to code or stored in environment variables in Git
- Services retrieve secrets at runtime using IAM role-based authentication
- Secret rotation is supported without requiring redeployment

### PII Handling

The system is GDPR-aware and implements the following protections:

1. **Minimal Data Collection**: Only PII directly required by functional requirements is stored (email address, trip preferences)
2. **Right to Deletion**: Users can delete their account, removing all associated PII within 30 days
3. **Privacy Policy**: A privacy policy page is accessible from registration/login flows before any data collection
4. **Data in Transit**: All client-server and service-to-service communication uses TLS 1.2 or higher

### Input Validation & Sanitization

```mermaid
graph LR
    Input[User Input]
    
    subgraph "API Layer"
        V1[Schema Validation]
        V2[SQL Injection Check]
        V3[XSS Prevention]
        V4[Path Traversal Check]
    end

    subgraph "AI Prompt Layer"
        P1[Prompt Validation]
        P2[Instruction Override Check]
        P3[System Prompt Extraction Check]
        P4[Off-topic Detection]
    end

    subgraph "AI Response Layer"
        S1[HTML/Script Stripping]
        S2[Executable Content Filter]
        S3[Safe Rendering]
    end

    Input --> V1
    V1 --> V2
    V2 --> V3
    V3 --> V4
    V4 --> |To AI| P1
    P1 --> P2
    P2 --> P3
    P3 --> P4
    P4 --> |AI Response| S1
    S1 --> S2
    S2 --> S3
    S3 --> |Persist/Render|DB[(Database)]

    V1 -->|400 Bad Request| Reject1[Rejected]
    P1 -->|400 Bad Request| Reject2[Rejected]

    style V1 fill:#e8f5e9
    style V2 fill:#e8f5e9
    style V3 fill:#e8f5e9
    style V4 fill:#e8f5e9
    style P1 fill:#fff4e6
    style P2 fill:#fff4e6
    style P3 fill:#fff4e6
    style P4 fill:#fff4e6
    style S1 fill:#e1f5ff
    style S2 fill:#e1f5ff
    style S3 fill:#e1f5ff
    style Reject1 fill:#ffebee
    style Reject2 fill:#ffebee
```

**Prompt Injection Protection** (NFR-SEC-007):
- All user input sent to the AI provider passes through a prompt-validation layer
- Rejects requests containing:
  - Instruction-override patterns (e.g., "ignore previous instructions")
  - Attempts to extract system prompts or internal configuration
  - Off-topic prompts unrelated to travel planning
- Rejected requests return `400 Bad Request` with correlation ID logged for security review

**Output Sanitization** (NFR-SEC-008):
- All AI-generated content is sanitized before rendering in browser or persisting to database
- HTML/script tags and executable content are stripped or escaped
- No AI-generated content is executed as code or used as raw SQL/template values

**API Input Validation** (NFR-SEC-005):
- All API inputs validated and sanitized before use
- SQL injection, XSS, and path-traversal payloads rejected with `400 Bad Request`
- Security-focused integration tests verify injection attempts are blocked

### Dependency Security

- No critical or high-severity CVEs permitted in direct or transitive dependencies at release time
- Backend: `gosec` must report zero high-severity findings on every PR
- Frontend: `npm audit --audit-level=high` must pass before merge
- Dependencies: `govulncheck` (Go) and `npm audit` (frontend) run in CI

### Secret Scanning

- No secrets, API keys, or credentials may be committed to the repository
- `gitleaks` secret scanner runs on every commit in CI
- Pre-commit hook and PR gate enforce zero secrets detected

## Security Testing

### Automated Testing

- Prompt injection test suite covering known attack patterns runs on every PR
- XSS payload tests verify AI-generated content sanitization
- SQL injection, path traversal, and other OWASP Top 10 vectors tested in integration suite
- OWASP ZAP baseline scan runs on staging environment before each release

### Manual Review

- OWASP LLM Top 10 checklist reviewed before each release
- Security-focused code review for all changes touching authentication, authorization, or data handling
- Disaster recovery drill conducted before first release

## Security Logging & Monitoring

### Event Logging

All security-relevant events are logged to AWS CloudWatch Logs in structured JSON format:
- **Authentication events**: Successful logins, failed login attempts, token refresh, logout (with correlation ID and timestamp)
- **Authorization denials**: 403 responses (with user ID, role, requested operation, and timestamp)
- **Input validation failures**: Prompt injection, XSS, SQL injection attempts (with correlation ID, user ID, and attack pattern detected)

**Sensitive Data Protection**: Logs MUST NOT contain JWT tokens, passwords, refresh tokens, API keys, or PII (except user ID for correlation).

### Log Retention

CloudWatch Logs retention period: **30 days** for all security event logs to balance incident response needs with cost efficiency.

### Metrics & Alarms

CloudWatch metrics are emitted for:
- Authentication failures (rate)
- Authorization denials (rate)
- Prompt injection detections (count)
- API error rate (5xx responses)

**CloudWatch Alarms** trigger when:
- Authentication failure rate exceeds 100 failures/minute for 5 consecutive minutes
- Prompt injection detection rate exceeds 10 events/minute for 5 consecutive minutes

**Alarm Response Strategy (MVP)**: Alarms trigger notifications for manual review and response. No automated IP blocking or account suspension is performed to avoid false-positive service disruptions.

<!-- PROMOTED:security END -->

## Implementation Status Note: Secrets Management

The Secrets Management target design above (all secrets in AWS Secrets Manager, IAM role-based
runtime retrieval, rotation without redeployment) now has a corresponding Terraform module
(`infra/modules/secrets/`, Sprint 3, issue #89) plus the RDS module's own auto-generated database
credentials secret. Neither has been applied against a real AWS account yet — no Secrets Manager
secret actually exists — per the project's AWS-cost-avoidance policy. See
[infra/README.md](../infra/README.md) ("Secrets Population" section) for the current implementation
status and the manual population steps required after the first real `terraform apply`.
