"""Jobs of every producer, newest first, for the admin browser (amendment 2026-10-04)."""

import sqlite3
from typing import Any

from worker.jobs.admin_job_row import admin_job_row
from worker.jobs.admin_job_select import ADMIN_JOB_SELECT
from worker.jobs.decode_jobs_cursor import decode_jobs_cursor
from worker.jobs.encode_jobs_cursor import encode_jobs_cursor
from worker.jobs.jobs_filter import JobsFilter


def list_admin_jobs(conn: sqlite3.Connection, wanted: JobsFilter) -> dict[str, Any]:
    """``{"jobs": [...], "next": cursor}``; ``next`` is null once a page comes back short."""
    clauses: list[str] = []
    params: list[Any] = []
    for column, value in (
        ("queue", wanted.queue),
        ("state", wanted.state),
        ("producer", wanted.producer),
    ):
        if value is not None:
            clauses.append(f"j.{column}=?")
            params.append(value)
    if wanted.before is not None:
        clauses.append("(j.created, j.id) < (?, ?)")
        params.extend(decode_jobs_cursor(wanted.before))
    where = f" WHERE {' AND '.join(clauses)}" if clauses else ""
    rows = conn.execute(
        f"{ADMIN_JOB_SELECT}{where} ORDER BY j.created DESC, j.id DESC LIMIT ?",
        (*params, wanted.limit),
    ).fetchall()
    full = len(rows) == wanted.limit
    return {
        "jobs": [admin_job_row(r) for r in rows],
        "next": encode_jobs_cursor(rows[-1]) if full and rows else None,
    }
