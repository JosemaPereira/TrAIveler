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

## Two Types of Memory

1. **Persistent Memory**
- Location: `.github/copilot-instructions.md`
- Role: Stable, foundational guidance (principles, workflows, responsibilities).
- Changes: Infrequent and deliberate.

2. **Working Memory**
- Location: `.github/memory/`
- Role: Day-to-day discoveries and emerging implementation knowledge.
- Changes: Frequent, session-driven updates.

## Directory Structure
- `session-notes.md`: Historical summaries of completed sessions (committed).
- `patterns-discovered.md`: Accumulated implementation and debugging patterns (committed).
- `scratch/working-notes.md`: Active session notes and in-progress thinking (NOT committed).
- `scratch/.gitignore`: Ignores everything in scratch to keep ephemeral work out of git.

## When to Use Each File
- While working: take notes in `scratch/working-notes.md`.
- When a reusable approach appears: add it to `patterns-discovered.md`.
- At the end of a session: summarize durable takeaways into `session-notes.md`.

## Language
All memory files are written in English, consistent with the project language policy.
