---
description: "Create GitHub issues for a specific sprint or task(s) — one issue per task, organized with epic labels and Projects"
agent: "product-manager"
tools: ['read', 'edit', 'execute', 'todo']
---

Act as a Product Manager and create GitHub issues for a specific scope — a single sprint or one/few tasks from `docs/roadmap.md`. Switch to the `product-manager` agent if not already active.

Scope input (required): ${input:scope:Required. Either a sprint (e.g. "Sprint 2" or "sprint:2") or one or more stable task IDs (e.g. "001-T003" or "001-T003, 002-T001").}

**IMPORTANT**: Read `.github/ISSUE-CREATION-GUIDELINES.md` before proceeding.

## Simplified Workflow (One Issue Per Task)

Instructions:
1. Load `docs/roadmap.md`. If it does not exist, tell the user to run `/build-roadmap` and `/plan-sprints` first and stop.

2. **Verify GitHub Project exists**: Run `gh project list --owner JosemaPereira`. If no project found, prompt user to create one:
   ```bash
   gh project create --title "TrAIveler MVP Development" --owner JosemaPereira
   ```

3. Resolve the scope:
   - If a sprint was given, select all rows whose `Sprint` matches.
   - If stable IDs were given, select exactly those rows.

4. **Identify tasks to create**:
   - One issue per roadmap task (1:1 mapping)
   - Skip any task that already has an `Issue` URL
   - Note the `Group` value for each task (used for Projects grouping)

5. **Verify required labels exist**:
   - Epic label (e.g., `epic:architecture-foundation`)
   - `spec:NNN`, `sprint:N`, `priority:P1/P2/P3`
   - `type:backend/frontend/infra/e2e`
   Create missing labels if needed.

6. Verify `gh` is authenticated (`gh auth status`). If not, ask user to run `gh auth login`.

7. **PREVIEW** (mandatory):
   - List all issues to be created
   - Show: Stable ID + Title + Labels + Group
   - Total count
   - Get explicit confirmation. Do NOT proceed without it.

8. **Create issues** (one per task):
   ```bash
   gh issue create \
     --repo JosemaPereira/capstone-project-ai-bootcamp \
     --title "<stable-id> — <title>" \
     --body "<issue-template>" \
     --label "epic:<name>,spec:<n>,sprint:<n>,priority:<P>,type:<type>"
   ```

9. **Add all issues to GitHub Project**:
   ```bash
   for issue_num in {start..end}; do
     gh project item-add <PROJECT-NUMBER> \
       --owner JosemaPereira \
       --url "https://github.com/JosemaPereira/capstone-project-ai-bootcamp/issues/$issue_num"
   done
   ```

10. **Update roadmap**: Write issue URLs back to `docs/roadmap.md` Issue column for each task.

11. **SET UP DEPENDENCY TRACKING** (mandatory):
    - Analyze the roadmap's "Depends on" column for all created issues
    - Create a shell script to establish relationships via cross-reference comments
    - For each dependency: add "Blocked by #N" comment on dependent issue, add "Blocks #A, #B, #C" comment on blocker issue
    - Execute the script to create visible relationships in GitHub UI
    - Report: "Dependency tracking complete: X blocking relationships established"
    
    Example script structure:
    ```bash
    # For each dependency in roadmap
    gh issue comment <blocked-issue> --body "**Dependency Tracking**: Blocked by #<blocker-issue>"
    ```

12. **Report**:
    - Created issues: `<stable-id> → <url>`
    - Skipped (already tracked): `<stable-id> → existing <url>`
    - Failures: `<stable-id> → ERROR: <message>`
    - Projects: Confirm all added successfully
    - Dependencies: `X blocking relationships established`

13. **Suggest next steps**:
    - In GitHub Projects, add "Group" custom field
    - Set Group values for grouped tasks (e.g., G-ARCH-SETUP-DIRS)
    - Create project views (Sprint Board, Epic Board, Group Board, Dependencies View)

Do NOT commit. Leave `docs/roadmap.md` staged and suggest `/commit-and-push` if approved.

---

## Notes

- **No parent issues**: Each roadmap task gets exactly one issue
- **Use Projects for grouping**: Set "Group" custom field in Projects UI
- **Epic labels**: All issues in sprint share epic label (e.g., `epic:architecture-foundation`)
- **Simple, scalable**: 23 tasks = 23 issues (not 29)
- **Dependency tracking is MANDATORY**: After creating issues, customize and run `.github/scripts/setup-issue-relationships.sh` to establish visible "Blocked by"/"Blocks" relationships in GitHub UI based on roadmap dependencies
