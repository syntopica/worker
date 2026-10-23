"""Add execution profile metadata without changing existing attempts."""

import sqlite3


def migrate_attempt_profiles(conn: sqlite3.Connection) -> None:
    """Schema v9: preserve the source job profile when an attempt overrides it."""
    columns = {row[1] for row in conn.execute("PRAGMA table_info(attempts)")}
    if "profile" not in columns:
        conn.execute("ALTER TABLE attempts ADD COLUMN profile TEXT")
