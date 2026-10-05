"""Grant the best eligible job to a node, minting a fenced attempt (spec 6)."""

import sqlite3
import uuid

from worker.config.worker_config import WorkerConfig
from worker.jobs.api_error import ApiError
from worker.jobs.cooling_runners import cooling_runners
from worker.jobs.fail_payload_lost import fail_payload_lost
from worker.jobs.inference_task_input import inference_task_input
from worker.jobs.lease import LEASE_TTL_S, Lease
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.load_candidates import load_candidates
from worker.jobs.load_job_input import load_job_input
from worker.jobs.openrouter_spent_today import openrouter_spent_today
from worker.jobs.saturated_runners import saturated_runners
from worker.policy.capped_queues import capped_queues
from worker.policy.held_for_remote import held_for_remote
from worker.policy.pick_job import pick_job
from worker.policy.pick_remote import pick_remote
from worker.policy.pick_task import pick_task
from worker.policy.runner_candidates import runner_candidates
from worker.policy.runner_route_profile import runner_route_profile
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
        candidates = load_candidates(conn, config, node.trust, now, req.kind)
        # A runner at its concurrency cap is passed over like a resting one.
        cooling = cooling_runners(conn, now) | saturated_runners(conn, config)
        # A queue past its daily OpenRouter cap leaves that rung for the day.
        capped = capped_queues(config, openrouter_spent_today(conn, now))
        if req.kind == "task":
            candidates = runner_candidates(candidates, config, cooling, (req.node, node.trust), now)
        elif req.kind == "openrouter":
            candidates = [c for c in candidates if c.queue not in capped]
        else:
            candidates = [
                c
                for c in candidates
                if not held_for_remote(c, config, (cooling, capped), (req.node, node.trust), now)
            ]
        recent = {
            r[0]: r[1]
            for r in conn.execute(
                "SELECT j.queue, count(*) FROM attempts a JOIN jobs j ON j.id=a.job_id WHERE a.started>? GROUP BY j.queue",
                (now - 3600,),
            )
        }
        for _ in range(_MAX_LOST_PER_LEASE):
            if req.kind == "task":
                chosen = pick_task(candidates, req, config, recent, cooling)
            elif req.kind == "openrouter":
                chosen = pick_remote(
                    candidates,
                    config,
                    recent,
                    now,
                    runner_first=lambda c: (
                        runner_route_profile(c, config, cooling, node.trust, req.node) is not None
                    ),
                )
            else:
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
        if req.kind == "task" and chosen.kind == "inference":
            job_input = inference_task_input(job_input)
        attempt_id = uuid.uuid4().hex
        conn.execute(
            # A task picked under a fallback profile keeps it: a wall is then
            # charged to the runner that actually ran. An escalated inference
            # job keeps its local model for a retry at home.
            "UPDATE jobs SET state='leased', generation=generation+1, lease_node=?, lease_attempt=?,"
            " lease_expires=?, updated=?, model=CASE WHEN kind='task' THEN ? ELSE model END WHERE id=?",
            (req.node, attempt_id, now + LEASE_TTL_S, now, chosen.model, chosen.job_id),
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
        chosen.job_id,
        attempt_id,
        generation,
        chosen.model,
        job_input,
        run_when,
        LEASE_TTL_S,
        req.kind,
        chosen.queue,
        chosen.privacy,
    )
