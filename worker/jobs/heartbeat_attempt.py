"""Renew a lease and record whether its attempt is running or draining."""

import sqlite3

from worker.jobs.check_fence import check_fence
from worker.jobs.lease import LEASE_TTL_S
from worker.store.transaction import transaction


def heartbeat_attempt(
    conn: sqlite3.Connection, attempt_id: str, generation: int, draining: bool, now: float
) -> None:
    """Raise ApiError(409) when fenced out; the node must then stop."""
    with transaction(conn):
        job = check_fence(conn, attempt_id, generation)
        state = "draining" if draining else "running"
        conn.execute(
            "UPDATE jobs SET state=?, lease_expires=?, updated=? WHERE id=?",
            (state, now + LEASE_TTL_S, now, job["id"]),
        )
