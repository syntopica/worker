"""Refuse a submission when the producer has too many outstanding jobs on a queue."""

import sqlite3

from worker.config.worker_config import WorkerConfig
from worker.jobs.api_error import ApiError
from worker.jobs.states import NOT_OUTSTANDING


def check_outstanding(
    conn: sqlite3.Connection, config: WorkerConfig, producer: str, queue: str
) -> None:
    """Raise ApiError(429) when the queue limit is reached."""
    placeholders = ",".join("?" * len(NOT_OUTSTANDING))
    count = conn.execute(
        f"SELECT count(*) FROM jobs WHERE producer=? AND queue=? AND acked IS NULL AND state NOT IN ({placeholders})",  # noqa: S608
        (producer, queue, *NOT_OUTSTANDING),
    ).fetchone()[0]
    if count >= config.queues[queue].max_outstanding:
        raise ApiError(429, "outstanding_limit")
