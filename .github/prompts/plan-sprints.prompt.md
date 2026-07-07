---
description: "Plan and prioritize the roadmap into sprints with mandatory consolidation: assign sprints, consolidate atomic tasks, set priorities, define sprint goals, surface risks and the critical path"
agent: "product-manager"
tools: ['read', 'search', 'edit', 'todo']
---

Act as a Product Manager and turn `docs/roadmap.md` into a prioritized, sprint-based plan with consolidated work items. Switch to the `product-manager` agent if not already active.

Planning input (optional): ${input:planning:Optional. Provide constraints such as number of sprints, sprint length, team size/velocity, deadlines, or priority focus. Leave blank to be asked.}

Instructions:
1. Load `docs/roadmap.md`. If it does not exist, tell the user to run `/build-roadmap` first and stop.
2. Confirm or ask for: number of sprints or sprint length, team size / velocity, fixed deadlines, and any priority focus. If unknown, propose conservative defaults and label them assumptions.
3. **Check for existing sprints**: Identify sprints that already have Issue URLs in roadmap — SKIP those as refinement evidence. Apply consolidation and planning ONLY to future sprints.
4. **MANDATORY CONSOLIDATION ANALYSIS** (for each NEW sprint without issues):
   - Group tasks by area/module (middleware, primitives, config, Terraform modules, etc.)
   - Apply consolidation rules: same spec + same tech + same context + size 2-4 tasks
   - Propose groups with reasoning and before/after stats (e.g., "Sprint 2: 40 tasks → 19 issues (-52%)")
   - Show which tasks will be grouped and which stay standalone
   - Get explicit approval before assigning Group values
5. Prioritize the tasks using the most fitting framework (MoSCoW / value-effort / RICE / WSJF). Show the ranking and the reasoning briefly.
6. Assign a `Sprint` value to each task, strictly respecting dependencies (never before a dependency) and the foundation-before-feature order, without overfilling sprint capacity. Keep members of the same `Group` in the same sprint.
7. Write `Group`, `Sprint`, and `Priority` back into `docs/roadmap.md`, and add or update a `## Sprint Plan` section containing, per sprint: a one-line Sprint Goal, the work items (groups + standalone tasks with counts), total size, consolidation stats, and key risks.
8. Do NOT create issues here. That is done later, scoped, via `/create-sprint-issues`.
9. Report: the sprint plan, consolidation results per sprint (tasks → issues), proposed groupings and their rationale, what was deferred and why, the critical path, and open risks.

**Consolidation Rules (MANDATORY)**:
- ✅ DO consolidate: Same spec + same area/module + same tech stack + shared context + size 2-4 tasks
- ❌ DON'T consolidate: Different tech stacks, critical blockers, different dependency chains, cross-spec
- **Target**: 10-25% ticket reduction per sprint via consolidation

**Note on Issue Creation**: When creating GitHub issues from sprint tasks (not in this workflow), **MUST** read `.github/ISSUE-CREATION-GUIDELINES.md` first. It defines consolidation patterns, epic vs issue classification, relationship management, label strategy, and issue body templates.

Do NOT commit. Leave `docs/roadmap.md` staged for review and suggest `/commit-and-push` if approved.
