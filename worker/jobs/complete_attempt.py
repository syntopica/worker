"""Settle one attempt: result, retry, preemption, split request or park (spec 6-7)."""

import dataclasses
import json
import sqlite3
import sys
import uuid

from worker.config.worker_config import WorkerConfig
from worker.jobs.check_fence import check_fence
from worker.jobs.completion_report import CompletionReport
from worker.jobs.fail_attempt import fail_attempt
from worker.jobs.fail_payload_lost import fail_payload_lost
from worker.jobs.known_error_code import known_error_code
from worker.jobs.load_job_input import load_job_input
from worker.jobs.open_shadow_group import open_shadow_group
from worker.jobs.output_matches_schema import output_matches_schema
from worker.jobs.schema_violation_path import schema_violation_path
from worker.jobs.settle_unsuccessful import settle_unsuccessful
from worker.store.transaction import transaction


def complete_attempt(  # noqa: PLR0913, PLR0917
    conn: sqlite3.Connection,
    config: WorkerConfig,
    attempt_id: str,
    generation: int,
    report: CompletionReport,
    now: float,
    *,
    node: str | None = None,
) -> str:
    """Return the job's new state; raise ApiError(409) for a fenced-out attempt.

    An unknown error code settles as ``executor_error`` after the fence check,
    so a valid completion is never lost to it and its text is never stored.
    """
    with transaction(conn):
        job = check_fence(conn, attempt_id, generation, node)
        code = known_error_code(report.error_code)
        if code != report.error_code:
            print("worker: unknown error code replaced", file=sys.stderr)
            report = dataclasses.replace(report, error_code=code)
        conn.execute(
            "UPDATE attempts SET ended=?, outcome=?, error=?, wall_s=?, tokens_in=?, tokens_out=?,"
            " provider=?, cost_usd=?, model=? WHERE id=?",
            (
                now,
                report.outcome,
                report.error_code,
                report.wall_s,
                report.usage.get("tokens_in"),
                report.usage.get("tokens_out"),
                report.executor.get("provider"),
                report.usage.get("cost_usd"),
                report.executor.get("model") or None,
                attempt_id,
            ),
        )
        if report.outcome != "succeeded":
            return settle_unsuccessful(conn, config, job, report, now)
        job_input = load_job_input(conn, job["id"])
        if job_input is None:
            fail_payload_lost(conn, job["id"], now)
            return "failed"
        if not output_matches_schema(job_input, report.output):
            conn.execute(
                "UPDATE attempts SET outcome='schema_violation', error='schema_violation' WHERE id=?",
                (attempt_id,),
            )
            path = schema_violation_path(job_input, report.output)
            extra = {"schema_path": path} if path else None
            return fail_attempt(conn, job, "schema_violation", now, extra)
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
        open_shadow_group(conn, config, job, job_input, report, now)
        return "succeeded"
