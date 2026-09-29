"""Reclaim leases whose node went silent, and expire jobs past their deadline."""

import sqlite3

from worker.jobs.finish_failed import finish_failed
from worker.store.transaction import transaction


def expire_leases(conn: sqlite3.Connection, now: float) -> int:
    """Return how many leases were reclaimed. A lost lease counts as an attempt."""
    reclaimed = 0
    with transaction(conn):
        for job in conn.execute(
            "SELECT * FROM jobs WHERE state IN ('leased','running','draining') AND lease_expires<?",
            (now,),
        ).fetchall():
            reclaimed += 1
            conn.execute(
                "UPDATE attempts SET ended=?, outcome='lost' WHERE id=?",
                (now, job["lease_attempt"]),
            )
            if job["attempts"] + 1 >= job["max_attempts"]:
                conn.execute("UPDATE jobs SET attempts=attempts+1 WHERE id=?", (job["id"],))
                finish_failed(conn, job, "lease_lost", now)
                continue
            conn.execute(
                "UPDATE jobs SET state='queued', attempts=attempts+1, lease_attempt=NULL, lease_expires=NULL,"
                " not_before=?, updated=? WHERE id=?",
                (now, now, job["id"]),
            )
        for job in conn.execute(
            "SELECT * FROM jobs WHERE state='queued' AND deadline IS NOT NULL AND deadline<=?",
            (now,),
        ).fetchall():
            finish_failed(conn, job, "deadline", now, control="expired")
    return reclaimed
