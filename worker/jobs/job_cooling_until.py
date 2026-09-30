"""When a queued task's runner stops resting, for the job's producer to see."""

import sqlite3

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
    runners = {config.profiles[n].runner for n in names if n in config.profiles}
    if not runners:
        return None
    ends = []
    for runner in runners:
        row = conn.execute(
            "SELECT until FROM cooldowns WHERE runner=? AND until>?", (runner, now)
        ).fetchone()
        if row is None:
            return None
        ends.append(float(row["until"]))
    return min(ends)
