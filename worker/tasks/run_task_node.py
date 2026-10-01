"""The task loop: one read-only task at a time, beside the inference node."""

import functools
import sys
import time
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.node.bridge_unreadable import BridgeUnreadable
from worker.node.hold_power_source import HoldPowerSource
from worker.node.host_state import HostState
from worker.node.sample_host_state import sample_host_state
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
) -> None:
    """A failing iteration logs its exception class, rests, and never stops the loop.

    Every ten minutes the runners' quotas are read through CodexBar and any
    spent one is reported, so the coordinator rests it (and uses a queue
    fallback) without first spending a call against its wall.
    """
    next_probe = 0.0
    if sample is None:
        sample = HoldPowerSource(
            BridgeUnreadable(
                functools.partial(
                    sample_host_state, min_free_pct=config.nodes[node_name].min_free_pct
                )
            )
        )
    while True:
        try:
            if current is not None:
                config = current()
            if config.runner_quota and clock() >= next_probe:
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
