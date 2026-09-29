"""Cancel a job; a running attempt is fenced out at its next heartbeat."""

import sqlite3

from worker.jobs.api_error import ApiError
from worker.jobs.delete_payloads import delete_payloads
from worker.jobs.states import SENSITIVE
from worker.store.transaction import transaction

_CANCELLABLE = ("queued", "split_requested", "leased", "running", "draining")


def cancel_job(conn: sqlite3.Connection, producer: str, job_id: str, now: float) -> str:
    """Return the resulting state; terminal jobs are left as they are."""
    with transaction(conn):
        job = conn.execute(
            "SELECT producer, state, privacy FROM jobs WHERE id=?", (job_id,)
        ).fetchone()
        if job is None or job["producer"] != producer:
            raise ApiError(404, "not_found")
        if job["state"] not in _CANCELLABLE:
            return str(job["state"])
        conn.execute(
            "UPDATE jobs SET state='cancelled', acked=?, finished=?, updated=?, lease_attempt=NULL, lease_expires=NULL WHERE id=?",
            (now, now, now, job_id),
        )
        if job["privacy"] in SENSITIVE:
            delete_payloads(conn, job_id)
    return "cancelled"
