"""Refuse any report from an attempt that no longer owns its job (spec 6)."""

import sqlite3

from worker.jobs.api_error import ApiError
from worker.jobs.states import LIVE


def check_fence(conn: sqlite3.Connection, attempt_id: str, generation: int) -> sqlite3.Row:
    """Return the job row, or raise ApiError(409, "stale_attempt")."""
    job: sqlite3.Row | None = conn.execute(
        "SELECT * FROM jobs WHERE lease_attempt=?", (attempt_id,)
    ).fetchone()
    if job is None or job["generation"] != generation or job["state"] not in LIVE:
        raise ApiError(409, "stale_attempt")
    return job
