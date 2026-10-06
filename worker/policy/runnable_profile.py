"""The profile a queued task would run under on this node right now."""

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.policy.profile_resting import profile_resting


def runnable_profile(
    c: Candidate, node: str, config: WorkerConfig, cooling: frozenset[str]
) -> str | None:
    """The task's own profile, else its queue's first usable fallback, else None.

    Usable means configured, not on demand, allowed on ``node`` and with a
    runner that is not resting after a quota wall. A pinned job never runs under a fallback.
    """
    queue = config.queues.get(c.queue)
    if queue is None:
        return None
    order = (c.model,) if c.pin else (c.model, *dict(queue.fallbacks).get(c.model, ()))
    for name in order:
        profile = config.profiles.get(name)
        if profile is None or profile.on_demand or profile_resting(profile, cooling):
            continue
        if profile.nodes is None or node in profile.nodes:
            return name
    return None
