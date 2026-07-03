---
description: "Create/sync GitHub issues from docs/roadmap.md: open issues for un-tracked tasks, write back their URLs, and reconcile labels — idempotent, gh CLI"
mode: "agent"
tools: ['read', 'edit', 'execute', 'todo']
---

Turn the consolidated roadmap into GitHub issues and keep them in sync. Read `docs/roadmap.md`, create issues for tasks that don't have one yet, write the issue URL back into the roadmap, and reconcile labels. This command is idempotent: rows that already have an `Issue` link are never recreated.

## Golden rules
1. Output and issue content in English (per the project language policy in `.github/copilot-instructions.md`).
2. External actions have no undo. Creating issues is irreversible — ALWAYS preview and get explicit confirmation before creating anything (see Step 3).
3. The `<stable-id>` (e.g. `001-T003`) is the anchor. It links a roadmap row to its issue. Every issue MUST carry its stable ID so re-runs can match without duplicating.
4. Idempotency: NEVER create an issue for a roadmap row that already has a non-empty `Issue` field. On re-runs, only un-tracked rows are created.
5. Do not modify `specs/**` or application code. You only read `docs/roadmap.md`, write issue URLs back into it, and call `gh`.

## Preconditions
- Verify `gh` is available (`gh --version`) and authenticated (`gh auth status`). If not, stop and tell the user to run `gh auth login`.
- Confirm the repo has a remote on GitHub (`gh repo view`). If not, stop and report.

## Step 1 — Parse the roadmap
Read `docs/roadmap.md` and load every task row with: stable ID, task title, spec origin, Priority, Status, Depends on, Parallel, Issue, Notes. Also read the `## Archived / Removed` section.

## Step 2 — Classify each row
- **TO CREATE** — `Issue` field is empty AND the row is not in Archived. These will become new issues.
- **ALREADY TRACKED** — `Issue` field has a URL. Skip creation; only reconcile labels (Step 5).
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

## Step 5 — Create issues (only for TO CREATE rows)
For each row to create, in roadmap order (foundation specs first, respecting `Depends on`):
1. Build the issue body in English with this structure:
   - A first line marker: `Stable-ID: <stable-id>` (used for matching on future runs).
   - `Spec:` link/path to the source spec (e.g. `specs/001-*/spec.md`).
   - `Task:` the task description.
   - `Depends on:` the stable IDs from the roadmap. If those dependencies already have issue URLs, reference them so the relationship is visible.
   - `Acceptance:` a short line pointing to the spec's acceptance criteria if available.
2. Create the issue:
   `gh issue create --title "<stable-id> — <task title>" --body "<body>" --label "<labels>"`
3. Capture the returned issue URL.
4. Write that URL back into the `Issue` column of the corresponding row in `docs/roadmap.md`.

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

Do NOT commit. Leave `docs/roadmap.md` staged (now containing the new issue URLs) for review and suggest `/commit-and-push` if approved.

## Relationship to other commands
- Run `/build-roadmap` first to consolidate/refresh `docs/roadmap.md`, then `/sync-issues` to push new rows to GitHub.
- Re-running `/build-roadmap` after issues exist preserves the `Issue` URLs (they are human/tooling-owned), so the two commands compose safely in a loop.