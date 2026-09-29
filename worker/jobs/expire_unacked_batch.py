"""Expire succeeded results that were never acknowledged in time (spec 9)."""

import sqlite3

from worker.config.queue_policy import QueuePolicy
from worker.jobs.add_control_result import add_control_result
from worker.jobs.delete_payloads import delete_payloads
from worker.jobs.states import SENSITIVE


def expire_unacked_batch(
    conn: sqlite3.Connection, policy: QueuePolicy, now: float, limit: int
) -> int:
    """Sensitive jobs after ``unacked_ttl``, the others at their retention deadline.

    Each becomes ``unacked_expired`` with a control result, and loses its payloads.
    Return how many jobs were expired.
    """
    marks = ",".join("?" * len(SENSITIVE))
    rows = conn.execute(
        "SELECT * FROM jobs WHERE state='succeeded' AND acked IS NULL AND queue=?"  # noqa: S608
        f" AND ((privacy IN ({marks}) AND finished<=?) OR (privacy NOT IN ({marks}) AND finished<=?))"
        " LIMIT ?",
        (
            policy.name,
            *SENSITIVE,
            now - policy.unacked_ttl_hours * 3600,
            *SENSITIVE,
            now - policy.retention_days * 86400,
            limit,
        ),
    ).fetchall()
    for job in rows:
        conn.execute(
            "UPDATE jobs SET state='unacked_expired', updated=? WHERE id=?", (now, job["id"])
        )
        conn.execute(
            "UPDATE results SET acked=? WHERE job_id=? AND acked IS NULL", (now, job["id"])
        )
        add_control_result(conn, job, "unacked_expired", {}, now)
        delete_payloads(conn, job["id"])
    return len(rows)
