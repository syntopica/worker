"""Set the Time Machine exclusion attribute on one file, once per file identity."""

import functools
import plistlib
import subprocess
import sys

_ATTRIBUTE = "com.apple.metadata:com_apple_backup_excludeItem"
_VALUE = plistlib.dumps("com.apple.backupd", fmt=plistlib.FMT_BINARY).hex()


@functools.lru_cache(maxsize=64)
def mark_backup_excluded(path: str, inode: int) -> None:  # noqa: ARG001
    """What ``tmutil addexclusion`` writes; ``inode`` re-marks a recreated file.

    A failure logs its class on one line and never raises.
    """
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
