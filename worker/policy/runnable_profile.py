"""The profile a queued task would run under on this node right now."""

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate


def runnable_profile(
    c: Candidate, node: str, config: WorkerConfig, cooling: frozenset[str]
) -> str | None:
    """The task's own profile, else its queue's first usable fallback, else None.

    Usable means configured, allowed on ``node`` and with a runner that is not
    resting after a quota wall.
    """
    queue = config.queues.get(c.queue)
    if queue is None:
        return None
    order = (c.model, *dict(queue.fallbacks).get(c.model, ()))
    for name in order:
        profile = config.profiles.get(name)
        if profile is None or profile.runner in cooling:
            continue
        if profile.nodes is None or node in profile.nodes:
            return name
    return None
