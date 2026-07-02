---
description: "Analyze changes, generate commit message, and push to feature branch"
mode: "agent"
tools: ['read', 'execute', 'todo']
---

Analyze current workspace changes, create a conventional commit, and push only to the user-provided feature branch.

Branch name input (required): ${input:branch-name:Required. Enter the target feature branch name (for example: feature/my-change).}

Instructions:
1. If branch name is empty, ask for it and stop.
2. Never commit to `main` or any branch other than the provided branch name.
3. Analyze changes with `git diff`.
4. Generate a descriptive conventional commit message (feat:, fix:, chore:, docs:, etc.) in English, using the Git Workflow rules in `.github/copilot-instructions.md`.
5. If the branch does not exist, create and switch: `git checkout -b <branch-name>`.
6. If the branch exists, switch to it: `git checkout <branch-name>`.
7. Stage all changes: `git add .`.
8. Commit with the generated message.
9. Push to the specified branch: `git push origin <branch-name>`.