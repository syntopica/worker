"""Shape one jobs row as the admin listing's metadata, never content."""

import sqlite3
from typing import Any

from worker.jobs.states import SAMPLING_PRODUCERS


def admin_job_row(r: sqlite3.Row) -> dict[str, Any]:
    """``r`` carries the jobs columns plus ``attempt_count``, its attempt rows."""
    return {
        "id": r["id"],
        "queue": r["queue"],
        "producer": r["producer"],
        "state": r["state"],
        "privacy": r["privacy"],
        "tier": r["tier"],
        "created": r["created"],
        "updated": r["updated"],
        "attempts": r["attempt_count"],
        "last_error": r["error"],
        "acked": r["acked"],
        "retry_of": r["retry_of"],
        "sampling": r["producer"] in SAMPLING_PRODUCERS,
    }
