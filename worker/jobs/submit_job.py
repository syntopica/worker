"""Admit one job: validate, deduplicate, bound, and store it (spec 6)."""

import sqlite3
import uuid

from worker.config.worker_config import WorkerConfig
from worker.jobs.admit_split_child import admit_split_child
from worker.jobs.api_error import ApiError
from worker.jobs.check_outstanding import check_outstanding
from worker.jobs.check_queue_grant import check_queue_grant
from worker.jobs.decode_body import decode_body
from worker.jobs.insert_job import insert_job
from worker.jobs.parse_submit_request import parse_submit_request
from worker.jobs.payload_hash import payload_hash
from worker.jobs.resolve_job_model import resolve_job_model
from worker.store.transaction import transaction


def submit_job(
    conn: sqlite3.Connection, config: WorkerConfig, producer: str, raw: bytes, now: float
) -> tuple[str, bool]:
    """Return ``(job_id, created)``; raise ApiError for every refusal."""
    if len(raw) > config.max_payload_bytes:
        raise ApiError(413, "payload_too_large")
    body = decode_body(raw)
    req = parse_submit_request(body)
    check_queue_grant(config, producer, req.queue)
    model = resolve_job_model(config, req)
    digest = payload_hash(body)
    with transaction(conn):
        existing = conn.execute(
            "SELECT id, payload_hash FROM jobs WHERE producer=? AND queue=? AND idempotency_key=?",
            (producer, req.queue, req.idempotency_key),
        ).fetchone()
        if existing is not None:
            if existing["payload_hash"] != digest:
                raise ApiError(409, "idempotency_conflict")
            return existing["id"], False
        if req.parent_id is not None:
            admit_split_child(conn, config, producer, req, now)
        else:
            check_outstanding(conn, config, producer, req.queue)
        job_id = uuid.uuid4().hex
        insert_job(conn, job_id, producer, req, digest, model, now)
    return job_id, True
