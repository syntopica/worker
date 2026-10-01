"""Choose the next inference job to escalate to OpenRouter (spec 8, rung 3)."""

import dataclasses
from collections.abc import Callable, Mapping, Sequence

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate


def pick_remote(
    candidates: Sequence[Candidate],
    config: WorkerConfig,
    recent: Mapping[str, int],
    now: float,
    *,
    runner_first: Callable[[Candidate], bool] = lambda _c: False,
) -> Candidate | None:
    """A job whose queue routes its model to OpenRouter and that waited long enough.

    Same ordering as ``pick_job``: priority, then weighted share, then age.
    The pick carries the remote model id in ``model``; the job keeps its own.
    The route's ``after_s`` applies only while ``runner_first`` says a runner
    would take the job; otherwise it escalates at once (executor ladder). A
    job pinned to an OpenRouter model goes there at once; other pins never.
    """
    eligible: list[tuple[Candidate, str]] = []
    for c in candidates:
        if c.pin is not None:
            kind, _, pinned = c.pin.partition(":")
            if kind == "openrouter":
                eligible.append((c, pinned))
            continue
        queue = config.queues.get(c.queue)
        route = queue.openrouter if queue is not None else None
        if route is None:
            continue
        if runner_first(c) and now - c.created < route.after_s:
            continue
        remote = dict(route.models).get(c.model)
        if remote is not None:
            eligible.append((c, remote))
    if not eligible:
        return None
    top = max(c.priority for c, _ in eligible)
    tier = [(c, r) for c, r in eligible if c.priority == top]
    chosen, remote = min(
        tier,
        key=lambda pair: (
            recent.get(pair[0].queue, 0) / config.queues[pair[0].queue].weight,
            pair[0].created,
        ),
    )
    return dataclasses.replace(chosen, model=remote)
