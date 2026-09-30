"""The node loop: sample, report, lease, run - one attempt at a time."""

import functools
import sys
import time
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.node.host_state import HostState
from worker.node.node_memory import NodeMemory
from worker.node.node_step import node_step
from worker.node.sample_host_state import sample_host_state

_REST_S = 30.0


def run_node(  # noqa: PLR0913, PLR0917
    config: WorkerConfig,
    node_name: str,
    link: Any,
    sample: Callable[[], HostState] | None = None,
    sleep: Callable[[float], None] = time.sleep,
    clock: Callable[[], float] = time.time,
    forever: bool = True,
    current: Callable[[], WorkerConfig] | None = None,
) -> None:
    """A failing iteration logs its exception class, rests, and never stops the loop.

    Without an explicit ``sample`` the node reads its own host with its
    configured ``min_free_pct``.
    """
    if sample is None:
        sample = functools.partial(
            sample_host_state, min_free_pct=config.nodes[node_name].min_free_pct
        )
    memory = NodeMemory()
    while True:
        try:
            if current is not None:
                config = current()
            rest = node_step(config, node_name, link, memory, sample, sleep, clock)
        except Exception as error:
            print(f"worker: node iteration failed: {type(error).__name__}", file=sys.stderr)
            rest = True
        if not forever:
            return
        if rest:
            sleep(_REST_S)
