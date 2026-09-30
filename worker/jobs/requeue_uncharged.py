"""Hand a job back after an ending that says nothing about the job."""

import sqlite3


def requeue_uncharged(conn: sqlite3.Connection, job: sqlite3.Row, code: str, now: float) -> str:
    """Return ``queued``; neither an attempt nor a preemption is charged and the job is due at once.

    A free endpoint's 429 (``rate_limited``, amendment phase 2b) or the node
    stopping under it (``node_shutdown``) is no evidence against the job.
    """
    conn.execute(
        "UPDATE jobs SET state='queued', error=?, not_before=?, updated=?,"
        " lease_attempt=NULL, lease_expires=NULL WHERE id=?",
        (code, now, now, job["id"]),
    )
    return "queued"
