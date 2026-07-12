---
description: "Create a pull request for the current feature branch"
mode: "agent"
tools: ['read', 'execute', 'todo']
---

Create a GitHub pull request for the current feature branch using the GitHub CLI (`gh`).

Base branch input (optional): ${input:base-branch:Optional. Target base branch for the PR. Leave blank to use main.}

Instructions:
1. Determine the current branch with `git rev-parse --abbrev-ref HEAD`.
2. If the current branch is `main` (or the chosen base branch), stop and ask the user
   to switch to a feature branch first.
3. If a base branch was not provided, default to `main`.
4. Ensure the current branch is pushed to the remote:
   - Run `git push -u origin <current-branch>` if it has no upstream, otherwise `git push`.
5. Verify `gh` is available (`gh --version`). If it is missing or not authenticated,
   stop and tell the user to run `gh auth login`.
6. Analyze the commits and diff of the current branch against the base branch to
   understand the change set.
7. **Read the PR template** from `.github/PULL_REQUEST_TEMPLATE.md` to understand the
   required structure.
8. Generate PR content in English following the template structure:
   - **Title**: Concise conventional format (feat:, fix:, chore:, docs:, etc.)
   - **Description**: Brief summary of what this PR accomplishes
   - **Stable IDs**: Extract from commit messages or branch name if available (e.g., 005-T024)
   - **Spec**: Reference the spec number (e.g., "005 — System Architecture")
   - **Sprint**: Identify sprint number if applicable (e.g., "Sprint 2")
   - **Group**: Group ID if tasks are consolidated (e.g., "G-SPRINT2-BACKEND-MIDDLEWARE") or "Standalone"
   - **Implementation Summary**: List key files changed and what was implemented
   - **Testing**: Describe test coverage, manual testing performed, verification steps
   - **Related Specifications**: Reference relevant specs/ and docs/ files
   - **Dependencies**: Note any PR dependencies (Depends on, Blocks, Related)
   - **Next Steps**: Recommend follow-up work if applicable
   - **Deployment Notes**: Note breaking changes, migrations, env vars
   - **Closes**: Reference GitHub issues this PR closes (if applicable)
   - **Review focus**: Highlight what reviewers should pay special attention to
9. Create the PR using the structured body:
   `gh pr create --base <base-branch> --head <current-branch> --title "<title>" --body "<body>"`
   
   **Note**: The body should be formatted as valid Markdown following the template structure.
   Use `\n` for line breaks in the `--body` argument.
10. Return the resulting PR URL.
11. Do NOT merge the PR. Leave it open for review.

## Template Compliance

The PR body MUST follow the structure defined in `.github/PULL_REQUEST_TEMPLATE.md`:
- All major sections included (Description, Implementation Summary, Testing, etc.)
- Proper Markdown formatting (headers with ##, bullet lists, checkboxes)
- English language throughout
- Stable IDs and Spec references when available from commits or roadmap
- Every issue/PR cross-reference (Dependencies, Next Steps, etc.) written as a bare `#N` **inside a
  Markdown list item**, per Issue/PR Reference Formatting below — never as a paragraph line

## Issue/PR Reference Formatting

GitHub auto-expands a bare `#N` issue/PR reference into a rich card (icon + full title + number)
only when that reference sits inside a Markdown list item (a `-`/`*`/numbered line). The exact same
`#N` syntax left in plain paragraph text renders as a short link with just the number — same repo,
same reference, different rendering, purely based on list-vs-paragraph placement.

So: **every** issue/PR reference in the PR body — `Depends on`, `Blocks`, `Related` in the
Dependencies section, `Next Steps`, anywhere else — MUST be written as a bare `#N` (never a
hand-typed title next to it, never a manually-built Markdown link) placed inside a list item, so
GitHub does the title expansion itself. `.github/PULL_REQUEST_TEMPLATE.md`'s `## Dependencies`
section is already formatted this way (each of Depends on/Blocks/Related as its own bullet) —
preserve that list structure, don't collapse it back into bold-label paragraph lines.

**Exception — the `Closes #<issue>` line.** Leave it exactly as `Closes #90` on its own line,
matching the template. Whether or not it also happens to render as an expanded card is irrelevant —
what matters is that the literal `Closes #<number>` text stays intact and unambiguous, since that's
what GitHub's issue-closing-keyword parser matches on merge.
