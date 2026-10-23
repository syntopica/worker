"""Rest a runner after a quota wall and requeue its task (amendment 2026-09-30)."""

import sqlite3

from worker.config.cooldown_key import cooldown_key
from worker.config.worker_config import WorkerConfig
from worker.jobs.completion_report import CompletionReport
from worker.jobs.trusted_wall_reset import trusted_wall_reset

_DEFAULT_COOLDOWN_S = 3600.0


def requeue_after_wall(
    conn: sqlite3.Connection,
    config: WorkerConfig,
    job: sqlite3.Row,
    report: CompletionReport,
    now: float,
) -> str:
    """Return ``queued``; the attempt is not charged and the job waits out the cooldown.

    A wall says nothing about the job, only about the account, so it is
    re-planned rather than failed (spec 8). A profile removed since the lease
    still rests the job for the default cooldown. A job whose queue names a
    fallback for its profile is due at once: the next pick runs it elsewhere.
    An inference job sent to a runner keeps its own model, so the runner is
    the reporting executor, and it is due at once for the next rung. Only the
    model that hit the wall rests when the profile pins one. A wall that says
    when it resets (``resets_in_s``) rests the runner that long when it is
    longer than the configured cooldown: agy's weekly quotas outlast it.
    """
    executor = report.executor
    resets_in_s = trusted_wall_reset(report.usage)
    inference = job["kind"] == "inference"
    attempt = conn.execute(
        "SELECT profile FROM attempts WHERE id=?", (job["lease_attempt"],)
    ).fetchone()
    selected = attempt["profile"] if attempt is not None else None
    profile = config.profiles.get(selected or job["model"])
    overridden = selected is not None and selected != job["model"]
    if inference:
        runner, model = str(executor.get("provider") or ""), executor.get("model") or None
    else:
        runner = profile.runner if profile is not None else ""
        model = profile.model if profile is not None else None
    until = now + max(config.runner_cooldown_s.get(runner, _DEFAULT_COOLDOWN_S), resets_in_s or 0.0)
    if runner:
        conn.execute(
            "INSERT INTO cooldowns (runner, until) VALUES (?, ?)"
            " ON CONFLICT(runner) DO UPDATE SET until=max(until, excluded.until)",
            (cooldown_key(runner, model), until),
        )
    queue = config.queues.get(job["queue"])
    has_fallback = (
        inference
        or overridden
        or (queue is not None and bool(dict(queue.fallbacks).get(job["model"])))
    )
    conn.execute(
        "UPDATE jobs SET state='queued', error='quota_wall', not_before=?, updated=?,"
        " lease_attempt=NULL, lease_expires=NULL WHERE id=?",
        (now if has_fallback else until, now, job["id"]),
    )
    return "queued"
