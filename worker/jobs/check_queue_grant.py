"""Refuse a job on a queue its producer is not granted, or that is not configured."""

from worker.config.worker_config import WorkerConfig
from worker.jobs.api_error import ApiError


def check_queue_grant(config: WorkerConfig, producer: str, queue: str) -> None:
    """Raise ApiError(403, "queue_not_granted")."""
    if queue not in config.producers.get(producer, frozenset()) or queue not in config.queues:
        raise ApiError(403, "queue_not_granted")
