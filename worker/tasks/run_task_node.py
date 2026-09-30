"""The task loop: one read-only task at a time, beside the inference node."""

import functools
import sys
import time
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.node.host_state import HostState
from worker.node.sample_host_state import sample_host_state
from worker.tasks.task_step import task_step

_REST_S = 30.0


def run_task_node(  # noqa: PLR0913, PLR0917
    config: WorkerConfig,
    node_name: str,
    link: Any,
    sample: Callable[[], HostState] | None = None,
    sleep: Callable[[float], None] = time.sleep,
    clock: Callable[[], float] = time.time,
    forever: bool = True,
    current: Callable[[], WorkerConfig] | None = None,
) -> None:
    """A failing iteration logs its exception class, rests, and never stops the loop."""
    if sample is None:
        sample = functools.partial(
            sample_host_state, min_free_pct=config.nodes[node_name].min_free_pct
        )
    while True:
        try:
            if current is not None:
                config = current()
            rest = task_step(config, node_name, link, sample, clock, sleep)
        except Exception as error:
            print(f"worker: task iteration failed: {type(error).__name__}", file=sys.stderr)
            rest = True
        if not forever:
            return
        if rest:
            sleep(_REST_S)
