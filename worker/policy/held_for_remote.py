"""Whether the local model should leave an inference job to a remote executor."""

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.policy.openrouter_may_take import openrouter_may_take
from worker.policy.runner_route_profile import runner_route_profile


def held_for_remote(
    c: Candidate,
    config: WorkerConfig,
    cooling: frozenset[str],
    node: tuple[str, str],
    now: float,
) -> bool:
    """Younger than the queue's ``local_after_s`` while a remote route could take it.

    ``node`` is the asking node's name and trust class.
    """
    queue = config.queues.get(c.queue)
    if queue is None or now - c.created >= queue.local_after_s:
        return False
    name, trust = node
    if runner_route_profile(c, config, cooling, trust, name) is not None:
        return True
    return openrouter_may_take(c, config, trust)
