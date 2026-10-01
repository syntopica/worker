"""Judged answer quality by queue and executor (amendment: shadow sampling)."""

import sqlite3
from typing import Any


def read_judgements(conn: sqlite3.Connection, since: float) -> list[dict[str, Any]]:
    """One row per queue, provider and model judged since ``since``, best mean first."""
    rows = conn.execute(
        "SELECT queue, provider, model, count(*) judged, round(avg(score), 2) mean_score,"
        " sum(best) best FROM judgements WHERE created>?"
        " GROUP BY 1, 2, 3 ORDER BY 1, 5 DESC",
        (since,),
    ).fetchall()
    return [dict(r) for r in rows]
