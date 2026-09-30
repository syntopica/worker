"""Decline, on the producer's behalf, a split request it never answered."""

import sqlite3

from worker.config.queue_policy import QueuePolicy
from worker.jobs.park_job import park_job


def expire_split_requests_batch(
    conn: sqlite3.Connection, policy: QueuePolicy, now: float, limit: int
) -> int:
    """Park jobs left in ``split_requested`` longer than the queue's ``unacked_ttl_hours``.

    Without this a producer that never reads its control results would hold
    the job, and a slot of its outstanding limit, forever. Its control result
    is marked acknowledged, as a decline would. Return how many were parked.
    """
    rows = conn.execute(
        "SELECT id FROM jobs WHERE state='split_requested' AND queue=? AND updated<=? LIMIT ?",
        (policy.name, now - policy.unacked_ttl_hours * 3600, limit),
    ).fetchall()
    for row in rows:
        conn.execute(
            "UPDATE results SET acked=? WHERE job_id=? AND control='split_requested'"
            " AND acked IS NULL",
            (now, row["id"]),
        )
        park_job(conn, row["id"], policy.name, policy.parked_min_idle_s, now)
    return len(rows)
