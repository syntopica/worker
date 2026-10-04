"""Acknowledge a job and every result it still has open, as a producer ack would."""

import sqlite3

from worker.jobs.delete_payloads import delete_payloads
from worker.jobs.states import SENSITIVE


def acknowledge_job(conn: sqlite3.Connection, job_id: str, privacy: str, now: float) -> None:
    """Run inside the caller's transaction; a sensitive job's payloads are deleted (spec 9)."""
    conn.execute("UPDATE results SET acked=? WHERE job_id=? AND acked IS NULL", (now, job_id))
    conn.execute(
        "UPDATE jobs SET acked=?, updated=? WHERE id=? AND acked IS NULL", (now, now, job_id)
    )
    if privacy in SENSITIVE:
        delete_payloads(conn, job_id)
