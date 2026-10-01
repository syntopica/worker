"""Move shadow groups forward: queue their judges, then store the verdicts."""

import sqlite3

from worker.config.worker_config import WorkerConfig
from worker.jobs.cancel_stale_shadows import cancel_stale_shadows
from worker.jobs.record_judgements_batch import record_judgements_batch
from worker.jobs.submit_judges_batch import submit_judges_batch
from worker.store.transaction import transaction

_BATCH = 50


def advance_shadow_groups(conn: sqlite3.Connection, config: WorkerConfig, now: float) -> int:
    """Return how many groups were judged or recorded; one short transaction per batch."""
    with transaction(conn):
        cancel_stale_shadows(conn, now)
    moved = 0
    for record in (False, True):
        while True:
            with transaction(conn):
                if record:
                    done = record_judgements_batch(conn, now, _BATCH)
                else:
                    done = submit_judges_batch(conn, config, now, _BATCH)
            moved += done
            if done < _BATCH:
                break
    return moved
