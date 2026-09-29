"""Settle a preempted attempt: requeue, park, ask for a split or exhaust."""

import sqlite3

from worker.config.worker_config import WorkerConfig
from worker.jobs.add_control_result import add_control_result
from worker.jobs.finish_failed import finish_failed


def preempt_job(
    conn: sqlite3.Connection, config: WorkerConfig, job: sqlite3.Row, now: float
) -> str:
    """Preemption spends no attempt. Return the job's new state."""
    if job["parked"]:
        runs = job["parked_runs"] + 1
        if runs >= config.max_parked_runs:
            finish_failed(conn, job, "preemption_exhausted", now)
            return "failed"
        conn.execute(
            "UPDATE jobs SET state='queued', parked_runs=?, parked_min_idle_s=parked_min_idle_s*2,"
            " updated=?, lease_attempt=NULL, lease_expires=NULL WHERE id=?",
            (runs, now, job["id"]),
        )
        return "queued"
    count = job["preemptions"] + 1
    state = "split_requested" if count >= config.split_after_preemptions else "queued"
    conn.execute(
        "UPDATE jobs SET state=?, preemptions=?, updated=?,"
        " lease_attempt=NULL, lease_expires=NULL WHERE id=?",
        (state, count, now, job["id"]),
    )
    if state == "split_requested":
        add_control_result(conn, job, "split_requested", {"preemptions": count}, now)
    return state
