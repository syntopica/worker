"""The executor providers a queue's ladder could ever give a job."""

from worker.config.worker_config import WorkerConfig


def ladder_providers(config: WorkerConfig, queue: str, privacy: str, model: str) -> frozenset[str]:
    """What the local, OpenRouter and runner rungs would report as ``executor.provider``.

    A rung counts when the class may use its executor and the queue routes the
    job there; cooldowns and node trust are runtime state and are not read.
    """
    allowed = config.privacy.get(privacy, frozenset())
    policy = config.queues[queue]
    providers: set[str] = set()
    if "ollama" in allowed:
        providers.add("ollama")
    if "openrouter" in allowed and policy.openrouter and dict(policy.openrouter.models).get(model):
        providers.add("openrouter")
    if "runner" in allowed and policy.runner:
        for name in (policy.runner.profile, *policy.runner.fallbacks):
            profile = config.profiles.get(name)
            if profile is not None and not profile.on_demand and privacy in profile.privacy:
                providers.add(profile.runner)
    return frozenset(providers)
