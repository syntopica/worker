"""Today's OpenRouter use against the key cap, for `worker status`."""

import sqlite3
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.jobs.openrouter_spent_today import openrouter_spent_today
from worker.policy.capped_queues import capped_queues


def openrouter_headroom(
    config: WorkerConfig, conn: sqlite3.Connection, now: float
) -> dict[str, Any]:
    """Attempts since 00:00 UTC, the strictest ``daily_key_cap`` and the queues already capped.

    The free key's daily allowance is spent by mid-morning on a busy day, after
    which the remote loop rests until midnight; this makes that visible.
    """
    spent = openrouter_spent_today(conn, now)
    caps = [
        q.openrouter.daily_key_cap
        for q in config.queues.values()
        if q.openrouter is not None and q.openrouter.daily_key_cap is not None
    ]
    return {
        "spent_today": sum(spent.values()),
        "key_cap": min(caps) if caps else None,
        "capped": sorted(capped_queues(config, spent)),
    }
