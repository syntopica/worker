"""What one machine may run and how it reaches its model server."""

from dataclasses import dataclass


@dataclass(frozen=True)
class NodePolicy:
    """``trust`` is ``owner``, ``server`` or ``guest``."""

    name: str
    trust: str
    idle_threshold_s: float
    memory_budget_gb: float
    ollama_url: str
    ollama_launchd_label: str
