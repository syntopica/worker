"""Refuse a shadow policy naming a missing model or profile."""

from collections.abc import Mapping

from worker.config.model_pin import ModelPin
from worker.config.queue_policy import QueuePolicy
from worker.config.task_profile import TaskProfile


def check_queue_shadows(
    queues: Mapping[str, QueuePolicy],
    models: Mapping[str, ModelPin],
    profiles: Mapping[str, TaskProfile],
) -> None:
    """Raise ValueError unless every local target, runner target and judge is configured."""
    for queue in queues.values():
        shadow = queue.shadow
        if shadow is None:
            continue
        if shadow.judge not in profiles:
            raise ValueError(f"queue {queue.name}: shadow.judge must be a configured profile")
        for target in shadow.targets:
            kind, _, name = target.partition(":")
            known = models if kind == "ollama" else profiles if kind == "runner" else None
            if known is not None and name not in known:
                raise ValueError(f"queue {queue.name}: shadow target {target} is not configured")
