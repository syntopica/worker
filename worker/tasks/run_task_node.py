"""The task loop: one read-only task at a time per slot, beside the inference node."""

import sys
import time
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.node.default_host_sampler import default_host_sampler
from worker.node.host_state import HostState
from worker.tasks.probe_runner_walls import probe_runner_walls
from worker.tasks.task_step import task_step

_REST_S = 30.0
_PROBE_S = 600.0


def run_task_node(  # noqa: PLR0913, PLR0917
    config: WorkerConfig,
    node_name: str,
    link: Any,
    sample: Callable[[], HostState] | None = None,
    sleep: Callable[[float], None] = time.sleep,
    clock: Callable[[], float] = time.time,
    forever: bool = True,
    current: Callable[[], WorkerConfig] | None = None,
    *,
    probe: bool = True,
) -> None:
    """A failing iteration logs its exception class, rests, and never stops the loop.

    Every ten minutes the runners' quotas are read through CodexBar and any
    spent one is reported, so the coordinator rests it (and uses a queue
    fallback) without first spending a call against its wall; with several
    slots only the first one probes.
    """
    next_probe = 0.0
    if sample is None:
        sample = default_host_sampler(config.nodes[node_name])
    while True:
        try:
            if current is not None:
                config = current()
            if probe and config.runner_quota and clock() >= next_probe:
                next_probe = clock() + _PROBE_S
                walls = probe_runner_walls(config, clock())
                if walls:
                    link.report_walls(walls)
            rest = task_step(config, node_name, link, sample, clock, sleep)
        except Exception as error:
            print(f"worker: task iteration failed: {type(error).__name__}", file=sys.stderr)
            rest = True
        if not forever:
            return
        if rest:
            sleep(_REST_S)
