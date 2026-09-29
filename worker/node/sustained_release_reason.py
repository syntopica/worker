"""Release reason where memory pressure counts only once it is sustained."""

import dataclasses

from worker.node.host_state import HostState
from worker.node.pressure_samples import PRESSURE_SAMPLES
from worker.node.release_reason import release_reason


def sustained_release_reason(state: HostState, run_when: str, streak: int) -> str | None:
    """``streak`` counts consecutive warn/critical samples, this one included.

    Below ``PRESSURE_SAMPLES`` a pressured sample is judged as if pressure were
    normal, so battery, input and unreadable state still release at once.
    """
    reason = release_reason(state, run_when)
    if reason == "memory_pressure" and streak < PRESSURE_SAMPLES:
        return release_reason(dataclasses.replace(state, pressure="normal"), run_when)
    return reason
