"""Settle an attempt that did not succeed: preempt, re-plan or charge it."""

import sqlite3

from worker.config.worker_config import WorkerConfig
from worker.jobs.completion_report import CompletionReport
from worker.jobs.fail_attempt import fail_attempt
from worker.jobs.preempt_job import preempt_job
from worker.jobs.requeue_after_wall import requeue_after_wall
from worker.jobs.requeue_uncharged import requeue_uncharged

_UNCHARGED = ("rate_limited", "node_shutdown")


def settle_unsuccessful(
    conn: sqlite3.Connection,
    config: WorkerConfig,
    job: sqlite3.Row,
    report: CompletionReport,
    now: float,
) -> str:
    """Return the job's new state.

    A quota wall, a remote rate limit or the node shutting down says nothing
    about the job, so it is re-planned without charging an attempt (or, for a
    shutdown, a preemption); any other failure is charged. A runner's empty
    answer to an inference job is not charged either: the runner rung skips
    the job afterwards, so the next executor answers it.
    """
    if report.error_code in _UNCHARGED or (
        report.error_code == "no_output" and job["kind"] == "inference"
    ):
        return requeue_uncharged(conn, job, report.error_code, now)
    if report.outcome == "preempted":
        return preempt_job(conn, config, job, now)
    if report.error_code == "quota_wall":
        return requeue_after_wall(conn, config, job, report.executor, now)
    return fail_attempt(conn, job, report.error_code or "executor_error", now)
