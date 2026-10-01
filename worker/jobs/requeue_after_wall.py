"""Rest a runner after a quota wall and requeue its task (amendment 2026-09-30)."""

import sqlite3
from typing import Any

from worker.config.cooldown_key import cooldown_key
from worker.config.worker_config import WorkerConfig

_DEFAULT_COOLDOWN_S = 3600.0


def requeue_after_wall(
    conn: sqlite3.Connection,
    config: WorkerConfig,
    job: sqlite3.Row,
    executor: dict[str, Any],
    now: float,
) -> str:
    """Return ``queued``; the attempt is not charged and the job waits out the cooldown.

    A wall says nothing about the job, only about the account, so it is
    re-planned rather than failed (spec 8). A profile removed since the lease
    still rests the job for the default cooldown. A job whose queue names a
    fallback for its profile is due at once: the next pick runs it elsewhere.
    An inference job sent to a runner keeps its own model, so the runner is
    the reporting executor, and it is due at once for the next rung. Only the
    model that hit the wall rests when the profile pins one.
    """
    inference = job["kind"] == "inference"
    profile = config.profiles.get(job["model"])
    if inference:
        runner, model = str(executor.get("provider") or ""), executor.get("model") or None
    else:
        runner = profile.runner if profile is not None else ""
        model = profile.model if profile is not None else None
    until = now + config.runner_cooldown_s.get(runner, _DEFAULT_COOLDOWN_S)
    if runner:
        conn.execute(
            "INSERT INTO cooldowns (runner, until) VALUES (?, ?)"
            " ON CONFLICT(runner) DO UPDATE SET until=max(until, excluded.until)",
            (cooldown_key(runner, model), until),
        )
    queue = config.queues.get(job["queue"])
    has_fallback = inference or (
        queue is not None and bool(dict(queue.fallbacks).get(job["model"]))
    )
    conn.execute(
        "UPDATE jobs SET state='queued', error='quota_wall', not_before=?, updated=?,"
        " lease_attempt=NULL, lease_expires=NULL WHERE id=?",
        (now if has_fallback else until, now, job["id"]),
    )
    return "queued"
