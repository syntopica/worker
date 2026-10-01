"""Store one member of a shadow group: a copy of the original job, pinned."""

import json
import sqlite3
import uuid
from typing import Any


def insert_shadow_job(  # noqa: PLR0913
    conn: sqlite3.Connection,
    original: sqlite3.Row,
    job_input: dict[str, Any],
    *,
    group: str,
    pin: str,
    model: str,
    kind: str = "inference",
    producer: str = "_shadow",
    max_attempts: int = 1,
    now: float,
) -> str:
    """Return the member's id; priority 0 and one attempt by default, never crowding real work.

    ``original`` supplies queue, class and tier; ``group`` is the shadowed job's id.
    """
    job_id = uuid.uuid4().hex
    conn.execute(
        "INSERT INTO jobs (id, producer, queue, kind, idempotency_key, payload_hash, priority, privacy,"
        " model, state, max_attempts, not_before, created, updated, tier, shadow_of, pin)"
        " VALUES (?, ?, ?, ?, ?, '', 0, ?, ?, 'queued', ?, ?, ?, ?, ?, ?, ?)",
        (
            job_id,
            producer,
            original["queue"],
            kind,
            f"{pin}:{group}",
            original["privacy"],
            model,
            max_attempts,
            now,
            now,
            now,
            original["tier"],
            group,
            pin,
        ),
    )
    conn.execute(
        "INSERT INTO p.inputs (job_id, body) VALUES (?, ?)",
        (job_id, json.dumps({"input": job_input})),
    )
    return job_id
