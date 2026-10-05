"""Whether the local model should leave an inference job to a remote executor."""

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.policy.openrouter_may_take import openrouter_may_take
from worker.policy.runner_route_profile import runner_route_profile


def held_for_remote(
    c: Candidate,
    config: WorkerConfig,
    resting: tuple[frozenset[str], frozenset[str]],
    node: tuple[str, str],
    now: float,
) -> bool:
    """Younger than the queue's ``local_after_s`` while a remote route could take it.

    ``node`` is the asking node's name and trust class. A shadow member is
    left alone unless it is pinned to the local model, which takes it at once.
    ``resting`` is the cooling runners and the queues past their daily
    OpenRouter cap. That rung cannot take a capped queue's job today, so
    nothing is gained by holding it for OpenRouter.
    """
    cooling, capped = resting
    queue = config.queues.get(c.queue)
    if c.pin is not None:
        return not c.pin.startswith("ollama:")
    if queue is None or now - c.created >= queue.local_after_s:
        return False
    name, trust = node
    if runner_route_profile(c, config, cooling, trust, name) is not None:
        return True
    return c.queue not in capped and openrouter_may_take(c, config, trust)
