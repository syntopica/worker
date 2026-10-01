"""A queued job as the scheduler sees it: no payload, only what ranking needs."""

from dataclasses import dataclass


@dataclass(frozen=True)
class Candidate:
    """``parked_min_idle_s`` is the current idle a parked job needs before it runs.

    ``pin`` names the only executor a shadow member may run on.
    """

    job_id: str
    queue: str
    model: str
    priority: int
    privacy: str
    created: float
    parked: bool
    parked_min_idle_s: float
    kind: str = "inference"
    pin: str | None = None
