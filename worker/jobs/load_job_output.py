"""Read a job's latest stored output from the payload file."""

import json
import sqlite3
from typing import Any


def load_job_output(conn: sqlite3.Connection, job_id: str) -> Any:
    """The newest result's output body, or None when no output row is left."""
    row = conn.execute(
        "SELECT o.body FROM results r JOIN p.outputs o ON o.result_id=r.result_id"
        " WHERE r.job_id=? ORDER BY r.seq DESC LIMIT 1",
        (job_id,),
    ).fetchone()
    return None if row is None else json.loads(row[0])
