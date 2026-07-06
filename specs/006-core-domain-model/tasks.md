# Tasks: Core Domain and Data Model Foundations

**Feature**: Core Domain and Data Model Foundations  
**Branch**: `006-core-domain-model`  
**Input**: Design documents from `specs/006-core-domain-model/`

## Implementation Strategy

This is a **documentation feature** that completes the canonical domain model reference by filling gaps in `docs/data-model.md`. Tasks focus on validation, review, and promotion rather than code implementation.

**MVP Scope**: User Story 1 (P1) — Core Entity Reference catalog is the minimum viable deliverable. Completing all P1 stories (US1-US3) provides a complete, production-ready domain model reference.

## Task Format

- **Checkbox**: `- [ ]` for incomplete tasks
- **Task ID**: Sequential (T001, T002, etc.)
- **[P] marker**: Task can run in parallel (no dependencies)
- **[Story] label**: US1, US2, US3 for user story tasks
- **File paths**: Exact paths to artifacts

---

## Phase 1: Setup & Validation Infrastructure

**Purpose**: Prepare documentation structure and validation tools

- [ ] T001 Verify specs/006-core-domain-model/ directory structure exists with spec.md, plan.md, research.md, data-model.md, quickstart.md, checklists/requirements.md
- [ ] T002 [P] Review research.md gap analysis and confirm all 6 gaps identified (missing entities, incomplete Destination, validation layers, indexes, state transitions, cascade behavior)
- [ ] T003 [P] Validate plan.md constitution check passed with no violations

**Checkpoint**: Documentation structure validated; ready for content review

---

## Phase 2: User Story 1 - Core Entity Reference (Priority: P1) 🎯 MVP

**Goal**: Complete entity catalog with all 16 entities having full attribute specifications, data types, nullability, and constraints

**Independent Test**: Developer can locate any entity definition and find complete attribute list with types, constraints, and default values

### Validation for User Story 1

- [ ] T004 [P] [US1] Verify all 16 entities documented in specs/006-core-domain-model/data-model.md (User, RefreshToken, JWTSigningKey, SecurityEvent, Plan, Subscription, Trip, Day, Activity, Collaborator, Suggestion, Destination, TravelStyle, TripTravelStyle, ConversationSession, ConversationMessage)
- [ ] T005 [P] [US1] Verify ConversationSession entity has complete definition (id, trip_id, started_at, completed_at, status, total_tokens, ai_provider, created_at attributes documented)
- [ ] T006 [P] [US1] Verify ConversationMessage entity has complete definition (id, session_id, role, content, token_count, timestamp attributes documented)
- [ ] T007 [P] [US1] Verify Destination entity includes geographic attributes (id, name, country, region, latitude, longitude, created_at attributes documented with coordinate constraints)
- [ ] T008 [US1] Confirm all 16 entities have data types specified for each attribute (UUID, VARCHAR with length, TEXT, INT, BIGINT, DECIMAL, TIMESTAMP, ENUM, BOOLEAN, JSONB)
- [ ] T009 [US1] Confirm all 16 entities have nullability specified for each attribute (NOT NULL or NULLABLE)
- [ ] T010 [US1] Confirm all 16 entities have constraint documentation (PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK constraints)
- [ ] T011 [US1] Validate spec.md acceptance scenarios 1-3 for US1 are met (User, Trip, Collaborator entities findable with complete definitions)

**Checkpoint**: All entities have complete attribute catalogs; developers can schema entities without ambiguity

---

## Phase 3: User Story 2 - Relationship and Constraint Understanding (Priority: P1)

**Goal**: Document all entity relationships with cardinality, foreign keys, and cascade behavior (CASCADE, SET NULL, RESTRICT)

**Independent Test**: Developer can trace any relationship and understand referential integrity rules and orphaned data prevention

### Validation for User Story 2

- [ ] T012 [P] [US2] Verify ERD in specs/006-core-domain-model/data-model.md includes all 16 entities with relationship lines showing cardinality (1:1, 1:many, many:many)
- [ ] T013 [P] [US2] Verify all foreign key attributes specify target table and column (e.g., `user_id` (UUID, FK → users.id))
- [ ] T014 [US2] Verify cascade behavior documented for all foreign keys in Cascade Behavior sections (User, Trip, Day, Activity, Collaborator, Suggestion, ConversationSession, ConversationMessage, Subscription, RefreshToken, SecurityEvent, TripTravelStyle)
- [ ] T015 [US2] Confirm Trip cascade behavior documents Days → CASCADE, Activities → CASCADE (via Day), Collaborator → CASCADE, Suggestion → CASCADE, ConversationSession → CASCADE
- [ ] T016 [US2] Confirm User cascade behavior documents RefreshToken → CASCADE, SecurityEvent → SET NULL, Trip → RESTRICT, Collaborator → CASCADE, Suggestion → CASCADE
- [ ] T017 [US2] Confirm unique constraints documented (User.email, RefreshToken.token_hash, Day (trip_id, day_number), Collaborator (trip_id, user_id), TripTravelStyle composite PK)
- [ ] T018 [US2] Validate spec.md acceptance scenarios 1-3 for US2 are met (Trip deletion cascades, User deletion cascade rules, Collaborator UNIQUE constraint documented)

**Checkpoint**: All relationships traceable; referential integrity rules clear

---

## Phase 4: User Story 3 - Business Rule Enforcement (Priority: P1)

**Goal**: Document business rules with three-layer validation tags ([DB], [Logic], [API]) and invariants across entities

**Independent Test**: Developer can identify where each validation rule should be enforced (database, service layer, or API boundary)

### Validation for User Story 3

- [ ] T019 [P] [US3] Verify all 16 entities have Validation Rules section with [DB], [Logic], [API] layer tags in specs/006-core-domain-model/data-model.md
- [ ] T020 [P] [US3] Verify User entity has validation rules tagged (email uniqueness [DB], password complexity [Logic], email format [API])
- [ ] T021 [P] [US3] Verify Trip entity has optimistic locking rules documented (version field, If-Match header requirement, 409 Conflict response)
- [ ] T022 [P] [US3] Verify Subscription entity has plan limit enforcement documented (Basic plan: 1 admin, 1 partner enforced [Logic])
- [ ] T023 [US3] Verify Invariants and Business Rules section documents 26 global rules (authentication, authorization, concurrency, data protection, plan limits, GDPR, state transition immutability)
- [ ] T024 [US3] Confirm password handling rules documented (bcrypt cost 12+, never return password_hash in API responses, never log passwords)
- [ ] T025 [US3] Confirm token handling rules documented (JWT RS256 with active keys, access token 24h expiry, refresh token 30d expiry, revocation on logout)
- [ ] T026 [US3] Validate spec.md acceptance scenarios 1-3 for US3 are met (authentication rules, trip editing rules, subscription limit rules findable)

**Checkpoint**: Business rules documented with enforcement layers; validation logic unambiguous

---

## Phase 5: User Story 4 - State Transition Clarity (Priority: P2)

**Goal**: Document forward-only state transitions for all stateful entities with immutable audit trail rationale

**Independent Test**: Developer can identify valid state transitions and understand that reversals require new records

### Validation for User Story 4

- [ ] T027 [P] [US4] Verify User entity State Transitions section documents Active → Deleted as forward-only with immutable audit trail rationale in specs/006-core-domain-model/data-model.md
- [ ] T028 [P] [US4] Verify Subscription entity State Transitions section documents stub_pending → active → cancelled as forward-only (no reactivation path)
- [ ] T029 [P] [US4] Verify Trip entity State Transitions section documents draft → published as forward-only (no unpublish operation)
- [ ] T030 [P] [US4] Verify Suggestion entity State Transitions section documents pending → approved/rejected as forward-only (resubmission creates new record)
- [ ] T031 [P] [US4] Verify ConversationSession entity State Transitions section documents in_progress → completed/abandoned as forward-only
- [ ] T032 [US4] Confirm all stateful entities include rationale explaining immutable audit trail requirement (prevents data tampering, preserves decision history)
- [ ] T033 [US4] Validate spec.md acceptance scenarios 1-3 for US4 are met (trip publishing, subscription management, suggestion workflow transitions documented as forward-only)

**Checkpoint**: State machines documented; immutable audit trail pattern clear

---

## Phase 6: User Story 5 - Concurrency and Versioning Strategy (Priority: P2)

**Goal**: Document optimistic locking strategy for versioned entities (Trip, Activity) with conflict detection and resolution

**Independent Test**: Developer can implement concurrent modification handling with version-based locking

### Validation for User Story 5

- [ ] T034 [P] [US5] Verify Trip entity has Concurrency Control section documenting version field, If-Match header, WHERE clause check, 409 Conflict response format in specs/006-core-domain-model/data-model.md
- [ ] T035 [P] [US5] Verify Activity entity has Concurrency Control section documenting same optimistic locking mechanism as Trip
- [ ] T036 [US5] Confirm Trip entity documents version increment strategy (SET version = version + 1 on successful update, atomic increment)
- [ ] T037 [US5] Confirm Concurrency invariants #6-8 documented (optimistic locking enforced, version numbers increment atomically, 409 Conflict with current version in response)
- [ ] T038 [US5] Validate spec.md acceptance scenarios 1-3 for US5 are met (trip updates with If-Match, activity updates with version increment, conflict handling with 409 response documented)

**Checkpoint**: Concurrency control strategy documented; safe concurrent access patterns clear

---

## Phase 7: Performance-Critical Indexes

**Goal**: Document indexes for entities with query-heavy access patterns (10 entities require index documentation per research.md)

**Independent Test**: Developer can identify which indexes to create for performant queries

- [ ] T039 [P] Verify User entity indexes documented (idx_users_email UNIQUE expression index for case-insensitive lookup, idx_users_subscription_id) in specs/006-core-domain-model/data-model.md
- [ ] T040 [P] Verify Trip entity indexes documented (idx_trips_creator_id for user's trip list, idx_trips_status for published/draft filtering)
- [ ] T041 [P] Verify Day entity indexes documented (idx_days_trip_id for trip detail queries, idx_days_destination_id for destination usage)
- [ ] T042 [P] Verify Activity entity indexes documented (idx_activities_day_id for day detail queries ordered by sequence)
- [ ] T043 [P] Verify Collaborator entity indexes documented (idx_collaborators_trip_id, idx_collaborators_user_id for collaboration queries)
- [ ] T044 [P] Verify Suggestion entity indexes documented (idx_suggestions_trip_id, idx_suggestions_author_id, idx_suggestions_status for suggestion management)
- [ ] T045 [P] Verify Subscription entity indexes documented (idx_subscriptions_user_id UNIQUE for one subscription per user, idx_subscriptions_status)
- [ ] T046 [P] Verify ConversationSession entity indexes documented (idx_conversation_sessions_trip_id, idx_conversation_sessions_status)
- [ ] T047 [P] Verify ConversationMessage entity indexes documented (idx_conversation_messages_session_id for message history chronological ordering)
- [ ] T048 [P] Verify Destination entity indexes documented (idx_destinations_country, optional spatial index for coordinates if PostGIS enabled)

**Checkpoint**: Performance-critical indexes documented; query optimization guidance clear

---

## Phase 8: Developer Reference Guide

**Goal**: Validate quickstart.md provides runnable scenarios that test domain model understanding

**Independent Test**: New developer can work through scenarios and verify understanding without asking questions

- [ ] T049 [P] Verify specs/006-core-domain-model/quickstart.md contains 12 validation scenarios covering all 5 user stories
- [ ] T050 [P] Verify Scenario 1 (three-layer validation) references User entity validation rules from data-model.md
- [ ] T051 [P] Verify Scenario 2 (optimistic locking) references Trip concurrency control from data-model.md
- [ ] T052 [P] Verify Scenario 3 (cascade deletes) references Trip cascade behavior from data-model.md
- [ ] T053 [P] Verify Scenario 4 (forward-only transitions) references Subscription state transitions from data-model.md
- [ ] T054 [P] Verify Scenario 5 (suggest-then-approve) references Collaborator and Suggestion entities from data-model.md
- [ ] T055 [P] Verify Scenario 6 (plan limits) references Plan and Collaborator business rules from data-model.md
- [ ] T056 [P] Verify Scenario 7 (GDPR deletion) references User cascade behavior and invariants #16-17 from data-model.md
- [ ] T057 [P] Verify Scenario 8 (geographic data) references Destination entity with coordinates from data-model.md
- [ ] T058 [P] Verify Scenario 9 (activity sequencing) references Activity sequence_order from data-model.md
- [ ] T059 [P] Verify Scenario 10 (token tracking) references ConversationSession and ConversationMessage from data-model.md
- [ ] T060 [P] Verify Scenario 11 (security audit trail) references SecurityEvent cascade behavior from data-model.md
- [ ] T061 [P] Verify Scenario 12 (JWT rotation) references JWTSigningKey rotation strategy from data-model.md

**Checkpoint**: Developer reference guide validated; self-service documentation complete

---

## Phase 9: Completeness Validation & Success Criteria

**Goal**: Validate all 8 success criteria from spec.md are met

**Independent Test**: Checklist review confirms zero gaps in documentation coverage

- [ ] T062 Validate SC-001: All 16 entities locatable without questions (entity catalog complete, organized alphabetically in data-model.md)
- [ ] T063 Validate SC-002: Zero ambiguity issues (all attributes have types, nullability, constraints; validation layers tagged)
- [ ] T064 Validate SC-003: 100% entity relationship coverage (ERD includes all entities, all FK relationships documented with cascade behavior)
- [ ] T065 Validate SC-004: 100% business rule coverage (all 16 entities have Business Rules section, 26 global invariants documented)
- [ ] T066 Validate SC-005: Migration sequence documented (16 migrations listed in order in data-model.md Database Migrations section)
- [ ] T067 Validate SC-006: Cascade behavior prevents FK violations (all foreign keys document CASCADE, SET NULL, or RESTRICT)
- [ ] T068 Validate SC-007: Concurrency-sensitive operations documented (Trip and Activity have Concurrency Control sections with optimistic locking)
- [ ] T069 Validate SC-008: Security-sensitive field handling documented (password_hash bcrypt rules, token storage rules, private key AWS Secrets Manager ARN, PII removal rules)
- [ ] T070 Update specs/006-core-domain-model/checklists/requirements.md with final validation status (all 16 checklist items passing)

**Checkpoint**: All success criteria validated; documentation complete and ready for promotion

---

## Phase 10: Documentation Promotion & Handoff

**Goal**: Promote completed documentation to docs/data-model.md and create PR for review

**Independent Test**: docs/data-model.md updated with gap-filled content; PR passes review

- [ ] T071 Review specs/006-core-domain-model/data-model.md for accuracy and completeness (all gaps from research.md filled)
- [ ] T072 Update docs/data-model.md with promoted content from specs/006-core-domain-model/data-model.md (preserve PROMOTED markers, update promotion date to 2026-07-06)
- [ ] T073 Add promotion metadata to docs/data-model.md header (<!-- Generated from specs/006-core-domain-model/data-model.md, Last promoted: 2026-07-06 -->)
- [ ] T074 [P] Update .github/memory/session-notes.md with session summary (document gaps filled, entities added, validation strategy)
- [ ] T075 [P] Update .github/memory/patterns-discovered.md if new reusable patterns identified (three-layer validation taxonomy, forward-only state transition pattern)
- [ ] T076 Create PR with title "feat(docs): complete core domain model with gap-filled entity catalog" targeting main branch
- [ ] T077 Add PR description summarizing 6 gaps filled, 16 entities documented, 8 success criteria met, and link to specs/006-core-domain-model/spec.md

**Checkpoint**: Documentation promoted; PR ready for team review

---

## Dependencies & Parallel Execution

**Dependency Graph** (User Story Completion Order):
```
Phase 1 (Setup) → Phase 2 (US1) → Phase 3 (US2) → Phase 4 (US3) → Phase 5 (US4) → Phase 6 (US5) → Phase 7 (Indexes) → Phase 8 (Quickstart) → Phase 9 (Validation) → Phase 10 (Promotion)
```

**Parallel Execution Examples**:

**Within Phase 2 (US1)**:
- T004, T005, T006, T007 can run in parallel (independent entity verification)
- T008, T009, T010 must run after T004-T007 (require complete entity list)

**Within Phase 3 (US2)**:
- T012, T013 can run in parallel (ERD verification, FK specification)
- T015, T016, T017 can run in parallel (independent cascade behavior verification)

**Phase 7 (Indexes)**:
- T039-T048 can all run in parallel (independent entity index verification)

**Phase 8 (Quickstart)**:
- T050-T061 can all run in parallel (independent scenario cross-reference verification)

**Phase 10 (Promotion)**:
- T074, T075 can run in parallel (independent memory file updates)

---

## Summary

- **Total Tasks**: 77
- **Parallelizable Tasks**: 48 (marked with [P])
- **Sequential Tasks**: 29
- **User Story Distribution**:
  - US1 (P1): 8 tasks — Core entity reference catalog
  - US2 (P1): 7 tasks — Relationship and cascade behavior
  - US3 (P1): 8 tasks — Business rules and validation layers
  - US4 (P2): 7 tasks — State transition immutability
  - US5 (P2): 5 tasks — Concurrency control strategy
  - Indexes: 10 tasks — Performance-critical index documentation
  - Quickstart: 13 tasks — Developer reference validation
  - Completeness: 9 tasks — Success criteria validation
  - Promotion: 7 tasks — Documentation handoff
  - Setup: 3 tasks — Infrastructure preparation

**MVP Scope**: Completing Phase 1-2 (US1) delivers the minimum viable entity catalog. Completing Phase 1-4 (US1-US3, all P1 stories) delivers a production-ready domain model reference.

**Estimated Effort**: ~2-3 hours for validation and promotion (documentation review, not implementation). Most artifacts already created in plan.md Phase 0-1.
