# Issue Relationship Management Scripts

This directory contains scripts for establishing dependency tracking between GitHub issues.

## Purpose

GitHub doesn't automatically track "Blocked by" / "Blocks" relationships from issue bodies. These scripts create visible cross-references through comments, making dependencies trackable in:
- GitHub issue timelines
- Linked issues sections  
- GitHub Projects dependency views

## Files

### `setup-issue-relationships.sh`
**TEMPLATE SCRIPT** — Must be customized for each sprint.
This is the working script you'll edit and run for every sprint to establish dependency tracking.

**Status**: Template (ready to customize)
**Last used**: Sprint 1 (see git history for reference implementation)

## Workflow for New Sprints (MANDATORY)

### 1. After Creating Sprint Issues

Once issues are created via `/create-sprint-issues`, dependency tracking is **MANDATORY**.

### 2. Customize the Script

Edit `.github/scripts/setup-issue-relationships.sh`:

1. **Analyze `docs/roadmap.md`** for your sprint:
   - Identify all tasks with non-empty "Depends on" column
   - Map Stable ID → Issue number (from Issue column)
   - Map Dependency Stable ID → Blocking issue number

2. **Add relationship calls**:
   ```bash
   # Foundation issue that blocks others
   add_blocks_comment <blocker-issue> "<blocked-1> <blocked-2> ..."
   
   # Each dependent issue
   add_relationship <blocked-issue> <blocker-issue> "<task-id>"
   ```

3. **Example** (Sprint 1 pattern):
   ```bash
   echo "Setting up backend foundation blocker..."
   add_blocks_comment 13 "14 15 16 25 27 30 31 32 33"
   add_relationship 14 13 "005-T005"
   add_relationship 15 13 "005-T013"
   # ... etc
   ```

3. **Add dependency mappings** (follow the pattern):
   ```bash
   # Foundation blocker with multiple dependents
   echo "Setting up #42 (006-T001) as blocker..."
   add_blocks_comment 42 "43 44 45"
   add_relationship 43 42 "006-T002"
   add_relationship 44 42 "006-T003"
   add_relationship 45 42 "006-T004"
   ```

### 3. Run the Script

```bash
.github/scripts/setup-sprint-N-relationships.sh
```

The script is **idempotent** — safe to re-run if relationships need updating.

### 4. Verify Results

Check a few issues in GitHub UI:
- Foundation issues should have "Blocks" comments listing dependents
- Dependent issues should have "Blocked by" comments
- Cross-references should appear in timelines

### 5. Document Dependencies

Create a dependency map (optional but recommended):

```bash
cp .github/SPRINT-1-DEPENDENCIES.md .github/SPRINT-N-DEPENDENCIES.md
```

Update with the new sprint's dependency graph.

## Example Mapping Process

Given roadmap entry:
```
| 006-T005 | Initialize handler | | 2 | P1 | ... | 006-T001 | no | https://.../issues/46 |
```

And you know:
- `006-T005` is issue #46
- `006-T001` is issue #42

Add to script:
```bash
add_relationship 46 42 "006-T005"  # Handler depends on directory
```

## Helper Functions

### `add_relationship(issue, blocker, task_id)`
Adds "Blocked by" comment to a dependent issue.
- `issue`: Number of the blocked issue
- `blocker`: Number of the blocking issue
- `task_id`: Stable ID for reference (e.g., "006-T005")

### `add_blocks_comment(blocker, blocked_issues...)`
Adds "Blocks" comment to a blocking issue with list of dependents.
- `blocker`: Number of the blocking issue
- `blocked_issues`: Space-separated list of issue numbers

## Common Patterns

### Single Blocker, Multiple Dependents
```bash
add_blocks_comment 42 "43 44 45 46"
add_relationship 43 42 "006-T002"
add_relationship 44 42 "006-T003"
add_relationship 45 42 "006-T004"
add_relationship 46 42 "006-T005"
```

### Multiple Blockers in Parallel
```bash
# Backend track
echo "Backend dependencies:"
add_blocks_comment 42 "43 44"
add_relationship 43 42 "006-T002"
add_relationship 44 42 "006-T003"

echo ""
echo "Frontend dependencies:"
add_blocks_comment 50 "51 52"
add_relationship 51 50 "007-T002"
add_relationship 52 50 "007-T003"
```

### Chain Dependencies (A → B → C)
```bash
# Issue 43 depends on 42
add_blocks_comment 42 "43"
add_relationship 43 42 "006-T002"

# Issue 44 depends on 43
add_blocks_comment 43 "44"
add_relationship 44 43 "006-T003"
```

## Troubleshooting

### Script fails with authentication error
```bash
gh auth status
gh auth login  # If not authenticated
```

### Relationships not showing in UI
- Check that comments were added (view issue in browser)
- GitHub may take a few seconds to process cross-references
- Refresh the page

### Duplicate comments
Script checks for existing relationships before adding. If duplicates appear:
1. Manually delete duplicate comments in GitHub UI
2. Script will not re-add them

### Need to remove a relationship
Manually delete the relationship comment in GitHub UI. The script won't re-add it.

## Best Practices

1. **Run immediately after issue creation** — Don't wait days
2. **Verify before committing** — Check a few issues in GitHub UI
3. **Document the dependency graph** — Create SPRINT-N-DEPENDENCIES.md
4. **Keep scripts** — Don't delete sprint scripts, they're reference material
5. **Update for critical path changes** — If dependencies change, re-run with updates

## References

- **Sprint 1 example**: `.github/scripts/setup-issue-relationships.sh`
- **Sprint 1 dependency map**: `.github/SPRINT-1-DEPENDENCIES.md`
- **Issue guidelines**: `.github/ISSUE-CREATION-GUIDELINES.md`
- **Sprint planning prompt**: `.github/prompts/plan-sprints.prompt.md`
- **Issue creation prompt**: `.github/prompts/create-sprint-issues.prompt.md`
