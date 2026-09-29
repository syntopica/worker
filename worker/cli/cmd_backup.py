"""`worker backup <dest>` - the metadata file only (spec 9)."""

import argparse
import contextlib
import os
import sqlite3
import sys
from pathlib import Path

from worker.cli.resolve_paths import resolve_paths
from worker.store.migrate_state import migrate_state


def cmd_backup(args: argparse.Namespace) -> int:
    """Online backup of meta.sqlite3 into a new 0600 file; payloads.sqlite3 is never copied."""
    _, state = resolve_paths()
    source_path = state / "meta.sqlite3"
    if not source_path.exists():
        print(f"worker: no store at {source_path}", file=sys.stderr)
        return 2
    migrate_state(state)
    dest = Path(args.dest)
    try:
        os.close(os.open(dest, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600))
    except FileExistsError:
        print(f"worker: {dest} already exists", file=sys.stderr)
        return 2
    try:
        with (
            contextlib.closing(sqlite3.connect(source_path)) as source,
            contextlib.closing(sqlite3.connect(dest)) as target,
        ):
            source.backup(target)
    except sqlite3.Error as error:
        dest.unlink(missing_ok=True)
        print(f"worker: backup failed: {type(error).__name__}", file=sys.stderr)
        return 1
    except BaseException:
        dest.unlink(missing_ok=True)
        raise
    return 0
