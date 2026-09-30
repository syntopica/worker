"""The usage ledger by provider, queue and UTC day (spec 13, ``costs``)."""

import sqlite3
from typing import Any


def read_costs(conn: sqlite3.Connection, since: float) -> list[dict[str, Any]]:
    """One row per provider, queue and day of attempts ended since ``since``.

    Failed and cancelled attempts count: they spent quota or money too.
    """
    rows = conn.execute(
        "SELECT coalesce(a.provider, 'unknown') provider, j.queue queue,"
        " date(a.ended, 'unixepoch') day, count(*) attempts,"
        " sum(a.outcome='succeeded') succeeded, sum(coalesce(a.tokens_in, 0)) tokens_in,"
        " sum(coalesce(a.tokens_out, 0)) tokens_out, sum(coalesce(a.cost_usd, 0)) cost_usd,"
        " sum(coalesce(a.wall_s, 0)) wall_s"
        " FROM attempts a JOIN jobs j ON j.id=a.job_id"
        " WHERE a.started>? AND a.ended IS NOT NULL"
        " GROUP BY provider, queue, day ORDER BY day, provider, queue",
        (since,),
    ).fetchall()
    return [dict(r) for r in rows]
