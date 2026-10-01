"""Open a shadow group for an answer just accepted (amendment: shadow sampling)."""

import json
import sqlite3
import uuid
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.jobs.completion_report import CompletionReport
from worker.jobs.insert_shadow_job import insert_shadow_job
from worker.jobs.shadow_sampled import shadow_sampled
from worker.jobs.target_allowed import target_allowed
from worker.jobs.target_answered import target_answered

_UNFINISHED = ("queued", "leased", "running", "draining", "split_requested")


def open_shadow_group(  # noqa: PLR0913, PLR0917
    conn: sqlite3.Connection,
    config: WorkerConfig,
    job: sqlite3.Row,
    job_input: dict[str, Any],
    report: CompletionReport,
    now: float,
) -> int:
    """Return how many members were queued; 0 when the job is not shadowed.

    The original answer is kept as an already succeeded member, so the
    producer's ack, which removes its own copy, cannot take it from the judge.
    """
    queue = config.queues.get(job["queue"])
    shadow = queue.shadow if queue is not None else None
    if shadow is None or job["kind"] != "inference" or job["shadow_of"] is not None:
        return 0
    if not shadow_sampled(job["id"], shadow.rate):
        return 0
    if not target_allowed(config, job["privacy"], f"runner:{shadow.judge}"):
        return 0
    marks = ",".join("?" * len(_UNFINISHED))
    pending = conn.execute(
        f"SELECT count(*) FROM jobs WHERE queue=? AND producer='_shadow' AND state IN ({marks})",  # noqa: S608
        (job["queue"], *_UNFINISHED),
    ).fetchone()[0]
    targets = [
        t
        for t in shadow.targets
        if target_allowed(config, job["privacy"], t)
        and not target_answered(config, t, report.executor)
    ]
    if not targets or pending + len(targets) > shadow.max_pending:
        return 0
    for target in targets:
        kind, _, name = target.partition(":")
        model = name if kind == "ollama" else job["model"]
        insert_shadow_job(conn, job, job_input, group=job["id"], pin=target, model=model, now=now)
    original = insert_shadow_job(
        conn, job, job_input, group=job["id"], pin="original", model=job["model"], now=now
    )
    conn.execute(
        "UPDATE jobs SET state='succeeded', finished=?, attempts=1 WHERE id=?", (now, original)
    )
    result_id = uuid.uuid4().hex
    conn.execute(
        "INSERT INTO results (result_id, job_id, producer, queue, executor, usage, created)"
        " VALUES (?, ?, '_shadow', ?, ?, ?, ?)",
        (
            result_id,
            original,
            job["queue"],
            json.dumps(report.executor),
            json.dumps(report.usage),
            now,
        ),
    )
    conn.execute(
        "INSERT INTO p.outputs (result_id, body) VALUES (?, ?)",
        (result_id, json.dumps(report.output)),
    )
    return len(targets)
