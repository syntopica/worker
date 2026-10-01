"""Sample a headless Linux server: host load stands in for the user."""

import sys
import time
from collections.abc import Callable

from worker.node.host_state import HostState
from worker.node.parse_loadavg import parse_loadavg
from worker.node.parse_meminfo_free_pct import parse_meminfo_free_pct
from worker.node.read_text_file import read_text_file


class LinuxHostSampler:
    """A server has no keyboard and no battery; its users are the services it hosts.

    ``idle_s`` is the time since the one-minute load average last exceeded
    ``max_load`` (no limit: since the sampler started), so a busy host blocks
    idle-only queues and releases their attempts exactly as a returning user
    does. The load includes the node's own model, so ``max_load`` must leave
    room for it. ``on_ac`` is always true. Memory counts as ``warn`` below
    ``min_free_pct`` available.
    """

    def __init__(
        self,
        max_load: float | None,
        min_free_pct: float,
        read: Callable[[str], str | None] = read_text_file,
        clock: Callable[[], float] = time.monotonic,
    ) -> None:
        self._max_load = max_load
        self._min_free_pct = min_free_pct
        self._read = read
        self._clock = clock
        self._quiet_since = clock()

    def __call__(self) -> HostState:
        """One sample; an unreadable file degrades its field to unreadable."""
        now = self._clock()
        raw_load = self._read("/proc/loadavg")
        raw_mem = self._read("/proc/meminfo")
        load = parse_loadavg(raw_load) if raw_load is not None else None
        free = parse_meminfo_free_pct(raw_mem) if raw_mem is not None else None
        failed = [n for n, v in (("loadavg", load), ("meminfo", free)) if v is None]
        if failed:
            print(f"worker: host readers failed: {','.join(failed)}", file=sys.stderr)
        if load is not None and self._max_load is not None and load > self._max_load:
            self._quiet_since = now
        idle = None if load is None else now - self._quiet_since
        return HostState(idle, True, self._pressure(free))

    def _pressure(self, free: float | None) -> str:
        """``unknown`` when unreadable, ``warn`` below the free threshold."""
        if free is None:
            return "unknown"
        return "warn" if free < self._min_free_pct else "normal"
