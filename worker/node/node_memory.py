"""What the node loop remembers between iterations."""

from dataclasses import dataclass, field


@dataclass
class NodeMemory:
    """``owned`` holds the models this node made resident or released for pressure.

    Only those are ever unloaded on pressure: a model resident at startup or
    loaded by another app belongs to someone else.
    """

    normal_since: float | None = None
    pressure_streak: int = 0
    unloaded: bool = False
    failed_model: str | None = None
    owned: set[str] = field(default_factory=set)
    backoff_until: float = 0.0
    backoff_s: float = 900.0
