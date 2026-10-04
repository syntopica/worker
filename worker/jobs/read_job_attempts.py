"""A job's attempts, oldest first, as the admin detail shows them."""

import sqlite3
from typing import Any


def read_job_attempts(conn: sqlite3.Connection, job_id: str) -> list[dict[str, Any]]:
    """Executor, outcome, allowlisted error code, times, tokens, wall and cost; never output."""
    rows = conn.execute(
        "SELECT node, provider, model, outcome, error, started, ended, tokens_in, tokens_out,"
        " wall_s, cost_usd FROM attempts WHERE job_id=? ORDER BY started, id",
        (job_id,),
    ).fetchall()
    return [dict(r) for r in rows]
