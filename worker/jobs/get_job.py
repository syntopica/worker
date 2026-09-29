"""One job's state and its latest unacknowledged result, for its own producer."""

import sqlite3
from typing import Any

from worker.jobs.api_error import ApiError
from worker.jobs.result_row_to_dict import result_row_to_dict


def get_job(conn: sqlite3.Connection, producer: str, job_id: str) -> dict[str, Any]:
    """Raise ApiError(404) for a missing job or one owned by another producer."""
    job = conn.execute(
        "SELECT id, producer, state, error FROM jobs WHERE id=?", (job_id,)
    ).fetchone()
    if job is None or job["producer"] != producer:
        raise ApiError(404, "not_found")
    latest = conn.execute(
        "SELECT r.*, o.body AS output FROM results r LEFT JOIN p.outputs o ON o.result_id=r.result_id"
        " WHERE r.job_id=? AND r.producer=? AND r.acked IS NULL ORDER BY r.seq DESC LIMIT 1",
        (job_id, producer),
    ).fetchone()
    result = result_row_to_dict(latest) if latest else None
    return {"id": job["id"], "state": job["state"], "error": job["error"], "result": result}
