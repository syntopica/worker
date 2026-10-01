"""Run the OpenRouter loop in several slots, each with its own coordinator link."""

import threading
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.remote.run_remote_node import run_remote_node

_JOIN_S = 18.0


def run_remote_slots(  # noqa: PLR0913, PLR0917
    config: WorkerConfig,
    node_name: str,
    links: list[Any],
    key: str,
    current: Callable[[], WorkerConfig] | None = None,
    run: Callable[..., None] = run_remote_node,
) -> None:
    """The first slot runs here; the others in threads that stop with it.

    When this loop is stopped (SIGTERM raises here) the other slots are told
    to stop, hand back their running job at the next heartbeat, and are
    waited for briefly so launchd does not kill them mid-report.
    """
    stop = threading.Event()

    def rest(seconds: float) -> None:
        stop.wait(seconds)

    threads = [
        threading.Thread(
            target=run,
            args=(config, node_name, link, key, rest),
            kwargs={"current": current, "stopping": stop.is_set},
            daemon=True,
        )
        for link in links[1:]
    ]
    for thread in threads:
        thread.start()
    try:
        run(config, node_name, links[0], key, current=current)
    finally:
        stop.set()
        for thread in threads:
            thread.join(_JOIN_S)
