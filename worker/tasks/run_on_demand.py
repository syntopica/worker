"""Drain an on-demand profile's jobs in several slots, then return."""

import threading
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.tasks.drain_on_demand import drain_on_demand


def run_on_demand(
    config: WorkerConfig,
    node_name: str,
    links: list[Any],
    profile: str,
    drain: Callable[..., int] = drain_on_demand,
) -> int:
    """The total jobs run; each slot stops on its own once its lease comes back empty.

    The first slot runs here, so a SIGTERM reports its attempt as a shutdown;
    the other slots' leases are left to expire.
    """
    counts = [0] * len(links)

    def run_slot(index: int) -> None:
        counts[index] = drain(config, node_name, links[index], profile)

    threads = [
        threading.Thread(target=run_slot, args=(i,), daemon=True) for i in range(1, len(links))
    ]
    for thread in threads:
        thread.start()
    run_slot(0)
    for thread in threads:
        thread.join()
    return sum(counts)
