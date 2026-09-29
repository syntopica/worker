"""Open the coordinator's two SQLite files as one connection."""

import os
import sqlite3
import threading
from pathlib import Path

from worker.store.exclude_payloads_from_backup import exclude_payloads_from_backup
from worker.store.migrate_store import migrate_store
from worker.store.protect_state_files import protect_state_files
from worker.store.schema_sql import SCHEMA

_UMASK_LOCK = threading.Lock()


def open_store(state_dir: Path) -> sqlite3.Connection:
    """Metadata in ``meta.sqlite3``; content in ``payloads.sqlite3``, attached as ``p``.

    Only the metadata file is ever backed up (spec 9), which is why content
    lives in a file of its own rather than in tables a backup cannot exclude.
    Files are created under umask 077 and then forced to 0600; sqlite gives
    later ``-wal``/``-shm`` files the mode of their database file.
    """
    with _UMASK_LOCK:
        previous = os.umask(0o077)
        try:
            state_dir.mkdir(parents=True, exist_ok=True)
            conn = sqlite3.connect(
                state_dir / "meta.sqlite3",
                timeout=10,
                isolation_level=None,
                check_same_thread=False,
            )
            conn.row_factory = sqlite3.Row
            conn.execute("PRAGMA journal_mode=WAL")
            conn.execute("PRAGMA synchronous=FULL")
            conn.execute("ATTACH DATABASE ? AS p", (str(state_dir / "payloads.sqlite3"),))
            conn.execute("PRAGMA p.journal_mode=WAL")
            conn.execute("PRAGMA p.synchronous=FULL")
            conn.execute("PRAGMA p.secure_delete=ON")
            conn.execute("PRAGMA p.journal_size_limit=0")
            conn.executescript(SCHEMA)
            migrate_store(conn)
        finally:
            os.umask(previous)
    protect_state_files(state_dir)
    exclude_payloads_from_backup(state_dir)
    return conn
