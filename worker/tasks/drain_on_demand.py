"""Run on-demand jobs through one coordinator link until none is left."""

import sys
import time
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.tasks.on_demand_step import on_demand_step


def drain_on_demand(
    config: WorkerConfig,
    node_name: str,
    link: Any,
    profile: str,
    step: Callable[..., bool] = on_demand_step,
) -> int:
    """The number of jobs this link ran; a failing iteration logs its class and ends it.

    Held completions are retried once more at the end, since no later lease
    will carry them.
    """
    ran = 0
    try:
        while step(config, node_name, link, profile, time.time, time.sleep):
            ran += 1
    except Exception as error:
        print(f"worker: on-demand iteration failed: {type(error).__name__}", file=sys.stderr)
    link.flush()
    return ran
