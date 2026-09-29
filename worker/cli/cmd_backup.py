"""`worker backup <dest>` - the metadata file only (spec 9)."""

import argparse
import sqlite3

from worker.cli.resolve_paths import resolve_paths


def cmd_backup(args: argparse.Namespace) -> int:
    """Uses SQLite's online backup of meta.sqlite3; payloads.sqlite3 is never copied."""
    _, state = resolve_paths()
    source = sqlite3.connect(state / "meta.sqlite3")
    target = sqlite3.connect(args.dest)
    with target:
        source.backup(target)
    source.close()
    target.close()
    return 0
