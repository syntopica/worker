"""Keep the payload file out of Time Machine (spec 9: content is never backed up)."""

import sys
from pathlib import Path

from worker.store.mark_backup_excluded import mark_backup_excluded

_NAMES = ("payloads.sqlite3", "payloads.sqlite3-wal", "payloads.sqlite3-shm")


def exclude_payloads_from_backup(state_dir: Path) -> None:
    """Only on macOS. The metadata file stays eligible: it is the file that is backed up."""
    # Read into a variable: a direct `sys.platform` test is resolved by mypy
    # per host, so on Linux the loop below reads as unreachable.
    current_platform: str = sys.platform
    if current_platform != "darwin":
        return
    for name in _NAMES:
        path = state_dir / name
        try:
            inode = path.stat().st_ino
        except OSError:
            continue
        mark_backup_excluded(str(path), inode)
