---
description: "Create GitHub issues for ONLY a specific sprint or specific task(s) from the roadmap — scoped, idempotent, confirm-first, gh CLI"
mode: "agent"
agent: "product-manager"
tools: ['read', 'edit', 'execute', 'todo']
---

Act as a Product Manager and create GitHub issues for a specific scope only — a single sprint or one/few tasks from `docs/roadmap.md`. Switch to the `product-manager` agent if not already active.

Scope input (required): ${input:scope:Required. Either a sprint (e.g. "Sprint 2" or "sprint:2") or one or more stable task IDs (e.g. "001-T003" or "001-T003, 002-T001").}

Instructions:
1. Load `docs/roadmap.md`. If it does not exist, tell the user to run `/build-roadmap` (and `/plan-sprints`) first and stop.
2. Resolve the scope:
   - If a sprint was given, select all rows whose `Sprint` matches.
   - If stable IDs were given, select exactly those rows (plus their group siblings — see next step).
3. Resolve WORK ITEMS (grouping, to avoid a fragmented backlog):
   - Rows sharing a non-empty `Group` value collapse into ONE work item = ONE issue; member tasks become a checklist. If a selected stable ID belongs to a group, include its group siblings so the whole work item is created together.
   - Rows with an empty `Group` are standalone work items (1 task = 1 issue).
   - Skip any work item where a member already has an `Issue` URL; fill any blank member with that existing URL instead of creating a duplicate. Report as skipped/already-tracked.
   - If several in-scope tasks are small and clearly related but ungrouped, you may PROPOSE grouping them first (with reasoning); on approval, set their `Group` in the roadmap before creating.
4. Verify `gh` is available and authenticated (`gh --version`, `gh auth status`). If not, stop and ask the user to run `gh auth login`.
5. Ensure required labels exist (create if missing): `spec:<n>`, `foundation`/`feature`, `P1`/`P2`/`P3`, `sprint:<n>`.
6. PREVIEW (mandatory): list exactly what will be created PER WORK ITEM — for groups show the group id, summary title, and member checklist; for standalone show stable ID + title — plus priority, sprint, labels, and the count of skipped/already-tracked. Get explicit confirmation. Do NOT proceed without it.
7. Create ONE issue per work item, in dependency order:
   - Standalone: `gh issue create --title "<stable-id> — <title>" --body "Stable-ID: <stable-id>\nSprint: <n>\nSpec: <path>\nTask: <desc>\nDepends on: <ids/urls>\nAcceptance: <ref>" --label "spec:<n>,<phase>,<priority>,sprint:<n>"`
   - Group: title `"<group> — <summary>"`; body starts `Group-ID: <group>` and `Stable-IDs: <id1>, <id2>...`, then a `- [ ] <stable-id>: <desc>` checklist per member, plus the union of dependencies and an acceptance ref.
   If a dependency already has an issue URL, reference it so the relationship is visible.
8. After each creation, write the returned URL back into the `Issue` column of EVERY member row of that work item in `docs/roadmap.md` (write-back in step, so an interruption is safe to resume).
9. NEVER create issues outside the requested scope.
10. Report: created issues per work item (`<group|stable-id> -> <url>`), skipped (already tracked), and any failures with ID + error for safe retry.

Do NOT commit. Leave `docs/roadmap.md` staged (now with the new issue URLs) and suggest `/commit-and-push` if approved.
