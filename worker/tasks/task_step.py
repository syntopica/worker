"""One iteration of the task loop: check the host, lease a task, run it, complete it."""

from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.node.host_state import HostState
from worker.node.release_reason import release_reason
from worker.tasks.run_leased_task import run_leased_task
from worker.tasks.task_admission_state import task_admission_state


def task_step(  # noqa: PLR0913, PLR0917
    config: WorkerConfig,
    node_name: str,
    link: Any,
    sample: Callable[[], HostState],
    clock: Callable[[], float],
    sleep: Callable[[float], None],
) -> bool:
    """Return True when the loop should rest before the next iteration.

    Only a known battery stops admission: the runner is a CLI talking to a
    remote model, so memory pressure and an unreadable reader do not, and an
    unreadable idle time leases as an active user. A running task is
    never preempted: it costs quota, and its lease is heartbeated to the end.
    """
    node = config.nodes[node_name]
    state = task_admission_state(sample())
    if release_reason(state, "active_ok") is not None:
        return True
    idle = state.idle_s or 0.0
    lease = link.lease_task(idle < node.idle_threshold_s, idle)
    if lease is None:
        return True
    run_leased_task(config, node_name, link, lease, clock, sleep)
    return False
