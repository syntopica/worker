"""The submit body an admin retry admits: the original job, re-asked."""

import json
import sqlite3
from typing import Any


def retry_submit_body(job: sqlite3.Row, job_input: dict[str, Any]) -> dict[str, Any]:
    """Same queue, kind, class, tier, priority, attempts, model, exclusions and input; no deadline.

    The key is derived from the original's id, so a repeated retry finds the same job.
    """
    excluded = json.loads(job["exclude_providers"] or "[]")
    return {
        "contract": 1,
        "kind": job["kind"],
        "queue": job["queue"],
        "idempotency_key": f"retry:{job['id']}",
        "priority": job["priority"],
        "privacy": job["privacy"],
        "tier": job["tier"],
        "max_attempts": job["max_attempts"],
        "requirements": {"models": [job["model"]], "exclude_providers": excluded},
        "input": job_input,
    }
