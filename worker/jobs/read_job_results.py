"""A job's results, oldest first, as content-free metadata for the admin detail."""

import json
import sqlite3
from typing import Any


def read_job_results(conn: sqlite3.Connection, job_id: str) -> list[dict[str, Any]]:
    """Control code, allowlisted detail, executor, usage, rating and times; never the output body."""
    rows = conn.execute(
        "SELECT result_id, control, detail, executor, usage, rating, created, acked"
        " FROM results WHERE job_id=? ORDER BY seq",
        (job_id,),
    ).fetchall()
    return [
        {
            "result_id": r["result_id"],
            "control": r["control"],
            "detail": json.loads(r["detail"]) if r["detail"] else None,
            "executor": json.loads(r["executor"]) if r["executor"] else None,
            "usage": json.loads(r["usage"]) if r["usage"] else None,
            "rating": r["rating"],
            "created": r["created"],
            "acked": r["acked"],
        }
        for r in rows
    ]
