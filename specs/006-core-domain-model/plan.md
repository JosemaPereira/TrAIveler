# Implementation Plan: Core Domain and Data Model Foundations

**Branch**: `006-core-domain-model` | **Date**: 2026-07-06 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/006-core-domain-model/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Complete the canonical domain model documentation by filling specific gaps in `docs/data-model.md`. Ensure all 16 core entities have complete attribute specifications, three-layer validation rules (database constraints, business logic, API validation), performance-critical index documentation, geographic attributes for Destination entity, and forward-only state transition rules with immutable audit trail enforcement. This documentation forms the authoritative reference for all feature implementation work.

## Technical Context

**Language/Version**: N/A (documentation artifact, not code)

**Primary Dependencies**: N/A (documentation references PostgreSQL 15.4+, Goose for migrations, AWS Secrets Manager)

**Storage**: PostgreSQL 15.4+ on Amazon RDS (documented target; not implementing in this feature)

**Testing**: Documentation completeness validation via checklist (specs/006-core-domain-model/checklists/requirements.md)

**Target Platform**: Documentation consumed by backend developers implementing Go 1.24+ services

**Project Type**: Technical documentation (domain model reference, not application code)

**Performance Goals**: Zero developer ambiguity (SC-001: zero Slack questions about entity structure during first sprint)

**Constraints**: Must maintain consistency with existing docs/data-model.md structure; must not duplicate content; gap-filling only

**Scale/Scope**: 16 core entities documented with complete attribute specifications, relationships, business rules, state transitions, validation layers, and performance-critical indexes

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Rationale |
|-----------|--------|-----------|
| **I. Test-First Development** | N/A | Documentation artifact; no code implementation. Completeness validated via checklist. |
| **II. Simplicity — KISS & DRY** | ✅ PASS | Gap-filling approach avoids duplication; reuses existing docs/data-model.md structure. |
| **III. Code Quality & Consistency** | ✅ PASS | English-only requirement applies; documentation follows project standards. |
| **IV. Accessible & Token-Driven UI** | N/A | No user interface; backend domain model documentation only. |
| **V. Secure Configuration** | ✅ PASS | Documents security requirements (password hashing, token storage, key rotation) without exposing secrets. |
| **Mandated Technology Stack** | ✅ PASS | Documents PostgreSQL 15.4+, Goose, AWS Secrets Manager per existing constitution. |

**Overall**: ✅ **PASS** — No violations. Documentation feature aligns with simplicity principle (gap-filling only) and consistency requirements (English-only, structured format).

## Project Structure

### Documentation (this feature)

```text
specs/006-core-domain-model/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (in progress)
├── research.md          # Phase 0: Gap analysis and documentation strategy
├── data-model.md        # Phase 1: Complete entity catalog with gaps filled
├── quickstart.md        # Phase 1: Developer reference guide
└── checklists/
    └── requirements.md  # Specification quality checklist (completed)
```

### Source Code (repository root)

This feature produces documentation updates, not application code. The primary artifact is an enhanced version of the existing domain model documentation.

```text
docs/
└── data-model.md        # Target for promotion: will be updated from specs/006-core-domain-model/data-model.md
```

**Structure Decision**: Documentation-only feature. Outputs are created in `specs/006-core-domain-model/` and will be promoted to `docs/data-model.md` after review. No application code, migrations, or tests are created by this feature.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations to justify. Constitution Check passed without exceptions.

---

## Phase 0: Research (Complete)

**Deliverable**: [research.md](research.md)

**Key Findings**:
1. **Missing Entities**: ConversationSession and ConversationMessage lack complete definitions
2. **Incomplete Destination**: Missing geographic coordinates (latitude, longitude, country, region)
3. **Validation Layers**: No explicit [DB]/[Logic]/[API] tags on validation rules
4. **Index Documentation**: Only 2 of 16 entities document performance-critical indexes
5. **State Transition Immutability**: Not explicitly documented as forward-only
6. **Cascade Behavior**: Inconsistently documented across relationships

**Strategy**: Gap-filling approach preserves existing docs/data-model.md structure; adds missing content only

---

## Phase 1: Design & Contracts (Complete)

**Deliverables**:
- [data-model.md](data-model.md) — Complete entity catalog with all gaps filled (16 entities)
- [quickstart.md](quickstart.md) — Developer validation guide with 12 runnable scenarios
- Contracts skipped — Documentation feature has no external interfaces

**Artifacts Created**:

### data-model.md Enhancements

✅ **ConversationSession entity** — Full definition with session lifecycle, token tracking, state transitions  
✅ **ConversationMessage entity** — Message structure with role, content, token count, sanitization rules  
✅ **Destination entity** — Geographic attributes (latitude, longitude, country, region) with coordinate constraints  
✅ **Three-layer validation tags** — All 16 entities document [DB], [Logic], [API] validation enforcement  
✅ **Performance-critical indexes** — 10 entities now have index documentation (User, Trip, Day, Activity, Collaborator, Suggestion, ConversationSession, ConversationMessage, Destination, Subscription)  
✅ **Forward-only state transitions** — All stateful entities (User, Subscription, Trip, Suggestion, ConversationSession) explicitly document immutable audit trail requirement  
✅ **Cascade behavior** — All foreign keys specify CASCADE, SET NULL, or RESTRICT behavior  

**Completion Metrics**:
- 16 entities with complete attribute specifications ✅
- 100% entity relationship coverage in ERD ✅
- 18 Invariants and Business Rules documented ✅
- 16 Database migrations sequenced ✅
- 6 gaps from research.md filled ✅

### quickstart.md Scenarios

12 validation scenarios covering:
- Three-layer validation (Scenario 1)
- Optimistic locking (Scenario 2)
- Cascade deletes (Scenario 3)
- Forward-only state transitions (Scenario 4)
- Suggest-then-approve workflow (Scenario 5)
- Plan limits (Scenario 6)
- GDPR-compliant deletion (Scenario 7)
- Geographic data (Scenario 8)
- Activity sequencing (Scenario 9)
- Token tracking (Scenario 10)
- Security audit trail (Scenario 11)
- Zero-downtime key rotation (Scenario 12)

---

## Phase 1: Constitution Re-Check

| Principle | Status | Rationale |
|-----------|--------|-----------|
| **I. Test-First Development** | N/A | Documentation artifact; no code implementation. |
| **II. Simplicity — KISS & DRY** | ✅ PASS | Gap-filling approach avoids duplication; consistent format with existing docs. |
| **III. Code Quality & Consistency** | ✅ PASS | English-only; structured format; validation layer taxonomy consistent. |
| **IV. Accessible & Token-Driven UI** | N/A | No user interface; backend domain model documentation only. |
| **V. Secure Configuration** | ✅ PASS | Documents security requirements without exposing implementation secrets. |
| **Mandated Technology Stack** | ✅ PASS | Documents PostgreSQL, Goose, AWS Secrets Manager per constitution. |

**Post-Design Assessment**: ✅ **PASS** — No new violations introduced during design phase. Documentation maintains consistency with existing structure and constitution requirements.
