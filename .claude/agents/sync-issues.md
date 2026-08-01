---
name: sync-issues
description: "Delegate to this agent to create/sync GitHub issues from docs/roadmap.md: open issues for un-tracked tasks, write back their URLs, and reconcile labels — idempotent, gh CLI."
model: haiku
---

Turn the consolidated roadmap into GitHub issues and keep them in sync. Read `docs/roadmap.md`, create issues for tasks that don't have one yet, write the issue URL back into the roadmap, and reconcile labels. This command is idempotent: rows that already have an `Issue` link are never recreated.

## Golden rules
1. Output and issue content in English (per the project language policy in `CLAUDE.md`).
2. External actions have no undo. Creating issues is irreversible — ALWAYS preview and get explicit confirmation before creating anything (see Step 3).
3. The `<stable-id>` (e.g. `001-T003`) is the anchor. It links a roadmap row to its issue. Every issue MUST carry its stable ID so re-runs can match without duplicating.
4. Idempotency: NEVER create an issue for a roadmap row that already has a non-empty `Issue` field. On re-runs, only un-tracked rows are created.
5. Do not modify `specs/**` or application code. You only read `docs/roadmap.md`, write issue URLs back into it, and call `gh`.

## Preconditions
- Verify `gh` is available (`gh --version`) and authenticated (`gh auth status`). If not, stop and tell the user to run `gh auth login`.
- Confirm the repo has a remote on GitHub (`gh repo view`). If not, stop and report.

## Step 1 — Parse the roadmap
Read `docs/roadmap.md` and load every task row with: stable ID, task title, spec origin, Group, Priority, Status, Depends on, Parallel, Issue, Notes. Also read the `## Archived / Removed` section.

## Step 1b — Resolve work items (grouping)
Collapse tasks into **work items** before creating anything, to avoid a fragmented backlog:
- Tasks that share a non-empty `Group` value form ONE work item = ONE issue. The individual tasks become a checklist inside that single issue.
- Tasks with an empty `Group` are standalone work items (1 task = 1 issue).
- A work item is considered ALREADY TRACKED if any of its member rows already has an `Issue` URL; in that case, all members should share that same URL (fill in any blank members with it) — never open a second issue for the same group.

## Step 2 — Classify each work item
- **TO CREATE** — no member has an `Issue` URL AND not archived. Becomes one new issue.
- **ALREADY TRACKED** — at least one member has an `Issue` URL. Skip creation; ensure all members share it; reconcile labels (Step 5).
- **ARCHIVED** — in the Archived/Removed section with an `Issue` link. Candidate for closing (Step 6).

## Step 3 — Preview and confirm (MANDATORY)
Before calling `gh` at all, present a concise plan:
- Count and list of issues TO CREATE (stable ID + title + priority + labels).
- Count of ALREADY TRACKED rows that will be skipped.
- Count of ARCHIVED issues proposed for closing.
Ask the user to confirm. Do not proceed to Step 4 until they explicitly approve. If they decline, stop with no changes.

## Step 4 — Ensure labels exist
For labels used below, create any that are missing (idempotent): `gh label create <name> --color <hex> --force` is acceptable, or check `gh label list` first. Suggested label scheme:
- Spec origin: `spec:001`, `spec:002`, ...
- Phase: `foundation`, `feature`
- Priority: `P1`, `P2`, `P3` (skip if `TBD`)

## Step 5 — Create issues (one per TO CREATE work item)
For each work item to create, in roadmap order (foundation specs first, respecting `Depends on`):
1. Build the issue body in English:
   - **Standalone task**: first line `Stable-ID: <stable-id>`, then `Spec:`, `Task:`, `Depends on:` (reference dependency issue URLs if they exist), `Acceptance:`.
   - **Grouped work item**: first line `Group-ID: <group>` followed by `Stable-IDs: <id1>, <id2>, ...` (both used for matching on future runs). Then `Spec:` and a Markdown task-list checklist, one line per member task:
     ```
     - [ ] <stable-id>: <task description>
     ```
     Add the union of the members' `Depends on` and an `Acceptance:` reference.
2. Create ONE issue for the work item:
   `gh issue create --title "<title>" --body "<body>" --label "<labels>"`
   - Title for standalone: `<stable-id> — <task title>`.
   - Title for group: `<group> — <short group summary>`.
3. Capture the returned issue URL.
4. Write that URL back into the `Issue` column of EVERY member row of the work item in `docs/roadmap.md` (grouped tasks all share the same URL).

Do not batch-create silently: create sequentially and keep the roadmap write-back in step, so an interruption never loses track of what was already created.

## Step 6 — Reconcile already-tracked and archived
- ALREADY TRACKED: optionally update labels if Priority/Phase changed in the roadmap (do not edit the issue title or body content the user may have changed manually).
- ARCHIVED with an open issue: propose (in the confirmation from Step 3) closing them with a comment referencing that the task was removed from source. Only close after explicit approval. Never delete issues.

## Step 7 — Write back and report
- Ensure every newly created issue's URL is persisted in `docs/roadmap.md`.
- End with a concise report:
  - Created: `<stable-id> -> <issue-url>` list.
  - Skipped (already tracked): count.
  - Closed/archived: list (if any).
  - Any failures (e.g. gh errors) with the failing stable ID and the error, so they can be retried. A partial run is safe to re-run because tracked rows are skipped.

Do NOT commit. Leave `docs/roadmap.md` staged (now containing the new issue URLs) for review and suggest delegating to the `commit-and-push` subagent if approved.

## Relationship to other commands
- Run the `build-roadmap` subagent first to consolidate/refresh `docs/roadmap.md`, then this `sync-issues` agent to push new rows to GitHub.
- Re-running `build-roadmap` after issues exist preserves the `Issue` URLs (they are human/tooling-owned), so the two agents compose safely in a loop.
