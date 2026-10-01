"""Refuse a runner route that names no configured task profile."""

from collections.abc import Mapping

from worker.config.queue_policy import QueuePolicy
from worker.config.task_profile import TaskProfile


def check_queue_runners(
    queues: Mapping[str, QueuePolicy], profiles: Mapping[str, TaskProfile]
) -> None:
    """Raise ValueError when a queue's runner route points at a missing profile."""
    for queue in queues.values():
        if queue.runner is not None and queue.runner.profile not in profiles:
            raise ValueError(f"queue {queue.name}: runner.profile must be a configured profile")
