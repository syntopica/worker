"""OpenRouter attempts per queue since 00:00 UTC, for the daily cap."""

import sqlite3

_DAY_S = 86400


def openrouter_spent_today(conn: sqlite3.Connection, now: float) -> dict[str, int]:
    """Attempts recorded with provider ``openrouter`` today, by queue.

    Only attempts whose provider is recorded count; those still in flight do
    not, which is the overshoot the cap allows (spec amendment 2026-10-05).
    """
    midnight = now - now % _DAY_S
    return {
        row[0]: row[1]
        for row in conn.execute(
            "SELECT j.queue, count(*) FROM attempts a JOIN jobs j ON j.id=a.job_id"
            " WHERE a.provider='openrouter' AND a.started>=? GROUP BY j.queue",
            (midnight,),
        )
    }
