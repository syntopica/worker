"""Bring a store forward to the current schema, in place."""

import sqlite3

from worker.store.added_columns import ADDED_COLUMNS
from worker.store.index_sql import INDEXES
from worker.store.schema_sql import SCHEMA
from worker.store.store_version import STORE_VERSION
from worker.store.transaction import transaction


def migrate_store(conn: sqlite3.Connection) -> None:
    """Forward and idempotent: add what is missing, never drop or rewrite data.

    A store from before ``payloads_deleted`` gains the column; finished jobs
    whose input is already gone are marked, so the first sweep after the
    upgrade does not revisit them. Later columns - the ledger's, the quality
    tier, the executor model and the producer rating - are added when absent. ``user_version`` is set last, in the same transaction.
    """
    with transaction(conn):
        for statement in filter(str.strip, SCHEMA.split(";")):
            conn.execute(statement)
        columns = {r[1] for r in conn.execute("PRAGMA main.table_info(jobs)")}
        if "payloads_deleted" not in columns:
            conn.execute("ALTER TABLE jobs ADD COLUMN payloads_deleted INTEGER NOT NULL DEFAULT 0")
            conn.execute(
                "UPDATE jobs SET payloads_deleted=1 WHERE finished IS NOT NULL"
                " AND id NOT IN (SELECT job_id FROM p.inputs)"
            )
        for table, column, sql_type in ADDED_COLUMNS:
            present = {r[1] for r in conn.execute(f"PRAGMA main.table_info({table})")}
            if column not in present:
                conn.execute(f"ALTER TABLE {table} ADD COLUMN {column} {sql_type}")
        for statement in INDEXES:
            conn.execute(statement)
        conn.execute(f"PRAGMA user_version={STORE_VERSION}")
