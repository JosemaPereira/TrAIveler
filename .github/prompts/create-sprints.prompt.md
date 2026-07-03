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
   - If stable IDs were given, select exactly those rows.
3. From the selected rows, keep ONLY those with an empty `Issue` field (idempotent — never recreate tracked rows). Report any selected rows already tracked as skipped.
4. Verify `gh` is available and authenticated (`gh --version`, `gh auth status`). If not, stop and ask the user to run `gh auth login`.
5. Ensure required labels exist (create if missing): `spec:<n>`, `foundation`/`feature`, `P1`/`P2`/`P3`, `sprint:<n>`.
6. PREVIEW (mandatory): list exactly what will be created — stable ID, title, priority, sprint, labels — plus the count of skipped/already-tracked rows. Get explicit confirmation. Do NOT proceed without it.
7. Create each issue in dependency order, carrying the stable-ID anchor in the body:
   `gh issue create --title "<stable-id> — <title>" --body "Stable-ID: <stable-id>\nSprint: <n>\nSpec: <path>\nTask: <desc>\nDepends on: <ids/urls>\nAcceptance: <ref>" --label "spec:<n>,<phase>,<priority>,sprint:<n>"`
   If a dependency already has an issue URL, reference it so the relationship is visible.
8. After each creation, write the returned URL back into the `Issue` column of that row in `docs/roadmap.md` (write-back in step, so an interruption is safe to resume).
9. NEVER create issues outside the requested scope.
10. Report: created issues (`<stable-id> -> <url>`), skipped (already tracked), and any failures with stable ID + error for safe retry.

Do NOT commit. Leave `docs/roadmap.md` staged (now with the new issue URLs) and suggest `/commit-and-push` if approved.