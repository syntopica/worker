"""Acknowledge a job's open control results on its producer's behalf."""

import sqlite3

from worker.config.worker_config import WorkerConfig
from worker.jobs.ack_result import ack_result
from worker.jobs.api_error import ApiError


def admin_ack_job(
    conn: sqlite3.Connection, config: WorkerConfig, job: sqlite3.Row, now: float
) -> None:
    """Each result is acked as the producer would, without decline or rating.

    Nothing open, or an open output, is ``409 not_ackable``: an output is the producer's.
    """
    rows = conn.execute(
        "SELECT result_id, control FROM results WHERE job_id=? AND acked IS NULL ORDER BY seq",
        (job["id"],),
    ).fetchall()
    if not rows or any(r["control"] is None for r in rows):
        raise ApiError(409, "not_ackable")
    for r in rows:
        ack_result(conn, config, job["producer"], job["id"], r["result_id"], False, now)
