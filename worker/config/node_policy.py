"""What one machine may run and how it reaches its model server."""

from dataclasses import dataclass

from worker.config.default_min_free_pct import DEFAULT_MIN_FREE_PCT


@dataclass(frozen=True)
class NodePolicy:
    """``trust`` is ``owner``, ``server`` or ``guest``.

    ``min_free_pct`` is the free memory below which a kernel pressure warning
    counts as pressure on this machine.
    """

    name: str
    trust: str
    idle_threshold_s: float
    memory_budget_gb: float
    ollama_url: str
    ollama_launchd_label: str
    min_free_pct: float = DEFAULT_MIN_FREE_PCT
