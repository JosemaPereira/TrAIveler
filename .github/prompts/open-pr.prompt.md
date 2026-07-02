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
7. Generate, in English:
   - A concise PR title in conventional format (feat:, fix:, chore:, docs:, etc.).
   - A short PR body with: a one-paragraph summary, a bullet list of key changes,
     and a "Testing" note describing how it was validated. Reference the related
     spec (e.g. `specs/001-*/spec.md`) if one exists.
8. Create the PR:
   `gh pr create --base <base-branch> --head <current-branch> --title "<title>" --body "<body>"`
9. Return the resulting PR URL.
10. Do NOT merge the PR. Leave it open for review.
