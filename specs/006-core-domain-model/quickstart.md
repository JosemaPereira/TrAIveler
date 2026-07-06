# Quickstart: Domain Model Validation Guide

**Feature**: Core Domain and Data Model Foundations  
**Date**: 2026-07-06  
**Purpose**: Developer reference guide with validation scenarios for testing domain model understanding

## Overview

This guide provides runnable validation scenarios that prove your understanding of the core domain model. Each scenario can be validated by consulting [data-model.md](data-model.md) and verifying the expected behavior matches the documented constraints.

---

## Validation Scenarios

### Scenario 1: User Account Creation with Validation Layers

**Goal**: Understand three-layer validation (DB, Logic, API) for User entity

**Test**:
1. Attempt to create user with email `"ADMIN@EXAMPLE.COM"`
2. Attempt to create second user with email `"admin@example.com"` (lowercase)

**Expected Outcome**:
- First user created successfully
- Second user rejected with unique constraint violation
- Validation layers triggered:
  - [API] Email format validated before reaching backend (RFC 5322 pattern)
  - [Logic] Password hashed with bcrypt cost 12+ before storage
  - [DB] UNIQUE constraint on LOWER(email) prevents duplicate (case-insensitive)

**Consult**: [data-model.md](data-model.md#user) — User validation rules and indexes

---

### Scenario 2: Trip Concurrent Modification Conflict

**Goal**: Understand optimistic locking with version counter on Trip entity

**Test**:
1. Admin retrieves trip with version=5
2. Admin A updates trip title (sends If-Match: 5)
3. Admin B simultaneously updates description (sends If-Match: 5)
4. First update succeeds, increments version to 6
5. Second update attempts to match version=5

**Expected Outcome**:
- Admin A: 200 OK, trip version now 6
- Admin B: 409 Conflict with response body `{id, version: 6, updated_at}`
- Admin B must refetch trip (version 6) and retry update with If-Match: 6

**Consult**: [data-model.md](data-model.md#trip) — Trip concurrency control section

---

### Scenario 3: Cascade Delete Across Multiple Levels

**Goal**: Understand CASCADE behavior when deleting Trip entity

**Test**:
1. Create Trip with 3 Days
2. Each Day has 2 Activities
3. Trip has 1 Collaborator and 2 Suggestions
4. Delete Trip

**Expected Outcome**:
- All 3 Days CASCADE deleted
- All 6 Activities CASCADE deleted (via Day → CASCADE)
- Collaborator record CASCADE deleted
- All 2 Suggestions CASCADE deleted
- ConversationSessions (if any) CASCADE deleted
- ConversationMessages CASCADE deleted (via ConversationSession → CASCADE)
- Destination records NOT deleted (Day → Destination uses SET NULL, not CASCADE)

**Consult**: [data-model.md](data-model.md#trip) — Trip cascade behavior section

---

### Scenario 4: Forward-Only State Transition (Subscription)

**Goal**: Understand immutable state transitions on Subscription entity

**Test**:
1. Create subscription with status='stub_pending'
2. Stub checkout completes, transition to 'active'
3. User cancels subscription, transition to 'cancelled'
4. User wants to reactivate (attempt transition 'cancelled' → 'active')

**Expected Outcome**:
- Transitions stub_pending → active: ✅ Allowed
- Transition active → cancelled: ✅ Allowed
- Transition cancelled → active: ❌ Blocked (forward-only rule)
- User must create new Subscription record to resume service

**Consult**: [data-model.md](data-model.md#subscription) — Subscription state transitions section

---

### Scenario 5: Partner Collaborator Suggestion Workflow

**Goal**: Understand suggest-then-approve collaboration model

**Test**:
1. Admin creates Trip
2. Admin invites Partner (creates Collaborator record)
3. Partner submits Suggestion (action='edit', target_type='activity', status='pending')
4. Partner attempts to edit Activity directly
5. Admin approves Suggestion

**Expected Outcome**:
- Collaborator record created with accepted_at=NULL (pending invitation)
- Suggestion created with status='pending'
- Direct Activity edit rejected (Partner role lacks permission)
- Admin approval triggers:
  - Suggestion status → 'approved'
  - Target Activity fields updated per payload
  - Activity.version incremented
  - Suggestion.reviewed_by = admin user_id, reviewed_at = NOW()

**Consult**: [data-model.md](data-model.md#collaborator), [data-model.md](data-model.md#suggestion)

---

### Scenario 6: Plan Limit Enforcement (Basic Plan)

**Goal**: Understand Basic plan collaborator limits

**Test**:
1. Admin creates Trip
2. Admin invites Partner A (creates Collaborator 1)
3. Admin invites Partner B (attempts to create Collaborator 2)

**Expected Outcome**:
- Collaborator 1 created successfully
- Collaborator 2 creation blocked by service layer (Basic plan: max 1 partner per trip)
- Error message: "Plan limit reached: Basic plan allows 1 partner collaborator per trip"

**Consult**: [data-model.md](data-model.md#plan), [data-model.md](data-model.md#collaborator) — Business rules section

---

### Scenario 7: User Account Deletion with PII Removal

**Goal**: Understand GDPR-compliant account deletion cascade behavior

**Test**:
1. User has active RefreshTokens, Trips, Collaborations, SecurityEvents
2. User requests account deletion

**Expected Outcome** (Immediate):
- All RefreshTokens CASCADE deleted (tokens invalidated)
- All Collaborator records CASCADE deleted (partnership invitations removed)
- All Suggestion records (as author) CASCADE deleted
- SecurityEvent.user_id SET NULL (audit log preserved, user anonymized)
- User.email, User.password_hash marked for PII removal

**Expected Outcome** (Within 30 days):
- User PII fields (email, password_hash) removed via background job
- User.id retained for referential integrity (anonymized)
- Trips: Must be transferred or explicitly deleted before account deletion (RESTRICT constraint blocks delete)

**Consult**: [data-model.md](data-model.md#user) — Cascade behavior and invariant #16-17

---

### Scenario 8: Destination Geographic Data for Mapping

**Goal**: Understand Destination entity coordinates and regional metadata

**Test**:
1. Create Destination with name="Tokyo", country="JP", latitude=35.6762, longitude=139.6503
2. Query Days by destination_id
3. Calculate distance between two Destinations using coordinates

**Expected Outcome**:
- Destination created with valid coordinates (DB CHECK constraint: lat [-90, +90], lon [-180, +180])
- Country validated as ISO 3166-1 alpha-2 code at API layer
- Days can reference Destination for location context
- Coordinates enable:
  - Map visualization (pin placement)
  - Distance calculations (haversine formula)
  - Regional filtering (country/region queries)
- No schema migration required for future mapping features

**Consult**: [data-model.md](data-model.md#destination) — Validation rules and business rules

---

### Scenario 9: Activity Sequence Ordering Within Day

**Goal**: Understand Activity sequence_order and display logic

**Test**:
1. Create Day with 4 Activities (sequence_order: 1, 2, 3, 4)
2. Delete Activity with sequence_order=2
3. Query remaining Activities
4. Insert new Activity with sequence_order=5

**Expected Outcome**:
- Activities returned in sequence_order ascending: [1, 3, 4, 5]
- Gaps in sequence allowed (no automatic reordering on delete)
- Frontend can:
  - Display activities in sequence_order
  - Implement drag-and-drop reordering (update sequence_order values)
  - Allow gaps or require contiguous numbering (business logic decision)
- sequence_order is display order, not structural constraint

**Consult**: [data-model.md](data-model.md#activity) — sequence_order attribute and business rules

---

### Scenario 10: Conversation Session Token Tracking

**Goal**: Understand ConversationSession token accumulation for cost analysis

**Test**:
1. Create ConversationSession for Trip (status='in_progress', total_tokens=0)
2. Add ConversationMessage (role='user', content='plan 5-day Tokyo trip', token_count=8)
3. Add ConversationMessage (role='assistant', content='<AI response>', token_count=450)
4. Add ConversationMessage (role='user', content='add Kyoto', token_count=4)
5. Add ConversationMessage (role='assistant', content='<AI response>', token_count=520)
6. Mark session as completed

**Expected Outcome**:
- ConversationSession.total_tokens = 982 (8 + 450 + 4 + 520)
- Token count used for:
  - Cost analysis (tokens * provider rate)
  - Usage metrics per trip
  - Session abandonment threshold (high token count = expensive session)
- Session status transitions: in_progress → completed (forward-only)

**Consult**: [data-model.md](data-model.md#conversationsession), [data-model.md](data-model.md#conversationmessage)

---

### Scenario 11: Security Event Audit Trail with User Anonymization

**Goal**: Understand SecurityEvent retention and user_id handling after account deletion

**Test**:
1. User triggers auth_login_failure event (SecurityEvent created with user_id)
2. Same user triggers validation_prompt_injection event
3. User deletes account
4. Query SecurityEvents for deleted user

**Expected Outcome**:
- SecurityEvents persist after user deletion (not CASCADE deleted)
- SecurityEvent.user_id SET NULL (user anonymized, but event preserved)
- Correlation_id, event_type, severity, timestamp, ip_address preserved
- Audit trail intact for 30 days (CloudWatch Logs retention)
- Use correlation_id (not user_id) for request tracing after account deletion

**Consult**: [data-model.md](data-model.md#securityevent) — Cascade behavior and invariant #23

---

### Scenario 12: Multi-Key JWT Rotation (Zero Downtime)

**Goal**: Understand JWTSigningKey rotation strategy with multiple active keys

**Test**:
1. System starts with key-A (status='active')
2. Generate key-B, insert with status='active'
3. Update application config: sign new tokens with key-B
4. Existing tokens signed with key-A remain valid
5. After 30 days, set key-A status='retired'

**Expected Outcome**:
- During overlap period:
  - New tokens signed with key-B
  - Tokens signed with key-A still validate (both keys status='active')
  - Zero service disruption
- After retirement:
  - Tokens signed with key-A rejected (key-A status='retired')
  - key-A record retained for audit only
- Backend retrieves private keys from AWS Secrets Manager at runtime (never from DB)

**Consult**: [data-model.md](data-model.md#jwtsigningkey) — Key rotation strategy section

---

## Quick Reference: Validation Layer Tags

All validation rules in [data-model.md](data-model.md) are tagged with enforcement layer:

- **[DB]**: Database constraint (CHECK, NOT NULL, UNIQUE, FK, ENUM) — Prevents data corruption at storage layer
- **[Logic]**: Service/domain layer business rule — Enforces complex domain invariants
- **[API]**: HTTP boundary input validation — Provides user-friendly error messages before backend processing

**Example**:
- [DB] `email` must be unique → UNIQUE constraint, database prevents duplicate insert
- [API] Email format validated → Returns 400 Bad Request with message "Invalid email format"
- [Logic] Password complexity checked → Returns 400 Bad Request with message "Password must contain uppercase, lowercase, and digit"

---

## Prerequisites for Implementation

Before implementing features based on this domain model:

1. ✅ Read [data-model.md](data-model.md) for complete entity specifications
2. ✅ Review [research.md](research.md) for gap analysis and design decisions
3. ✅ Consult invariants section for cross-entity business rules
4. ✅ Check cascade behavior to understand delete implications
5. ✅ Verify validation layer tags to avoid redundant validation
6. ✅ Review state transitions to implement forward-only state machines

---

## Expected Outcomes

After working through these scenarios, developers should be able to:

- [ ] Locate entity definitions without asking questions (SC-001)
- [ ] Understand foreign key relationships and cascade behavior (SC-002)
- [ ] Implement optimistic locking for concurrent modifications (SC-007)
- [ ] Enforce three-layer validation consistently (SC-008)
- [ ] Design schema migrations without constraint conflicts (SC-005)
- [ ] Handle state transitions with immutable audit trails (FR-003)

**Next Steps**: Proceed to implementation with `/speckit.tasks` to generate dependency-ordered tasks.
