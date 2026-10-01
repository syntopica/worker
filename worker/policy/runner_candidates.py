"""The candidates the task loop may lease: tasks, and inference routed to a runner."""

import dataclasses
from collections.abc import Sequence

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.policy.runner_route_profile import runner_route_profile


def runner_candidates(
    candidates: Sequence[Candidate],
    config: WorkerConfig,
    cooling: frozenset[str],
    node: tuple[str, str],
    now: float,
) -> list[Candidate]:
    """Tasks as they are; an inference job carries its route's profile in ``model``.

    An inference job is dropped while younger than its route's ``after_s`` or
    when the route is unusable; the lease keeps the job's own model. ``node``
    is the asking node's name and trust class.
    """
    kept: list[Candidate] = []
    for c in candidates:
        if c.kind != "inference":
            kept.append(c)
            continue
        queue = config.queues.get(c.queue)
        route = queue.runner if queue is not None else None
        if route is None or now - c.created < route.after_s:
            continue
        profile = runner_route_profile(c, config, cooling, node[1], node[0])
        if profile is not None:
            kept.append(dataclasses.replace(c, model=profile))
    return kept
