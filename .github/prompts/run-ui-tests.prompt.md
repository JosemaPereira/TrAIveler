---
description: "Run UI tests and summarize failures"
mode: "agent"
agent: "test-engineer"
tools: ['read', 'execute', 'todo']
---

Run the project's UI tests and summarize outcomes with failure classification. Switches to the test-engineer agent.

Instructions:
1. Ensure UI test dependencies are installed (for example, browser binaries for the
   chosen framework). Install them first if missing.
2. Ensure the application (backend and/or frontend) is running before executing UI tests,
   using the project's start command.
3. Run the UI tests using the project's command (see `docs/testing-guidelines.md`).
4. If dependency install fails, stop immediately and report an environment blocker,
   including the failing command and key error lines. Do not run the tests.
5. Summarize pass/fail results clearly.
6. For failures, classify the likely root cause as one of:
   - application code
   - test code
   - environment
7. For each failure, propose the smallest next step to validate and fix.
