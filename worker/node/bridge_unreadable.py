"""Bridge one unreadable idle or pressure sample with the last reading."""

import dataclasses
import time
from collections.abc import Callable

from worker.node.host_state import HostState

_BRIDGE_S = 30.0


class BridgeUnreadable:
    """Wrap a sampler so a failed idle or pressure read reuses one at most thirty seconds old.

    Under load the host readers occasionally fail one sample; on 2026-10-01
    that caused 17 of 20 preemptions. Each reader runs in turn with a five
    second timeout, so a sample in which they all time out lasts about 25 s:
    the window covers one such sample, and a reading that old delays a
    release by at most that long. The idle time is reused as read, never
    advanced, so it cannot cross the return threshold on its own.
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
