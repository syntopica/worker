"""Append one row to the admin audit table: who did what to which job, never content."""

import sqlite3


def record_audit(  # noqa: PLR0913, PLR0917
    conn: sqlite3.Connection, action: str, job_id: str, privacy: str, principal: str, now: float
) -> None:
    """``action`` is ``reveal``, ``cancel``, ``retry`` or ``ack``; inside the caller's transaction, if any."""
    conn.execute(
        "INSERT INTO audit (action, job_id, privacy, principal, created) VALUES (?, ?, ?, ?, ?)",
        (action, job_id, privacy, principal, now),
    )
