---
description: "Distill durable decisions from the foundational specs into persistent project context (constitution + docs), so every future plan/task/feature inherits them automatically"
mode: "agent"
tools: ['read', 'search', 'edit', 'todo']
---

Promote the durable decisions made across the foundational specs into the project's persistent context, so that all subsequent `/speckit.plan`, `/speckit.tasks`, and implementation work inherits them automatically instead of re-reading each spec.

## Why this exists
SpecKit reads the constitution and the CURRENT spec, but does NOT automatically read the other specs. Foundational specs (vision, NFRs, cloud/IaC, security, architecture, domain, API standards) are one-time discovery. Their durable output must live where everything reads it: the constitution and `docs/*.md` referenced by `.github/copilot-instructions.md`.

## Golden rules
1. Output in English (per the project language policy in `.github/copilot-instructions.md`).
2. Do NOT invent decisions. Only promote what the foundational specs actually state. If a spec left something open, flag it — do not fill the gap.
3. Do NOT modify `specs/**` (they remain the detailed source of record). You only write to the constitution and `docs/`.
4. Additive and idempotent: on re-runs, update the promoted sections in place; do not duplicate. Preserve any human edits outside the marked sections.
5. Route each decision to the right home (see mapping). Principles/non-negotiables → constitution. Reference material → docs.

## Step 1 — Identify foundational specs
Scan `specs/*/spec.md` (and their `plan.md`, `data-model.md`, `research.md` where present). Treat the foundation-phase specs as sources: product vision/scope, NFRs, cloud/environments, security/auth, architecture/tech, domain/data model, and API standards (whichever exist). Report which were found and which are missing.

## Step 2 — Extract durable decisions
From each foundational spec, extract the decisions that features will depend on, e.g.:
- Vision/scope → product purpose, target users, MVP boundary, explicit out-of-scope.
- NFRs → measurable targets (performance, availability, accessibility level, observability).
- Cloud/environments → provider, environment topology, CI/CD approach, secrets strategy.
- Security/auth → auth model, roles/permissions, data protection rules, input-validation principles.
- Architecture/tech → chosen stack + rationale, component boundaries, integration rules.
- Domain/data model → core entities, relationships, key invariants and business rules.
- API standards → naming, versioning, error format, pagination conventions.

## Step 3 — Route and write
Promote into the right destination, wrapping generated content between markers
`<!-- PROMOTED:<topic> START -->` and `<!-- PROMOTED:<topic> END -->` so re-runs update in place:

- **Constitution** (`.specify/memory/constitution.md`) — non-negotiable principles and standards: security rules, mandated stack, testing/quality gates, architectural constraints that every feature must obey. Respect the constitution's amendment/versioning convention if present (bump version, note the change).
- **docs/architecture.md** — architecture overview, component boundaries, integration rules, and the key diagrams' source (or links to them).
- **docs/data-model.md** — core entities, relationships, invariants, base business rules.
- **docs/nfrs.md** — the measurable NFR targets and how they're validated.
- **docs/security.md** — auth/authorization model, roles, data protection, validation principles.
- **docs/cloud-and-environments.md** — provider, environments, CI/CD, secrets strategy.
- Reuse existing docs where they already cover a topic (e.g. keep functional-requirements.md, ui-guidelines.md, coding-guidelines.md, testing-guidelines.md as-is; add only what's missing).

## Step 4 — Wire references
Update `.github/copilot-instructions.md` so its "Documentation References" section lists every promoted doc with a one-line description, ensuring the agent reads them before generating code, tests, plans, or tasks. Do not remove existing references.

## Step 5 — Report
Summarize: specs promoted, destinations written/updated (constitution + which docs), any decisions that were left open/ambiguous in the specs (flagged for the user to resolve), and a reminder that new features will now inherit this context. Recommend running this again whenever a foundational spec is refined.

Do NOT commit. Leave changes staged for review and suggest `/commit-and-push` if approved.