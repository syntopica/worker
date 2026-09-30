"""Whether one task candidate may run on the requesting node right now."""

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.jobs.lease_request import LeaseRequest
from worker.policy.runnable_profile import runnable_profile


def task_eligible(
    c: Candidate, req: LeaseRequest, config: WorkerConfig, cooling: frozenset[str]
) -> bool:
    """Queue run policy, then a usable profile: its own or a queue fallback.

    No memory check: the runner is a CLI talking to a remote model. A profile
    removed from the configuration runs nothing, like a removed queue.
    """
    queue = config.queues.get(c.queue)
    if queue is None or c.model not in config.profiles:
        return False
    if req.user_active and queue.run_when != "active_ok":
        return False
    return runnable_profile(c, req.node, config, cooling) is not None
