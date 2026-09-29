"""Apply the retention table of spec 9 in short, bounded transactions."""

import sqlite3

from worker.config.worker_config import WorkerConfig
from worker.jobs.delete_payloads_batch import delete_payloads_batch
from worker.jobs.expire_unacked_batch import expire_unacked_batch
from worker.jobs.release_jobs_batch import release_jobs_batch
from worker.jobs.retention_policies import retention_policies
from worker.store.transaction import transaction

_BATCH = 500


def sweep_retention(conn: sqlite3.Connection, config: WorkerConfig, now: float) -> int:
    """Return how many jobs had their payloads deleted.

    Each step selects only rows it still has to change, so a pass over jobs
    already swept writes nothing, and no transaction holds more than one batch.
    """
    swept = 0
    for policy in retention_policies(conn, config):
        for step in (expire_unacked_batch, delete_payloads_batch, release_jobs_batch):
            while True:
                with transaction(conn):
                    done = step(conn, policy, now, _BATCH)
                if step is not release_jobs_batch:
                    swept += done
                if done < _BATCH:
                    break
    if swept:
        conn.execute("PRAGMA p.wal_checkpoint(TRUNCATE)")
    return swept
