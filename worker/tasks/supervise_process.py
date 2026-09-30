"""Wait for a runner while keeping its lease alive and its clock bounded."""

import subprocess
from collections.abc import Callable

from worker.tasks.kill_group import kill_group

HEARTBEAT_EVERY_S = 20.0
_POLL_S = 2.0


def supervise_process(
    proc: subprocess.Popen[bytes],
    timeout_s: float,
    heartbeat: Callable[[], bool],
    clock: Callable[[], float],
    sleep: Callable[[float], None],
) -> str:
    """``exited``, ``fenced`` (the lease is someone else's now) or ``timeout``.

    The lease TTL is 60 s, so a heartbeat every 20 s survives two lost ones.
    Fenced and timed-out runners are killed as a group before returning.
    """
    started = last_beat = clock()
    while proc.poll() is None:
        sleep(_POLL_S)
        now = clock()
        if now - started > timeout_s:
            kill_group(proc)
            return "timeout"
        if now - last_beat >= HEARTBEAT_EVERY_S:
            last_beat = now
            if not heartbeat():
                kill_group(proc)
                return "fenced"
    return "exited"
