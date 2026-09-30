"""Choose the next task: priority, then weighted share across queues, then age."""

import dataclasses
from collections.abc import Mapping, Sequence

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.jobs.lease_request import LeaseRequest
from worker.policy.runnable_profile import runnable_profile
from worker.policy.task_eligible import task_eligible


def pick_task(
    candidates: Sequence[Candidate],
    req: LeaseRequest,
    config: WorkerConfig,
    recent: Mapping[str, int],
    cooling: frozenset[str],
) -> Candidate | None:
    """Same ordering as ``pick_job`` without model residency, which tasks lack.

    The pick carries the profile it will run under in ``model``: its own, or
    a queue fallback while its runner rests.
    """
    eligible = [c for c in candidates if task_eligible(c, req, config, cooling)]
    if not eligible:
        return None
    top = max(c.priority for c in eligible)
    tier = [c for c in eligible if c.priority == top]
    chosen = min(
        tier, key=lambda c: (recent.get(c.queue, 0) / config.queues[c.queue].weight, c.created)
    )
    profile = runnable_profile(chosen, req.node, config, cooling)
    return dataclasses.replace(chosen, model=profile or chosen.model)
