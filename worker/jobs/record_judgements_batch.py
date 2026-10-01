"""Store the scores of each finished judge, mapped back to the executors."""

import json
import sqlite3

from worker.jobs.load_job_input import load_job_input


def record_judgements_batch(conn: sqlite3.Connection, now: float, limit: int) -> int:
    """Return how many judges were recorded; a group is recorded once."""
    judges = conn.execute(
        "SELECT j.id, j.shadow_of, j.queue, o.body FROM jobs j"
        " JOIN results r ON r.job_id=j.id JOIN p.outputs o ON o.result_id=r.result_id"
        " WHERE j.pin='judge' AND j.state='succeeded'"
        " AND j.shadow_of NOT IN (SELECT group_id FROM judgements) LIMIT ?",
        (limit,),
    ).fetchall()
    for judge in judges:
        task = load_job_input(conn, judge["id"]) or {}
        verdict = json.loads(judge["body"]).get("json") or {}
        scores = verdict.get("scores") or {}
        for letter, executor in (task.get("labels") or {}).items():
            conn.execute(
                "INSERT INTO judgements (group_id, queue, provider, model, score, best, created)"
                " VALUES (?, ?, ?, ?, ?, ?, ?)",
                (
                    judge["shadow_of"],
                    judge["queue"],
                    str(executor.get("provider") or "unknown"),
                    str(executor.get("model") or "unknown"),
                    int(scores.get(letter, 0)),
                    int(verdict.get("best") == letter),
                    now,
                ),
            )
    return len(judges)
