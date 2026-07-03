# Project Workflow Guide

> The end-to-end operating manual for this project: how it was set up, which tooling
> was added and what problem each piece solves, and the repeatable flow from a spec to
> a shipped, tracked piece of work. Written to live inside the repo so any contributor
> (or agent) can follow it without external context.

---

## Table of contents

1. [Philosophy in one paragraph](#1-philosophy-in-one-paragraph)
2. [How the project was bootstrapped (manual setup)](#2-how-the-project-was-bootstrapped-manual-setup)
3. [SpecKit: what it is and why we added it](#3-speckit-what-it-is-and-why-we-added-it)
4. [The agentic kit: agents, commands, and the needs they solve](#4-the-agentic-kit-agents-commands-and-the-needs-they-solve)
5. [Foundational phase: specs before features](#5-foundational-phase-specs-before-features)
6. [Propagating context across the whole project](#6-propagating-context-across-the-whole-project)
7. [Consolidating tasks into a single roadmap](#7-consolidating-tasks-into-a-single-roadmap)
8. [Planning sprints and prioritizing](#8-planning-sprints-and-prioritizing)
9. [Turning roadmap work into GitHub issues](#9-turning-roadmap-work-into-github-issues)
10. [Implementing an issue](#10-implementing-an-issue)
11. [Documentation](#11-documentation)
12. [Scaling: new tasks or new foundations later](#12-scaling-new-tasks-or-new-foundations-later)
13. [Duplicate and repeated work](#13-duplicate-and-repeated-work)
14. [Quick command reference](#14-quick-command-reference)
15. [Golden rules](#15-golden-rules)

---

## 1. Philosophy in one paragraph

We practice **spec-driven, context-first development**. Nothing is built from a vague
prompt: intent is captured in specs, durable decisions are promoted to persistent
project context, work is consolidated and prioritized in a single roadmap, and only
then turned into tracked issues and implemented test-first. Every automated step is
**idempotent** (safe to re-run) and **reconciling** (updates in place instead of
duplicating), so the process scales without turning into busywork.

---

## 2. How the project was bootstrapped (manual setup)

The following were done once, by hand, before any agent involvement:

1. **Repository + branch discipline.** Cloned the empty repo; work happens on
   `feature/<name>` branches, never directly on `main`.
2. **Stack decision, made explicitly** and recorded in the README, so agents build on a
   chosen stack instead of guessing.
3. **`.github/copilot-instructions.md`** — the project "system prompt". It holds project
   context, the mandatory language policy, development principles (TDD, small increments),
   the git workflow, references to the `docs/` context files, and the memory-system rules.
   Everything the agent should always know lives here or is linked from here.
4. **Context docs in `docs/`** — `functional-requirements.md`, `coding-guidelines.md`,
   `testing-guidelines.md`, `ui-guidelines.md`. These are the standards the agent inherits
   on every task (this is "context engineering": document once, reuse everywhere).

**Language policy (mandatory):** conversation with the agent may be English or Spanish,
but ALL committed artifacts — code, comments, docs, specs, commit messages, branch names —
must be in English, with no language mixing inside a file.

---

## 3. SpecKit: what it is and why we added it

**Problem it solves:** ad-hoc prompting produces code with no traceable link back to
intent, and no consistent way to go from an idea to acceptance-tested work.

**SpecKit** formalizes that path into a reproducible pipeline of slash commands, each
producing versioned artifacts inside `specs/NNN-*/`:

```
/speckit.constitution   → .specify/memory/constitution.md  (governance, versioned principles)
/speckit.specify        → spec.md  (+ creates a NNN-* branch)
/speckit.clarify        → resolves ambiguities BEFORE planning (highest-ROI step)
/speckit.plan           → plan.md, data-model.md, research.md, contracts/
/speckit.tasks          → tasks.md (atomic tasks, phases, [P] parallel flags, dependencies)
/speckit.implement      → executes tasks.md (code + tests + docs)
```

Install (once per project):

```bash
uv tool install specify-cli --from git+https://github.com/github/spec-kit.git
specify init --here --force --integration copilot
```

> Note: the flag is `--integration` (not the deprecated `--ai`). SpecKit inserts its
> context into `.github/copilot-instructions.md` between `SPECKIT` markers, coexisting
> with the hand-written content.

**SpecKit's limit:** it operates *inside a single spec*. It does not consolidate,
prioritize, or organize work *across* specs, and it does not read the other specs when
planning a feature. Those gaps are exactly what the agentic kit fills.

---

## 4. The agentic kit: agents, commands, and the needs they solve

Custom agents (`.github/agents/*.agent.md`) and slash commands
(`.github/prompts/*.prompt.md`) were added on top of SpecKit. Each agent has an explicit
role, preferred model, tool set, and **scope boundaries** (what it must NOT do), which is
what keeps agents from making unrequested changes.

### Agents

| Agent | Role | Key boundary |
|---|---|---|
| `tdd-developer` | Implements features test-first (Red-Green-Refactor) | Won't touch lint/unrelated code while fixing tests |
| `code-reviewer` | Lint, quality, idiomatic patterns | Behavior-preserving; adds tests first if behavior may change |
| `test-engineer` | Integration & UI tests, failure triage | Owns tests only; limits UI runs to a few high-value journeys |
| `technical-writer` | READMEs + code documentation | Docs only — never changes program behavior |
| `product-manager` | Sprint planning, prioritization, backlog hygiene, scoped issue creation | Confirms before any external (issue) write; never invents tasks |

### Commands and the need each one solves

| Command | Need it solves |
|---|---|
| `/promote-foundations` | SpecKit doesn't share one spec's decisions with the rest — this promotes durable decisions to the constitution + `docs/` so every future plan/task inherits them |
| `/build-roadmap` | SpecKit produces isolated `tasks.md` per spec — this consolidates them into one reconciled `docs/roadmap.md` without duplicating or losing human curation |
| `/plan-sprints` | The roadmap needs prioritization and sprinting — the PM ranks work, groups tasks, and assigns sprints with goals and risks |
| `/create-sprint-issues` | Creating every issue at once floods the backlog — this creates issues for ONLY a chosen sprint or task, honoring task grouping |
| `/sync-issues` | When you do want to push all pending work — creates issues for every untracked roadmap row at once |
| `/implement-feature` | Turns an issue/task into working, tested code via strict TDD |
| `/create-ui-tests` | Adds a small set of high-value UI journey tests |
| `/run-ui-tests` | Runs UI tests and classifies failures by root cause |
| `/technical-writer` | Keeps READMEs and code docs in sync across areas |
| `/commit-and-push` | SpecKit doesn't commit — this makes a conventional commit and pushes to a feature branch |
| `/open-pr` | Opens a PR for the current branch (via `gh`) |

Two cross-cutting concepts the kit introduced into the roadmap:

- **Stable IDs** (`001-T003`) — anchor every task to its source and its future issue, so
  reconciliation and issue creation never duplicate.
- **Task grouping** (`Group` column) — several small, related tasks can share a `Group`
  value and become ONE issue with a checklist, keeping the backlog lean. Grouping is a
  human/PM decision; automation preserves it but never groups on its own.

---

## 5. Foundational phase: specs before features

Before any user-facing feature, the project is grounded with foundation specs, in this
dependency order (each informs the next):

1. Product vision & scope
2. Non-functional requirements (NFRs)
3. Cloud & environments / IaC strategy
4. Security & auth foundations
5. Architecture, tech alignment & foundations
6. Domain / data model foundations
7. API design standards *(optional, if API-heavy)*

**Flow for each foundation spec** (batch the whole foundation phase before moving on,
since these are tightly coupled and clarifying one can change another):

```
/speckit.specify → /speckit.clarify → /speckit.plan → /speckit.tasks
```

Rule of thumb for what goes where:
- "How do we build, and by what rules?" → **constitution**.
- "What is the system, and under what constraints?" → **foundation specs**.
- "What does the user do?" → **feature specs** (comes later).

---

## 6. Propagating context across the whole project

**The critical step.** SpecKit reads the constitution and the current spec, but NOT the
other specs. If architecture/NFR/security decisions live only inside `specs/005-*/spec.md`,
future feature plans are blind to them.

**Solution — run `/promote-foundations`** once the foundation specs are stable (and again
whenever a foundation spec is refined). It distills durable decisions into the two places
the whole project reads:

- `.specify/memory/constitution.md` — non-negotiable principles/standards.
- `docs/*.md` (architecture, data-model, nfrs, security, cloud) — reference material,
  wired into `.github/copilot-instructions.md`.

After this, every `/speckit.plan`, `/speckit.tasks`, and implementation inherits the
foundational context automatically — no re-reading specs by hand. This is the mechanism
that guarantees implementations don't lose whole-project context.

> Tip: run it early with the foundation specs you already have (don't wait for all seven).
> It surfaces any decisions the specs left ambiguous while they're still cheap to fix.

---

## 7. Consolidating tasks into a single roadmap

Each spec's `tasks.md` is isolated. To get one prioritizable view:

```
/build-roadmap
```

This reconciles all `specs/*/tasks.md` into `docs/roadmap.md`:
- **ADD** new tasks, **UPDATE** changed ones (title/deps), leave **UNCHANGED** ones alone,
  and **REMOVE** tasks refined away (archiving any that already had an issue/progress).
- **Preserves human-owned fields** (`Group`, `Sprint`, `Priority`, `Status`, `Issue`,
  `Notes`) across runs — re-running never wipes your planning.

Because it's reconciling, you run it after every batch of `/speckit.tasks` — it consolidates
incrementally, never from scratch.

---

## 8. Planning sprints and prioritizing

```
/plan-sprints
```

The `product-manager` agent:
- Prioritizes with the fitting framework (MoSCoW / value-effort / RICE / WSJF) and explains
  the trade-offs.
- Right-sizes the backlog: proposes grouping small related tasks into single work items.
- Assigns each task a `Sprint`, respecting dependencies and capacity, and writes a one-line
  Sprint Goal per sprint plus risks and the critical path.
- Writes `Group`, `Sprint`, and `Priority` back into `docs/roadmap.md` (no issues created here).

---

## 9. Turning roadmap work into GitHub issues

Two options; both are idempotent, group-aware, require `gh auth login` once, and preview +
ask for confirmation before creating anything (external actions have no undo):

- **Scoped (recommended for a lean backlog):**
  ```
  /create-sprint-issues     # e.g. "Sprint 1", or a specific stable ID
  ```
  Creates issues only for the chosen sprint or task. Grouped tasks become one issue with a
  checklist.

- **All at once:**
  ```
  /sync-issues              # creates issues for every untracked roadmap row
  ```

Both write the created issue URL back into the `Issue` column of the roadmap, so a re-run
never recreates them.

---

## 10. Implementing an issue

```
/implement-feature        # give it the issue number/URL or paste its content
```

Auto-switches to `tdd-developer`, which reads the issue, finds its `Stable-ID`, locates the
task in `specs/*/tasks.md` + the roadmap, and implements it test-first. If the issue is a
group, it works through the checklist. Then validate and ship:

```
/create-ui-tests → /run-ui-tests → (manual validation) → /commit-and-push → /open-pr
```

Skip the UI steps for non-UI work. Always validate the generated tests yourself — an agent
finishing a task is not proof it works.

---

## 11. Documentation

```
/technical-writer         # optional scope: frontend | backend | iac | e2e
```

Updates area READMEs and adds code documentation (docstrings/JSDoc/TSDoc, interface/class
docs, key inline comments). Documentation only — it never changes behavior. Run it after a
feature is implemented and validated, so it documents real code rather than intentions.

---

## 12. Scaling: new tasks or new foundations later

**A new feature or capability during development** — treat it like any spec:

```
/speckit.specify → /speckit.clarify → /speckit.plan → /speckit.tasks
/build-roadmap                 # reconciles: ADDS the new tasks, keeps existing planning
/plan-sprints                  # optional: re-prioritize / re-sprint
/create-sprint-issues          # create the new issues (scoped)
```

**A refined or brand-new foundation spec** — same as above, plus:

```
/promote-foundations           # re-distill durable decisions into constitution + docs
```

so the new/changed foundational decisions propagate to everything downstream.

Because `/build-roadmap` is reconciling and all issue commands are idempotent, scaling never
means redoing prior work: new specs add, refined specs update, obsolete tasks archive, and
your priorities/sprints/issue links are preserved.

---

## 13. Duplicate and repeated work

Three layers keep the backlog clean:

- **Stable IDs** prevent technical duplicates — issue commands never create two issues for
  the same task ID.
- **Task grouping** absorbs several similar tasks within a spec into one work item.
- **Semantic duplicates across specs** (e.g. "configure logging" appearing in two specs) are
  caught by the `product-manager` during `/build-roadmap` + `/plan-sprints`: it flags
  equivalent tasks and proposes consolidating them (grouping, or marking one as dependent).
  This needs human judgment, so the PM proposes and you approve.

---

## 14. Quick command reference

**Planning & context**
```
/speckit.constitution        governance / versioned principles
/speckit.specify             capability spec (+ branch)
/speckit.clarify             resolve ambiguities (do not skip)
/speckit.plan                technical plan
/speckit.tasks               atomic tasks
/promote-foundations         durable decisions → constitution + docs
/build-roadmap               consolidate specs' tasks → docs/roadmap.md
/plan-sprints                prioritize, group, assign sprints
```

**Issues & delivery**
```
/create-sprint-issues        issues for one sprint/task (scoped, group-aware)
/sync-issues                 issues for all untracked rows
/implement-feature           implement an issue (TDD)
/create-ui-tests             add high-value UI tests
/run-ui-tests                run UI tests + triage
/technical-writer            update READMEs + code docs
/commit-and-push             conventional commit + push
/open-pr                     open a PR (gh)
```

**End-to-end happy path**
```
foundations:  (per spec) specify → clarify → plan → tasks
              then:  /promote-foundations → /build-roadmap → /plan-sprints
delivery:     /create-sprint-issues → /implement-feature → /create-ui-tests
              → /run-ui-tests → /technical-writer → /commit-and-push → /open-pr
scaling:      new spec → …tasks → /build-roadmap (+ /promote-foundations if foundational)
```

---

## 15. Golden rules

- **Don't skip `/speckit.clarify`** — it prevents the most rework.
- **Run `/promote-foundations`** after foundation specs stabilize (and after any refinement),
  or downstream work loses whole-project context.
- **`/build-roadmap` is reconciling** — run it freely; it preserves your curation.
- **Group small related tasks** to keep the backlog lean; grouping is your decision.
- **Confirm before external writes** — issue creation has no undo.
- **Validate AI-generated tests** — try to break the function; a passing suite isn't proof.
- **English-only artifacts**, regardless of conversation language.
- **Never commit to `main`** — feature branches and PRs only.
- **New chat per unrelated task / each `/speckit.*`** — avoid dragging stale context.