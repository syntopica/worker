"""Bridge one or two unreadable idle or pressure samples with the last reading."""

import dataclasses
import time
from collections.abc import Callable

from worker.node.host_state import HostState

_BRIDGE_S = 10.0


class BridgeUnreadable:
    """Wrap a sampler so a failed idle or pressure read reuses one at most ten seconds old.

    Under load ``ioreg`` and ``sysctl`` occasionally fail one two-second
    sample; on 2026-10-01 that caused 17 of 20 preemptions. A reading ten
    seconds old delays a release by at most that long. The idle time is
    reused as read, never advanced, so it cannot cross the return threshold
    on its own.
    """

    def __init__(
        self, sample: Callable[[], HostState], clock: Callable[[], float] = time.monotonic
    ) -> None:
        self._sample = sample
        self._clock = clock
        self._idle: tuple[float, float] | None = None
        self._pressure: tuple[float, str] | None = None

    def __call__(self) -> HostState:
        """One sample, with failed idle or pressure filled from a recent reading."""
        state = self._sample()
        now = self._clock()
        if state.idle_s is not None:
            self._idle = (now, state.idle_s)
        elif self._idle is not None and now - self._idle[0] <= _BRIDGE_S:
            state = dataclasses.replace(state, idle_s=self._idle[1])
        if state.pressure != "unknown":
            self._pressure = (now, state.pressure)
        elif self._pressure is not None and now - self._pressure[0] <= _BRIDGE_S:
            state = dataclasses.replace(state, pressure=self._pressure[1])
        return state
