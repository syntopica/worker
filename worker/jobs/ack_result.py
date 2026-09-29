"""Acknowledge a result; a declined split parks the job instead (spec 6-7)."""

import sqlite3

from worker.config.worker_config import WorkerConfig
from worker.jobs.api_error import ApiError
from worker.jobs.delete_payloads import delete_payloads
from worker.jobs.queue_p90_runtime import queue_p90_runtime
from worker.jobs.states import CONTROL_TERMINAL, SENSITIVE
from worker.store.transaction import transaction


def ack_result(  # noqa: PLR0913, PLR0917
    conn: sqlite3.Connection,
    config: WorkerConfig,
    producer: str,
    job_id: str,
    result_id: str,
    decline: bool,
    now: float,
) -> None:
    """Record the ack; raise ApiError(404) when the result is not this producer's."""
    with transaction(conn):
        result = conn.execute(
            "SELECT control FROM results WHERE result_id=? AND job_id=? AND producer=?",
            (result_id, job_id, producer),
        ).fetchone()
        if result is None:
            raise ApiError(404, "not_found")
        job = conn.execute(
            "SELECT queue, privacy, state FROM jobs WHERE id=?", (job_id,)
        ).fetchone()
        conn.execute("UPDATE results SET acked=? WHERE result_id=?", (now, result_id))
        if result["control"] == "split_requested":
            if decline and job["state"] == "split_requested":
                p90 = queue_p90_runtime(conn, job["queue"]) or 0.0
                need = max(1.5 * p90, config.queues[job["queue"]].parked_min_idle_s)
                conn.execute(
                    "UPDATE jobs SET state='queued', parked=1, parked_min_idle_s=?, updated=? WHERE id=?",
                    (need, now, job_id),
                )
            return
        if result["control"] is None or result["control"] in CONTROL_TERMINAL:
            conn.execute("UPDATE jobs SET acked=?, updated=? WHERE id=?", (now, now, job_id))
            if job["privacy"] in SENSITIVE:
                delete_payloads(conn, job_id)
