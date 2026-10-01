"""Queue a judge task for each shadow group whose members have all finished."""

import json
import random
import sqlite3
import string

from worker.config.worker_config import WorkerConfig
from worker.jobs.insert_shadow_job import insert_shadow_job
from worker.jobs.judge_task_input import judge_task_input
from worker.jobs.load_job_input import load_job_input

_UNFINISHED = "('queued', 'leased', 'running', 'draining', 'split_requested')"


def submit_judges_batch(
    conn: sqlite3.Connection, config: WorkerConfig, now: float, limit: int
) -> int:
    """Return how many groups were settled: judged, or closed with too few answers.

    Letters are shuffled per group, so position never tells the judge which
    executor wrote an answer; the letter map stays in the judge's own input.
    A group with fewer than two answers gets a cancelled judge as its marker.
    """
    groups = conn.execute(
        "SELECT shadow_of FROM jobs WHERE shadow_of IS NOT NULL GROUP BY shadow_of"  # noqa: S608
        f" HAVING sum(pin='judge')=0 AND sum(state IN {_UNFINISHED})=0 LIMIT ?",
        (limit,),
    ).fetchall()
    for (group,) in groups:
        rows = conn.execute(
            "SELECT j.id, j.queue, j.privacy, j.tier, r.executor, o.body FROM jobs j"
            " JOIN results r ON r.job_id=j.id JOIN p.outputs o ON o.result_id=r.result_id"
            " WHERE j.shadow_of=? AND j.state='succeeded' ORDER BY j.id",
            (group,),
        ).fetchall()
        first = conn.execute("SELECT * FROM jobs WHERE shadow_of=? LIMIT 1", (group,)).fetchone()
        queue = config.queues.get(first["queue"])
        job_input = load_job_input(conn, rows[0]["id"]) if rows else None
        if len(rows) < 2 or queue is None or queue.shadow is None or job_input is None:  # noqa: PLR2004
            judge = insert_shadow_job(
                conn,
                first,
                {},
                group=group,
                pin="judge",
                model="",
                kind="task",
                producer="_judge",
                now=now,
            )
            conn.execute("UPDATE jobs SET state='cancelled', finished=? WHERE id=?", (now, judge))
            continue
        order = list(range(len(rows)))
        random.Random(group).shuffle(order)  # noqa: S311 - blinding, not secrecy
        letters = string.ascii_uppercase[: len(rows)]
        answers = {
            letters[i]: str(json.loads(rows[k]["body"]).get("text", ""))
            for i, k in enumerate(order)
        }
        labels = {letters[i]: json.loads(rows[k]["executor"]) for i, k in enumerate(order)}
        task = {**judge_task_input(job_input["messages"], answers), "labels": labels}
        insert_shadow_job(
            conn,
            first,
            task,
            group=group,
            pin="judge",
            model=queue.shadow.judge,
            kind="task",
            producer="_judge",
            max_attempts=2,
            now=now,
        )
    return len(groups)
