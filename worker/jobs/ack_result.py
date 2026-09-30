"""Acknowledge a result; a declined split parks the job instead (spec 6-7)."""

import sqlite3

from worker.config.worker_config import WorkerConfig
from worker.jobs.api_error import ApiError
from worker.jobs.delete_payloads import delete_payloads
from worker.jobs.park_job import park_job
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
    rating: str | None = None,
) -> None:
    """Record the ack and any rating; raise ApiError(404) when the result is not this producer's.

    A later ack may add or change the rating: the producer learns late that
    a result it kept needed editing.
    """
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
        conn.execute(
            "UPDATE results SET acked=? WHERE result_id=? AND acked IS NULL", (now, result_id)
        )
        if rating is not None:
            conn.execute("UPDATE results SET rating=? WHERE result_id=?", (rating, result_id))
        if result["control"] == "split_requested":
            if decline and job["state"] == "split_requested":
                # A queue removed from config still parks: the job waits for it.
                policy = config.queues.get(job["queue"])
                floor = policy.parked_min_idle_s if policy else 0.0
                park_job(conn, job_id, job["queue"], floor, now)
            return
        if result["control"] is None or result["control"] in CONTROL_TERMINAL:
            conn.execute(
                "UPDATE jobs SET acked=?, updated=? WHERE id=? AND acked IS NULL",
                (now, now, job_id),
            )
            if job["privacy"] in SENSITIVE:
                delete_payloads(conn, job_id)
