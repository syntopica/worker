"""Apply the retention table of spec 9 to every job."""

import sqlite3

from worker.config.worker_config import WorkerConfig
from worker.jobs.add_control_result import add_control_result
from worker.jobs.delete_payloads import delete_payloads
from worker.jobs.states import SENSITIVE
from worker.store.transaction import transaction

_INSPECTION_S = 24 * 3600.0


def sweep_retention(conn: sqlite3.Connection, config: WorkerConfig, now: float) -> int:
    """Return how many jobs had their payloads deleted."""
    swept = 0
    marks = ",".join("?" * len(SENSITIVE))
    with transaction(conn):
        for job in conn.execute(
            f"SELECT * FROM jobs WHERE state='succeeded' AND acked IS NULL AND privacy IN ({marks})",  # noqa: S608
            SENSITIVE,
        ).fetchall():
            if now - job["finished"] >= config.queues[job["queue"]].unacked_ttl_hours * 3600:
                conn.execute(
                    "UPDATE jobs SET state='unacked_expired', updated=? WHERE id=?",
                    (now, job["id"]),
                )
                conn.execute(
                    "UPDATE results SET acked=? WHERE job_id=? AND acked IS NULL", (now, job["id"])
                )
                add_control_result(conn, job, "unacked_expired", {}, now)
                delete_payloads(conn, job["id"])
                swept += 1
        for job in conn.execute(
            "SELECT id, queue, privacy, state, finished, acked FROM jobs WHERE finished IS NOT NULL"
        ).fetchall():
            ended = job["acked"] or job["finished"]
            sensitive_done = (
                job["privacy"] in SENSITIVE
                and job["state"] != "succeeded"
                and now - job["finished"] >= _INSPECTION_S
            )
            retained_out = (
                job["acked"] is not None
                and now - ended >= config.queues[job["queue"]].retention_days * 86400
            )
            if sensitive_done or retained_out:
                delete_payloads(conn, job["id"])
                swept += 1
    return swept
