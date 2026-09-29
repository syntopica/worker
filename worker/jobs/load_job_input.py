"""Read a job's stored input from the payload file."""

import json
import sqlite3
from typing import Any


def load_job_input(conn: sqlite3.Connection, job_id: str) -> dict[str, Any] | None:
    """The ``input`` object, or None when the payload row is missing."""
    row = conn.execute("SELECT body FROM p.inputs WHERE job_id=?", (job_id,)).fetchone()
    if row is None:
        return None
    job_input: dict[str, Any] = json.loads(row[0])["input"]
    return job_input
