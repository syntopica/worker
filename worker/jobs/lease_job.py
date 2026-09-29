"""Grant the best eligible job to a node, minting a fenced attempt (spec 6)."""

import sqlite3
import uuid

from worker.config.worker_config import WorkerConfig
from worker.jobs.api_error import ApiError
from worker.jobs.fail_payload_lost import fail_payload_lost
from worker.jobs.lease import LEASE_TTL_S, Lease
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.load_candidates import load_candidates
from worker.jobs.load_job_input import load_job_input
from worker.policy.pick_job import pick_job
from worker.store.transaction import transaction

_MAX_LOST_PER_LEASE = 16


def lease_job(
    conn: sqlite3.Connection, config: WorkerConfig, req: LeaseRequest, now: float
) -> Lease | None:
    """Return a Lease, or None when nothing is eligible.

    A pick whose input is gone fails as ``payload_lost`` and the next one is
    tried, at most ``_MAX_LOST_PER_LEASE`` times in one call.
    """
    node = config.nodes.get(req.node)
    if node is None:
        raise ApiError(403, "unknown_node")
    with transaction(conn):
        candidates = load_candidates(conn, config, node.trust, now)
        recent = {
            r[0]: r[1]
            for r in conn.execute(
                "SELECT j.queue, count(*) FROM attempts a JOIN jobs j ON j.id=a.job_id WHERE a.started>? GROUP BY j.queue",
                (now - 3600,),
            )
        }
        for _ in range(_MAX_LOST_PER_LEASE):
            chosen = pick_job(candidates, req, config, recent, now)
            if chosen is None:
                return None
            job_input = load_job_input(conn, chosen.job_id)
            if job_input is not None:
                break
            fail_payload_lost(conn, chosen.job_id, now)
            candidates = [c for c in candidates if c.job_id != chosen.job_id]
        else:
            return None
        attempt_id = uuid.uuid4().hex
        conn.execute(
            "UPDATE jobs SET state='leased', generation=generation+1, lease_node=?, lease_attempt=?,"
            " lease_expires=?, updated=? WHERE id=?",
            (req.node, attempt_id, now + LEASE_TTL_S, now, chosen.job_id),
        )
        generation = conn.execute(
            "SELECT generation FROM jobs WHERE id=?", (chosen.job_id,)
        ).fetchone()[0]
        conn.execute(
            "INSERT INTO attempts (id, job_id, generation, node, started) VALUES (?, ?, ?, ?, ?)",
            (attempt_id, chosen.job_id, generation, req.node, now),
        )
    run_when = config.queues[chosen.queue].run_when
    return Lease(
        chosen.job_id, attempt_id, generation, chosen.model, job_input, run_when, LEASE_TTL_S
    )
