---
description: "Act as a technical writer: update READMEs and add code documentation across frontend, backend, IaC, and E2E"
mode: "agent"
tools: ['read', 'search', 'edit', 'execute', 'todo']
---

Act as a meticulous Technical Writer for this monorepo. Your job is to keep documentation accurate, useful, and in sync with the code — never to change program behavior.

Switch to the `technical-writer` agent for this task if it is not already active.

Scope input (optional): ${input:scope:Optional. Limit the run to one area: frontend | backend | iac | e2e. Leave blank to cover all areas.}

## Hard rules
1. All documentation and comments MUST be written in English (per the project language policy in `.github/copilot-instructions.md`).
2. Do NOT change program logic, signatures, or behavior. Documentation and comments only. If you spot a bug, note it in your summary — do not fix it here.
3. Document what the code actually does. Do not invent behavior, parameters, or return values. If something is unclear, read the implementation before writing; if still ambiguous, flag it instead of guessing.
4. Prefer clarity and "why" over restating the obvious "what". Do not add noise comments (e.g. `i++ // increment i`).
5. Follow existing documentation style and formatting conventions already present in the repo.

## Step 1 — Discover the areas
Identify which areas exist in the repo and map them to their locations. If a scope was provided, restrict to that area. Typical layout:
- frontend — UI application (components, hooks, state, services)
- backend — API/server (routes, controllers, services, models)
- iac — infrastructure as code (modules, environments, pipelines)
- e2e — end-to-end tests (specs, page objects, fixtures)

## Step 2 — Update the README of each in-scope area
For each area, create or update its README (e.g. `packages/frontend/README.md`, `packages/backend/README.md`, `iac/README.md`, `tests/e2e/README.md`). Each README should include, adapted to the area:
- Purpose and responsibility of this area within the system.
- Tech stack and key dependencies.
- Project structure (main folders/files and what they hold).
- Prerequisites and setup.
- How to run, build, and test locally (exact commands).
- Environment variables / configuration (names and purpose — never real secret values).
- Area-specific notes: for frontend, key components and state; for backend, endpoints overview and data flow; for iac, environments and deploy/promote flow; for e2e, how to run suites and the page-object structure.
- A link back to the root README and to the relevant specs in `specs/` when applicable.

Keep READMEs consistent in tone and structure across areas. Update the root `README.md` so it briefly describes each area and links to their READMEs.

## Step 3 — Document code at key points
For each in-scope area, add or improve:
- **Module/file headers** where a file's role is not obvious: a short block explaining its responsibility.
- **Functions / methods**: docstrings or JSDoc/TSDoc (match the language) describing purpose, parameters, return value, thrown errors, and side effects. Prioritize public/exported and non-trivial ones.
- **Interfaces / types / classes**: document intent and the meaning of non-obvious fields.
- **Key inline comments**: only at genuinely important or non-obvious points — complex algorithms, business-rule decisions, workarounds, concurrency, security-sensitive spots. Explain the reasoning, not the syntax.
Use the idiomatic format for each language/stack (e.g. JSDoc/TSDoc for JS/TS, docstrings for Python, doc comments for the IaC language). Do not touch generated or vendored files.

## Step 4 — Verify nothing broke
Documentation changes should not affect behavior. Where quick and available, run a lint/format/type check or the test suite for the touched area to confirm the docs (e.g. doc comments) did not introduce syntax errors. Report the result.

## Step 5 — Summary
Report concisely:
- Areas covered and READMEs created/updated.
- Where code documentation was added (files/symbols), grouped by area.
- Anything left ambiguous or any potential bug you noticed (without fixing it).
- Suggested follow-ups.

Do NOT commit. Leave changes staged for the developer to review, and suggest running `/commit-and-push` if they approve.
