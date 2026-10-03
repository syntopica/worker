"""Finished attempts by time bucket, queue, provider and outcome (spec 13, ``activity``)."""

import sqlite3
from typing import Any

_HOUR_S = 3600.0
_MAX_HOURS = 168
_HOURLY_UP_TO = 48
_COARSE_BUCKET_S = 6 * _HOUR_S


def read_activity(conn: sqlite3.Connection, now: float, hours: float) -> dict[str, Any]:
    """Aggregates only: counts, seconds and tokens, never a job id or content.

    ``hours`` is clamped to 1..168; up to 48 hours buckets are an hour wide,
    longer spans six hours. ``sampling`` marks shadow copies and their judges.
    ``error`` is kept for every outcome but success (a preemption's reason too).
    """
    span = min(max(hours, 1.0), _MAX_HOURS)
    bucket_s = _HOUR_S if span <= _HOURLY_UP_TO else _COARSE_BUCKET_S
    since = now - span * _HOUR_S
    rows = conn.execute(
        "SELECT cast(a.ended / ? AS INTEGER) * ? bucket, j.queue queue,"
        " coalesce(a.provider, 'unknown') provider,"
        " j.producer IN ('_shadow', '_judge') sampling, a.outcome outcome,"
        " CASE WHEN a.outcome = 'succeeded' THEN NULL ELSE a.error END error,"
        " count(*) attempts, sum(coalesce(a.wall_s, 0)) wall_s,"
        " sum(coalesce(a.tokens_in, 0)) tokens_in, sum(coalesce(a.tokens_out, 0)) tokens_out"
        " FROM attempts a JOIN jobs j ON j.id = a.job_id"
        " WHERE a.ended > ? AND a.ended <= ?"
        " GROUP BY 1, 2, 3, 4, 5, 6 ORDER BY 1, 2, 3, 4, 5, 6",
        (bucket_s, bucket_s, since, now),
    ).fetchall()
    return {
        "since": since,
        "bucket_s": bucket_s,
        "rows": [
            {**dict(r), "bucket": float(r["bucket"]), "sampling": bool(r["sampling"])} for r in rows
        ],
    }
