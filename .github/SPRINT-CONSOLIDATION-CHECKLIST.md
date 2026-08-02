# Sprint Task Consolidation Checklist

> **MANDATORY** — Apply this checklist BEFORE creating GitHub issues for any sprint.

## Purpose

Prevent issue fragmentation by consolidating related tasks into single work items BEFORE creating GitHub issues. This reduces:
- Issue proliferation (62% reduction in Sprint 2)
- PR/review overhead
- Context switching
- Backlog clutter

## When to Use

**Before creating issues for ANY sprint** — particularly at:
1. Sprint planning (before `/create-sprint-issues`)
2. Sprint closure (document lessons for next sprint)
3. Mid-sprint refinement (if adding new work)

## Consolidation Checklist

### Step 1: Load Sprint Tasks

```bash
# Read roadmap for target sprint
grep "| <sprint-number> | P1 | Backlog |" docs/roadmap.md
```

Count total tasks assigned to the sprint.

### Step 2: Apply Consolidation Rules

For each potential group, verify ALL criteria:

**✅ DO consolidate when ALL of**:
- [ ] Same spec (never cross spec boundaries)
- [ ] Same area/module (e.g., all in `internal/middleware/`)
- [ ] Same tech stack (all Go OR all React, never mixed)
- [ ] Shared context (config files, middleware package, primitives)
- [ ] Optimal size (2-4 tasks = 1 reviewable PR, ~150-200 LOC total)
- [ ] No blocking dependencies between grouped tasks
- [ ] Independent tracking not required (not critical blockers)

**❌ DON'T consolidate when**:
- [ ] Different tech stacks (Go + TypeScript = separate)
- [ ] Critical blocker task (needs individual visibility)
- [ ] Different dependency chains
- [ ] Already large task (>200 LOC estimate)
- [ ] Cross-spec boundary

### Step 3: Propose Consolidation Groups

Document proposed groups with reasoning:

```markdown
## Sprint X Consolidation Analysis

**Total tasks**: X
**Proposed work items**: Y (Z% reduction)

### Backend Groups
1. G-SPRINTX-BACKEND-<AREA>: TXX-TYY (N tasks)
   - Reasoning: Same module, shared context, ~150 LOC
   - Files: backend/internal/<area>/...

### Frontend Groups
2. G-SPRINTX-FRONTEND-<AREA>: TXX-TYY (N tasks)
   - Reasoning: Related primitives, same styling approach
   - Files: frontend/src/components/...

### Standalone Tasks
- TXX: <reason for not grouping>
```

### Step 4: Calculate Impact Metrics

```markdown
| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Tasks | X | X | - |
| Issues | X | Y | -Z% |
| PRs | ~X | ~Y | More coherent |
```

### Step 5: Update Roadmap Group Column

For each consolidated group:

```bash
# Update Group column in docs/roadmap.md
| ID | Task | Group | Sprint | ... |
|----|------|-------|--------|-----|
| TXX | ... | G-SPRINTX-<AREA> | X | ... |
| TYY | ... | G-SPRINTX-<AREA> | X | ... |
```

### Step 6: Validate Consolidation

Run validation checks:

```bash
# Verify no cross-spec groups
# Verify all group members share Sprint value
# Verify optimal group sizes (2-4 tasks preferred)
# Verify dependencies allow grouping
```

### Step 7: Create Issues

Only after consolidation is complete:

```bash
# Use PM agent or manual creation
# One issue per work item (group or standalone)
# Group issues include checklist of member tasks
```

## Sprint 2 Example (Approved Pattern)

### Analysis Results

**Total tasks**: 37
**Work items**: 14 (62% reduction)

### Groups Applied

**Backend (18 tasks → 6 work items)**:
1. **G-SPRINT2-BACKEND-MIDDLEWARE** (T024-T028) — 5 tasks
   - Same package: `internal/middleware/`
   - Files: request_id.go, logger.go, recovery.go, cors.go, body_size.go
2. **Standalone: T029** — Database client (foundation, deserves visibility)
3. **G-SPRINT2-BACKEND-AI** (T030-T032) — 3 tasks
   - Same package: `internal/ai/`
   - Files: client.go, validator.go, sanitizer.go
4. **G-SPRINT2-BACKEND-ERRORS** (T033-T034) — 2 tasks
   - Same package: `internal/errors/`
   - Files: handler.go, types.go
5. **G-SPRINT2-BACKEND-HTTP-SERVER** (T035-T037) — 3 tasks
   - Same file: `cmd/api/main.go`
   - Logical unit: HTTP server + DB integration + healthcheck
6. **G-SPRINT2-BACKEND-EXAMPLE** (T038-T041) — 4 tasks
   - Same package: `internal/example/`
   - Complete pattern: model → repository → service → handler

**Frontend (19 tasks → 8 work items)**:
1. **G-SPRINT2-FRONTEND-TOKENS** (T042-T043) — 2 tasks
   - Same directory: `src/styles/`
   - Related: tokens.css + global.css
2. **G-SPRINT2-FRONTEND-API-CONFIG** (T044-T045) — 2 tasks
   - Same directory: `src/lib/`
   - Related: api-client.ts + query-client.ts
3. **Standalone: T046** — Auth store (core feature, deserves visibility)
4. **G-SPRINT2-FRONTEND-APP-SHELL** (T047, T059-T060) — 3 tasks
   - Same file: `src/App.tsx`
   - Logical unit: App setup + router + error boundary
5. **G-SPRINT2-FRONTEND-PRIMITIVES-CORE** (T048-T051) — 4 tasks
   - Same directory: `src/components/primitives/`
   - Related: Button, Input, Card with CSS modules
6. **G-SPRINT2-FRONTEND-PRIMITIVES-STATE** (T052-T054) — 3 tasks
   - Same directory: `src/components/primitives/`
   - Related: Loading, Error, Empty state components
7. **Standalone: T055** — Form composite (bridges primitives, single file)
8. **G-SPRINT2-FRONTEND-INFRASTRUCTURE** (T056-T058) — 3 tasks
   - Setup tasks: .gitkeep, ErrorBoundary, routes/index.tsx
   - Small combined size (~50-80 LOC)

### Impact

- **Before**: 37 potential issues
- **After**: 14 work items
- **Reduction**: -62% fewer issues to manage
- **PR coherence**: Each group = 1 reviewable PR with shared context

## Sprint Closure: Prepare for Next Sprint

At the end of each sprint, add a section to this file:

```markdown
### Sprint X Lessons Learned

**What worked well**:
- List successful consolidation patterns
- Note optimal group sizes discovered

**What to improve**:
- List missed consolidation opportunities
- Note groups that were too large/small

**Patterns for Sprint X+1**:
- Carry forward proven consolidation patterns
```

## References

- [PM Workflow Consolidation](./.github/PM-WORKFLOW-CONSOLIDATION.md) — Detailed consolidation guide
- [CLAUDE.md](../CLAUDE.md) — Task consolidation policy
- [Issue Creation Guidelines](./.github/ISSUE-CREATION-GUIDELINES.md) — GitHub issue workflow

---

**Last Updated**: 2026-07-08 (Sprint 2 consolidation applied)
**Next Review**: Sprint 2 closure (add lessons learned)
