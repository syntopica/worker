"""Bridge a transiently unreadable power source with the last reading."""

import dataclasses
import time
from collections.abc import Callable

from worker.node.host_state import HostState

_HOLD_S = 120.0


class HoldPowerSource:
    """Wrap a sampler so a failed ``pmset`` read reuses a reading at most two minutes old.

    ``pmset -g ps`` failed nine times on 2026-09-30 even after one immediate
    retry, and each failure preempted the running attempt as
    ``host_state_unreadable``. The power source changes rarely; trusting a
    two-minute-old reading costs at most two minutes on battery.
    """

    def __init__(
        self, sample: Callable[[], HostState], clock: Callable[[], float] = time.monotonic
    ) -> None:
        self._sample = sample
        self._clock = clock
        self._last: tuple[float, bool] | None = None

    def __call__(self) -> HostState:
        """One sample, with ``on_ac`` filled from a recent reading when it failed."""
        state = self._sample()
        now = self._clock()
        if state.on_ac is not None:
            self._last = (now, state.on_ac)
            return state
        if self._last is not None and now - self._last[0] <= _HOLD_S:
            return dataclasses.replace(state, on_ac=self._last[1])
        return state
