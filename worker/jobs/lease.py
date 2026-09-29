"""A granted lease, as returned to a node."""

from dataclasses import asdict, dataclass
from typing import Any

LEASE_TTL_S = 60.0


@dataclass(frozen=True)
class Lease:
    """Heartbeat and completion must echo ``attempt_id`` and ``generation``."""

    job_id: str
    attempt_id: str
    generation: int
    model: str
    input: dict[str, Any]
    run_when: str
    ttl_s: float

    def to_json(self) -> dict[str, Any]:
        """The wire form of the lease."""
        return asdict(self)
