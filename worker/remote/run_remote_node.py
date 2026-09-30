"""The OpenRouter loop: one escalated inference job at a time."""

import sys
import time
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.remote.remote_step import remote_step

_ERROR_REST_S = 60.0


def run_remote_node(  # noqa: PLR0913, PLR0917
    config: WorkerConfig,
    node_name: str,
    link: Any,
    key: str,
    sleep: Callable[[float], None] = time.sleep,
    clock: Callable[[], float] = time.time,
    *,
    forever: bool = True,
    current: Callable[[], WorkerConfig] | None = None,
) -> None:
    """A failing iteration logs its exception class, rests, and never stops the loop."""
    while True:
        try:
            if current is not None:
                config = current()
            rest = remote_step(config, node_name, link, key, clock)
        except Exception as error:
            print(f"worker: remote iteration failed: {type(error).__name__}", file=sys.stderr)
            rest = _ERROR_REST_S
        if not forever:
            return
        sleep(rest)
