"""The audit rows since a moment, newest first."""

import sqlite3
from typing import Any


def read_audit(conn: sqlite3.Connection, since: float) -> list[dict[str, Any]]:
    """``{time, action, job_id, class, principal}`` per row."""
    rows = conn.execute(
        "SELECT created, action, job_id, privacy, principal FROM audit"
        " WHERE created>? ORDER BY seq DESC",
        (since,),
    ).fetchall()
    return [
        {
            "time": r["created"],
            "action": r["action"],
            "job_id": r["job_id"],
            "class": r["privacy"],
            "principal": r["principal"],
        }
        for r in rows
    ]
