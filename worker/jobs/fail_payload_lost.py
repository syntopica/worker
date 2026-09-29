"""Fail a job whose input is gone from the payload file (spec 9 restore rule)."""

import sqlite3

from worker.jobs.delete_payloads import delete_payloads
from worker.jobs.finish_failed import finish_failed


def fail_payload_lost(conn: sqlite3.Connection, job_id: str, now: float) -> None:
    """``failed`` with ``payload_lost``; the producer resubmits it."""
    job = conn.execute("SELECT * FROM jobs WHERE id=?", (job_id,)).fetchone()
    finish_failed(conn, job, "payload_lost", now)
    delete_payloads(conn, job_id)
