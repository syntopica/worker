"""Shape one jobs row as the admin listing's metadata, never content."""

import sqlite3
from typing import Any

from worker.jobs.states import SAMPLING_PRODUCERS


def admin_job_row(r: sqlite3.Row) -> dict[str, Any]:
    """``r`` carries the jobs columns, ``attempt_count`` and the attempt totals.

    Token, cost and wall totals sum the settled attempts and are null when
    none reported one; ``last_*`` describe the newest attempt, running or not.
    """
    return {
        "id": r["id"],
        "queue": r["queue"],
        "producer": r["producer"],
        "kind": r["kind"],
        "model": r["model"],
        "priority": r["priority"],
        "state": r["state"],
        "privacy": r["privacy"],
        "tier": r["tier"],
        "created": r["created"],
        "updated": r["updated"],
        "finished": r["finished"],
        "deadline": r["deadline"],
        "lease_node": r["lease_node"],
        "lease_expires": r["lease_expires"],
        "parent_id": r["parent_id"],
        "preemptions": r["preemptions"],
        "attempts": r["attempt_count"],
        "tokens_in": r["tokens_in"],
        "tokens_out": r["tokens_out"],
        "cost_usd": r["cost_usd"],
        "wall_s": r["wall_s"],
        "last_model": r["last_model"],
        "last_provider": r["last_provider"],
        "last_started": r["last_started"],
        "last_outcome": r["last_outcome"],
        "last_error": r["error"],
        "acked": r["acked"],
        "retry_of": r["retry_of"],
        "sampling": r["producer"] in SAMPLING_PRODUCERS,
    }
