# Roadmap Reconciliation Report

**Date**: 2026-07-06  
**Reconciled with**: spec 008 — Authentication & Collaboration User Experience  
**Source**: `specs/008-auth-collaboration-ux/tasks.md`  
**Target**: `docs/roadmap.md`

---

## Summary

Successfully reconciled roadmap with spec 008. This spec is a **feature implementation** that delivers the complete authentication and collaboration user experience with 206 tasks across 10 phases, including paid/free user registration, login, password management, trip collaboration, suggestions workflow, and subscription lifecycle.

**Previous reconciliation**: spec 006 (Foundation Phase documentation)  
**This reconciliation**: spec 008 (Feature Phase implementation)

---

## Changes Made

### Operations Count

| Operation | Count | Details |
|-----------|-------|---------|
| **ADD** | 206 | All tasks from spec 008 added to Feature Phase |
| **UPDATE** | 0 | No existing tasks modified |
| **UNCHANGED** | 539 | All tasks from specs 001-007 preserved (Foundation Phase complete) |
| **REMOVE** | 0 | No tasks removed |
| **ARCHIVED** | 0 | No tasks archived |

**Total roadmap size**: 745 tasks (previously 539, +206 from spec 008)

---

## Spec 008 Details

**Title**: Authentication & Collaboration User Experience  
**Type**: Feature Implementation (Feature Phase)  
**User Stories**: 6 (US1-US3 Priority P1, US4-US6 Priority P2-P3)  
**Success Criteria**: 8  
**Phases**: 10

### Task Distribution by Phase

| Phase | Description | Task Count | Task IDs |
|-------|-------------|------------|----------|
| 1 | Setup (Shared Infrastructure) | 8 | 008-T001 to 008-T008 |
| 2 | Foundational (Blocking Prerequisites) | 28 | 008-T009 to 008-T036 |
| 3 | US1: Paid User Registration & First Trip Creation (P1) 🎯 MVP | 34 | 008-T037 to 008-T070 |
| 4 | US2: Free User Registration & Accepting Collaboration Invite (P1) | 26 | 008-T071 to 008-T096 |
| 5 | US3: Returning User Login & Trip Management (P1) | 26 | 008-T097 to 008-T122 |
| 6 | US4: Collaboration: Invite & Manage Suggestions (P2) | 17 | 008-T123 to 008-T139 |
| 7 | US5: Password Management & Account Security (P2) | 21 | 008-T140 to 008-T160 |
| 8 | US6: UI Component Library & Design System Basics (P3) | 14 | 008-T161 to 008-T174 |
| 9 | Subscription Lifecycle Management (P2) | 16 | 008-T175 to 008-T190 |
| 10 | Polish & Cross-Cutting Concerns | 16 | 008-T191 to 008-T206 |

### MVP Scope (Phases 1+2+3+5)

**MVP tasks**: 82 tasks  
**MVP User Stories**: US1 (Paid User), US3 (Login & Trip Management)  
**Deferred to post-MVP**: US2 (Free Users), US4 (Suggestions), US5 (Password Reset), US6 (Design System), Subscription Lifecycle, Polish

### Task Groups Created

| Group Name | Task Count | Phase(s) | Purpose |
|------------|------------|----------|---------|
| G-008-SETUP | 7 | 1 | Project structure initialization (backend, frontend, e2e, infra) |
| G-008-DATABASE | 2 | 2 | PostgreSQL connection and migrations setup |
| G-008-MIGRATIONS | 7 | 2 | Database migrations for all 7 tables |
| G-008-JWT | 3 | 2 | JWT generation, validation, refresh with RS256 |
| G-008-PASSWORD | 2 | 2 | Password hashing (bcrypt cost 12) and validation |
| G-008-RATELIMIT | 2 | 2 | Progressive delay rate limiter for auth endpoints |
| G-008-MIDDLEWARE | 3 | 2 | Auth, request ID, and rate limit middleware |
| G-008-FRONTEND-API | 2 | 2 | Axios instance with interceptors and error handler |
| G-008-US1-MODELS | 3 | 3 | User, Subscription, Trip models for MVP |
| G-008-US1-REPOS | 3 | 3 | User, Subscription, Stub Payment repositories |
| G-008-US1-TRIP | 2 | 3 | Trip repository and stubbed itinerary generation |
| G-008-US1-PRIMITIVES | 5 | 3 | Button, Input, Label, ErrorMessage primitives |
| G-008-US1-COMPOSITES | 1 | 3 | Card composite component |
| G-008-US1-COMPONENTS | 2 | 3 | LoadingSpinner, EmptyState feature components |
| G-008-US1-API | 2 | 3 | authApi and tripsApi service layers |
| G-008-US1-TRIP-HOOKS | 2 | 3 | useTrips and useCreateTrip React hooks |
| G-008-US2-MODELS | 2 | 4 | Collaborator, Suggestion models |
| G-008-US2-HANDLERS | 4 | 4 | Invite, accept, leave, list invitations handlers |
| G-008-US2-PRIMITIVES | 1 | 4 | Badge primitive for user status |
| G-008-US2-COMPOSITES | 2 | 4 | Modal, Banner composites |
| G-008-US2-HOOKS | 3 | 4 | Invitations, accept, leave trip hooks |
| G-008-US2-API | 1 | 4 | collaborationApi service layer |
| G-008-US3-MODELS | 1 | 5 | LoginRequest/LoginResponse models |
| G-008-US3-HANDLERS | 8 | 5 | Login, logout, trips list/detail/update/delete handlers |
| G-008-US3-HOOKS | 3 | 5 | useLogin, useLogout, useTripDetail hooks |
| G-008-US3-TRIP-HOOKS | 3 | 5 | Update, delete trip hooks |
| G-008-US3-COMPONENTS | 1 | 5 | DeleteTripModal component |
| G-008-US4-REPOS | 1 | 6 | Suggestion repository with stub apply logic |
| G-008-US4-HANDLERS | 4 | 6 | Create, list, approve, reject suggestion handlers |
| G-008-US4-HOOKS | 4 | 6 | Suggestions CRUD hooks |
| G-008-US5-REPOS | 2 | 7 | PasswordResetToken, RefreshToken repositories |
| G-008-US5-HANDLERS | 4 | 7 | Password reset, change, token refresh handlers |
| G-008-US5-HOOKS | 3 | 7 | Password reset/change, token refresh hooks |
| G-008-US5-COMPONENTS | 4 | 7 | Password reset/change forms and pages |
| G-008-US6-PRIMITIVES | 4 | 8 | Checkbox, Select, Textarea, Link primitives |
| G-008-US6-COMPOSITES | 2 | 8 | Toast, Tooltip composites |
| G-008-US6-AUDITS | 3 | 8 | Accessibility audits (forms, keyboard, contrast) |
| G-008-US6-DOCS | 1 | 8 | Design system documentation |
| G-008-SUBSCRIPTION | 3 | 9 | Cancel, renew subscription handlers |
| G-008-SUBSCRIPTION-HOOKS | 3 | 9 | Cancel, renew, upgrade subscription hooks |
| G-008-POLISH-LOGGING | 2 | 10 | Error and request logging |
| G-008-POLISH-SECURITY | 1 | 10 | Security headers middleware |
| G-008-POLISH-VALIDATION | 1 | 10 | Quickstart validation |
| G-008-POLISH-CLEANUP | 2 | 10 | Code cleanup (backend, frontend) |
| G-008-POLISH-DOCS | 2 | 10 | README updates |
| G-008-POLISH-E2E | 6 | 10 | End-to-end tests for all user stories |

**Standalone tasks**: 83 (40.3%)  
**Grouped tasks**: 123 (59.7%)

### Priority Distribution

| Priority | Count | Percentage |
|----------|-------|------------|
| P1 | 96 | 46.6% |
| P2 | 96 | 46.6% |
| P3 | 14 | 6.8% |

### Parallelization Opportunities

**Parallelizable tasks**: 101 (49.0%)  
**Sequential tasks**: 105 (51.0%)

Phases 1-2 (Setup & Foundational) have highest parallelization (75%), while phases 3-10 (User Stories) are more sequential due to dependencies on backend/frontend integration.

---

## Default Values Applied

All spec 008 tasks added to roadmap with following defaults:

| Field | Default Value | Rationale |
|-------|---------------|-----------|
| **Sprint** | _empty_ | Not yet assigned to sprint |
| **Status** | Backlog | Not yet started |
| **Issue** | _empty_ | GitHub issues not yet created |
| **Notes** | _empty_ | No additional context |
| **Group** | _as defined in spec_ | 47 groups created for logical task organization |
| **Priority** | _as defined in spec_ | P1 for MVP (phases 1-3, 5), P2-P3 for post-MVP |
| **Depends on** | _as defined in spec_ | Dependencies preserved from task file |
| **Parallel** | _as defined in spec_ | 101 tasks marked as parallelizable with `[P]` flag |

---

## Cross-Spec Dependencies

Spec 008 depends on foundational work from earlier specs:

- **Spec 004**: Security model (JWT RS256, bcrypt cost 12, rate limiting, security events)
- **Spec 005**: Architecture patterns (middleware chain, repository/service/handler layers, error handling, pgx database client)
- **Spec 006**: Data model (User, Subscription, Trip, Collaborator, Suggestion entities with optimistic locking)
- **Spec 007**: API standards (resource naming, error format, HTTP status codes, rate limit headers)

These dependencies are explicitly noted in the roadmap's cross-spec note for spec 008.

---

## Roadmap Structure

The roadmap now has two major phases:

### Foundation Phase (Specs 001-007)
- **539 tasks** across 7 foundational specifications
- Ordered by cross-spec dependency: vision → NFRs → cloud → security → architecture → domain → API standards
- All specs fully reconciled and up-to-date

### Feature Phase (Specs 008+)
- **206 tasks** from spec 008 (Authentication & Collaboration UX)
- First feature spec added to roadmap
- Implements security model, architecture patterns, and API standards from foundation

**Total roadmap**: 745 tasks (539 foundation + 206 feature)

---

## Cross-Spec Dependencies

Spec 006 has **NO code dependencies** on other specs. It is a documentation-only feature that completes the canonical domain model reference. However, all future code-generation specs (backend, frontend, IaC) will depend on this spec for entity definitions.

### Upstream Dependencies (specs 006 depends on)
- None (spec 006 uses already-promoted content from specs 001-005)

### Downstream Dependencies (specs that depend on 006)
- **All future code specs**: Backend (007+), Frontend (008+), IaC (009+), E2E (010+)
- Reason: Complete entity definitions with validation layers, indexes, and business rules required for code generation

---

## Critical Path Impact

### Current Critical Path
Foundation Phase minimum sequential chain (specs 001-005):
```
001-T001 → 002-T001 → 003-T001 → 004-T001 → 005-T001 → [MVP code features]
```

### Updated Critical Path
Spec 006 does **NOT modify** the critical path. It is a parallel documentation track that enhances the domain model reference but is not a blocking dependency for MVP code work.

However, spec 006 should be **completed before major code generation** to ensure all entities, validation rules, and business logic are documented. Recommended insertion point:

```
001-T001 → 002-T001 → 003-T001 → 004-T001 → 005-T001 → 006-T001 → [MVP code features]
```

Estimated impact: +0 days to critical path (documentation can happen in parallel with early setup tasks)

---

## Validation Results

✅ All 77 tasks from spec 006 added to roadmap  
✅ Task IDs formatted as 006-T001 through 006-T077  
✅ All groups preserved (8 groups created)  
✅ All dependencies preserved (48 dependency relationships maintained)  
✅ Parallel flags preserved (41 tasks marked as parallelizable)  
✅ Priority levels preserved (P1 for MVP-critical phases, P2 for enhancements)  
✅ Spec 006 inserted in Foundation Phase before "## Feature Phase" marker  
✅ Roadmap header updated to "Last reconciled: 2026-07-06 (updated with spec 006)"  
✅ Cross-spec note added explaining documentation-only nature

---

## Human Attention Required

### 1. New Spec Review
- **Action**: Review spec 006 section in roadmap for accuracy
- **Location**: `docs/roadmap.md` lines 712-1000+ (new section before Feature Phase)
- **Questions**: Are the 10 phases logically ordered? Are the 8 groups useful for sprint planning?

### 2. Priority Review
- **Action**: Confirm P1/P2 split aligns with MVP scope
- **Current**: Phases 1-4 (US1-US3) are P1, phases 5-6 (US4-US5) are P2, phases 7-8 (Indexes, Quickstart) are P2, phases 9-10 (Validation, Promotion) are P1
- **Consideration**: Should all documentation be P1 since it blocks code generation? Or is P2 acceptable for enhancement docs?

### 3. Sprint Assignment
- **Action**: Assign spec 006 tasks to sprints
- **Current**: All 77 tasks have empty Sprint field
- **Recommendation**: Documentation-only feature, could complete in 1-2 sprints (phases 1-6 in sprint 1, phases 7-10 in sprint 2)

### 4. Issue Creation
- **Action**: Create GitHub issues for spec 006 tasks
- **Current**: All 77 tasks have empty Issue field
- **Recommendation**: Use `speckit.taskstoissues` agent or create issues manually from task table

### 5. Dependency Validation
- **Action**: Verify spec 006 dependencies before starting work
- **Check**: Are docs from specs 001-005 already promoted? (Yes, per session context)
- **Check**: Is `specs/006-core-domain-model/data-model.md` gap-filled? (Yes, per session context)

### 6. Status Tracking
- **Action**: Update task status as work progresses
- **Current**: All 77 tasks are "Backlog"
- **Process**: Move to "In Progress" when started, "Done" when completed, update `specs/006-core-domain-model/tasks.md` accordingly

---

## Roadmap Statistics (After Reconciliation)

| Metric | Value | Change |
|--------|-------|--------|
| **Total Tasks** | 539 | +77 (+16.7%) |
| **Foundation Phase Tasks** | 539 | +77 (+16.7%) |
| **Feature Phase Tasks** | 0 | No change |
| **Total Specs** | 6 | +1 |
| **Grouped Tasks** | 373 | +49 |
| **Standalone Tasks** | 166 | +28 |
| **Parallelizable Tasks** | ~295 | +41 (estimate) |

---

## Next Steps

### Immediate (Today)
1. ✅ Delete temporary files from promotion workflow (already completed)
2. ✅ Review this report and validate spec 006 insertion (completed)
3. Commit roadmap changes: `git add docs/roadmap.md .github/*.md && git commit -m "feat(roadmap): add spec 006 tasks and update reports"`

### Short-term (This Week)
1. Review spec 006 priorities and adjust if needed
2. Assign spec 006 tasks to sprint(s)
3. Create GitHub issues for spec 006 (use `speckit.taskstoissues` or manual creation)
4. Update project board with new task count

### Medium-term (Next Sprint)
1. Execute spec 006 documentation tasks
2. Update task status in roadmap as work progresses
3. Track blockers and dependencies
4. Update critical path if spec 006 reveals new requirements

---

## Appendix: Spec 006 Task Summary Table

(For quick reference when reviewing roadmap)

| Phase | Tasks | Groups | Priority | Parallelizable | Key Deliverables |
|-------|-------|--------|----------|----------------|------------------|
| 1 | 3 | 0 | P1 | 2/3 | Setup validation infrastructure |
| 2 | 8 | 1 | P1 | 5/8 | Verify all 16 entities documented |
| 3 | 7 | 1 | P1 | 3/7 | Document cascade behavior |
| 4 | 8 | 1 | P1 | 5/8 | Document three-layer validation |
| 5 | 7 | 1 | P2 | 6/7 | Document state transitions |
| 6 | 5 | 1 | P2 | 2/5 | Document concurrency control |
| 7 | 10 | 1 | P2 | 10/10 | Document performance indexes |
| 8 | 13 | 1 | P2 | 13/13 | Create developer quickstart |
| 9 | 9 | 1 | P1 | 8/9 | Validate success criteria |
| 10 | 7 | 0 | P1 | 2/7 | Promote docs and create PR |

---

**Report generated**: 2026-07-06  
**Agent**: GitHub Copilot (Claude Sonnet 4.5)  
**Workflow**: `#prompt:build-roadmap.prompt.md`  

---

## Maintenance Notes

This report should be updated when:
- New specs are added to the roadmap
- Tasks from spec 006 change status or priority
- Critical path analysis changes based on new dependencies
- Cross-spec dependencies are discovered or modified

See [PROMOTION-REPORT.md](PROMOTION-REPORT.md) for details on promoted foundational decisions from specs 001-006.
