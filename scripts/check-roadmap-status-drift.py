#!/usr/bin/env python3
"""Detect drift between docs/roadmap.md's Status column and live GitHub issue state.

docs/roadmap.md holds one markdown table per Spec/Phase, each with columns
`| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |`.
`Status` is human-owned (one of Backlog | Ready | In Progress | In Review | Done |
Superseded); `Issue` links to the GitHub issue tracking that row, and multiple rows
commonly share the same issue (a "group" ticket).

Drift = a row whose linked issue's actual GitHub state disagrees with its Status:
  - issue is CLOSED but Status is neither Done nor Superseded, or
  - issue is OPEN but Status IS Done.

Rows with no Issue URL are out of scope entirely (nothing to compare against).
Rows with Status Superseded are also out of scope entirely: superseded work is
intentionally not tracked to an issue's lifecycle (see CLAUDE.md / roadmap Legend).

A group issue can close while one of its own rows is deliberately left
unfinished (e.g. a GitHub-plan API limitation blocks just that row — see
009-T019/002-T050). This repo's existing convention for that case is a Notes
cell starting with "**Blocked**" (see docs/roadmap.md); such rows are also
out of scope for the closed-but-not-Done check.

Exit code is non-zero if any drift is found, so this can be used as a CI gate
(mirrors scripts/check-agent-drift.py's style: a small, dependency-free script
with a pure parsing/comparison core and a single network boundary at the edge).
"""

import argparse
import json
import re
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
DEFAULT_ROADMAP_PATH = REPO_ROOT / "docs" / "roadmap.md"
DEFAULT_REPO = "JosemaPereira/TrAIveler"

# Matches roadmap task IDs like "001-T001" or "009-T026".
ID_PATTERN = re.compile(r"^\d{3}-T\d+$")

# Matches a well-formed GitHub issue URL and captures the issue number.
ISSUE_URL_PATTERN = re.compile(r"^https://github\.com/[^/\s]+/[^/\s]+/issues/(\d+)$")

STATUS_DONE = "Done"
STATUS_SUPERSEDED = "Superseded"
BLOCKED_NOTES_PREFIX = "**Blocked**"

# Column indices after splitting a table row on "|" (index 0 and the last
# index are the empty strings before/after the leading/trailing pipe).
COL_ID = 1
COL_STATUS = 6
COL_ISSUE = 9
COL_NOTES = 10


@dataclass(frozen=True)
class RoadmapRow:
    """One parsed data row from a docs/roadmap.md task table."""

    id: str
    status: str
    issue_url: str
    issue_number: int
    notes: str = ""


@dataclass(frozen=True)
class DriftEntry:
    """A roadmap row whose Status disagrees with its linked issue's live GitHub state."""

    row: RoadmapRow
    github_state: str

    def format(self) -> str:
        return (
            f"{self.row.id}: roadmap Status={self.row.status!r} but issue "
            f"#{self.row.issue_number} is {self.github_state} ({self.row.issue_url})"
        )


def parse_roadmap_rows(text: str) -> list[RoadmapRow]:
    """Parse every task-table data row out of docs/roadmap.md's markdown content.

    Skips non-table lines, header rows, separator rows, rows whose ID column
    doesn't look like a task ID, and rows with no well-formed Issue URL (empty
    Issue cell, or anything that isn't a `.../issues/<number>` GitHub link).
    """
    rows: list[RoadmapRow] = []
    for line in text.splitlines():
        if not line.startswith("|"):
            continue

        cells = [cell.strip() for cell in line.split("|")]
        if len(cells) <= max(COL_ID, COL_STATUS, COL_ISSUE, COL_NOTES):
            continue

        row_id = cells[COL_ID]
        if not ID_PATTERN.match(row_id):
            continue

        issue_match = ISSUE_URL_PATTERN.match(cells[COL_ISSUE])
        if not issue_match:
            continue

        rows.append(
            RoadmapRow(
                id=row_id,
                status=cells[COL_STATUS],
                issue_url=cells[COL_ISSUE],
                issue_number=int(issue_match.group(1)),
                notes=cells[COL_NOTES],
            )
        )
    return rows


def compute_drift(rows: list[RoadmapRow], issue_states: dict[int, str]) -> list[DriftEntry]:
    """Compare each row's Status against its linked issue's real GitHub state.

    `issue_states` maps issue number -> "OPEN" | "CLOSED". A row whose issue
    number is absent from that map (not returned by the `gh` query) is skipped
    with a warning on stderr rather than reported as drift, since that is a
    data problem (stale/typo'd issue number), not a status-vs-reality mismatch.
    """
    drift: list[DriftEntry] = []
    for row in rows:
        if row.status == STATUS_SUPERSEDED:
            continue
        if row.notes.startswith(BLOCKED_NOTES_PREFIX):
            continue

        state = issue_states.get(row.issue_number)
        if state is None:
            print(
                f"warning: {row.id} links issue #{row.issue_number}, "
                "not found via gh issue list",
                file=sys.stderr,
            )
            continue

        if state == "CLOSED" and row.status != STATUS_DONE:
            drift.append(DriftEntry(row=row, github_state=state))
        elif state == "OPEN" and row.status == STATUS_DONE:
            drift.append(DriftEntry(row=row, github_state=state))

    return drift


def fetch_issue_states(repo: str) -> dict[int, str]:
    """Fetch every issue's live state from GitHub in one call (the network boundary).

    Uses `gh issue list --state all` rather than one call per roadmap row —
    there are hundreds of rows but far fewer distinct issues.
    """
    result = subprocess.run(
        [
            "gh",
            "issue",
            "list",
            "--repo",
            repo,
            "--state",
            "all",
            "--json",
            "number,state",
            "--limit",
            "1000",
        ],
        check=True,
        capture_output=True,
        text=True,
    )
    issues = json.loads(result.stdout)
    return {issue["number"]: issue["state"] for issue in issues}


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument(
        "--repo",
        default=DEFAULT_REPO,
        help=f"GitHub repo (owner/name) to check issue states against (default: {DEFAULT_REPO}).",
    )
    parser.add_argument(
        "--roadmap",
        default=str(DEFAULT_ROADMAP_PATH),
        help=f"Path to roadmap markdown file (default: {DEFAULT_ROADMAP_PATH}).",
    )
    args = parser.parse_args()

    roadmap_text = Path(args.roadmap).read_text()
    rows = parse_roadmap_rows(roadmap_text)
    issue_states = fetch_issue_states(args.repo)
    drift = compute_drift(rows, issue_states)

    if not drift:
        print("No roadmap-status drift found.")
        return 0

    print(f"Found {len(drift)} roadmap-status drift row(s):")
    for entry in drift:
        print(f"  - {entry.format()}")
    return 1


if __name__ == "__main__":
    sys.exit(main())
