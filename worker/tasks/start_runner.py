"""Start a runner in its own process group, in its workspace, with its environment."""

import os
import subprocess
import threading
from pathlib import Path
from typing import IO

from worker.config.task_profile import TaskProfile
from worker.tasks.feed_stdin import feed_stdin
from worker.tasks.runner_invocation import RunnerInvocation


def start_runner(
    profile: TaskProfile,
    invocation: RunnerInvocation,
    workspace: Path,
    streams: tuple[IO[bytes], IO[bytes]],
) -> subprocess.Popen[bytes]:
    """Stdin is fed from a thread so a large prompt cannot block this loop."""
    env = {k: v for k, v in os.environ.items() if k not in profile.env_unset}
    proc = subprocess.Popen(
        invocation.argv,
        cwd=workspace,
        stdin=subprocess.PIPE if invocation.stdin is not None else subprocess.DEVNULL,
        stdout=streams[0],
        stderr=streams[1],
        env=env,
        start_new_session=True,
    )
    if invocation.stdin is not None and proc.stdin is not None:
        pipe, text = proc.stdin, invocation.stdin
        threading.Thread(target=feed_stdin, args=(pipe, text), daemon=True).start()
    return proc
