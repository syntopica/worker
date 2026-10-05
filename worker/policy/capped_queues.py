"""Queues that spent their daily OpenRouter cap."""

from collections.abc import Mapping

from worker.config.worker_config import WorkerConfig


def capped_queues(config: WorkerConfig, spent: Mapping[str, int]) -> frozenset[str]:
    """Queues whose route sets ``daily_cap`` and whose ``spent`` count reached it."""
    capped: set[str] = set()
    for name, queue in config.queues.items():
        route = queue.openrouter
        cap = route.daily_cap if route is not None else None
        if cap is not None and spent.get(name, 0) >= cap:
            capped.add(name)
    return frozenset(capped)
