"""Columns added after a table's first release: ``(table, column, SQL type)``."""

ADDED_COLUMNS = (
    ("attempts", "provider", "TEXT"),
    ("attempts", "cost_usd", "REAL"),
    ("jobs", "tier", "TEXT NOT NULL DEFAULT 'basic'"),
    ("attempts", "model", "TEXT"),
    ("results", "rating", "TEXT"),
    ("jobs", "shadow_of", "TEXT"),
    ("jobs", "pin", "TEXT"),
    ("jobs", "retry_of", "TEXT"),
)
