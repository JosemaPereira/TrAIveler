# Feature Specification: Core Domain and Data Model Foundations

**Feature Branch**: `006-core-domain-model`

**Created**: 2026-07-06

**Status**: Draft

**Input**: User description: "Define the core domain and data model foundations. As an engineering team, we want the canonical set of core entities, their attributes, relationships, invariants, and the base business rules that govern them. Include the conceptual data model and the key domain constraints. This is the bridge to feature work, so keep it focused on the shared domain — do not define specific endpoints, screens, or feature behavior yet."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Core Entity Reference (Priority: P1)

As a backend developer implementing API endpoints, I need a complete reference of all core entities with their attributes, types, and constraints so I can create consistent database schemas and validate data correctly.

**Why this priority**: Without accurate entity definitions, every developer creates slightly different implementations, leading to data inconsistencies and runtime errors. This is foundational to all feature work.

**Independent Test**: Can be validated by reviewing the entity catalog and confirming each entity has defined attributes, data types, constraints, and nullability rules. Delivers immediate value by eliminating ambiguity in schema design.

**Acceptance Scenarios**:

1. **Given** a developer needs to implement user authentication, **When** they consult the domain model, **Then** they find complete User entity definition including password hashing requirements, role constraints, and email validation rules
2. **Given** a developer needs to create trip management features, **When** they consult the domain model, **Then** they find Trip entity with ownership model, status transitions, and version control requirements
3. **Given** a developer needs to implement collaboration features, **When** they consult the domain model, **Then** they find Collaborator entity with partner constraints and invitation workflow rules

---

### User Story 2 - Relationship and Constraint Understanding (Priority: P1)

As a developer implementing business logic, I need clear documentation of entity relationships, cardinality constraints, and cascade behaviors so I can enforce referential integrity and prevent orphaned data.

**Why this priority**: Incorrect relationship handling causes data corruption, orphaned records, and failed transactions. This prevents critical bugs and ensures data consistency.

**Independent Test**: Can be validated by reviewing relationship diagrams and confirming all foreign keys, cascade rules, and junction tables are documented. Delivers value by preventing relationship implementation errors.

**Acceptance Scenarios**:

1. **Given** a developer implements trip deletion, **When** they consult relationship rules, **Then** they understand Days and Activities cascade delete with the Trip
2. **Given** a developer implements user account deletion, **When** they consult relationship rules, **Then** they understand RefreshTokens cascade delete but Trip ownership requires special handling
3. **Given** a developer implements collaborator management, **When** they consult relationship rules, **Then** they understand the UNIQUE constraint on (trip_id, user_id) and cardinality limits per plan

---

### User Story 3 - Business Rule Enforcement (Priority: P1)

As a developer implementing service layer logic, I need explicit documentation of business rules and invariants so I can enforce them consistently across all operations and prevent invalid state.

**Why this priority**: Business rules encoded inconsistently across features lead to security vulnerabilities, data corruption, and violated assumptions. Canonical rules prevent these issues.

**Independent Test**: Can be validated by reviewing business rule catalog and confirming each rule has enforcement requirements and validation logic. Delivers value by ensuring consistent rule application.

**Acceptance Scenarios**:

1. **Given** a developer implements authentication, **When** they consult business rules, **Then** they find password complexity requirements, bcrypt cost factor, and token expiration policies
2. **Given** a developer implements trip editing, **When** they consult business rules, **Then** they find optimistic locking requirements, version increment rules, and admin-only modification constraints
3. **Given** a developer implements subscription limits, **When** they consult business rules, **Then** they find partner invite limits per plan and enforcement layer requirements

---

### User Story 4 - State Transition Clarity (Priority: P2)

As a developer implementing stateful entities, I need documented state transition rules so I can implement valid state machines and prevent illegal transitions.

**Why this priority**: Undocumented state transitions lead to entities in invalid states, broken workflows, and unhandled edge cases. This ensures state integrity. Forward-only transitions with immutable audit trails prevent data tampering.

**Independent Test**: Can be validated by reviewing state machine diagrams for each stateful entity and confirming all valid forward transitions are documented, and reversals are marked as requiring new record creation. Delivers value by preventing state corruption.

**Acceptance Scenarios**:

1. **Given** a developer implements trip publishing, **When** they consult state transition rules, **Then** they understand the draft → published transition is forward-only with any preconditions
2. **Given** a developer implements subscription management, **When** they consult state transition rules, **Then** they understand stub_pending → active → cancelled transitions are forward-only (no reactivation; new subscription required)
3. **Given** a developer implements suggestion workflow, **When** they consult state transition rules, **Then** they understand pending → approved/rejected transitions are forward-only; resubmission requires creating a new suggestion record

---

### User Story 5 - Concurrency and Versioning Strategy (Priority: P2)

As a developer implementing concurrent operations, I need clear documentation of optimistic locking strategy, version fields, and conflict resolution so I can handle concurrent modifications safely.

**Why this priority**: Concurrent modifications without proper locking lead to lost updates and data races. This prevents concurrency bugs in multi-user scenarios.

**Independent Test**: Can be validated by reviewing concurrency control documentation and confirming which entities use versioning, how conflicts are detected, and how they're resolved. Delivers value by ensuring safe concurrent access.

**Acceptance Scenarios**:

1. **Given** a developer implements trip updates, **When** they consult concurrency rules, **Then** they understand version-based optimistic locking with If-Match headers
2. **Given** a developer implements activity updates, **When** they consult concurrency rules, **Then** they understand version increment on successful updates and 409 Conflict on mismatch
3. **Given** a developer implements conflict handling, **When** they consult concurrency rules, **Then** they understand response format requirements and client retry expectations

---

### Edge Cases

- What happens when a user has existing trips and their account is deleted? (User deletion must handle trip ownership transfer or cascading deletion)
- How does the system handle concurrent modifications by admin and partner? (Optimistic locking on admin changes, suggestion workflow for partner changes)
- What happens when subscription plan limits change mid-flight? (Existing collaborators grandfathered, new invites blocked)
- How are orphaned records prevented when cascade deletes span multiple levels? (Database-level CASCADE constraints ensure automatic cleanup)
- What happens when JWT signing key rotation fails mid-process? (Multiple active keys support allows rollback; old key remains valid)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST define all core entities with complete attribute specifications including name, data type, nullability, uniqueness, and default values, with validation requirements documented at three layers: database constraints (CHECK, NOT NULL, UNIQUE), business logic rules, and API input validation
- **FR-002**: System MUST document all entity relationships with cardinality (1:1, 1:many, many:many), foreign key references, and cascade behavior (CASCADE, SET NULL, RESTRICT)
- **FR-003**: System MUST specify all business rules governing entity lifecycle including creation constraints, update validations, forward-only state transitions (reversals require new records for immutable audit trail), and deletion requirements
- **FR-004**: System MUST define all database constraints including primary keys, foreign keys, unique constraints, check constraints, and performance-critical indexes (foreign keys used in queries, frequently filtered columns)
- **FR-005**: System MUST document all invariants that must hold true across entity states including referential integrity rules, data consistency requirements, and concurrency control
- **FR-006**: System MUST specify optimistic locking strategy for entities supporting concurrent modifications including version field semantics and conflict resolution
- **FR-007**: System MUST define authentication and authorization entities (User, RefreshToken, JWTSigningKey) with security requirements including password hashing, token expiration, and key rotation
- **FR-008**: System MUST define subscription and plan entities with limit enforcement rules including collaborator quotas and feature gating
- **FR-009**: System MUST define trip domain entities (Trip, Day, Activity, Destination) with ownership model, content hierarchy, and sequence ordering; Destination MUST include geographic coordinates (latitude, longitude) and country/region for mapping features
- **FR-010**: System MUST define collaboration entities (Collaborator, Suggestion) with role-based permissions and approval workflow
- **FR-011**: System MUST define audit and security entities (SecurityEvent) with event types, severity levels, and retention policies
- **FR-012**: System MUST document migration strategy including migration tool, sequencing requirements, and rollback procedures
- **FR-013**: System MUST specify data protection rules including PII handling, credential storage, secret management, and GDPR compliance
- **FR-014**: System MUST define conversation domain entities (ConversationSession, ConversationMessage) for AI-assisted trip planning with session management and message history
- **FR-015**: System MUST specify travel style entities (TravelStyle, TripTravelStyle) with seed values and many-to-many relationship model

### Key Entities *(include if feature involves data)*

- **User**: Authenticated account with role-based permissions (admin/partner), email, password hash, subscription link
- **RefreshToken**: Long-lived token for session management with hash storage, expiration, and revocation
- **JWTSigningKey**: RSA key pair for token signing with multi-key rotation support and AWS Secrets Manager integration
- **SecurityEvent**: Structured audit log for authentication, authorization, validation failures, and concurrency conflicts
- **Plan**: Subscription tier defining account limits (basic plan in MVP)
- **Subscription**: Links user to plan with payment stub reference and status tracking
- **Trip**: Top-level journey container owned by admin with versioning for optimistic locking
- **Day**: Calendar day within trip associated with destination and sequence number
- **Activity**: Specific event within day with type classification (visit/food/logistics/transfer), AI-generation flag, and versioning
- **Collaborator**: Partner invite granting view/suggest permissions with acceptance tracking
- **Suggestion**: Partner-proposed modification requiring admin approval with payload and review status
- **Destination**: Geographic location with name, coordinates (latitude, longitude), and country/region for mapping and filtering
- **TravelStyle**: Lookup table for travel preferences (gastronomy, sports, technology, museums, film)
- **TripTravelStyle**: Many-to-many join linking trips to multiple travel styles with assignment tracking
- **ConversationSession**: AI planning session associated with trip for tracking generated content
- **ConversationMessage**: Individual message within conversation session with role and content

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of developers can locate entity definitions without asking questions (measured by zero Slack questions about entity structure during first sprint)
- **SC-002**: Zero data model ambiguity issues reported during implementation review (measured by PR comments requiring clarification)
- **SC-003**: 100% entity relationship coverage in documentation (measured by ERD completeness check: all FK references mapped)
- **SC-004**: 100% business rule coverage for core entities (measured by checklist: each entity has documented creation/update/delete rules)
- **SC-005**: All database migrations execute successfully in sequence without conflicts (measured by clean migration run on fresh database)
- **SC-006**: Zero foreign key violation errors in integration tests (measured by test suite passing all referential integrity scenarios)
- **SC-007**: 100% of concurrency-sensitive operations implement documented locking strategy (measured by code review checklist for Trip and Activity updates)
- **SC-008**: All security-sensitive fields (passwords, tokens, keys) have documented handling requirements (measured by security checklist completion)

## Clarifications

### Session 2026-07-06

- Q: What incremental value should this feature deliver beyond what already exists in docs/data-model.md? → A: Identify and fill specific gaps in docs/data-model.md (missing entities, incomplete business rules, undocumented indexes)
- Q: What level of index documentation completeness is required? → A: Document only performance-critical indexes (foreign keys used in queries, frequently filtered columns)
- Q: Should the Destination entity include geographic coordinates and additional metadata for mapping/display, or remain minimal with just name and ID? → A: Include coordinates (latitude, longitude) and country/region for mapping features
- Q: Should state transition documentation include reversibility rules (can states be undone) and re-transition policies (can rejected items be resubmitted)? → A: Document forward-only transitions; reversals require new records (immutable audit trail)
- Q: What validation layer should be documented in the domain model specification? → A: Document all three layers explicitly (database constraints, business logic rules, API validation requirements)

## Assumptions

- PostgreSQL 15.4+ is the target database platform with UUID, JSONB, and ENUM support
- Existing documentation in `docs/data-model.md` provides the foundation; this spec identifies and fills specific gaps in entity completeness, business rule documentation, and index specifications
- Goose is the established migration tool for sequential schema evolution
- AWS Secrets Manager is the established secret storage for private keys and credentials
- bcrypt with cost factor 12+ is the standard for password hashing
- RS256 with 2048-bit RSA keys is the standard for JWT signing
- Multi-key rotation strategy is required for zero-downtime key updates
- WCAG 2.1 AA accessibility and design token usage apply to frontend only; this spec focuses on domain model
- English is the language for all technical artifacts including entity names, attribute names, and documentation
