"""Read the ready jobs a node of one trust class may see (privacy applied here)."""

import sqlite3

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.policy.privacy_allows import privacy_allows

_PER_QUEUE = 200


def load_candidates(
    conn: sqlite3.Connection, config: WorkerConfig, trust: str, now: float, kind: str = "inference"
) -> list[Candidate]:
    """Queued, due and not past deadline, of one kind, from configured queues only.

    At most ``_PER_QUEUE`` per queue, best priority first: a single row cap
    let one queue's backlog - or a removed queue's - hide every other queue.
    Inference runs on the ``ollama`` executor and tasks on ``runner`` (spec 8);
    the ``openrouter`` loop sees inference jobs its executor may take.
    """
    executor = {"task": "runner", "openrouter": "openrouter"}.get(kind, "ollama")
    kind = "inference" if kind == "openrouter" else kind
    allowed = [c for c in config.privacy if privacy_allows(config, c, executor, trust)]
    queues = list(config.queues)
    if not allowed or not queues:
        return []
    marks = ",".join("?" * len(allowed))
    queue_marks = ",".join("?" * len(queues))
    rows = conn.execute(
        "SELECT id, queue, model, priority, privacy, created, parked, parked_min_idle_s FROM ("  # noqa: S608
        " SELECT *, row_number() OVER (PARTITION BY queue ORDER BY priority DESC, created) AS rank"
        " FROM jobs WHERE state='queued' AND kind=? AND not_before<=?"
        f" AND (deadline IS NULL OR deadline>?) AND privacy IN ({marks}) AND queue IN ({queue_marks})"
        ") WHERE rank<=? ORDER BY priority DESC, created",
        (kind, now, now, *allowed, *queues, _PER_QUEUE),
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
        if privacy_allows(config, r["privacy"], executor, trust)
    ]
