"""When a queued task's runner stops resting, for the job's producer to see."""

import sqlite3

from worker.config.cooldown_key import cooldown_key
from worker.config.worker_config import WorkerConfig


def job_cooling_until(
    conn: sqlite3.Connection, config: WorkerConfig, job_id: str, now: float
) -> float | None:
    """The end of the cooldown holding a queued task back, or None if nothing is.

    A producer waiting on one job cannot otherwise tell a queue parked behind a
    quota wall from a busy one, and would wait out its whole budget on every
    job it submitted during the wall. With queue fallbacks the job is held
    only while every runner it could use rests, until the first one returns.
    """
    job = conn.execute(
        "SELECT kind, model, queue, state FROM jobs WHERE id=?", (job_id,)
    ).fetchone()
    if job is None or job["kind"] != "task" or job["state"] != "queued":
        return None
    queue = config.queues.get(job["queue"])
    names = (job["model"], *(dict(queue.fallbacks).get(job["model"], ()) if queue else ()))
    profiles = [config.profiles[n] for n in names if n in config.profiles]
    if not profiles:
        return None
    ends = []
    for profile in profiles:
        keys = (profile.runner, cooldown_key(profile.runner, profile.model))
        row = conn.execute(
            "SELECT max(until) FROM cooldowns WHERE runner IN (?, ?) AND until>?", (*keys, now)
        ).fetchone()
        if row[0] is None:
            return None
        ends.append(float(row[0]))
    return min(ends)
