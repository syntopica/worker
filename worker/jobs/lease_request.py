"""What a node reports when it asks for work."""

from dataclasses import dataclass


@dataclass(frozen=True)
class LeaseRequest:
    """``free_gb`` is the node budget minus the resident model's footprint.

    ``kind`` is the job kind this loop runs: ``inference`` or ``task``.
    ``profile`` names the on-demand profile a task loop was started for.
    """

    node: str
    resident_model: str | None
    user_active: bool
    free_gb: float
    current_idle_s: float
    kind: str = "inference"
    profile: str | None = None
