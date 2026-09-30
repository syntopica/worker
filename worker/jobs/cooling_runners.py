"""The runners resting after a quota wall (amendment 2026-09-30, phase 2a)."""

import sqlite3


def cooling_runners(conn: sqlite3.Connection, now: float) -> frozenset[str]:
    """Runners whose cooldown has not ended yet."""
    rows = conn.execute("SELECT runner FROM cooldowns WHERE until>?", (now,)).fetchall()
    return frozenset(r[0] for r in rows)
