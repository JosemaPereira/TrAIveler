#!/usr/bin/env python3
"""Unit tests for check-roadmap-status-drift.py's pure parsing/comparison logic.

Only the pure logic is tested here (row parsing, issue-number extraction,
drift comparison) — the `gh issue list` network boundary (fetch_issue_states)
is intentionally not covered by these tests, mirroring the mandate to mock or
inject that boundary rather than hit the network from a unit test.

The script under test is named with a hyphen (matching this repo's existing
scripts/check-agent-drift.py convention), so it isn't a valid Python module
name and is loaded dynamically via importlib rather than a plain `import`.
"""

import importlib.util
import sys
import unittest
from pathlib import Path

SCRIPT_PATH = Path(__file__).resolve().parent / "check-roadmap-status-drift.py"

_spec = importlib.util.spec_from_file_location("check_roadmap_status_drift", SCRIPT_PATH)
drift_module = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(drift_module)


HEADER = "| ID | Task | Group | Sprint | Priority | Status | Depends on | Parallel | Issue | Notes |"
SEPARATOR = "|----|------|-------|--------|----------|--------|------------|----------|-------|-------|"


def make_row(
    row_id="001-T001",
    status="Done",
    issue="https://github.com/JosemaPereira/TrAIveler/issues/42",
    notes="Some notes",
):
    """Build one markdown table data-row line with the given field values."""
    return f"| {row_id} | Some task | | 6 | P1 | {status} | - | no | {issue} | {notes} |"


class TestParseRoadmapRows(unittest.TestCase):
    """Behavior of parse_roadmap_rows: which lines become RoadmapRow entries."""

    def test_valid_row_with_issue_url_is_parsed(self):
        # Arrange
        text = "\n".join([HEADER, SEPARATOR, make_row()])

        # Act
        rows = drift_module.parse_roadmap_rows(text)

        # Assert
        self.assertEqual(len(rows), 1)
        self.assertEqual(rows[0].id, "001-T001")
        self.assertEqual(rows[0].status, "Done")
        self.assertEqual(rows[0].issue_number, 42)

    def test_header_and_separator_rows_are_skipped(self):
        # Arrange
        text = "\n".join([HEADER, SEPARATOR])

        # Act
        rows = drift_module.parse_roadmap_rows(text)

        # Assert
        self.assertEqual(rows, [])

    def test_row_with_empty_issue_cell_is_skipped(self):
        # Arrange
        text = make_row(issue="")

        # Act
        rows = drift_module.parse_roadmap_rows(text)

        # Assert
        self.assertEqual(rows, [])

    def test_row_with_non_url_issue_cell_is_skipped(self):
        # Arrange
        text = make_row(issue="TBD")

        # Act
        rows = drift_module.parse_roadmap_rows(text)

        # Assert
        self.assertEqual(rows, [])

    def test_non_table_line_is_skipped(self):
        # Arrange
        text = "This is prose, not a table row."

        # Act
        rows = drift_module.parse_roadmap_rows(text)

        # Assert
        self.assertEqual(rows, [])

    def test_multiple_rows_sharing_the_same_issue_are_all_parsed(self):
        # Arrange
        text = "\n".join([
            make_row(row_id="008-T037", issue="https://github.com/JosemaPereira/TrAIveler/issues/167"),
            make_row(row_id="008-T038", issue="https://github.com/JosemaPereira/TrAIveler/issues/167"),
        ])

        # Act
        rows = drift_module.parse_roadmap_rows(text)

        # Assert
        self.assertEqual(len(rows), 2)
        self.assertTrue(all(row.issue_number == 167 for row in rows))

    def test_escaped_pipe_in_notes_column_does_not_corrupt_earlier_columns(self):
        # Arrange: a literal "\|" inside the Notes cell (seen for real in
        # docs/roadmap.md, e.g. "role: 'admin'\|'partner'") must not shift
        # the ID/Status/Issue columns that come before it.
        text = (
            "| 005-T046 | Task | | 2 | P1 | Done | - | yes | "
            "https://github.com/JosemaPereira/TrAIveler/issues/61 | "
            "notes with 'admin'\\|'partner' inside |"
        )

        # Act
        rows = drift_module.parse_roadmap_rows(text)

        # Assert
        self.assertEqual(len(rows), 1)
        self.assertEqual(rows[0].id, "005-T046")
        self.assertEqual(rows[0].status, "Done")
        self.assertEqual(rows[0].issue_number, 61)


class TestComputeDrift(unittest.TestCase):
    """Behavior of compute_drift: the Status-vs-GitHub-state comparison rules."""

    def test_closed_issue_with_non_done_status_is_drift(self):
        # Arrange
        row = drift_module.RoadmapRow(
            id="001-T001", status="Backlog",
            issue_url="https://github.com/JosemaPereira/TrAIveler/issues/1", issue_number=1,
        )

        # Act
        drift = drift_module.compute_drift([row], {1: "CLOSED"})

        # Assert
        self.assertEqual(len(drift), 1)
        self.assertEqual(drift[0].row.id, "001-T001")
        self.assertEqual(drift[0].github_state, "CLOSED")

    def test_closed_issue_with_done_status_is_not_drift(self):
        # Arrange
        row = drift_module.RoadmapRow(
            id="001-T001", status="Done",
            issue_url="https://github.com/JosemaPereira/TrAIveler/issues/1", issue_number=1,
        )

        # Act
        drift = drift_module.compute_drift([row], {1: "CLOSED"})

        # Assert
        self.assertEqual(drift, [])

    def test_open_issue_with_done_status_is_not_drift(self):
        # Arrange: a row is flipped to Done in the same PR that closes its
        # issue — the issue is still OPEN until that PR merges, so this must
        # never be reported as drift (it would trip on every closing PR).
        row = drift_module.RoadmapRow(
            id="002-T047", status="Done",
            issue_url="https://github.com/JosemaPereira/TrAIveler/issues/173", issue_number=173,
        )

        # Act
        drift = drift_module.compute_drift([row], {173: "OPEN"})

        # Assert
        self.assertEqual(drift, [])

    def test_open_issue_with_in_progress_status_is_not_drift(self):
        # Arrange
        row = drift_module.RoadmapRow(
            id="002-T047", status="In Progress",
            issue_url="https://github.com/JosemaPereira/TrAIveler/issues/173", issue_number=173,
        )

        # Act
        drift = drift_module.compute_drift([row], {173: "OPEN"})

        # Assert
        self.assertEqual(drift, [])

    def test_closed_issue_with_superseded_status_is_not_drift(self):
        # Arrange: superseded rows are intentionally not tracked to an
        # issue's lifecycle, even when the linked issue is closed.
        row = drift_module.RoadmapRow(
            id="001-T012", status="Superseded",
            issue_url="https://github.com/JosemaPereira/TrAIveler/issues/168", issue_number=168,
            notes="Some notes",
        )

        # Act
        drift = drift_module.compute_drift([row], {168: "CLOSED"})

        # Assert
        self.assertEqual(drift, [])

    def test_closed_issue_with_backlog_status_but_blocked_notes_is_not_drift(self):
        # Arrange: a row can stay genuinely unfinished (e.g. a GitHub-plan
        # limitation) while its parent group issue is closed because the
        # issue's *other* rows did complete. "**Blocked**" at the start of
        # Notes is this repo's existing convention for that case (e.g.
        # 009-T019, 002-T050) — it must not be reported as drift.
        row = drift_module.RoadmapRow(
            id="009-T019", status="Backlog",
            issue_url="https://github.com/JosemaPereira/TrAIveler/issues/172", issue_number=172,
            notes="**Blocked**, re-confirmed 2026-07-31: same 403 as the ruleset pattern.",
        )

        # Act
        drift = drift_module.compute_drift([row], {172: "CLOSED"})

        # Assert
        self.assertEqual(drift, [])

    def test_closed_issue_with_backlog_status_and_unblocked_notes_is_drift(self):
        # Arrange: without the "**Blocked**" marker, a closed-but-not-Done
        # row is still reported as drift as normal.
        row = drift_module.RoadmapRow(
            id="001-T001", status="Backlog",
            issue_url="https://github.com/JosemaPereira/TrAIveler/issues/1", issue_number=1,
            notes="Some notes",
        )

        # Act
        drift = drift_module.compute_drift([row], {1: "CLOSED"})

        # Assert
        self.assertEqual(len(drift), 1)

    def test_unknown_issue_number_is_skipped_without_being_reported_as_drift(self):
        # Arrange: gh issue list didn't return this number (e.g. stale/typo'd link).
        row = drift_module.RoadmapRow(
            id="999-T999", status="Done",
            issue_url="https://github.com/JosemaPereira/TrAIveler/issues/999999", issue_number=999999,
        )

        # Act
        drift = drift_module.compute_drift([row], {})

        # Assert
        self.assertEqual(drift, [])

    def test_no_rows_yields_no_drift(self):
        # Act
        drift = drift_module.compute_drift([], {1: "OPEN"})

        # Assert
        self.assertEqual(drift, [])


if __name__ == "__main__":
    unittest.main()
