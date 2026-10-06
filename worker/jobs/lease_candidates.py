"""Narrow loaded candidates to what the asking loop may lease."""

from collections.abc import Sequence

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.jobs.lease_request import LeaseRequest
from worker.policy.held_for_remote import held_for_remote
from worker.policy.on_demand_candidates import on_demand_candidates
from worker.policy.runner_candidates import runner_candidates


def lease_candidates(  # noqa: PLR0913, PLR0917
    candidates: Sequence[Candidate],
    config: WorkerConfig,
    req: LeaseRequest,
    trust: str,
    walls: tuple[frozenset[str], frozenset[str]],
    now: float,
) -> list[Candidate]:
    """On demand, task, OpenRouter or local inference, by what the loop asked for.

    ``walls`` are the resting runners and the queues past their OpenRouter cap.
    """
    cooling, capped = walls
    node = (req.node, trust)
    if req.profile is not None:
        return on_demand_candidates(candidates, config, cooling, node, req.profile)
    if req.kind == "task":
        return runner_candidates(candidates, config, cooling, node, now)
    if req.kind == "openrouter":
        return [c for c in candidates if c.queue not in capped]
    return [c for c in candidates if not held_for_remote(c, config, walls, node, now)]
