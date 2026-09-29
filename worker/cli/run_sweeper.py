"""The coordinator's periodic lease-reclaim and retention pass."""

import sys
import time
from collections.abc import Callable
from pathlib import Path

from worker.config.worker_config import WorkerConfig
from worker.jobs.expire_leases import expire_leases
from worker.jobs.sweep_retention import sweep_retention
from worker.store.open_store import open_store

_SWEEP_S = 60.0


def run_sweeper(
    state: Path,
    config: WorkerConfig,
    sleep: Callable[[float], None] = time.sleep,
    clock: Callable[[], float] = time.time,
    forever: bool = True,
) -> None:
    """A failing pass logs its exception class only and never stops the loop."""
    while True:
        try:
            conn = open_store(state)
            try:
                expire_leases(conn, clock())
                sweep_retention(conn, config, clock())
            finally:
                conn.close()
        except Exception as error:
            print(f"worker: sweeper pass failed: {type(error).__name__}", file=sys.stderr)
        if not forever:
            return
        sleep(_SWEEP_S)
