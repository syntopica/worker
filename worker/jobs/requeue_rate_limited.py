"""Hand an escalated job back after a remote rate limit (amendment phase 2b)."""

import sqlite3


def requeue_rate_limited(conn: sqlite3.Connection, job: sqlite3.Row, now: float) -> str:
    """Return ``queued``; the attempt is not charged and the job is due at once.

    A free endpoint's 429 says nothing about the job, and the job's own model
    may be about to run it locally.
    """
    conn.execute(
        "UPDATE jobs SET state='queued', error='rate_limited', not_before=?, updated=?,"
        " lease_attempt=NULL, lease_expires=NULL WHERE id=?",
        (now, now, job["id"]),
    )
    return "queued"
