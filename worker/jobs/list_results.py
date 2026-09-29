"""The producer's unacknowledged results, in completion order (spec 6)."""

import sqlite3
from typing import Any

from worker.jobs.result_row_to_dict import result_row_to_dict


def list_results(
    conn: sqlite3.Connection, producer: str, queue: str, after: int, limit: int
) -> list[dict[str, Any]]:
    """``after`` is the last ``seq`` the producer has seen; ``limit`` is clamped to 1-100.

    An output row whose content is gone is never listed as a bare success.
    """
    rows = conn.execute(
        "SELECT r.*, o.body AS output FROM results r LEFT JOIN p.outputs o ON o.result_id=r.result_id"
        " WHERE r.producer=? AND r.queue=? AND r.seq>? AND r.acked IS NULL"
        " AND (r.control IS NOT NULL OR o.result_id IS NOT NULL) ORDER BY r.seq LIMIT ?",
        (producer, queue, after, max(1, min(limit, 100))),
    ).fetchall()
    return [result_row_to_dict(r) for r in rows]
