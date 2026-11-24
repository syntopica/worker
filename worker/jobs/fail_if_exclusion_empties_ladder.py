"""Fail a just-admitted job whose excluded providers leave its ladder empty."""

import json
import sqlite3

from worker.config.worker_config import WorkerConfig
from worker.jobs.finish_failed import finish_failed
from worker.policy.ladder_providers import ladder_providers


def fail_if_exclusion_empties_ladder(
    conn: sqlite3.Connection, config: WorkerConfig, job_id: str, now: float
) -> None:
    """Run inside the submit transaction; fail the job as ``no_eligible_executor``.

    Only an exclusion that removes every provider the ladder had counts: a
    ladder already empty without it waits as before. No other executor is tried.
    """
    job = conn.execute("SELECT * FROM jobs WHERE id=?", (job_id,)).fetchone()
    excluded = set(json.loads(job["exclude_providers"] or "[]"))
    ladder = ladder_providers(config, job["queue"], job["privacy"], job["model"])
    if excluded and ladder and not ladder - excluded:
        finish_failed(conn, job, "no_eligible_executor", now)
