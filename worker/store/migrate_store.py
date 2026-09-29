"""Bring an existing store forward to the current schema, in place."""

import sqlite3

from worker.store.index_sql import INDEXES
from worker.store.transaction import transaction


def migrate_store(conn: sqlite3.Connection) -> None:
    """Forward and idempotent: add what is missing, never drop or rewrite data.

    A store from before ``payloads_deleted`` gains the column; finished jobs
    whose input is already gone are marked, so the first sweep after the
    upgrade does not revisit them.
    """
    with transaction(conn):
        columns = {r[1] for r in conn.execute("PRAGMA main.table_info(jobs)")}
        if "payloads_deleted" not in columns:
            conn.execute("ALTER TABLE jobs ADD COLUMN payloads_deleted INTEGER NOT NULL DEFAULT 0")
            conn.execute(
                "UPDATE jobs SET payloads_deleted=1 WHERE finished IS NOT NULL"
                " AND id NOT IN (SELECT job_id FROM p.inputs)"
            )
        for statement in INDEXES:
            conn.execute(statement)
