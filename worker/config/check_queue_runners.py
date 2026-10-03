"""Refuse a runner route that names no configured task profile."""

from collections.abc import Mapping

from worker.config.queue_policy import QueuePolicy
from worker.config.task_profile import TaskProfile


def check_queue_runners(
    queues: Mapping[str, QueuePolicy], profiles: Mapping[str, TaskProfile]
) -> None:
    """Raise ValueError when a queue's runner route or a fallback points at a missing profile."""
    for queue in queues.values():
        route = queue.runner
        if route is not None and any(p not in profiles for p in (route.profile, *route.fallbacks)):
            raise ValueError(f"queue {queue.name}: runner profiles must be configured profiles")
