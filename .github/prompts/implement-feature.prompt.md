---
description: "Implement a feature using strict test-first TDD"
mode: "agent"
agent: "tdd-developer"
tools: ['search', 'read', 'edit', 'execute', 'web', 'todo']
---

Implement a feature using a strict TDD execution flow. Switches to the tdd-developer agent.

Feature or task input (required): ${input:feature:Required. Describe the feature or task to implement (or point to a spec/tasks file, e.g. specs/001-*/tasks.md).}

## MANDATORY Prerequisites (Memory Loading)

Before implementing, load project memory to ensure continuity. `session-notes.md` and
`patterns-discovered.md` are large, append-only logs (tens of KB) — read them targeted, not in full,
to avoid burning most of the context window before any code is touched:

1. **Load Working Notes in full**: `read_file(".github/memory/scratch/working-notes.md")`
   - Small by design. Check for in-progress work and avoid conflicts with the current session.
2. **Targeted lookup in Patterns Discovered**: `grep` for keywords from the feature/task at hand
   (component, domain, layer, tech e.g. "Zustand", "JWT", "handler") against
   `.github/memory/patterns-discovered.md` and read only the matching entries.
   - Only fall back to reading the full file if the grep yields no relevant hits and the task
     touches a foundational/cross-cutting area (auth, data model, API conventions).
3. **Targeted lookup in Session Notes**: `grep` the same keywords against
   `.github/memory/session-notes.md`; if nothing matches, read only the most recent
   (non-"Compacted") sprint section rather than the whole history.
4. **Confirm Memory Load**: State which entries/sections were used (or that none matched) before proceeding.

**Why this is mandatory:**
- Prevents duplicate implementations
- Ensures consistency with established patterns
- Avoids conflicts with in-progress work
- Maintains code quality standards

**Why targeted, not full-file, reads:** these two files are shared, append-only project memory that
grows every sprint (currently ~97KB / ~77KB). Reading both in full on every subagent invocation was
a major contributor to running this agent at >150k context.

## Implementation Instructions

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
