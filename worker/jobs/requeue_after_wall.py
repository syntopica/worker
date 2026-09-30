"""Rest a runner after a quota wall and requeue its task (amendment 2026-09-30)."""

import sqlite3

from worker.config.worker_config import WorkerConfig

_DEFAULT_COOLDOWN_S = 3600.0


def requeue_after_wall(
    conn: sqlite3.Connection, config: WorkerConfig, job: sqlite3.Row, now: float
) -> str:
    """Return ``queued``; the attempt is not charged and the job waits out the cooldown.

    A wall says nothing about the job, only about the account, so it is
    re-planned rather than failed (spec 8). A profile removed since the lease
    still rests the job for the default cooldown.
    """
    profile = config.profiles.get(job["model"])
    runner = profile.runner if profile is not None else ""
    until = now + config.runner_cooldown_s.get(runner, _DEFAULT_COOLDOWN_S)
    if runner:
        conn.execute(
            "INSERT INTO cooldowns (runner, until) VALUES (?, ?)"
            " ON CONFLICT(runner) DO UPDATE SET until=max(until, excluded.until)",
            (runner, until),
        )
    conn.execute(
        "UPDATE jobs SET state='queued', error='quota_wall', not_before=?, updated=?,"
        " lease_attempt=NULL, lease_expires=NULL WHERE id=?",
        (until, now, job["id"]),
    )
    return "queued"
