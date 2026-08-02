# Development Memory System

## Purpose
The memory system captures patterns, decisions, and lessons discovered during
implementation, testing, and debugging. It gives you and the AI agent a shared,
evolving record of how this codebase behaves in practice.

Use this memory system to:
- Preserve proven approaches so they can be reused quickly.
- Document trade-offs and decisions to avoid repeating analysis.
- Capture troubleshooting lessons from failures and fixes.
- Improve consistency across iterative, feedback-driven development cycles.

## Single Source of Truth for Any AI Tool

This system is **tool-agnostic**: any AI tool used in this repository reads and writes the exact
same files under `.github/memory/`. There is no per-tool copy of memory — Claude Code is currently
the only AI tool used on this project.

This file is the **canonical protocol**. `CLAUDE.md` summarizes Claude Code's obligations under it,
but the rules themselves live here only — so update the protocol here first, and it never drifts
out of sync.

## Two Types of Memory

1. **Persistent Memory**

   - Location: `CLAUDE.md` (canonical — Claude Code is the sole AI tool for this project).
   - Role: Stable, foundational guidance (principles, workflows, responsibilities).
   - Changes: Infrequent and deliberate, made in `CLAUDE.md` first.

2. **Working Memory**

   - Location: `.github/memory/`
   - Role: Day-to-day discoveries and emerging implementation knowledge.
   - Changes: Frequent, session-driven updates.

## Directory Structure
- `session-notes.md`: Historical summaries of completed sessions (committed).
- `patterns-discovered.md`: Accumulated implementation and debugging patterns (committed).
- `scratch/working-notes.md`: Active session notes and in-progress thinking (NOT committed).
- `scratch/.gitignore`: Ignores everything in scratch to keep ephemeral work out of git.

## Avoiding Conflicts & Stale Entries Across Concurrent Sessions

Because multiple Claude Code sessions can run concurrently on different branches (and this
protocol remains tool-agnostic in case another AI tool is added later), follow these rules so
entries never clash or silently overwrite each other:

- **Append-only.** New entries go at the end of the relevant section. Never edit, reorder, or
  delete another session's entry — not even to "clean it up."
- **Tag every new entry with the tool that wrote it.** Add a `**Tool**: Claude Code` line next to
  the date in every new `### Session: ...` and `### Pattern Name` entry (see templates in each
  file). This makes provenance and freshness obvious at a glance. Historical entries tagged with a
  different tool (e.g. GitHub Copilot, used on this project before the migration tracked in issue
  #212) are kept as-is — append-only applies regardless of which tool wrote the entry.
- **Sync before writing.** Pull the latest committed version of the memory files before appending,
  so a new entry lands after the true last entry rather than a stale local copy.
- **On merge conflicts, keep both sides.** If a git merge conflicts inside these files, never
  resolve by discarding one tool's addition — concatenate both entries and move on. Exact ordering
  is not critical since every entry is dated.
- **Compaction preserves both tools' substance.** When a phase's entries pile up and get folded
  into a `(Compacted)` summary block, whoever compacts must carry forward the substance of every
  tool's entries, not just their own.

## Session Start Protocol

**MANDATORY for any AI tool working in this repository (Claude Code, or others added later).
Every new session must begin by loading memory files in this order:**

1. Read `session-notes.md` — understand what has been built and decided
2. Read `patterns-discovered.md` — review proven implementation patterns
3. Read `scratch/working-notes.md` — check for in-progress work
4. Confirm loading with a status message before proceeding with any task

## Prerequisites for Implementation Workflows

**MANDATORY: Before using `/implement-feature` or any SpecKit implementation workflow:**

1. **Load Session Notes** — `read_file(".github/memory/session-notes.md")` to understand project history
2. **Load Patterns Discovered** — `read_file(".github/memory/patterns-discovered.md")` to apply proven solutions
3. **Load Working Notes** — `read_file(".github/memory/scratch/working-notes.md")` to check for in-progress work
4. **Confirm Memory Load** — Output confirmation message before proceeding

**Why this is critical:**
- Prevents re-discovering known solutions
- Avoids conflicting with in-progress work
- Ensures consistency with project patterns
- Maintains continuity across sessions

**Failure to load memory may result in:**
- Duplicate implementations
- Pattern violations
- Inconsistent code style
- Wasted development time

## When to Use Each File
- **While working**: take notes in `scratch/working-notes.md`.
- **When a reusable approach appears**: add it to `patterns-discovered.md`.
- **At the end of a session**: summarize durable takeaways into `session-notes.md`.

## Language
All memory files are written in English, consistent with the project language policy.
