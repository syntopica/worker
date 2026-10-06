"""One iteration of the task loop: check the host, lease a task, run it, complete it."""

import dataclasses
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.node.host_state import HostState
from worker.node.release_reason import release_reason
from worker.tasks.run_leased_task import run_leased_task


def task_step(  # noqa: PLR0913, PLR0917
    config: WorkerConfig,
    node_name: str,
    link: Any,
    sample: Callable[[], HostState],
    clock: Callable[[], float],
    sleep: Callable[[float], None],
) -> bool:
    """Return True when the loop should rest before the next iteration.

    Battery and an unreadable host stop admission; memory pressure does not,
    since the runner is a CLI talking to a remote model. A running task is
    never preempted: it costs quota, and its lease is heartbeated to the end.
    """
    node = config.nodes[node_name]
    state = sample()
    if release_reason(dataclasses.replace(state, pressure="normal"), "active_ok") is not None:
        return True
    idle = state.idle_s or 0.0
    lease = link.lease_task(idle < node.idle_threshold_s, idle)
    if lease is None:
        return True
    run_leased_task(config, node_name, link, lease, clock, sleep)
    return False
