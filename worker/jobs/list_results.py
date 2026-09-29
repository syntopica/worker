"""The producer's unacknowledged results, in completion order (spec 6)."""

import sqlite3
from typing import Any

from worker.jobs.result_row_to_dict import result_row_to_dict


def list_results(
    conn: sqlite3.Connection, producer: str, queue: str, after: int, limit: int
) -> list[dict[str, Any]]:
    """``after`` is the last ``seq`` the producer has seen."""
    rows = conn.execute(
        "SELECT r.*, o.body AS output FROM results r LEFT JOIN p.outputs o ON o.result_id=r.result_id"
        " WHERE r.producer=? AND r.queue=? AND r.seq>? AND r.acked IS NULL ORDER BY r.seq LIMIT ?",
        (producer, queue, after, min(limit, 100)),
    ).fetchall()
    return [result_row_to_dict(r) for r in rows]
