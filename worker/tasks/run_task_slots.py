"""Run the task loop in several slots, each with its own coordinator link."""

import threading
import time
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.tasks.run_task_node import run_task_node
from worker.tasks.stoppable_sleep import stoppable_sleep

_JOIN_S = 12.0


def run_task_slots(
    config: WorkerConfig,
    node_name: str,
    links: list[Any],
    current: Callable[[], WorkerConfig] | None = None,
    run: Callable[..., None] = run_task_node,
) -> None:
    """The first slot runs here and alone probes quotas; the others run in threads.

    When this loop is stopped (SIGTERM raises here) the other slots stop
    within one poll: their runners are killed and their jobs handed back,
    all waited for together for at most ``_JOIN_S``, inside launchd's 20 s.
    """
    stop = threading.Event()
    threads = [
        threading.Thread(
            target=run,
            args=(config, node_name, link),
            kwargs={"sleep": stoppable_sleep(stop), "current": current, "probe": False},
            daemon=True,
        )
        for link in links[1:]
    ]
    for thread in threads:
        thread.start()
    try:
        run(config, node_name, links[0], current=current)
    finally:
        stop.set()
        deadline = time.monotonic() + _JOIN_S
        for thread in threads:
            thread.join(max(0.0, deadline - time.monotonic()))
