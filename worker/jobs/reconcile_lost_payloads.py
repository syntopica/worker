"""Startup check: no live job may point at a payload that is gone (spec 9)."""

import sqlite3

from worker.jobs.fail_payload_lost import fail_payload_lost
from worker.store.transaction import transaction

_LIVE_STATES = ("queued", "leased", "running", "draining", "split_requested")


def reconcile_lost_payloads(conn: sqlite3.Connection, now: float) -> int:
    """Fail every non-terminal job whose input is missing; return how many."""
    marks = ",".join("?" * len(_LIVE_STATES))
    with transaction(conn):
        ids = [
            r[0]
            for r in conn.execute(
                f"SELECT id FROM jobs WHERE state IN ({marks})"  # noqa: S608
                " AND id NOT IN (SELECT job_id FROM p.inputs)",
                _LIVE_STATES,
            )
        ]
        for job_id in ids:
            fail_payload_lost(conn, job_id, now)
    return len(ids)
