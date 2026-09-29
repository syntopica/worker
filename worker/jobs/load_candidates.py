"""Read the ready jobs a node of one trust class may see (privacy applied here)."""

import sqlite3

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.policy.privacy_allows import privacy_allows


def load_candidates(
    conn: sqlite3.Connection, config: WorkerConfig, trust: str, now: float
) -> list[Candidate]:
    """Queued, due and not past deadline; at most 500, best priority first."""
    rows = conn.execute(
        "SELECT id, queue, model, priority, privacy, created, parked, parked_min_idle_s FROM jobs"
        " WHERE state='queued' AND not_before<=? AND (deadline IS NULL OR deadline>?)"
        " ORDER BY priority DESC, created LIMIT 500",
        (now, now),
    ).fetchall()
    return [
        Candidate(
            r["id"],
            r["queue"],
            r["model"],
            r["priority"],
            r["privacy"],
            r["created"],
            bool(r["parked"]),
            r["parked_min_idle_s"],
        )
        for r in rows
        if privacy_allows(config, r["privacy"], "ollama", trust)
    ]
