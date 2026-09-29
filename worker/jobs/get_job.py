"""One job's state and its latest unacknowledged result, for its own producer."""

import sqlite3
from typing import Any

from worker.jobs.api_error import ApiError
from worker.jobs.list_results import list_results


def get_job(conn: sqlite3.Connection, producer: str, job_id: str) -> dict[str, Any]:
    """Raise ApiError(404) for a missing job or one owned by another producer."""
    job = conn.execute(
        "SELECT id, producer, queue, state, error FROM jobs WHERE id=?", (job_id,)
    ).fetchone()
    if job is None or job["producer"] != producer:
        raise ApiError(404, "not_found")
    mine = [r for r in list_results(conn, producer, job["queue"], 0, 100) if r["job_id"] == job_id]
    return {
        "id": job["id"],
        "state": job["state"],
        "error": job["error"],
        "result": mine[-1] if mine else None,
    }
