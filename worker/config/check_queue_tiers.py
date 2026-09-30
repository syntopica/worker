"""Refuse a tier route that names an unknown model or widens a task's reach."""

from collections.abc import Collection, Mapping

from worker.config.check_stand_in import check_stand_in
from worker.config.queue_policy import QueuePolicy
from worker.config.task_profile import TaskProfile


def check_queue_tiers(
    queues: Mapping[str, QueuePolicy],
    models: Collection[str],
    profiles: Mapping[str, TaskProfile],
) -> None:
    """Raise ValueError unless every tier model is configured and every stand-in is allowed."""
    for queue in queues.values():
        for tier, route in queue.tiers:
            unknown = [m for m in route.models if m not in models]
            if unknown:
                raise ValueError(f"queue {queue.name}: tiers.{tier} names unknown models {unknown}")
            for primary, other in route.profiles:
                check_stand_in(queue, primary, other, profiles)
