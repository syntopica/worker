"""The host state the task loop admits work on: only a known battery blocks."""

import dataclasses

from worker.node.host_state import HostState


def task_admission_state(state: HostState) -> HostState:
    """Unreadable idle reads as an active user, unreadable power as mains.

    The runners are remote CLIs, so memory pressure is ignored and an
    unreadable reader must not stop every runner queue (spec amendment
    2026-11-05); an active user still limits the lease to ``active_ok`` queues.
    """
    return dataclasses.replace(
        state,
        idle_s=0.0 if state.idle_s is None else state.idle_s,
        on_ac=True if state.on_ac is None else state.on_ac,
        pressure="normal",
    )
