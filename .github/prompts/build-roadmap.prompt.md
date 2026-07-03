---
description: "Consolidate all spec tasks into a single prioritized roadmap, reconciling against the existing roadmap (add/update/remove) and preserving human metadata"
mode: "agent"
tools: ['read', 'search', 'edit', 'todo']
---

Build and maintain a single consolidated, prioritized roadmap at `docs/roadmap.md` by reconciling the current state of all spec task files. This command is idempotent: running it repeatedly must converge, never duplicate, and never discard human-entered metadata.

## Golden rules
1. Output in English (per the project language policy in `.github/copilot-instructions.md`).
2. This is a RECONCILIATION, not a regeneration. Do NOT rebuild `docs/roadmap.md` from scratch. Diff sources against the existing roadmap and apply the minimal set of add/update/remove operations.
3. NEVER overwrite human-owned fields (Sprint, Priority, Status, Phase, Issue link, Notes). Source files are the source of truth ONLY for task existence, title, and dependencies. Everything a human curates is preserved across runs.
4. Do not create GitHub issues here. Produce a roadmap that a downstream command (or SpecKit) can consume to create issues. Leave the `Issue` field empty for tasks that have no issue yet.
5. Do not modify any `specs/**` files or any application code. You only read specs and write `docs/roadmap.md`.

## Stable task identity (the anchor for reconciliation)
Every roadmap task has a global stable ID: `<spec-number>-<source-task-id>`.
- `<spec-number>` = the numeric prefix of the spec folder (e.g. `001` from `specs/001-product-vision/`).
- `<source-task-id>` = the task identifier used inside that spec's `tasks.md` (e.g. `T003`). If a task has no explicit id, derive a stable slug from its title and record it; keep it stable across runs.
- Example global ID: `001-T003`.
This ID is what links a roadmap row to its source task AND to a future GitHub issue. It must never change for the same underlying task.

## Task grouping (avoid backlog fragmentation)
To prevent a bloated backlog of tiny tickets, tasks can be grouped into a single **work item** that becomes ONE issue. The roadmap has a human-owned `Group` column:
- Tasks that share a `Group` value (e.g. `G-AUTH-1`) are handled by a single issue; the individual tasks become a checklist inside that issue.
- A blank `Group` means the task is its own standalone work item (1 task = 1 issue).
- Grouping is a human/PM decision curated in the roadmap. `/build-roadmap` NEVER auto-groups or ungroups; it only preserves whatever `Group` values already exist. Group tasks only when they share context, the same area, and are small enough that separate tickets would add noise. Do not group across different specs or across a dependency boundary that must be tracked separately.

## Step 1 — Discover sources
- Find every spec task file: `specs/*/tasks.md`.
- Record each spec's number, folder name, and human-readable title (from the spec's `spec.md` if available).
- Detect specs that exist but have no `tasks.md` yet (report them as "pending tasks").
- Detect brand-new spec folders not yet represented in `docs/roadmap.md`.

## Step 2 — Parse tasks from each source
For each `tasks.md`, extract every task with: source task id, title/description, phase/grouping if present, `[P]` parallelizable flag, and intra-spec dependencies. Normalize into candidate rows keyed by the global stable ID.

## Step 3 — Load the existing roadmap (if any)
If `docs/roadmap.md` exists, parse its task table into a map keyed by global stable ID, capturing BOTH source-derived fields and human-owned fields (Group, Sprint, Priority, Status, Phase, Issue, Notes). If it does not exist, treat the existing set as empty and create the file in Step 5.

## Step 4 — Reconcile (compute the diff)
Classify every global ID:
- **ADD** — present in a source `tasks.md`, absent from the roadmap. Insert a new row. Set source-derived fields from the source. Set human-owned fields to sensible defaults: Priority = inherit from spec/phase default or `TBD`, Status = `Backlog`, Issue = empty.
- **UPDATE** — present in both, but source-derived fields (title or dependencies) changed. Refresh ONLY the source-derived fields. Preserve Group, Sprint, Priority, Status, Phase override, Issue, and Notes exactly. Append a short note in a "Changed" log if the title materially changed.
- **UNCHANGED** — present in both, source identical. Leave the row untouched.
- **REMOVE** — present in the roadmap, no longer in any source (refined away). Do NOT hard-delete blindly:
  - If the task has NO Issue link and Status is `Backlog`/`TBD`: remove the row.
  - If the task HAS an Issue link OR Status is beyond Backlog (e.g. In Progress/Done): move it to an `## Archived / Removed` section with a reason (`removed from source on <date>`) and keep its Issue link, so nothing done or tracked is silently lost. Flag it for the human to close/clean up.

Also reconcile whole specs: if a spec folder was deleted, archive its tasks under the same rules.

## Step 5 — Write `docs/roadmap.md`
Produce a clean, human-readable AND machine-parseable file with this structure:

```markdown
# Project Roadmap

> Generated and reconciled by /build-roadmap. Source of truth for task existence,
> titles, and dependencies is specs/*/tasks.md. Priority, Status, Phase, Issue, and
> Notes are human-owned and preserved across runs. Do not hand-edit the stable IDs.

Last reconciled: <YYYY-MM-DD>

## Legend
- Group: shared value = tasks handled by ONE issue (checklist inside); empty = standalone (1 task = 1 issue)
- Priority: P1 (critical) | P2 | P3 | TBD
- Status: Backlog | Ready | In Progress | In Review | Done
- Issue: link to the tracker issue once created (empty = not yet created). Grouped tasks share the same issue URL.

## Foundation phase (specs 001-00X)
<!-- Ordered by cross-spec dependency: vision -> nfrs -> cloud/iac -> security -> architecture -> domain -->

### Spec 001 — <title>  (source: specs/001-*/tasks.md)
| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|
| 001-T001 | ... | | - | P1 | Backlog | - | no | | |

## Feature phase
### Spec 00N — <title>  (source: specs/00N-*/tasks.md)
| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |
|----|------|-------|--------|----------|--------|------------|----------|-------|-------|

## Critical path
<!-- Ordered list of the must-do-in-sequence tasks/specs derived from dependencies -->

## Specs without tasks yet
<!-- specs that exist but have no tasks.md -->

## Archived / Removed
<!-- tasks removed from source but preserved because they had issues or progress -->
| ID | Task | Reason | Issue |
|----|------|--------|-------|
```

Ordering rules:
- Group by spec, with foundation specs before feature specs.
- Within foundation, order specs by cross-spec dependency (vision -> NFRs -> cloud/IaC -> security -> architecture -> domain), consistent with the project's foundation sequence.
- Keep the `Depends on` column populated so a downstream issue-creation command can set issue relationships.

## Step 6 — Reconciliation report
End with a concise summary of what changed this run:
- Specs discovered (and any new/removed since last run).
- Counts: added / updated / unchanged / removed / archived.
- Tasks flagged for human attention (archived items with open issues, priority still `TBD`, specs without tasks).
- Reminder: to turn roadmap rows into issues, run the issue-creation command; only rows with an empty `Issue` field should be created.

Do NOT commit. Leave `docs/roadmap.md` staged for review and suggest `/commit-and-push` if approved.
