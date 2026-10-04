"""The opaque ``next`` cursor of the admin job listing."""

import sqlite3


def encode_jobs_cursor(last: sqlite3.Row) -> str:
    """``<created>:<id>`` of the page's last row; ``repr`` round-trips the float exactly."""
    return f"{last['created']!r}:{last['id']}"
