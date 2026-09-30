"""Settle an attempt that did not succeed: preempt, re-plan or charge it."""

import sqlite3

from worker.config.worker_config import WorkerConfig
from worker.jobs.completion_report import CompletionReport
from worker.jobs.fail_attempt import fail_attempt
from worker.jobs.preempt_job import preempt_job
from worker.jobs.requeue_after_wall import requeue_after_wall
from worker.jobs.requeue_rate_limited import requeue_rate_limited


def settle_unsuccessful(
    conn: sqlite3.Connection,
    config: WorkerConfig,
    job: sqlite3.Row,
    report: CompletionReport,
    now: float,
) -> str:
    """Return the job's new state.

    A quota wall or a remote rate limit says nothing about the job, so it is
    re-planned without charging an attempt; any other failure is charged.
    """
    if report.outcome == "preempted":
        return preempt_job(conn, config, job, now)
    if report.error_code == "quota_wall" and job["kind"] == "task":
        return requeue_after_wall(conn, config, job, now)
    if report.error_code == "rate_limited":
        return requeue_rate_limited(conn, job, now)
    return fail_attempt(conn, job, report.error_code or "executor_error", now)
