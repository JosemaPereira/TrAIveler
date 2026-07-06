# Research: Core Domain Model Gap Analysis

**Feature**: Core Domain and Data Model Foundations  
**Date**: 2026-07-06  
**Purpose**: Identify specific gaps in `docs/data-model.md` and define strategy for filling them

## Gap Analysis

### 1. Missing Entity Definitions

**Gap**: ConversationSession and ConversationMessage entities appear in ERD but lack complete attribute specifications

**Current State**: Entities referenced in mermaid diagram with basic relationship mapping  
**Required State**: Full entity definitions with attributes, data types, nullability, constraints, business rules, indexes  
**Impact**: Backend developers cannot implement conversation tracking without complete specifications

**Decision**: Add complete entity definitions following the same format as existing entities (User, Trip, etc.)

**Attributes to Document**:
- **ConversationSession**: id, trip_id (FK), started_at, completed_at, status, AI provider metadata
- **ConversationMessage**: id, session_id (FK), role (system/user/assistant), content, timestamp, token_count

---

### 2. Incomplete Destination Entity

**Gap**: Destination entity referenced in ERD and Day entity but lacks complete attribute specification

**Current State**: Mentioned as foreign key in Day entity; no standalone definition  
**Required State**: Full entity with geographic attributes per clarification (latitude, longitude, country, region)  
**Impact**: Trip planning features requiring location data (maps, distance calculations) cannot be implemented

**Decision**: Add Destination entity with geographic coordinates and regional metadata

**Attributes to Document**:
- `id` (UUID, PK)
- `name` (VARCHAR) — city or location name
- `country` (VARCHAR) — ISO 3166-1 alpha-2 country code
- `region` (VARCHAR, NULLABLE) — state/province/region
- `latitude` (DECIMAL(9,6)) — geographic coordinate
- `longitude` (DECIMAL(9,6)) — geographic coordinate

**Rationale**: Enables future map visualization and distance-based features without schema migration

---

### 3. Missing Three-Layer Validation Documentation

**Gap**: Current docs document business rules but do not explicitly specify validation layers (database constraints vs. business logic vs. API validation)

**Current State**: Business rules section documents what must be validated but not where  
**Required State**: Each validation rule tagged with enforcement layer: [DB], [Logic], [API]  
**Impact**: Developers implement redundant validations or miss critical constraints; no clear separation of concerns

**Decision**: Add explicit validation layer tags to all validation rules

**Validation Layers**:
- **[DB]**: Database constraints (CHECK, NOT NULL, UNIQUE, FK) — prevent data corruption at storage layer
- **[Logic]**: Service/domain layer business rules — enforce complex domain invariants
- **[API]**: Input validation at HTTP boundary — provide user-friendly error messages

**Example Format**:
```
**Validation Rules**:
- [DB] Email must be unique (UNIQUE constraint on users.email)
- [Logic] Password must contain uppercase, lowercase, and digit (validated before hash)
- [API] Email format validated against RFC 5322 (returns 400 with user-friendly message)
```

---

### 4. Incomplete Index Documentation

**Gap**: Only RefreshToken and SecurityEvent entities document performance-critical indexes

**Current State**: 2 of 16 entities have index specifications  
**Required State**: All entities with performance-critical indexes documented (foreign keys used in queries, frequently filtered columns)  
**Impact**: Queries on large tables (Trips, Days, Activities, Collaborators, Suggestions) may perform poorly

**Decision**: Document performance-critical indexes for entities with query-heavy access patterns

**Entities Requiring Index Documentation**:
- **User**: `idx_users_email` (UNIQUE) — login queries
- **Trip**: `idx_trips_creator_id` — user's trip list queries
- **Day**: `idx_days_trip_id`, `idx_days_destination_id` — trip detail queries
- **Activity**: `idx_activities_day_id` — day detail queries
- **Collaborator**: `idx_collaborators_trip_id`, `idx_collaborators_user_id` — collaboration queries
- **Suggestion**: `idx_suggestions_trip_id`, `idx_suggestions_author_id`, `idx_suggestions_status` — suggestion management
- **ConversationSession**: `idx_conversation_sessions_trip_id` — session lookup
- **ConversationMessage**: `idx_conversation_messages_session_id` — message history queries

**Rationale**: Focus on indexes that support actual query patterns; avoid over-indexing (write performance cost)

---

### 5. Unclear State Transition Immutability

**Gap**: State transitions documented but not explicitly marked as forward-only with immutable audit trail requirement

**Current State**: State transitions mentioned (e.g., draft → published, pending → approved) but reversibility unclear  
**Required State**: Each stateful entity's transitions explicitly documented as forward-only; reversals require new records  
**Impact**: Developers may implement state reset functionality that violates audit trail requirements

**Decision**: Add explicit forward-only transition documentation with immutable audit trail justification

**Stateful Entities to Update**:
- **User**: Active → Deleted (forward-only; no reactivation)
- **Subscription**: stub_pending → active → cancelled (forward-only; new subscription for reactivation)
- **Trip**: draft → published (forward-only; unpublishing requires new draft)
- **Suggestion**: pending → approved/rejected (forward-only; resubmission creates new record)

**Format**:
```
**State Transitions** (Forward-Only):
- draft → published: Admin publishes trip; no reverse transition (create new draft for modifications)
- Rationale: Immutable audit trail prevents data tampering and preserves decision history
```

---

### 6. Missing Cascade Delete Documentation

**Gap**: Foreign key CASCADE behavior mentioned inconsistently; not all relationships specify cascade rules

**Current State**: Some entities mention CASCADE (e.g., RefreshToken → User), others don't specify  
**Required State**: All foreign keys document cascade behavior (CASCADE, SET NULL, RESTRICT)  
**Impact**: Orphaned records or unexpected deletions during entity removal operations

**Decision**: Document cascade behavior for all foreign key relationships

**Cascade Strategy**:
- **CASCADE**: Child records deleted with parent (e.g., RefreshToken when User deleted, Day when Trip deleted)
- **SET NULL**: Reference nullified (e.g., SecurityEvent.user_id when User deleted for audit preservation)
- **RESTRICT**: Prevent parent deletion if children exist (not used in MVP; all deletions cascade or nullify)

---

## Documentation Strategy

### Approach: Gap-Filling, Not Duplication

**Principle**: Preserve existing docs/data-model.md structure; add missing content only

**Workflow**:
1. Create complete entity catalog in `specs/006-core-domain-model/data-model.md` with all gaps filled
2. Use consistent format from existing entities (User, Trip, RefreshToken) as template
3. Add validation layer tags [DB], [Logic], [API] to all entities
4. Document performance-critical indexes for high-query entities
5. Add forward-only state transition sections with audit trail rationale
6. Document cascade behavior for all foreign keys

**Review Checklist**:
- [ ] All 16 entities have complete attribute specifications
- [ ] ConversationSession and ConversationMessage fully defined
- [ ] Destination includes latitude, longitude, country, region
- [ ] All validation rules tagged with enforcement layer
- [ ] Performance-critical indexes documented (8 entities)
- [ ] State transitions marked as forward-only where applicable
- [ ] All foreign keys specify cascade behavior

---

## Decision Log

| Decision | Rationale | Alternative Rejected |
|----------|-----------|----------------------|
| Include geographic coordinates in Destination | Enables map features without schema migration | Defer to future: requires migration later, blocks features |
| Three-layer validation tags [DB]/[Logic]/[API] | Clear separation of concerns; prevents redundancy | Document rules only: ambiguous enforcement location |
| Performance-critical indexes only | Balances query speed with write cost | Index all FKs: excessive write overhead for low-query tables |
| Forward-only state transitions | Immutable audit trail prevents tampering | Allow reversals: loses decision history, enables data manipulation |
| Gap-filling approach | Preserves existing structure; avoids duplication | Full rewrite: duplicates effort, loses promoted markers |

---

## Validation Criteria

**Success Metrics** (from spec.md Success Criteria):
- SC-001: Zero Slack questions about entity structure during first sprint
- SC-002: Zero data model ambiguity issues in PR reviews
- SC-003: 100% entity relationship coverage (ERD completeness)
- SC-004: 100% business rule coverage for core entities
- SC-005: Clean migration execution on fresh database
- SC-006: Zero foreign key violations in integration tests
- SC-007: 100% concurrency-sensitive operations use documented locking
- SC-008: All security-sensitive fields have handling requirements

**Completion Criteria**:
- All gaps identified in this research are addressed in data-model.md
- Checklist items in specs/006-core-domain-model/checklists/requirements.md remain passing
- Documentation format consistent with existing docs/data-model.md structure
