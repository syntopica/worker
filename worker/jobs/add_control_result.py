"""Append a control result to the job's results feed (spec 6)."""

import json
import sqlite3
import uuid
from typing import Any


def add_control_result(
    conn: sqlite3.Connection, job: sqlite3.Row, control: str, detail: dict[str, Any], now: float
) -> str:
    """Return the new result_id. ``detail`` must hold allowlisted fields only."""
    result_id = uuid.uuid4().hex
    conn.execute(
        "INSERT INTO results (result_id, job_id, producer, queue, control, detail, created) VALUES (?, ?, ?, ?, ?, ?, ?)",
        (result_id, job["id"], job["producer"], job["queue"], control, json.dumps(detail), now),
    )
    return result_id
