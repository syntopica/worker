"""Queues that spent their daily OpenRouter cap."""

from collections.abc import Mapping

from worker.config.worker_config import WorkerConfig


def capped_queues(config: WorkerConfig, spent: Mapping[str, int]) -> frozenset[str]:
    """Queues that reached their own ``daily_cap`` or their ``daily_key_cap``.

    ``daily_key_cap`` is compared with every queue's ``spent`` together.
    """
    capped: set[str] = set()
    key_spent = sum(spent.values())
    for name, queue in config.queues.items():
        route = queue.openrouter
        if route is None:
            continue
        if route.daily_cap is not None and spent.get(name, 0) >= route.daily_cap:
            capped.add(name)
        if route.daily_key_cap is not None and key_spent >= route.daily_key_cap:
            capped.add(name)
    return frozenset(capped)
