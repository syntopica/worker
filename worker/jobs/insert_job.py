"""Store one admitted job and its input."""

import json
import sqlite3

from worker.jobs.submit_request import SubmitRequest


def insert_job(  # noqa: PLR0913, PLR0917
    conn: sqlite3.Connection,
    job_id: str,
    producer: str,
    req: SubmitRequest,
    digest: str,
    model: str,
    now: float,
    retry_of: str | None = None,
) -> None:
    """Run inside the caller's transaction; the input goes to the payload file."""
    conn.execute(
        "INSERT INTO jobs (id, producer, queue, kind, idempotency_key, payload_hash, priority, privacy, model,"
        " state, max_attempts, not_before, deadline, parent_id, created, updated, tier, retry_of)"
        " VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'queued', ?, ?, ?, ?, ?, ?, ?, ?)",
        (
            job_id,
            producer,
            req.queue,
            req.kind,
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
            req.tier,
            retry_of,
        ),
    )
    conn.execute(
        "INSERT INTO p.inputs (job_id, body) VALUES (?, ?)",
        (job_id, json.dumps({"input": req.input})),
    )
