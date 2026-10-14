"""Charge a failed attempt to the job: retry with backoff or fail for good."""

import sqlite3
from typing import Any

from worker.jobs.finish_failed import finish_failed
from worker.jobs.retry_backoff import retry_backoff

# A runner that reports its service unavailable was seen failing for 5-10
# minutes at a time (2026-10-11); a retry 30 s later lands in the same outage.
_UNAVAILABLE_BACKOFF_S = 600.0


def fail_attempt(
    conn: sqlite3.Connection,
    job: sqlite3.Row,
    code: str,
    now: float,
    extra: dict[str, Any] | None = None,
) -> str:
    """Return the job's new state; ``extra`` reaches the final control result only.

    A ``runner_unavailable`` retry waits at least ten minutes, past the outage.
    """
    attempts = job["attempts"] + 1
    conn.execute("UPDATE jobs SET attempts=? WHERE id=?", (attempts, job["id"]))
    if attempts >= job["max_attempts"]:
        finish_failed(conn, job, code, now, extra=extra)
        return "failed"
    delay = retry_backoff(attempts)
    if code == "runner_unavailable":
        delay = max(delay, _UNAVAILABLE_BACKOFF_S)
    conn.execute(
        "UPDATE jobs SET state='queued', error=?, not_before=?, updated=?,"
        " lease_attempt=NULL, lease_expires=NULL WHERE id=?",
        (code, now + delay, now, job["id"]),
    )
    return "queued"
