"""One on-demand iteration: lease a job for the named profile, run it, complete it."""

from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.tasks.run_leased_task import run_leased_task


def on_demand_step(  # noqa: PLR0913, PLR0917
    config: WorkerConfig,
    node_name: str,
    link: Any,
    profile: str,
    clock: Callable[[], float],
    sleep: Callable[[float], None],
) -> bool:
    """False once nothing is left for the profile, or its runner is resting.

    No host check: the owner started this run, so neither activity nor idle
    time holds it back.
    """
    lease = link.lease_task(False, 0.0, profile)
    if lease is None:
        return False
    run_leased_task(config, node_name, link, lease, clock, sleep)
    return True
