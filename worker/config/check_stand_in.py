"""Refuse a profile that would widen what a task admitted under another may reach."""

from collections.abc import Mapping

from worker.config.queue_policy import QueuePolicy
from worker.config.task_profile import TaskProfile


def check_stand_in(
    queue: QueuePolicy, primary: str, other: str, profiles: Mapping[str, TaskProfile]
) -> None:
    """Raise ValueError unless ``other`` is granted and no broader than ``primary``.

    It must accept every privacy class ``primary`` accepts (the job was
    admitted against ``primary``) and read from the same input root, so the
    same manifest means the same files.
    """
    if any(n not in queue.profiles or n not in profiles for n in (primary, other)):
        raise ValueError(f"queue {queue.name}: {primary} and {other} must be granted profiles")
    base = profiles[primary]
    if not base.privacy <= profiles[other].privacy:
        raise ValueError(f"queue {queue.name}: {other} must accept {primary}'s privacy")
    if base.input_root is not None and profiles[other].input_root != base.input_root:
        raise ValueError(f"queue {queue.name}: {other} must share {primary}'s input_root")
