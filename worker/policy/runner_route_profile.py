"""The task profile a queue's runner route would run an inference job under now."""

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.policy.privacy_allows import privacy_allows
from worker.policy.profile_resting import profile_resting


def runner_route_profile(
    c: Candidate, config: WorkerConfig, cooling: frozenset[str], trust: str, node: str
) -> str | None:
    """The route's profile, or None when the queue has none or it is unusable.

    Usable means configured, allowed on ``node``, its runner not resting
    after a quota wall, and the job's class allowed both by the profile and
    for ``runner`` executors. Other rungs ask with their own node: a profile
    pinned elsewhere does not hold them back. A job a runner answered
    with nothing skips the rung: the next executor takes it.
    """
    if c.error == "no_output":
        return None
    queue = config.queues.get(c.queue)
    route = queue.runner if queue is not None else None
    profile = config.profiles.get(route.profile) if route is not None else None
    if route is None or profile is None or profile_resting(profile, cooling):
        return None
    if profile.nodes is not None and node not in profile.nodes:
        return None
    if c.privacy not in profile.privacy:
        return None
    if not privacy_allows(config, c.privacy, "runner", trust):
        return None
    return route.profile
