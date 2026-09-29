"""The retention policy of every queue that has jobs, configured or not."""

import sqlite3
import sys

from worker.config.parse_queue_policy import parse_queue_policy
from worker.config.queue_policy import QueuePolicy
from worker.config.worker_config import WorkerConfig


def retention_policies(conn: sqlite3.Connection, config: WorkerConfig) -> list[QueuePolicy]:
    """A queue no longer in the configuration keeps the defaults and is named once per pass.

    The distinct queues are walked through the ``jobs_queue`` index, one seek
    per queue, so the pass never scans the jobs table.
    """
    policies = dict(config.queues)
    name = conn.execute("SELECT min(queue) FROM jobs").fetchone()[0]
    while name is not None:
        if name not in policies:
            print(
                f"worker: queue {name} is not configured; default retention applies",
                file=sys.stderr,
            )
            policies[name] = parse_queue_policy(name, {})
        name = conn.execute("SELECT min(queue) FROM jobs WHERE queue>?", (name,)).fetchone()[0]
    return list(policies.values())
