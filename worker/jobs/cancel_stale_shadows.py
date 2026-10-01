"""Close shadow members and judges no executor took within a day."""

import sqlite3

_STALE_S = 86400.0


def cancel_stale_shadows(conn: sqlite3.Connection, now: float) -> int:
    """Return how many were cancelled, so an unreachable target cannot hold a group open."""
    return conn.execute(
        "UPDATE jobs SET state='cancelled', finished=?, updated=?"
        " WHERE producer IN ('_shadow', '_judge') AND state='queued' AND created<?",
        (now, now, now - _STALE_S),
    ).rowcount
