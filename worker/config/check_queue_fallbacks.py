"""Refuse a fallback profile that would widen what a task may reach."""

from collections.abc import Mapping

from worker.config.check_stand_in import check_stand_in
from worker.config.queue_policy import QueuePolicy
from worker.config.task_profile import TaskProfile


def check_queue_fallbacks(
    queues: Mapping[str, QueuePolicy], profiles: Mapping[str, TaskProfile]
) -> None:
    """Raise ValueError unless every fallback is granted by its queue and no broader."""
    for queue in queues.values():
        for primary, alternatives in queue.fallbacks:
            if primary not in queue.profiles or primary not in profiles:
                raise ValueError(f"queue {queue.name}: fallbacks must name granted profiles")
            for name in alternatives:
                check_stand_in(queue, primary, name, profiles)
