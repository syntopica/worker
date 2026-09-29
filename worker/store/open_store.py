"""Open the coordinator's two SQLite files as one connection."""

import sqlite3
from pathlib import Path

from worker.store.schema_sql import SCHEMA


def open_store(state_dir: Path) -> sqlite3.Connection:
    """Metadata in ``meta.sqlite3``; content in ``payloads.sqlite3``, attached as ``p``.

    Only the metadata file is ever backed up (spec 9), which is why content
    lives in a file of its own rather than in tables a backup cannot exclude.
    """
    state_dir.mkdir(parents=True, exist_ok=True)
    conn = sqlite3.connect(
        state_dir / "meta.sqlite3", timeout=10, isolation_level=None, check_same_thread=False
    )
    conn.row_factory = sqlite3.Row
    conn.execute("PRAGMA journal_mode=WAL")
    conn.execute("PRAGMA synchronous=FULL")
    conn.execute("ATTACH DATABASE ? AS p", (str(state_dir / "payloads.sqlite3"),))
    conn.execute("PRAGMA p.journal_mode=WAL")
    conn.execute("PRAGMA p.synchronous=FULL")
    conn.execute("PRAGMA p.secure_delete=ON")
    conn.executescript(SCHEMA)
    return conn
