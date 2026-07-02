---
description: "Implement a feature using strict test-first TDD"
mode: "agent"
agent: "tdd-developer"
tools: ['search', 'read', 'edit', 'execute', 'web', 'todo']
---

Implement a feature using a strict TDD execution flow. Switches to the tdd-developer agent.

Feature or task input (required): ${input:feature:Required. Describe the feature or task to implement (or point to a spec/tasks file, e.g. specs/001-*/tasks.md).}

Instructions:
1. Read the relevant context in `.github/copilot-instructions.md` and any referenced
   spec/plan/tasks files under `specs/`.
2. Break the work into small, testable increments.
3. For each increment, follow Red-Green-Refactor:
   - Write the failing test first (RED) and confirm it fails for the right reason.
   - Implement the minimal code to pass (GREEN).
   - Refactor while keeping tests green (REFACTOR).
4. Scope boundary: do NOT create or run end-to-end UI tests in this prompt.
   Use `/create-ui-tests` and `/run-ui-tests` for that.
5. Do NOT commit or push changes. Use `/commit-and-push` for that.
6. All code, comments, and test names MUST be in English.
7. Stop after completing the increments and report:
   - What was implemented and which tests were added.
   - Suggested next command(s):
     - If UI workflow is required: `/create-ui-tests` -> `/run-ui-tests` -> `/commit-and-push`
     - Otherwise: `/commit-and-push`
