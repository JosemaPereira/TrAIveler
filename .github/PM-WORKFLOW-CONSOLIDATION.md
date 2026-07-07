# PM Workflow: Consolidation-First Sprint Planning

**Last updated**: 2026-07-07  
**Supersedes**: Previous sprint planning approach (post Sprint 1 consolidation lesson)

## Core Principle

**Consolidate BEFORE creating issues, not after.**

Small, atomic tasks that share context, technology, or purpose should be grouped into a single work item BEFORE issue creation. This reduces ticket waste, improves backlog health, and creates appropriately-sized PRs for review.

## Consolidation Rules

### When to Consolidate (Group)

Group tasks into ONE issue when they meet ALL of these criteria:

1. **Same spec** — Never cross spec boundaries
2. **Same area/module** — Share the same code area, directory, or logical module
3. **Same tech stack** — Same language, framework, or toolset (don't mix Go + React)
4. **Shared context** — Benefit from being worked on together (e.g., related config files, same API layer)
5. **Similar size** — Each task is small enough that 2-4 fit comfortably in one PR
6. **No blocking dependencies between them** — Must be executable in parallel or sequence without external blockers
7. **Independent tracking not required** — Don't group if separate issues need distinct status tracking for planning

**Optimal group size**: 2-4 tasks (1 PR, reviewable in one sitting)

### When NOT to Consolidate

Keep tasks as standalone issues when:

- **Critical blocker** — Task blocks many others and needs visibility (e.g., foundation setup)
- **Different tech stacks** — Mixing languages/frameworks (Go linting + ESLint = separate)
- **Different dependency chains** — One is blocked, the other isn't
- **Large task** — Already substantial enough for a full PR
- **Independent status tracking** — Needs separate issue for sprint board visibility
- **Cross-spec** — Tasks from different specs must stay separate

### Terraform Config Files Example

✅ **DO consolidate**:
- `backend.tf` + `versions.tf` — Both are foundational Terraform config, 2-3 lines each, natural pair
- `.env.example` files for backend + frontend (if in same setup phase)

❌ **DON'T consolidate**:
- `golangci-lint.yml` + `eslint.config.js` — Different tech stacks, different dependency chains
- Foundation directory setup + linter config — Different phases, different blockers

## Sprint Planning Workflow (Updated)

### Phase 1: Load and Analyze

1. Load `docs/roadmap.md` (source of truth)
2. Identify target sprint(s) for planning (e.g., Sprint 2-10)
3. Extract all unassigned or target sprint tasks
4. Read relevant `specs/*/spec.md` for value context

### Phase 2: Consolidation Analysis (NEW - MANDATORY)

For each sprint's task list:

1. **Group by area**: Sort tasks by module/directory/layer (e.g., middleware, primitives, config)
2. **Apply consolidation rules**: Identify consolidation candidates using the rules above
3. **Propose groups**: Present grouping recommendations with reasoning
4. **Get approval**: Confirm with user before assigning `Group` values
5. **Document groups**: Assign `Group` column values in `docs/roadmap.md` (e.g., `G-SPRINT2-MIDDLEWARE`)

**Example consolidation proposal**:

```
Sprint 2 Consolidation Candidates:

1. Backend Middleware (5 tasks → 1 issue)
   - 005-T024: request_id.go
   - 005-T025: logger.go
   - 005-T026: recovery.go
   - 005-T027: cors.go
   - 005-T028: body_size.go
   Reason: Same module (middleware), same tech (Go), same pattern (Chi middleware), 
           all parallel-executable, total ~150 lines of code
   Group: G-SPRINT2-BACKEND-MIDDLEWARE

2. Frontend Primitives (7 tasks → 2 issues)
   Group A (4 tasks → 1 issue):
   - 005-T048: Button.tsx
   - 005-T049: Button.module.css
   - 005-T050: Input.tsx
   - 005-T051: Card.tsx
   Reason: Core primitives, same pattern, same dependencies
   Group: G-SPRINT2-PRIMITIVES-CORE

   Group B (3 tasks → 1 issue):
   - 005-T052: LoadingSpinner.tsx
   - 005-T053: ErrorMessage.tsx
   - 005-T054: EmptyState.tsx
   Reason: State components, same pattern, same dependencies
   Group: G-SPRINT2-PRIMITIVES-STATE

Result: Sprint 2 tasks 40 → 30 issues (-25% ticket reduction)
```

### Phase 3: Prioritization

Apply prioritization framework (MoSCoW, RICE, Value vs Effort) to work items (groups + standalone tasks)

### Phase 4: Sprint Assignment

1. Respect dependencies and critical path
2. Assign `Sprint` column values to tasks
3. Balance capacity per sprint (state team size/velocity assumptions)
4. Define Sprint Goal per sprint (one-line outcome)
5. Document in `## Sprint Plan` section of roadmap

### Phase 5: Issue Creation (Scoped)

When asked to create issues for a specific sprint:

1. **Select scope**: Only tasks for requested sprint(s)
2. **Resolve groups**: Collapse grouped tasks into single work items
3. **Skip existing**: Don't recreate if `Issue` URL already present
4. **Preview**: Show what will be created (group + standalone breakdown)
5. **Get confirmation**: MUST confirm before any `gh` write
6. **Create**: One issue per work item (grouped = checklist inside)
7. **Update roadmap**: Write issue URLs back to ALL member rows

**Never create issues for unplanned sprints. Never create all sprints at once unless explicitly requested.**

## Grouping Mechanics

### Group Column

- **Human-owned field** in `docs/roadmap.md`
- **Shared value** = tasks handled by ONE issue (e.g., `G-SPRINT2-MIDDLEWARE`)
- **Blank value** = standalone task (1 task = 1 issue)
- **Must be same spec** — Group ID format: `G-<SPRINT>-<AREA>` or `G-<SPEC>-<AREA>`

### Issue Content for Groups

```markdown
Title: G-SPRINT2-BACKEND-MIDDLEWARE — Create Chi middleware package

Body:
Group-ID: G-SPRINT2-BACKEND-MIDDLEWARE
Stable-IDs: 005-T024, 005-T025, 005-T026, 005-T027, 005-T028
Sprint: 2
Spec: specs/005-system-architecture/
Phase: Phase 3 — Backend Service Architecture

## Tasks

- [ ] 005-T024: Create backend/internal/middleware/request_id.go (UUID v4, X-Request-ID)
- [ ] 005-T025: Create backend/internal/middleware/logger.go (slog JSON, correlation ID)
- [ ] 005-T026: Create backend/internal/middleware/recovery.go (panic recovery, stack trace)
- [ ] 005-T027: Create backend/internal/middleware/cors.go (configurable origins)
- [ ] 005-T028: Create backend/internal/middleware/body_size.go (10 MB limit, 413 response)

## Dependencies

Depends on: #15 (005-T013 — Config package)

## Acceptance Criteria

See specs/005-system-architecture/tasks.md for detailed acceptance criteria per task.

Labels: spec:005, phase:3, priority:p1, sprint:2
```

## Tools and Commands

- **Manual grouping**: Edit `docs/roadmap.md` `Group` column, assign shared value
- **Bulk consolidation**: Use `.github/scripts/consolidate-issues.sh` pattern
- **Issue creation**: Use PM agent "create issues for Sprint N" command
- **Sync**: Use `/sync-issues` for creating ALL untracked (use sparingly)

## Success Metrics

- **Ticket reduction**: Aim for 10-25% fewer issues via consolidation
- **PR reviewability**: Grouped issues = 1 PR, reviewable in 30-60 min
- **Backlog health**: No proliferation of <5-line atomic tasks as separate issues
- **Context preservation**: Grouped tasks share context, reducing context-switching

## Pattern: Sprint 1 Terraform Consolidation

**Before**: 23 issues (001-T022 backend.tf, 001-T023 versions.tf as separate)  
**After**: 22 issues (001-T022 consolidated both Terraform config files)  
**Reason**: Both files are 2-3 lines, foundational Terraform config, natural pair  
**Lesson**: Config files that are naturally paired should be grouped BEFORE issue creation

## Next Steps

1. Apply this workflow to Sprints 2-10
2. Re-analyze task groupings per sprint
3. Update `Group` column in roadmap
4. Create issues sprint-by-sprint with consolidated groups
5. Document grouping decisions in sprint plan section

---

**Retrospective Note**: Sprint 1 created 23 issues, then consolidated 2 post-creation. This workflow prevents that redundancy by consolidating FIRST, then creating issues.
