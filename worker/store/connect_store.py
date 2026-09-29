"""Connect to the coordinator's two SQLite files, without touching the schema."""

import os
import sqlite3
import threading
from pathlib import Path

_UMASK_LOCK = threading.Lock()


def connect_store(state_dir: Path) -> sqlite3.Connection:
    """Metadata in ``meta.sqlite3``; content in ``payloads.sqlite3``, attached as ``p``.

    Only the steps that may create a file run under umask 077 (and its lock);
    sqlite gives later ``-wal``/``-shm`` files the mode of their database file.
    No statement here waits for the writer lock.
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
            conn.execute("ATTACH DATABASE ? AS p", (str(state_dir / "payloads.sqlite3"),))
        finally:
            os.umask(previous)
    conn.row_factory = sqlite3.Row
    conn.execute("PRAGMA journal_mode=WAL")
    conn.execute("PRAGMA synchronous=FULL")
    conn.execute("PRAGMA p.journal_mode=WAL")
    conn.execute("PRAGMA p.synchronous=FULL")
    conn.execute("PRAGMA p.secure_delete=ON")
    conn.execute("PRAGMA p.journal_size_limit=0")
    return conn
