"""When a queued task's runner stops resting, for the job's producer to see."""

import sqlite3

from worker.config.worker_config import WorkerConfig


def job_cooling_until(
    conn: sqlite3.Connection, config: WorkerConfig, job_id: str, now: float
) -> float | None:
    """The end of the cooldown holding a queued task back, or None if nothing is.

    A producer waiting on one job cannot otherwise tell a queue parked behind a
    quota wall from a busy one, and would wait out its whole budget on every
    job it submitted during the wall.
    """
    job = conn.execute("SELECT kind, model, state FROM jobs WHERE id=?", (job_id,)).fetchone()
    if job is None or job["kind"] != "task" or job["state"] != "queued":
        return None
    profile = config.profiles.get(job["model"])
    if profile is None:
        return None
    row = conn.execute(
        "SELECT until FROM cooldowns WHERE runner=? AND until>?", (profile.runner, now)
    ).fetchone()
    return None if row is None else float(row["until"])
