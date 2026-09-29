"""Settle one attempt: result, retry, preemption, split request or park (spec 6-7)."""

import json
import sqlite3
import uuid

from worker.config.worker_config import WorkerConfig
from worker.jobs.check_fence import check_fence
from worker.jobs.completion_report import CompletionReport
from worker.jobs.fail_attempt import fail_attempt
from worker.jobs.fail_payload_lost import fail_payload_lost
from worker.jobs.load_job_input import load_job_input
from worker.jobs.output_matches_schema import output_matches_schema
from worker.jobs.preempt_job import preempt_job
from worker.store.transaction import transaction


def complete_attempt(  # noqa: PLR0913, PLR0917
    conn: sqlite3.Connection,
    config: WorkerConfig,
    attempt_id: str,
    generation: int,
    report: CompletionReport,
    now: float,
) -> str:
    """Return the job's new state; raise ApiError(409) for a fenced-out attempt."""
    with transaction(conn):
        job = check_fence(conn, attempt_id, generation)
        conn.execute(
            "UPDATE attempts SET ended=?, outcome=?, error=?, wall_s=?, tokens_in=?, tokens_out=? WHERE id=?",
            (
                now,
                report.outcome,
                report.error_code,
                report.wall_s,
                report.usage.get("tokens_in"),
                report.usage.get("tokens_out"),
                attempt_id,
            ),
        )
        if report.outcome == "preempted":
            return preempt_job(conn, config, job, now)
        if report.outcome != "succeeded":
            return fail_attempt(conn, job, report.error_code or "executor_error", now)
        job_input = load_job_input(conn, job["id"])
        if job_input is None:
            fail_payload_lost(conn, job["id"], now)
            return "failed"
        if not output_matches_schema(job_input, report.output):
            conn.execute(
                "UPDATE attempts SET outcome='schema_violation', error='schema_violation' WHERE id=?",
                (attempt_id,),
            )
            return fail_attempt(conn, job, "schema_violation", now)
        result_id = uuid.uuid4().hex
        conn.execute(
            "INSERT INTO results (result_id, job_id, producer, queue, executor, usage, created) VALUES (?, ?, ?, ?, ?, ?, ?)",
            (
                result_id,
                job["id"],
                job["producer"],
                job["queue"],
                json.dumps(report.executor),
                json.dumps(report.usage),
                now,
            ),
        )
        conn.execute(
            "INSERT INTO p.outputs (result_id, body) VALUES (?, ?)",
            (result_id, json.dumps(report.output)),
        )
        conn.execute(
            "UPDATE jobs SET state='succeeded', error=NULL, finished=?, updated=?,"
            " lease_attempt=NULL, lease_expires=NULL WHERE id=?",
            (now, now, job["id"]),
        )
        return "succeeded"
