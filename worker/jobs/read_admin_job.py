"""One job's admin detail: metadata, attempts and which payloads are still stored."""

import sqlite3
from typing import Any

from worker.jobs.admin_job_row import admin_job_row
from worker.jobs.load_admin_job import load_admin_job
from worker.jobs.read_job_attempts import read_job_attempts


def read_admin_job(conn: sqlite3.Connection, job_id: str) -> dict[str, Any]:
    """Raise ApiError(404) for an unknown id; never reads a payload body."""
    detail = admin_job_row(load_admin_job(conn, job_id))
    detail["attempt_details"] = read_job_attempts(conn, job_id)
    detail["has_input"] = (
        conn.execute("SELECT 1 FROM p.inputs WHERE job_id=?", (job_id,)).fetchone() is not None
    )
    detail["has_output"] = (
        conn.execute(
            "SELECT 1 FROM results r JOIN p.outputs o ON o.result_id=r.result_id WHERE r.job_id=?",
            (job_id,),
        ).fetchone()
        is not None
    )
    return detail
