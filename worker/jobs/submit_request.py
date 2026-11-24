"""A validated ``POST /v1/jobs`` body."""

from dataclasses import dataclass
from typing import Any


@dataclass(frozen=True)
class SubmitRequest:
    """``models`` is the ordered preference list from ``requirements``; empty for a task.

    ``tier`` is ``basic`` or ``strong``; the queue decides what each runs on.
    ``exclude_providers`` are the executor providers that must never run the job.
    """

    queue: str
    idempotency_key: str
    priority: int
    privacy: str
    max_attempts: int
    deadline: float | None
    models: tuple[str, ...]
    input: dict[str, Any]
    parent_id: str | None
    kind: str = "inference"
    tier: str = "basic"
    exclude_providers: tuple[str, ...] = ()
