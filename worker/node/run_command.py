"""Run a read-only system command; None on any failure (spec 7: failure = busy)."""

import errno
import subprocess
import sys
from pathlib import Path


def run_command(args: list[str], timeout: float = 5.0) -> str | None:
    """Stdout on exit 0, else None.

    A failure is logged as the command's base name and the kind of failure
    (timeout, exit status or OS error name) and nothing else, so a flapping
    host reader can be told apart from a slow one.
    """
    name = Path(args[0]).name
    try:
        done = subprocess.run(args, capture_output=True, text=True, timeout=timeout, check=False)
    except subprocess.TimeoutExpired:
        print(f"worker: command failed: {name} timeout {timeout:g}s", file=sys.stderr)
        return None
    except OSError as error:
        kind = errno.errorcode.get(error.errno or 0, "unknown")
        print(f"worker: command failed: {name} oserror {kind}", file=sys.stderr)
        return None
    if done.returncode != 0:
        print(f"worker: command failed: {name} exit {done.returncode}", file=sys.stderr)
        return None
    return done.stdout
