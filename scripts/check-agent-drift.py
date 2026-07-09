#!/usr/bin/env python3
"""Detect drift between hand-ported Claude Code subagents and their GitHub Copilot sources.

Compares the current sha256 of every file listed in scripts/agent-port-manifest.json
against the hash recorded there. Reports, per pair:
  - in sync        both sides match the manifest
  - claude changed  .claude/agents/<name>.md changed, Copilot side not yet updated
  - copilot changed one or more .github/agents|prompts files changed, Claude side not yet updated
  - both changed    edited on both sides independently — needs manual reconciliation
  - missing file    a file the manifest expects no longer exists

Exit code is non-zero if any pair is not "in sync", so this can be used as a check/gate.

After manually reconciling a drifted pair (porting the change to the other side,
following the adaptation rules in CLAUDE.md's "Local Agents & Workflows" section),
re-run with --update to refresh the stored hashes for that pair.
"""

import argparse
import hashlib
import json
import sys
from datetime import date, datetime, timezone
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
MANIFEST_PATH = Path(__file__).resolve().parent / "agent-port-manifest.json"


def sha256_of(path: Path) -> str | None:
    if not path.exists():
        return None
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load_manifest() -> dict:
    return json.loads(MANIFEST_PATH.read_text())


def save_manifest(manifest: dict) -> None:
    MANIFEST_PATH.write_text(json.dumps(manifest, indent=2) + "\n")


def check_pair(pair: dict) -> tuple[str, dict[str, tuple[str | None, str | None]]]:
    """Returns (status, {relpath: (recorded_hash, current_hash)}) for changed/missing files."""
    changes: dict[str, tuple[str | None, str | None]] = {}
    claude_changed = False
    copilot_changed = False

    all_files = [pair["claude_file"], *pair["copilot_files"]]
    for relpath in all_files:
        recorded = pair["hashes"].get(relpath)
        current = sha256_of(REPO_ROOT / relpath)
        if current != recorded:
            changes[relpath] = (recorded, current)
            if relpath == pair["claude_file"]:
                claude_changed = True
            else:
                copilot_changed = True

    if not changes:
        return "in sync", changes
    if claude_changed and copilot_changed:
        return "both changed", changes
    if claude_changed:
        return "claude changed", changes
    return "copilot changed", changes


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--update",
        metavar="NAME",
        help="After manually reconciling pair NAME, refresh its stored hashes to match "
        "the current files (or pass 'all' to refresh every pair).",
    )
    args = parser.parse_args()

    manifest = load_manifest()

    if args.update:
        targets = manifest["pairs"] if args.update == "all" else [
            p for p in manifest["pairs"] if p["name"] == args.update
        ]
        if not targets:
            print(f"No pair named '{args.update}' in {MANIFEST_PATH}", file=sys.stderr)
            return 2
        for pair in targets:
            for relpath in [pair["claude_file"], *pair["copilot_files"]]:
                current = sha256_of(REPO_ROOT / relpath)
                if current is None:
                    print(f"warning: {relpath} does not exist, leaving its hash unset", file=sys.stderr)
                    pair["hashes"].pop(relpath, None)
                else:
                    pair["hashes"][relpath] = current
            pair["last_synced"] = date.today().isoformat()
        save_manifest(manifest)
        print(f"Updated hashes for: {', '.join(p['name'] for p in targets)}")
        return 0

    any_drift = False
    for pair in manifest["pairs"]:
        status, changes = check_pair(pair)
        if status == "in sync":
            continue
        any_drift = True
        print(f"[{status}] {pair['name']}")
        for relpath, (recorded, current) in changes.items():
            if current is None:
                print(f"    MISSING  {relpath}")
            elif recorded is None:
                print(f"    NEW      {relpath}")
            else:
                print(f"    CHANGED  {relpath}")
        print()

    if not any_drift:
        print(f"All {len(manifest['pairs'])} hand-ported agent pairs are in sync.")
        return 0

    print(
        "Drift detected. Port the change to the other side by hand (see CLAUDE.md's "
        "\"Local Agents & Workflows\" section for the adaptation rules), then run:\n"
        "  python3 scripts/check-agent-drift.py --update <name>"
    )
    return 1


if __name__ == "__main__":
    sys.exit(main())
