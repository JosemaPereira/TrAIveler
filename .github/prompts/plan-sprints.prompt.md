---
description: "Plan and prioritize the roadmap into sprints with a Product Manager: assign sprints, set priorities, define sprint goals, surface risks and the critical path"
mode: "agent"
agent: "product-manager"
tools: ['read', 'search', 'edit', 'todo']
---

Act as a Product Manager and turn `docs/roadmap.md` into a prioritized, sprint-based plan. Switch to the `product-manager` agent if not already active.

Planning input (optional): ${input:planning:Optional. Provide constraints such as number of sprints, sprint length, team size/velocity, deadlines, or priority focus. Leave blank to be asked.}

Instructions:
1. Load `docs/roadmap.md`. If it does not exist, tell the user to run `/build-roadmap` first and stop.
2. Confirm or ask for: number of sprints or sprint length, team size / velocity, fixed deadlines, and any priority focus. If unknown, propose conservative defaults and label them assumptions.
3. Prioritize the tasks using the most fitting framework (MoSCoW / value-effort / RICE / WSJF). Show the ranking and the reasoning briefly.
4. Backlog hygiene: identify small, related tasks (same spec, same area, shared context) that would fragment the backlog as separate tickets. PROPOSE grouping them into single work items, show the reasoning, and on approval set their `Group` value so they will become one issue with a checklist. Do not group across specs or across dependency boundaries that need separate tracking.
5. Assign a `Sprint` value to each task, strictly respecting dependencies (never before a dependency) and the foundation-before-feature order, without overfilling sprint capacity. Keep members of the same `Group` in the same sprint.
6. Write `Group`, `Sprint`, and `Priority` back into `docs/roadmap.md`, and add or update a `## Sprint Plan` section containing, per sprint: a one-line Sprint Goal, the work items (groups + standalone tasks), total size, and key risks.
7. Do NOT create issues here. That is done later, scoped, via `/create-sprint-issues`.
8. Report: the sprint plan, proposed groupings and their rationale, what was deferred and why, the critical path, and open risks.

Do NOT commit. Leave `docs/roadmap.md` staged for review and suggest `/commit-and-push` if approved.
