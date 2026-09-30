"""Park a job that will not be split: queued, runnable only in a long enough idle period (spec 7)."""

import sqlite3

from worker.jobs.queue_p90_runtime import queue_p90_runtime


def park_job(
    conn: sqlite3.Connection, job_id: str, queue: str, parked_min_idle_s: float, now: float
) -> None:
    """Require an idle period of 1.5 times the queue's p90 runtime, at least ``parked_min_idle_s``."""
    p90 = queue_p90_runtime(conn, queue) or 0.0
    need = max(1.5 * p90, parked_min_idle_s)
    conn.execute(
        "UPDATE jobs SET state='queued', parked=1, parked_min_idle_s=?, updated=? WHERE id=?",
        (need, now, job_id),
    )
