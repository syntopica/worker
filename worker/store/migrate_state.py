"""Migrate the store in a state directory, for the entry points that own it."""

from pathlib import Path

from worker.store.connect_store import connect_store
from worker.store.migrate_store import migrate_store
from worker.store.protect_state_files import protect_state_files


def migrate_state(state_dir: Path) -> None:
    """Create or upgrade the store once; per-request opens never migrate."""
    conn = connect_store(state_dir)
    try:
        migrate_store(conn)
    finally:
        conn.close()
    protect_state_files(state_dir)
