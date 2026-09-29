"""Charge a failed attempt to the job: retry with backoff or fail for good."""

import sqlite3

from worker.jobs.finish_failed import finish_failed
from worker.jobs.retry_backoff import retry_backoff


def fail_attempt(conn: sqlite3.Connection, job: sqlite3.Row, code: str, now: float) -> str:
    """Return the job's new state."""
    attempts = job["attempts"] + 1
    conn.execute("UPDATE jobs SET attempts=? WHERE id=?", (attempts, job["id"]))
    if attempts >= job["max_attempts"]:
        finish_failed(conn, job, code, now)
        return "failed"
    conn.execute(
        "UPDATE jobs SET state='queued', error=?, not_before=?, updated=?,"
        " lease_attempt=NULL, lease_expires=NULL WHERE id=?",
        (code, now + retry_backoff(attempts), now, job["id"]),
    )
    return "queued"
