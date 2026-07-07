# GitHub Issue Creation Guidelines

> **Updated approach**: Consolidate BEFORE creating issues. One work item = one issue.

**Last updated**: 2026-07-07 (added consolidation policy)

---

## Consolidation-First Workflow (MANDATORY)

**Key Principle**: Consolidate atomic tasks into work items BEFORE creating issues to reduce ticket waste and improve backlog health.

### Why Consolidate?

**Lesson from Sprint 1**: Creating 40+ atomic tasks as separate issues leads to:
- ❌ Ticket proliferation (hard to navigate backlog)
- ❌ Context switching overhead (5 tiny PRs vs 1 cohesive PR)
- ❌ Review fatigue (reviewing 5 x 30-line PRs vs 1 x 150-line PR)

**Sprint 2 with consolidation**: 40 tasks → 19 issues (-52% reduction)

### Core Rules

1. **Consolidate FIRST, Create SECOND** ⚠️  
   Identify consolidation opportunities BEFORE creating any issues
   
2. **One Work Item = One Issue**  
   A work item is either:
   - **Standalone task**: 1 roadmap row = 1 issue
   - **Consolidated group**: 2-4 roadmap rows = 1 issue with checklist
   
3. **Use Group Column in Roadmap**  
   Mark consolidated tasks with shared `Group` value in `docs/roadmap.md` (e.g., `G-BACKEND-MIDDLEWARE`)
   
4. **Group = Checklist in Issue**  
   Grouped tasks become checklist items inside one issue
   
5. **Use Epic Labels for High-Level Organization**  
   Tag all issues with `epic:name` for feature-level grouping
   
6. **Dependency Tracking is MANDATORY** ⚠️  
   After creating issues, MUST run dependency tracking script to establish "Blocked by" / "Blocks" relationships

---

## Consolidation Rules (MANDATORY Before Issue Creation)

### ✅ DO Consolidate When Tasks Meet ALL Criteria

1. **Same spec** — Never cross spec boundaries
2. **Same area/module** — Same directory or logical component (e.g., all middleware)
3. **Same tech stack** — All Go OR all React (never mix languages)
4. **Shared context** — Benefit from being worked together (config files, primitives set)
5. **Similar size** — Each task small enough that 2-4 fit in one reviewable PR
6. **No blocking dependencies** — Can be executed in parallel or sequence without external blockers
7. **Independent tracking not required** — Don't need separate issue status for sprint planning

**Optimal group size**: 2-4 tasks (1 PR, reviewable in 30-60 min)

### ❌ DON'T Consolidate When

- Different tech stacks (Go linting + ESLint = separate)
- Critical blocker task (needs visibility as standalone)
- Different dependency chains (one blocked, other not)
- Already large task (>200 LOC)
- Cross-spec boundary (never group across specs)
- Independent status tracking needed

### Examples

**✅ Good Consolidation** (Chi Middleware):
```
Group: G-BACKEND-MIDDLEWARE (5 tasks → 1 issue)
- request_id.go (UUID generation)
- logger.go (slog JSON)
- recovery.go (panic recovery)
- cors.go (CORS config)
- body_size.go (10 MB limit)

Reason: Same module, same pattern, parallel executable, ~150 LOC total
```

**❌ Bad Consolidation** (Mixed Stacks):
```
DON'T group: golangci-lint.yml + eslint.config.js
Reason: Different tech stacks, different dependency chains
```

**✅ Good Consolidation** (Terraform Config):
```
Group: G-TERRAFORM-CONFIG (2 tasks → 1 issue)
- backend.tf (S3 state backend)
- versions.tf (Terraform/provider versions)

Reason: Both foundational Terraform config, 2-3 lines each, natural pair
```

---

## How to Mark Consolidations in Roadmap

Edit `docs/roadmap.md` and add Group value to the `Group` column:

```markdown
| ID | Task | Group | Sprint | Priority | ... |
|----|------|-------|--------|----------|-----|
| 005-T024 | request_id.go | G-BACKEND-MIDDLEWARE | 2 | P1 | ... |
| 005-T025 | logger.go | G-BACKEND-MIDDLEWARE | 2 | P1 | ... |
| 005-T026 | recovery.go | G-BACKEND-MIDDLEWARE | 2 | P1 | ... |
```

Tasks with the same Group value become ONE issue.

---

## Epic Organization Using Labels

### Epic Label Format: `epic:short-name`

Examples:
- `epic:architecture-foundation` — Groups all architecture setup tasks
- `epic:auth-security` — Groups authentication & security implementation
- `epic:trip-generation` — Groups trip creation & AI integration
- `epic:collaboration` — Groups collaboration & suggestions features

### When to Create an Epic Label:

1. **Scope**: Feature or component spanning 20+ tasks
   - Example: Architecture Foundation (23 tasks across 2 phases)
   - Example: Authentication & Security (64 tasks across multiple specs)

2. **Duration**: Work spans multiple sprints (2+ sprints)
   - Example: Spec 004-008 security implementation (Sprints 5-9)

3. **Business Value**: Represents a complete capability or user story
   - Example: "Paid User Registration & First Trip Creation"
   - Example: "Infrastructure Provisioning & Deployment"

### Issue Creation with Epic Labels

Every issue gets:
- **Epic label**: `epic:architecture-foundation`
- **Spec label**: `spec:005`
- **Sprint label**: `sprint:1`
- **Priority label**: `priority:P1`
- **Type label**: `type:backend` / `type:frontend` / `type:infra`
- **Domain label** (optional): `domain:auth` / `domain:trips` / etc.

---

## Work Package Grouping (Using Projects)

### Check for Active Project

Before creating issues, verify if a GitHub Project exists:

```bash
# List all projects for the repository
gh project list --owner josema.pereiracih --format json

# If no project exists, create one
gh project create --title "TrAIveler MVP Development" --owner josema.pereiracih
```

### Project Structure

**Board Views**:
1. **Sprint Board** — Group by sprint (sprint:1, sprint:2, etc.)
2. **Epic Board** — Group by epic label (epic:architecture-foundation, etc.)
3. **Priority Board** — Group by priority (priority:P1, P2, P3)
4. **Team Board** — Group by type (type:backend, type:frontend, type:infra)

**Custom Fields**:
- **Sprint**: Number field (1-10 for MVP)
- **Epic**: Text field (Architecture Foundation, Auth & Security, etc.)
- **Story Points**: Number field (optional, for velocity tracking)
- **Blocked**: Status field (Yes/No)

### Adding Issues to Project

```bash
# Add issue to project automatically when created
gh issue create \
  --title "005-T001 — Create backend directory structure" \
  --project "TrAIveler MVP Development" \
  --label "epic:architecture-foundation,spec:005,sprint:1,priority:P1,type:backend"

# Or add existing issue to project
gh project item-add <PROJECT-ID> --owner josema.pereiracih --url <ISSUE-URL>
```

### Project Workflow States

Use GitHub Projects status field:

| Status | Description | When to Use |
|--------|-------------|-------------|
| **Backlog** | Not yet ready to start | Default for new issues |
| **Ready** | All dependencies met, can start | Dependencies resolved |
| **In Progress** | Actively being worked on | PR opened or work started |
| **In Review** | PR created, awaiting review | PR submitted |
| **Done** | Merged and closed | PR merged |

## GitHub Projects (v2) Organization

### Project Setup

Before creating issues, verify GitHub Project exists:

```bash
# List projects
gh project list --owner JosemaPereira

# Create if needed
gh project create --title "TrAIveler MVP Development" --owner JosemaPereira
```

### Project Custom Fields

Add these custom fields to your project:

| Field Name | Type | Values | Purpose |
|------------|------|--------|---------|
| **Group** | Text | G-ARCH-SETUP-DIRS, etc. | Work package grouping |
| **Sprint** | Number | 1-10 | Sprint assignment |
| **Story Points** | Number | 1-8 | Effort estimation (optional) |
| **Blocked** | Single Select | Yes, No | Dependency tracking |

### Project Views (Recommended)

**1. Sprint Board** (Primary)
```
View: Board
Group by: Sprint
Filter: sprint:<current>
Status columns: Backlog → Ready → In Progress → Review → Done
```

**2. Epic View**
```
View: Table  
Group by: Labels (filter epic:*)
Show: Issue, Status, Assignee, Sprint
Sort by: Sprint ascending
```

**3. Group View** (Work Packages)
```
View: Board
Group by: Group custom field
Filter: sprint:1
Shows: G-ARCH-SETUP-DIRS, G-ARCH-SETUP-INIT, etc.
```

**4. Dependency View**
```
View: Table
Columns: Issue, Blocked By, Blocks, Status
Filter: Blocked = Yes
```

---

## Simplified Issue Creation Workflow

### Step 1: Prepare Sprint Scope

From `docs/roadmap.md`, identify:
- Tasks for the sprint (e.g., Sprint 1 = 23 tasks from 005-T001 to 005-T023)
- Epic label (e.g., `epic:architecture-foundation`)
- Group assignments from roadmap `Group` column

### Step 2: Create Issues (One Per Task)

For each task, create a single issue:

```bash
gh issue create \
  --repo JosemaPereira/capstone-project-ai-bootcamp \
  --title "005-T001 — Create backend directory structure" \
  --body "$(cat issue-template.md)" \
  --label "epic:architecture-foundation,spec:005,sprint:1,priority:P1,type:backend"
```

### Step 3: Add to Project

```bash
# Get project ID
gh project list --owner JosemaPereira

# Add all sprint issues
for issue_num in {13..35}; do
  gh project item-add 2 --owner JosemaPereira \
    --url "https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/$issue_num"
done
```

### Step 4: Set Custom Fields (In Projects UI)

For grouped tasks, manually add "Group" field value:
- Issues #20-22 → Set Group = `G-ARCH-SETUP-DIRS`
- Issues #23-24 → Set Group = `G-ARCH-SETUP-INIT`
- etc.

### Step 5: Update Roadmap

Write issue URLs back to `docs/roadmap.md` Issue column.

---

## Sprint Organization Example

### Sprint 1 (Simplified Approach)

**23 tasks → 23 issues**

**Standalone Issues** (no group):
- #13: 005-T001 — Backend directory structure
- #14: 005-T005 — Initialize Go module
- #15: 005-T013 — Create config loader
- #16: 005-T020 — Configure golangci-lint
- #17: 005-T021 — Configure ESLint/Prettier
- #18: 005-T022 — Terraform backend.tf
- #19: 005-T023 — Terraform versions.tf

**Grouped Issues** (share Group value):
- **G-ARCH-SETUP-DIRS**: #20-22 (3 tasks)
- **G-ARCH-SETUP-INIT**: #23-24 (2 tasks)
- **G-ARCH-SETUP-CONFIG**: #25-26 (2 tasks)
- **G-ARCH-SETUP-DOCS**: #27-29 (3 tasks)
- **G-ARCH-FOUNDATIONAL-DOCKER**: #30-32 (3 tasks)
- **G-ARCH-FOUNDATIONAL-CI**: #33-35 (3 tasks)

**How to View Groups**:
```
In Projects:
1. Add custom field "Group" 
2. Create board view grouped by "Group"
3. See all G-ARCH-SETUP-DIRS tasks together
```

**No parent issues needed!** GitHub Projects handles the grouping.

---

## Issue Relationships

### Relationship Types

Use issue description sections to document relationships:

```markdown
## Dependencies & Relationships

### Blocks
- #123 — Trip service implementation
- #124 — API handlers

### Blocked By
- #115 — Database migrations (must complete first)
- #110 — Config loader (required)

### Related To
- #130 — User repository (parallel work)
- #135 — Subscription service (shared patterns)
```

GitHub will:
- Auto-link issue numbers (#123 becomes clickable)
- Show relationship graph in issue sidebar
- Track blocking/blocked status

---

## Label Strategy

### Epic Labels (Primary Grouping)

| Epic Label | Description | Sprints | Tasks |
|------------|-------------|---------|-------|
| `epic:architecture-foundation` | System Architecture setup | 1 | 23 |
| `epic:architecture-backend` | Backend patterns | 2 | 18 |
| `epic:architecture-frontend` | Frontend patterns | 2 | 19 |
| `epic:architecture-infra` | Infrastructure modules | 2-3 | 49 |
| `epic:integration-observability` | Integration layer | 4 | 29 |
| `epic:auth-security` | Authentication & security | 5 | 64 |
| `epic:data-layer` | Repositories & models | 6 | 22 |
| `epic:frontend-shell` | UI components | 7 | 28 |
| `epic:trip-generation` | Core MVP features | 8 | 44 |
| `epic:security-hardening` | NFR gates | 9 | 56 |
| `epic:deployment` | AWS infrastructure | 10 | 44 |

### Standard Labels

**Spec Labels**: `spec:001`, `spec:002`, `spec:003`, etc.
- Links issue to specification document

**Sprint Labels**: `sprint:1`, `sprint:2`, ..., `sprint:10`
- Indicates target sprint for GitHub Projects filtering

**Priority Labels**: `priority:P1`, `priority:P2`, `priority:P3`
- P1 = Critical (MVP blocker)
- P2 = Important (post-MVP)
- P3 = Nice-to-have

**Type Labels**: `type:backend`, `type:frontend`, `type:infra`, `type:e2e`
- Indicates primary implementation area

**Domain Labels**: `domain:auth`, `domain:trips`, `domain:collaboration`, `domain:ai`, `domain:security`, `domain:observability`
- Functional domain grouping

**Status Labels**: `status:blocked`, `status:ready`, `status:needs-review`
- Workflow state indicators (supplement GitHub Projects status)

---

## Sprint Issue Creation Workflow (Deprecated)

### Step 1: Verify GitHub Project Exists

```bash
# Check if project exists
gh project list --owner josema.pereiracih

# If not, create it
gh project create \
  --title "TrAIveler MVP Development" \
  --owner josema.pereiracih \
  --format json > project-info.json

# Get project ID for later use
PROJECT_ID=$(cat project-info.json | jq -r '.id')
```

### Step 2: Create Epic Labels

For Sprint 1, create the epic label if it doesn't exist:

```bash
# Create epic label
gh label create "epic:architecture-foundation" \
  --description "System Architecture Foundation (Spec 005 Phase 1-2)" \
  --color "0E8A16"
```

Epic label colors:
- Foundation epics: `0E8A16` (green)
- Feature epics: `1D76DB` (blue)
- Security epics: `D93F0B` (red)
- Infrastructure epics: `FBCA04` (yellow)

### Step 3: Group Tasks into Work Items

Following the roadmap's `Group` column:
- Tasks with same `Group` value → Create parent issue + individual sub-issues
- Tasks with empty `Group` → Create standalone issue

Sprint 1 work items:
- 6 parent issues (groups with 2-3 sub-issues each)
- 5 standalone issues
- **Total: 11 parent + 18 sub-issues = 29 GitHub issues for 23 tasks**

### Step 4: Create Standalone Issues First

```bash
# Standalone issue
gh issue create \
  --title "005-T001 — Create backend directory structure" \
  --body "$(cat standalone-issue-template.md)" \
  --label "epic:architecture-foundation,spec:005,sprint:1,priority:P1,type:backend" \
  --project "$PROJECT_ID"
```

### Step 5: Create Sub-Issues for Groups

For each group, create individual issues first (to get issue numbers), then create parent:

```bash
# Create sub-issues first
gh issue create \
  --title "005-T002 — Create frontend directory structure" \
  --body "$(cat sub-issue-template.md)" \
  --label "epic:architecture-foundation,spec:005,sprint:1,priority:P1,type:frontend" \
  --project "$PROJECT_ID"
# Returns: #101

gh issue create \
  --title "005-T003 — Create e2e directory structure" \
  --body "$(cat sub-issue-template.md)" \
  --label "epic:architecture-foundation,spec:005,sprint:1,priority:P1,type:e2e" \
  --project "$PROJECT_ID"
# Returns: #102

gh issue create \
  --title "005-T004 — Create infrastructure directory structure" \
  --body "$(cat sub-issue-template.md)" \
  --label "epic:architecture-foundation,spec:005,sprint:1,priority:P1,type:infra" \
  --project "$PROJECT_ID"
# Returns: #103

# Now create parent with sub-issue references
gh issue create \
  --title "G-ARCH-SETUP-DIRS — Project Directory Structure Setup" \
  --body "$(cat parent-issue-template.md)" \
  --label "epic:architecture-foundation,spec:005,sprint:1,priority:P1,group" \
  --project "$PROJECT_ID"
```

Parent issue body includes tasklist:
```markdown
## Tasks

- [ ] #101 — 005-T002: Create frontend directory structure
- [ ] #102 — 005-T003: Create e2e directory structure
- [ ] #103 — 005-T004: Create infrastructure directory structure
```

GitHub automatically creates parent-child relationships from `- [ ] #<number>`.

### Step 6: Update Roadmap with Issue URLs

After creating each issue, write the URL back to roadmap:

```bash
# For standalone task
# Update 005-T001 Issue column with URL

# For grouped tasks
# Update 005-T002, 005-T003, 005-T004 Issue columns with their individual URLs
# Update Group row Issue column with parent URL (if roadmap tracks groups separately)
```

</details>

---

## Issue Body Template

Use this template for all issues:

```markdown
## Task Description

**Stable ID**: 005-T001
**Spec**: 005 — System Architecture and Technology Stack
**Phase**: Phase 1 — Setup
**Sprint**: 1
**Epic**: Architecture Foundation
**Priority**: P1
**Group**: (blank for standalone, or G-ARCH-SETUP-DIRS, etc.)

[Brief description of what needs to be done]

## Implementation Details

[Specific technical details, code snippets, file paths, etc.]

## Acceptance Criteria

- [ ] [Criterion 1]
- [ ] [Criterion 2]
- [ ] [Criterion 3]

## Dependencies & Relationships

### Blocks
- #[issue-number] — [Description]

### Blocked By
- #[issue-number] — [Description]

### Related To
- #[issue-number] — [Description]

## Reference

- **Spec**: \`specs/[spec-folder]/plan.md\`
- **Roadmap**: \`docs/roadmap.md\` ([stable-id])
```

---

## Example: Standalone Issue

```markdown
## Task Description

**Stable ID**: 005-T001
**Spec**: 005 — System Architecture
**Sprint**: 1
**Epic**: Architecture Foundation
**Priority**: P1

Create backend directory structure per plan.md specifications.

## Implementation Details

Directory structure to create:
\`\`\`
backend/
├── cmd/api/              # Main entry point
├── internal/
│   ├── middleware/       # HTTP middleware
│   ├── database/         # DB connection
│   ├── ai/               # AI client
│   └── errors/           # Error handling
├── pkg/                  # Shared packages
├── config/               # Configuration
├── migrations/           # SQL migrations
└── tests/
    ├── integration/      # Integration tests
    └── fixtures/         # Test data
\`\`\`

## Acceptance Criteria

- [ ] All directories created with correct structure
- [ ] Matches \`specs/005-system-architecture/plan.md\`
- [ ] \`.gitkeep\` files in empty directories
- [ ] Structure documented in backend/README.md

## Dependencies & Relationships

### Blocks
- #104 — Initialize Go module (needs directory structure)
- #111 — Create config loader (needs pkg/ directory)

### Blocked By
- None (first task in sprint)

### Related To
- #101 — Frontend directory (same pattern, different area)
- #102 — E2E directory (same pattern, different area)

## Reference

- **Spec**: \`specs/005-system-architecture/plan.md\`
- **Roadmap**: \`docs/roadmap.md\` (005-T001)
```

---

## Relationship Management During Sprint

### When Starting Work on an Issue

1. **Check "Blocked By" section** — ensure all dependencies are closed
2. **Update GitHub Projects status** — move from "Backlog" to "In Progress"
3. **Add comment** — "Starting work on this issue"
4. **Link PR** — create draft PR early, use "Closes #123" in description

### When Completing an Issue

1. **Close with PR merge** — use "Closes #123" keyword in PR
2. **GitHub auto-updates** — Projects status changes to "Done"
3. **Check "Blocks" section** — comment on blocked issues that dependency is resolved
4. **Update roadmap** — change Status to "Done", verify Issue URL exists

### When an Issue is Blocked

1. **Add `status:blocked` label**
2. **Update Projects** — set Blocked field to "Yes"
3. **Comment** with reason: "Blocked by #115 (migrations incomplete)"
4. **Update "Blocked By" section** in description
5. **Notify blocking issue** — add comment requesting status update

---

## GitHub Projects Views

### Recommended Project Views

**1. Sprint Board** (Kanban):
```
Group by: Sprint (sprint:1, sprint:2, ...)
Columns: Backlog → Ready → In Progress → In Review → Done
Filter: sprint:<current-sprint>
```

**2. Epic Board** (Grouped):
```
Group by: Epic Label (epic:architecture-foundation, epic:auth-security, ...)
Show: All issues with epic labels
Useful for: Seeing progress across epics
```

**3. Dependencies View** (Table):
```
Columns: Issue | Blocks | Blocked By | Status | Assignee
Sort by: Blocked (Yes first)
Useful for: Identifying blockers
```

**4. Team Velocity** (Chart):
```
X-axis: Sprint
Y-axis: Completed tasks
Filter: Closed issues by sprint
Useful for: Velocity tracking
```

### Querying in Projects

```bash
# View all blocked issues
epic:architecture-foundation status:blocked

# View ready-to-start issues (no blockers)
sprint:1 -status:blocked -label:status:blocked

# View all sub-issues of a parent
parent:#100

# View issues by type
type:backend sprint:1
```

---

## Tools & Automation

### GitHub CLI Commands

```bash
# List all issues in a sprint
gh issue list --label "sprint:1" --state all

# List issues by epic
gh issue list --label "epic:architecture-foundation" --state open

# Find blocked issues
gh issue list --label "status:blocked" --state open

# Update issue labels
gh issue edit 123 --add-label "status:ready" --remove-label "status:blocked"

# Link issues in comments
gh issue comment 123 --body "Unblocked by #115 (migrations complete)"

# Close issue with reference
gh issue close 123 --comment "Completed in PR #456"

# Add issue to project
gh project item-add 2 --owner JosemaPereira --url https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/123

# Set up dependency tracking (cross-references)
# Run the provided script to add relationship comments for all issues
.github/scripts/setup-issue-relationships.sh
```

### Bulk Operations

```bash
# Add all Sprint 1 issues to project
for issue in $(gh issue list --label "sprint:1" --json number -q '.[].number'); do
  gh project item-add 2 --owner JosemaPereira --url "https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/$issue"
done

# Update all blocked issues to ready when unblocked
gh issue list --label "status:blocked" --json number -q '.[].number' | xargs -I {} \
  gh issue edit {} --remove-label "status:blocked" --add-label "status:ready"

# Set up dependency tracking for a sprint
# (Creates cross-reference comments showing blocked-by/blocks relationships)
.github/scripts/setup-issue-relationships.sh
```

### Dependency Tracking Script

**MANDATORY** for all sprints: After creating issues, establish dependency tracking.

The script template (`.github/scripts/setup-issue-relationships.sh`) must be customized for each sprint:

1. **Read the roadmap** — Identify all "Depends on" relationships for your sprint
2. **Map to issue numbers** — Match stable IDs to created issue numbers
3. **Customize script** — Add `add_relationship()` calls for each dependency
4. **Run script** — Execute to create cross-reference comments

```bash
# 1. Edit the script with your sprint's dependencies
vim .github/scripts/setup-issue-relationships.sh

# 2. Run it
chmod +x .github/scripts/setup-issue-relationships.sh
.github/scripts/setup-issue-relationships.sh
```

**What it does:**
- Adds "Blocked by #N" comments to dependent issues
- Adds "Blocks #A, #B, #C" comments to foundation issues  
- Creates cross-references visible in GitHub timeline
- Makes dependencies visible in issue lists and Projects
- Idempotent (safe to re-run)

**Why it's required:**
- GitHub's native relationship UI requires manual linking OR cross-reference comments
- This script automates the cross-referencing
- Makes dependency chains visible without manual clicking
- Essential for dependency tracking in Projects

---

## Sprint Planning Checklist

### Before Creating Issues

- [ ] ✅ Verify GitHub Project exists (`gh project list`)
- [ ] ✅ Create epic label if new (`gh label create "epic:name"`)
- [ ] ✅ Review task groups from roadmap `Group` column (for Projects organization)
- [ ] ✅ Check cross-spec dependencies (e.g., 002 extends 001 CI)
- [ ] ✅ Verify critical path dependencies (foundation before features)
- [ ] ✅ Prepare issue template for tasks
- [ ] ✅ Create standard labels if missing (`spec:NNN`, `sprint:N`, `priority:PN`, `type:*`)

### During Issue Creation

- [ ] ✅ Create one issue per task (1:1 mapping with roadmap)
- [ ] ✅ Document all relationships (Blocks/Blocked By/Related To)
- [ ] ✅ Add appropriate labels (epic, spec, sprint, priority, type)
- [ ] ✅ Add all issues to GitHub Project
- [ ] ✅ Update roadmap `Issue` column with URLs

### After Issue Creation

- [ ] ✅ **MANDATORY: Customize and run dependency tracking script** (`.github/scripts/setup-issue-relationships.sh`)
- [ ] ✅ Verify dependency chains complete (check cross-references in issues)
- [ ] ✅ Add `status:blocked` label to issues with unmet dependencies
- [ ] ✅ Confirm all issues appear in Project board
- [ ] ✅ Update sprint plan in roadmap with issue counts and dependency summary
- [ ] ✅ Set up Project custom fields (Group, Story Points, Blocked)
- [ ] ✅ Configure Project views (Sprint Board, Epic Board, Group Board, Dependencies View)
- [ ] ✅ Set Group values for grouped tasks in Projects UI

---

## Sprint 1 Example (Simplified Approach)

### What Was Created

**Epic**: Architecture Foundation  
**Sprint**: 1  
**Total**: 23 issues (one per task)

**Standalone Issues** (no Group):
- #13: 005-T001 — Backend directory structure
- #14: 005-T005 — Initialize Go module  
- #15: 005-T013 — Create config loader
- #16: 005-T020 — Configure golangci-lint
- #17: 005-T021 — Configure ESLint/Prettier
- #18: 005-T022 — Terraform backend.tf
- #19: 005-T023 — Terraform versions.tf

**Grouped Issues** (share Group value in Projects):
- **G-ARCH-SETUP-DIRS** (3 tasks): #20, #21, #22
- **G-ARCH-SETUP-INIT** (2 tasks): #23, #24
- **G-ARCH-SETUP-CONFIG** (2 tasks): #25, #26
- **G-ARCH-SETUP-DOCS** (3 tasks): #27, #28, #29
- **G-ARCH-FOUNDATIONAL-DOCKER** (3 tasks): #30, #31, #32
- **G-ARCH-FOUNDATIONAL-CI** (3 tasks): #33, #34, #35

### How to View Groups in Projects

1. Add custom field "Group" (text type)
2. For issues #20-22, set Group = `G-ARCH-SETUP-DIRS`
3. For issues #23-24, set Group = `G-ARCH-SETUP-INIT`
4. Create board view grouped by "Group"

**Result**: See all related tasks together without parent issues!

### Lessons Learned

**❌ What we tried first**: Created 6 parent issues + 23 task issues = 29 total  
**✅ What works better**: Created 23 task issues only, use Projects for grouping

**Why simplified is better**:
- Less overhead (23 issues instead of 29)
- No duplicate tracking (no parent % completion vs. task status)
- Projects views provide all needed grouping
- Easier to update (just task issues)
- Scales better (Sprint 2 = 62 issues, not 82)

---

## Tools & Automation

### GitHub CLI Commands

```bash
# List all issues in a sprint
gh issue list --label "sprint:1" --state all

# List issues by epic
gh issue list --label "epic:architecture-foundation" --state open

# Find blocked issues
gh issue list --label "status:blocked" --state open

# Update issue labels
gh issue edit 123 --add-label "status:ready" --remove-label "status:blocked"

# Link issues in comments
gh issue comment 123 --body "Unblocked by #115 (migrations complete)"

# Close issue with reference
gh issue close 123 --comment "Completed in PR #456"

# Add issue to project
gh project item-add 2 --owner JosemaPereira --url https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/123
```

### Bulk Operations

```bash
# Add all Sprint 1 issues to project
for issue in $(gh issue list --label "sprint:1" --json number -q '.[].number'); do
  gh project item-add 2 --owner JosemaPereira --url "https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/$issue"
done

# Close multiple parent issues (if simplifying)
for issue_num in {36..41}; do
  gh issue close $issue_num --repo JosemaPereira/capstone-project-ai-bootcamp \
    --comment "Closing parent issue - using Projects for grouping instead"
done
```

---

❌ **DON'T**: Create separate epic issues
- Use `epic:name` labels instead
- GitHub Projects provides epic-level views

❌ **DON'T**: Use milestones for sprints
- Use `sprint:N` labels instead
## Anti-Patterns to Avoid

❌ **DON'T**: Create parent/epic issues + task issues  
**Why**: Over-engineering — adds unnecessary overhead (29 issues instead of 23)  
**DO**: Create one issue per task, use Projects for grouping

❌ **DON'T**: Create separate epic issues  
**Why**: Epic labels + Projects views provide same organization  
**DO**: Use `epic:name` labels instead

❌ **DON'T**: Use milestones for sprints  
**Why**: Labels are more flexible and better for filtering  
**DO**: Use `sprint:N` labels instead (milestones for releases like v1.0)

❌ **DON'T**: Skip relationship documentation  
**Why**: Loses critical dependency tracking  
**DO**: Always fill "Dependencies & Relationships" section

❌ **DON'T**: Create circular dependencies  
**Why**: Deadlocks the workflow  
**DO**: Review dependency graph — Foundation → Features → Polish (never backwards)

❌ **DON'T**: Forget to add issues to GitHub Project  
**Why**: Loses sprint visibility  
**DO**: Use `gh project item-add` after creation

❌ **DON'T**: Create issues without confirmation  
**Why**: Irreversible, hard to clean up mistakes  
**DO**: Always preview and get explicit approval first

---

## Quick Reference Card

| Scenario | Action |
|----------|--------|
| **New epic scope** | Create `epic:name` label |
| **Standalone task** | Create single issue with epic label |
| **Task group (2-10 tasks)** | Create issues with same Group value, use Projects to visualize |
| **Cross-spec dependency** | Document in "Blocked By" section |
| **Foundation work** | Use `epic:architecture-foundation` label |
| **User story** | Use `epic:trip-generation` or similar epic label |
| **Before starting work** | Check "Blocked By", ensure dependencies closed |
| **After completing work** | Close issue, update roadmap Status to "Done" |
| **Need to see epic progress** | Use GitHub Projects Epic Board view |
| **Need to see sprint progress** | Use GitHub Projects Sprint Board view |
| **Need to see work packages** | Use GitHub Projects Group Board view |
| **Find blocked issues** | Filter by `status:blocked` label |

---

**Last updated**: 2026-07-06  
**Maintained by**: Product Manager Agent  
**References**: 
- `docs/roadmap.md` — Source of truth for task tracking
- `.github/prompts/plan-sprints.prompt.md` — Sprint planning workflow
- `.github/prompts/create-sprint-issues.prompt.md` — Simplified issue creation workflow
- [GitHub Projects Documentation](https://docs.github.com/en/issues/planning-and-tracking-with-projects)
- [GitHub Labels Documentation](https://docs.github.com/en/issues/using-labels-and-milestones-to-track-work)
