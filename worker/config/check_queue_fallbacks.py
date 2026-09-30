"""Refuse a fallback profile that would widen what a task may reach."""

from collections.abc import Mapping

from worker.config.queue_policy import QueuePolicy
from worker.config.task_profile import TaskProfile


def check_queue_fallbacks(
    queues: Mapping[str, QueuePolicy], profiles: Mapping[str, TaskProfile]
) -> None:
    """Raise ValueError unless every fallback is granted by its queue and no broader.

    A fallback must accept every privacy class its primary accepts (the job
    was admitted against the primary) and read from the same input root, so
    the same manifest means the same files.
    """
    for queue in queues.values():
        for primary, alternatives in queue.fallbacks:
            names = (primary, *alternatives)
            if any(n not in queue.profiles or n not in profiles for n in names):
                raise ValueError(f"queue {queue.name}: fallbacks must name granted profiles")
            base = profiles[primary]
            for name in alternatives:
                other = profiles[name]
                if not base.privacy <= other.privacy:
                    raise ValueError(f"queue {queue.name}: {name} must accept {primary}'s privacy")
                if base.input_root is not None and other.input_root != base.input_root:
                    raise ValueError(
                        f"queue {queue.name}: {name} must share {primary}'s input_root"
                    )
