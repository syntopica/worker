"""Delete the payloads that retention says are due, each job exactly once (spec 9)."""

import sqlite3

from worker.config.queue_policy import QueuePolicy
from worker.jobs.delete_payloads import delete_payloads
from worker.jobs.states import SENSITIVE

_INSPECTION_S = 24 * 3600.0


def delete_payloads_batch(
    conn: sqlite3.Connection, policy: QueuePolicy, now: float, limit: int
) -> int:
    """Sensitive non-successes after the inspection window; everything after retention.

    An unacknowledged success is left to ``expire_unacked_batch``, which tells
    the producer. Return how many jobs lost their payloads.
    """
    inspected = now - _INSPECTION_S
    kept = now - policy.retention_days * 86400
    marks = ",".join("?" * len(SENSITIVE))
    rows = conn.execute(
        "SELECT id FROM jobs WHERE payloads_deleted=0 AND queue=? AND finished<=?"  # noqa: S608
        " AND NOT (state='succeeded' AND acked IS NULL)"
        f" AND ((privacy IN ({marks}) AND state!='succeeded' AND finished<=?)"
        " OR coalesce(acked, finished)<=?) LIMIT ?",
        (policy.name, max(inspected, kept), *SENSITIVE, inspected, kept, limit),
    ).fetchall()
    for (job_id,) in rows:
        delete_payloads(conn, job_id)
    return len(rows)
