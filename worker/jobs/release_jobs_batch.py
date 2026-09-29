"""Forget terminal jobs past retention, which releases their idempotency keys (spec 6)."""

import sqlite3

from worker.config.queue_policy import QueuePolicy
from worker.jobs.states import RELEASABLE


def release_jobs_batch(
    conn: sqlite3.Connection, policy: QueuePolicy, now: float, limit: int
) -> int:
    """A job whose payloads are gone and whose last transition or ack is older than retention.

    Its results and attempts go with it; status needs only the last hour.
    Return how many jobs were removed.
    """
    marks = ",".join("?" * len(RELEASABLE))
    ids = [
        r[0]
        for r in conn.execute(
            "SELECT id FROM jobs WHERE payloads_deleted=1 AND queue=? AND updated<=?"  # noqa: S608
            f" AND (state IN ({marks}) OR (state='succeeded' AND acked IS NOT NULL)) LIMIT ?",
            (policy.name, now - policy.retention_days * 86400, *RELEASABLE, limit),
        )
    ]
    if ids:
        where = ",".join("?" * len(ids))
        conn.execute(f"DELETE FROM results WHERE job_id IN ({where})", ids)  # noqa: S608
        conn.execute(f"DELETE FROM attempts WHERE job_id IN ({where})", ids)  # noqa: S608
        conn.execute(f"DELETE FROM jobs WHERE id IN ({where})", ids)  # noqa: S608
    return len(ids)
