"""How one queue is scheduled and retained."""

from dataclasses import dataclass


@dataclass(frozen=True)
class QueuePolicy:
    """``run_when`` is ``idle`` or ``active_ok``."""

    name: str
    run_when: str
    weight: int
    max_outstanding: int
    retention_days: int
    unacked_ttl_hours: int
    parked_min_idle_s: float
    max_model_age_s: float
