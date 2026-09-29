"""Admit one job: validate, deduplicate, bound, and store it (spec 6)."""

import json
import sqlite3
import uuid

from worker.config.worker_config import WorkerConfig
from worker.jobs.admit_split_child import admit_split_child
from worker.jobs.api_error import ApiError
from worker.jobs.check_outstanding import check_outstanding
from worker.jobs.parse_submit_request import parse_submit_request
from worker.jobs.payload_hash import payload_hash
from worker.store.transaction import transaction


def submit_job(
    conn: sqlite3.Connection, config: WorkerConfig, producer: str, raw: bytes, now: float
) -> tuple[str, bool]:
    """Return ``(job_id, created)``; raise ApiError for every refusal."""
    if len(raw) > config.max_payload_bytes:
        raise ApiError(413, "payload_too_large")
    body = json.loads(raw)
    req = parse_submit_request(body)
    if (
        req.queue not in config.producers.get(producer, frozenset())
        or req.queue not in config.queues
    ):
        raise ApiError(403, "queue_not_granted")
    model = next((m for m in req.models if m in config.models), None)
    if model is None:
        raise ApiError(400, "unknown_model")
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
        conn.execute(
            "INSERT INTO jobs (id, producer, queue, kind, idempotency_key, payload_hash, priority, privacy, model,"
            " state, max_attempts, not_before, deadline, parent_id, created, updated)"
            " VALUES (?, ?, ?, 'inference', ?, ?, ?, ?, ?, 'queued', ?, ?, ?, ?, ?, ?)",
            (
                job_id,
                producer,
                req.queue,
                req.idempotency_key,
                digest,
                req.priority,
                req.privacy,
                model,
                req.max_attempts,
                now,
                req.deadline,
                req.parent_id,
                now,
                now,
            ),
        )
        conn.execute(
            "INSERT INTO p.inputs (job_id, body) VALUES (?, ?)",
            (job_id, json.dumps({"input": req.input})),
        )
    return job_id, True
