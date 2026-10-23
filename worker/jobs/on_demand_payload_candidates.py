"""Keep task overrides that need no files or self-evaluation."""

import sqlite3
from collections.abc import Sequence

from worker.config.task_profile import TaskProfile
from worker.jobs.candidate import Candidate
from worker.jobs.load_job_input import load_job_input


def on_demand_payload_candidates(
    conn: sqlite3.Connection, candidates: Sequence[Candidate], profile: TaskProfile
) -> list[Candidate]:
    """Filter task manifests before selection; missing payloads use normal reconciliation."""
    kept = []
    for candidate in candidates:
        if candidate.kind == "task":
            task = load_job_input(conn, candidate.job_id) or {}
            if task.get("inputs"):
                continue
            if candidate.pin == "judge" and any(
                label.get("provider") == profile.runner and label.get("model") == profile.model
                for label in task.get("labels", {}).values()
            ):
                continue
        kept.append(candidate)
    return kept
