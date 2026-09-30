"""Kill a runner and everything it started."""

import contextlib
import os
import signal
import subprocess


def kill_group(proc: subprocess.Popen[bytes]) -> None:
    """SIGKILL the process group, then reap; the runner leader may already be gone."""
    with contextlib.suppress(ProcessLookupError, PermissionError):
        os.killpg(proc.pid, signal.SIGKILL)
    with contextlib.suppress(subprocess.TimeoutExpired):
        proc.wait(timeout=10)
