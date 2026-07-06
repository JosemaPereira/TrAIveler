# Tasks: API Design Standards

**Input**: Design documents from `/specs/007-api-design-standards/`

**Prerequisites**: plan.md (complete), spec.md (complete), research.md (complete), data-model.md (complete), contracts/ (complete)

**Tests**: This feature includes validation tests to verify endpoint compliance with standards, not traditional TDD tests of endpoints themselves.

**Organization**: Tasks are grouped by user story to enable independent implementation and delivery of standards documentation, tooling, and enforcement.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Documentation**: `docs/`, `specs/007-api-design-standards/`
- **Backend tests**: `backend/tests/integration/`
- **PR templates**: `.github/`
- **Constitution**: `.specify/memory/`

---

## Phase 1: Setup (Documentation Foundation)

**Purpose**: Establish directory structure and baseline documentation

- [ ] T001 Verify specs/007-api-design-standards/ structure is complete (plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md)
- [ ] T002 [P] Create backend/tests/integration/ directory if not exists
- [ ] T003 [P] Read existing .github/pull_request_template.md to understand current structure

---

## Phase 2: Foundational (Standards Document Promotion)

**Purpose**: Make API standards the authoritative project-wide reference. MUST be complete before user story work.

**⚠️ CRITICAL**: This phase creates the reference document that all subsequent tasks depend on.

- [ ] T004 Promote specs/007-api-design-standards/contracts/api-design-standards.md to docs/api-design-standards.md (authoritative location)
- [ ] T005 Verify docs/api-design-standards.md includes all 15 sections with examples (resource naming, URL structure, versioning, request/response format, error format, status codes, pagination, filtering, sorting, rate limiting, auth headers, timestamps, endpoint patterns, compliance, references)
- [ ] T006 [P] Update .github/copilot-instructions.md Documentation References section to include docs/api-design-standards.md with description
- [ ] T007 Audit existing endpoints in specs/001-product-vision-scope/contracts/api.md against docs/api-design-standards.md
- [ ] T008 Audit existing endpoints in specs/004-security-auth-model/contracts/api.md against docs/api-design-standards.md
- [ ] T009 Document any deviations found in audits (create specs/007-api-design-standards/audit-report.md with list of compliant vs. non-compliant patterns)

**Checkpoint**: Standards document is now the project-wide reference; existing endpoints audited for compliance

---

## Phase 3: User Story 1 - Backend Developer Creates Consistent Endpoints (Priority: P1) 🎯 MVP

**Goal**: Backend developers can design compliant endpoints within 15 minutes by consulting the standards document, without asking teammates or studying existing code.

**Independent Test**: Give a backend developer the standards document and ask them to design a new endpoint. Success = endpoint follows all conventions without additional guidance.

### US1: Standards Accessibility & Examples

- [ ] T010 [P] [US1] Verify docs/api-design-standards.md Table of Contents has anchor links to all 15 sections
- [ ] T011 [P] [US1] Verify each standard section includes ✅ DO and ❌ DON'T examples with code snippets
- [ ] T012 [P] [US1] Verify error codes catalog in docs/api-design-standards.md lists all 11 machine-readable codes (invalid_request, validation_failed, authentication_required, forbidden, not_found, conflict, rate_limit_exceeded, internal_error, etc.)
- [ ] T013 [US1] Verify endpoint patterns section includes 7 reusable templates (List Resources, Get Single, Create, Update Full, Update Partial, Delete, Action on Resource)
- [ ] T014 [US1] Create docs/api-design-standards.md quick reference card section at top (1-page summary of all conventions for printing/bookmarking)

### US1: Resource Naming & URL Structure Clarity

- [ ] T015 [P] [US1] Verify resource naming section documents plural nouns, lowercase, hyphenated compounds with 5+ examples
- [ ] T016 [P] [US1] Verify URL nesting section documents max 2 levels with nested vs. top-level endpoint guidance
- [ ] T017 [US1] Add decision tree diagram to docs/api-design-standards.md: "Should this resource be nested or top-level?"

### US1: Request/Response Format Clarity

- [ ] T018 [P] [US1] Verify request format section documents snake_case fields, ISO 8601 timestamps, required vs. optional fields
- [ ] T019 [P] [US1] Verify response format section documents flat JSON for singles, envelope for lists with pagination metadata
- [ ] T020 [US1] Add side-by-side comparison in docs/api-design-standards.md: correct vs. incorrect field naming examples

### US1: Validation Scenarios for Self-Testing

- [ ] T021 [US1] Verify specs/007-api-design-standards/quickstart.md includes 10 validation scenarios covering all standards
- [ ] T022 [US1] Add "How to validate your endpoint" checklist to docs/api-design-standards.md referencing quickstart.md scenarios

**Checkpoint**: Backend developers can now design compliant endpoints using docs/api-design-standards.md without external help

---

## Phase 4: User Story 2 - Frontend Developer Integrates Predictably (Priority: P2)

**Goal**: Frontend developers can write generic API client code (error handling, pagination, filtering) that works consistently across all endpoints.

**Independent Test**: Provide frontend developer with standards document and sample endpoint signatures. Success = they predict request/response formats without testing each endpoint.

### US2: Predictable Error Handling Documentation

- [ ] T023 [P] [US2] Verify error format section in docs/api-design-standards.md includes full TypeScript interface for error response structure
- [ ] T024 [P] [US2] Add frontend integration example to docs/api-design-standards.md showing generic error handler using error codes
- [ ] T025 [US2] Document in docs/api-design-standards.md which error codes map to which user-facing messages (UX guidance)

### US2: Predictable Pagination Documentation

- [ ] T026 [P] [US2] Verify pagination section includes complete response envelope structure with all metadata fields
- [ ] T027 [P] [US2] Add frontend integration example to docs/api-design-standards.md showing generic pagination component using response metadata
- [ ] T028 [US2] Document edge cases in docs/api-design-standards.md: empty lists, page beyond total_pages, per_page validation

### US2: Predictable Filtering & Sorting Documentation

- [ ] T029 [P] [US2] Verify filtering section documents all operators ([eq], [ne], [gt], [gte], [lt], [lte], [in], [like]) with URLSearchParams examples
- [ ] T030 [P] [US2] Verify sorting section documents minus prefix for descending, comma-separated multi-field with priority order
- [ ] T031 [US2] Add frontend integration example to docs/api-design-standards.md showing generic query builder constructing filter/sort params

### US2: Rate Limiting Transparency Documentation

- [ ] T032 [P] [US2] Verify rate limiting section documents all response headers (X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset)
- [ ] T033 [US2] Add frontend integration example to docs/api-design-standards.md showing how to read rate limit headers and implement proactive throttling

**Checkpoint**: Frontend developers can now write generic API client code using docs/api-design-standards.md patterns

---

## Phase 5: User Story 3 - Technical Lead Reviews API Changes (Priority: P2)

**Goal**: Code reviewers can identify convention violations in under 5 minutes by comparing PR against standards document.

**Independent Test**: Ask technical lead to review PR with new endpoint code against standards. Success = identify all violations quickly and objectively.

### US3: PR Template Enhancement

- [ ] T034 [US3] Read current .github/pull_request_template.md structure
- [ ] T035 [US3] Add "API Standards Compliance" section to .github/pull_request_template.md with checklist (if PR touches backend/ or adds/modifies API endpoints)
- [ ] T036 [US3] Create API standards compliance checklist in .github/pull_request_template.md covering 10 categories (resource naming, URL structure, versioning, request format, response format, error format, status codes, pagination, filtering, rate limiting)

### US3: Code Review Guidance

- [ ] T037 [P] [US3] Create docs/api-review-checklist.md with detailed verification steps for each standard (what to look for, common violations, how to verify)
- [ ] T038 [P] [US3] Add "For Code Reviewers" section to docs/api-design-standards.md with quick verification tips
- [ ] T039 [US3] Update .github/pull_request_template.md to link to docs/api-review-checklist.md in API Standards Compliance section

### US3: Exception Process Documentation

- [ ] T040 [US3] Verify docs/api-design-standards.md Exceptions section documents process: document why, what alternative, get technical lead approval
- [ ] T041 [US3] Add exception template to docs/api-design-standards.md (required fields: endpoint, standard violated, reason, alternative approach, approver)

**Checkpoint**: Code reviewers can now efficiently verify API standards compliance using PR template checklist and review guidance

---

## Phase 6: User Story 4 - API Consumer Learns the System (Priority: P3)

**Goal**: External API consumers can predict endpoint behavior after learning 2-3 endpoints, reducing learning curve and integration time.

**Independent Test**: Show external developer 2-3 endpoint docs. Success = they predict how to use other endpoints based on learned patterns.

### US4: External-Facing Documentation

- [ ] T042 [P] [US4] Create docs/api-getting-started.md for external developers (assumes no internal context) with introduction to API philosophy
- [ ] T043 [P] [US4] Add "Key Conventions at a Glance" section to docs/api-getting-started.md (10-item list of most important patterns)
- [ ] T044 [US4] Add 3 complete endpoint examples to docs/api-getting-started.md demonstrating all major conventions (List with pagination, Create with validation error, Get single with success)

### US4: Pattern Recognition Documentation

- [ ] T045 [P] [US4] Add "If you know this... then you know that" section to docs/api-getting-started.md (pattern transfer examples)
- [ ] T046 [P] [US4] Document consistency guarantees in docs/api-getting-started.md: all lists paginate identically, all errors structured identically, all timestamps formatted identically
- [ ] T047 [US4] Add FAQ section to docs/api-getting-started.md addressing common API consumer questions (How do I handle errors? How do I paginate? How do I filter?)

### US4: API Reference Structure

- [ ] T048 [US4] Update docs/api-design-standards.md to include "For API Consumers" callouts highlighting patterns that benefit external developers
- [ ] T049 [US4] Link docs/api-getting-started.md from docs/api-design-standards.md and README.md (make discoverable)

**Checkpoint**: External API consumers can now learn the system efficiently using docs/api-getting-started.md

---

## Phase 7: Polish & Enforcement (Automated Validation)

**Purpose**: Automated validation of standards compliance to reduce manual review burden

### Integration Tests for Standards Validation

- [ ] T050 [P] Create backend/tests/integration/api_standards_test.go file structure
- [ ] T051 [P] Implement test helper functions in backend/tests/integration/api_standards_test.go: assertErrorFormat(), assertPaginationFormat(), assertStatusCode(), assertRateLimitHeaders()
- [ ] T052 [P] Write integration test in backend/tests/integration/api_standards_test.go: TestErrorResponseFormat validates error structure with all required fields (error, message, request_id)
- [ ] T053 [P] Write integration test in backend/tests/integration/api_standards_test.go: TestPaginationFormat validates list responses have data envelope and pagination metadata
- [ ] T054 [P] Write integration test in backend/tests/integration/api_standards_test.go: TestStatusCodeSemantics validates GET returns 200, POST returns 201, DELETE returns 204
- [ ] T055 [P] Write integration test in backend/tests/integration/api_standards_test.go: TestValidationErrorIncludesFields validates 422 responses include fields array with field-level errors
- [ ] T056 [P] Write integration test in backend/tests/integration/api_standards_test.go: TestRateLimitHeaders validates all responses include X-RateLimit-* headers
- [ ] T057 [P] Write integration test in backend/tests/integration/api_standards_test.go: TestTimestampFormat validates timestamps use ISO 8601 with UTC (Z suffix)
- [ ] T058 [P] Write integration test in backend/tests/integration/api_standards_test.go: TestFieldNaming validates response fields use snake_case (not camelCase or PascalCase)
- [ ] T059 [P] Write integration test in backend/tests/integration/api_standards_test.go: TestRequestIDCorrelation validates error response request_id matches X-Request-ID header

### Linter Rules (Automated Checks)

- [ ] T060 [P] Research golangci-lint custom linter options for URL structure validation
- [ ] T061 [P] Document linter integration plan in specs/007-api-design-standards/audit-report.md (which checks can be automated, which remain manual)
- [ ] T062 [P] Add comment to backend/.golangci.yml noting future custom linter rules for API standards (placeholder for future automation)

### Documentation & Training

- [ ] T063 [P] Update README.md to link to docs/api-design-standards.md in Documentation section
- [ ] T064 [P] Update docs/coding-guidelines.md to reference docs/api-design-standards.md for API-specific conventions
- [ ] T065 [P] Update specs/007-api-design-standards/audit-report.md with final compliance status of existing endpoints (list compliant, non-compliant, grandfathered exceptions)

### Validation Execution

- [ ] T066 Run all 10 validation scenarios from specs/007-api-design-standards/quickstart.md against staging environment
- [ ] T067 Document validation results in specs/007-api-design-standards/audit-report.md (which scenarios pass, which need fixes)
- [ ] T068 Run backend/tests/integration/api_standards_test.go test suite and verify all tests pass (or document expected failures for non-compliant endpoints)

### Constitution Update (If Needed)

- [ ] T069 Review .specify/memory/constitution.md to determine if API standards should be elevated to non-negotiable principles
- [ ] T070 If constitution update warranted, document proposal in specs/007-api-design-standards/constitution-amendment-proposal.md (rationale, proposed changes, impact)

**Checkpoint**: Standards compliance is now enforced through automated tests, PR templates, and review checklists

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories (standards doc must be promoted before US work)
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - US1 (P1), US2 (P2), US3 (P2) can proceed in parallel after Phase 2
  - US4 (P3) can start after Phase 2 but benefits from US1-US3 completion (references their work)
- **Polish (Phase 7)**: Can start after US1 (needs standards doc), runs parallel to US2-US4

### User Story Dependencies

- **User Story 1 (P1)**: Standards document promoted and accessible - No dependencies on other stories
- **User Story 2 (P2)**: Standards document complete with examples - May reference US1 examples but independently testable
- **User Story 3 (P2)**: Standards document and PR template - No dependencies on other stories
- **User Story 4 (P3)**: Complete standards documentation - Benefits from US1-US3 being done (references their patterns)

### Within Each User Story

- **US1**: Documentation tasks can run in parallel [P] (different sections), then integration tasks
- **US2**: Documentation examples build on US1 standards, all marked [P] can run in parallel
- **US3**: PR template before review checklist, review checklist before exception process
- **US4**: Getting started guide references standards doc, FAQ references patterns

### Parallel Opportunities

- **Setup (Phase 1)**: T002, T003 can run in parallel
- **Foundational (Phase 2)**: T006, T007, T008 can run in parallel after T004-T005 complete
- **User Story 1**: T010-T012, T015-T016, T018-T019 can run in parallel (different doc sections)
- **User Story 2**: T023-T024, T026-T027, T029-T030, T032 can run in parallel (independent examples)
- **User Story 3**: T037-T038 can run in parallel
- **User Story 4**: T042-T043, T045-T046 can run in parallel
- **Polish (Phase 7)**: T050-T059 (all integration tests) can run in parallel, T060-T062 (linter research) can run in parallel, T063-T065 (documentation updates) can run in parallel

---

## Parallel Example: User Story 1 (Backend Developer Standards)

```bash
# Launch all documentation verification tasks in parallel:
Task T010: "Verify Table of Contents anchor links"
Task T011: "Verify DO/DON'T examples in each section"
Task T012: "Verify error codes catalog completeness"

# Then launch resource naming tasks in parallel:
Task T015: "Verify resource naming conventions with examples"
Task T016: "Verify URL nesting max 2 levels guidance"

# Then launch request/response format tasks in parallel:
Task T018: "Verify request format snake_case, ISO 8601"
Task T019: "Verify response format flat vs. envelope"
```

---

## Parallel Example: Polish Phase (Integration Tests)

```bash
# Launch all integration test implementation tasks in parallel:
Task T052: "Test error response format"
Task T053: "Test pagination format"
Task T054: "Test status code semantics"
Task T055: "Test validation error fields array"
Task T056: "Test rate limit headers"
Task T057: "Test timestamp format"
Task T058: "Test field naming snake_case"
Task T059: "Test request ID correlation"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only) 🎯

1. Complete Phase 1: Setup (verify structure)
2. Complete Phase 2: Foundational (promote standards doc to docs/, audit existing endpoints)
3. Complete Phase 3: User Story 1 (backend developer standards access)
4. **STOP and VALIDATE**: Give backend developer the standards doc, ask them to design an endpoint
5. If they succeed in < 15 minutes without help, MVP is complete!

### Incremental Delivery

1. **Foundation**: Setup + Foundational → Standards doc promoted and existing endpoints audited
2. **+US1**: Backend developers have clear standards → Can design compliant endpoints independently
3. **+US2**: Frontend developers have predictable patterns → Can write generic API client code
4. **+US3**: Code reviewers have checklists → Can efficiently verify compliance
5. **+US4**: External consumers have learning guide → Can adopt API quickly
6. **+Polish**: Automated tests enforce standards → Reduce manual review burden

Each phase delivers measurable value independently.

### Parallel Team Strategy

With multiple team members:

1. **Everyone**: Complete Setup + Foundational together (standards doc promotion is blocking)
2. **Once Foundational done**:
   - **PM**: User Story 1 (standards accessibility for backend devs)
   - **Backend Engineer**: User Story 2 (frontend integration examples) + Phase 7 (integration tests)
   - **Technical Lead**: User Story 3 (code review tooling)
   - **Technical Writer**: User Story 4 (external consumer documentation)
3. Stories complete and integrate independently

---

## Success Metrics (from spec.md)

- **SC-001**: 100% of new endpoints follow standards (verified via PR checklist) → Enabled by Phase 5 (US3)
- **SC-002**: Backend devs design compliant endpoint in < 15 min → Enabled by Phase 3 (US1)
- **SC-003**: Reviewers identify violations in < 5 min → Enabled by Phase 5 (US3)
- **SC-004**: Frontend devs write generic client code → Enabled by Phase 4 (US2)
- **SC-005**: Standards doc has 2+ complete examples → Verified by Phase 3 (US1) tasks
- **SC-006**: Zero ambiguity (all questions answered) → Verified by Phase 6 (US4) FAQ
- **SC-007**: New team members apply conventions in 1 day → Enabled by Phase 6 (US4)

---

## Notes

- This is a **documentation feature**, not a code implementation feature
- Tests in Phase 7 validate that **endpoints comply with standards**, not traditional TDD tests
- [P] tasks can run in parallel (different files, independent work)
- [Story] label maps task to user story for traceability
- Each user story delivers independent value (can stop after any story)
- Foundational phase (standards doc promotion) is CRITICAL and blocks all user story work
- Integration tests in Phase 7 can run parallel to US2-US4 (only needs US1 standards doc)
- Constitution amendment (T069-T070) is optional and may be deferred post-MVP

---

## Total Task Count

- **Phase 1 (Setup)**: 3 tasks
- **Phase 2 (Foundational)**: 6 tasks
- **Phase 3 (US1 - P1)**: 13 tasks
- **Phase 4 (US2 - P2)**: 11 tasks
- **Phase 5 (US3 - P2)**: 8 tasks
- **Phase 6 (US4 - P3)**: 8 tasks
- **Phase 7 (Polish)**: 21 tasks

**Total**: 70 tasks (38 parallelizable [P] = 54%)

**MVP Scope** (Phases 1-3 only): 22 tasks
**Full Feature** (All phases): 70 tasks
