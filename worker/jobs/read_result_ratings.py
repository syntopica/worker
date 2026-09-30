"""Producer ratings by queue, tier, provider and model (spec amendment: quality tiers)."""

import sqlite3
from typing import Any


def read_result_ratings(conn: sqlite3.Connection, since: float) -> list[dict[str, Any]]:
    """One row per queue, tier and executor of results created since ``since``."""
    rows = conn.execute(
        "SELECT r.queue queue, coalesce(j.tier, 'basic') tier,"
        " coalesce(json_extract(r.executor, '$.provider'), 'unknown') provider,"
        " coalesce(nullif(json_extract(r.executor, '$.model'), ''), j.model, 'unknown') model,"
        " count(*) results, sum(r.rating IS NOT NULL) rated, sum(r.rating IS 'good') good,"
        " sum(r.rating IS 'edited') edited, sum(r.rating IS 'discarded') discarded"
        " FROM results r LEFT JOIN jobs j ON j.id=r.job_id"
        " WHERE r.created>? AND r.control IS NULL"
        " GROUP BY 1, 2, 3, 4 ORDER BY 1, 2, 3, 4",
        (since,),
    ).fetchall()
    return [dict(r) for r in rows]
