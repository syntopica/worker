"""The producer's unacknowledged results, in completion order (spec 6)."""

import json
import sqlite3
from typing import Any


def list_results(
    conn: sqlite3.Connection, producer: str, queue: str, after: int, limit: int
) -> list[dict[str, Any]]:
    """``after`` is the last ``seq`` the producer has seen."""
    rows = conn.execute(
        "SELECT r.*, o.body AS output FROM results r LEFT JOIN p.outputs o ON o.result_id=r.result_id"
        " WHERE r.producer=? AND r.queue=? AND r.seq>? AND r.acked IS NULL ORDER BY r.seq LIMIT ?",
        (producer, queue, after, min(limit, 100)),
    ).fetchall()
    return [
        {
            "seq": r["seq"],
            "result_id": r["result_id"],
            "job_id": r["job_id"],
            "control": r["control"],
            "detail": json.loads(r["detail"]) if r["detail"] else None,
            "output": json.loads(r["output"]) if r["output"] else None,
            "executor": json.loads(r["executor"]) if r["executor"] else None,
            "usage": json.loads(r["usage"]) if r["usage"] else None,
        }
        for r in rows
    ]
