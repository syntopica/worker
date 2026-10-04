"""One job's admin metadata row, whoever its producer."""

import sqlite3

from worker.jobs.admin_job_select import ADMIN_JOB_SELECT
from worker.jobs.api_error import ApiError


def load_admin_job(conn: sqlite3.Connection, job_id: str) -> sqlite3.Row:
    """Raise ApiError(404) for an unknown id."""
    row: sqlite3.Row | None = conn.execute(
        f"{ADMIN_JOB_SELECT} WHERE j.id=?",
        (job_id,),
    ).fetchone()
    if row is None:
        raise ApiError(404, "not_found")
    return row
