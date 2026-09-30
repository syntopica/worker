"""Set the Time Machine exclusion attribute on one file, once per file identity."""

import plistlib
import subprocess
import sys

from worker.store.excluded_files import EXCLUDED_FILES

_ATTRIBUTE = "com.apple.metadata:com_apple_backup_excludeItem"
_VALUE = plistlib.dumps("com.apple.backupd", fmt=plistlib.FMT_BINARY).hex()


def mark_backup_excluded(path: str, inode: int) -> None:
    """What ``tmutil addexclusion`` writes; ``inode`` re-marks a recreated file.

    Only a success is remembered, so a failure is retried on the next open.
    A failure logs its class on one line and never raises.
    """
    if (path, inode) in EXCLUDED_FILES:
        return
    try:
        done = subprocess.run(
            ["/usr/bin/xattr", "-wx", _ATTRIBUTE, _VALUE, path],
            capture_output=True,
            timeout=5,
            check=False,
        )
    except (OSError, subprocess.TimeoutExpired) as error:
        print(f"worker: backup exclusion failed: {type(error).__name__}", file=sys.stderr)
        return
    if done.returncode != 0:
        print(f"worker: backup exclusion failed: exit {done.returncode}", file=sys.stderr)
        return
    EXCLUDED_FILES.add((path, inode))
