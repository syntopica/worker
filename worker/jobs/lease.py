"""A granted lease, as returned to a node."""

from dataclasses import asdict, dataclass
from typing import Any

LEASE_TTL_S = 60.0


@dataclass(frozen=True)
class Lease:
    """Heartbeat and completion must echo ``attempt_id`` and ``generation``.

    ``model`` is the executor target: a pinned model, a task's profile, or an
    OpenRouter model id. ``queue`` and ``privacy`` let a remote executor apply
    the queue's route and the class's data-handling rule.
    """

    job_id: str
    attempt_id: str
    generation: int
    model: str
    input: dict[str, Any]
    run_when: str
    ttl_s: float
    kind: str = "inference"
    queue: str = ""
    privacy: str = ""

    def to_json(self) -> dict[str, Any]:
        """The wire form of the lease."""
        return asdict(self)
