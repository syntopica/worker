"""How one queue is scheduled and retained."""

from dataclasses import dataclass, field

from worker.config.open_router_route import OpenRouterRoute


@dataclass(frozen=True)
class QueuePolicy:
    """``run_when`` is ``idle`` or ``active_ok``; ``profiles`` are the task profiles it grants.

    ``fallbacks`` maps a granted profile to the profiles, in order, a task may
    run under while its own runner rests after a quota wall. ``openrouter``
    lets its inference jobs escalate to free remote endpoints.
    """

    name: str
    run_when: str
    weight: int
    max_outstanding: int
    retention_days: int
    unacked_ttl_hours: int
    parked_min_idle_s: float
    max_model_age_s: float
    profiles: frozenset[str] = field(default_factory=frozenset)
    fallbacks: tuple[tuple[str, tuple[str, ...]], ...] = ()
    openrouter: OpenRouterRoute | None = None
