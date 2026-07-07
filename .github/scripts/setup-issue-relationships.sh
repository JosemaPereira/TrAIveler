#!/bin/bash
# Setup GitHub issue relationships for sprint tasks
# Creates cross-references and adds tracking comments for dependencies
# 
# Usage: ./setup-issue-relationships.sh
#        This is a TEMPLATE script - customize for each sprint

REPO_OWNER="JosemaPereira"
REPO_NAME="capstone-project-ai-bootcamp"

echo "============================================"
echo "GitHub Issue Dependency Tracking Script"
echo "============================================"
echo ""

# Function to add blocking relationship comment
# This creates visible cross-references in GitHub's UI
add_relationship() {
  local issue=$1
  local blocker=$2
  local task_id=$3
  
  # Check if relationship comment already exists
  existing=$(gh issue view $issue --repo "$REPO_OWNER/$REPO_NAME" --json comments --jq ".comments[] | select(.body | contains(\"Blocked by #$blocker\"))" 2>/dev/null)
  
  if [ -n "$existing" ]; then
    echo "  ⏭️  Relationship already tracked: #$issue blocked by #$blocker"
    return 0
  fi
  
  # Add comment to create cross-reference
  gh issue comment $issue \
    --repo "$REPO_OWNER/$REPO_NAME" \
    --body "**Dependency Tracking**: This issue is blocked by #$blocker

This task ($task_id) requires completion of the blocking issue before work can begin." \
    && echo "  ✅ Added: #$issue blocked by #$blocker" \
    || echo "  ❌ Failed: #$issue blocked by #$blocker"
}

# Function to add "blocks" relationship comment on the blocking issue
add_blocks_comment() {
  local blocker=$1
  shift
  local blocked_issues="$@"
  
  # Check if blocks comment already exists
  existing=$(gh issue view $blocker --repo "$REPO_OWNER/$REPO_NAME" --json comments --jq ".comments[] | select(.body | contains(\"This issue blocks\"))" 2>/dev/null)
  
  if [ -n "$existing" ]; then
    echo "  ⏭️  Blocks tracking already exists on #$blocker"
    return 0
  fi
  
  # Format blocked issues list
  blocked_list=$(echo "$blocked_issues" | tr ' ' '\n' | sed 's/^/- #/' | tr '\n' ' ' | sed 's/ $/\n/')
  
  # Add comment listing what this issue blocks
  gh issue comment $blocker \
    --repo "$REPO_OWNER/$REPO_NAME" \
    --body "**Blocks the following issues**:
$blocked_list

These issues depend on completion of this task." \
    && echo "  ✅ Added blocks tracking to #$blocker" \
    || echo "  ❌ Failed to add blocks tracking to #$blocker"
}

echo "⚠️  TEMPLATE SCRIPT - Customize for Your Sprint"
echo ""
echo "To set up dependencies for a sprint:"
echo "  1. Read docs/roadmap.md to identify dependencies"
echo "  2. For each 'Depends on' relationship, note:"
echo "     - Issue number of dependent task"
echo "     - Issue number of blocking task"
echo "     - Task ID (e.g., 005-T020)"
echo "  3. Add calls below using this pattern:"
echo ""
echo "     # Foundation issue that blocks others"
echo "     add_blocks_comment <blocker-issue> \"<blocked-1> <blocked-2> <blocked-3>\""
echo "     "
echo "     # Dependent issues"
echo "     add_relationship <blocked-issue> <blocker-issue> \"<task-id>\""
echo ""
echo "============================================"
echo ""

# EXAMPLE (Sprint 1 pattern - adapt for your sprint):
#
# echo "Setting up foundation blocker..."
# add_blocks_comment 50 "51 52 53 54"  # Issue #50 blocks 4 issues
# add_relationship 51 50 "002-T010"    # Issue #51 blocked by #50
# add_relationship 52 50 "002-T011"    # Issue #52 blocked by #50
# add_relationship 53 50 "002-T012"    # Issue #53 blocked by #50
# add_relationship 54 50 "002-T013"    # Issue #54 blocked by #50
#
# echo ""
# echo "Setting up another blocker..."
# add_blocks_comment 55 "56 57"        # Issue #55 blocks 2 issues
# add_relationship 56 55 "002-T020"    # Issue #56 blocked by #55
# add_relationship 57 55 "002-T021"    # Issue #57 blocked by #55

# ADD YOUR SPRINT'S DEPENDENCY SETUP HERE:
# (Uncomment and customize the example above)

echo ""
echo "============================================"
echo "⚠️  No dependencies configured yet"
echo "============================================"
echo ""
echo "Customize this script by adding dependency"
echo "relationships based on your sprint's roadmap."
echo ""
echo "For Sprint 1 example, run:"
echo "  git show HEAD~1:.github/scripts/setup-issue-relationships.sh"
echo ""
