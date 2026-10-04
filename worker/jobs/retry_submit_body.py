"""The submit body an admin retry admits: the original job, re-asked."""

import sqlite3
from typing import Any


def retry_submit_body(job: sqlite3.Row, job_input: dict[str, Any]) -> dict[str, Any]:
    """Same queue, kind, class, tier, priority, attempts, model and input; no deadline.

    The key is derived from the original's id, so a repeated retry finds the same job.
    """
    return {
        "contract": 1,
        "kind": job["kind"],
        "queue": job["queue"],
        "idempotency_key": f"retry:{job['id']}",
        "priority": job["priority"],
        "privacy": job["privacy"],
        "tier": job["tier"],
        "max_attempts": job["max_attempts"],
        "requirements": {"models": [job["model"]]},
        "input": job_input,
    }
