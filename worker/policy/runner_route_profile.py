"""The task profile a queue's runner route would run an inference job under now."""

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.policy.privacy_allows import privacy_allows
from worker.policy.route_profile_usable import route_profile_usable


def runner_route_profile(
    c: Candidate, config: WorkerConfig, cooling: frozenset[str], trust: str, node: str
) -> str | None:
    """The route's profile, else its first usable fallback, else None.

    Usable means configured, allowed on ``node``, its runner or model not
    resting after a quota wall, and the job's class allowed by the profile.
    The class must also be allowed for ``runner`` executors. Other rungs ask
    with their own node: a profile pinned elsewhere does not hold them back.
    A job a runner answered with nothing skips the rung: the next executor
    takes it.
    """
    if c.error == "no_output":
        return None
    queue = config.queues.get(c.queue)
    route = queue.runner if queue is not None else None
    if route is None or not privacy_allows(config, c.privacy, "runner", trust):
        return None
    for name in (route.profile, *route.fallbacks):
        if route_profile_usable(config.profiles.get(name), c.privacy, cooling, node):
            return name
    return None
