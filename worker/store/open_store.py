"""Open an already migrated store for one request or one sweep pass."""

import sqlite3
from pathlib import Path

from worker.store.connect_store import connect_store
from worker.store.exclude_payloads_from_backup import exclude_payloads_from_backup
from worker.store.protect_state_files import protect_state_files
from worker.store.store_not_migrated_error import StoreNotMigratedError
from worker.store.store_version import STORE_VERSION


def open_store(state_dir: Path) -> sqlite3.Connection:
    """Never migrates and never begins a transaction; only reads ``user_version``.

    Only the metadata file is ever backed up (spec 9), which is why content
    lives in a file of its own. Raise StoreNotMigratedError on an older store:
    the entry points run ``migrate_state`` first.
    """
    conn = connect_store(state_dir)
    found = conn.execute("PRAGMA user_version").fetchone()[0]
    if found < STORE_VERSION:
        conn.close()
        raise StoreNotMigratedError(found, STORE_VERSION)
    protect_state_files(state_dir)
    exclude_payloads_from_backup(state_dir)
    return conn
