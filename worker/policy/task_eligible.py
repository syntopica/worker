"""Whether one task candidate may run on the requesting node right now."""

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.jobs.lease_request import LeaseRequest


def task_eligible(
    c: Candidate, req: LeaseRequest, config: WorkerConfig, cooling: frozenset[str]
) -> bool:
    """Queue run policy, the profile's node allowlist, and the runner's cooldown.

    No memory check: the runner is a CLI talking to a remote model. A profile
    removed from the configuration runs nothing, like a removed queue.
    """
    queue = config.queues.get(c.queue)
    profile = config.profiles.get(c.model)
    if queue is None or profile is None:
        return False
    if req.user_active and queue.run_when != "active_ok":
        return False
    if profile.nodes is not None and req.node not in profile.nodes:
        return False
    return profile.runner not in cooling
