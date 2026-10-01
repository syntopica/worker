"""Whether a queue's OpenRouter route could take an inference job."""

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.policy.privacy_allows import privacy_allows


def openrouter_may_take(c: Candidate, config: WorkerConfig, trust: str) -> bool:
    """The route maps the job's model and its class may go to ``openrouter``."""
    queue = config.queues.get(c.queue)
    route = queue.openrouter if queue is not None else None
    if route is None or dict(route.models).get(c.model) is None:
        return False
    return privacy_allows(config, c.privacy, "openrouter", trust)
