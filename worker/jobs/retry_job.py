"""Admit a finished job's copy and acknowledge the original (amendment 2026-10-04)."""

import sqlite3
import uuid

from worker.config.worker_config import WorkerConfig
from worker.jobs.acknowledge_job import acknowledge_job
from worker.jobs.api_error import ApiError
from worker.jobs.check_outstanding import check_outstanding
from worker.jobs.check_queue_grant import check_queue_grant
from worker.jobs.insert_job import insert_job
from worker.jobs.load_job_input import load_job_input
from worker.jobs.parse_submit_request import parse_submit_request
from worker.jobs.payload_hash import payload_hash
from worker.jobs.record_audit import record_audit
from worker.jobs.resolve_job_model import resolve_job_model
from worker.jobs.retry_submit_body import retry_submit_body
from worker.jobs.states import RETRYABLE
from worker.store.transaction import transaction


def retry_job(
    conn: sqlite3.Connection, config: WorkerConfig, job_id: str, principal: str, now: float
) -> tuple[str, bool]:
    """Return ``(new_id, created)``; every refusal is an ApiError and changes nothing.

    The copy is admitted as a submit: queue grant, task grant, privacy ceiling
    and outstanding limit. The original is acknowledged first, in the same
    transaction, so it no longer counts against that limit.
    """
    job = conn.execute("SELECT * FROM jobs WHERE id=?", (job_id,)).fetchone()
    if job is None:
        raise ApiError(404, "not_found")
    if job["state"] not in RETRYABLE:
        raise ApiError(409, "not_retryable")
    job_input = load_job_input(conn, job_id)
    if job_input is None:
        raise ApiError(410, "content_gone")
    body = retry_submit_body(job, job_input)
    req = parse_submit_request(body)
    check_queue_grant(config, job["producer"], req.queue)
    model = resolve_job_model(config, req)
    with transaction(conn):
        existing = conn.execute(
            "SELECT id FROM jobs WHERE producer=? AND queue=? AND idempotency_key=?",
            (job["producer"], req.queue, req.idempotency_key),
        ).fetchone()
        if existing is not None:
            return str(existing["id"]), False
        acknowledge_job(conn, job_id, job["privacy"], now)
        check_outstanding(conn, config, job["producer"], req.queue)
        new_id = uuid.uuid4().hex
        insert_job(conn, new_id, job["producer"], req, payload_hash(body), model, now, job_id)
        record_audit(conn, "retry", job_id, job["privacy"], principal, now)
    return new_id, True
