---
description: "Create UI tests for required critical user journeys"
mode: "agent"
agent: "test-engineer"
tools: ['search', 'read', 'edit', 'execute', 'todo']
---

Create or update UI tests for critical user journeys with strong stability and maintainability practices. Switches to the test-engineer agent.

Journeys input (optional): ${input:journeys:Optional. Comma-separated journeys. Leave blank to use defaults.}

Instructions:
1. If journeys are not provided, use sensible defaults for the app's primary flows
   (for example: create, edit, toggle/update, delete, and core error-state handling).
2. HARD LIMIT: Create a maximum of 5 UI test cases in this run (target 3-5 total).
3. Include at least 1 error-path test within the 3-5 total.
4. If candidate scenarios exceed 5, select the highest-risk 5 and list the deferred
   scenarios instead of creating more tests.
5. Generate or update tests using the project's UI test framework (see
   `docs/testing-guidelines.md`).
6. Prefer stable, accessibility-first selectors and state-based waits.
7. Apply Page Object Model best practices:
   - Put reusable interactions/selectors in page objects/helpers.
   - Keep test files focused on scenario intent and assertions.
   - Avoid duplicate selectors and repeated interaction flows.
8. Before finishing, count authored test cases and reduce to <= 5 if over the limit.
9. All test code and names MUST be in English.
10. Report files changed and scenarios covered.
