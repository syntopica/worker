"""Free quality signals by queue, tier, provider and model (spec amendment: quality tiers)."""

import sqlite3
from typing import Any


def read_attempt_quality(conn: sqlite3.Connection, since: float) -> list[dict[str, Any]]:
    """One row per queue, tier, provider and model of attempts ended since ``since``.

    The model is the one the executor reported, else the job's own (a task's
    profile, or an attempt that never completed). Rate-limited and quota-wall
    attempts are not the model's fault and are left out of ``failed``.
    """
    rows = conn.execute(
        "SELECT j.queue queue, j.tier tier, coalesce(a.provider, 'unknown') provider,"
        " coalesce(a.model, j.model) model, count(*) attempts,"
        " sum(a.outcome='succeeded') succeeded,"
        " sum(a.outcome='schema_violation') schema_violations,"
        " sum(a.outcome='failed' AND coalesce(a.error, '') NOT IN ('rate_limited', 'quota_wall'))"
        " failed, sum(a.outcome='preempted') preempted, avg(a.wall_s) mean_wall_s"
        " FROM attempts a JOIN jobs j ON j.id=a.job_id"
        " WHERE a.started>? AND a.ended IS NOT NULL"
        " GROUP BY 1, 2, 3, 4 ORDER BY 1, 2, 3, 4",
        (since,),
    ).fetchall()
    return [dict(r) for r in rows]
